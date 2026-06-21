package match

import (
	"context"
	"errors"
	"fmt"
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
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Take(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return UserSignals{}, ErrProfileRequired
		}
		return UserSignals{}, fmt.Errorf("load profile: %w", err)
	}
	var form questionnaire.Questionnaire
	if err := r.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, "completed").Order("version DESC, created_at DESC").Take(&form).Error; err != nil {
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
	if err := r.db.WithContext(ctx).
		Where("status = ? AND creator_id <> ?", activity.StatusRecruiting, userID).
		Order("created_at DESC").
		Find(&candidates).Error; err != nil {
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
			questionnaireReference := questionnaireID
			record := Record{
				UserID:           userID,
				ActivityID:       recommendation.Activity.ID,
				QuestionnaireID:  &questionnaireReference,
				Score:            recommendation.Score,
				Explanation:      Explanation{DetailScores: recommendation.DetailScores, Reason: recommendation.Reason},
				AlgorithmVersion: AlgorithmVersion,
				Status:           "recommended",
				CreatedAt:        now,
				UpdatedAt:        now,
			}
			if questionnaireID == uuid.Nil {
				record.QuestionnaireID = nil
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "user_id"}, {Name: "activity_id"}, {Name: "algorithm_version"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"questionnaire_id", "score", "explanation", "status", "updated_at",
				}),
			}).Create(&record).Error; err != nil {
				return fmt.Errorf("upsert match for activity %s: %w", recommendation.Activity.ID, err)
			}
		}
		return nil
	})
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
			Score:            record.Score,
			DetailScores:     record.Explanation.DetailScores,
			Reason:           record.Explanation.Reason,
			AlgorithmVersion: record.AlgorithmVersion,
			Status:           record.Status,
			CreatedAt:        record.CreatedAt,
			UpdatedAt:        record.UpdatedAt,
		})
	}
	return results, nil
}
