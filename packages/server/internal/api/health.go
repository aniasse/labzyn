package api

import (
	"context"
	"net/http"
	"time"

	"github.com/aniasse/labzyn/server/internal/httpx"
)

// handleHealth reports liveness and database connectivity.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	dbStatus := "ok"
	status := http.StatusOK
	if err := s.pool.Ping(ctx); err != nil {
		dbStatus = "down"
		status = http.StatusServiceUnavailable
	}

	httpx.JSON(w, status, map[string]any{
		"status":   map[bool]string{true: "ok", false: "degraded"}[status == http.StatusOK],
		"service":  "labzyn-server",
		"database": dbStatus,
	})
}
