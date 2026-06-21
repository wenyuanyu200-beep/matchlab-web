package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func corsTestEngine(origins []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(CORS(origins))
	engine.GET("/api/resource", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true}})
	})
	return engine
}

func TestCORSAllowsConfiguredOriginAndPreflight(t *testing.T) {
	engine := corsTestEngine([]string{"https://matchlab.example", "http://139.224.119.187"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/resource", nil)
	request.Header.Set("Origin", "https://matchlab.example")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	request.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "https://matchlab.example" {
		t.Fatalf("allow origin=%q", got)
	}
	if !strings.Contains(recorder.Header().Get("Vary"), "Origin") {
		t.Fatalf("vary=%q", recorder.Header().Get("Vary"))
	}
	if !strings.Contains(recorder.Header().Get("Access-Control-Allow-Headers"), "Authorization") {
		t.Fatalf("allow headers=%q", recorder.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestCORSDoesNotAllowUnconfiguredOrigin(t *testing.T) {
	engine := corsTestEngine([]string{"https://matchlab.example"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	request.Header.Set("Origin", "https://attacker.example")

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected allow origin=%q", got)
	}
}
