package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"matchlab/backend/internal/user"
)

type fakeRepository struct {
	byEmail   map[string]user.User
	byID      map[uuid.UUID]user.User
	createErr error
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{byEmail: map[string]user.User{}, byID: map[uuid.UUID]user.User{}}
}

func (r *fakeRepository) Create(_ context.Context, model *user.User) error {
	if r.createErr != nil {
		return r.createErr
	}
	if _, exists := r.byEmail[model.Email]; exists {
		return user.ErrEmailExists
	}
	if model.ID == uuid.Nil {
		model.ID = uuid.New()
	}
	if model.CreatedAt.IsZero() {
		model.CreatedAt = time.Now().UTC()
		model.UpdatedAt = model.CreatedAt
	}
	r.byEmail[model.Email] = *model
	r.byID[model.ID] = *model
	return nil
}

func (r *fakeRepository) FindByEmail(_ context.Context, email string) (*user.User, error) {
	model, ok := r.byEmail[email]
	if !ok {
		return nil, user.ErrNotFound
	}
	return &model, nil
}

func (r *fakeRepository) FindByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	model, ok := r.byID[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	return &model, nil
}

type fakeIssuer struct {
	token  string
	err    error
	issued user.User
}

func (i *fakeIssuer) Issue(model user.User) (string, error) {
	i.issued = model
	return i.token, i.err
}

func TestRegisterNormalizesAndHashesPassword(t *testing.T) {
	repository := newFakeRepository()
	service := NewService(repository, &fakeIssuer{})

	result, err := service.Register(context.Background(), RegisterInput{
		Email: "  TEST@Example.COM ", Password: "12345678",
		Nickname: " 测试用户 ", School: " 西南大学 ",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	stored := repository.byEmail["test@example.com"]
	if stored.PasswordHash == "12345678" || bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("12345678")) != nil {
		t.Fatal("password was not stored as a bcrypt hash")
	}
	if stored.Role != "user" || stored.Nickname != "测试用户" || stored.School != "西南大学" {
		t.Fatalf("unexpected stored user: %#v", stored)
	}
	if result.Email != stored.Email || result.ID != stored.ID {
		t.Fatalf("unexpected public result: %#v", result)
	}
}

func TestRegisterValidatesEmailAndPassword(t *testing.T) {
	service := NewService(newFakeRepository(), &fakeIssuer{})
	tests := []RegisterInput{
		{Email: "", Password: "12345678"},
		{Email: "not-an-email", Password: "12345678"},
		{Email: "test@example.com", Password: "1234567"},
		{Email: "test@example.com", Password: "密码密码密码"},
	}

	for _, input := range tests {
		if _, err := service.Register(context.Background(), input); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("input %#v: expected ErrInvalidInput, got %v", input, err)
		}
	}
}

func TestRegisterAcceptsEightUnicodeCharacters(t *testing.T) {
	service := NewService(newFakeRepository(), &fakeIssuer{})
	_, err := service.Register(context.Background(), RegisterInput{
		Email: "test@example.com", Password: "密码密码密码密码",
	})
	if err != nil {
		t.Fatalf("expected eight Unicode characters to pass: %v", err)
	}
}

func TestRegisterRejectsPasswordBeyondBcryptLimit(t *testing.T) {
	service := NewService(newFakeRepository(), &fakeIssuer{})
	_, err := service.Register(context.Background(), RegisterInput{
		Email: "test@example.com", Password: strings.Repeat("a", 73),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestRegisterReportsDuplicateEmail(t *testing.T) {
	repository := newFakeRepository()
	repository.createErr = user.ErrEmailExists
	service := NewService(repository, &fakeIssuer{})

	_, err := service.Register(context.Background(), RegisterInput{Email: "test@example.com", Password: "12345678"})
	if !errors.Is(err, user.ErrEmailExists) {
		t.Fatalf("expected ErrEmailExists, got %v", err)
	}
}

func TestLoginReturnsTokenAndSafeUser(t *testing.T) {
	repository := newFakeRepository()
	hash, _ := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
	model := user.User{ID: uuid.New(), Email: "test@example.com", PasswordHash: string(hash), Role: "user"}
	repository.byEmail[model.Email] = model
	repository.byID[model.ID] = model
	issuer := &fakeIssuer{token: "signed-token"}
	service := NewService(repository, issuer)

	result, err := service.Login(context.Background(), LoginInput{Email: " TEST@example.com ", Password: "12345678"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if result.Token != "signed-token" || result.User.ID != model.ID || issuer.issued.ID != model.ID {
		t.Fatalf("unexpected login result: %#v", result)
	}
}

func TestLoginHidesWhetherEmailExists(t *testing.T) {
	repository := newFakeRepository()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	model := user.User{ID: uuid.New(), Email: "test@example.com", PasswordHash: string(hash), Role: "user"}
	repository.byEmail[model.Email] = model
	service := NewService(repository, &fakeIssuer{token: "token"})

	for _, input := range []LoginInput{
		{Email: "missing@example.com", Password: "12345678"},
		{Email: "test@example.com", Password: "wrong-password"},
	} {
		if _, err := service.Login(context.Background(), input); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("input %#v: expected ErrInvalidCredentials, got %v", input, err)
		}
	}
}

func TestCurrentUserReturnsPublicUser(t *testing.T) {
	repository := newFakeRepository()
	model := user.User{ID: uuid.New(), Email: "test@example.com", PasswordHash: "hidden", Role: "user"}
	repository.byID[model.ID] = model
	service := NewService(repository, &fakeIssuer{})

	result, err := service.CurrentUser(context.Background(), model.ID)
	if err != nil {
		t.Fatalf("current user: %v", err)
	}
	if result.ID != model.ID || result.Email != model.Email {
		t.Fatalf("unexpected user: %#v", result)
	}
}
