package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"matchlab/backend/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeService struct {
	usersFilter UsersFilter
	page        Page
	updateErr   error
}

func (f *fakeService) Stats(context.Context) (Stats, error) { return Stats{UsersCount: 2}, nil }
func (f *fakeService) Users(_ context.Context, filter UsersFilter) ([]AdminUser, error) {
	f.usersFilter = filter
	return []AdminUser{{ID: uuid.New(), Email: "safe@example.com", Role: "user"}}, nil
}
func (f *fakeService) Activities(context.Context, ActivitiesFilter) ([]AdminActivity, error) {
	return []AdminActivity{}, nil
}
func (f *fakeService) Applications(context.Context, ApplicationsFilter) ([]AdminApplication, error) {
	return []AdminApplication{}, nil
}
func (f *fakeService) Feedbacks(_ context.Context, page Page) ([]Feedback, error) {
	f.page = page
	return []Feedback{}, nil
}
func (f *fakeService) UpdateUserRole(context.Context, uuid.UUID, uuid.UUID, string) (*AdminUser, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return &AdminUser{ID: uuid.New(), Role: "admin"}, nil
}

func TestHandlerListsSafeUsersAndParsesFilters(t *testing.T) {
	service := &fakeService{}
	engine := handlerEngine(service, uuid.New())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/users?keyword=lab&role=admin&limit=10&offset=5", nil)

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.usersFilter.Keyword != "lab" || service.usersFilter.Role != "admin" || service.usersFilter.Limit != 10 || service.usersFilter.Offset != 5 {
		t.Fatalf("unexpected filter: %#v", service.usersFilter)
	}
	if strings.Contains(strings.ToLower(recorder.Body.String()), "password") {
		t.Fatalf("response contains password field: %s", recorder.Body.String())
	}
}

func TestHandlerReturnsFeedbacksAsArray(t *testing.T) {
	engine := handlerEngine(&fakeService{}, uuid.New())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/feedbacks", nil)

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var body struct {
		Data struct {
			Feedbacks []Feedback `json:"feedbacks"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.Feedbacks == nil || len(body.Data.Feedbacks) != 0 {
		t.Fatalf("feedbacks must be an empty array: %s", recorder.Body.String())
	}
}

func TestHandlerRejectsInvalidPaginationAndSelfDemotion(t *testing.T) {
	currentID := uuid.New()
	service := &fakeService{updateErr: ErrSelfDemotion}
	engine := handlerEngine(service, currentID)

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/users?limit=nope", nil))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid pagination status=%d body=%s", recorder.Code, recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/users/"+currentID.String()+"/role", strings.NewReader(`{"role":"user"}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("self-demotion status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func handlerEngine(service Service, currentID uuid.UUID) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(auth.ContextUserIDKey, currentID.String())
		c.Next()
	})
	handler := NewHandlerWithService(service)
	engine.GET("/users", handler.Users)
	engine.GET("/feedbacks", handler.Feedbacks)
	engine.POST("/users/:id/role", handler.UpdateUserRole)
	return engine
}
