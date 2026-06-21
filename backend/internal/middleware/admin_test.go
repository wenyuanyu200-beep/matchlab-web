package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"matchlab/backend/internal/auth"
)

func adminTestEngine(role *string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	if role != nil {
		engine.Use(func(c *gin.Context) {
			c.Set(auth.ContextRoleKey, *role)
			c.Next()
		})
	}
	engine.Use(RequireAdmin())
	engine.GET("/admin", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	return engine
}

func TestRequireAdminAllowsAdminRole(t *testing.T) {
	role := "admin"
	recorder := httptest.NewRecorder()
	adminTestEngine(&role).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestRequireAdminRejectsUserAndMissingRole(t *testing.T) {
	userRole := "user"
	for _, role := range []*string{&userRole, nil} {
		recorder := httptest.NewRecorder()
		adminTestEngine(role).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/admin", nil))
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("role=%v status=%d body=%s", role, recorder.Code, recorder.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] != "forbidden" {
			t.Fatalf("unexpected response: %v %s", err, recorder.Body.String())
		}
	}
}

func TestRequireAuthThenAdminRejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequireAuth(auth.NewTokenManager("test-secret")), RequireAdmin())
	engine.GET("/admin", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.Header.Set("Authorization", "Bearer invalid")

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
