package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	// ErrNotFound means the requested user does not exist.
	ErrNotFound = errors.New("user not found")
	// ErrEmailExists means a normalized email already belongs to an account.
	ErrEmailExists = errors.New("email already exists")
	// ErrUnavailable means no user storage is available.
	ErrUnavailable = errors.New("user repository unavailable")
)

// Repository describes the persistence operations needed by authentication.
type Repository interface {
	Create(ctx context.Context, model *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
}

// GormRepository stores users with GORM.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a user repository. A nil database produces
// ErrUnavailable rather than a panic so the health route can remain online.
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// Create inserts a user.
func (r *GormRepository) Create(ctx context.Context, model *User) error {
	if r.db == nil {
		return ErrUnavailable
	}
	return mapRepositoryError(r.db.WithContext(ctx).Create(model).Error)
}

// FindByEmail returns a user by normalized, case-insensitive email.
func (r *GormRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	var model User
	err := r.db.WithContext(ctx).
		Where("LOWER(email) = ?", strings.ToLower(strings.TrimSpace(email))).
		First(&model).Error
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return &model, nil
}

// FindByID returns a user by UUID.
func (r *GormRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	var model User
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, mapRepositoryError(err)
	}
	return &model, nil
}

func mapRepositoryError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == "23505" {
		return ErrEmailExists
	}
	return fmt.Errorf("user repository: %w", err)
}
