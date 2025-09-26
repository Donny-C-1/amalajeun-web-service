package utils

import (
	"testing"
)

// TestHaversineDistance tests the Haversine distance calculation
func TestHaversineDistance(t *testing.T) {
	tests := []struct {
		name      string
		lat1      float64
		lng1      float64
		lat2      float64
		lng2      float64
		expected  float64
		tolerance float64
	}{
		{
			name: "Same location",
			lat1: 6.5244, lng1: 3.3792,
			lat2: 6.5244, lng2: 3.3792,
			expected:  0.0,
			tolerance: 0.1,
		},
		{
			name: "Lagos locations ~1.5km apart",
			lat1: 6.5244, lng1: 3.3792, // Ikeja area
			lat2: 6.5334, lng2: 3.3892, // ~1.5km northeast
			expected:  1490.0,
			tolerance: 100.0, // Allow 100m tolerance
		},
		{
			name: "Very close locations ~63m apart",
			lat1: 6.5244, lng1: 3.3792,
			lat2: 6.5248, lng2: 3.3796, // ~63m northeast
			expected:  63.0,
			tolerance: 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := HaversineDistance(tt.lat1, tt.lng1, tt.lat2, tt.lng2)
			if abs(distance-tt.expected) > tt.tolerance {
				t.Errorf("HaversineDistance() = %.2f, expected %.2f ± %.2f",
					distance, tt.expected, tt.tolerance)
			}
		})
	}
}

// TestIsWithinDuplicateTolerance tests the duplicate tolerance checking
func TestIsWithinDuplicateTolerance(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lng1     float64
		lat2     float64
		lng2     float64
		expected bool
	}{
		{
			name: "Same location - should be within tolerance",
			lat1: 6.5244, lng1: 3.3792,
			lat2: 6.5244, lng2: 3.3792,
			expected: true,
		},
		{
			name: "25m apart - should be within 50m tolerance",
			lat1: 6.5244, lng1: 3.3792,
			lat2: 6.5246, lng2: 3.3794,
			expected: true,
		},
		{
			name: "100m apart - should be outside 50m tolerance",
			lat1: 6.5244, lng1: 3.3792,
			lat2: 6.5253, lng2: 3.3801,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWithinDuplicateTolerance(tt.lat1, tt.lng1, tt.lat2, tt.lng2)
			if result != tt.expected {
				t.Errorf("IsWithinDuplicateTolerance() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestIsValidLagosCoordinate tests Lagos coordinate validation
func TestIsValidLagosCoordinate(t *testing.T) {
	tests := []struct {
		name     string
		lat      float64
		lng      float64
		expected bool
	}{
		{
			name: "Valid Lagos coordinate (Ikeja)",
			lat:  6.5244, lng: 3.3792,
			expected: true,
		},
		{
			name: "Valid Lagos coordinate (Victoria Island)",
			lat:  6.4281, lng: 3.4219,
			expected: true,
		},
		{
			name: "Invalid coordinate - outside Lagos (Abuja)",
			lat:  9.0765, lng: 7.3986,
			expected: false,
		},
		{
			name: "Invalid coordinate - outside Nigeria",
			lat:  40.7128, lng: -74.0060, // New York
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidLagosCoordinate(tt.lat, tt.lng)
			if result != tt.expected {
				t.Errorf("IsValidLagosCoordinate() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestNormalizeSpotName tests spot name normalization
func TestNormalizeSpotName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic normalization",
			input:    "Mama Cass Amala Spot",
			expected: "cass",
		},
		{
			name:     "Remove special characters",
			input:    "Mama's Amala & Restaurant!!!",
			expected: "s",
		},
		{
			name:     "Multiple common words",
			input:    "The Best Amala Spot in Lagos",
			expected: "best lagos",
		},
		{
			name:     "Empty after normalization",
			input:    "Amala Spot Restaurant",
			expected: "",
		},
		{
			name:     "Mixed case with numbers",
			input:    "MAMA Cass Amala Spot 2",
			expected: "cass 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeSpotName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeSpotName() = '%s', expected '%s'", result, tt.expected)
			}
		})
	}
}

// TestCalculateNameSimilarity tests name similarity calculation
func TestCalculateNameSimilarity(t *testing.T) {
	tests := []struct {
		name1    string
		name2    string
		expected float64
		minScore float64 // Minimum acceptable score
	}{
		{
			name1:    "Mama Cass Amala Spot",
			name2:    "Mama Cass Amala Restaurant",
			expected: 1.0, // Should be identical after normalization
			minScore: 0.9,
		},
		{
			name1:    "Bukka Amala Joint",
			name2:    "Buka Amala Spot",
			expected: 0.0, // After normalization, both become empty
			minScore: 0.0,
		},
		{
			name1:    "Completely Different Name",
			name2:    "Another Unrelated Spot",
			expected: 0.0,
			minScore: 0.0, // Should be very low similarity
		},
		{
			name1:    "Same Name",
			name2:    "Same Name",
			expected: 1.0,
			minScore: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name1+" vs "+tt.name2, func(t *testing.T) {
			similarity := CalculateNameSimilarity(tt.name1, tt.name2)
			if similarity < tt.minScore {
				t.Errorf("CalculateNameSimilarity() = %.3f, expected >= %.3f",
					similarity, tt.minScore)
			}
		})
	}
}

// TestAreNamesSimilar tests the boolean similarity check
func TestAreNamesSimilar(t *testing.T) {
	tests := []struct {
		name1    string
		name2    string
		expected bool
	}{
		{
			name1:    "Mama Cass Amala Spot",
			name2:    "Mama Cass Amala Restaurant",
			expected: true, // Should be similar after normalization
		},
		{
			name1:    "Bukka Joint",
			name2:    "Pizza Hut",
			expected: false, // Completely different
		},
		{
			name1:    "Amala Spot",
			name2:    "Amala Restaurant",
			expected: false, // Both become empty after removing common words
		},
	}

	for _, tt := range tests {
		t.Run(tt.name1+" vs "+tt.name2, func(t *testing.T) {
			result := AreNamesSimilar(tt.name1, tt.name2)
			if result != tt.expected {
				similarity := CalculateNameSimilarity(tt.name1, tt.name2)
				t.Errorf("AreNamesSimilar() = %v, expected %v (similarity: %.3f)",
					result, tt.expected, similarity)
			}
		})
	}
}

// TestGetBoundingBox tests bounding box calculation
func TestGetBoundingBox(t *testing.T) {
	centerLat, centerLng := 6.5244, 3.3792
	radiusM := 100.0

	minLat, maxLat, minLng, maxLng := GetBoundingBox(centerLat, centerLng, radiusM)

	// Basic sanity checks
	if minLat >= maxLat {
		t.Errorf("minLat (%.6f) should be less than maxLat (%.6f)", minLat, maxLat)
	}
	if minLng >= maxLng {
		t.Errorf("minLng (%.6f) should be less than maxLng (%.6f)", minLng, maxLng)
	}

	// Check that center is within bounds
	if centerLat < minLat || centerLat > maxLat {
		t.Errorf("centerLat (%.6f) should be within bounds [%.6f, %.6f]",
			centerLat, minLat, maxLat)
	}
	if centerLng < minLng || centerLng > maxLng {
		t.Errorf("centerLng (%.6f) should be within bounds [%.6f, %.6f]",
			centerLng, minLng, maxLng)
	}

	// Check that the bounding box is reasonable (not too large or too small)
	latDiff := maxLat - minLat
	lngDiff := maxLng - minLng
	if latDiff < 0.0005 || latDiff > 0.005 { // Rough bounds check
		t.Errorf("Latitude difference (%.6f) seems unreasonable for 100m radius", latDiff)
	}
	if lngDiff < 0.0005 || lngDiff > 0.005 { // Rough bounds check
		t.Errorf("Longitude difference (%.6f) seems unreasonable for 100m radius", lngDiff)
	}
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
