package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aniasse/labzyn/server/internal/httpx"
	"github.com/aniasse/labzyn/server/internal/llm"
)

const maxAgentTurns = 6

const agentSystemPrompt = `You are Labzyn, an agentic copilot for scientific research on HPC.
You help researchers manage their project data and run compute jobs.

You have tools to:
- inspect and query the workspace data model (objects and records),
- create records,
- submit jobs to the HPC gateway (Slurm/TAOUEY) and read their status/results.

Rules:
- Use tools to read real data; never invent records, statuses, or results.
- When asked to run something, generate a correct shell/sbatch script and submit it.
- Be concise. Report what you actually did (ids, statuses).`

// agentStep records a single tool invocation for the UI trace.
type agentStep struct {
	Tool      string          `json:"tool"`
	Arguments json.RawMessage `json:"arguments"`
	Result    string          `json:"result"`
}

// handleAgentStatus tells the UI whether the agent is configured.
func (s *Server) handleAgentStatus(w http.ResponseWriter, _ *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]any{"enabled": s.llm.Enabled()})
}

func (s *Server) handleAgentChat(w http.ResponseWriter, r *http.Request) {
	if !s.llm.Enabled() {
		httpx.Error(w, http.StatusServiceUnavailable,
			"LLM not configured — set LABZYN_LLM_API_KEY (and base URL/model) to enable the agent")
		return
	}

	var in struct {
		Messages []llm.Message `json:"messages"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	wsID := workspaceID(r)
	tools := agentTools()

	msgs := append([]llm.Message{{Role: "system", Content: agentSystemPrompt}}, in.Messages...)
	steps := []agentStep{}

	for turn := 0; turn < maxAgentTurns; turn++ {
		asst, err := s.llm.Chat(r.Context(), msgs, tools)
		if err != nil {
			httpx.Error(w, http.StatusBadGateway, err.Error())
			return
		}
		msgs = append(msgs, asst)

		if len(asst.ToolCalls) == 0 {
			httpx.JSON(w, http.StatusOK, map[string]any{"reply": asst.Content, "steps": steps})
			return
		}

		for _, tc := range asst.ToolCalls {
			content := s.runAgentTool(r.Context(), wsID, tc.Function.Name, tc.Function.Arguments)
			steps = append(steps, agentStep{
				Tool:      tc.Function.Name,
				Arguments: json.RawMessage(tc.Function.Arguments),
				Result:    content,
			})
			msgs = append(msgs, llm.Message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Name:       tc.Function.Name,
				Content:    content,
			})
		}
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"reply": "I reached the step limit before finishing. Please refine the request.",
		"steps": steps,
	})
}

// runAgentTool executes a tool and returns a JSON string result (errors in-band).
func (s *Server) runAgentTool(ctx context.Context, wsID, name, argsJSON string) string {
	args := json.RawMessage(argsJSON)
	if argsJSON == "" {
		args = json.RawMessage(`{}`)
	}
	result, err := s.execTool(ctx, wsID, name, args)
	if err != nil {
		b, _ := json.Marshal(map[string]string{"error": err.Error()})
		return string(b)
	}
	b, _ := json.Marshal(result)
	return string(b)
}

// agentTools exposes the MCP tools to the LLM in OpenAI function format.
func agentTools() []llm.Tool {
	defs := mcpToolDefs()
	tools := make([]llm.Tool, 0, len(defs))
	for _, d := range defs {
		tools = append(tools, llm.Tool{
			Type: "function",
			Function: llm.Function{
				Name:        d.Name,
				Description: d.Description,
				Parameters:  d.InputSchema,
			},
		})
	}
	return tools
}
