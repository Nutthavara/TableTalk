// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr      string
	DatabaseURL   string
	OIDCIssuerURL string
	OIDCAudience  string
	UploadDir     string
	PublicBaseURL string
	// APIDocsEnabled serves Swagger UI at /docs. Off unless explicitly enabled (development).
	APIDocsEnabled bool
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:       getenv("HTTP_ADDR", ":8081"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		OIDCIssuerURL:  os.Getenv("OIDC_ISSUER_URL"),
		OIDCAudience:   getenv("OIDC_AUDIENCE", "tabletalk-api"),
		UploadDir:      getenv("UPLOAD_DIR", "./uploads"),
		PublicBaseURL:  strings.TrimRight(getenv("PUBLIC_BASE_URL", "http://localhost:8081"), "/"),
		APIDocsEnabled: os.Getenv("API_DOCS_ENABLED") == "true",
	}

	var missing []string
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.OIDCIssuerURL == "" {
		missing = append(missing, "OIDC_ISSUER_URL")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
