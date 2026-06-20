package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aniasse/labzyn/server/internal/auth"
	"github.com/aniasse/labzyn/server/internal/core"
	"github.com/aniasse/labzyn/server/internal/httpx"
)

const tokenTTL = 7 * 24 * time.Hour

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email         string `json:"email"`
		Password      string `json:"password"`
		FirstName     string `json:"firstName"`
		LastName      string `json:"lastName"`
		WorkspaceName string `json:"workspaceName"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	if in.Email == "" || len(in.Password) < 6 {
		httpx.Error(w, http.StatusBadRequest, "email and a password of at least 6 characters are required")
		return
	}
	wsName := strings.TrimSpace(in.WorkspaceName)
	if wsName == "" {
		wsName = "My workspace"
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	user, ws, err := s.core.Signup(r.Context(), in.Email, hash, in.FirstName, in.LastName, wsName, subdomain(wsName))
	if err != nil {
		if isUniqueViolation(err) {
			httpx.Error(w, http.StatusConflict, "an account with this email already exists")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "could not create account")
		return
	}

	s.issueAuth(w, http.StatusCreated, user, []core.Workspace{ws}, ws.ID)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))

	user, err := s.core.GetUserByEmail(r.Context(), in.Email)
	if err != nil || !auth.CheckPassword(user.PasswordHash, in.Password) {
		httpx.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	workspaces, err := s.core.ListUserWorkspaces(r.Context(), user.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not load workspaces")
		return
	}
	current := ""
	if len(workspaces) > 0 {
		current = workspaces[0].ID
	}
	s.issueAuth(w, http.StatusOK, user, workspaces, current)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, err := s.core.GetUserByID(r.Context(), userID(r))
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "user not found")
		return
	}
	workspaces, err := s.core.ListUserWorkspaces(r.Context(), user.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not load workspaces")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"user":               user,
		"workspaces":         workspaces,
		"currentWorkspaceId": workspaceID(r),
	})
}

// issueAuth signs a token and writes the standard auth response.
func (s *Server) issueAuth(w http.ResponseWriter, status int, user core.User, workspaces []core.Workspace, current string) {
	token, err := auth.NewToken(s.jwtSecret, user.ID, tokenTTL)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	httpx.JSON(w, status, map[string]any{
		"token":              token,
		"user":               user,
		"workspaces":         workspaces,
		"currentWorkspaceId": current,
	})
}

// subdomain builds a unique-ish subdomain from a workspace name.
func subdomain(name string) string {
	base := strings.Trim(slugRe.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if base == "" {
		base = "ws"
	}
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return base + "-" + hex.EncodeToString(b)
}
