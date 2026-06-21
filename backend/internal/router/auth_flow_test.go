package router

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"matchlab/backend/internal/user"
)

type memoryRepository struct {
	byEmail map[string]user.User
	byID    map[uuid.UUID]user.User
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{byEmail: map[string]user.User{}, byID: map[uuid.UUID]user.User{}}
}

func (r *memoryRepository) Create(_ context.Context, model *user.User) error {
	if _, exists := r.byEmail[model.Email]; exists {
		return user.ErrEmailExists
	}
	model.ID = uuid.New()
	model.CreatedAt = time.Now().UTC()
	model.UpdatedAt = model.CreatedAt
	r.byEmail[model.Email] = *model
	r.byID[model.ID] = *model
	return nil
}

func (r *memoryRepository) FindByEmail(_ context.Context, email string) (*user.User, error) {
	model, exists := r.byEmail[email]
	if !exists {
		return nil, user.ErrNotFound
	}
	return &model, nil
}

func (r *memoryRepository) FindByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	model, exists := r.byID[id]
	if !exists {
		return nil, user.ErrNotFound
	}
	return &model, nil
}

func jsonRequest(method, target, body string) *http.Request {
	request := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	return request
}

func TestAuthenticationFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := newMemoryRepository()
	engine := New(Dependencies{Users: repository, JWTSecret: "router-test-secret"})

	registerRecorder := httptest.NewRecorder()
	engine.ServeHTTP(registerRecorder, jsonRequest(http.MethodPost, "/api/auth/register", `{
		"email":"TEST@example.com","password":"12345678","nickname":"测试用户","school":"西南大学"
	}`))
	if registerRecorder.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d: %s", registerRecorder.Code, registerRecorder.Body.String())
	}
	stored := repository.byEmail["test@example.com"]
	if stored.PasswordHash == "12345678" || bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("12345678")) != nil {
		t.Fatal("register did not persist a bcrypt hash")
	}

	loginRecorder := httptest.NewRecorder()
	engine.ServeHTTP(loginRecorder, jsonRequest(http.MethodPost, "/api/auth/login", `{
		"email":"test@example.com","password":"12345678"
	}`))
	if loginRecorder.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d: %s", loginRecorder.Code, loginRecorder.Body.String())
	}
	var loginResponse struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginRecorder.Body.Bytes(), &loginResponse); err != nil || loginResponse.Data.Token == "" {
		t.Fatalf("login did not return token: %v, %s", err, loginRecorder.Body.String())
	}

	meRecorder := httptest.NewRecorder()
	meRequest := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	meRequest.Header.Set("Authorization", "Bearer "+loginResponse.Data.Token)
	engine.ServeHTTP(meRecorder, meRequest)
	if meRecorder.Code != http.StatusOK {
		t.Fatalf("me: expected 200, got %d: %s", meRecorder.Code, meRecorder.Body.String())
	}
	if bytes.Contains(meRecorder.Body.Bytes(), []byte("password")) || !bytes.Contains(meRecorder.Body.Bytes(), []byte("test@example.com")) {
		t.Fatalf("me returned unsafe or incomplete response: %s", meRecorder.Body.String())
	}
}

func TestAuthenticationRoutesReturn503WithoutDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := New(Dependencies{JWTSecret: "router-test-secret"})
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, jsonRequest(http.MethodPost, "/api/auth/register", `{
		"email":"test@example.com","password":"12345678"
	}`))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", recorder.Code, recorder.Body.String())
	}
}
