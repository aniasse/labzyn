// Package config loads runtime configuration from the environment.
package config

import "os"

// Config holds the server's runtime configuration.
type Config struct {
	Port        string
	DatabaseURL string
	GatewayURL  string
	JWTSecret   string
}

// Load reads configuration from environment variables, applying sane defaults
// for local development.
func Load() Config {
	return Config{
		Port:        env("LABZYN_PORT", "4000"),
		DatabaseURL: env("LABZYN_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/labzyn?sslmode=disable"),
		GatewayURL:  env("LABZYN_GATEWAY_URL", "http://localhost:8088"),
		// Dev default; MUST be overridden in production.
		JWTSecret: env("LABZYN_JWT_SECRET", "dev-insecure-secret-change-me"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
