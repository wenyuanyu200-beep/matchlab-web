package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"matchlab/backend/internal/auth"
	"matchlab/backend/internal/user"
)

func TestHealthRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)

	New(Dependencies{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["ok"] != true || body["message"] != "MatchLab API running" {
		t.Fatalf("unexpected response: %#v", body)
	}
}

func TestQuestionnaireAndMatchRoutesAreRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const secret = "router-test-secret"
	token, err := auth.NewTokenManager(secret).Issue(user.User{ID: uuid.New(), Role: "user"})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	engine := New(Dependencies{JWTSecret: secret})
	tests := []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodPost, path: "/api/questionnaires", body: `{"mode":"activity","answers":{}}`},
		{method: http.MethodGet, path: "/api/me/profile"},
		{method: http.MethodPost, path: "/api/match/recommend", body: `{"target_type":"activity"}`},
		{method: http.MethodGet, path: "/api/me/matches"},
	}
	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.path, bytes.NewBufferString(test.body))
			request.Header.Set("Authorization", "Bearer "+token)
			request.Header.Set("Content-Type", "application/json")

			engine.ServeHTTP(recorder, request)

			if recorder.Code == http.StatusNotFound {
				t.Fatalf("route returned 404: %s", recorder.Body.String())
			}
			if recorder.Code != http.StatusServiceUnavailable {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestRecommendationRoutesRequireAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := New(Dependencies{JWTSecret: "router-test-secret"})
	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/questionnaires"},
		{method: http.MethodGet, path: "/api/me/profile"},
		{method: http.MethodPost, path: "/api/match/recommend"},
		{method: http.MethodGet, path: "/api/me/matches"},
	}
	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.path, bytes.NewBufferString(`{}`))
			request.Header.Set("Content-Type", "application/json")

			engine.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}
