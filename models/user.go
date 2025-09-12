package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	GoogleID    string         `json:"google_id" gorm:"uniqueIndex;not null"`
	Email       string         `json:"email" gorm:"uniqueIndex;not null"`
	Name        string         `json:"name" gorm:"not null"`
	AvatarURL   string         `json:"avatar_url"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	LastLoginAt *time.Time     `json:"last_login_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Spots   []Spot   `json:"spots,omitempty" gorm:"foreignKey:UserID"`
	Reviews []Review `json:"reviews,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate will set a UUID rather than numeric ID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
