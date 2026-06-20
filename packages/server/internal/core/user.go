package core

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// ErrNotFound is returned when a user/membership lookup finds nothing.
var ErrNotFound = errors.New("not found")

// User is an account. PasswordHash is never serialized to clients.
type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	PasswordHash string `json:"-"`
}

func scanUser(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.PasswordHash)
	return u, err
}

const userCols = `id::text, email, first_name, last_name, coalesce(password_hash, '')`

// GetUserByEmail looks up a user by email.
func (r *Repo) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+userCols+` FROM core."user" WHERE email = $1 AND deleted_at IS NULL`, email)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

// GetUserByID looks up a user by id.
func (r *Repo) GetUserByID(ctx context.Context, id string) (User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+userCols+` FROM core."user" WHERE id = $1 AND deleted_at IS NULL`, id)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

// Signup creates a user, a workspace, and the membership linking them, all in
// one transaction. Returns the created user and workspace.
func (r *Repo) Signup(
	ctx context.Context,
	email, passwordHash, firstName, lastName, workspaceName, subdomain string,
) (User, Workspace, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return User{}, Workspace{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after commit

	u, err := scanUser(tx.QueryRow(ctx,
		`INSERT INTO core."user" (email, first_name, last_name, password_hash)
		 VALUES ($1,$2,$3,$4) RETURNING `+userCols,
		email, firstName, lastName, passwordHash))
	if err != nil {
		return User{}, Workspace{}, err
	}

	w, err := scanWorkspace(tx.QueryRow(ctx,
		`INSERT INTO core.workspace (display_name, subdomain) VALUES ($1,$2)
		 RETURNING `+workspaceCols, workspaceName, subdomain))
	if err != nil {
		return User{}, Workspace{}, err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO core.user_workspace (user_id, workspace_id, role) VALUES ($1,$2,'admin')`,
		u.ID, w.ID); err != nil {
		return User{}, Workspace{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, Workspace{}, err
	}
	return u, w, nil
}

// ListUserWorkspaces returns the workspaces a user belongs to.
func (r *Repo) ListUserWorkspaces(ctx context.Context, userID string) ([]Workspace, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT w.id::text, w.display_name, w.subdomain, w.created_at, w.updated_at
		 FROM core.workspace w
		 JOIN core.user_workspace uw ON uw.workspace_id = w.id
		 WHERE uw.user_id = $1 AND w.deleted_at IS NULL
		 ORDER BY w.created_at ASC`, userID)
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

// IsMember reports whether a user belongs to a workspace.
func (r *Repo) IsMember(ctx context.Context, userID, workspaceID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM core.user_workspace WHERE user_id = $1 AND workspace_id = $2)`,
		userID, workspaceID).Scan(&exists)
	return exists, err
}

// DefaultWorkspaceID returns the user's first workspace id, or "" if none.
func (r *Repo) DefaultWorkspaceID(ctx context.Context, userID string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`SELECT w.id::text FROM core.workspace w
		 JOIN core.user_workspace uw ON uw.workspace_id = w.id
		 WHERE uw.user_id = $1 AND w.deleted_at IS NULL
		 ORDER BY w.created_at ASC LIMIT 1`, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return id, err
}

// AddMembership links a user to a workspace with a role (idempotent).
func (r *Repo) AddMembership(ctx context.Context, userID, workspaceID, role string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO core.user_workspace (user_id, workspace_id, role) VALUES ($1,$2,$3)
		 ON CONFLICT (user_id, workspace_id) DO NOTHING`, userID, workspaceID, role)
	return err
}
