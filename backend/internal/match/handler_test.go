package match

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

	"matchlab/backend/internal/activity"
	"matchlab/backend/internal/auth"
	"matchlab/backend/internal/questionnaire"
)

type memoryRepository struct {
	signals       UserSignals
	signalsErr    error
	candidates    []activity.Activity
	persisted     []Recommendation
	persistedUser uuid.UUID
	saved         []SavedRecommendation
}

func (r *memoryRepository) LoadSignals(_ context.Context, _ uuid.UUID) (UserSignals, error) {
	return r.signals, r.signalsErr
}

func (r *memoryRepository) ListCandidates(_ context.Context, _ uuid.UUID) ([]activity.Activity, error) {
	return r.candidates, nil
}

func (r *memoryRepository) UpsertMatches(_ context.Context, userID uuid.UUID, _ uuid.UUID, recommendations []Recommendation) error {
	r.persistedUser = userID
	r.persisted = append([]Recommendation(nil), recommendations...)
	return nil
}

func (r *memoryRepository) ListMatches(_ context.Context, _ uuid.UUID) ([]SavedRecommendation, error) {
	return r.saved, nil
}

func matchEngine(repository Repository, userID uuid.UUID) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(auth.ContextUserIDKey, userID.String())
		c.Next()
	})
	handler := NewHandler(repository)
	engine.POST("/api/match/recommend", handler.Recommend)
	engine.GET("/api/me/matches", handler.CurrentMatches)
	return engine
}

func TestRecommendExcludesOwnActivitiesAndPersistsRankedResults(t *testing.T) {
	userID := uuid.New()
	repository := &memoryRepository{
		signals: UserSignals{QuestionnaireID: uuid.New(), Answers: questionnaire.Answers{Interests: questionnaire.StringList{"硬件"}}},
		candidates: []activity.Activity{
			{ID: uuid.New(), CreatorID: userID, Title: "本人活动", Tags: activity.StringList{"硬件"}, Status: activity.StatusRecruiting},
			{ID: uuid.New(), CreatorID: uuid.New(), Title: "软件活动", Tags: activity.StringList{"软件"}, Status: activity.StatusRecruiting},
			{ID: uuid.New(), CreatorID: uuid.New(), Title: "硬件活动", Tags: activity.StringList{"硬件"}, Status: activity.StatusRecruiting},
		},
	}
	engine := matchEngine(repository, userID)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/match/recommend", bytes.NewBufferString(`{"target_type":"activity","limit":10}`))
	request.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if repository.persistedUser != userID || len(repository.persisted) != 2 {
		t.Fatalf("persisted user=%s results=%d", repository.persistedUser, len(repository.persisted))
	}
	if repository.persisted[0].Activity.Title != "硬件活动" || repository.persisted[0].Score != 30 {
		t.Fatalf("unexpected ranking: %#v", repository.persisted)
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte("detail_scores")) || !bytes.Contains(recorder.Body.Bytes(), []byte("reason")) {
		t.Fatalf("response misses explanation: %s", recorder.Body.String())
	}
}

func TestRecommendDefaultsLimitAndRejectsInvalidRequests(t *testing.T) {
	userID := uuid.New()
	repository := &memoryRepository{signals: UserSignals{QuestionnaireID: uuid.New()}}
	for i := 0; i < 12; i++ {
		repository.candidates = append(repository.candidates, activity.Activity{ID: uuid.New(), CreatorID: uuid.New(), Title: "活动", CreatedAt: time.Now().Add(time.Duration(i) * time.Second)})
	}
	engine := matchEngine(repository, userID)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/match/recommend", bytes.NewBufferString(`{"target_type":"activity"}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || len(repository.persisted) != 10 {
		t.Fatalf("default limit: status=%d persisted=%d body=%s", recorder.Code, len(repository.persisted), recorder.Body.String())
	}

	for _, body := range []string{`{"target_type":"user"}`, `{"target_type":"activity","limit":0}`, `{"target_type":"activity","limit":51}`} {
		recorder = httptest.NewRecorder()
		request = httptest.NewRequest(http.MethodPost, "/api/match/recommend", bytes.NewBufferString(body))
		request.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("body=%s status=%d response=%s", body, recorder.Code, recorder.Body.String())
		}
	}
}

func TestRecommendRequiresProfile(t *testing.T) {
	repository := &memoryRepository{signalsErr: ErrProfileRequired}
	engine := matchEngine(repository, uuid.New())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/match/recommend", bytes.NewBufferString(`{"target_type":"activity"}`))
	request.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] != "profile_required" {
		t.Fatalf("unexpected body: %v %s", err, recorder.Body.String())
	}
}

func TestCurrentMatchesReturnsSavedActivityResults(t *testing.T) {
	repository := &memoryRepository{saved: []SavedRecommendation{{
		ID:               uuid.New(),
		Activity:         activity.Activity{ID: uuid.New(), Title: "智能硬件项目"},
		Score:            88,
		DetailScores:     DetailScores{Interest: 26, Skill: 22, Type: 20, Time: 8, Goal: 12},
		Reason:           "推荐原因：兴趣和技能与活动内容契合。",
		AlgorithmVersion: AlgorithmVersion,
	}}}
	engine := matchEngine(repository, uuid.New())
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/me/matches", nil))

	if recorder.Code != http.StatusOK || !bytes.Contains(recorder.Body.Bytes(), []byte("智能硬件项目")) || !bytes.Contains(recorder.Body.Bytes(), []byte(`"score":88`)) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

var _ Repository = (*memoryRepository)(nil)
