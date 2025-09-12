package routes

import (
	"github.com/donny-c-1/amalajeun/auth"
	"github.com/donny-c-1/amalajeun/handlers"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine) {
	// Home route
	router.GET("/", handlers.HomeHandler)

	// API version 1 group
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint
		v1.GET("/health", handlers.AuthHealth) // Updated to show auth status

		// Authentication routes
		authGroup := v1.Group("/auth")
		{
			authGroup.GET("/google", handlers.GoogleLogin)                        // GET /api/v1/auth/google
			authGroup.GET("/google/callback", handlers.GoogleCallback)            // GET /api/v1/auth/google/callback
			authGroup.GET("/profile", auth.AuthMiddleware(), handlers.GetProfile) // GET /api/v1/auth/profile (protected)
			authGroup.POST("/logout", handlers.Logout)                            // POST /api/v1/auth/logout
		}

		// Public routes (no authentication required)
		spots := v1.Group("/spots")
		{
			spots.GET("", handlers.GetSpots)    // GET /api/v1/spots
			spots.GET("/:id", handlers.GetSpot) // GET /api/v1/spots/:id
		}

		reviews := v1.Group("/reviews")
		{
			reviews.GET("/:spotId", handlers.GetReviewsBySpot) // GET /api/v1/reviews/:spotId
		}

		// Protected routes (authentication required)
		protected := v1.Group("")
		protected.Use(auth.AuthMiddleware())
		{
			// Protected spot routes
			protected.POST("/spots", handlers.CreateSpot)             // POST /api/v1/spots
			protected.PATCH("/spots/:id/verify", handlers.VerifySpot) // PATCH /api/v1/spots/:id/verify

			// Protected review routes
			protected.POST("/reviews", handlers.CreateReview) // POST /api/v1/reviews
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
