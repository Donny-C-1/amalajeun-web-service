package handlers

import (
	"net/http"
	"strconv"

	"github.com/donny-c-1/amalajeun/database"
	"github.com/donny-c-1/amalajeun/models"
	"github.com/gin-gonic/gin"
)

// CreateReview handles POST /reviews - add a review for a spot
func CreateReview(c *gin.Context) {
	var review models.Review

	// Bind JSON request to review struct
	if err := c.ShouldBindJSON(&review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Validate rating range (should be handled by binding, but double-check)
	if review.Rating < 1 || review.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rating must be between 1 and 5",
		})
		return
	}

	// Check if the spot exists
	var spot models.Spot
	if err := database.DB.First(&spot, review.SpotID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Spot not found",
		})
		return
	}

	// Create review in database
	if err := database.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create review",
			"details": err.Error(),
		})
		return
	}

	// Load the spot information for the response
	database.DB.Preload("Spot").First(&review, review.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Review created successfully",
		"data":    review,
	})
}

// GetReviewsBySpot handles GET /reviews/:spotId - list reviews for a given spot
func GetReviewsBySpot(c *gin.Context) {
	spotID := c.Param("spotId")

	// Convert spotId to uint
	spotIDUint, err := strconv.ParseUint(spotID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid spot ID",
		})
		return
	}

	// Check if the spot exists
	var spot models.Spot
	if err := database.DB.First(&spot, uint(spotIDUint)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Spot not found",
		})
		return
	}

	// Get query parameters for pagination and sorting
	limit := c.DefaultQuery("limit", "20")
	offset := c.DefaultQuery("offset", "0")
	sortBy := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	// Convert limit and offset to integers
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		limitInt = 20
	}
	if limitInt > 50 {
		limitInt = 50 // Cap at 50 for performance
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil || offsetInt < 0 {
		offsetInt = 0
	}

	// Validate sort parameters
	validSortFields := map[string]bool{
		"created_at": true,
		"rating":     true,
		"user_name":  true,
	}
	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	if order != "asc" && order != "desc" {
		order = "desc"
	}

	var reviews []models.Review

	// Build and execute query
	query := database.DB.Where("spot_id = ?", uint(spotIDUint)).
		Order(sortBy + " " + order).
		Limit(limitInt).
		Offset(offsetInt)

	if err := query.Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch reviews",
			"details": err.Error(),
		})
		return
	}

	// Get total count for pagination info
	var total int64
	database.DB.Model(&models.Review{}).Where("spot_id = ?", uint(spotIDUint)).Count(&total)

	// Calculate average rating
	var avgRating float64
	database.DB.Model(&models.Review{}).
		Where("spot_id = ?", uint(spotIDUint)).
		Select("AVG(rating)").
		Scan(&avgRating)

	c.JSON(http.StatusOK, gin.H{
		"data": reviews,
		"spot": gin.H{
			"id":             spot.ID,
			"name":           spot.Name,
			"average_rating": avgRating,
		},
		"pagination": gin.H{
			"total":  total,
			"limit":  limitInt,
			"offset": offsetInt,
		},
	})
}
