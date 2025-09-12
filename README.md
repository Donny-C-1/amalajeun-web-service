# Amala Jeun Backend API

A Go-based backend API for Amala Jeun - a Yelp-like platform for discovering Amala spots in Lagos, Nigeria. The API includes AI agent capabilities for adding, verifying, and discovering spots.

## Features

- **Spot Management**: Add, list, and verify Amala spots
- **Review System**: Add and retrieve reviews for spots
- **Multiple Sources**: Support for user-generated, AI agent, and scraper-sourced data
- **Verification System**: Mark spots as verified
- **Pagination**: Efficient data retrieval with pagination support
- **Filtering**: Filter spots by verification status and source

## Tech Stack

- **Language**: Go 1.25.1
- **Framework**: Gin (HTTP web framework)
- **Database**: PostgreSQL
- **ORM**: GORM
- **Architecture**: Clean architecture with separate packages for models, handlers, routes, and database

## Project Structure

```
AmalaJeun/
├── main.go                 # Application entry point
├── go.mod                  # Go module dependencies
├── models/                 # Data models
│   ├── spot.go            # Spot model
│   └── review.go          # Review model
├── database/              # Database configuration
│   └── database.go        # Connection and migration setup
├── handlers/              # HTTP request handlers
│   ├── spot_handlers.go   # Spot-related endpoints
│   └── review_handlers.go # Review-related endpoints
└── routes/                # Route definitions
    └── routes.go          # API route setup
```

## Database Setup

### Prerequisites
- PostgreSQL installed and running
- Database created (default: `amalajeun`)

### Environment Variables
Set the following environment variables (optional, defaults provided):

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=amalajeun
DB_SSLMODE=disable
PORT=8080
GIN_MODE=debug
```

## Installation & Running

1. **Clone and navigate to the project**:
   ```bash
   cd AmalaJeun
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Run the application**:
   ```bash
   go run main.go
   ```

   Or build and run:
   ```bash
   go build -o amalajeun.exe .
   ./amalajeun.exe
   ```

4. **Verify the server is running**:
   ```bash
   curl http://localhost:8080/api/v1/health
   ```

## API Endpoints

### Health Check
- **GET** `/api/v1/health` - Check API status

### Spots
- **POST** `/spots` - Create a new Amala spot
- **GET** `/spots` - List all spots (with pagination and filtering)
- **GET** `/spots/:id` - Get a specific spot with reviews
- **PATCH** `/spots/:id/verify` - Mark a spot as verified

### Reviews
- **POST** `/reviews` - Add a review for a spot
- **GET** `/reviews/:spotId` - Get all reviews for a specific spot

## API Usage Examples

### Create a Spot
```bash
curl -X POST http://localhost:8080/spots \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Mama Cass Amala Spot",
    "address": "123 Lagos Street, Ikeja, Lagos",
    "latitude": 6.5244,
    "longitude": 3.3792,
    "added_by": "user123",
    "source": "user"
  }'
```

### List Spots with Filtering
```bash
# Get all spots
curl http://localhost:8080/spots

# Get verified spots only
curl http://localhost:8080/spots?verified=true

# Get spots with pagination
curl http://localhost:8080/spots?limit=10&offset=0

# Get spots by source
curl http://localhost:8080/spots?source=agent
```

### Get a Specific Spot
```bash
curl http://localhost:8080/spots/1
```

### Verify a Spot
```bash
curl -X PATCH http://localhost:8080/spots/1/verify
```

### Add a Review
```bash
curl -X POST http://localhost:8080/reviews \
  -H "Content-Type: application/json" \
  -d '{
    "spot_id": 1,
    "user_name": "john_doe",
    "rating": 5,
    "comment": "Amazing amala! Best in Lagos!"
  }'
```

### Get Reviews for a Spot
```bash
# Get all reviews for spot ID 1
curl http://localhost:8080/reviews/1

# Get reviews with pagination and sorting
curl http://localhost:8080/reviews/1?limit=5&offset=0&sort=rating&order=desc
```

## Data Models

### Spot Model
```json
{
  "id": 1,
  "name": "Mama Cass Amala Spot",
  "address": "123 Lagos Street, Ikeja, Lagos",
  "latitude": 6.5244,
  "longitude": 3.3792,
  "added_by": "user123",
  "verified": false,
  "source": "user",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Review Model
```json
{
  "id": 1,
  "spot_id": 1,
  "user_name": "john_doe",
  "rating": 5,
  "comment": "Amazing amala! Best in Lagos!",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

## Error Handling

The API returns appropriate HTTP status codes:
- **200**: Success
- **201**: Created
- **400**: Bad Request (invalid data)
- **404**: Not Found
- **500**: Internal Server Error

Error responses include details:
```json
{
  "error": "Invalid request data",
  "details": "Field validation error details"
}
```

## Development

### Adding New Features
1. Add models in `models/` package
2. Create handlers in `handlers/` package
3. Register routes in `routes/routes.go`
4. Update database migrations in `database/database.go`

### Database Migrations
The application automatically runs migrations on startup. Models are defined using GORM tags for automatic schema generation.

## Production Deployment

1. Set `GIN_MODE=release`
2. Configure proper database credentials
3. Set up proper CORS policies
4. Use a process manager like systemd or Docker
5. Set up proper logging and monitoring

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.