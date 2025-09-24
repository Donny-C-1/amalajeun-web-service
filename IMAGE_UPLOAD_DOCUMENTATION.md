# Image Upload Functionality Documentation

## Overview

This document describes the image upload functionality added to the AmalaJeun project. Users can now upload images for spots, which are stored in Cloudinary and associated with the correct spot records in the database.

## Changes Made

### 1. Dependencies Added

- **Cloudinary Go SDK**: `github.com/cloudinary/cloudinary-go/v2` - For handling image uploads to Cloudinary
- **Additional dependencies**: `github.com/creasty/defaults`, `github.com/gorilla/schema` (automatically added)

### 2. Database Schema Changes

#### Updated Spot Model (`models/spot.go`)

- **Added `StringArray` type**: Custom type for handling JSON arrays in PostgreSQL
- **Added `Images` field**: `StringArray` field to store multiple Cloudinary URLs per spot
  ```go
  Images StringArray `json:"images" gorm:"type:jsonb;default:'[]'"`
  ```

The `StringArray` type includes:
- `Scan()` method for database deserialization
- `Value()` method for database serialization
- Proper JSON marshaling/unmarshaling

### 3. New Service Package

#### Cloudinary Service (`services/cloudinary/cloudinary_service.go`)

**Features:**
- **Environment-based configuration**: Uses `CLOUDINARY_URL` environment variable
- **File validation**: 
  - Supported formats: JPEG, PNG, GIF, WebP
  - Maximum file size: 5MB
- **Organized storage**: Images stored in `amalajeun/spots/{spot_id}/` folders
- **Image optimization**: Automatic resizing and format optimization
- **Unique naming**: Timestamp-based filename generation to prevent conflicts

**Key Methods:**
- `NewCloudinaryService()`: Initialize service with environment credentials
- `UploadImage()`: Upload image with validation and return secure URL
- `DeleteImage()`: Delete image from Cloudinary (for future use)

### 4. New API Endpoint

#### Upload Spot Image Handler (`handlers/spot_handlers.go`)

**Endpoint**: `POST /api/v1/spots/:id/images`

**Features:**
- **Authentication required**: Uses existing JWT middleware
- **Multipart form data**: Accepts image file with key `image`
- **Spot validation**: Verifies spot exists before upload
- **Database update**: Adds image URL to spot's images array
- **Comprehensive error handling**: Detailed error messages for different failure scenarios

**Request Format:**
```
POST /api/v1/spots/{spot_id}/images
Content-Type: multipart/form-data
Authorization: Bearer {jwt_token}

Form Data:
- image: [image file]
```

**Response Format:**
```json
{
  "message": "Image uploaded successfully",
  "image_url": "https://res.cloudinary.com/...",
  "spot_id": 123,
  "total_images": 2,
  "uploaded_by": {
    "id": "user-uuid",
    "name": "User Name",
    "email": "user@example.com"
  }
}
```

### 5. Routing Updates

#### Routes Configuration (`routes/routes.go`)

- **Added protected route**: `POST /api/v1/spots/:id/images` → `handlers.UploadSpotImage`
- **Maintains consistency**: Uses same authentication middleware as other protected endpoints

#### API Documentation (`main.go`)

- **Updated startup logs**: Added new endpoint to API documentation display

## Environment Configuration

### Required Environment Variables

```bash
# Cloudinary Configuration
CLOUDINARY_URL=cloudinary://api_key:api_secret@cloud_name

# Alternative format (if CLOUDINARY_URL not available)
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
```

## Usage Examples

### 1. Upload Image using cURL

```bash
curl -X POST \
  http://localhost:8080/api/v1/spots/123/images \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "image=@/path/to/your/image.jpg"
```

### 2. Upload Image using JavaScript (Frontend)

```javascript
const formData = new FormData();
formData.append('image', imageFile);

fetch('/api/v1/spots/123/images', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${jwtToken}`
  },
  body: formData
})
.then(response => response.json())
.then(data => console.log('Upload successful:', data))
.catch(error => console.error('Upload failed:', error));
```

### 3. Retrieve Spot with Images

```bash
curl -X GET http://localhost:8080/api/v1/spots/123
```

Response will include the images array:
```json
{
  "data": {
    "id": 123,
    "name": "Great Amala Spot",
    "address": "Lagos, Nigeria",
    "images": [
      "https://res.cloudinary.com/demo/image/upload/v1234567890/amalajeun/spots/123/image1.jpg",
      "https://res.cloudinary.com/demo/image/upload/v1234567890/amalajeun/spots/123/image2.jpg"
    ],
    // ... other spot fields
  }
}
```

## Error Handling

### Common Error Responses

1. **Authentication Required** (401):
   ```json
   {"error": "Authentication required to upload images"}
   ```

2. **Invalid Spot ID** (400):
   ```json
   {"error": "Invalid spot ID"}
   ```

3. **Spot Not Found** (404):
   ```json
   {"error": "Spot not found"}
   ```

4. **No Image File** (400):
   ```json
   {
     "error": "No image file provided",
     "details": "Please provide an image file with the key 'image'"
   }
   ```

5. **Invalid File Type** (400):
   ```json
   {
     "error": "Failed to upload image",
     "details": "invalid file type. Only JPEG, PNG, GIF, and WebP images are allowed"
   }
   ```

6. **File Too Large** (400):
   ```json
   {
     "error": "Failed to upload image",
     "details": "file size exceeds 5MB limit"
   }
   ```

7. **Cloudinary Configuration Error** (500):
   ```json
   {
     "error": "Failed to initialize image upload service",
     "details": "CLOUDINARY_URL environment variable is required"
   }
   ```

## Database Migration

The new `Images` field will be automatically added to existing spots when the application starts, thanks to GORM's auto-migration feature. Existing spots will have an empty images array (`[]`) by default.

## Security Considerations

1. **Authentication Required**: Only authenticated users can upload images
2. **File Type Validation**: Only image files are accepted
3. **File Size Limits**: 5MB maximum to prevent abuse
4. **Organized Storage**: Images are organized by spot ID in Cloudinary
5. **Secure URLs**: All image URLs are HTTPS by default

## Performance Considerations

1. **Direct Upload**: Images go directly to Cloudinary (no local storage)
2. **Image Optimization**: Automatic compression and format optimization
3. **JSON Storage**: Efficient storage of image URLs in PostgreSQL JSONB
4. **Indexed Queries**: Spot lookups remain fast with existing indexes

## Future Enhancements

1. **Image Deletion**: Add endpoint to remove images from spots
2. **Image Ordering**: Allow users to reorder images
3. **Image Metadata**: Store additional metadata (captions, alt text)
4. **Bulk Upload**: Support multiple image uploads in single request
5. **Image Thumbnails**: Generate different sizes for various use cases

## Testing

The implementation has been tested for:
- ✅ Successful compilation
- ✅ Route registration
- ✅ Handler function creation
- ✅ Database schema updates
- ✅ Service integration

For full testing, ensure:
1. Set up Cloudinary account and configure environment variables
2. Test with various image formats and sizes
3. Verify database updates after successful uploads
4. Test error scenarios (invalid files, missing authentication, etc.)