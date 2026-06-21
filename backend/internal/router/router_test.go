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
	data, ok := body["data"].(map[string]any)
	if !ok || data["ok"] != true || data["message"] != "MatchLab API running" {
		t.Fatalf("unexpected response: %#v", body)
	}
}

func TestRouterReturnsJSONForUnknownRouteAndWrongMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := New(Dependencies{})
	tests := []struct {
		method string
		path   string
		status int
		code   string
	}{
		{http.MethodGet, "/api/does-not-exist", http.StatusNotFound, "not_found"},
		{http.MethodDelete, "/api/health", http.StatusMethodNotAllowed, "method_not_allowed"},
	}
	for _, test := range tests {
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, httptest.NewRequest(test.method, test.path, nil))
		if recorder.Code != test.status {
			t.Fatalf("%s %s status=%d body=%s", test.method, test.path, recorder.Code, recorder.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] != test.code || body["message"] == "" {
			t.Fatalf("%s %s response=%s err=%v", test.method, test.path, recorder.Body.String(), err)
		}
	}
}

func TestRouterAppliesConfiguredCORSOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := New(Dependencies{CORSAllowedOrigins: []string{"https://matchlab.example"}})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/health", nil)
	request.Header.Set("Origin", "https://matchlab.example")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent || recorder.Header().Get("Access-Control-Allow-Origin") != "https://matchlab.example" {
		t.Fatalf("status=%d headers=%v body=%s", recorder.Code, recorder.Header(), recorder.Body.String())
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

func TestAdminRoutesAreRegisteredAndProtected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const secret = "admin-router-test-secret"
	engine := New(Dependencies{JWTSecret: secret})
	adminToken, err := auth.NewTokenManager(secret).Issue(user.User{ID: uuid.New(), Role: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	userToken, err := auth.NewTokenManager(secret).Issue(user.User{ID: uuid.New(), Role: "user"})
	if err != nil {
		t.Fatal(err)
	}
	targetID := uuid.New().String()
	tests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/admin/stats", ""},
		{http.MethodGet, "/api/admin/users", ""},
		{http.MethodGet, "/api/admin/activities", ""},
		{http.MethodGet, "/api/admin/applications", ""},
		{http.MethodGet, "/api/admin/feedbacks", ""},
		{http.MethodPost, "/api/admin/users/" + targetID + "/role", `{"role":"admin"}`},
	}
	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			assertRouteStatus(t, engine, test.method, test.path, test.body, "", http.StatusUnauthorized)
			assertRouteStatus(t, engine, test.method, test.path, test.body, userToken, http.StatusForbidden)
			assertRouteStatus(t, engine, test.method, test.path, test.body, adminToken, http.StatusServiceUnavailable)
		})
	}
}

func assertRouteStatus(t *testing.T, engine http.Handler, method, path, body, token string, want int) {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	engine.ServeHTTP(recorder, request)
	if recorder.Code != want {
		t.Fatalf("status=%d want=%d body=%s", recorder.Code, want, recorder.Body.String())
	}
}
