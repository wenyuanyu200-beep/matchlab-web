package config

import "testing"

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
}
