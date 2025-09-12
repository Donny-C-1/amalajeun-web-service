package models

import (
	"time"

	"gorm.io/gorm"
)

// Source represents the source of a spot entry
type Source string

const (
	SourceUser    Source = "user"
	SourceAgent   Source = "agent"
	SourceScraper Source = "scraper"
)

// Spot represents an Amala spot in Lagos
type Spot struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null" binding:"required"`
	Address   string         `json:"address" gorm:"not null" binding:"required"`
	Latitude  float64        `json:"latitude" gorm:"not null" binding:"required"`
	Longitude float64        `json:"longitude" gorm:"not null" binding:"required"`
	AddedBy   string         `json:"added_by" gorm:"not null" binding:"required"`
	Verified  bool           `json:"verified" gorm:"default:false"`
	Source    Source         `json:"source" gorm:"type:varchar(20);not null" binding:"required"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationship
	Reviews []Review `json:"reviews,omitempty" gorm:"foreignKey:SpotID"`
}

// TableName specifies the table name for the Spot model
func (Spot) TableName() string {
	return "spots"
}
