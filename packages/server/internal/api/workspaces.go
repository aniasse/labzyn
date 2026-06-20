package api

import (
	"net/http"
	"strings"

	"github.com/aniasse/labzyn/server/internal/httpx"
)

func (s *Server) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var in struct {
		DisplayName string `json:"displayName"`
		Subdomain   string `json:"subdomain"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	in.Subdomain = strings.TrimSpace(in.Subdomain)
	if in.DisplayName == "" || in.Subdomain == "" {
		httpx.Error(w, http.StatusBadRequest, "displayName and subdomain are required")
		return
	}
	ws, err := s.core.CreateWorkspace(r.Context(), in.DisplayName, in.Subdomain)
	if err != nil {
		httpx.Error(w, http.StatusConflict, "could not create workspace (subdomain taken?)")
		return
	}
	httpx.JSON(w, http.StatusCreated, ws)
}

func (s *Server) handleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	list, err := s.core.ListWorkspaces(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list workspaces")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"data": list})
}
