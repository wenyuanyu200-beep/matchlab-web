package config

import (
	"strings"
	"testing"
)

func TestLoadUsesSafeDefaults(t *testing.T) {
	t.Setenv("SERVER_HOST", "")
	t.Setenv("SERVER_PORT", "")
	t.Setenv("GIN_MODE", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_SSLMODE", "")

	cfg := Load()

	if cfg.Address() != "127.0.0.1:8080" {
		t.Fatalf("expected default address, got %q", cfg.Address())
	}
	if cfg.Database.Configured() {
		t.Fatal("database must not be configured without credentials")
	}
}

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv("SERVER_HOST", "0.0.0.0")
	t.Setenv("SERVER_PORT", "9000")
	t.Setenv("GIN_MODE", "release")
	t.Setenv("DB_HOST", "db.internal")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_NAME", "matchlab")
	t.Setenv("DB_USER", "matchlab_user")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("CORS_ALLOWED_ORIGINS", " https://matchlab.example, http://139.224.119.187,https://matchlab.example ")

	cfg := Load()

	if cfg.Address() != "0.0.0.0:9000" {
		t.Fatalf("unexpected address: %q", cfg.Address())
	}
	if cfg.GinMode != "release" {
		t.Fatalf("unexpected Gin mode: %q", cfg.GinMode)
	}
	if !cfg.Database.Configured() {
		t.Fatal("expected database to be configured")
	}
	if len(cfg.CORSAllowedOrigins) != 2 || cfg.CORSAllowedOrigins[0] != "https://matchlab.example" || cfg.CORSAllowedOrigins[1] != "http://139.224.119.187" {
		t.Fatalf("unexpected CORS origins: %#v", cfg.CORSAllowedOrigins)
	}
}

func TestLoadUsesDevelopmentJWTSecretByDefault(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	cfg := Load()

	if cfg.JWTSecret == "" {
		t.Fatal("expected a development JWT secret")
	}
	if !cfg.UsesDevelopmentJWTSecret() {
		t.Fatal("expected default secret to be marked as development-only")
	}
}

func TestLoadReadsJWTSecretFromEnvironment(t *testing.T) {
	t.Setenv("JWT_SECRET", "production-test-secret")

	cfg := Load()

	if cfg.JWTSecret != "production-test-secret" {
		t.Fatalf("unexpected JWT secret: %q", cfg.JWTSecret)
	}
	if cfg.UsesDevelopmentJWTSecret() {
		t.Fatal("custom JWT secret must not be marked as development-only")
	}
}

func TestValidateRuntimeRejectsUnsafeReleaseConfiguration(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "development JWT secret",
			cfg:  Config{GinMode: "release", JWTSecret: developmentJWTSecret, Database: configuredDatabase()},
			want: "JWT_SECRET",
		},
		{
			name: "placeholder JWT secret",
			cfg:  Config{GinMode: "release", JWTSecret: "replace_with_a_long_random_secret", Database: configuredDatabase()},
			want: "JWT_SECRET",
		},
		{
			name: "incomplete database",
			cfg:  Config{GinMode: "release", JWTSecret: strings.Repeat("s", 32)},
			want: "database",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.cfg.ValidateRuntime(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error=%v want substring %q", err, test.want)
			}
		})
	}
}

func TestValidateRuntimeAllowsSafeReleaseAndDegradedDebug(t *testing.T) {
	release := Config{GinMode: "release", JWTSecret: strings.Repeat("s", 32), Database: configuredDatabase()}
	if err := release.ValidateRuntime(); err != nil {
		t.Fatalf("safe release rejected: %v", err)
	}
	debug := Config{GinMode: "debug", JWTSecret: developmentJWTSecret}
	if err := debug.ValidateRuntime(); err != nil {
		t.Fatalf("debug configuration rejected: %v", err)
	}
}

func configuredDatabase() Database {
	return Database{Host: "db", Port: "5432", Name: "matchlab", User: "matchlab", Password: "secret"}
}
