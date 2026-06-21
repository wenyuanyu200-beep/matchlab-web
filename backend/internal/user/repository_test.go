package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func TestMapRepositoryErrorMapsUniqueViolation(t *testing.T) {
	err := mapRepositoryError(&pgconn.PgError{Code: "23505"})

	if !errors.Is(err, ErrEmailExists) {
		t.Fatalf("expected ErrEmailExists, got %v", err)
	}
}

func TestMapRepositoryErrorMapsRecordNotFound(t *testing.T) {
	err := mapRepositoryError(gorm.ErrRecordNotFound)

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGormRepositoryWithoutDatabaseIsUnavailable(t *testing.T) {
	repository := NewGormRepository(nil)

	if err := repository.Create(context.Background(), &User{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("create: expected ErrUnavailable, got %v", err)
	}
	if _, err := repository.FindByEmail(context.Background(), "test@example.com"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("find by email: expected ErrUnavailable, got %v", err)
	}
	if _, err := repository.FindByID(context.Background(), uuid.New()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("find by ID: expected ErrUnavailable, got %v", err)
	}
}
