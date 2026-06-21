package user

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPublicOmitsPasswordHash(t *testing.T) {
	now := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)
	model := User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "secret-hash",
		Nickname:     "测试用户",
		Role:         "user",
		School:       "西南大学",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	public := model.Public()
	encoded, err := json.Marshal(public)
	if err != nil {
		t.Fatalf("marshal public user: %v", err)
	}

	if strings.Contains(string(encoded), "password") || strings.Contains(string(encoded), "secret-hash") {
		t.Fatalf("public user exposed password data: %s", encoded)
	}
	if public.ID != model.ID || public.Email != model.Email || public.Nickname != model.Nickname ||
		public.Role != model.Role || public.School != model.School ||
		!public.CreatedAt.Equal(now) || !public.UpdatedAt.Equal(now) {
		t.Fatalf("public user lost fields: %#v", public)
	}
}
