package places

import (
	"fmt"
	"log"
	"time"

	"github.com/donny-c-1/amalajeun/database"
	"github.com/donny-c-1/amalajeun/models"
	"github.com/donny-c-1/amalajeun/services/duplicate"
	"gorm.io/gorm"
)

// SpotRepository handles database operations for spots
type SpotRepository struct {
	db               *gorm.DB
	duplicateService *duplicate.DuplicateService
}

// NewSpotRepository creates a new spot repository
func NewSpotRepository() *SpotRepository {
	return &SpotRepository{
		db:               database.DB,
		duplicateService: duplicate.NewDuplicateService(database.DB),
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
// Now includes duplicate prevention logic before database insertion
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

	// Check for duplicates using the new duplicate detection system
	duplicateResult, err := r.duplicateService.CheckForDuplicate(spot)
	if err != nil {
		log.Printf("ERROR: Duplicate check failed for Google Place spot '%s': %v", spot.Name, err)
		// Continue with creation if duplicate check fails (fail-safe approach)
	} else if duplicateResult.IsDuplicate {
		log.Printf("DUPLICATE_PREVENTED: Google Places spot '%s' (%s) rejected - %s",
			spot.Name, googlePlace.PlaceID, duplicateResult.Reason)
		return nil, fmt.Errorf("duplicate spot detected: %s (matches existing spot ID: %d)",
			duplicateResult.Reason, duplicateResult.ExistingSpot.ID)
	}

	result := r.db.Create(spot)
	if result.Error != nil {
		return nil, result.Error
	}

	log.Printf("SPOT_CREATED: New spot from Google Places - '%s' (ID: %d, PlaceID: %s)",
		spot.Name, spot.ID, spot.PlaceID)

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

// CreateSpotWithDuplicateCheck creates a new spot with comprehensive duplicate prevention
// This method is used for user-submitted spots and manual spot creation
func (r *SpotRepository) CreateSpotWithDuplicateCheck(spot *models.Spot) (*models.Spot, error) {
	// Perform duplicate check before creation
	duplicateResult, err := r.duplicateService.CheckForDuplicate(spot)
	if err != nil {
		log.Printf("ERROR: Duplicate check failed for spot '%s': %v", spot.Name, err)
		// Continue with creation if duplicate check fails (fail-safe approach)
	} else if duplicateResult.IsDuplicate {
		log.Printf("DUPLICATE_PREVENTED: User spot '%s' rejected - %s",
			spot.Name, duplicateResult.Reason)
		return nil, fmt.Errorf("duplicate spot detected: %s (matches existing spot ID: %d)",
			duplicateResult.Reason, duplicateResult.ExistingSpot.ID)
	}

	// Create the spot in database
	result := r.db.Create(spot)
	if result.Error != nil {
		return nil, result.Error
	}

	log.Printf("SPOT_CREATED: New user spot - '%s' (ID: %d, Source: %s)",
		spot.Name, spot.ID, spot.Source)

	return spot, nil
}

// FindSpotsNearLocation finds spots within a specified radius of a location
// This method is used for geospatial duplicate detection and location-based queries
func (r *SpotRepository) FindSpotsNearLocation(lat, lng, radiusM float64, limit int) ([]models.Spot, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}

	// Use bounding box for efficient database query
	minLat, maxLat, minLng, maxLng := getBoundingBox(lat, lng, radiusM)

	var spots []models.Spot
	result := r.db.Where("latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?",
		minLat, maxLat, minLng, maxLng).
		Where("deleted_at IS NULL").
		Limit(limit).
		Find(&spots)

	return spots, result.Error
}

// FindSpotsByNamePattern finds spots with names similar to the given pattern
// Uses database ILIKE for basic pattern matching - more sophisticated matching
// is handled by the duplicate service
func (r *SpotRepository) FindSpotsByNamePattern(namePattern string, limit int) ([]models.Spot, error) {
	if limit <= 0 {
		limit = 20 // Default limit
	}

	var spots []models.Spot
	result := r.db.Where("name ILIKE ?", "%"+namePattern+"%").
		Where("deleted_at IS NULL").
		Limit(limit).
		Find(&spots)

	return spots, result.Error
}

// CheckForDuplicates returns duplicate check result for a spot without creating it
// Useful for validation and showing users potential duplicates before submission
func (r *SpotRepository) CheckForDuplicates(spot *models.Spot) (*duplicate.DuplicateResult, error) {
	return r.duplicateService.CheckForDuplicate(spot)
}

// FindSimilarSpots returns spots similar to the given spot with similarity scores
// Useful for administrative purposes and duplicate management
func (r *SpotRepository) FindSimilarSpots(spot *models.Spot, limit int) ([]duplicate.DuplicateResult, error) {
	return r.duplicateService.FindSimilarSpots(spot, limit)
}

// getBoundingBox calculates a bounding box around a point for efficient database queries
// This is a local helper function for the repository
func getBoundingBox(centerLat, centerLng, radiusM float64) (minLat, maxLat, minLng, maxLng float64) {
	// Convert radius from meters to degrees (approximate)
	// 1 degree latitude ≈ 111,320 meters
	// 1 degree longitude ≈ 111,320 * cos(latitude) meters
	latDelta := radiusM / 111320.0
	lngDelta := radiusM / (111320.0 * (0.017453292519943295 * centerLat)) // cos(lat in radians)

	minLat = centerLat - latDelta
	maxLat = centerLat + latDelta
	minLng = centerLng - lngDelta
	maxLng = centerLng + lngDelta

	return minLat, maxLat, minLng, maxLng
}
