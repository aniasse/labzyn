package twentyapi

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aniasse/labzyn/server/internal/core"
	"github.com/aniasse/labzyn/server/internal/hpc"
	"github.com/aniasse/labzyn/server/internal/metadata"
	"github.com/aniasse/labzyn/server/internal/records"
	"github.com/graphql-go/graphql"
)

// gatewayURL is where the Labzyn HPC gateway (→ TAOUEY/Slurm) listens.
func gatewayURL() string {
	if v := os.Getenv("LABZYN_HPC_GATEWAY_URL"); v != "" {
		return v
	}
	return "http://localhost:8088"
}

// addRunHpcMutations wires the Labzyn-specific bridge between a Run record and
// the HPC gateway: submitRun sends the script, refreshRun polls and harvests
// results. These are what make Labzyn more than a CRM running on Twenty.
func addRunHpcMutations(
	mutationFields graphql.Fields,
	runNode *graphql.Object,
	coreRepo *core.Repo,
	metaRepo *metadata.Repo,
	recRepo *records.Repo,
) {
	gw := hpc.New(gatewayURL())

	runObject := func(ctx context.Context) (string, metadata.Object, error) {
		wsID, err := coreRepo.DefaultWorkspaceID(ctx, userIDFromCtx(ctx))
		if err != nil {
			return "", metadata.Object{}, err
		}
		obj, err := metaRepo.GetObjectByPlural(ctx, wsID, "runs")
		return wsID, obj, err
	}

	// submitRun(id): submit the Run's script to the gateway, store the job id.
	mutationFields["submitRun"] = &graphql.Field{
		Type: runNode,
		Args: graphql.FieldConfigArgument{"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)}},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			wsID, obj, err := runObject(p.Context)
			if err != nil {
				return nil, err
			}
			id, _ := p.Args["id"].(string)
			rec, err := recRepo.Get(p.Context, obj.ID, wsID, id)
			if err != nil {
				return nil, err
			}
			script := asString(rec.Data["script"])
			if script == "" {
				return nil, fmt.Errorf("run has no script to submit")
			}
			job, err := gw.Submit(p.Context, hpc.SubmitRequest{
				Name:      asString(rec.Data["name"]),
				Script:    script,
				Resources: hpc.Resources{Partition: asString(rec.Data["partition"])},
			})
			if err != nil {
				return nil, err
			}
			rec, err = recRepo.Update(p.Context, obj.ID, wsID, id, map[string]any{
				"hpcJobId": job.ID,
				"status":   mapJobStatus(job.Status),
			})
			if err != nil {
				return nil, err
			}
			return recordNode(rec), nil
		},
	}

	// refreshRun(id): poll the gateway; on completion store status + a Result.
	mutationFields["refreshRun"] = &graphql.Field{
		Type: runNode,
		Args: graphql.FieldConfigArgument{"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)}},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			wsID, obj, err := runObject(p.Context)
			if err != nil {
				return nil, err
			}
			id, _ := p.Args["id"].(string)
			rec, err := recRepo.Get(p.Context, obj.ID, wsID, id)
			if err != nil {
				return nil, err
			}
			jobID := asString(rec.Data["hpcJobId"])
			if jobID == "" {
				return nil, fmt.Errorf("run has not been submitted")
			}
			job, err := gw.Status(p.Context, jobID)
			if err != nil {
				return nil, err
			}
			status := mapJobStatus(job.Status)

			if status == "COMPLETED" {
				if res, err := gw.FetchResults(p.Context, jobID); err == nil {
					if rObj, err := metaRepo.GetObjectByPlural(p.Context, wsID, "results"); err == nil {
						fileName, fileSize := "", 0
						if len(res.Files) > 0 {
							fileName = res.Files[0].Name
							fileSize = int(res.Files[0].Size)
						}
						_, _ = recRepo.Create(p.Context, rObj.ID, wsID, map[string]any{
							"name":          asString(rec.Data["name"]) + " — output",
							"stdout":        res.Stdout,
							"fileName":      fileName,
							"fileSizeBytes": fileSize,
						})
					}
				}
			}

			rec, err = recRepo.Update(p.Context, obj.ID, wsID, id, map[string]any{"status": status})
			if err != nil {
				return nil, err
			}
			return recordNode(rec), nil
		},
	}
}

func asString(v any) string { s, _ := v.(string); return s }

// mapJobStatus normalises gateway/Slurm states to the Run status options.
func mapJobStatus(s string) string {
	switch strings.ToUpper(s) {
	case "PENDING", "QUEUED", "PD":
		return "PENDING"
	case "RUNNING", "R":
		return "RUNNING"
	case "COMPLETED", "SUCCESS", "DONE", "CD":
		return "COMPLETED"
	case "FAILED", "ERROR", "CANCELLED", "F":
		return "FAILED"
	default:
		return "RUNNING"
	}
}
