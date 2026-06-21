// Package ddl is the dynamic-DDL engine: it materializes table-backed objects
// as real Postgres tables (one schema per workspace, typed columns per field).
// This is the high-performance storage path, mirroring Twenty's approach.
package ddl

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aniasse/labzyn/server/internal/metadata"
	"github.com/jackc/pgx/v5/pgxpool"
)

var identRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

// SchemaName returns the per-workspace schema identifier (unquoted).
func SchemaName(workspaceID string) string { return schemaName(workspaceID) }

// Relation returns the safely-quoted "schema"."table" for an object.
func Relation(workspaceID, table string) (string, error) {
	t, err := quote(table)
	if err != nil {
		return "", err
	}
	return `"` + schemaName(workspaceID) + `".` + t, nil
}

// QuoteIdent validates and double-quotes an identifier (column name).
func QuoteIdent(ident string) (string, error) { return quote(ident) }

// Engine runs schema/table DDL.
type Engine struct{ pool *pgxpool.Pool }

// New builds a DDL engine.
func New(pool *pgxpool.Pool) *Engine { return &Engine{pool: pool} }

// schemaName derives a safe per-workspace schema identifier from a UUID.
func schemaName(workspaceID string) string {
	return "ws_" + strings.ReplaceAll(workspaceID, "-", "")
}

func quote(ident string) (string, error) {
	if !identRe.MatchString(ident) {
		return "", fmt.Errorf("unsafe identifier %q", ident)
	}
	return `"` + ident + `"`, nil
}

func (e *Engine) qualified(workspaceID, table string) (string, error) {
	t, err := quote(table)
	if err != nil {
		return "", err
	}
	return `"` + schemaName(workspaceID) + `".` + t, nil
}

// EnsureTable creates the workspace schema and the object's table (with the
// system columns and one typed column per field) if they do not exist.
func (e *Engine) EnsureTable(ctx context.Context, workspaceID string, obj metadata.Object, fields []metadata.Field) error {
	rel, err := e.qualified(workspaceID, obj.NameSingular)
	if err != nil {
		return err
	}

	cols := []string{
		`"id" uuid PRIMARY KEY DEFAULT uuid_generate_v4()`,
		`"created_at" timestamptz NOT NULL DEFAULT now()`,
		`"updated_at" timestamptz NOT NULL DEFAULT now()`,
		`"deleted_at" timestamptz`,
	}
	for _, f := range fields {
		def, err := columnDef(f)
		if err != nil {
			return err
		}
		cols = append(cols, def)
	}

	stmt := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS "%s"; CREATE TABLE IF NOT EXISTS %s (%s);`,
		schemaName(workspaceID), rel, strings.Join(cols, ", "))

	conn, err := e.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	_, err = conn.Conn().PgConn().Exec(ctx, stmt).ReadAll()
	return err
}

// AddColumn adds a column for a new field.
func (e *Engine) AddColumn(ctx context.Context, workspaceID string, obj metadata.Object, field metadata.Field) error {
	rel, err := e.qualified(workspaceID, obj.NameSingular)
	if err != nil {
		return err
	}
	def, err := columnDef(field)
	if err != nil {
		return err
	}
	_, err = e.pool.Exec(ctx, fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s`, rel, def))
	return err
}

// DropColumn removes a field's column.
func (e *Engine) DropColumn(ctx context.Context, workspaceID string, obj metadata.Object, fieldName string) error {
	rel, err := e.qualified(workspaceID, obj.NameSingular)
	if err != nil {
		return err
	}
	col, err := quote(fieldName)
	if err != nil {
		return err
	}
	_, err = e.pool.Exec(ctx, fmt.Sprintf(`ALTER TABLE %s DROP COLUMN IF EXISTS %s`, rel, col))
	return err
}

// DropTable removes the object's table.
func (e *Engine) DropTable(ctx context.Context, workspaceID string, obj metadata.Object) error {
	rel, err := e.qualified(workspaceID, obj.NameSingular)
	if err != nil {
		return err
	}
	_, err = e.pool.Exec(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS %s`, rel))
	return err
}

func columnDef(f metadata.Field) (string, error) {
	col, err := quote(f.Name)
	if err != nil {
		return "", err
	}
	def := col + " " + f.Type.PostgresColumnType()
	if !f.IsNullable {
		def += " NOT NULL"
	}
	if f.IsUnique {
		def += " UNIQUE"
	}
	return def, nil
}
