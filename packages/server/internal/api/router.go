// Package api wires HTTP routes to handlers.
package api

import (
	"net/http"
	"time"

	"github.com/aniasse/labzyn/server/internal/core"
	"github.com/aniasse/labzyn/server/internal/httpx"
	"github.com/aniasse/labzyn/server/internal/metadata"
	"github.com/aniasse/labzyn/server/internal/records"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server holds dependencies shared by handlers.
type Server struct {
	pool *pgxpool.Pool
	core *core.Repo
	meta *metadata.Repo
	rec  *records.Repo
}

// NewServer builds the HTTP handler tree.
func NewServer(pool *pgxpool.Pool) http.Handler {
	s := &Server{
		pool: pool,
		core: core.NewRepo(pool),
		meta: metadata.NewRepo(pool),
		rec:  records.NewRepo(pool),
	}

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

	r.Route("/api/v1", func(r chi.Router) {
		// Workspaces are tenant-level: no workspace context required.
		r.Post("/workspaces", s.handleCreateWorkspace)
		r.Get("/workspaces", s.handleListWorkspaces)

		// Everything below operates within a resolved workspace.
		r.Group(func(r chi.Router) {
			r.Use(s.workspaceCtx)

			// Metadata catalog.
			r.Post("/metadata/objects", s.handleCreateObject)
			r.Get("/metadata/objects", s.handleListObjects)
			r.Get("/metadata/objects/{nameSingular}", s.handleGetObject)
			r.Post("/metadata/objects/{nameSingular}/fields", s.handleCreateField)

			// Generic record CRUD, addressed by the object's plural name.
			r.Get("/{objectPlural}", s.handleListRecords)
			r.Post("/{objectPlural}", s.handleCreateRecord)
			r.Get("/{objectPlural}/{id}", s.handleGetRecord)
			r.Patch("/{objectPlural}/{id}", s.handleUpdateRecord)
			r.Delete("/{objectPlural}/{id}", s.handleDeleteRecord)
		})
	})

	return r
}
