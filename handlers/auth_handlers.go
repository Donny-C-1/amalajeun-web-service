package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/donny-c-1/amalajeun/auth"
	"github.com/donny-c-1/amalajeun/database"
	"github.com/donny-c-1/amalajeun/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// OAuth state store (in production, use Redis or database)
var oauthStates = make(map[string]time.Time)

// generateState generates a random state parameter for OAuth2
func generateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// cleanExpiredStates removes expired OAuth states
func cleanExpiredStates() {
	now := time.Now()
	for state, expiry := range oauthStates {
		if now.After(expiry) {
			delete(oauthStates, state)
		}
	}
}

// GoogleLogin initiates the Google OAuth2 login flow
func GoogleLogin(c *gin.Context) {
	// Generate state parameter for CSRF protection
	state, err := generateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate OAuth state",
		})
		return
	}

	// Store state with expiry (10 minutes)
	oauthStates[state] = time.Now().Add(10 * time.Minute)

	// Clean expired states periodically
	cleanExpiredStates()

	// Generate OAuth2 URL
	url := auth.GoogleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": url,
		"message":  "Redirect user to this URL to authenticate with Google",
	})
}

// GoogleCallback handles the OAuth2 callback from Google
func GoogleCallback(c *gin.Context) {
	// Get authorization code from query parameters
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Authorization code is required",
		})
		return
	}

	// Verify state parameter
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "State parameter is required",
		})
		return
	}

	// Check if state exists and is not expired
	expiry, exists := oauthStates[state]
	if !exists || time.Now().After(expiry) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid or expired state parameter",
		})
		return
	}

	// Remove used state
	delete(oauthStates, state)

	// Exchange authorization code for access token
	ctx := context.Background()
	token, err := auth.GoogleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to exchange authorization code",
			"details": err.Error(),
		})
		return
	}

	// Get user information from Google
	googleUser, err := auth.GetGoogleUser(ctx, token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user information from Google",
			"details": err.Error(),
		})
		return
	}

	// Check if user exists in database
	var user models.User
	result := database.DB.Where("google_id = ?", googleUser.ID).First(&user)

	if result.Error != nil {
		// User doesn't exist, create new user
		user = *auth.ConvertGoogleUserToModel(googleUser)
		now := time.Now()
		user.LastLoginAt = &now

		if err := database.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create user",
				"details": err.Error(),
			})
			return
		}
	} else {
		// User exists, update last login time
		now := time.Now()
		if err := database.DB.Model(&user).Update("last_login_at", now).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update user login time",
				"details": err.Error(),
			})
			return
		}
	}

	// Generate JWT token
	jwtToken, err := auth.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate authentication token",
			"details": err.Error(),
		})
		return
	}

	c.SetCookie(
		"amalajeun_token",
		jwtToken,
		86400, // 1day
		"/",
		"https://amalajeun.vercel.app",
		true,
		false,
	)
	c.Redirect(http.StatusFound, "https://amalajeun.vercel.app/map")
}

// GetProfile returns the current user's profile information
func GetProfile(c *gin.Context) {
	user, exists := auth.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":            user.ID,
			"email":         user.Email,
			"name":          user.Name,
			"avatar_url":    user.AvatarURL,
			"google_id":     user.GoogleID,
			"created_at":    user.CreatedAt,
			"last_login_at": user.LastLoginAt,
		},
	})
}

// Logout handles user logout (client-side token removal)
func Logout(c *gin.Context) {
	// In a stateless JWT system, logout is handled client-side by removing the token
	// For enhanced security, you could implement token blacklisting in Redis/database
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
		"note":    "Please remove the JWT token from client storage",
	})
}

// Health check with authentication info
func AuthHealth(c *gin.Context) {
	user, authenticated := auth.GetUserFromContext(c)

	response := gin.H{
		"status":  "ok",
		"service": "Amala Jeun Auth API",
		"version": "1.0.0",
	}

	if authenticated {
		response["user"] = gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		}
		response["authenticated"] = true
	} else {
		response["authenticated"] = false
	}

	c.JSON(http.StatusOK, response)
}
