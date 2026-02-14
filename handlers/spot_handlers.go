package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/donny-c-1/amalajeun/auth"
	"github.com/donny-c-1/amalajeun/database"
	"github.com/donny-c-1/amalajeun/models"
	"github.com/donny-c-1/amalajeun/services/cloudinary"
	"github.com/donny-c-1/amalajeun/services/places"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateSpot handles POST /spots - add a new Amala spot
// Now includes comprehensive duplicate prevention logic
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

	// Use the repository with duplicate checking instead of direct database creation
	spotRepo := places.NewSpotRepository()
	createdSpot, err := spotRepo.CreateSpotWithDuplicateCheck(&spot)
	if err != nil {
		// Check if this is a duplicate error
		if isDuplicateError(err) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Duplicate spot detected",
				"message": "A similar spot already exists in this location",
				"details": err.Error(),
			})
			return
		}

		// Other database errors
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create spot",
			"details": err.Error(),
		})
		return
	}

	// Load user information for response
	database.DB.Preload("User").First(createdSpot, createdSpot.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Spot created successfully",
		"data":    createdSpot,
	})
}

// isDuplicateError checks if an error is related to duplicate detection
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "duplicate spot detected") ||
		strings.Contains(errMsg, "similar spot exists") ||
		strings.Contains(errMsg, "matches existing spot")
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
		switch verified {
		case "true":
			query = query.Where("status = ?", models.StatusVerified)
		case "false":
			query = query.Where("status != ?", models.StatusVerified)
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
		switch verified {
		case "true":
			countQuery = countQuery.Where("status = ?", models.StatusVerified)
		case "false":
			countQuery = countQuery.Where("status != ?", models.StatusVerified)
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

// VerifySpot handles PATCH /spots/:id/verify - verify a spot
// New workflow: Requires 3 unique user verifications to mark spot as verified
// Each user can only verify a spot once (enforced by unique constraint)
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

	// Check if user has already verified this spot
	var existingVerification models.SpotVerification
	result := database.DB.Where("spot_id = ? AND user_id = ?", spotID, user.ID).First(&existingVerification)

	if result.Error == nil {
		// Verification already exists for this user and spot
		c.JSON(http.StatusConflict, gin.H{
			"error":   "You have already verified this spot",
			"message": "Each user can only verify a spot once",
		})
		return
	}

	// Create new verification record
	verification := models.SpotVerification{
		SpotID: uint(spotID),
		UserID: user.ID,
	}

	if err := database.DB.Create(&verification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to record verification",
			"details": err.Error(),
		})
		return
	}

	// Count total verifications for this spot
	var verificationCount int64
	if err := database.DB.Model(&models.SpotVerification{}).
		Where("spot_id = ?", spotID).
		Count(&verificationCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to count verifications",
			"details": err.Error(),
		})
		return
	}

	// Update spot status if we reach 3 verifications
	if verificationCount >= 3 {
		updates := map[string]interface{}{
			"status":     models.StatusVerified,
			"updated_at": time.Now(),
		}

		if err := database.DB.Model(&spot).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update spot status",
				"details": err.Error(),
			})
			return
		}
	}

	// Reload the spot to get updated data
	database.DB.Preload("User").First(&spot, uint(spotID))

	c.JSON(http.StatusOK, gin.H{
		"message":            "Verification recorded successfully",
		"data":               spot,
		"verification_count": verificationCount,
		"verified_by": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
		"spot_status": spot.Status,
	})
}

// UploadSpotImage handles POST /spots/:id/images - upload an image for a specific spot
func UploadSpotImage(c *gin.Context) {
	// Get authenticated user
	user, exists := auth.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required to upload images",
		})
		return
	}

	// Get spot ID from URL parameter
	id := c.Param("id")
	spotID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid spot ID",
		})
		return
	}

	// Check if spot exists
	var spot models.Spot
	if err := database.DB.First(&spot, uint(spotID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Spot not found",
		})
		return
	}

	// Get the uploaded file
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "No image file provided",
			"details": "Please provide an image file with the key 'image'",
		})
		return
	}
	defer file.Close()

	// Check if Cloudinary service is initialized
	if cloudinary.GlobalCloudinaryService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Image upload service not initialized",
		})
		return
	}

	// Upload image to Cloudinary using the global service instance
	imageURL, err := cloudinary.GlobalCloudinaryService.UploadImage(c.Request.Context(), file, header, uint(spotID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to upload image",
			"details": err.Error(),
		})
		return
	}

	// Add the new image URL to the spot's images array
	spot.Images = append(spot.Images, imageURL)
	spot.UpdatedAt = time.Now()
	if spot.PlaceID == ""{
		spot.PlaceID = uuid.New().String() // Ensure PlaceID is set
	}

	// Update the spot in the database
	if err := database.DB.Save(&spot).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save image URL to database",
			"details": err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message":      "Image uploaded successfully",
		"image_url":    imageURL,
		"spot_id":      spotID,
		"total_images": len(spot.Images),
		"uploaded_by": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}
