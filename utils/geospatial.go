package utils

import (
	"math"
)

const (
	// EarthRadiusKm represents Earth's radius in kilometers
	EarthRadiusKm = 6371.0

	// DuplicateDistanceToleranceM represents the distance tolerance in meters for duplicate detection
	// Two spots within this distance with similar names are considered duplicates
	DuplicateDistanceToleranceM = 50.0

	// Lagos approximate bounds for validation
	LagosMinLat = 6.4
	LagosMaxLat = 6.7
	LagosMinLng = 3.2
	LagosMaxLng = 3.6
)

// HaversineDistance calculates the great-circle distance between two points
// on Earth given their latitude and longitude coordinates using the Haversine formula.
// Returns distance in meters.
func HaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180

	// Calculate differences
	deltaLat := lat2Rad - lat1Rad
	deltaLng := lng2Rad - lng1Rad

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Distance in kilometers, convert to meters
	distanceKm := EarthRadiusKm * c
	return distanceKm * 1000
}

// IsWithinDuplicateTolerance checks if two coordinates are within the duplicate detection tolerance
func IsWithinDuplicateTolerance(lat1, lng1, lat2, lng2 float64) bool {
	distance := HaversineDistance(lat1, lng1, lat2, lng2)
	return distance <= DuplicateDistanceToleranceM
}

// IsValidLagosCoordinate validates if coordinates are within Lagos bounds
// This helps prevent obviously invalid coordinates from being processed
func IsValidLagosCoordinate(lat, lng float64) bool {
	return lat >= LagosMinLat && lat <= LagosMaxLat &&
		lng >= LagosMinLng && lng <= LagosMaxLng
}

// GetBoundingBox calculates a bounding box around a point for efficient database queries
// Returns min/max latitude and longitude for a square around the point
// This is used to pre-filter database results before applying Haversine distance
func GetBoundingBox(centerLat, centerLng, radiusM float64) (minLat, maxLat, minLng, maxLng float64) {
	// Convert radius from meters to degrees (approximate)
	// 1 degree latitude ≈ 111,320 meters
	// 1 degree longitude ≈ 111,320 * cos(latitude) meters
	latDelta := radiusM / 111320.0
	lngDelta := radiusM / (111320.0 * math.Cos(centerLat*math.Pi/180))

	minLat = centerLat - latDelta
	maxLat = centerLat + latDelta
	minLng = centerLng - lngDelta
	maxLng = centerLng + lngDelta

	return minLat, maxLat, minLng, maxLng
}
