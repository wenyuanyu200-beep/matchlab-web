package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"matchlab/backend/internal/user"
)

var (
	// ErrInvalidInput means registration data does not meet the API contract.
	ErrInvalidInput = errors.New("invalid authentication input")
	// ErrInvalidCredentials intentionally covers both unknown email and wrong password.
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// RegisterInput contains account registration fields.
type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	School   string `json:"school"`
}

// LoginInput contains email/password credentials.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResult contains the access token and safe user data.
type LoginResult struct {
	Token string          `json:"token"`
	User  user.PublicUser `json:"user"`
}

// Service implements authentication business rules.
type Service struct {
	users  user.Repository
	tokens TokenIssuer
}

// NewService creates an authentication service.
func NewService(users user.Repository, tokens TokenIssuer) *Service {
	return &Service{users: users, tokens: tokens}
}

// Register validates and persists a new account.
func (s *Service) Register(ctx context.Context, input RegisterInput) (user.PublicUser, error) {
	email := normalizeEmail(input.Email)
	if !validEmail(email) || utf8.RuneCountInString(input.Password) < 8 || len([]byte(input.Password)) > 72 {
		return user.PublicUser{}, ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return user.PublicUser{}, fmt.Errorf("hash password: %w", err)
	}
	model := user.User{
		Email:        email,
		PasswordHash: string(hash),
		Nickname:     strings.TrimSpace(input.Nickname),
		Role:         "user",
		School:       strings.TrimSpace(input.School),
	}
	if err := s.users.Create(ctx, &model); err != nil {
		return user.PublicUser{}, err
	}
	return model.Public(), nil
}

// Login validates credentials and creates an access token.
func (s *Service) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	email := normalizeEmail(input.Email)
	if !validEmail(email) || input.Password == "" {
		return LoginResult{}, ErrInvalidCredentials
	}
	model, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(model.PasswordHash), []byte(input.Password)) != nil {
		return LoginResult{}, ErrInvalidCredentials
	}
	token, err := s.tokens.Issue(*model)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{Token: token, User: model.Public()}, nil
}

// CurrentUser loads the currently authenticated account.
func (s *Service) CurrentUser(ctx context.Context, id uuid.UUID) (user.PublicUser, error) {
	model, err := s.users.FindByID(ctx, id)
	if err != nil {
		return user.PublicUser{}, err
	}
	return model.Public(), nil
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validEmail(value string) bool {
	if value == "" {
		return false
	}
	parsed, err := mail.ParseAddress(value)
	return err == nil && parsed.Address == value
}
