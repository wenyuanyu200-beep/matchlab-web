package questionnaire

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrUnavailable     = errors.New("questionnaire repository unavailable")
	ErrProfileNotFound = errors.New("profile not found")
)

type Repository interface {
	Submit(ctx context.Context, userID uuid.UUID, mode string, answers Answers, generated GeneratedProfile) (*Questionnaire, *Profile, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*Profile, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Submit(ctx context.Context, userID uuid.UUID, mode string, answers Answers, generated GeneratedProfile) (*Questionnaire, *Profile, error) {
	if r.db == nil {
		return nil, nil, ErrUnavailable
	}
	var questionnaire Questionnaire
	var profile Profile
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var account struct {
			ID       uuid.UUID
			Nickname string
		}
		if err := tx.Table("users").Clauses(clause.Locking{Strength: "UPDATE"}).Select("id, nickname").Where("id = ?", userID).Take(&account).Error; err != nil {
			return fmt.Errorf("read user: %w", err)
		}

		var maxVersion int
		if err := tx.Model(&Questionnaire{}).Where("user_id = ?", userID).Select("COALESCE(MAX(version), 0)").Scan(&maxVersion).Error; err != nil {
			return fmt.Errorf("read questionnaire version: %w", err)
		}
		now := time.Now().UTC()
		questionnaire = Questionnaire{
			UserID:      userID,
			Mode:        mode,
			Version:     maxVersion + 1,
			Answers:     answers,
			Scores:      scoresMap(generated.Scores),
			Status:      "completed",
			CompletedAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := tx.Create(&questionnaire).Error; err != nil {
			return fmt.Errorf("create questionnaire: %w", err)
		}

		displayName := strings.TrimSpace(account.Nickname)
		if displayName == "" {
			displayName = "校园协作者"
		}
		candidate := Profile{
			UserID:      userID,
			DisplayName: displayName,
			ProfileType: generated.ProfileType,
			Tags:        generated.Tags,
			Scores:      generated.Scores,
			Summary:     generated.Summary,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"profile_type", "tags", "scores", "summary", "updated_at"}),
		}).Create(&candidate).Error; err != nil {
			return fmt.Errorf("upsert profile: %w", err)
		}
		if err := tx.Where("user_id = ?", userID).Take(&profile).Error; err != nil {
			return fmt.Errorf("read profile: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("submit questionnaire: %w", err)
	}
	return &questionnaire, &profile, nil
}

func (r *GormRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	var profile Profile
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Take(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return &profile, nil
}

func scoresMap(scores ProfileScores) JSONMap {
	return JSONMap{
		"interest_score":      scores.InterestScore,
		"skill_score":         scores.SkillScore,
		"time_score":          scores.TimeScore,
		"goal_score":          scores.GoalScore,
		"communication_score": scores.CommunicationScore,
	}
}
