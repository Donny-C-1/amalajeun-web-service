package cloudinary

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// CloudinaryService handles image uploads to Cloudinary
type CloudinaryService struct {
	client *cloudinary.Cloudinary
}

// NewCloudinaryService creates a new Cloudinary service instance
func NewCloudinaryService() (*CloudinaryService, error) {
	// Get Cloudinary URL from environment variable
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	if cloudinaryURL == "" {
		return nil, fmt.Errorf("CLOUDINARY_URL environment variable is required")
	}

	// Initialize Cloudinary client
	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary client: %w", err)
	}

	return &CloudinaryService{
		client: cld,
	}, nil
}

// UploadImage uploads an image file to Cloudinary and returns the secure URL
func (cs *CloudinaryService) UploadImage(ctx context.Context, file multipart.File, header *multipart.FileHeader, spotID uint) (string, error) {
	// Validate file type
	if !cs.isValidImageType(header.Filename) {
		return "", fmt.Errorf("invalid file type. Only JPEG, PNG, GIF, and WebP images are allowed")
	}

	// Validate file size (5MB limit)
	const maxSize = 5 * 1024 * 1024 // 5MB in bytes
	if header.Size > maxSize {
		return "", fmt.Errorf("file size exceeds 5MB limit")
	}

	// Generate a unique public ID for the image
	publicID := fmt.Sprintf("amalajeun/spots/%d/%s", spotID, cs.generateUniqueFilename(header.Filename))

	// Upload parameters
	uploadParams := uploader.UploadParams{
		PublicID:     publicID,
		Folder:       "amalajeun/spots",
		ResourceType: "image",
		Tags:         []string{"spot", fmt.Sprintf("spot_%d", spotID)},
		Context: map[string]string{
			"spot_id": fmt.Sprintf("%d", spotID),
		},
		Transformation: "c_limit,w_1920,h_1080,q_auto,f_auto", // Optimize images
	}

	// Upload the file
	result, err := cs.client.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload image to Cloudinary: %w", err)
	}

	// Return the secure URL
	return result.SecureURL, nil
}

// isValidImageType checks if the file has a valid image extension
func (cs *CloudinaryService) isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

// generateUniqueFilename creates a unique filename to avoid conflicts
func (cs *CloudinaryService) generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	name := strings.TrimSuffix(originalFilename, ext)

	// Clean the filename (remove special characters)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	// Add timestamp for uniqueness
	return fmt.Sprintf("%s_%d%s", name, getCurrentTimestamp(), ext)
}

// getCurrentTimestamp returns current Unix timestamp
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// DeleteImage deletes an image from Cloudinary using its public ID
func (cs *CloudinaryService) DeleteImage(ctx context.Context, imageURL string) error {
	// Extract public ID from Cloudinary URL
	publicID := cs.extractPublicIDFromURL(imageURL)
	if publicID == "" {
		return fmt.Errorf("invalid Cloudinary URL")
	}

	// Delete the image
	_, err := cs.client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "image",
	})

	if err != nil {
		return fmt.Errorf("failed to delete image from Cloudinary: %w", err)
	}

	return nil
}

// extractPublicIDFromURL extracts the public ID from a Cloudinary URL
func (cs *CloudinaryService) extractPublicIDFromURL(url string) string {
	// Example URL: https://res.cloudinary.com/demo/image/upload/v1234567890/amalajeun/spots/1/image.jpg
	parts := strings.Split(url, "/")
	if len(parts) < 7 {
		return ""
	}

	// Find the version part (starts with 'v' followed by numbers)
	versionIndex := -1
	for i, part := range parts {
		if strings.HasPrefix(part, "v") && len(part) > 1 {
			versionIndex = i
			break
		}
	}

	if versionIndex == -1 || versionIndex+1 >= len(parts) {
		return ""
	}

	// Join everything after the version as the public ID (without extension)
	publicIDParts := parts[versionIndex+1:]
	publicID := strings.Join(publicIDParts, "/")

	// Remove file extension
	if lastDot := strings.LastIndex(publicID, "."); lastDot != -1 {
		publicID = publicID[:lastDot]
	}

	return publicID
}
