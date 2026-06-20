package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/aniasse/labzyn/server/internal/httpx"
)

type ctxKey string

const workspaceIDKey ctxKey = "workspaceID"

// nameRe validates object/field identifiers (used to build safe SQL fragments).
var nameRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

func validName(s string) bool { return nameRe.MatchString(s) }

// workspaceCtx resolves the current workspace from the X-Workspace-Id header,
// falling back to the oldest workspace in development (pre-auth).
func (s *Server) workspaceCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Workspace-Id")
		if id == "" {
			var err error
			id, err = s.core.OldestWorkspaceID(r.Context())
			if err != nil {
				httpx.Error(w, http.StatusInternalServerError, "could not resolve workspace")
				return
			}
			if id == "" {
				httpx.Error(w, http.StatusBadRequest,
					"no workspace found; create one via POST /api/v1/workspaces or set X-Workspace-Id")
				return
			}
		}
		ctx := context.WithValue(r.Context(), workspaceIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func workspaceID(r *http.Request) string {
	id, _ := r.Context().Value(workspaceIDKey).(string)
	return id
}

// decodeJSON reads a JSON body (capped at 1 MiB) into v.
func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
