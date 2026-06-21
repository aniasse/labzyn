package twentyapi

import (
	_ "embed"
	"net/http"

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
	coreHandler     http.Handler
	metadataHandler http.Handler
}

// coreResolver is the root resolver for /graphql (records come later).
type coreResolver struct{}

func (*coreResolver) Ping() *string { s := "labzyn-core"; return &s }

// NewServer constructs the Twenty-compatible GraphQL server.
func NewServer(pool *pgxpool.Pool, jwtSecret string) (*Server, error) {
	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers(), graphql.UseStringDescriptions()}

	metaResolver := &MetadataResolver{core: core.NewRepo(pool), secret: jwtSecret}
	metaSchema, err := graphql.ParseSchema(metadataSDL, metaResolver, opts...)
	if err != nil {
		return nil, err
	}
	coreSchema, err := graphql.ParseSchema(coreSDL, &coreResolver{}, opts...)
	if err != nil {
		return nil, err
	}

	return &Server{
		coreHandler:     &relay.Handler{Schema: coreSchema},
		metadataHandler: &relay.Handler{Schema: metaSchema},
	}, nil
}

// CoreHandler serves POST /graphql.
func (s *Server) CoreHandler() http.Handler { return cors(s.coreHandler) }

// MetadataHandler serves POST /metadata.
func (s *Server) MetadataHandler() http.Handler { return cors(s.metadataHandler) }

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
