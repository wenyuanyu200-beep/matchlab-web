package match

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"matchlab/backend/internal/activity"
	"matchlab/backend/internal/questionnaire"
)

type serviceRepository struct {
	signals         UserSignals
	candidates      []activity.Activity
	persisted       []Recommendation
	persistedUser   uuid.UUID
	persistedFormID uuid.UUID
	saved           []SavedRecommendation
	listUser        uuid.UUID
}

func (r *serviceRepository) LoadSignals(_ context.Context, _ uuid.UUID) (UserSignals, error) {
	return r.signals, nil
}

func (r *serviceRepository) ListCandidates(_ context.Context, _ uuid.UUID) ([]activity.Activity, error) {
	return r.candidates, nil
}

func (r *serviceRepository) UpsertMatches(_ context.Context, userID, questionnaireID uuid.UUID, recommendations []Recommendation) error {
	r.persistedUser = userID
	r.persistedFormID = questionnaireID
	r.persisted = append([]Recommendation(nil), recommendations...)
	return nil
}

func (r *serviceRepository) ListMatches(_ context.Context, userID uuid.UUID) ([]SavedRecommendation, error) {
	r.listUser = userID
	return r.saved, nil
}

func TestServiceRecommendExcludesOwnActivityRanksAndPersists(t *testing.T) {
	userID := uuid.New()
	formID := uuid.New()
	repository := &serviceRepository{
		signals: UserSignals{QuestionnaireID: formID, Answers: questionnaire.Answers{Interests: questionnaire.StringList{"硬件"}}},
		candidates: []activity.Activity{
			{ID: uuid.New(), CreatorID: userID, Title: "本人活动", Tags: activity.StringList{"硬件"}, Status: activity.StatusRecruiting},
			{ID: uuid.New(), CreatorID: uuid.New(), Title: "软件活动", Tags: activity.StringList{"软件"}, Status: activity.StatusRecruiting},
			{ID: uuid.New(), CreatorID: uuid.New(), Title: "硬件活动", Tags: activity.StringList{"硬件"}, Status: activity.StatusRecruiting},
		},
	}
	service := NewService(repository)

	recommendations, err := service.Recommend(context.Background(), userID, " activity ", nil)

	if err != nil {
		t.Fatalf("recommend: %v", err)
	}
	if len(recommendations) != 2 || recommendations[0].Activity.Title != "硬件活动" {
		t.Fatalf("unexpected recommendations: %#v", recommendations)
	}
	if repository.persistedUser != userID || repository.persistedFormID != formID || len(repository.persisted) != 2 {
		t.Fatalf("persistence call mismatch: %#v", repository)
	}
}

func TestServiceRecommendDefaultsLimitToTen(t *testing.T) {
	repository := &serviceRepository{signals: UserSignals{QuestionnaireID: uuid.New()}}
	for i := 0; i < 12; i++ {
		repository.candidates = append(repository.candidates, activity.Activity{ID: uuid.New(), CreatorID: uuid.New(), CreatedAt: time.Now().Add(time.Duration(i) * time.Second)})
	}

	recommendations, err := NewService(repository).Recommend(context.Background(), uuid.New(), "activity", nil)

	if err != nil || len(recommendations) != 10 {
		t.Fatalf("recommendations=%d err=%v", len(recommendations), err)
	}
}

func TestServiceRecommendRejectsInvalidTargetAndLimit(t *testing.T) {
	service := NewService(&serviceRepository{})
	zero := 0
	tooMany := 51
	tests := []struct {
		target string
		limit  *int
		want   error
	}{
		{target: "user", want: ErrInvalidTarget},
		{target: "activity", limit: &zero, want: ErrInvalidLimit},
		{target: "activity", limit: &tooMany, want: ErrInvalidLimit},
	}
	for _, test := range tests {
		_, err := service.Recommend(context.Background(), uuid.New(), test.target, test.limit)
		if !errors.Is(err, test.want) {
			t.Fatalf("target=%s limit=%v err=%v, want %v", test.target, test.limit, err, test.want)
		}
	}
}

func TestServiceMyMatchesDelegatesUserID(t *testing.T) {
	userID := uuid.New()
	repository := &serviceRepository{saved: []SavedRecommendation{{ID: uuid.New()}}}

	matches, err := NewService(repository).MyMatches(context.Background(), userID)

	if err != nil || len(matches) != 1 || repository.listUser != userID {
		t.Fatalf("matches=%#v user=%s err=%v", matches, repository.listUser, err)
	}
}

var _ Repository = (*serviceRepository)(nil)
