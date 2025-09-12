package routes

import (
	"github.com/donny-c-1/amalajeun/handlers"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine) {
	// API version 1 group
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "Amala Jeun API is running",
				"version": "1.0.0",
			})
		})

		// Spot routes
		spots := v1.Group("/spots")
		{
			spots.POST("", handlers.CreateSpot)             // POST /api/v1/spots
			spots.GET("", handlers.GetSpots)                // GET /api/v1/spots
			spots.GET("/:id", handlers.GetSpot)             // GET /api/v1/spots/:id
			spots.PATCH("/:id/verify", handlers.VerifySpot) // PATCH /api/v1/spots/:id/verify
		}

		// Review routes
		reviews := v1.Group("/reviews")
		{
			reviews.POST("", handlers.CreateReview)            // POST /api/v1/reviews
			reviews.GET("/:spotId", handlers.GetReviewsBySpot) // GET /api/v1/reviews/:spotId
		}
	}

	// Root level routes (for backward compatibility if needed)
	router.POST("/spots", handlers.CreateSpot)
	router.GET("/spots", handlers.GetSpots)
	router.GET("/spots/:id", handlers.GetSpot)
	router.PATCH("/spots/:id/verify", handlers.VerifySpot)
	router.POST("/reviews", handlers.CreateReview)
	router.GET("/reviews/:spotId", handlers.GetReviewsBySpot)
}
