package places

import (
	"time"

	"github.com/donny-c-1/amalajeun/database"
	"github.com/donny-c-1/amalajeun/models"
	"gorm.io/gorm"
)

// SpotRepository handles database operations for spots
type SpotRepository struct {
	db *gorm.DB
}

// NewSpotRepository creates a new spot repository
func NewSpotRepository() *SpotRepository {
	return &SpotRepository{
		db: database.DB,
	}
}

// FindByPlaceID finds a spot by its Google Places ID
func (r *SpotRepository) FindByPlaceID(placeID string) (*models.Spot, error) {
	var spot models.Spot
	result := r.db.Where("place_id = ?", placeID).First(&spot)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Spot not found
		}
		return nil, result.Error
	}

	return &spot, nil
}

// CreateFromGooglePlace creates a new spot from a Google Place result
func (r *SpotRepository) CreateFromGooglePlace(googlePlace GooglePlace) (*models.Spot, error) {
	now := time.Now()

	spot := &models.Spot{
		Name:      googlePlace.Name,
		Address:   googlePlace.Address,
		Latitude:  googlePlace.Latitude,
		Longitude: googlePlace.Longitude,
		PlaceID:   googlePlace.PlaceID,
		Source:    models.SourceScraper,
		Status:    models.StatusPendingVerification,
		LastSeen:  &now,
		CreatedAt: now,
		UpdatedAt: now,
		// UserID and AddedBy remain nil/empty for scraper sources
	}

	result := r.db.Create(spot)
	if result.Error != nil {
		return nil, result.Error
	}

	return spot, nil
}

// UpdateLastSeen updates the last_seen timestamp for a spot
func (r *SpotRepository) UpdateLastSeen(spot *models.Spot) error {
	now := time.Now()
	spot.LastSeen = &now
	spot.UpdatedAt = now

	result := r.db.Model(spot).Updates(map[string]interface{}{
		"last_seen":  spot.LastSeen,
		"updated_at": spot.UpdatedAt,
	})

	return result.Error
}

// CountPendingVerification returns the number of spots pending verification
func (r *SpotRepository) CountPendingVerification() (int64, error) {
	var count int64
	result := r.db.Model(&models.Spot{}).
		Where("status = ?", models.StatusPendingVerification).
		Count(&count)

	return count, result.Error
}
