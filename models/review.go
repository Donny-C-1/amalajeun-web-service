package models

import (
	"time"

	"gorm.io/gorm"
)

// Review represents a review for an Amala spot
type Review struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	SpotID    uint           `json:"spot_id" gorm:"not null" binding:"required"`
	UserName  string         `json:"user_name" gorm:"not null" binding:"required"`
	Rating    int            `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5" binding:"required,min=1,max=5"`
	Comment   string         `json:"comment" gorm:"type:text"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationship
	Spot Spot `json:"spot,omitempty" gorm:"foreignKey:SpotID"`
}

// TableName specifies the table name for the Review model
func (Review) TableName() string {
	return "reviews"
}
