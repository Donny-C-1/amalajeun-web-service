package database

import (
	"fmt"
	"log"
	"os"

	"github.com/donny-c-1/amalajeun/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GetDefaultConfig returns default database configuration
// In production, these should come from environment variables
func GetDefaultConfig() Config {
	return Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "amalajeun"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Connect establishes database connection
func Connect() error {
	connStr := os.Getenv("DATABASE_URL")

	var err error
	DB, err = gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established successfully")
	return nil
}

// Migrate runs auto-migration for all models
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database connection not established")
	}

	err := DB.AutoMigrate(
		&models.User{}, // Add User model first for foreign key relationships
		&models.Spot{},
		&models.Review{},
		&models.SpotVerification{}, // New: Track spot verifications for the 3-user verification workflow
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")

	// Apply performance indexes for duplicate prevention
	if err := ApplyDuplicatePreventionIndexes(); err != nil {
		log.Printf("Warning: Failed to apply duplicate prevention indexes: %v", err)
		// Don't fail the migration if indexes fail - they're performance optimizations
	}

	return nil
}

// ApplyDuplicatePreventionIndexes creates database indexes for optimal duplicate detection performance
func ApplyDuplicatePreventionIndexes() error {
	if DB == nil {
		return fmt.Errorf("database connection not established")
	}

	log.Println("Applying duplicate prevention performance indexes...")

	// Index for geospatial queries (latitude, longitude)
	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_spots_location
		ON spots (latitude, longitude)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create location index: %w", err)
	}

	// Index for name-based searches combined with status filtering
	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_spots_name_status
		ON spots (name, status)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create name-status index: %w", err)
	}

	// Index for status and source filtering
	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_spots_status_source
		ON spots (status, source)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create status-source index: %w", err)
	}

	// Partial index for active spots (non-deleted)
	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_spots_active
		ON spots (id, created_at)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create active spots index: %w", err)
	}

	// Index for last_seen timestamp queries
	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_spots_last_seen
		ON spots (last_seen DESC)
		WHERE deleted_at IS NULL AND last_seen IS NOT NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create last_seen index: %w", err)
	}

	// Composite index for user-specific queries
	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_spots_user_source
		ON spots (user_id, source)
		WHERE deleted_at IS NULL AND user_id IS NOT NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create user-source index: %w", err)
	}

	log.Println("Duplicate prevention indexes applied successfully")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
