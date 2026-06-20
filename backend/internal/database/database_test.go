package database

import (
	"strings"
	"testing"

	"matchlab/backend/internal/config"
)

func TestDSNContainsPostgresSettings(t *testing.T) {
	cfg := config.Database{
		Host:     "127.0.0.1",
		Port:     "5432",
		Name:     "matchlab",
		User:     "matchlab_user",
		Password: "p@ss word",
		SSLMode:  "disable",
	}

	dsn := DSN(cfg)

	for _, expected := range []string{
		"host=127.0.0.1", "port=5432", "dbname=matchlab",
		"user=matchlab_user", "password='p@ss word'", "sslmode=disable",
	} {
		if !strings.Contains(dsn, expected) {
			t.Fatalf("DSN %q does not contain %q", dsn, expected)
		}
	}
}

func TestSafeDescriptionOmitsPassword(t *testing.T) {
	cfg := config.Database{
		Host: "db", Port: "5432", Name: "matchlab", User: "app", Password: "do-not-log",
	}

	description := SafeDescription(cfg)

	if strings.Contains(description, cfg.Password) {
		t.Fatal("safe description exposed the password")
	}
}
