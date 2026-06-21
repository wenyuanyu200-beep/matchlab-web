// Package config loads application configuration from environment variables.
package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const developmentJWTSecret = "matchlab-development-secret-change-before-production"

// Config contains process-level configuration.
type Config struct {
	ServerHost string
	ServerPort string
	GinMode    string
	JWTSecret  string
	Database   Database
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
		ServerHost: envOrDefault("SERVER_HOST", "127.0.0.1"),
		ServerPort: envOrDefault("SERVER_PORT", "8080"),
		GinMode:    envOrDefault("GIN_MODE", "debug"),
		JWTSecret:  envOrDefault("JWT_SECRET", developmentJWTSecret),
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
