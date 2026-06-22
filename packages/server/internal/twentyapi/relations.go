package twentyapi

import (
	"github.com/aniasse/labzyn/server/internal/core"
	"github.com/aniasse/labzyn/server/internal/metadata"
	"github.com/aniasse/labzyn/server/internal/records"
	"github.com/graphql-go/graphql"
)

// addRelationField adds a relation field to a node's field map. MANY_TO_ONE
// resolves the single related record via the join column; ONE_TO_MANY resolves
// a connection of records pointing back through the join column.
func addRelationField(
	nf graphql.Fields,
	f metadata.Field,
	nodes map[string]*graphql.Object,
	conns map[string]*graphql.Object,
	filters map[string]*graphql.InputObject,
	coreRepo *core.Repo,
	metaRepo *metadata.Repo,
	recRepo *records.Repo,
	tableStore *records.TableStore,
) {
	rel := parseRelation(f.Options)
	if rel == nil {
		return
	}
	join, targetPlural := rel.JoinColumnName, rel.TargetPlural

	if rel.Type == "MANY_TO_ONE" {
		target := nodes[rel.Target]
		if target == nil {
			return
		}
		nf[f.Name] = &graphql.Field{
			Type: target,
			Resolve: func(p graphql.ResolveParams) (any, error) {
				src, _ := p.Source.(map[string]any)
				fk, _ := src[join].(string)
				if fk == "" {
					return nil, nil
				}
				wsID, obj, err := resolveTarget(p, coreRepo, metaRepo, targetPlural)
				if err != nil {
					return nil, nil
				}
				var rec records.Record
				if obj.Storage == "table" {
					rec, err = tableStore.Get(p.Context, obj, wsID, fk)
				} else {
					rec, err = recRepo.Get(p.Context, obj.ID, wsID, fk)
				}
				if err != nil {
					return nil, nil
				}
				return recordNode(rec), nil
			},
		}
		return
	}

	// ONE_TO_MANY
	targetConn := conns[rel.Target]
	if targetConn == nil {
		return
	}
	args := graphql.FieldConfigArgument{
		"first":  &graphql.ArgumentConfig{Type: graphql.Int},
		"after":  &graphql.ArgumentConfig{Type: graphql.String},
		"offset": &graphql.ArgumentConfig{Type: graphql.Int},
	}
	if fi := filters[rel.Target]; fi != nil {
		args["filter"] = &graphql.ArgumentConfig{Type: fi}
	}
	nf[f.Name] = &graphql.Field{
		Type: targetConn,
		Args: args,
		Resolve: func(p graphql.ResolveParams) (any, error) {
			src, _ := p.Source.(map[string]any)
			thisID, _ := src["id"].(string)
			if thisID == "" {
				return emptyConnection(), nil
			}
			wsID, obj, err := resolveTarget(p, coreRepo, metaRepo, targetPlural)
			if err != nil {
				return emptyConnection(), nil
			}
			opt := records.ListOptions{
				Limit:   argInt(p.Args, "first", 50),
				Offset:  argInt(p.Args, "offset", 0),
				Filters: map[string]string{join: thisID},
			}
			var recs []records.Record
			var total int
			if obj.Storage == "table" {
				of, _ := metaRepo.ListFields(p.Context, obj.ID)
				names := map[string]struct{}{}
				for _, ff := range of {
					names[ff.Name] = struct{}{}
				}
				recs, total, err = tableStore.List(p.Context, obj, names, wsID, opt)
			} else {
				recs, total, err = recRepo.List(p.Context, obj.ID, wsID, opt)
			}
			if err != nil {
				return emptyConnection(), nil
			}
			return buildConnection(recs, total, opt.Offset), nil
		},
	}
}

func resolveTarget(p graphql.ResolveParams, coreRepo *core.Repo, metaRepo *metadata.Repo, plural string) (string, metadata.Object, error) {
	wsID, err := coreRepo.DefaultWorkspaceID(p.Context, userIDFromCtx(p.Context))
	if err != nil {
		return "", metadata.Object{}, err
	}
	obj, err := metaRepo.GetObjectByPlural(p.Context, wsID, plural)
	return wsID, obj, err
}
