package admin

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"matchlab/backend/internal/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestAdminUserFromModelNeverExposesPassword(t *testing.T) {
	got := adminUserFromModel(user.User{
		ID:           uuid.New(),
		Email:        "admin@example.com",
		PasswordHash: "secret-hash",
		Nickname:     "Admin",
		Role:         "admin",
		School:       "MatchLab University",
	})

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(encoded))
	if strings.Contains(text, "password") || strings.Contains(text, "secret-hash") {
		t.Fatalf("safe user response leaked password data: %s", text)
	}
}

func TestGormRepositoryWithoutDatabaseIsUnavailable(t *testing.T) {
	repository := NewGormRepository(nil)
	ctx := context.Background()

	_, err := repository.Stats(ctx)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Stats() error = %v, want ErrUnavailable", err)
	}
	_, err = repository.Users(ctx, UsersFilter{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Users() error = %v, want ErrUnavailable", err)
	}
	_, err = repository.Activities(ctx, ActivitiesFilter{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Activities() error = %v, want ErrUnavailable", err)
	}
	_, err = repository.Applications(ctx, ApplicationsFilter{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Applications() error = %v, want ErrUnavailable", err)
	}
	_, err = repository.Feedbacks(ctx, Page{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Feedbacks() error = %v, want ErrUnavailable", err)
	}
	_, err = repository.UpdateUserRole(ctx, uuid.New(), "admin")
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("UpdateUserRole() error = %v, want ErrUnavailable", err)
	}
}

func TestMapRepositoryError(t *testing.T) {
	if !errors.Is(mapRepositoryError(gorm.ErrRecordNotFound), ErrNotFound) {
		t.Fatal("record-not-found was not mapped to ErrNotFound")
	}
}
