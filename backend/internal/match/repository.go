package match

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"matchlab/backend/internal/activity"
	"matchlab/backend/internal/questionnaire"
)

var (
	ErrUnavailable     = errors.New("match repository unavailable")
	ErrProfileRequired = errors.New("profile required")
)

type Repository interface {
	LoadSignals(ctx context.Context, userID uuid.UUID) (UserSignals, error)
	ListCandidates(ctx context.Context, userID uuid.UUID) ([]activity.Activity, error)
	UpsertMatches(ctx context.Context, userID, questionnaireID uuid.UUID, recommendations []Recommendation) error
	ListMatches(ctx context.Context, userID uuid.UUID) ([]SavedRecommendation, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) LoadSignals(ctx context.Context, userID uuid.UUID) (UserSignals, error) {
	if r.db == nil {
		return UserSignals{}, ErrUnavailable
	}
	var profile questionnaire.Profile
	profileResult := r.db.WithContext(ctx).Where("user_id = ?", userID).Take(&profile)
	slog.Info("match_recommend_db_query",
		"stage", "database",
		"query", "load_profile",
		"user_id", userID.String(),
		"activity_id", "",
		"rows_affected", profileResult.RowsAffected,
		"found", profileResult.Error == nil,
	)
	if err := profileResult.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return UserSignals{}, ErrProfileRequired
		}
		return UserSignals{}, fmt.Errorf("load profile: %w", err)
	}
	var form questionnaire.Questionnaire
	formResult := r.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, "completed").Order("version DESC, created_at DESC").Take(&form)
	slog.Info("match_recommend_db_query",
		"stage", "database",
		"query", "load_questionnaire",
		"user_id", userID.String(),
		"activity_id", "",
		"rows_affected", formResult.RowsAffected,
		"found", formResult.Error == nil,
	)
	if err := formResult.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return UserSignals{}, ErrProfileRequired
		}
		return UserSignals{}, fmt.Errorf("load questionnaire: %w", err)
	}
	return UserSignals{Profile: profile, QuestionnaireID: form.ID, Answers: form.Answers}, nil
}

func (r *GormRepository) ListCandidates(ctx context.Context, userID uuid.UUID) ([]activity.Activity, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	var candidates []activity.Activity
	result := r.db.WithContext(ctx).
		Where("status = ? AND creator_id <> ?", activity.StatusRecruiting, userID).
		Order("created_at DESC").
		Find(&candidates)
	slog.Info("match_recommend_db_query",
		"stage", "database",
		"query", "list_candidates",
		"user_id", userID.String(),
		"activity_id", "",
		"rows_affected", result.RowsAffected,
		"found", result.Error == nil,
	)
	if err := result.Error; err != nil {
		return nil, fmt.Errorf("list recommendation candidates: %w", err)
	}
	return candidates, nil
}

func (r *GormRepository) UpsertMatches(ctx context.Context, userID, questionnaireID uuid.UUID, recommendations []Recommendation) error {
	if r.db == nil {
		return ErrUnavailable
	}
	if len(recommendations) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		for _, recommendation := range recommendations {
			record := newRecord(userID, questionnaireID, recommendation, now)
			result := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "user_id"}, {Name: "activity_id"}, {Name: "algorithm_version"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"target_id", "target_type", "questionnaire_id", "algorithm", "score", "detail_scores", "reason", "status", "updated_at",
				}),
			}).Create(&record)
			slog.Info("match_recommend_db_query",
				"stage", "database",
				"query", "upsert_match",
				"user_id", userID.String(),
				"activity_id", recommendation.Activity.ID.String(),
				"rows_affected", result.RowsAffected,
				"score", recommendation.Score,
				"algorithm_version", AlgorithmVersion,
				"saved", result.Error == nil,
			)
			if err := result.Error; err != nil {
				return fmt.Errorf("upsert match for activity %s: %w", recommendation.Activity.ID, err)
			}
		}
		return nil
	})
}

func newRecord(userID, questionnaireID uuid.UUID, recommendation Recommendation, now time.Time) Record {
	questionnaireReference := questionnaireID
	record := Record{
		UserID:           userID,
		ActivityID:       recommendation.Activity.ID,
		TargetID:         recommendation.Activity.ID,
		TargetType:       "activity",
		QuestionnaireID:  &questionnaireReference,
		Algorithm:        "rules",
		AlgorithmVersion: AlgorithmVersion,
		Score:            float64(recommendation.Score),
		DetailScores:     recommendation.DetailScores,
		Reason:           recommendation.Reason,
		Status:           "recommended",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if questionnaireID == uuid.Nil {
		record.QuestionnaireID = nil
	}
	return record
}

func (r *GormRepository) ListMatches(ctx context.Context, userID uuid.UUID) ([]SavedRecommendation, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	var records []Record
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND algorithm_version = ?", userID, AlgorithmVersion).
		Order("score DESC, updated_at DESC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list matches: %w", err)
	}
	if len(records) == 0 {
		return []SavedRecommendation{}, nil
	}
	ids := make([]uuid.UUID, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ActivityID)
	}
	var activities []activity.Activity
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("load matched activities: %w", err)
	}
	byID := make(map[uuid.UUID]activity.Activity, len(activities))
	for _, candidate := range activities {
		byID[candidate.ID] = candidate
	}
	results := make([]SavedRecommendation, 0, len(records))
	for _, record := range records {
		candidate, exists := byID[record.ActivityID]
		if !exists {
			continue
		}
		results = append(results, SavedRecommendation{
			ID:               record.ID,
			Activity:         candidate,
			Score:            int(math.Round(record.Score)),
			DetailScores:     record.DetailScores,
			Reason:           record.Reason,
			AlgorithmVersion: record.AlgorithmVersion,
			Status:           record.Status,
			CreatedAt:        record.CreatedAt,
			UpdatedAt:        record.UpdatedAt,
		})
	}
	return results, nil
}
