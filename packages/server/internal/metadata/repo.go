package metadata

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a catalog entry does not exist.
var ErrNotFound = errors.New("not found")

// Repo provides metadata (object/field) persistence.
type Repo struct{ pool *pgxpool.Pool }

// NewRepo builds a metadata repository.
func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

// CreateObjectInput carries the fields needed to create an object.
type CreateObjectInput struct {
	NameSingular  string
	NamePlural    string
	LabelSingular string
	LabelPlural   string
	Labels        []byte // jsonb, may be nil
	Icon          string
	Description   string
	Storage       string // "jsonb" (default) | "table"
}

const objectCols = `id::text, workspace_id::text, name_singular, name_plural,
	label_singular, label_plural, labels, icon, description, is_system, is_active,
	storage, created_at, updated_at`

func scanObject(row pgx.Row) (Object, error) {
	var o Object
	err := row.Scan(&o.ID, &o.WorkspaceID, &o.NameSingular, &o.NamePlural,
		&o.LabelSingular, &o.LabelPlural, &o.Labels, &o.Icon, &o.Description,
		&o.IsSystem, &o.IsActive, &o.Storage, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}

// CreateObject inserts a new object in the given workspace.
func (r *Repo) CreateObject(ctx context.Context, workspaceID string, in CreateObjectInput) (Object, error) {
	labels := in.Labels
	if labels == nil {
		labels = []byte("{}")
	}
	icon := in.Icon
	if icon == "" {
		icon = "IconList"
	}
	storage := in.Storage
	if storage == "" {
		storage = "jsonb"
	}
	row := r.pool.QueryRow(ctx,
		`INSERT INTO metadata.object
		   (workspace_id, name_singular, name_plural, label_singular, label_plural, labels, icon, description, storage)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 RETURNING `+objectCols,
		workspaceID, in.NameSingular, in.NamePlural, in.LabelSingular, in.LabelPlural,
		labels, icon, in.Description, storage)
	return scanObject(row)
}

// ListAllObjects returns every active object across all workspaces (used to
// build the dynamic records GraphQL schema).
func (r *Repo) ListAllObjects(ctx context.Context) ([]Object, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+objectCols+` FROM metadata.object
		 WHERE is_active = true ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Object{}
	for rows.Next() {
		o, err := scanObject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// ListObjects returns active objects for a workspace.
func (r *Repo) ListObjects(ctx context.Context, workspaceID string) ([]Object, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+objectCols+` FROM metadata.object
		 WHERE workspace_id = $1 AND is_active = true
		 ORDER BY created_at ASC`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Object{}
	for rows.Next() {
		o, err := scanObject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *Repo) getObjectBy(ctx context.Context, workspaceID, column, value string) (Object, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+objectCols+` FROM metadata.object
		 WHERE workspace_id = $1 AND `+column+` = $2 AND is_active = true`,
		workspaceID, value)
	o, err := scanObject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Object{}, ErrNotFound
	}
	return o, err
}

// GetObjectBySingular looks up an object by its singular name.
func (r *Repo) GetObjectBySingular(ctx context.Context, workspaceID, name string) (Object, error) {
	return r.getObjectBy(ctx, workspaceID, "name_singular", name)
}

// GetObjectByPlural looks up an object by its plural name (used in record URLs).
func (r *Repo) GetObjectByPlural(ctx context.Context, workspaceID, name string) (Object, error) {
	return r.getObjectBy(ctx, workspaceID, "name_plural", name)
}

// UpdateObjectInput carries editable object attributes (nil = unchanged).
type UpdateObjectInput struct {
	LabelSingular *string
	LabelPlural   *string
	Icon          *string
	Description   *string
}

// UpdateObject edits an object's labels/icon/description.
func (r *Repo) UpdateObject(ctx context.Context, workspaceID, nameSingular string, in UpdateObjectInput) (Object, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE metadata.object SET
		   label_singular = coalesce($3, label_singular),
		   label_plural   = coalesce($4, label_plural),
		   icon           = coalesce($5, icon),
		   description    = coalesce($6, description),
		   updated_at     = now()
		 WHERE workspace_id = $1 AND name_singular = $2 AND is_active = true
		 RETURNING `+objectCols,
		workspaceID, nameSingular, in.LabelSingular, in.LabelPlural, in.Icon, in.Description)
	o, err := scanObject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Object{}, ErrNotFound
	}
	return o, err
}

// DeleteObject hard-deletes an object; fields and records cascade.
func (r *Repo) DeleteObject(ctx context.Context, workspaceID, nameSingular string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM metadata.object WHERE workspace_id = $1 AND name_singular = $2`,
		workspaceID, nameSingular)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateField edits a field's label (scoped to its object).
func (r *Repo) UpdateField(ctx context.Context, objectID, fieldID, label string) (Field, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE metadata.field SET label = $3, updated_at = now()
		 WHERE id = $1 AND object_id = $2 AND is_active = true
		 RETURNING `+fieldCols, fieldID, objectID, label)
	f, err := scanField(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Field{}, ErrNotFound
	}
	return f, err
}

// DeleteField hard-deletes a field (scoped to its object).
func (r *Repo) DeleteField(ctx context.Context, objectID, fieldID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM metadata.field WHERE id = $1 AND object_id = $2 AND is_system = false`,
		fieldID, objectID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateFieldInput carries the fields needed to create a field.
type CreateFieldInput struct {
	Name         string
	Type         FieldType
	Label        string
	Labels       []byte
	Icon         string
	Description  string
	IsNullable   bool
	IsUnique     bool
	DefaultValue []byte
	Options      []byte
	Position     int
}

const fieldCols = `id::text, object_id::text, name, type, label, labels, icon,
	description, is_nullable, is_unique, is_system, is_active, default_value,
	options, position, created_at, updated_at`

func scanField(row pgx.Row) (Field, error) {
	var f Field
	err := row.Scan(&f.ID, &f.ObjectID, &f.Name, &f.Type, &f.Label, &f.Labels,
		&f.Icon, &f.Description, &f.IsNullable, &f.IsUnique, &f.IsSystem,
		&f.IsActive, &f.DefaultValue, &f.Options, &f.Position, &f.CreatedAt, &f.UpdatedAt)
	return f, err
}

// CreateField inserts a new field on an object.
func (r *Repo) CreateField(ctx context.Context, objectID string, in CreateFieldInput) (Field, error) {
	labels := in.Labels
	if labels == nil {
		labels = []byte("{}")
	}
	row := r.pool.QueryRow(ctx,
		`INSERT INTO metadata.field
		   (object_id, name, type, label, labels, icon, description, is_nullable, is_unique, default_value, options, position)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		 RETURNING `+fieldCols,
		objectID, in.Name, string(in.Type), in.Label, labels, in.Icon, in.Description,
		in.IsNullable, in.IsUnique, nullableJSON(in.DefaultValue), nullableJSON(in.Options), in.Position)
	return scanField(row)
}

// ListFields returns active fields for an object, ordered by position.
func (r *Repo) ListFields(ctx context.Context, objectID string) ([]Field, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+fieldCols+` FROM metadata.field
		 WHERE object_id = $1 AND is_active = true
		 ORDER BY position ASC, created_at ASC`, objectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Field{}
	for rows.Next() {
		f, err := scanField(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// nullableJSON returns nil for empty input so it inserts SQL NULL.
func nullableJSON(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}
