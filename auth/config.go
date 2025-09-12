package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/donny-c-1/amalajeun/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuth2 configuration
var GoogleOAuthConfig *oauth2.Config

// GoogleUser represents the user information returned by Google OAuth2
type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	GoogleID string `json:"google_id"`
}

// UserContextKey is used for storing user information in request context
type UserContextKey string

const UserKey UserContextKey = "user"

// InitOAuthConfig initializes the Google OAuth2 configuration
func InitOAuthConfig() error {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET environment variables are required")
	}

	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/google/callback"
	}

	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return nil
}

// GetGoogleUser fetches user information from Google using the access token
func GetGoogleUser(ctx context.Context, accessToken string) (*GoogleUser, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google API returned status: %d", resp.StatusCode)
	}

	var googleUser GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &googleUser, nil
}

// ConvertGoogleUserToModel converts GoogleUser to models.User
func ConvertGoogleUserToModel(googleUser *GoogleUser) *models.User {
	return &models.User{
		GoogleID:  googleUser.ID,
		Email:     googleUser.Email,
		Name:      googleUser.Name,
		AvatarURL: googleUser.Picture,
	}
}
