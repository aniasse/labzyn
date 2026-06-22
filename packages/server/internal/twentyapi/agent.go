package twentyapi

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/aniasse/labzyn/server/internal/llm"
)

// runPlan is the structured job an agent (or the fallback) derives from a prompt.
type runPlan struct {
	Name      string `json:"name"`
	Partition string `json:"partition"`
	CPUs      int    `json:"cpus"`
	GPUs      int    `json:"gpus"`
	TimeMin   int    `json:"timeMin"`
	Script    string `json:"script"`
}

const agentSystemPrompt = `You are Labzyn's HPC assistant for the TAOUEY cluster (Slurm).
Given a researcher's request, produce ONE JSON object describing a batch job:
{"name":"<short-kebab-slug>","partition":"cpu" or "gpu","cpus":<int>,"gpus":<int>,"timeMin":<int minutes>,"script":"<complete bash sbatch script>"}
The script MUST start with #!/bin/bash, include #SBATCH directives matching the resources, and the commands to fulfil the request.
Output ONLY the JSON object — no prose, no markdown fences.`

// generateRunPlan turns a natural-language request into a runnable Slurm job.
// It uses the configured LLM when available, otherwise a deterministic fallback
// so the agent is demonstrable without an API key.
func generateRunPlan(ctx context.Context, c *llm.Client, prompt string) runPlan {
	if c != nil && c.Enabled() {
		if p, ok := planFromLLM(ctx, c, prompt); ok {
			return p
		}
	}
	return fallbackPlan(prompt)
}

func planFromLLM(ctx context.Context, c *llm.Client, prompt string) (runPlan, bool) {
	msg, err := c.Chat(ctx, []llm.Message{
		{Role: "system", Content: agentSystemPrompt},
		{Role: "user", Content: prompt},
	}, nil)
	if err != nil {
		return runPlan{}, false
	}
	var p runPlan
	if err := json.Unmarshal([]byte(extractJSON(msg.Content)), &p); err != nil || strings.TrimSpace(p.Script) == "" {
		return runPlan{}, false
	}
	if p.Name == "" {
		p.Name = slugify(prompt)
	}
	if p.Partition != "gpu" {
		p.Partition = "cpu"
	}
	return p, true
}

// fallbackPlan builds a reasonable job from the prompt without an LLM: it guesses
// the partition from keywords and scaffolds an sbatch script around the request.
func fallbackPlan(prompt string) runPlan {
	gpu := mentionsAny(prompt, "gpu", "cuda", "torch", "tensorflow", "jax", "burn", "neural", "deep learning")
	partition, gpus := "cpu", 0
	if gpu {
		partition, gpus = "gpu", 1
	}
	name := slugify(prompt)
	if name == "" {
		name = "labzyn-run"
	}
	var b strings.Builder
	b.WriteString("#!/bin/bash\n")
	b.WriteString("#SBATCH --job-name=" + name + "\n")
	b.WriteString("#SBATCH --partition=" + partition + "\n")
	b.WriteString("#SBATCH --time=00:30:00\n")
	b.WriteString("#SBATCH --cpus-per-task=4\n")
	if gpus > 0 {
		b.WriteString("#SBATCH --gres=gpu:1\n")
	}
	b.WriteString("\n# Request: " + strings.TrimSpace(prompt) + "\n")
	b.WriteString("echo \"[labzyn] running on $(hostname) — partition " + partition + "\"\n")
	b.WriteString("# TODO: replace with the actual command for this job\n")
	return runPlan{Name: name, Partition: partition, CPUs: 4, GPUs: gpus, TimeMin: 30, Script: b.String()}
}

var (
	fenceRe = regexp.MustCompile("(?s)```[a-zA-Z]*\\n?(.*?)```")
	slugRe  = regexp.MustCompile(`[^a-z0-9]+`)
)

// extractJSON pulls a JSON object out of an LLM reply (handles ``` fences).
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if m := fenceRe.FindStringSubmatch(s); m != nil {
		s = m[1]
	}
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j >= i {
			return s[i : j+1]
		}
	}
	return s
}

func slugify(s string) string {
	words := strings.Fields(strings.ToLower(s))
	if len(words) > 5 {
		words = words[:5]
	}
	return strings.Trim(slugRe.ReplaceAllString(strings.Join(words, "-"), "-"), "-")
}

func mentionsAny(s string, terms ...string) bool {
	low := strings.ToLower(s)
	for _, t := range terms {
		if strings.Contains(low, t) {
			return true
		}
	}
	return false
}
