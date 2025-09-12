package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/donny-c-1/amalajeun/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWT secret key
var jwtSecret []byte

// InitJWT initializes the JWT secret key
func InitJWT() error {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}
	jwtSecret = []byte(secret)
	return nil
}

// GenerateToken generates a JWT token for a user
func GenerateToken(user *models.User) (string, error) {
	// Set token expiry (24 hours by default)
	expiryHours := 24
	if expiryStr := os.Getenv("JWT_EXPIRY_HOURS"); expiryStr != "" {
		if hours, err := time.ParseDuration(expiryStr + "h"); err == nil {
			expiryHours = int(hours.Hours())
		}
	}

	claims := jwt.MapClaims{
		"sub":       user.ID.String(),
		"email":     user.Email,
		"name":      user.Name,
		"google_id": user.GoogleID,
		"exp":       time.Now().Add(time.Hour * time.Duration(expiryHours)).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ExtractUserFromToken extracts user information from JWT token
func ExtractUserFromToken(tokenString string) (*models.User, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	googleID, _ := claims["google_id"].(string)

	return &models.User{
		ID:       userID,
		Email:    email,
		Name:     name,
		GoogleID: googleID,
	}, nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return authHeader[7:], nil
}
