package twentyapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aniasse/labzyn/server/internal/core"
	"github.com/aniasse/labzyn/server/internal/llm"
	"github.com/aniasse/labzyn/server/internal/metadata"
	"github.com/aniasse/labzyn/server/internal/records"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	gqlhandler "github.com/graphql-go/handler"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UUID scalar (behaves like String) for the core records schema.
var uuidScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:       "UUID",
	Serialize:  func(v any) any { return v },
	ParseValue: func(v any) any { return v },
	ParseLiteral: func(v ast.Value) any {
		if s, ok := v.(*ast.StringValue); ok {
			return s.Value
		}
		return nil
	},
})

// filterIsEnum is nestjs-query's IS comparison (NULL / NOT_NULL).
var filterIsEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "FilterIs",
	Values: graphql.EnumValueConfigMap{
		"NULL":     &graphql.EnumValueConfig{Value: "NULL"},
		"NOT_NULL": &graphql.EnumValueConfig{Value: "NOT_NULL"},
	},
})

// Shared filter comparison inputs (nestjs-query style: eq, in, ilike, …).
func cmpInput(name string, scalar graphql.Input, extra ...string) *graphql.InputObject {
	fields := graphql.InputObjectConfigFieldMap{
		"eq":  &graphql.InputObjectFieldConfig{Type: scalar},
		"neq": &graphql.InputObjectFieldConfig{Type: scalar},
		"in":  &graphql.InputObjectFieldConfig{Type: graphql.NewList(scalar)},
		"is":  &graphql.InputObjectFieldConfig{Type: filterIsEnum},
	}
	for _, e := range extra {
		fields[e] = &graphql.InputObjectFieldConfig{Type: scalar}
	}
	return graphql.NewInputObject(graphql.InputObjectConfig{Name: name, Fields: fields})
}

var (
	uuidFilterCmp   = cmpInput("UUIDFilterComparison", uuidScalar, "gt", "gte", "lt", "lte")
	stringFilterCmp = cmpInput("StringFilterComparison", graphql.String, "ilike", "like", "gt", "gte", "lt", "lte")
	floatFilterCmp  = cmpInput("FloatFilterComparison", graphql.Float, "gt", "gte", "lt", "lte")
	boolFilterCmp   = cmpInput("BooleanFilterComparison", graphql.Boolean)
	dateFilterCmp   = cmpInput("DateFilterComparison", graphql.String, "gt", "gte", "lt", "lte")
)

// fullNameOutput is the FULL_NAME composite field type (firstName/lastName).
var fullNameOutput = graphql.NewObject(graphql.ObjectConfig{
	Name: "FullName",
	Fields: graphql.Fields{
		"firstName": &graphql.Field{Type: graphql.String},
		"lastName":  &graphql.Field{Type: graphql.String},
	},
})

func filterCmpFor(t metadata.FieldType) *graphql.InputObject {
	switch t {
	case metadata.FieldNumber, metadata.FieldNumeric, metadata.FieldRating:
		return floatFilterCmp
	case metadata.FieldBoolean:
		return boolFilterCmp
	case metadata.FieldDate, metadata.FieldDateTime:
		return dateFilterCmp
	default:
		return stringFilterCmp
	}
}

// BuildCoreHandler builds the dynamic records GraphQL handler for /graphql.
func BuildCoreHandler(pool *pgxpool.Pool, llmClient *llm.Client) (http.Handler, error) {
	schema, err := buildCoreSchema(pool, llmClient)
	if err != nil {
		return nil, err
	}
	return gqlhandler.New(&gqlhandler.Config{Schema: &schema, Pretty: true}), nil
}

// buildCoreSchema builds the dynamic per-object records schema (/graphql).
// One query field + connection type per object, resolved from the records repo.
func buildCoreSchema(pool *pgxpool.Pool, llmClient *llm.Client) (graphql.Schema, error) {
	coreRepo := core.NewRepo(pool)
	metaRepo := metadata.NewRepo(pool)
	recRepo := records.NewRepo(pool)
	tableStore := records.NewTableStore(pool)

	pageInfo := graphql.NewObject(graphql.ObjectConfig{
		Name: "PageInfo",
		Fields: graphql.Fields{
			"hasNextPage":     &graphql.Field{Type: graphql.Boolean},
			"hasPreviousPage": &graphql.Field{Type: graphql.Boolean},
			"startCursor":     &graphql.Field{Type: graphql.String},
			"endCursor":       &graphql.Field{Type: graphql.String},
		},
	})

	objs, err := metaRepo.ListAllObjects(context.Background())
	if err != nil {
		return graphql.Schema{}, err
	}

	queryFields := graphql.Fields{
		"ping": &graphql.Field{
			Type:    graphql.String,
			Resolve: func(graphql.ResolveParams) (any, error) { return "labzyn-core", nil },
		},
	}
	mutationFields := graphql.Fields{}

	nodes := map[string]*graphql.Object{}
	seen := map[string]bool{}
	for _, o := range objs {
		fields, err := metaRepo.ListFields(context.Background(), o.ID)
		if err != nil {
			return graphql.Schema{}, err
		}
		seen[o.NamePlural] = true
		nodes[o.NameSingular] = addObjectQuery(queryFields, mutationFields, pageInfo, o, fields, coreRepo, metaRepo, recRepo, tableStore)
	}
	// Twenty's standard objects (people, companies, workspaceMembers): queries
	// resolve empty (they are not DB-backed yet).
	for _, so := range standardObjects() {
		if seen[so.Object.NamePlural] {
			continue
		}
		addObjectQuery(queryFields, mutationFields, pageInfo, so.Object, so.Fields, coreRepo, metaRepo, recRepo, tableStore)
	}

	// Labzyn HPC pipeline + AI agent: submitRun/refreshRun/planRun on the Run node.
	if runNode, ok := nodes["run"]; ok {
		addRunHpcMutations(mutationFields, runNode, coreRepo, metaRepo, recRepo, llmClient)
	}

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    graphql.NewObject(graphql.ObjectConfig{Name: "Query", Fields: queryFields}),
		Mutation: graphql.NewObject(graphql.ObjectConfig{Name: "Mutation", Fields: mutationFields}),
	})
}

func addObjectQuery(
	queryFields graphql.Fields,
	mutationFields graphql.Fields,
	pageInfo *graphql.Object,
	o metadata.Object,
	fields []metadata.Field,
	coreRepo *core.Repo,
	metaRepo *metadata.Repo,
	recRepo *records.Repo,
	tableStore *records.TableStore,
) *graphql.Object {
	typeName := title(o.NameSingular)

	nodeFields := graphql.Fields{
		"id":        &graphql.Field{Type: graphql.ID},
		"createdAt": &graphql.Field{Type: graphql.String},
		"updatedAt": &graphql.Field{Type: graphql.String},
	}
	for _, f := range fields {
		nodeFields[f.Name] = &graphql.Field{Type: gqlType(f.Type)}
	}
	node := graphql.NewObject(graphql.ObjectConfig{Name: typeName, Fields: nodeFields})

	edge := graphql.NewObject(graphql.ObjectConfig{
		Name: typeName + "Edge",
		Fields: graphql.Fields{
			"node":   &graphql.Field{Type: node},
			"cursor": &graphql.Field{Type: graphql.String},
		},
	})
	conn := graphql.NewObject(graphql.ObjectConfig{
		Name: typeName + "Connection",
		Fields: graphql.Fields{
			"edges":      &graphql.Field{Type: graphql.NewList(edge)},
			"pageInfo":   &graphql.Field{Type: pageInfo},
			"totalCount": &graphql.Field{Type: graphql.Int},
		},
	})

	// Filter input: and/or/not + per-field comparisons (nestjs-query style).
	var filterInput *graphql.InputObject
	filterInput = graphql.NewInputObject(graphql.InputObjectConfig{
		Name: typeName + "FilterInput",
		Fields: graphql.InputObjectConfigFieldMapThunk(func() graphql.InputObjectConfigFieldMap {
			m := graphql.InputObjectConfigFieldMap{
				"and":       &graphql.InputObjectFieldConfig{Type: graphql.NewList(filterInput)},
				"or":        &graphql.InputObjectFieldConfig{Type: graphql.NewList(filterInput)},
				"not":       &graphql.InputObjectFieldConfig{Type: filterInput},
				"id":        &graphql.InputObjectFieldConfig{Type: uuidFilterCmp},
				"createdAt": &graphql.InputObjectFieldConfig{Type: dateFilterCmp},
				"updatedAt": &graphql.InputObjectFieldConfig{Type: dateFilterCmp},
				"deletedAt": &graphql.InputObjectFieldConfig{Type: dateFilterCmp},
			}
			for _, f := range fields {
				m[f.Name] = &graphql.InputObjectFieldConfig{Type: filterCmpFor(f.Type)}
			}
			return m
		}),
	})
	orderByInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   typeName + "OrderByInput",
		Fields: graphql.InputObjectConfigFieldMap{"id": &graphql.InputObjectFieldConfig{Type: graphql.String}},
	})

	plural := o.NamePlural
	queryFields[plural] = &graphql.Field{
		Type: conn,
		Args: graphql.FieldConfigArgument{
			"filter":  &graphql.ArgumentConfig{Type: filterInput},
			"orderBy": &graphql.ArgumentConfig{Type: graphql.NewList(orderByInput)},
			"first":   &graphql.ArgumentConfig{Type: graphql.Int},
			"after":   &graphql.ArgumentConfig{Type: graphql.String},
			"before":  &graphql.ArgumentConfig{Type: graphql.String},
			"last":    &graphql.ArgumentConfig{Type: graphql.Int},
			"offset":  &graphql.ArgumentConfig{Type: graphql.Int},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			wsID, err := coreRepo.DefaultWorkspaceID(p.Context, userIDFromCtx(p.Context))
			if err != nil {
				return nil, err
			}
			obj, err := metaRepo.GetObjectByPlural(p.Context, wsID, plural)
			if err != nil {
				return emptyConnection(), nil
			}
			objFields, err := metaRepo.ListFields(p.Context, obj.ID)
			if err != nil {
				return nil, err
			}

			idEq := ""
			fieldFilters := map[string]string{}
			if fm, ok := p.Args["filter"].(map[string]any); ok {
				parseFilter(fm, &idEq, fieldFilters)
			}

			// findOne: filter on a single id → fetch directly.
			if idEq != "" {
				var rec records.Record
				if obj.Storage == "table" {
					rec, err = tableStore.Get(p.Context, obj, wsID, idEq)
				} else {
					rec, err = recRepo.Get(p.Context, obj.ID, wsID, idEq)
				}
				if err != nil {
					return emptyConnection(), nil
				}
				return buildConnection([]records.Record{rec}, 1, 0), nil
			}

			opt := records.ListOptions{
				Limit:   argInt(p.Args, "first", 50),
				Offset:  argInt(p.Args, "offset", 0),
				Filters: fieldFilters,
			}
			var recs []records.Record
			var total int
			if obj.Storage == "table" {
				names := map[string]struct{}{}
				for _, f := range objFields {
					names[f.Name] = struct{}{}
				}
				recs, total, err = tableStore.List(p.Context, obj, names, wsID, opt)
			} else {
				recs, total, err = recRepo.List(p.Context, obj.ID, wsID, opt)
			}
			if err != nil {
				return nil, err
			}
			return buildConnection(recs, total, opt.Offset), nil
		},
	}

	// Singular findOne field: {singular}(filter) → one record or null.
	queryFields[o.NameSingular] = &graphql.Field{
		Type: node,
		Args: graphql.FieldConfigArgument{"filter": &graphql.ArgumentConfig{Type: filterInput}},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			wsID, err := coreRepo.DefaultWorkspaceID(p.Context, userIDFromCtx(p.Context))
			if err != nil {
				return nil, err
			}
			obj, err := metaRepo.GetObjectByPlural(p.Context, wsID, plural)
			if err != nil {
				return nil, nil
			}
			idEq := ""
			ff := map[string]string{}
			if fm, ok := p.Args["filter"].(map[string]any); ok {
				parseFilter(fm, &idEq, ff)
			}
			var rec records.Record
			if idEq != "" {
				if obj.Storage == "table" {
					rec, err = tableStore.Get(p.Context, obj, wsID, idEq)
				} else {
					rec, err = recRepo.Get(p.Context, obj.ID, wsID, idEq)
				}
				if err != nil {
					return nil, nil
				}
				return recordNode(rec), nil
			}
			// otherwise first matching record
			var recs []records.Record
			opt := records.ListOptions{Limit: 1, Filters: ff}
			if obj.Storage == "table" {
				of, _ := metaRepo.ListFields(p.Context, obj.ID)
				names := map[string]struct{}{}
				for _, f := range of {
					names[f.Name] = struct{}{}
				}
				recs, _, err = tableStore.List(p.Context, obj, names, wsID, opt)
			} else {
				recs, _, err = recRepo.List(p.Context, obj.ID, wsID, opt)
			}
			if err != nil || len(recs) == 0 {
				return nil, nil
			}
			return recordNode(recs[0]), nil
		},
	}

	// ── Mutations: create / update / delete ─────────────────────────────────
	createInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   typeName + "CreateInput",
		Fields: inputFieldsFor(fields),
	})
	updateInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   typeName + "UpdateInput",
		Fields: inputFieldsFor(fields),
	})

	resolveObj := func(ctx context.Context) (string, metadata.Object, error) {
		wsID, err := coreRepo.DefaultWorkspaceID(ctx, userIDFromCtx(ctx))
		if err != nil {
			return "", metadata.Object{}, err
		}
		obj, err := metaRepo.GetObjectByPlural(ctx, wsID, plural)
		return wsID, obj, err
	}

	mutationFields["create"+typeName] = &graphql.Field{
		Type: node,
		Args: graphql.FieldConfigArgument{"data": &graphql.ArgumentConfig{Type: graphql.NewNonNull(createInput)}},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			wsID, obj, err := resolveObj(p.Context)
			if err != nil {
				return nil, err
			}
			data, id := splitRecordInput(p.Args["data"])
			var rec records.Record
			if obj.Storage == "table" {
				rec, err = tableStore.CreateWithID(p.Context, obj, wsID, id, data)
			} else {
				rec, err = recRepo.CreateWithID(p.Context, obj.ID, wsID, id, data)
			}
			if err != nil {
				return nil, err
			}
			return recordNode(rec), nil
		},
	}

	mutationFields["update"+typeName] = &graphql.Field{
		Type: node,
		Args: graphql.FieldConfigArgument{
			"id":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			"data": &graphql.ArgumentConfig{Type: graphql.NewNonNull(updateInput)},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			wsID, obj, err := resolveObj(p.Context)
			if err != nil {
				return nil, err
			}
			id, _ := p.Args["id"].(string)
			patch, _ := splitRecordInput(p.Args["data"])
			var rec records.Record
			if obj.Storage == "table" {
				rec, err = tableStore.Update(p.Context, obj, wsID, id, patch)
			} else {
				rec, err = recRepo.Update(p.Context, obj.ID, wsID, id, patch)
			}
			if err != nil {
				return nil, err
			}
			return recordNode(rec), nil
		},
	}

	mutationFields["delete"+typeName] = &graphql.Field{
		Type: node,
		Args: graphql.FieldConfigArgument{"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)}},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			wsID, obj, err := resolveObj(p.Context)
			if err != nil {
				return nil, err
			}
			id, _ := p.Args["id"].(string)
			var rec records.Record
			if obj.Storage == "table" {
				rec, _ = tableStore.Get(p.Context, obj, wsID, id)
				err = tableStore.SoftDelete(p.Context, obj, wsID, id)
			} else {
				rec, _ = recRepo.Get(p.Context, obj.ID, wsID, id)
				err = recRepo.SoftDelete(p.Context, obj.ID, wsID, id)
			}
			if err != nil {
				return nil, err
			}
			return recordNode(rec), nil
		},
	}

	return node
}

// inputFieldsFor builds the input fields for create/update (all optional + id).
func inputFieldsFor(fields []metadata.Field) graphql.InputObjectConfigFieldMap {
	m := graphql.InputObjectConfigFieldMap{"id": &graphql.InputObjectFieldConfig{Type: graphql.String}}
	for _, f := range fields {
		m[f.Name] = &graphql.InputObjectFieldConfig{Type: gqlType(f.Type)}
	}
	return m
}

// splitRecordInput pops id/system keys from a record input, returning data + id.
func splitRecordInput(arg any) (map[string]any, string) {
	in, _ := arg.(map[string]any)
	data := map[string]any{}
	id := ""
	for k, v := range in {
		switch k {
		case "id":
			id, _ = v.(string)
		case "createdAt", "updatedAt", "deletedAt":
			// system columns, ignore
		default:
			data[k] = v
		}
	}
	return data, id
}

// recordNode builds a GraphQL node map from a record.
func recordNode(r records.Record) map[string]any {
	node := map[string]any{
		"id":        r.ID,
		"createdAt": r.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": r.UpdatedAt.UTC().Format(time.RFC3339),
	}
	for k, v := range r.Data {
		node[k] = v
	}
	return node
}

func buildConnection(recs []records.Record, total, offset int) map[string]any {
	edges := make([]map[string]any, 0, len(recs))
	for _, r := range recs {
		node := map[string]any{
			"id":        r.ID,
			"createdAt": r.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt": r.UpdatedAt.UTC().Format(time.RFC3339),
		}
		for k, v := range r.Data {
			node[k] = v
		}
		edges = append(edges, map[string]any{"node": node, "cursor": r.ID})
	}
	start, end := "", ""
	if len(recs) > 0 {
		start = recs[0].ID
		end = recs[len(recs)-1].ID
	}
	return map[string]any{
		"totalCount": total,
		"edges":      edges,
		"pageInfo": map[string]any{
			"hasNextPage":     offset+len(recs) < total,
			"hasPreviousPage": offset > 0,
			"startCursor":     start,
			"endCursor":       end,
		},
	}
}

func emptyConnection() map[string]any {
	return map[string]any{
		"totalCount": 0,
		"edges":      []map[string]any{},
		"pageInfo":   map[string]any{"hasNextPage": false, "hasPreviousPage": false, "startCursor": "", "endCursor": ""},
	}
}

// parseFilter walks a nestjs-query filter, collecting an id equality (for
// findOne) and per-field equality filters. and/or are flattened; not is skipped.
func parseFilter(f map[string]any, idEq *string, ff map[string]string) {
	for k, v := range f {
		switch k {
		case "and", "or":
			if arr, ok := v.([]any); ok {
				for _, e := range arr {
					if m, ok := e.(map[string]any); ok {
						parseFilter(m, idEq, ff)
					}
				}
			}
		case "not":
			// negation not supported yet
		case "id":
			if cmp, ok := v.(map[string]any); ok {
				if eq, ok := cmp["eq"].(string); ok {
					*idEq = eq
				}
			}
		default:
			if cmp, ok := v.(map[string]any); ok {
				if eq, ok := cmp["eq"]; ok && eq != nil {
					ff[k] = fmt.Sprint(eq)
				}
			}
		}
	}
}

func gqlType(t metadata.FieldType) graphql.Output {
	switch t {
	case metadata.FieldNumber, metadata.FieldNumeric, metadata.FieldRating:
		return graphql.Float
	case metadata.FieldBoolean:
		return graphql.Boolean
	case metadata.FieldMultiSelect, metadata.FieldArray:
		return graphql.NewList(graphql.String)
	case "FULL_NAME":
		return fullNameOutput
	default:
		return graphql.String
	}
}

func argInt(args map[string]any, key string, def int) int {
	if v, ok := args[key].(int); ok && v > 0 {
		return v
	}
	return def
}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
