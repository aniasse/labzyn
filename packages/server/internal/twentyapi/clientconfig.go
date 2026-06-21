// Package twentyapi implements the Twenty backend contract (REST /client-config
// + GraphQL /graphql and /metadata) so Twenty's frontend runs against Go.
package twentyapi

import (
	_ "embed"
	"net/http"
)

// clientConfigJSON is the exact /client-config payload Twenty's frontend expects
// at boot (captured from a reference Twenty server).
//
//go:embed client-config.json
var clientConfigJSON []byte

// HandleClientConfig serves the boot configuration.
func HandleClientConfig(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(clientConfigJSON)
}
