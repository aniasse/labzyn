package api

import (
	"net/http"
	"strings"

	"github.com/aniasse/labzyn/server/internal/httpx"
)

func (s *Server) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var in struct {
		DisplayName string `json:"displayName"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	if in.DisplayName == "" {
		httpx.Error(w, http.StatusBadRequest, "displayName is required")
		return
	}

	ws, err := s.core.CreateWorkspace(r.Context(), in.DisplayName, subdomain(in.DisplayName))
	if err != nil {
		httpx.Error(w, http.StatusConflict, "could not create workspace")
		return
	}
	if err := s.core.AddMembership(r.Context(), userID(r), ws.ID, "admin"); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not add membership")
		return
	}
	httpx.JSON(w, http.StatusCreated, ws)
}

func (s *Server) handleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	list, err := s.core.ListUserWorkspaces(r.Context(), userID(r))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list workspaces")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"data": list})
}
