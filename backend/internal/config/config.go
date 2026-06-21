// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const developmentJWTSecret = "matchlab-development-secret-change-before-production"

// Config contains process-level configuration.
type Config struct {
	ServerHost         string
	ServerPort         string
	GinMode            string
	JWTSecret          string
	CORSAllowedOrigins []string
	Database           Database
}

// Database contains PostgreSQL connection settings.
type Database struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

// Load reads .env when present and then resolves configuration from the
// environment. Existing environment variables take precedence over .env.
func Load() Config {
	_ = godotenv.Load()

	return Config{
		ServerHost:         envOrDefault("SERVER_HOST", "127.0.0.1"),
		ServerPort:         envOrDefault("SERVER_PORT", "8080"),
		GinMode:            envOrDefault("GIN_MODE", "debug"),
		JWTSecret:          envOrDefault("JWT_SECRET", developmentJWTSecret),
		CORSAllowedOrigins: splitUnique(os.Getenv("CORS_ALLOWED_ORIGINS")),
		Database: Database{
			Host:     strings.TrimSpace(os.Getenv("DB_HOST")),
			Port:     envOrDefault("DB_PORT", "5432"),
			Name:     strings.TrimSpace(os.Getenv("DB_NAME")),
			User:     strings.TrimSpace(os.Getenv("DB_USER")),
			Password: os.Getenv("DB_PASSWORD"),
			SSLMode:  envOrDefault("DB_SSLMODE", "disable"),
		},
	}
}

// UsesDevelopmentJWTSecret reports whether the unsafe development fallback is active.
func (c Config) UsesDevelopmentJWTSecret() bool {
	return c.JWTSecret == developmentJWTSecret
}

// IsRelease reports whether production runtime safeguards must be enforced.
func (c Config) IsRelease() bool {
	return strings.EqualFold(strings.TrimSpace(c.GinMode), "release")
}

// ValidateRuntime rejects configurations that would start a broken or unsafe production API.
func (c Config) ValidateRuntime() error {
	if !c.IsRelease() {
		return nil
	}
	secret := strings.TrimSpace(c.JWTSecret)
	if c.UsesDevelopmentJWTSecret() || secret == "replace_with_a_long_random_secret" || len(secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be a non-placeholder secret of at least 32 characters in release mode")
	}
	if !c.Database.Configured() {
		return fmt.Errorf("database configuration must be complete in release mode")
	}
	return nil
}

// Address returns the HTTP listen address.
func (c Config) Address() string {
	return c.ServerHost + ":" + c.ServerPort
}

// Configured reports whether sufficient PostgreSQL settings were supplied.
func (d Database) Configured() bool {
	return d.Host != "" && d.Port != "" && d.Name != "" && d.User != "" && d.Password != ""
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func splitUnique(value string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, item := range strings.Split(value, ",") {
		origin := strings.TrimRight(strings.TrimSpace(item), "/")
		if origin == "" {
			continue
		}
		if _, exists := seen[origin]; exists {
			continue
		}
		seen[origin] = struct{}{}
		result = append(result, origin)
	}
	return result
}
