package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Review represents a review for an Amala spot
type Review struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	SpotID    uint           `json:"spot_id" gorm:"not null" binding:"required"`
	UserID    *uuid.UUID     `json:"user_id" gorm:"type:uuid;index"`
	UserName  string         `json:"user_name" gorm:"not null" binding:"required"` // Keep for backward compatibility during migration
	Rating    int            `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5" binding:"required,min=1,max=5"`
	Comment   string         `json:"comment" gorm:"type:text"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Spot Spot  `json:"spot,omitempty" gorm:"foreignKey:SpotID"`
}

// TableName specifies the table name for the Review model
func (Review) TableName() string {
	return "reviews"
}
