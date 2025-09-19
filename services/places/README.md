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
- **Duplicate Prevention**: Uses `place_id` to avoid duplicate entries
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