package questionnaire

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var ErrInvalidMode = errors.New("mode must be activity")

type Service interface {
	Submit(ctx context.Context, userID uuid.UUID, mode string, answers Answers) (*Questionnaire, *Profile, error)
	Profile(ctx context.Context, userID uuid.UUID) (*Profile, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) Submit(ctx context.Context, userID uuid.UUID, mode string, answers Answers) (*Questionnaire, *Profile, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != "activity" {
		return nil, nil, ErrInvalidMode
	}
	answers = normalizeAnswers(answers)
	generated := GenerateProfile(mode, answers)
	return s.repository.Submit(ctx, userID, mode, answers, generated)
}

func (s *service) Profile(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	return s.repository.GetProfile(ctx, userID)
}
