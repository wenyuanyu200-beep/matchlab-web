package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"matchlab/backend/internal/auth"
	"matchlab/backend/internal/user"
)

func TestRequireAuthRejectsMissingOrMalformedHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, header := range []string{"", "Basic abc", "Bearer", "Bearer invalid"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/protected", nil)
		if header != "" {
			request.Header.Set("Authorization", header)
		}
		engine := gin.New()
		engine.GET("/protected", RequireAuth(auth.NewTokenManager("secret")), func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})

		engine.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("header %q: expected 401, got %d", header, recorder.Code)
		}
		var body map[string]any
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] != "invalid_token" {
			t.Fatalf("header %q: unexpected body %s", header, recorder.Body.String())
		}
	}
}

func TestRequireAuthStoresUserIDAndRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := auth.NewTokenManager("secret")
	model := user.User{ID: uuid.New(), Role: "admin"}
	token, err := manager.Issue(model)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	engine := gin.New()
	engine.GET("/protected", RequireAuth(manager), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.GetString(auth.ContextUserIDKey),
			"role":    c.GetString(auth.ContextRoleKey),
		})
	})

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["user_id"] != model.ID.String() || body["role"] != "admin" {
		t.Fatalf("unexpected context values: %#v", body)
	}
}
