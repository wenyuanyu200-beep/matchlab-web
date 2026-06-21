package questionnaire

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type StringList []string

func (s StringList) Value() (driver.Value, error) { return jsonValue(s, StringList{}) }
func (s *StringList) Scan(value any) error        { return scanJSON(value, s) }

type Answers struct {
	Interests          StringList `json:"interests"`
	Skills             StringList `json:"skills"`
	AvailableTime      string     `json:"available_time"`
	ActivityTypes      StringList `json:"activity_types"`
	Goal               string     `json:"goal"`
	CommunicationStyle string     `json:"communication_style"`
}

func (a Answers) Value() (driver.Value, error) { return jsonValue(a, Answers{}) }
func (a *Answers) Scan(value any) error        { return scanJSON(value, a) }

type ProfileScores struct {
	InterestScore      int `json:"interest_score"`
	SkillScore         int `json:"skill_score"`
	TimeScore          int `json:"time_score"`
	GoalScore          int `json:"goal_score"`
	CommunicationScore int `json:"communication_score"`
}

func (s ProfileScores) Value() (driver.Value, error) { return jsonValue(s, ProfileScores{}) }
func (s *ProfileScores) Scan(value any) error        { return scanJSON(value, s) }

type Questionnaire struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	Mode        string    `gorm:"size:32;not null;default:activity" json:"mode"`
	Version     int       `gorm:"not null;default:1" json:"version"`
	Answers     Answers   `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"answers"`
	Scores      JSONMap   `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"scores"`
	Status      string    `gorm:"size:32;not null;default:completed" json:"status"`
	CompletedAt time.Time `gorm:"column:completed_at" json:"completed_at"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`
}

func (Questionnaire) TableName() string { return "questionnaires" }

type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) { return jsonValue(m, JSONMap{}) }
func (m *JSONMap) Scan(value any) error        { return scanJSON(value, m) }

type Profile struct {
	ID          uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID      uuid.UUID     `gorm:"column:user_id;type:uuid;not null;unique" json:"user_id"`
	DisplayName string        `gorm:"column:display_name;size:80;not null" json:"-"`
	ProfileType string        `gorm:"column:profile_type;size:32;not null;default:activity" json:"profile_type"`
	Tags        StringList    `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"tags"`
	Scores      ProfileScores `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"scores"`
	Summary     string        `gorm:"type:text;not null;default:''" json:"summary"`
	CreatedAt   time.Time     `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"not null" json:"updated_at"`
}

func (Profile) TableName() string { return "profiles" }

type GeneratedProfile struct {
	ProfileType string        `json:"profile_type"`
	Tags        StringList    `json:"tags"`
	Scores      ProfileScores `json:"scores"`
	Summary     string        `json:"summary"`
}

func jsonValue(value any, empty any) (driver.Value, error) {
	if value == nil {
		value = empty
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return string(encoded), nil
}

func scanJSON(value any, destination any) error {
	if value == nil {
		return nil
	}
	var raw []byte
	switch typed := value.(type) {
	case []byte:
		raw = typed
	case string:
		raw = []byte(typed)
	default:
		return fmt.Errorf("scan JSON: unsupported type %T", value)
	}
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, destination)
}
