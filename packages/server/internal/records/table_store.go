package records

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aniasse/labzyn/server/internal/ddl"
	"github.com/aniasse/labzyn/server/internal/metadata"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TableStore is the high-performance record backend: it reads/writes a real
// per-workspace Postgres table with typed columns (storage = "table").
type TableStore struct{ pool *pgxpool.Pool }

// NewTableStore builds a table-backed store.
func NewTableStore(pool *pgxpool.Pool) *TableStore { return &TableStore{pool: pool} }

// systemCols are the columns that are not user data fields.
var systemCols = map[string]struct{}{
	"id": {}, "created_at": {}, "updated_at": {}, "deleted_at": {},
}

// selectExpr returns "id::text, created_at, updated_at, to_jsonb(t.*) AS row".
const selectExpr = `"id"::text, "created_at", "updated_at", to_jsonb(t.*) AS row`

func scanTableRecord(row pgx.Row) (Record, error) {
	var r Record
	var raw map[string]any
	if err := row.Scan(&r.ID, &r.CreatedAt, &r.UpdatedAt, &raw); err != nil {
		return Record{}, err
	}
	for k := range systemCols {
		delete(raw, k)
	}
	r.Data = raw
	return r, nil
}

// List returns records from the object's real table.
func (s *TableStore) List(ctx context.Context, obj metadata.Object, fieldNames map[string]struct{}, workspaceID string, opt ListOptions) ([]Record, int, error) {
	rel, err := ddl.Relation(workspaceID, obj.NameSingular)
	if err != nil {
		return nil, 0, err
	}

	where := []string{`"deleted_at" IS NULL`}
	args := []any{}
	for field, value := range opt.Filters {
		if _, ok := fieldNames[field]; !ok {
			continue
		}
		col, err := ddl.QuoteIdent(field)
		if err != nil {
			return nil, 0, err
		}
		args = append(args, value)
		where = append(where, fmt.Sprintf("%s::text = $%d", col, len(args)))
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	if err := s.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT count(*) FROM %s t WHERE %s`, rel, whereSQL), args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	order, err := tableOrderClause(opt, fieldNames)
	if err != nil {
		return nil, 0, err
	}
	limit := opt.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	args = append(args, limit, opt.Offset)
	q := fmt.Sprintf(`SELECT %s FROM %s t WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		selectExpr, rel, whereSQL, order, len(args)-1, len(args))

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []Record{}
	for rows.Next() {
		rec, err := scanTableRecord(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, rec)
	}
	return out, total, rows.Err()
}

func tableOrderClause(opt ListOptions, fieldNames map[string]struct{}) (string, error) {
	dir := "ASC"
	if opt.Desc {
		dir = "DESC"
	}
	switch opt.OrderBy {
	case "", "created_at":
		return `"created_at" ` + dir, nil
	case "updated_at":
		return `"updated_at" ` + dir, nil
	}
	if _, ok := fieldNames[opt.OrderBy]; !ok {
		return "", fmt.Errorf("unknown orderBy field: %s", opt.OrderBy)
	}
	col, err := ddl.QuoteIdent(opt.OrderBy)
	if err != nil {
		return "", err
	}
	return col + " " + dir, nil // native typed column sort
}

// Create inserts a record into the real table.
func (s *TableStore) Create(ctx context.Context, obj metadata.Object, workspaceID string, data map[string]any) (Record, error) {
	rel, err := ddl.Relation(workspaceID, obj.NameSingular)
	if err != nil {
		return Record{}, err
	}

	cols := []string{}
	placeholders := []string{}
	args := []any{}
	for name, val := range data {
		col, err := ddl.QuoteIdent(name)
		if err != nil {
			return Record{}, err
		}
		args = append(args, val)
		cols = append(cols, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
	}

	var id string
	var q string
	if len(cols) == 0 {
		q = fmt.Sprintf(`INSERT INTO %s DEFAULT VALUES RETURNING "id"::text`, rel)
	} else {
		q = fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) RETURNING "id"::text`,
			rel, strings.Join(cols, ", "), strings.Join(placeholders, ", "))
	}
	if err := s.pool.QueryRow(ctx, q, args...).Scan(&id); err != nil {
		return Record{}, err
	}
	return s.Get(ctx, obj, workspaceID, id)
}

// Get returns one record from the real table.
func (s *TableStore) Get(ctx context.Context, obj metadata.Object, workspaceID, id string) (Record, error) {
	rel, err := ddl.Relation(workspaceID, obj.NameSingular)
	if err != nil {
		return Record{}, err
	}
	row := s.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM %s t WHERE "id" = $1 AND "deleted_at" IS NULL`, selectExpr, rel), id)
	rec, err := scanTableRecord(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrNotFound
	}
	return rec, err
}

// Update sets the provided columns and returns the updated record.
func (s *TableStore) Update(ctx context.Context, obj metadata.Object, workspaceID, id string, patch map[string]any) (Record, error) {
	rel, err := ddl.Relation(workspaceID, obj.NameSingular)
	if err != nil {
		return Record{}, err
	}
	sets := []string{`"updated_at" = now()`}
	args := []any{id}
	for name, val := range patch {
		col, err := ddl.QuoteIdent(name)
		if err != nil {
			return Record{}, err
		}
		args = append(args, val)
		sets = append(sets, fmt.Sprintf("%s = $%d", col, len(args)))
	}
	q := fmt.Sprintf(`UPDATE %s SET %s WHERE "id" = $1 AND "deleted_at" IS NULL RETURNING "id"::text`,
		rel, strings.Join(sets, ", "))
	var rid string
	err = s.pool.QueryRow(ctx, q, args...).Scan(&rid)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrNotFound
	}
	if err != nil {
		return Record{}, err
	}
	return s.Get(ctx, obj, workspaceID, id)
}

// SoftDelete marks a record deleted.
func (s *TableStore) SoftDelete(ctx context.Context, obj metadata.Object, workspaceID, id string) error {
	rel, err := ddl.Relation(workspaceID, obj.NameSingular)
	if err != nil {
		return err
	}
	tag, err := s.pool.Exec(ctx,
		fmt.Sprintf(`UPDATE %s SET "deleted_at" = now() WHERE "id" = $1 AND "deleted_at" IS NULL`, rel), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
