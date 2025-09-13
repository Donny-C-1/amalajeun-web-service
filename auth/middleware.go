package auth

import (
	"net/http"
	"strings"

	"github.com/donny-c-1/amalajeun/database"
	"github.com/donny-c-1/amalajeun/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthMiddleware validates JWT tokens and adds user context to requests
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization header required",
				"message": "Please provide a valid JWT token in the Authorization header",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization header",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate token and extract user info
		user, err := ExtractUserFromToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid or expired token",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Verify user exists in database
		var dbUser models.User
		if err := database.DB.Where("id = ?", user.ID).First(&dbUser).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "User not found",
				"message": "The user associated with this token no longer exists",
			})
			c.Abort()
			return
		}

		// Add user to request context
		c.Set(string(UserKey), &dbUser)
		c.Next()
	}
}

// OptionalAuthMiddleware allows requests without authentication but adds user context if token is provided
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			// Don't abort, just continue without user context
			c.Next()
			return
		}

		// Validate token and extract user info
		user, err := ExtractUserFromToken(tokenString)
		if err != nil {
			// Don't abort, just continue without user context
			c.Next()
			return
		}

		// Verify user exists in database
		var dbUser models.User
		if err := database.DB.Where("id = ?", user.ID).First(&dbUser).Error; err != nil {
			// Don't abort, just continue without user context
			c.Next()
			return
		}

		// Add user to request context
		c.Set(string(UserKey), &dbUser)
		c.Next()
	}
}

// GetUserFromContext extracts user from Gin context
func GetUserFromContext(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get(string(UserKey))
	if !exists {
		return nil, false
	}

	userPtr, ok := user.(*models.User)
	if !ok {
		return nil, false
	}

	return userPtr, true
}

// RequireUserID middleware ensures the user ID in the URL matches the authenticated user
// Useful for user-specific operations like updating their own profile
func RequireUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := GetUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		userIDParam := c.Param("userId")
		if userIDParam == "" {
			c.Next()
			return
		}

		userID, err := uuid.Parse(userIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user ID format",
			})
			c.Abort()
			return
		}

		if user.ID != userID {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied",
				"message": "You can only access your own resources",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMiddleware handles CORS for frontend requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Allow specific origins
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
			"https://amalajeun.vercel.app",
		}

		// Check if origin is allowed
		allowOrigin := ""
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowOrigin = origin
				break
			}
		}

		// If origin is not in allowed list, allow it (for development flexibility)
		if allowOrigin == "" && strings.HasPrefix(origin, "http://localhost:") {
			allowOrigin = origin
		}

		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
