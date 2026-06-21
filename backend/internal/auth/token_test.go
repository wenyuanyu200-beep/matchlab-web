package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"matchlab/backend/internal/user"
)

func TestTokenManagerIssuesAndParsesSevenDayToken(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	manager := NewTokenManager("test-secret")
	manager.now = func() time.Time { return now }
	model := user.User{ID: uuid.New(), Role: "admin"}

	token, err := manager.Issue(model)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	if claims.Subject != model.ID.String() || claims.Role != "admin" || claims.Issuer != "matchlab" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.Time.Equal(now.Add(7*24*time.Hour)) {
		t.Fatalf("unexpected expiry: %#v", claims.ExpiresAt)
	}
}

func TestTokenManagerRejectsExpiredToken(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	manager := NewTokenManager("test-secret")
	manager.now = func() time.Time { return now }
	token, err := manager.Issue(user.User{ID: uuid.New(), Role: "user"})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	manager.now = func() time.Time { return now.Add(8 * 24 * time.Hour) }

	if _, err := manager.Parse(token); err == nil {
		t.Fatal("expected expired token to fail")
	}
}

func TestTokenManagerRejectsTamperedToken(t *testing.T) {
	manager := NewTokenManager("test-secret")
	token, err := manager.Issue(user.User{ID: uuid.New(), Role: "user"})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	if _, err := manager.Parse(token + "tampered"); err == nil {
		t.Fatal("expected tampered token to fail")
	}
}

func TestTokenManagerRejectsNonHS256Algorithm(t *testing.T) {
	now := time.Now().UTC()
	unsigned := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{
		Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.NewString(),
			Issuer:    "matchlab",
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	})
	token, err := unsigned.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("create unsigned token: %v", err)
	}

	if _, err := NewTokenManager("test-secret").Parse(token); err == nil {
		t.Fatal("expected non-HS256 token to fail")
	}
}
