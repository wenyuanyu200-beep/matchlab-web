package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"matchlab/backend/internal/user"
)

const tokenLifetime = 7 * 24 * time.Hour

// Claims are the authenticated values carried by MatchLab access tokens.
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// TokenIssuer creates access tokens.
type TokenIssuer interface {
	Issue(model user.User) (string, error)
}

// TokenParser validates access tokens.
type TokenParser interface {
	Parse(token string) (*Claims, error)
}

// TokenManager issues and validates HMAC SHA-256 JWTs.
type TokenManager struct {
	secret []byte
	now    func() time.Time
}

// NewTokenManager creates a token manager using the supplied secret.
func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{secret: []byte(secret), now: time.Now}
}

// Issue creates a seven-day access token for a user.
func (m *TokenManager) Issue(model user.User) (string, error) {
	now := m.now().UTC()
	claims := Claims{
		Role: model.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   model.ID.String(),
			Issuer:    "matchlab",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenLifetime)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// Parse validates an access token and returns its claims.
func (m *TokenManager) Parse(raw string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		raw,
		claims,
		func(_ *jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer("matchlab"),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return nil, fmt.Errorf("invalid access token subject: %w", err)
	}
	if claims.Role != "user" && claims.Role != "admin" {
		return nil, fmt.Errorf("invalid access token role")
	}
	return claims, nil
}
