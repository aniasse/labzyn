package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aniasse/labzyn/server/internal/hpc"
	"github.com/aniasse/labzyn/server/internal/httpx"
	"github.com/aniasse/labzyn/server/internal/records"
)

// Minimal MCP (Model Context Protocol) server over JSON-RPC 2.0 / HTTP POST.
// Exposes the workspace's data model and the HPC gateway as agent tools.

const mcpProtocolVersion = "2025-06-18"

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type toolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

func obj(props map[string]any, required ...string) map[string]any {
	m := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		m["required"] = required
	}
	return m
}

func mcpToolDefs() []toolDef {
	str := map[string]any{"type": "string"}
	num := map[string]any{"type": "number"}
	return []toolDef{
		{"list_objects", "List the objects (tables) defined in the workspace.", obj(map[string]any{})},
		{"query_records", "List records of an object. Use the object's plural name.",
			obj(map[string]any{"object": str, "limit": num, "orderBy": str, "order": str}, "object")},
		{"create_record", "Create a record for an object (plural name). data is an object of field values.",
			obj(map[string]any{"object": str, "data": map[string]any{"type": "object"}}, "object", "data")},
		{"submit_hpc_job", "Submit a job to the HPC gateway (Slurm/TAOUEY). Returns a job id.",
			obj(map[string]any{"name": str, "script": str, "partition": str, "cpus": num, "gpus": num, "memMb": num, "timeMin": num}, "name", "script")},
		{"get_hpc_job_status", "Get the status of a submitted HPC job.",
			obj(map[string]any{"jobId": str}, "jobId")},
		{"fetch_hpc_job_results", "Fetch stdout/stderr/files of an HPC job.",
			obj(map[string]any{"jobId": str}, "jobId")},
	}
}

func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	var req rpcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.JSON(w, http.StatusOK, rpcResponse{JSONRPC: "2.0", Error: &rpcError{-32700, "parse error"}})
		return
	}

	// Notifications (no id) get no response body.
	if len(req.ID) == 0 && req.Method != "tools/call" {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	resp := rpcResponse{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		resp.Result = map[string]any{
			"protocolVersion": mcpProtocolVersion,
			"serverInfo":      map[string]any{"name": "labzyn-mcp", "version": "0.1.0"},
			"capabilities":    map[string]any{"tools": map[string]any{}},
		}
	case "ping":
		resp.Result = map[string]any{}
	case "tools/list":
		resp.Result = map[string]any{"tools": mcpToolDefs()}
	case "tools/call":
		resp.Result, resp.Error = s.mcpCallTool(r.Context(), workspaceID(r), req.Params)
	default:
		resp.Error = &rpcError{-32601, "method not found: " + req.Method}
	}
	httpx.JSON(w, http.StatusOK, resp)
}

func (s *Server) mcpCallTool(ctx context.Context, wsID string, params json.RawMessage) (any, *rpcError) {
	var p struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcError{-32602, "invalid params"}
	}

	result, err := s.execTool(ctx, wsID, p.Name, p.Arguments)
	if err != nil {
		// Tool errors are reported in-band (isError), per MCP convention.
		return map[string]any{
			"content": []map[string]any{{"type": "text", "text": err.Error()}},
			"isError": true,
		}, nil
	}
	text, _ := json.Marshal(result)
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": string(text)}},
	}, nil
}

func (s *Server) execTool(ctx context.Context, wsID, name string, args json.RawMessage) (any, error) {
	switch name {
	case "list_objects":
		return s.meta.ListObjects(ctx, wsID)

	case "query_records":
		var a struct {
			Object  string `json:"object"`
			Limit   int    `json:"limit"`
			OrderBy string `json:"orderBy"`
			Order   string `json:"order"`
		}
		_ = json.Unmarshal(args, &a)
		o, err := s.meta.GetObjectByPlural(ctx, wsID, a.Object)
		if err != nil {
			return nil, err
		}
		opt := records.ListOptions{Limit: a.Limit, Desc: a.Order == "desc"}
		// Validate orderBy against the catalog (the value reaches raw SQL) and
		// apply a typed cast so numbers/dates sort by value, not text.
		if a.OrderBy != "" {
			switch a.OrderBy {
			case "created_at", "updated_at", "position":
				opt.OrderBy = a.OrderBy
			default:
				fields, err := s.meta.ListFields(ctx, o.ID)
				if err != nil {
					return nil, err
				}
				for _, f := range fields {
					if f.Name == a.OrderBy {
						opt.OrderBy = a.OrderBy
						opt.OrderByCast = f.Type.SortCast()
						break
					}
				}
				if opt.OrderBy == "" {
					return nil, errUnknownField(a.OrderBy)
				}
			}
		}
		list, total, err := s.rec.List(ctx, o.ID, wsID, opt)
		if err != nil {
			return nil, err
		}
		return map[string]any{"records": list, "total": total}, nil

	case "create_record":
		var a struct {
			Object string         `json:"object"`
			Data   map[string]any `json:"data"`
		}
		_ = json.Unmarshal(args, &a)
		o, err := s.meta.GetObjectByPlural(ctx, wsID, a.Object)
		if err != nil {
			return nil, err
		}
		fields, err := s.meta.ListFields(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		if err := records.Validate(fields, a.Data, true); err != nil {
			return nil, err
		}
		return s.rec.Create(ctx, o.ID, wsID, a.Data)

	case "submit_hpc_job":
		var a struct {
			Name      string `json:"name"`
			Script    string `json:"script"`
			Partition string `json:"partition"`
			CPUs      int    `json:"cpus"`
			GPUs      int    `json:"gpus"`
			MemMB     int    `json:"memMb"`
			TimeMin   int    `json:"timeMin"`
		}
		_ = json.Unmarshal(args, &a)
		return s.hpc.Submit(ctx, hpc.SubmitRequest{
			Name:   a.Name,
			Script: a.Script,
			Resources: hpc.Resources{
				Partition: a.Partition, CPUs: a.CPUs, GPUs: a.GPUs, MemMB: a.MemMB, TimeMin: a.TimeMin,
			},
		})

	case "get_hpc_job_status":
		var a struct {
			JobID string `json:"jobId"`
		}
		_ = json.Unmarshal(args, &a)
		return s.hpc.Status(ctx, a.JobID)

	case "fetch_hpc_job_results":
		var a struct {
			JobID string `json:"jobId"`
		}
		_ = json.Unmarshal(args, &a)
		return s.hpc.FetchResults(ctx, a.JobID)

	default:
		return nil, errUnknownTool(name)
	}
}

type errUnknownTool string

func (e errUnknownTool) Error() string { return "unknown tool: " + string(e) }

type errUnknownField string

func (e errUnknownField) Error() string { return "unknown orderBy field: " + string(e) }
