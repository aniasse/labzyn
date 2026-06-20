// Package records implements generic, metadata-driven CRUD over the JSONB
// record store (core.record).
package records

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a record does not exist.
var ErrNotFound = errors.New("not found")

// Record is a single stored row.
type Record struct {
	ID        string         `json:"id"`
	Data      map[string]any `json:"data"`
	Position  float64        `json:"position"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// Repo provides record persistence.
type Repo struct{ pool *pgxpool.Pool }

// NewRepo builds a records repository.
func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

// ListOptions controls listing.
type ListOptions struct {
	Limit       int
	Offset      int
	OrderBy     string            // "created_at" | "updated_at" | "position" | a data field
	OrderByCast string            // Postgres cast for a data-field sort (e.g. "numeric"); "" = text
	Desc        bool              // sort direction
	Filters     map[string]string // data field -> equality value
}

const recordCols = `id::text, data, position, created_at, updated_at`

func scanRecord(row pgx.Row) (Record, error) {
	var r Record
	err := row.Scan(&r.ID, &r.Data, &r.Position, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

// List returns records for an object plus the total count (ignoring paging).
func (r *Repo) List(ctx context.Context, objectID, workspaceID string, opt ListOptions) ([]Record, int, error) {
	where := []string{"object_id = $1", "workspace_id = $2", "deleted_at IS NULL"}
	args := []any{objectID, workspaceID}

	for field, value := range opt.Filters {
		args = append(args, value)
		where = append(where, fmt.Sprintf("data->>'%s' = $%d", field, len(args)))
	}
	whereSQL := strings.Join(where, " AND ")

	// Total count.
	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT count(*) FROM core.record WHERE `+whereSQL, args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Order by (identifiers are pre-validated by the caller).
	order := orderClause(opt.OrderBy, opt.OrderByCast, opt.Desc)

	limit := opt.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	args = append(args, limit, opt.Offset)
	q := fmt.Sprintf(
		`SELECT %s FROM core.record WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		recordCols, whereSQL, order, len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []Record{}
	for rows.Next() {
		rec, err := scanRecord(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, rec)
	}
	return out, total, rows.Err()
}

func orderClause(orderBy, cast string, desc bool) string {
	dir := "ASC"
	if desc {
		dir = "DESC"
	}
	switch orderBy {
	case "created_at", "updated_at", "position":
		return orderBy + " " + dir
	case "":
		return "created_at " + dir
	default:
		// data field; orderBy and cast are validated against the catalog by the
		// caller. Cast typed fields (e.g. numbers) so they sort by value, not text.
		if cast != "" {
			return fmt.Sprintf("(data->>'%s')::%s %s", orderBy, cast, dir)
		}
		return fmt.Sprintf("data->>'%s' %s", orderBy, dir)
	}
}

// Create inserts a record.
func (r *Repo) Create(ctx context.Context, objectID, workspaceID string, data map[string]any) (Record, error) {
	if data == nil {
		data = map[string]any{}
	}
	row := r.pool.QueryRow(ctx,
		`INSERT INTO core.record (object_id, workspace_id, data) VALUES ($1,$2,$3)
		 RETURNING `+recordCols, objectID, workspaceID, data)
	return scanRecord(row)
}

// Get returns a single record.
func (r *Repo) Get(ctx context.Context, objectID, workspaceID, id string) (Record, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+recordCols+` FROM core.record
		 WHERE id = $1 AND object_id = $2 AND workspace_id = $3 AND deleted_at IS NULL`,
		id, objectID, workspaceID)
	rec, err := scanRecord(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrNotFound
	}
	return rec, err
}

// Update merges patch into the record's data (top-level keys).
func (r *Repo) Update(ctx context.Context, objectID, workspaceID, id string, patch map[string]any) (Record, error) {
	if patch == nil {
		patch = map[string]any{}
	}
	row := r.pool.QueryRow(ctx,
		`UPDATE core.record SET data = data || $4, updated_at = now()
		 WHERE id = $1 AND object_id = $2 AND workspace_id = $3 AND deleted_at IS NULL
		 RETURNING `+recordCols, id, objectID, workspaceID, patch)
	rec, err := scanRecord(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrNotFound
	}
	return rec, err
}

// SoftDelete marks a record deleted.
func (r *Repo) SoftDelete(ctx context.Context, objectID, workspaceID, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE core.record SET deleted_at = now()
		 WHERE id = $1 AND object_id = $2 AND workspace_id = $3 AND deleted_at IS NULL`,
		id, objectID, workspaceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
