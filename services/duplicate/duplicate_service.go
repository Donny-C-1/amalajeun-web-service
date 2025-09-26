package duplicate

import (
	"fmt"
	"log"

	"github.com/donny-c-1/amalajeun/models"
	"github.com/donny-c-1/amalajeun/utils"
	"gorm.io/gorm"
)

// DuplicateResult represents the result of a duplicate check
type DuplicateResult struct {
	IsDuplicate    bool         `json:"is_duplicate"`
	ExistingSpot   *models.Spot `json:"existing_spot,omitempty"`
	Distance       float64      `json:"distance,omitempty"`        // Distance in meters
	NameSimilarity float64      `json:"name_similarity,omitempty"` // Similarity score 0.0-1.0
	Reason         string       `json:"reason,omitempty"`          // Human-readable reason
}

// DuplicateService handles duplicate detection logic for spots
type DuplicateService struct {
	db *gorm.DB
}

// NewDuplicateService creates a new duplicate detection service
func NewDuplicateService(db *gorm.DB) *DuplicateService {
	return &DuplicateService{
		db: db,
	}
}

// CheckForDuplicate performs comprehensive duplicate detection for a new spot
// It checks both name similarity and geospatial proximity against existing spots
func (s *DuplicateService) CheckForDuplicate(newSpot *models.Spot) (*DuplicateResult, error) {
	// Validate input coordinates
	if !utils.IsValidLagosCoordinate(newSpot.Latitude, newSpot.Longitude) {
		log.Printf("DUPLICATE_CHECK: Invalid coordinates for spot '%s': lat=%.6f lng=%.6f",
			newSpot.Name, newSpot.Latitude, newSpot.Longitude)
		return &DuplicateResult{
			IsDuplicate: false,
			Reason:      "Invalid coordinates - outside Lagos bounds",
		}, nil
	}

	// Find potentially similar spots using geospatial proximity
	candidateSpots, err := s.findNearbySpots(newSpot.Latitude, newSpot.Longitude)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby spots: %w", err)
	}

	log.Printf("DUPLICATE_CHECK: Found %d candidate spots within proximity for '%s'",
		len(candidateSpots), newSpot.Name)

	// Check each candidate for duplicate criteria
	for _, candidate := range candidateSpots {
		result := s.evaluateDuplicateCandidate(newSpot, &candidate)
		if result.IsDuplicate {
			log.Printf("DUPLICATE_DETECTED: spot='%s' matches existing spot_id=%d (name='%s') distance=%.1fm name_similarity=%.2f reason='%s'",
				newSpot.Name, candidate.ID, candidate.Name, result.Distance, result.NameSimilarity, result.Reason)
			return result, nil
		}
	}

	log.Printf("DUPLICATE_CHECK: No duplicates found for spot '%s' at lat=%.6f lng=%.6f",
		newSpot.Name, newSpot.Latitude, newSpot.Longitude)

	return &DuplicateResult{
		IsDuplicate: false,
		Reason:      "No similar spots found within tolerance",
	}, nil
}

// findNearbySpots retrieves spots within a reasonable distance for duplicate checking
// Uses bounding box for efficient database querying, then applies precise distance filtering
func (s *DuplicateService) findNearbySpots(lat, lng float64) ([]models.Spot, error) {
	// Calculate bounding box for efficient database query
	// Use a larger radius for initial filtering to ensure we don't miss edge cases
	searchRadiusM := utils.DuplicateDistanceToleranceM * 2 // 100m search radius
	minLat, maxLat, minLng, maxLng := utils.GetBoundingBox(lat, lng, searchRadiusM)

	var spots []models.Spot

	// Query database with bounding box filter
	// Only check against verified and pending verification spots (not rejected)
	result := s.db.Where("latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?",
		minLat, maxLat, minLng, maxLng).
		Where("status IN (?)", []models.Status{models.StatusVerified, models.StatusPendingVerification}).
		Where("deleted_at IS NULL").
		Find(&spots)

	if result.Error != nil {
		return nil, result.Error
	}

	// Apply precise distance filtering using Haversine formula
	var nearbySpots []models.Spot
	for _, spot := range spots {
		distance := utils.HaversineDistance(lat, lng, spot.Latitude, spot.Longitude)
		if distance <= searchRadiusM {
			nearbySpots = append(nearbySpots, spot)
		}
	}

	return nearbySpots, nil
}

// evaluateDuplicateCandidate checks if a candidate spot is a duplicate of the new spot
func (s *DuplicateService) evaluateDuplicateCandidate(newSpot *models.Spot, candidate *models.Spot) *DuplicateResult {
	// Calculate distance between spots
	distance := utils.HaversineDistance(
		newSpot.Latitude, newSpot.Longitude,
		candidate.Latitude, candidate.Longitude,
	)

	// Calculate name similarity
	nameSimilarity := utils.CalculateNameSimilarity(newSpot.Name, candidate.Name)

	// Determine if this is a duplicate based on our criteria:
	// 1. Distance within tolerance AND high name similarity
	// 2. Very close distance (< 25m) with moderate name similarity
	// 3. Identical normalized names regardless of distance (within search radius)

	var isDuplicate bool
	var reason string

	if distance <= utils.DuplicateDistanceToleranceM && nameSimilarity >= utils.NameSimilarityThreshold {
		// Primary duplicate criteria: close distance + similar names
		isDuplicate = true
		reason = fmt.Sprintf("Similar names (%.2f similarity) within %dm tolerance", nameSimilarity, int(utils.DuplicateDistanceToleranceM))
	} else if distance <= 25.0 && nameSimilarity >= 0.6 {
		// Very close spots with moderate name similarity
		isDuplicate = true
		reason = fmt.Sprintf("Very close proximity (%.1fm) with moderate name similarity (%.2f)", distance, nameSimilarity)
	} else if nameSimilarity >= 0.95 && distance <= 100.0 {
		// Nearly identical names within reasonable distance
		isDuplicate = true
		reason = fmt.Sprintf("Nearly identical names (%.2f similarity) within 100m", nameSimilarity)
	}

	return &DuplicateResult{
		IsDuplicate:    isDuplicate,
		ExistingSpot:   candidate,
		Distance:       distance,
		NameSimilarity: nameSimilarity,
		Reason:         reason,
	}
}

// FindSimilarSpots returns all spots that are similar to the given spot
// This is useful for administrative purposes or showing users potential duplicates
func (s *DuplicateService) FindSimilarSpots(spot *models.Spot, limit int) ([]DuplicateResult, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	candidateSpots, err := s.findNearbySpots(spot.Latitude, spot.Longitude)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby spots: %w", err)
	}

	var results []DuplicateResult
	for _, candidate := range candidateSpots {
		// Skip the same spot if it already exists in database
		if spot.ID != 0 && candidate.ID == spot.ID {
			continue
		}

		result := s.evaluateDuplicateCandidate(spot, &candidate)

		// Include all results with some similarity (not just duplicates)
		if result.NameSimilarity > 0.3 || result.Distance <= 200.0 {
			results = append(results, *result)
		}

		// Respect the limit
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// IsDuplicateSpot is a convenience method that returns only the boolean result
func (s *DuplicateService) IsDuplicateSpot(spot *models.Spot) (bool, error) {
	result, err := s.CheckForDuplicate(spot)
	if err != nil {
		return false, err
	}
	return result.IsDuplicate, nil
}
