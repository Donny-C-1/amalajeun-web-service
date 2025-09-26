package places

import (
	"log"
	"time"

	"github.com/donny-c-1/amalajeun/models"
)

// PlacesService handles the business logic for place discovery and management
type PlacesService struct {
	placesClient *GooglePlacesClient
	spotRepo     *SpotRepository
}

// NewPlacesService creates a new places service
func NewPlacesService() *PlacesService {
	return &PlacesService{
		placesClient: NewGooglePlacesClient(),
		spotRepo:     NewSpotRepository(),
	}
}

// RunPlacesDiscovery discovers Amala spots from Google Places API and adds them to the verification queue
func (s *PlacesService) RunPlacesDiscovery() error {
	log.Println("Starting Amala places discovery...")

	// Search for Amala spots in Lagos
	googlePlaces, err := s.placesClient.SearchAmalaInLagos()
	if err != nil {
		log.Printf("Error searching Google Places API: %v", err)
		return err
	}

	var newSpotsCount int
	var updatedSpotsCount int

	// Process each discovered place
	for _, googlePlace := range googlePlaces {
		err := s.processDiscoveredPlace(googlePlace)
		if err != nil {
			log.Printf("Error processing place %s: %v", googlePlace.PlaceID, err)
			continue // Continue with next place even if one fails
		}

		// Track statistics
		// Check if this was a new spot or an update
		existingSpot, err := s.spotRepo.FindByPlaceID(googlePlace.PlaceID)
		if err != nil {
			log.Printf("Error checking if spot was created: %v", err)
			continue
		}

		if existingSpot != nil && existingSpot.LastSeen != nil {
			// Spot already existed, this was an update
			updatedSpotsCount++
		} else {
			// This was a new spot
			newSpotsCount++
		}
	}

	// Log discovery results
	log.Printf("Places discovery completed. New spots: %d, Updated spots: %d, Total processed: %d",
		newSpotsCount, updatedSpotsCount, len(googlePlaces))

	// Get current count of pending verification spots
	pendingCount, err := s.spotRepo.CountPendingVerification()
	if err != nil {
		log.Printf("Error counting pending verification spots: %v", err)
	} else {
		log.Printf("Total spots pending verification: %d", pendingCount)
	}

	return nil
}

// processDiscoveredPlace handles an individual discovered place from Google Places API
// Now includes enhanced duplicate prevention logic that goes beyond just PlaceID checking
func (s *PlacesService) processDiscoveredPlace(googlePlace GooglePlace) error {
	// First check if place already exists by PlaceID (existing logic)
	existingSpot, err := s.spotRepo.FindByPlaceID(googlePlace.PlaceID)
	if err != nil {
		return err
	}

	if existingSpot != nil {
		// Place exists by PlaceID, update last_seen timestamp
		log.Printf("PLACES_UPDATE: Updating last_seen for existing spot: %s (PlaceID: %s, ID: %d)",
			existingSpot.Name, existingSpot.PlaceID, existingSpot.ID)
		return s.spotRepo.UpdateLastSeen(existingSpot)
	}

	// Place doesn't exist by PlaceID, but check for duplicates using name + location
	// This catches cases where the same physical location has different PlaceIDs
	log.Printf("PLACES_DISCOVERY: Processing new place: %s (%s)", googlePlace.Name, googlePlace.PlaceID)

	// CreateFromGooglePlace now includes duplicate checking internally
	newSpot, err := s.spotRepo.CreateFromGooglePlace(googlePlace)
	if err != nil {
		// Check if this is a duplicate error
		if isDuplicateError(err) {
			log.Printf("PLACES_DUPLICATE: Skipping duplicate place from Google Places: %s (%s) - %v",
				googlePlace.Name, googlePlace.PlaceID, err)
			// Return nil to continue processing other places - this is not a fatal error
			return nil
		}
		// Other errors should be returned
		return err
	}

	log.Printf("PLACES_SUCCESS: Successfully added new spot to verification queue: %s (ID: %d)",
		newSpot.Name, newSpot.ID)
	return nil
}

// isDuplicateError checks if an error is related to duplicate detection
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error message contains duplicate-related keywords
	errMsg := err.Error()
	return contains(errMsg, "duplicate spot detected") ||
		contains(errMsg, "similar spot exists") ||
		contains(errMsg, "matches existing spot")
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				indexOfSubstring(s, substr) >= 0)))
}

// indexOfSubstring finds the index of a substring in a string
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// GetPendingVerificationSpots returns all spots pending verification
func (s *PlacesService) GetPendingVerificationSpots() ([]models.Spot, error) {
	var spots []models.Spot
	result := s.spotRepo.db.
		Where("status = ?", models.StatusPendingVerification).
		Order("created_at DESC").
		Find(&spots)

	return spots, result.Error
}

// VerifySpot changes a spot's status from pending to verified
func (s *PlacesService) VerifySpot(spotID uint) error {
	result := s.spotRepo.db.Model(&models.Spot{}).
		Where("id = ? AND status = ?", spotID, models.StatusPendingVerification).
		Updates(map[string]interface{}{
			"status":     models.StatusVerified,
			"verified":   true,
			"updated_at": time.Now(),
		})

	return result.Error
}

// RejectSpot changes a spot's status from pending to rejected
func (s *PlacesService) RejectSpot(spotID uint) error {
	result := s.spotRepo.db.Model(&models.Spot{}).
		Where("id = ? AND status = ?", spotID, models.StatusPendingVerification).
		Updates(map[string]interface{}{
			"status":     models.StatusRejected,
			"updated_at": time.Now(),
		})

	return result.Error
}
