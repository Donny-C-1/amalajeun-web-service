package main

import (
	"fmt"
	"log"
	"os"

	"github.com/donny-c-1/amalajeun/auth"
	"github.com/donny-c-1/amalajeun/database"
	"github.com/donny-c-1/amalajeun/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load env vars
	if err := godotenv.Load(); err != nil {
		fmt.Println(err)
	}

	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize database connection
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run database migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize authentication system
	if err := auth.InitOAuthConfig(); err != nil {
		log.Fatalf("Failed to initialize OAuth configuration: %v", err)
	}

	if err := auth.InitJWT(); err != nil {
		log.Fatalf("Failed to initialize JWT: %v", err)
	}

	// Initialize Gin router
	router := gin.Default()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add CORS middleware for frontend
	router.Use(auth.CORSMiddleware())

	// Setup routes
	routes.SetupRoutes(router)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Starting Amala Jeun API server on port %s", port)
	log.Printf("Health check available at: http://localhost:%s/api/v1/health", port)
	log.Printf("API documentation:")
	log.Printf("  POST   /spots                 - Create a new spot")
	log.Printf("  GET    /spots                 - List all spots")
	log.Printf("  GET    /spots/:id             - Get a specific spot")
	log.Printf("  PATCH  /spots/:id/verify      - Verify a spot")
	log.Printf("  POST   /reviews               - Create a review")
	log.Printf("  GET    /reviews/:spotId       - Get reviews for a spot")

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
