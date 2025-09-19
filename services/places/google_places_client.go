package places

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// GooglePlace represents a place from Google Places API
type GooglePlace struct {
	PlaceID   string  `json:"place_id"`
	Name      string  `json:"name"`
	Address   string  `json:"formatted_address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	MapsURL   string  `json:"maps_url"`
}

// Modern Google Places API response structures
type ModernGooglePlacesResponse struct {
	Places []struct {
		ID          string `json:"id"`
		DisplayName struct {
			Text string `json:"text"`
		} `json:"displayName"`
		FormattedAddress string `json:"formattedAddress"`
		Location         struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location"`
	} `json:"places"`
}

// SearchTextRequest represents the request body for the modern Google Places API
type SearchTextRequest struct {
	TextQuery string `json:"textQuery"`
}

// GooglePlacesClient handles interactions with Google Places API
type GooglePlacesClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewGooglePlacesClient creates a new Google Places API client
func NewGooglePlacesClient() *GooglePlacesClient {
	apiKey := os.Getenv("GOOGLE_PLACES_API_KEY")
	if apiKey == "" {
		log.Println("WARNING: GOOGLE_PLACES_API_KEY environment variable not set")
	}

	return &GooglePlacesClient{
		apiKey:  apiKey,
		baseURL: "https://places.googleapis.com/v1/places:searchText",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchAmalaInLagos searches for Amala spots in Lagos using the modern Google Places API
func (c *GooglePlacesClient) SearchAmalaInLagos() ([]GooglePlace, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("google Places API key not configured")
	}

	// Prepare request body
	requestBody := SearchTextRequest{
		TextQuery: "Amala in Lagos, Nigeria",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create POST request
	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", c.apiKey)
	req.Header.Set("X-Goog-FieldMask", "places.id,places.displayName,places.location,places.formattedAddress")

	// Make HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make Google Places API request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google Places API returned non-200 status: %d - %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var placesResponse ModernGooglePlacesResponse
	if err := json.Unmarshal(body, &placesResponse); err != nil {
		return nil, fmt.Errorf("failed to parse Google Places response: %w", err)
	}

	// Convert to our GooglePlace format
	var places []GooglePlace
	for _, result := range placesResponse.Places {
		place := GooglePlace{
			PlaceID:   result.ID,
			Name:      result.DisplayName.Text,
			Address:   result.FormattedAddress,
			Latitude:  result.Location.Latitude,
			Longitude: result.Location.Longitude,
			MapsURL:   fmt.Sprintf("https://www.google.com/maps/place/?q=place_id:%s", result.ID),
		}
		places = append(places, place)
	}

	log.Printf("Found %d Amala spots in Lagos via modern Google Places API", len(places))
	return places, nil
}
