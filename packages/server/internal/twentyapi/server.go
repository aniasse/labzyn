package twentyapi

import (
	"context"
	_ "embed"
	"net/http"
	"strings"

	"github.com/aniasse/labzyn/server/internal/auth"
	"github.com/aniasse/labzyn/server/internal/core"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed metadata.graphql
var metadataSDL string

//go:embed core.graphql
var coreSDL string

// Server builds and serves Twenty's GraphQL endpoints (core + metadata).
type Server struct {
	secret          string
	coreHandler     http.Handler
	metadataHandler http.Handler
}

// NewServer constructs the Twenty-compatible GraphQL server.
func NewServer(pool *pgxpool.Pool, jwtSecret string) (*Server, error) {
	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers(), graphql.UseStringDescriptions()}
	repo := core.NewRepo(pool)

	metaSchema, err := graphql.ParseSchema(metadataSDL, &MetadataResolver{core: repo, secret: jwtSecret}, opts...)
	if err != nil {
		return nil, err
	}
	coreSchema, err := graphql.ParseSchema(coreSDL, &CoreResolver{core: repo, secret: jwtSecret}, opts...)
	if err != nil {
		return nil, err
	}

	return &Server{
		secret:          jwtSecret,
		coreHandler:     &relay.Handler{Schema: coreSchema},
		metadataHandler: &relay.Handler{Schema: metaSchema},
	}, nil
}

// CoreHandler serves POST /graphql.
func (s *Server) CoreHandler() http.Handler { return cors(s.withAuth(s.coreHandler)) }

// MetadataHandler serves POST /metadata.
func (s *Server) MetadataHandler() http.Handler { return cors(s.withAuth(s.metadataHandler)) }

// ── Auth context ─────────────────────────────────────────────────────────────

type ctxKey string

const userIDKey ctxKey = "userID"

// withAuth extracts the Bearer token (if any) and puts the user id in context.
func (s *Server) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			if uid, err := auth.ParseToken(s.secret, strings.TrimPrefix(header, "Bearer ")); err == nil {
				r = r.WithContext(context.WithValue(r.Context(), userIDKey, uid))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func userIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey).(string)
	return id
}

// cors allows the Twenty frontend (different origin) to call these endpoints.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
