package admin

import (
	"time"

	"github.com/google/uuid"
)

type Stats struct {
	UsersCount                int64 `json:"users_count"`
	ActivitiesCount           int64 `json:"activities_count"`
	ApplicationsCount         int64 `json:"applications_count"`
	MatchesCount              int64 `json:"matches_count"`
	QuestionnairesCount       int64 `json:"questionnaires_count"`
	FeedbacksCount            int64 `json:"feedbacks_count"`
	RecruitingActivitiesCount int64 `json:"recruiting_activities_count"`
	PendingApplicationsCount  int64 `json:"pending_applications_count"`
	ApprovedApplicationsCount int64 `json:"approved_applications_count"`
}

type Page struct {
	Limit  int
	Offset int
}

type UsersFilter struct {
	Keyword string
	Role    string
	Page
}

type ActivitiesFilter struct {
	Keyword string
	Type    string
	Status  string
	Page
}

type ApplicationsFilter struct {
	Status string
	Page
}

type AdminUser struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Nickname  string    `json:"nickname"`
	Role      string    `json:"role"`
	School    string    `json:"school"`
	CreatedAt time.Time `json:"created_at"`
}

type PersonSummary struct {
	Nickname string `json:"nickname"`
	School   string `json:"school"`
}

type AdminActivity struct {
	ID            uuid.UUID     `json:"id"`
	Title         string        `json:"title"`
	Type          string        `json:"type"`
	Status        string        `json:"status"`
	CreatorID     uuid.UUID     `json:"creator_id"`
	Creator       PersonSummary `json:"creator"`
	RequiredCount int           `json:"required_count"`
	JoinedCount   int           `json:"joined_count"`
	CreatedAt     time.Time     `json:"created_at"`
}

type AdminApplication struct {
	ID            uuid.UUID     `json:"id"`
	ActivityID    uuid.UUID     `json:"activity_id"`
	ActivityTitle string        `json:"activity_title"`
	ApplicantID   uuid.UUID     `json:"applicant_id"`
	Applicant     PersonSummary `json:"applicant"`
	Reason        string        `json:"reason"`
	MatchScore    int           `json:"match_score"`
	Status        string        `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
}

type Feedback struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	ActivityID *uuid.UUID `json:"activity_id,omitempty"`
	MatchID    *uuid.UUID `json:"match_id,omitempty"`
	Rating     int16      `json:"rating"`
	Comment    *string    `json:"comment,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
