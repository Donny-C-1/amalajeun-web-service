# Places Discovery Service

This service handles automatic discovery of Amala spots using the Google Places API and manages the verification queue.

## Overview

The service consists of three main components:

1. **GooglePlacesClient** - Handles API calls to Google Places Text Search
2. **SpotRepository** - Database operations for spot management
3. **PlacesService** - Business logic for place discovery and verification

## Features

- **Automatic Discovery**: Searches for "Amala" spots in Lagos daily
- **Verification Queue**: New spots are added with `pending_verification` status
- **Advanced Duplicate Prevention**: Multi-layered duplicate detection using name similarity + geospatial proximity
- **Legacy PlaceID Support**: Maintains `place_id` uniqueness for Google Places API compatibility
- **Last Seen Tracking**: Updates timestamp when existing spots are rediscovered
- **Error Handling**: Graceful error handling with comprehensive logging

## Environment Variables

Required environment variables:
- `GOOGLE_PLACES_API_KEY` - Your Google Places API key
- `DATABASE_URL` - PostgreSQL connection string

## Usage

The service runs automatically:
1. **On startup** - Immediate discovery run
2. **Daily** - Scheduled discovery at midnight

### Manual Testing

You can test the service manually:

```go
package main

import (
	"log"
	"os"

	"github.com/donny-c-1/amalajeun/database"
	"github.com/donny-c-1/amalajeun/services/places"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// Run discovery
	service := places.NewPlacesService()
	if err := service.RunPlacesDiscovery(); err != nil {
		log.Fatal("Discovery failed:", err)
	}

	log.Println("Discovery completed successfully")
}
```

## API Integration Details

The modern Google Places API uses POST requests with specific headers and field masks. The response is parsed into the following structure:

```go
type GooglePlace struct {
	PlaceID   string  `json:"place_id"`
	Name      string  `json:"name"`
	Address   string  `json:"formatted_address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	MapsURL   string  `json:"maps_url"`
}
```

### Required Headers:
- `Content-Type: application/json`
- `X-Goog-Api-Key: <your_api_key>`
- `X-Goog-FieldMask: "places.id,places.displayName,places.location,places.formattedAddress"`

### Request Body:
```json
{
  "textQuery": "Amala in Lagos, Nigeria"
}
```

### API Endpoint:
- `POST https://places.googleapis.com/v1/places:searchText`

## Database Changes

The Spot model has been enhanced with:

- `Status` field with values: `pending_verification`, `verified`, `rejected`
- `PlaceID` field for Google Places API identifier (unique)
- `LastSeen` timestamp for tracking rediscovery

### Performance Indexes

The following indexes have been added for optimal duplicate detection performance:

- `idx_spots_location`: Composite index on (latitude, longitude) for geospatial queries
- `idx_spots_name_status`: Index on (name, status) for name-based duplicate detection
- `idx_spots_status_source`: Index on (status, source) for filtering queries
- `idx_spots_active`: Partial index for active (non-deleted) spots
- `idx_spots_last_seen`: Index on last_seen for scraper tracking
- `idx_spots_user_source`: Composite index for user-specific queries

## Duplicate Prevention System

### Overview

The service now implements a sophisticated duplicate prevention system that goes beyond simple PlaceID matching:

### Detection Criteria

**Primary Duplicate Detection:**
- **Name Similarity**: Uses fuzzy string matching with Levenshtein distance
- **Geospatial Proximity**: 50-meter tolerance using Haversine distance calculation
- **Combined Scoring**: Spots are considered duplicates if they meet both criteria

**Advanced Detection Rules:**
1. **High Similarity + Close Distance**: Name similarity ≥ 80% within 50m
2. **Very Close Proximity**: Distance ≤ 25m with moderate name similarity (≥ 60%)
3. **Nearly Identical Names**: Name similarity ≥ 95% within 100m

### Text Processing

**Name Normalization:**
- Case-insensitive comparison
- Removal of common words ("amala", "spot", "restaurant", "mama", etc.)
- Special character normalization
- Whitespace standardization

**Fuzzy Matching Algorithm:**
- Levenshtein distance calculation (70% weight)
- Word overlap similarity (30% weight)
- Configurable similarity threshold (default: 0.8)

### Geospatial Calculations

**Distance Calculation:**
- Haversine formula for great-circle distance
- 50-meter duplicate tolerance (hardcoded)
- Lagos coordinate validation bounds
- Efficient bounding box pre-filtering

**Performance Optimization:**
- Bounding box queries reduce database load
- Precise Haversine calculation only on filtered results
- Database indexes on latitude/longitude coordinates

### Integration Points

**Places Service (`places_service.go`):**
- Enhanced `processDiscoveredPlace()` with duplicate checking
- Graceful handling of duplicate errors (continues processing)
- Comprehensive logging of duplicate detection decisions

**Spot Repository (`spot_repository.go`):**
- New `CreateSpotWithDuplicateCheck()` method for user submissions
- `FindSpotsNearLocation()` for geospatial queries
- `CheckForDuplicates()` for validation without creation
- `FindSimilarSpots()` for administrative purposes

**Spot Handlers (`spot_handlers.go`):**
- Updated `CreateSpot()` with duplicate prevention
- HTTP 409 Conflict response for detected duplicates
- Detailed error messages with duplicate information

### Logging and Monitoring

**Comprehensive Logging:**
```
DUPLICATE_DETECTED: spot='Mama Cass Amala' lat=6.524400 lng=3.379200 matches existing spot_id=123 distance=25.5m name_similarity=0.87 reason='Similar names (0.87 similarity) within 50m tolerance'

DUPLICATE_PREVENTED: User spot 'New Amala Spot' rejected - Similar names (0.82 similarity) within 50m tolerance (matches existing spot ID: 456)

PLACES_DUPLICATE: Skipping duplicate place from Google Places: 'Amala Joint' (place_abc123) - duplicate spot detected: Nearly identical names (0.96 similarity) within 100m (matches existing spot ID: 789)
```

**Error Handling:**
- Fail-safe approach: continues operation if duplicate check fails
- Detailed error messages for API consumers
- Distinction between duplicate errors and system errors

### Backward Compatibility

**Non-Breaking Changes:**
- All existing APIs remain functional
- PlaceID uniqueness constraint preserved
- Existing database schema unchanged
- New duplicate logic applies only to future spots

**Migration Strategy:**
- Automatic index creation during application startup
- No data migration required
- Gradual improvement of duplicate detection over time

## Error Handling

The service includes comprehensive error handling:
- API connection failures are logged but don't stop the application
- Individual place processing failures are logged but processing continues
- Database errors are properly handled and logged

## Monitoring

Check application logs for:
- Discovery start/end times
- Number of new spots discovered
- Number of existing spots updated
- Any errors encountered
- Total spots in verification queue