package activity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	StatusRecruiting = "recruiting"
	StatusFull       = "full"
	StatusClosed     = "closed"

	ApplicationPending   = "pending"
	ApplicationApproved  = "approved"
	ApplicationRejected  = "rejected"
	ApplicationCancelled = "cancelled"
)

type StringList []string

func (s StringList) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	encoded, err := json.Marshal([]string(s))
	if err != nil {
		return nil, err
	}
	return string(encoded), nil
}

func (s *StringList) Scan(value any) error {
	if value == nil {
		*s = StringList{}
		return nil
	}
	var raw []byte
	switch typed := value.(type) {
	case []byte:
		raw = typed
	case string:
		raw = []byte(typed)
	default:
		return fmt.Errorf("scan StringList: unsupported type %T", value)
	}
	var decoded []string
	if len(raw) == 0 {
		decoded = []string{}
	} else if err := json.Unmarshal(raw, &decoded); err != nil {
		return err
	}
	*s = decoded
	return nil
}

type Activity struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatorID     uuid.UUID  `gorm:"column:creator_id;type:uuid;not null" json:"creator_id"`
	Title         string     `gorm:"size:120;not null" json:"title"`
	Type          string     `gorm:"size:64;not null;default:project" json:"type"`
	Description   string     `gorm:"type:text;not null" json:"description"`
	RequiredCount int        `gorm:"column:required_count;not null;default:2" json:"required_count"`
	JoinedCount   int        `gorm:"column:joined_count;not null;default:0" json:"joined_count"`
	Tags          StringList `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"tags"`
	PreferredTags StringList `gorm:"column:preferred_tags;type:jsonb;not null;default:'[]'::jsonb" json:"preferred_tags"`
	TimeText      string     `gorm:"column:time_text;size:120;not null;default:''" json:"time_text"`
	LocationText  string     `gorm:"column:location_text;size:160;not null;default:''" json:"location_text"`
	Status        string     `gorm:"size:32;not null;default:recruiting" json:"status"`
	CreatedAt     time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"not null" json:"updated_at"`
}

func (Activity) TableName() string { return "activities" }

type Application struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ActivityID  uuid.UUID `gorm:"column:activity_id;type:uuid;not null" json:"activity_id"`
	ApplicantID uuid.UUID `gorm:"column:applicant_id;type:uuid;not null" json:"applicant_id"`
	Reason      string    `gorm:"type:text;not null;default:''" json:"reason"`
	MatchScore  int       `gorm:"column:match_score;not null;default:0" json:"match_score"`
	Status      string    `gorm:"size:32;not null;default:pending" json:"status"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`
}

func (Application) TableName() string { return "applications" }

type CreatorSummary struct {
	ID       uuid.UUID `json:"id"`
	Nickname string    `json:"nickname"`
	School   string    `json:"school"`
}

type ActivityWithCreator struct {
	Activity
	Creator CreatorSummary `json:"creator"`
}

type ApplicationWithActivity struct {
	Application
	ActivityTitle  string `json:"activity_title"`
	ActivityStatus string `json:"activity_status"`
}

type ApplicationWithApplicant struct {
	Application
	Applicant CreatorSummary `json:"applicant"`
}
