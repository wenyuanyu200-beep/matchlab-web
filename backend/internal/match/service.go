package match

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"matchlab/backend/internal/activity"
)

var (
	ErrInvalidTarget = errors.New("target_type must be activity")
	ErrInvalidLimit  = errors.New("limit must be between 1 and 50")
)

type RecommendFailure struct {
	Stage   string
	Message string
	Err     error
}

func (e *RecommendFailure) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return e.Message + ": " + e.Err.Error()
}

func (e *RecommendFailure) Unwrap() error {
	return e.Err
}

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
		slog.Warn("match_recommend_failed",
			"stage", "input",
			"user_id", userID.String(),
			"activity_id", "",
			"target_type", targetType,
			"error", ErrInvalidTarget.Error(),
		)
		return nil, recommendFailure("input", ErrInvalidTarget.Error(), ErrInvalidTarget)
	}
	limit := 10
	if requestedLimit != nil {
		limit = *requestedLimit
	}
	if limit < 1 || limit > 50 {
		slog.Warn("match_recommend_failed",
			"stage", "input",
			"user_id", userID.String(),
			"activity_id", "",
			"limit", limit,
			"error", ErrInvalidLimit.Error(),
		)
		return nil, recommendFailure("input", ErrInvalidLimit.Error(), ErrInvalidLimit)
	}
	signals, err := s.repository.LoadSignals(ctx, userID)
	if err != nil {
		slog.Error("match_recommend_failed",
			"stage", "database",
			"user_id", userID.String(),
			"activity_id", "",
			"query", "load_signals",
			"error", err.Error(),
		)
		return nil, recommendFailure("database", recommendMessage(err), err)
	}
	slog.Info("match_recommend_database_result",
		"stage", "database",
		"user_id", userID.String(),
		"activity_id", "",
		"query", "load_signals",
		"profile_found", signals.Profile.ID != uuid.Nil,
		"questionnaire_id", signals.QuestionnaireID.String(),
		"questionnaire_found", signals.QuestionnaireID != uuid.Nil,
		"answers_interests_count", len(signals.Answers.Interests),
		"answers_skills_count", len(signals.Answers.Skills),
	)
	candidates, err := s.repository.ListCandidates(ctx, userID)
	if err != nil {
		slog.Error("match_recommend_failed",
			"stage", "database",
			"user_id", userID.String(),
			"activity_id", "",
			"query", "list_candidates",
			"error", err.Error(),
		)
		return nil, recommendFailure("database", "failed to load recommendation candidates", err)
	}
	eligible := make([]activity.Activity, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.CreatorID == userID || (candidate.Status != "" && candidate.Status != activity.StatusRecruiting) {
			continue
		}
		eligible = append(eligible, candidate)
	}
	activityIDs := make([]string, 0, len(eligible))
	for _, candidate := range eligible {
		activityIDs = append(activityIDs, candidate.ID.String())
	}
	slog.Info("match_recommend_candidate_result",
		"stage", "database",
		"user_id", userID.String(),
		"activity_id", strings.Join(activityIDs, ","),
		"candidate_count", len(candidates),
		"eligible_count", len(eligible),
		"candidates_nil", candidates == nil,
		"eligible_nil", eligible == nil,
	)
	slog.Info("match_recommend_km_diagnostics",
		"stage", "matching",
		"user_id", userID.String(),
		"activity_id", strings.Join(activityIDs, ","),
		"algorithm", "rules",
		"algorithm_version", AlgorithmVersion,
		"km_matrix_rows", matrixRows(eligible),
		"km_matrix_cols", len(eligible),
	)
	recommendations := RankActivities(signals, eligible, limit)
	if err := s.repository.UpsertMatches(ctx, userID, signals.QuestionnaireID, recommendations); err != nil {
		slog.Error("match_recommend_failed",
			"stage", "database",
			"user_id", userID.String(),
			"activity_id", strings.Join(recommendationActivityIDs(recommendations), ","),
			"query", "upsert_matches",
			"recommendation_count", len(recommendations),
			"error", err.Error(),
		)
		return nil, recommendFailure("database", "failed to save recommendation results", err)
	}
	slog.Info("match_recommend_success",
		"stage", "matching",
		"user_id", userID.String(),
		"activity_id", strings.Join(recommendationActivityIDs(recommendations), ","),
		"recommendation_count", len(recommendations),
		"recommendations_nil", recommendations == nil,
	)
	return recommendations, nil
}

func (s *service) MyMatches(ctx context.Context, userID uuid.UUID) ([]SavedRecommendation, error) {
	return s.repository.ListMatches(ctx, userID)
}

func recommendFailure(stage, message string, err error) error {
	return &RecommendFailure{Stage: stage, Message: message, Err: err}
}

func recommendMessage(err error) string {
	if errors.Is(err, ErrProfileRequired) {
		return "submit a questionnaire before requesting recommendations"
	}
	if errors.Is(err, ErrUnavailable) {
		return "match service unavailable"
	}
	return "database query failed during recommendation"
}

func matrixRows(eligible []activity.Activity) int {
	if len(eligible) == 0 {
		return 0
	}
	return 1
}

func recommendationActivityIDs(recommendations []Recommendation) []string {
	ids := make([]string, 0, len(recommendations))
	for _, recommendation := range recommendations {
		ids = append(ids, recommendation.Activity.ID.String())
	}
	return ids
}
