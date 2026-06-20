package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/aniasse/labzyn/server/internal/auth"
	"github.com/aniasse/labzyn/server/internal/httpx"
)

type ctxKey string

const (
	userIDKey      ctxKey = "userID"
	workspaceIDKey ctxKey = "workspaceID"
)

// nameRe validates object/field identifiers (used to build safe SQL fragments).
var nameRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

func validName(s string) bool { return nameRe.MatchString(s) }

// requireAuth validates the Bearer token, resolves the active workspace, and
// puts the user and workspace ids in the request context.
func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			httpx.Error(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		uid, err := auth.ParseToken(s.jwtSecret, strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			httpx.Error(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// Resolve workspace: honor X-Workspace-Id only if the user is a member,
		// otherwise fall back to the user's default workspace.
		wsID := ""
		if requested := r.Header.Get("X-Workspace-Id"); requested != "" {
			ok, err := s.core.IsMember(r.Context(), uid, requested)
			if err != nil {
				httpx.Error(w, http.StatusInternalServerError, "could not verify workspace access")
				return
			}
			if !ok {
				httpx.Error(w, http.StatusForbidden, "not a member of this workspace")
				return
			}
			wsID = requested
		} else {
			wsID, err = s.core.DefaultWorkspaceID(r.Context(), uid)
			if err != nil {
				httpx.Error(w, http.StatusInternalServerError, "could not resolve workspace")
				return
			}
		}

		ctx := context.WithValue(r.Context(), userIDKey, uid)
		ctx = context.WithValue(ctx, workspaceIDKey, wsID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func userID(r *http.Request) string {
	id, _ := r.Context().Value(userIDKey).(string)
	return id
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
