package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/donny-c-1/amalajeun/auth"
	"github.com/donny-c-1/amalajeun/database"
	"github.com/donny-c-1/amalajeun/models"
	"github.com/gin-gonic/gin"
)

// CreateSpot handles POST /spots - add a new Amala spot
func CreateSpot(c *gin.Context) {
	// Get authenticated user
	user, exists := auth.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required to create spots",
		})
		return
	}

	var spot models.Spot

	// Bind JSON request to spot struct
	if err := c.ShouldBindJSON(&spot); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Validate source enum
	if spot.Source != models.SourceUser && spot.Source != models.SourceAgent && spot.Source != models.SourceScraper {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid source. Must be one of: user, agent, scraper",
		})
		return
	}

	// Set user ID and keep backward compatibility with AddedBy
	spot.UserID = &user.ID
	spot.AddedBy = user.Name // Keep for backward compatibility
	spot.CreatedAt = time.Now()
	spot.UpdatedAt = time.Now()

	// Create spot in database
	if err := database.DB.Create(&spot).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create spot",
			"details": err.Error(),
		})
		return
	}

	// Load user information for response
	database.DB.Preload("User").First(&spot, spot.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Spot created successfully",
		"data":    spot,
	})
}

// GetSpots handles GET /spots - list all spots
func GetSpots(c *gin.Context) {
	var spots []models.Spot

	// Get query parameters for filtering/pagination
	verified := c.Query("verified")
	source := c.Query("source")
	limit := c.DefaultQuery("limit", "50")
	offset := c.DefaultQuery("offset", "0")

	// Convert limit and offset to integers
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		limitInt = 50
	}
	if limitInt > 100 {
		limitInt = 100 // Cap at 100 for performance
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil || offsetInt < 0 {
		offsetInt = 0
	}

	// Build query
	query := database.DB.Model(&models.Spot{})

	// Apply filters
	if verified != "" {
		if verified == "true" {
			query = query.Where("verified = ?", true)
		} else if verified == "false" {
			query = query.Where("verified = ?", false)
		}
	}

	if source != "" {
		query = query.Where("source = ?", source)
	}

	// Execute query with pagination
	if err := query.Limit(limitInt).Offset(offsetInt).Find(&spots).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch spots",
			"details": err.Error(),
		})
		return
	}

	// Get total count for pagination info
	var total int64
	countQuery := database.DB.Model(&models.Spot{})
	if verified != "" {
		if verified == "true" {
			countQuery = countQuery.Where("verified = ?", true)
		} else if verified == "false" {
			countQuery = countQuery.Where("verified = ?", false)
		}
	}
	if source != "" {
		countQuery = countQuery.Where("source = ?", source)
	}
	countQuery.Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"data": spots,
		"pagination": gin.H{
			"total":  total,
			"limit":  limitInt,
			"offset": offsetInt,
		},
	})
}

// GetSpot handles GET /spots/:id - fetch a single spot
func GetSpot(c *gin.Context) {
	id := c.Param("id")

	// Convert id to uint
	spotID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid spot ID",
		})
		return
	}

	var spot models.Spot

	// Find spot with reviews
	if err := database.DB.Preload("Reviews").First(&spot, uint(spotID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Spot not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": spot,
	})
}

// VerifySpot handles PATCH /spots/:id/verify - mark a spot as verified
func VerifySpot(c *gin.Context) {
	// Get authenticated user
	user, exists := auth.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required to verify spots",
		})
		return
	}

	id := c.Param("id")

	// Convert id to uint
	spotID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid spot ID",
		})
		return
	}

	var spot models.Spot

	// Find the spot first
	if err := database.DB.First(&spot, uint(spotID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Spot not found",
		})
		return
	}

	// Update verified status and updated timestamp
	updates := map[string]interface{}{
		"verified":   true,
		"updated_at": time.Now(),
	}

	if err := database.DB.Model(&spot).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to verify spot",
			"details": err.Error(),
		})
		return
	}

	// Reload the spot to get updated data
	database.DB.Preload("User").First(&spot, uint(spotID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Spot verified successfully",
		"data":    spot,
		"verified_by": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}
