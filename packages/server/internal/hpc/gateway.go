// Package hpc is a client for the Labzyn HPC gateway (:8088 → TAOUEY/Slurm).
package hpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Resources describes the compute a job asks for.
type Resources struct {
	Partition string `json:"partition,omitempty"`
	CPUs      int    `json:"cpus,omitempty"`
	GPUs      int    `json:"gpus,omitempty"`
	MemMB     int    `json:"memMb,omitempty"`
	TimeMin   int    `json:"timeMin,omitempty"`
}

// SubmitRequest is the payload to submit a job.
type SubmitRequest struct {
	Name      string    `json:"name"`
	Script    string    `json:"script"`
	Resources Resources `json:"resources"`
}

// Job is the gateway's view of a submitted job.
type Job struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Resources   Resources `json:"resources"`
	SubmittedAt string    `json:"submittedAt"`
}

// Results holds a job's outputs.
type Results struct {
	JobID  string   `json:"jobId"`
	Status string   `json:"status"`
	Stdout string   `json:"stdout"`
	Stderr string   `json:"stderr"`
	Files  []string `json:"files"`
}

// Client talks to the gateway over HTTP.
type Client struct {
	baseURL string
	http    *http.Client
}

// New builds a gateway client.
func New(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 15 * time.Second}}
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("gateway unreachable: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		msg, _ := io.ReadAll(res.Body)
		return fmt.Errorf("gateway error %d: %s", res.StatusCode, string(msg))
	}
	if out != nil {
		return json.NewDecoder(res.Body).Decode(out)
	}
	return nil
}

// Submit sends a job to the gateway.
func (c *Client) Submit(ctx context.Context, req SubmitRequest) (Job, error) {
	var job Job
	err := c.do(ctx, http.MethodPost, "/jobs", req, &job)
	return job, err
}

// Status returns a job's current state.
func (c *Client) Status(ctx context.Context, id string) (Job, error) {
	var job Job
	err := c.do(ctx, http.MethodGet, "/jobs/"+id, nil, &job)
	return job, err
}

// FetchResults returns a job's outputs.
func (c *Client) FetchResults(ctx context.Context, id string) (Results, error) {
	var r Results
	err := c.do(ctx, http.MethodGet, "/jobs/"+id+"/results", nil, &r)
	return r, err
}
