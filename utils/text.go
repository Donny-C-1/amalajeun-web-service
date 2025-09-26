package utils

import (
	"regexp"
	"strings"
	"unicode"
)

const (
	// NameSimilarityThreshold represents the minimum similarity score (0.0-1.0)
	// for two spot names to be considered similar enough to be duplicates
	NameSimilarityThreshold = 0.8
)

var (
	// Common words to remove from spot names for better comparison
	// These words are frequently used in restaurant names but don't add uniqueness
	commonWords = []string{
		"amala", "spot", "restaurant", "eatery", "place", "joint",
		"mama", "buka", "canteen", "kitchen", "food", "house",
		"the", "and", "&", "n", "of", "at", "in", "on",
	}

	// Regex for cleaning text - removes special characters and extra spaces
	cleanTextRegex = regexp.MustCompile(`[^\p{L}\p{N}\s]+`)
	spaceRegex     = regexp.MustCompile(`\s+`)
)

// NormalizeSpotName preprocesses a spot name for comparison by:
// - Converting to lowercase
// - Removing special characters
// - Removing common words that don't add uniqueness
// - Trimming and normalizing whitespace
func NormalizeSpotName(name string) string {
	if name == "" {
		return ""
	}

	// Convert to lowercase and remove special characters
	normalized := strings.ToLower(name)
	normalized = cleanTextRegex.ReplaceAllString(normalized, " ")
	normalized = spaceRegex.ReplaceAllString(normalized, " ")
	normalized = strings.TrimSpace(normalized)

	// Split into words and filter out common words
	words := strings.Fields(normalized)
	var filteredWords []string

	for _, word := range words {
		if !isCommonWord(word) && len(word) >= 1 {
			filteredWords = append(filteredWords, word)
		}
	}

	return strings.Join(filteredWords, " ")
}

// isCommonWord checks if a word is in the common words list
func isCommonWord(word string) bool {
	for _, common := range commonWords {
		if word == common {
			return true
		}
	}
	return false
}

// CalculateNameSimilarity calculates similarity between two spot names using
// a combination of normalized Levenshtein distance and word overlap.
// Returns a score between 0.0 (completely different) and 1.0 (identical)
func CalculateNameSimilarity(name1, name2 string) float64 {
	if name1 == "" || name2 == "" {
		return 0.0
	}

	// Normalize both names
	norm1 := NormalizeSpotName(name1)
	norm2 := NormalizeSpotName(name2)

	// If either normalized name is empty, return 0
	if norm1 == "" || norm2 == "" {
		return 0.0
	}

	// If normalized names are identical, return 1
	if norm1 == norm2 {
		return 1.0
	}

	// Calculate Levenshtein distance
	levenshteinSimilarity := calculateLevenshteinSimilarity(norm1, norm2)

	// Calculate word overlap similarity
	wordOverlapSimilarity := calculateWordOverlapSimilarity(norm1, norm2)

	// Combine both metrics with weights
	// Levenshtein is better for character-level differences
	// Word overlap is better for word order differences
	combinedSimilarity := (levenshteinSimilarity * 0.7) + (wordOverlapSimilarity * 0.3)

	return combinedSimilarity
}

// calculateLevenshteinSimilarity calculates normalized Levenshtein similarity
func calculateLevenshteinSimilarity(s1, s2 string) float64 {
	distance := levenshteinDistance(s1, s2)
	maxLen := max(len(s1), len(s2))

	if maxLen == 0 {
		return 1.0
	}

	return 1.0 - (float64(distance) / float64(maxLen))
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	r1 := []rune(s1)
	r2 := []rune(s2)

	len1 := len(r1)
	len2 := len(r2)

	// Create a matrix to store distances
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	// Initialize first row and column
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	// Fill the matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				min(matrix[i-1][j]+1, matrix[i][j-1]+1), // deletion, insertion
				matrix[i-1][j-1]+cost,                   // substitution
			)
		}
	}

	return matrix[len1][len2]
}

// calculateWordOverlapSimilarity calculates similarity based on word overlap
func calculateWordOverlapSimilarity(s1, s2 string) float64 {
	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	// Count overlapping words
	wordSet1 := make(map[string]bool)
	for _, word := range words1 {
		wordSet1[word] = true
	}

	overlap := 0
	for _, word := range words2 {
		if wordSet1[word] {
			overlap++
		}
	}

	// Calculate Jaccard similarity (intersection / union)
	union := len(words1) + len(words2) - overlap
	if union == 0 {
		return 1.0
	}

	return float64(overlap) / float64(union)
}

// AreNamesSimilar checks if two spot names are similar enough to be considered duplicates
func AreNamesSimilar(name1, name2 string) bool {
	similarity := CalculateNameSimilarity(name1, name2)
	return similarity >= NameSimilarityThreshold
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RemoveAccents removes accents and diacritics from text for better matching
func RemoveAccents(s string) string {
	var result strings.Builder
	for _, r := range s {
		// Convert accented characters to their base form
		switch {
		case unicode.Is(unicode.Mn, r): // Nonspacing marks
			// Skip combining characters
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}
