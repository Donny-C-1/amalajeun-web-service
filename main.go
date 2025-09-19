package main

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"os/signal"

	"github.com/donny-c-1/amalajeun/auth"
	"github.com/donny-c-1/amalajeun/database"
	"github.com/donny-c-1/amalajeun/routes"
	"github.com/donny-c-1/amalajeun/services/places"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
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
		log.Fatalln("PORT must be set in env")
	}

	// Start the Cron Job for places discovery
	cronjob := cron.New()

	// Run immediately on startup and then daily
	placesService := places.NewPlacesService()

	// Run immediately on startup
	log.Println("Running initial places discovery on startup...")
	if err := placesService.RunPlacesDiscovery(); err != nil {
		log.Printf("Initial places discovery failed: %v", err)
	} else {
		log.Println("Initial places discovery completed successfully")
	}

	// Schedule daily runs
	_, err := cronjob.AddFunc("@daily", func() {
		log.Println("Starting scheduled daily places discovery...")
		if err := placesService.RunPlacesDiscovery(); err != nil {
			log.Printf("Daily places discovery failed: %v", err)
		} else {
			log.Println("Daily places discovery completed successfully")
		}
	})
	if err != nil {
		log.Fatalln("Error scheduling discovery job:", err)
	}

	cronjob.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

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

	// Run Gin in a subroutine
	go func() {
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Block until signal received
	<-quit

	fmt.Println("Stopping Cron Job")
	ctx := cronjob.Stop()
	<-ctx.Done()
	fmt.Println("Cron Job Stopped")
}
