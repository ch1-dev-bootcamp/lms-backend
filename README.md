# LMS Backend

A Learning Management System backend built with Go.

## Project Structure

```
lms-backend/
├── cmd/                    # Application entry points
│   └── main.go            # Main application
├── internal/              # Private application code
│   ├── handlers/          # HTTP handlers
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Data models
│   ├── services/          # Business logic
│   └── database/          # Database layer
├── pkg/                   # Public library code
│   └── config/            # Configuration management
├── api/                   # API definitions
├── configs/               # Configuration files
├── migrations/            # Database migrations
├── go.mod                 # Go module file
└── README.md              # This file
```

## Getting Started

### Option 1: Using Docker (Recommended)

1. Make sure you have Docker and Docker Compose installed
2. Clone the repository
3. Start all services:
   ```bash
   docker-compose up -d
   ```

This will start:
- PostgreSQL database on port 5432
- Redis cache on port 6379
- LMS Backend API on port 8080

### Option 2: Local Development

1. Make sure you have Go installed (version 1.19 or later)
2. Install PostgreSQL and Redis locally
3. Clone the repository
4. Set up environment variables (copy from .env.example)
5. Run the application:
   ```bash
   go run cmd/main.go
   ```

The server will start on port 8080 by default.

### Docker Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Rebuild and start
docker-compose up --build -d

# Stop and remove volumes (WARNING: deletes data)
docker-compose down -v
```

## Environment Variables

- `SERVER_PORT`: Server port (default: 8080)
- `SERVER_HOST`: Server host (default: localhost)
- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database user (default: postgres)
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name (default: lms)
- `DB_SSLMODE`: Database SSL mode (default: disable)

## API Endpoints

- `GET /health` - Health check
- `GET /api/v1/` - API root