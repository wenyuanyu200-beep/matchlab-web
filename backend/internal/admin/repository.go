package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"matchlab/backend/internal/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUnavailable = errors.New("admin repository unavailable")
	ErrNotFound    = errors.New("admin resource not found")
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (r *GormRepository) Stats(ctx context.Context) (Stats, error) {
	if r.db == nil {
		return Stats{}, ErrUnavailable
	}
	var stats Stats
	counts := []struct {
		table string
		where string
		value any
		dest  *int64
	}{
		{"users", "", nil, &stats.UsersCount},
		{"activities", "", nil, &stats.ActivitiesCount},
		{"applications", "", nil, &stats.ApplicationsCount},
		{"matches", "", nil, &stats.MatchesCount},
		{"questionnaires", "", nil, &stats.QuestionnairesCount},
		{"feedbacks", "", nil, &stats.FeedbacksCount},
		{"activities", "status = ?", "recruiting", &stats.RecruitingActivitiesCount},
		{"applications", "status = ?", "pending", &stats.PendingApplicationsCount},
		{"applications", "status = ?", "approved", &stats.ApprovedApplicationsCount},
	}
	for _, count := range counts {
		query := r.db.WithContext(ctx).Table(count.table)
		if count.where != "" {
			query = query.Where(count.where, count.value)
		}
		if err := query.Count(count.dest).Error; err != nil {
			return Stats{}, fmt.Errorf("count %s: %w", count.table, mapRepositoryError(err))
		}
	}
	return stats, nil
}

func (r *GormRepository) Users(ctx context.Context, filter UsersFilter) ([]AdminUser, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	users := []AdminUser{}
	query := r.db.WithContext(ctx).Table("users").
		Select("id, email, nickname, role, school, created_at")
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where("(email ILIKE ? OR nickname ILIKE ? OR school ILIKE ?)", like, like, like)
	}
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}
	err := query.Order("created_at DESC, id DESC").Limit(filter.Limit).Offset(filter.Offset).Scan(&users).Error
	if err != nil {
		return nil, fmt.Errorf("list users: %w", mapRepositoryError(err))
	}
	return users, nil
}

type activityRow struct {
	ID              uuid.UUID
	Title           string
	Type            string
	Status          string
	CreatorID       uuid.UUID
	CreatorNickname string
	CreatorSchool   string
	RequiredCount   int
	JoinedCount     int
	CreatedAt       time.Time
}

func (r *GormRepository) Activities(ctx context.Context, filter ActivitiesFilter) ([]AdminActivity, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	rows := []activityRow{}
	query := r.db.WithContext(ctx).Table("activities AS a").
		Select(`a.id, a.title, a.type, a.status, a.creator_id,
			u.nickname AS creator_nickname, u.school AS creator_school,
			a.required_count, a.joined_count, a.created_at`).
		Joins("JOIN users AS u ON u.id = a.creator_id")
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where("(a.title ILIKE ? OR a.description ILIKE ?)", like, like)
	}
	if filter.Type != "" {
		query = query.Where("a.type = ?", filter.Type)
	}
	if filter.Status != "" {
		query = query.Where("a.status = ?", filter.Status)
	}
	if err := query.Order("a.created_at DESC, a.id DESC").Limit(filter.Limit).Offset(filter.Offset).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list activities: %w", mapRepositoryError(err))
	}
	activities := make([]AdminActivity, 0, len(rows))
	for _, row := range rows {
		activities = append(activities, AdminActivity{
			ID: row.ID, Title: row.Title, Type: row.Type, Status: row.Status,
			CreatorID:     row.CreatorID,
			Creator:       PersonSummary{Nickname: row.CreatorNickname, School: row.CreatorSchool},
			RequiredCount: row.RequiredCount, JoinedCount: row.JoinedCount, CreatedAt: row.CreatedAt,
		})
	}
	return activities, nil
}

type applicationRow struct {
	ID                uuid.UUID
	ActivityID        uuid.UUID
	ActivityTitle     string
	ApplicantID       uuid.UUID
	ApplicantNickname string
	ApplicantSchool   string
	Reason            string
	MatchScore        int
	Status            string
	CreatedAt         time.Time
}

func (r *GormRepository) Applications(ctx context.Context, filter ApplicationsFilter) ([]AdminApplication, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	rows := []applicationRow{}
	query := r.db.WithContext(ctx).Table("applications AS ap").
		Select(`ap.id, ap.activity_id, a.title AS activity_title, ap.applicant_id,
			u.nickname AS applicant_nickname, u.school AS applicant_school,
			ap.reason, ap.match_score, ap.status, ap.created_at`).
		Joins("JOIN activities AS a ON a.id = ap.activity_id").
		Joins("JOIN users AS u ON u.id = ap.applicant_id")
	if filter.Status != "" {
		query = query.Where("ap.status = ?", filter.Status)
	}
	if err := query.Order("ap.created_at DESC, ap.id DESC").Limit(filter.Limit).Offset(filter.Offset).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list applications: %w", mapRepositoryError(err))
	}
	applications := make([]AdminApplication, 0, len(rows))
	for _, row := range rows {
		applications = append(applications, AdminApplication{
			ID: row.ID, ActivityID: row.ActivityID, ActivityTitle: row.ActivityTitle,
			ApplicantID: row.ApplicantID,
			Applicant:   PersonSummary{Nickname: row.ApplicantNickname, School: row.ApplicantSchool},
			Reason:      row.Reason, MatchScore: row.MatchScore, Status: row.Status, CreatedAt: row.CreatedAt,
		})
	}
	return applications, nil
}

func (r *GormRepository) Feedbacks(ctx context.Context, page Page) ([]Feedback, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	feedbacks := []Feedback{}
	err := r.db.WithContext(ctx).Table("feedbacks").
		Select("id, user_id, activity_id, match_id, rating, comment, created_at").
		Order("created_at DESC, id DESC").Limit(page.Limit).Offset(page.Offset).Scan(&feedbacks).Error
	if err != nil {
		return nil, fmt.Errorf("list feedbacks: %w", mapRepositoryError(err))
	}
	return feedbacks, nil
}

func (r *GormRepository) UpdateUserRole(ctx context.Context, targetID uuid.UUID, role string) (*AdminUser, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	var updated AdminUser
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Table("users").Where("id = ?", targetID).Update("role", role)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Table("users").Select("id, email, nickname, role, school, created_at").Where("id = ?", targetID).Take(&updated).Error
	})
	if err != nil {
		return nil, fmt.Errorf("update user role: %w", mapRepositoryError(err))
	}
	return &updated, nil
}

func adminUserFromModel(value user.User) AdminUser {
	return AdminUser{
		ID: value.ID, Email: value.Email, Nickname: value.Nickname,
		Role: value.Role, School: value.School, CreatedAt: value.CreatedAt,
	}
}

func mapRepositoryError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
