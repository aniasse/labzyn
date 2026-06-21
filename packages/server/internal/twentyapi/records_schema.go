package twentyapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/aniasse/labzyn/server/internal/core"
	"github.com/aniasse/labzyn/server/internal/metadata"
	"github.com/aniasse/labzyn/server/internal/records"
	"github.com/graphql-go/graphql"
	gqlhandler "github.com/graphql-go/handler"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BuildCoreHandler builds the dynamic records GraphQL handler for /graphql.
func BuildCoreHandler(pool *pgxpool.Pool) (http.Handler, error) {
	schema, err := buildCoreSchema(pool)
	if err != nil {
		return nil, err
	}
	return gqlhandler.New(&gqlhandler.Config{Schema: &schema, Pretty: true}), nil
}

// buildCoreSchema builds the dynamic per-object records schema (/graphql).
// One query field + connection type per object, resolved from the records repo.
func buildCoreSchema(pool *pgxpool.Pool) (graphql.Schema, error) {
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

	for _, o := range objs {
		fields, err := metaRepo.ListFields(context.Background(), o.ID)
		if err != nil {
			return graphql.Schema{}, err
		}
		addObjectQuery(queryFields, pageInfo, o, fields, coreRepo, metaRepo, recRepo, tableStore)
	}

	return graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{Name: "Query", Fields: queryFields}),
	})
}

func addObjectQuery(
	queryFields graphql.Fields,
	pageInfo *graphql.Object,
	o metadata.Object,
	fields []metadata.Field,
	coreRepo *core.Repo,
	metaRepo *metadata.Repo,
	recRepo *records.Repo,
	tableStore *records.TableStore,
) {
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

	// Permissive filter/orderBy inputs (contents ignored for now; accept {}/[]).
	filterInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   typeName + "FilterInput",
		Fields: graphql.InputObjectConfigFieldMap{"id": &graphql.InputObjectFieldConfig{Type: graphql.String}},
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

			opt := records.ListOptions{Limit: argInt(p.Args, "first", 50), Offset: argInt(p.Args, "offset", 0)}
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

func gqlType(t metadata.FieldType) graphql.Output {
	switch t {
	case metadata.FieldNumber, metadata.FieldNumeric, metadata.FieldRating:
		return graphql.Float
	case metadata.FieldBoolean:
		return graphql.Boolean
	case metadata.FieldMultiSelect, metadata.FieldArray:
		return graphql.NewList(graphql.String)
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
