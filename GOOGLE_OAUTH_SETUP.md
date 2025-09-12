# Google OAuth2 Setup Guide

This guide will help you set up Google OAuth2 authentication for your Amala Jeun backend application.

## Prerequisites

1. A Google Cloud Console account
2. Your Amala Jeun backend running locally or deployed

## Step 1: Create a Google Cloud Project

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Click on the project dropdown and select "New Project"
3. Enter a project name (e.g., "Amala Jeun")
4. Click "Create"

## Step 2: Enable Google+ API

1. In your Google Cloud Console, go to "APIs & Services" > "Library"
2. Search for "Google+ API" and click on it
3. Click "Enable" to enable the Google+ API for your project

## Step 3: Create OAuth2 Credentials

1. Go to "APIs & Services" > "Credentials"
2. Click "Create Credentials" > "OAuth 2.0 Client IDs"
3. If prompted, configure the OAuth consent screen:
   - User Type: External
   - App name: Amala Jeun
   - User support email: Your email
   - Developer contact information: Your email
   - Click "Save and Continue"
4. On the "Scopes" page, you can skip for now
5. On the "Test users" page, add your email as a test user
6. Click "Save and Continue"
7. Now create the OAuth client ID:
   - Application type: Web application
   - Name: Amala Jeun Web Client
   - Authorized JavaScript origins:
     - `http://localhost:3000` (for development)
     - `http://localhost:8080` (for backend)
     - Add your production domains when ready
   - Authorized redirect URIs:
     - `http://localhost:8080/auth/google/callback` (for development)
     - Add your production callback URL when ready
8. Click "Create"
9. Copy the Client ID and Client Secret - you'll need these!

## Step 4: Configure Environment Variables

Update your `.env` file with the credentials from Step 3:

```env
# Google OAuth2 Configuration
GOOGLE_CLIENT_ID=your-actual-google-client-id-here
GOOGLE_CLIENT_SECRET=your-actual-google-client-secret-here
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production-make-it-long-and-random
JWT_EXPIRY_HOURS=24

# Frontend Configuration
FRONTEND_URL=http://localhost:3000
```

## Step 5: Test the OAuth Flow

1. Start your Amala Jeun backend:
   ```bash
   go run main.go
   ```

2. Test the OAuth endpoints:
   - Health check: `GET http://localhost:8080/api/v1/health`
   - Initiate OAuth: `GET http://localhost:8080/api/v1/auth/google`

3. The OAuth flow should work as follows:
   - Frontend calls `/api/v1/auth/google` to get the authorization URL
   - User is redirected to Google for authentication
   - Google redirects back to `/api/v1/auth/google/callback`
   - Backend processes the callback and returns a JWT token
   - Frontend stores the token for authenticated requests

## Step 6: Using Authentication in Frontend

Here's how to use the JWT token in your frontend requests:

```javascript
// Store token after login
const token = response.data.token;
localStorage.setItem('authToken', token);

// Use token in API requests
const headers = {
  'Authorization': `Bearer ${token}`,
  'Content-Type': 'application/json'
};

fetch('/api/v1/spots', {
  method: 'POST',
  headers: headers,
  body: JSON.stringify(spotData)
});
```

## API Endpoints

### Authentication Endpoints

- `GET /api/v1/auth/google` - Initiate Google OAuth login
- `GET /api/v1/auth/google/callback` - Handle OAuth callback (internal)
- `GET /api/v1/auth/profile` - Get current user profile (requires auth)
- `POST /api/v1/auth/logout` - Logout user

### Protected Endpoints (require authentication)

- `POST /api/v1/spots` - Create a new spot
- `POST /api/v1/reviews` - Create a review
- `PATCH /api/v1/spots/:id/verify` - Verify a spot

### Public Endpoints (no authentication required)

- `GET /api/v1/spots` - List all spots
- `GET /api/v1/spots/:id` - Get spot details
- `GET /api/v1/reviews/:spotId` - Get reviews for a spot
- `GET /api/v1/health` - Health check

## Security Notes

1. **Never commit your `.env` file** to version control
2. **Use HTTPS in production** - OAuth2 requires secure connections
3. **Change the JWT secret** in production to a long, random string
4. **Validate redirect URIs** in production to prevent open redirect attacks
5. **Use environment-specific credentials** (dev/staging/prod)

## Troubleshooting

### Common Issues

1. **"redirect_uri_mismatch" error**
   - Make sure the redirect URI in Google Console matches exactly
   - Include the full URL with protocol and port

2. **"invalid_client" error**
   - Check that your Client ID and Secret are correct
   - Ensure there are no extra spaces or characters

3. **CORS errors**
   - Make sure your frontend domain is in the allowed origins
   - Check that the CORS middleware is properly configured

4. **JWT token errors**
   - Verify the JWT_SECRET is set correctly
   - Check token expiration (default 24 hours)

### Testing OAuth Flow

You can test the OAuth flow manually:

1. Visit: `http://localhost:8080/api/v1/auth/google`
2. This should return a JSON with an `auth_url`
3. Visit that URL in your browser
4. Complete the Google authentication
5. You should be redirected back with a JWT token

## Production Deployment

When deploying to production:

1. Update the redirect URI in Google Console to your production domain
2. Add your production domain to authorized origins
3. Use strong, unique secrets for JWT_SECRET
4. Consider implementing token refresh logic
5. Set up proper logging and monitoring
6. Consider using Redis for OAuth state storage in production

## Support

If you encounter issues:

1. Check the application logs for detailed error messages
2. Verify all environment variables are set correctly
3. Ensure your Google Cloud project is properly configured
4. Test with the health endpoint to verify the service is running

The authentication system is now fully integrated with your Amala Jeun backend! 🎉