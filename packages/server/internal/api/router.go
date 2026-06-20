// Package api wires HTTP routes to handlers.
package api

import (
	"net/http"
	"time"

	"github.com/aniasse/labzyn/server/internal/httpx"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server holds dependencies shared by handlers.
type Server struct {
	pool *pgxpool.Pool
}

// NewServer builds the HTTP handler tree.
func NewServer(pool *pgxpool.Pool) http.Handler {
	s := &Server{pool: pool}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/healthz", s.handleHealth)
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{
			"service": "labzyn-server",
			"version": "0.1.0",
		})
	})

	// API v1 — metadata & records land here in the next milestone.
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
			httpx.JSON(w, http.StatusOK, map[string]string{"api": "v1"})
		})
	})

	return r
}
