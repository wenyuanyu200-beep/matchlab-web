package circle

import (
	"time"

	"github.com/google/uuid"
)

type Circle struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatorID uuid.UUID `gorm:"column:creator_id" json:"creator_id"`
	Name string `json:"name"`
	Description string `json:"description"`
	Category string `json:"category"`
	Status string `json:"status"`
	MemberCount int `gorm:"column:member_count" json:"member_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MemberSummary struct {
	ID uuid.UUID `json:"id"`
	Nickname string `json:"nickname"`
	School string `json:"school"`
	Role string `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type Channel struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CircleID uuid.UUID `gorm:"column:circle_id" json:"circle_id"`
	Name string `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CircleID uuid.UUID `gorm:"column:circle_id" json:"circle_id"`
	ChannelID uuid.UUID `gorm:"column:channel_id" json:"channel_id"`
	SenderID uuid.UUID `gorm:"column:sender_id" json:"sender_id"`
	Content string `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	SenderNickname string `gorm:"column:sender_nickname;->" json:"sender_nickname"`
}

type AdminFilter struct { Status string }
