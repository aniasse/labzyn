// Package httpx provides small HTTP response helpers used across handlers.
package httpx

import (
	"encoding/json"
	"net/http"
)

// JSON writes v as a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// Error writes a JSON error envelope: {"error": {"message": "..."}}.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]any{"error": map[string]string{"message": message}})
}
