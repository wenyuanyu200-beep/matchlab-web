package match

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"matchlab/backend/internal/activity"
)

var (
	ErrInvalidTarget = errors.New("target_type must be activity")
	ErrInvalidLimit  = errors.New("limit must be between 1 and 50")
)

type Service interface {
	Recommend(ctx context.Context, userID uuid.UUID, targetType string, limit *int) ([]Recommendation, error)
	MyMatches(ctx context.Context, userID uuid.UUID) ([]SavedRecommendation, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) Recommend(ctx context.Context, userID uuid.UUID, targetType string, requestedLimit *int) ([]Recommendation, error) {
	if strings.ToLower(strings.TrimSpace(targetType)) != "activity" {
		return nil, ErrInvalidTarget
	}
	limit := 10
	if requestedLimit != nil {
		limit = *requestedLimit
	}
	if limit < 1 || limit > 50 {
		return nil, ErrInvalidLimit
	}
	signals, err := s.repository.LoadSignals(ctx, userID)
	if err != nil {
		return nil, err
	}
	candidates, err := s.repository.ListCandidates(ctx, userID)
	if err != nil {
		return nil, err
	}
	eligible := make([]activity.Activity, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.CreatorID == userID || (candidate.Status != "" && candidate.Status != activity.StatusRecruiting) {
			continue
		}
		eligible = append(eligible, candidate)
	}
	recommendations := RankActivities(signals, eligible, limit)
	if err := s.repository.UpsertMatches(ctx, userID, signals.QuestionnaireID, recommendations); err != nil {
		return nil, err
	}
	return recommendations, nil
}

func (s *service) MyMatches(ctx context.Context, userID uuid.UUID) ([]SavedRecommendation, error) {
	return s.repository.ListMatches(ctx, userID)
}
