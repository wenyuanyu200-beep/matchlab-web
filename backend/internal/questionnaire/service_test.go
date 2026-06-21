package questionnaire

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type serviceRepository struct {
	submitCalls int
	lastUserID  uuid.UUID
	lastMode    string
	lastAnswers Answers
	generated   GeneratedProfile
	profile     *Profile
}

func (r *serviceRepository) Submit(_ context.Context, userID uuid.UUID, mode string, answers Answers, generated GeneratedProfile) (*Questionnaire, *Profile, error) {
	r.submitCalls++
	r.lastUserID = userID
	r.lastMode = mode
	r.lastAnswers = answers
	r.generated = generated
	now := time.Now().UTC()
	questionnaire := &Questionnaire{ID: uuid.New(), UserID: userID, Mode: mode, Answers: answers, CreatedAt: now}
	profile := &Profile{ID: uuid.New(), UserID: userID, ProfileType: generated.ProfileType, Tags: generated.Tags, Scores: generated.Scores, Summary: generated.Summary, CreatedAt: now}
	r.profile = profile
	return questionnaire, profile, nil
}

func (r *serviceRepository) GetProfile(_ context.Context, _ uuid.UUID) (*Profile, error) {
	if r.profile == nil {
		return nil, ErrProfileNotFound
	}
	return r.profile, nil
}

func TestServiceSubmitGeneratesProfileAndPersists(t *testing.T) {
	repository := &serviceRepository{}
	service := NewService(repository)
	userID := uuid.New()

	questionnaire, profile, err := service.Submit(context.Background(), userID, " activity ", Answers{
		Interests: StringList{" 硬件 ", "硬件"},
		Skills:    StringList{"嵌入式"},
	})

	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if repository.submitCalls != 1 || repository.lastUserID != userID || repository.lastMode != "activity" {
		t.Fatalf("unexpected repository call: %#v", repository)
	}
	if questionnaire.Mode != "activity" || len(profile.Tags) != 2 || profile.Tags[0] != "硬件" {
		t.Fatalf("questionnaire=%#v profile=%#v", questionnaire, profile)
	}
}

func TestServiceSubmitRejectsUnsupportedMode(t *testing.T) {
	repository := &serviceRepository{}
	service := NewService(repository)

	_, _, err := service.Submit(context.Background(), uuid.New(), "team", Answers{})

	if !errors.Is(err, ErrInvalidMode) || repository.submitCalls != 0 {
		t.Fatalf("err=%v calls=%d", err, repository.submitCalls)
	}
}

func TestServiceProfileDelegatesToRepository(t *testing.T) {
	repository := &serviceRepository{profile: &Profile{ID: uuid.New(), UserID: uuid.New()}}
	service := NewService(repository)

	profile, err := service.Profile(context.Background(), repository.profile.UserID)

	if err != nil || profile.ID != repository.profile.ID {
		t.Fatalf("profile=%#v err=%v", profile, err)
	}
}

var _ Repository = (*serviceRepository)(nil)
