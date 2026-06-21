package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"matchlab/backend/internal/user"
)

func performHandlerRequest(method, target, body string, handler gin.HandlerFunc, setup ...func(*gin.Context)) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(method, target, bytes.NewBufferString(body))
	context.Request.Header.Set("Content-Type", "application/json")
	for _, configure := range setup {
		configure(context)
	}
	handler(context)
	return recorder
}

func TestRegisterHandlerReturnsWrappedSafeUser(t *testing.T) {
	repository := newFakeRepository()
	handler := NewHandler(NewService(repository, &fakeIssuer{}))
	recorder := performHandlerRequest(http.MethodPost, "/api/auth/register", `{
		"email":"test@example.com","password":"12345678","nickname":"测试用户","school":"西南大学"
	}`, handler.Register)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "password") {
		t.Fatalf("response exposed password data: %s", recorder.Body.String())
	}
	var response struct {
		Data struct {
			User user.PublicUser `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.User.Email != "test@example.com" || response.Data.User.Role != "user" {
		t.Fatalf("unexpected user: %#v", response.Data.User)
	}
}

func TestRegisterHandlerMapsDuplicateAndUnavailable(t *testing.T) {
	for _, test := range []struct {
		name       string
		repository user.Repository
		status     int
		code       string
	}{
		{name: "duplicate", repository: &fakeRepository{byEmail: map[string]user.User{}, byID: map[uuid.UUID]user.User{}, createErr: user.ErrEmailExists}, status: http.StatusConflict, code: "email_exists"},
		{name: "unavailable", repository: user.NewGormRepository(nil), status: http.StatusServiceUnavailable, code: "service_unavailable"},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler := NewHandler(NewService(test.repository, &fakeIssuer{}))
			recorder := performHandlerRequest(http.MethodPost, "/api/auth/register", `{"email":"test@example.com","password":"12345678"}`, handler.Register)
			if recorder.Code != test.status || !strings.Contains(recorder.Body.String(), `"error":"`+test.code+`"`) {
				t.Fatalf("unexpected response %d: %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestLoginHandlerReturns401ForInvalidCredentials(t *testing.T) {
	handler := NewHandler(NewService(newFakeRepository(), &fakeIssuer{}))
	recorder := performHandlerRequest(http.MethodPost, "/api/auth/login", `{"email":"missing@example.com","password":"12345678"}`, handler.Login)

	if recorder.Code != http.StatusUnauthorized || !strings.Contains(recorder.Body.String(), `"error":"invalid_credentials"`) {
		t.Fatalf("unexpected response %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestLoginHandlerReturnsTokenEnvelope(t *testing.T) {
	repository := newFakeRepository()
	hash, _ := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
	model := user.User{ID: uuid.New(), Email: "test@example.com", PasswordHash: string(hash), Role: "user"}
	repository.byEmail[model.Email] = model
	repository.byID[model.ID] = model
	handler := NewHandler(NewService(repository, &fakeIssuer{token: "signed-token"}))
	recorder := performHandlerRequest(http.MethodPost, "/api/auth/login", `{"email":"test@example.com","password":"12345678"}`, handler.Login)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"token":"signed-token"`) {
		t.Fatalf("unexpected response %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestMeHandlerReturnsCurrentUser(t *testing.T) {
	repository := newFakeRepository()
	model := user.User{ID: uuid.New(), Email: "test@example.com", PasswordHash: "hidden", Role: "user"}
	repository.byID[model.ID] = model
	handler := NewHandler(NewService(repository, &fakeIssuer{}))
	recorder := performHandlerRequest(http.MethodGet, "/api/me", "", handler.Me, func(c *gin.Context) {
		c.Set(ContextUserIDKey, model.ID.String())
	})

	if recorder.Code != http.StatusOK || strings.Contains(recorder.Body.String(), "hidden") {
		t.Fatalf("unexpected response %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestMeHandlerMapsDeletedUserTo404(t *testing.T) {
	handler := NewHandler(NewService(newFakeRepository(), &fakeIssuer{}))
	recorder := performHandlerRequest(http.MethodGet, "/api/me", "", handler.Me, func(c *gin.Context) {
		c.Set(ContextUserIDKey, uuid.NewString())
	})

	if recorder.Code != http.StatusNotFound || !strings.Contains(recorder.Body.String(), `"error":"user_not_found"`) {
		t.Fatalf("unexpected response %d: %s", recorder.Code, recorder.Body.String())
	}
}
