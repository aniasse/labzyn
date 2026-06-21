// Package config loads runtime configuration from the environment.
package config

import "os"

// Config holds the server's runtime configuration.
type Config struct {
	Port        string
	DatabaseURL string
	GatewayURL  string
	JWTSecret   string

	// LLM (OpenAI-compatible: OpenAI, Groq, Mistral, …). Agent is disabled
	// until an API key is set.
	LLMBaseURL string
	LLMAPIKey  string
	LLMModel   string
}

// Load reads configuration from environment variables, applying sane defaults
// for local development.
func Load() Config {
	return Config{
		Port:        env("LABZYN_PORT", "4000"),
		DatabaseURL: env("LABZYN_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/labzyn?sslmode=disable"),
		GatewayURL:  env("LABZYN_GATEWAY_URL", "http://localhost:8088"),
		// Dev default; MUST be overridden in production.
		JWTSecret:  env("LABZYN_JWT_SECRET", "dev-insecure-secret-change-me"),
		LLMBaseURL: env("LABZYN_LLM_BASE_URL", "https://api.groq.com/openai/v1"),
		LLMAPIKey:  env("LABZYN_LLM_API_KEY", ""),
		LLMModel:   env("LABZYN_LLM_MODEL", "llama-3.3-70b-versatile"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
