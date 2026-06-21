package questionnaire

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

	"matchlab/backend/internal/auth"
)

type memoryRepository struct {
	questionnaires []Questionnaire
	profiles       map[uuid.UUID]Profile
	lastUserID     uuid.UUID
	getProfileErr  error
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{profiles: make(map[uuid.UUID]Profile)}
}

func (r *memoryRepository) Submit(_ context.Context, userID uuid.UUID, mode string, answers Answers, generated GeneratedProfile) (*Questionnaire, *Profile, error) {
	r.lastUserID = userID
	now := time.Now().UTC()
	questionnaire := Questionnaire{ID: uuid.New(), UserID: userID, Mode: mode, Version: len(r.questionnaires) + 1, Answers: answers, Status: "completed", CompletedAt: now, CreatedAt: now, UpdatedAt: now}
	r.questionnaires = append(r.questionnaires, questionnaire)
	profile, exists := r.profiles[userID]
	if !exists {
		profile = Profile{ID: uuid.New(), UserID: userID, CreatedAt: now}
	}
	profile.ProfileType = generated.ProfileType
	profile.Tags = generated.Tags
	profile.Scores = generated.Scores
	profile.Summary = generated.Summary
	profile.UpdatedAt = now
	r.profiles[userID] = profile
	return &questionnaire, &profile, nil
}

func (r *memoryRepository) GetProfile(_ context.Context, userID uuid.UUID) (*Profile, error) {
	if r.getProfileErr != nil {
		return nil, r.getProfileErr
	}
	profile, ok := r.profiles[userID]
	if !ok {
		return nil, ErrProfileNotFound
	}
	return &profile, nil
}

func questionnaireEngine(repository Repository, userID uuid.UUID) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(auth.ContextUserIDKey, userID.String())
		c.Next()
	})
	handler := NewHandler(repository)
	engine.POST("/api/questionnaires", handler.Submit)
	engine.GET("/api/me/profile", handler.Profile)
	return engine
}

func TestSubmitUsesAuthenticatedUserAndReturnsQuestionnaireAndProfile(t *testing.T) {
	userID := uuid.New()
	repository := newMemoryRepository()
	engine := questionnaireEngine(repository, userID)
	body := `{"mode":"activity","answers":{"interests":["电赛","STM32"],"skills":["嵌入式"],"available_time":"周末下午","activity_types":["competition"],"goal":"参加比赛","communication_style":"稳定沟通"}}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/questionnaires", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if repository.lastUserID != userID {
		t.Fatalf("repository user=%s, want %s", repository.lastUserID, userID)
	}
	var response struct {
		Data struct {
			Questionnaire Questionnaire `json:"questionnaire"`
			Profile       Profile       `json:"profile"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.Questionnaire.UserID != userID || response.Data.Profile.UserID != userID {
		t.Fatalf("unexpected response: %#v", response.Data)
	}
}

func TestSubmitUpdatesExistingProfile(t *testing.T) {
	userID := uuid.New()
	repository := newMemoryRepository()
	engine := questionnaireEngine(repository, userID)
	for _, interest := range []string{"电赛", "机器人"} {
		recorder := httptest.NewRecorder()
		body := `{"mode":"activity","answers":{"interests":["` + interest + `"]}}`
		request := httptest.NewRequest(http.MethodPost, "/api/questionnaires", bytes.NewBufferString(body))
		request.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusCreated {
			t.Fatalf("submit %q: status=%d body=%s", interest, recorder.Code, recorder.Body.String())
		}
	}
	if len(repository.profiles) != 1 || len(repository.questionnaires) != 2 {
		t.Fatalf("profiles=%d questionnaires=%d", len(repository.profiles), len(repository.questionnaires))
	}
	if got := repository.profiles[userID].Tags; len(got) != 1 || got[0] != "机器人" {
		t.Fatalf("profile was not updated: %v", got)
	}
}

func TestCurrentProfileReturnsProfile(t *testing.T) {
	userID := uuid.New()
	repository := newMemoryRepository()
	repository.profiles[userID] = Profile{ID: uuid.New(), UserID: userID, ProfileType: "activity", Tags: StringList{"硬件"}, Summary: "校园项目协作"}
	engine := questionnaireEngine(repository, userID)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/me/profile", nil))

	if recorder.Code != http.StatusOK || !bytes.Contains(recorder.Body.Bytes(), []byte("校园项目协作")) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSubmitRejectsInvalidBody(t *testing.T) {
	engine := questionnaireEngine(newMemoryRepository(), uuid.New())
	for _, body := range []string{`{`, `{"mode":"","answers":{}}`, `{"mode":"team","answers":{}}`} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/questionnaires", bytes.NewBufferString(body))
		request.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("body=%s status=%d response=%s", body, recorder.Code, recorder.Body.String())
		}
	}
}

func TestCurrentProfileMapsNotFound(t *testing.T) {
	repository := newMemoryRepository()
	repository.getProfileErr = ErrProfileNotFound
	engine := questionnaireEngine(repository, uuid.New())
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/me/profile", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] != "profile_not_found" {
		t.Fatalf("unexpected body: %v, %s", err, recorder.Body.String())
	}
}

func TestCurrentProfileMapsUnavailable(t *testing.T) {
	repository := newMemoryRepository()
	repository.getProfileErr = ErrUnavailable
	engine := questionnaireEngine(repository, uuid.New())
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/me/profile", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

var _ Repository = (*memoryRepository)(nil)
