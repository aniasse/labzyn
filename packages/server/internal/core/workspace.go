// Package core holds tenant-level primitives (workspaces, users) shared by the
// metadata and records layers.
package core

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Workspace is a tenant.
type Workspace struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"displayName"`
	Subdomain   string    `json:"subdomain"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Repo provides workspace persistence.
type Repo struct{ pool *pgxpool.Pool }

// NewRepo builds a core repository.
func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

const workspaceCols = `id::text, display_name, subdomain, created_at, updated_at`

func scanWorkspace(row pgx.Row) (Workspace, error) {
	var w Workspace
	err := row.Scan(&w.ID, &w.DisplayName, &w.Subdomain, &w.CreatedAt, &w.UpdatedAt)
	return w, err
}

// CreateWorkspace inserts a workspace.
func (r *Repo) CreateWorkspace(ctx context.Context, displayName, subdomain string) (Workspace, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO core.workspace (display_name, subdomain) VALUES ($1, $2)
		 RETURNING `+workspaceCols, displayName, subdomain)
	return scanWorkspace(row)
}

// ListWorkspaces returns all non-deleted workspaces, oldest first.
func (r *Repo) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+workspaceCols+` FROM core.workspace
		 WHERE deleted_at IS NULL ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Workspace{}
	for rows.Next() {
		w, err := scanWorkspace(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// GetWorkspace returns a workspace by id.
func (r *Repo) GetWorkspace(ctx context.Context, id string) (Workspace, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+workspaceCols+` FROM core.workspace
		 WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanWorkspace(row)
}

// OldestWorkspaceID returns the id of the oldest workspace, or "" if none exist.
// Used as the default tenant in development before auth is wired.
func (r *Repo) OldestWorkspaceID(ctx context.Context) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`SELECT id::text FROM core.workspace
		 WHERE deleted_at IS NULL ORDER BY created_at ASC LIMIT 1`).Scan(&id)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return id, err
}
