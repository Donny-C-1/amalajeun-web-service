package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Source represents the source of a spot entry
type Source string

const (
	SourceUser    Source = "user"
	SourceAgent   Source = "agent"
	SourceScraper Source = "scraper"
)

// Status represents the verification status of a spot
type Status string

const (
	StatusPendingVerification Status = "pending_verification"
	StatusVerified            Status = "verified"
	StatusRejected            Status = "rejected"
)

// Spot represents an Amala spot in Lagos
type Spot struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null" binding:"required"`
	Address   string         `json:"address" gorm:"not null" binding:"required"`
	Latitude  float64        `json:"latitude" gorm:"not null" binding:"required"`
	Longitude float64        `json:"longitude" gorm:"not null" binding:"required"`
	UserID    *uuid.UUID     `json:"user_id" gorm:"type:uuid;index"`
	AddedBy   string         `json:"added_by" gorm:"not null" binding:"required"` // Keep for backward compatibility during migration
	Verified  bool           `json:"verified" gorm:"default:false"`
	Source    Source         `json:"source" gorm:"type:varchar(20);not null" binding:"required"`
	Status    Status         `json:"status" gorm:"type:varchar(20);default:'pending_verification'"`
	PlaceID   string         `json:"place_id" gorm:"type:varchar(255);uniqueIndex"`
	LastSeen  *time.Time     `json:"last_seen"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User    *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Reviews []Review `json:"reviews,omitempty" gorm:"foreignKey:SpotID"`
}

// TableName specifies the table name for the Spot model
func (Spot) TableName() string {
	return "spots"
}
