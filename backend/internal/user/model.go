package user

import (
	"time"

	"github.com/google/uuid"
)

// User is the persisted account model.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Email        string    `gorm:"size:255;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;size:255;not null" json:"-"`
	Nickname     string    `gorm:"size:80" json:"nickname"`
	Role         string    `gorm:"size:32;not null;default:user" json:"role"`
	School       string    `gorm:"size:120" json:"school"`
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`
}

// PublicUser contains only fields safe to return through the API.
type PublicUser struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Nickname  string    `json:"nickname"`
	Role      string    `json:"role"`
	School    string    `json:"school"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName pins the GORM model to the existing users table.
func (User) TableName() string {
	return "users"
}

// Public returns a representation that never contains the password hash.
func (u User) Public() PublicUser {
	return PublicUser{
		ID:        u.ID,
		Email:     u.Email,
		Nickname:  u.Nickname,
		Role:      u.Role,
		School:    u.School,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
