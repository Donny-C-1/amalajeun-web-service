package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SpotVerification represents a user's verification of a spot
// This model tracks which users have verified which spots to implement
// the new verification workflow requiring 3 unique verifications
//
// Key features:
// - Composite unique constraint on (SpotID, UserID) prevents duplicate verifications
// - Foreign key relationships to Spot and User models
// - Timestamps for tracking verification activity
// - Soft delete support for audit purposes
type SpotVerification struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	SpotID    uint           `json:"spot_id" gorm:"not null;index:idx_spot_user_unique"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index:idx_spot_user_unique"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Spot *Spot `json:"spot,omitempty" gorm:"foreignKey:SpotID"`
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for the SpotVerification model
func (SpotVerification) TableName() string {
	return "spot_verifications"
}

// BeforeCreate sets the timestamp before creating a new verification record
func (sv *SpotVerification) BeforeCreate(tx *gorm.DB) error {
	sv.CreatedAt = time.Now()
	sv.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate updates the timestamp before updating a verification record
func (sv *SpotVerification) BeforeUpdate(tx *gorm.DB) error {
	sv.UpdatedAt = time.Now()
	return nil
}
