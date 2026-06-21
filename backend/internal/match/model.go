package match

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"matchlab/backend/internal/activity"
	"matchlab/backend/internal/questionnaire"
)

const AlgorithmVersion = "activity-rules-v1"

type DetailScores struct {
	Interest int `json:"interest"`
	Skill    int `json:"skill"`
	Type     int `json:"type"`
	Time     int `json:"time"`
	Goal     int `json:"goal"`
}

type Explanation struct {
	DetailScores DetailScores `json:"detail_scores"`
	Reason       string       `json:"reason"`
}

func (e Explanation) Value() (driver.Value, error) {
	encoded, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return string(encoded), nil
}

func (e *Explanation) Scan(value any) error {
	if value == nil {
		*e = Explanation{}
		return nil
	}
	var raw []byte
	switch typed := value.(type) {
	case []byte:
		raw = typed
	case string:
		raw = []byte(typed)
	default:
		return fmt.Errorf("scan explanation: unsupported type %T", value)
	}
	return json.Unmarshal(raw, e)
}

type UserSignals struct {
	Profile         questionnaire.Profile
	QuestionnaireID uuid.UUID
	Answers         questionnaire.Answers
}

type Recommendation struct {
	Activity     activity.Activity `json:"activity"`
	Score        int               `json:"score"`
	DetailScores DetailScores      `json:"detail_scores"`
	Reason       string            `json:"reason"`
}

type Record struct {
	ID               uuid.UUID   `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID           uuid.UUID   `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	ActivityID       uuid.UUID   `gorm:"column:activity_id;type:uuid;not null" json:"activity_id"`
	QuestionnaireID  *uuid.UUID  `gorm:"column:questionnaire_id;type:uuid" json:"questionnaire_id,omitempty"`
	Score            float64     `gorm:"not null" json:"score"`
	Explanation      Explanation `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"-"`
	AlgorithmVersion string      `gorm:"column:algorithm_version;size:32;not null" json:"algorithm_version"`
	Status           string      `gorm:"size:32;not null;default:recommended" json:"status"`
	CreatedAt        time.Time   `gorm:"not null" json:"created_at"`
	UpdatedAt        time.Time   `gorm:"not null" json:"updated_at"`
}

func (Record) TableName() string { return "matches" }

type SavedRecommendation struct {
	ID               uuid.UUID         `json:"id"`
	Activity         activity.Activity `json:"activity"`
	Score            int               `json:"score"`
	DetailScores     DetailScores      `json:"detail_scores"`
	Reason           string            `json:"reason"`
	AlgorithmVersion string            `json:"algorithm_version"`
	Status           string            `json:"status"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}
