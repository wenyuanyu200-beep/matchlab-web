package activity

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrUnavailable         = errors.New("activity repository unavailable")
	ErrNotFound            = errors.New("activity not found")
	ErrApplicationNotFound = errors.New("application not found")
	ErrDuplicateApply      = errors.New("duplicate application")
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidState        = errors.New("invalid state")
)

type Repository interface {
	CreateActivity(ctx context.Context, model *Activity) error
	ListActivities(ctx context.Context, filter ListFilter) ([]ActivityWithCreator, error)
	GetActivity(ctx context.Context, id uuid.UUID) (*ActivityWithCreator, error)
	ListMyActivities(ctx context.Context, creatorID uuid.UUID) ([]Activity, error)
	CreateApplication(ctx context.Context, model *Application) error
	ListMyApplications(ctx context.Context, applicantID uuid.UUID) ([]ApplicationWithActivity, error)
	ListActivityApplications(ctx context.Context, activityID, requesterID uuid.UUID) ([]ApplicationWithApplicant, error)
	ApproveApplication(ctx context.Context, applicationID, requesterID uuid.UUID) (*Application, error)
	RejectApplication(ctx context.Context, applicationID, requesterID uuid.UUID) (*Application, error)
}

type ListFilter struct {
	Type    string
	Status  string
	Keyword string
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) CreateActivity(ctx context.Context, model *Activity) error {
	if r.db == nil {
		return ErrUnavailable
	}
	return mapError(r.db.WithContext(ctx).Create(model).Error)
}

func (r *GormRepository) ListActivities(ctx context.Context, filter ListFilter) ([]ActivityWithCreator, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	status := filter.Status
	if status == "" {
		status = StatusRecruiting
	}
	var rows []activityCreatorRow
	query := r.db.WithContext(ctx).Table("activities AS a").
		Select(activityCreatorSelect()).
		Joins("JOIN users AS u ON u.id = a.creator_id").
		Where("a.status = ?", status)
	if filter.Type != "" {
		query = query.Where("a.\"type\" = ?", filter.Type)
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where("(a.title ILIKE ? OR a.description ILIKE ?)", like, like)
	}
	if err := query.Order("a.created_at DESC").Scan(&rows).Error; err != nil {
		return nil, mapError(err)
	}
	return rowsToActivities(rows), nil
}

func (r *GormRepository) GetActivity(ctx context.Context, id uuid.UUID) (*ActivityWithCreator, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	var row activityCreatorRow
	err := r.db.WithContext(ctx).Table("activities AS a").
		Select(activityCreatorSelect()).
		Joins("JOIN users AS u ON u.id = a.creator_id").
		Where("a.id = ?", id).
		First(&row).Error
	if err != nil {
		return nil, mapError(err)
	}
	result := row.toActivity()
	return &result, nil
}

func (r *GormRepository) ListMyActivities(ctx context.Context, creatorID uuid.UUID) ([]Activity, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	var activities []Activity
	err := r.db.WithContext(ctx).Where("creator_id = ?", creatorID).Order("created_at DESC").Find(&activities).Error
	return activities, mapError(err)
}

func (r *GormRepository) CreateApplication(ctx context.Context, model *Application) error {
	if r.db == nil {
		return ErrUnavailable
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var act Activity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&act, "id = ?", model.ActivityID).Error; err != nil {
			return mapError(err)
		}
		if act.CreatorID == model.ApplicantID {
			return ErrForbidden
		}
		if act.Status != StatusRecruiting {
			return ErrInvalidState
		}
		return mapError(tx.Create(model).Error)
	})
}

func (r *GormRepository) ListMyApplications(ctx context.Context, applicantID uuid.UUID) ([]ApplicationWithActivity, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	var rows []applicationActivityRow
	err := r.db.WithContext(ctx).Table("applications AS ap").
		Select("ap.id, ap.activity_id, ap.applicant_id, ap.reason, ap.match_score, ap.status, ap.created_at, ap.updated_at, a.title AS activity_title, a.status AS activity_status").
		Joins("JOIN activities AS a ON a.id = ap.activity_id").
		Where("ap.applicant_id = ?", applicantID).
		Order("ap.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]ApplicationWithActivity, 0, len(rows))
	for _, row := range rows {
		out = append(out, ApplicationWithActivity{Application: row.Application, ActivityTitle: row.ActivityTitle, ActivityStatus: row.ActivityStatus})
	}
	return out, nil
}

func (r *GormRepository) ListActivityApplications(ctx context.Context, activityID, requesterID uuid.UUID) ([]ApplicationWithApplicant, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	var act Activity
	if err := r.db.WithContext(ctx).First(&act, "id = ?", activityID).Error; err != nil {
		return nil, mapError(err)
	}
	if act.CreatorID != requesterID {
		return nil, ErrForbidden
	}
	var rows []applicationApplicantRow
	err := r.db.WithContext(ctx).Table("applications AS ap").
		Select("ap.id, ap.activity_id, ap.applicant_id, ap.reason, ap.match_score, ap.status, ap.created_at, ap.updated_at, u.nickname AS applicant_nickname, u.school AS applicant_school").
		Joins("JOIN users AS u ON u.id = ap.applicant_id").
		Where("ap.activity_id = ?", activityID).
		Order("ap.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]ApplicationWithApplicant, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toApplication())
	}
	return out, nil
}

func (r *GormRepository) ApproveApplication(ctx context.Context, applicationID, requesterID uuid.UUID) (*Application, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	var updated Application
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var app Application
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&app, "id = ?", applicationID).Error; err != nil {
			return mapApplicationError(err)
		}
		var act Activity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&act, "id = ?", app.ActivityID).Error; err != nil {
			return mapError(err)
		}
		if act.CreatorID != requesterID {
			return ErrForbidden
		}
		if app.Status == ApplicationApproved {
			updated = app
			return nil
		}
		if app.Status != ApplicationPending {
			return ErrInvalidState
		}
		if act.Status != StatusRecruiting {
			return ErrInvalidState
		}
		app.Status = ApplicationApproved
		if err := tx.Save(&app).Error; err != nil {
			return mapError(err)
		}
		act.JoinedCount++
		if act.JoinedCount >= act.RequiredCount {
			act.Status = StatusFull
		}
		if err := tx.Save(&act).Error; err != nil {
			return mapError(err)
		}
		updated = app
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *GormRepository) RejectApplication(ctx context.Context, applicationID, requesterID uuid.UUID) (*Application, error) {
	if r.db == nil {
		return nil, ErrUnavailable
	}
	var updated Application
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var app Application
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&app, "id = ?", applicationID).Error; err != nil {
			return mapApplicationError(err)
		}
		var act Activity
		if err := tx.First(&act, "id = ?", app.ActivityID).Error; err != nil {
			return mapError(err)
		}
		if act.CreatorID != requesterID {
			return ErrForbidden
		}
		if app.Status == ApplicationRejected {
			updated = app
			return nil
		}
		if app.Status != ApplicationPending {
			return ErrInvalidState
		}
		app.Status = ApplicationRejected
		if err := tx.Save(&app).Error; err != nil {
			return mapError(err)
		}
		updated = app
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

type activityCreatorRow struct {
	Activity
	CreatorNickname string
	CreatorSchool   string
}

func activityCreatorSelect() string {
	return "a.id, a.creator_id, a.title, a.\"type\", a.description, a.required_count, a.joined_count, a.tags, a.preferred_tags, a.time_text, a.location_text, a.status, a.created_at, a.updated_at, u.nickname AS creator_nickname, u.school AS creator_school"
}

func rowsToActivities(rows []activityCreatorRow) []ActivityWithCreator {
	out := make([]ActivityWithCreator, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toActivity())
	}
	return out
}

func (r activityCreatorRow) toActivity() ActivityWithCreator {
	return ActivityWithCreator{
		Activity: r.Activity,
		Creator: CreatorSummary{ID: r.CreatorID, Nickname: r.CreatorNickname, School: r.CreatorSchool},
	}
}

type applicationActivityRow struct {
	Application
	ActivityTitle  string
	ActivityStatus string
}

type applicationApplicantRow struct {
	Application
	ApplicantNickname string
	ApplicantSchool   string
}

func (r applicationApplicantRow) toApplication() ApplicationWithApplicant {
	return ApplicationWithApplicant{
		Application: r.Application,
		Applicant:  CreatorSummary{ID: r.ApplicantID, Nickname: r.ApplicantNickname, School: r.ApplicantSchool},
	}
}

func mapApplicationError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrApplicationNotFound
	}
	return mapError(err)
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == "23505" {
		return ErrDuplicateApply
	}
	return fmt.Errorf("activity repository: %w", err)
}
