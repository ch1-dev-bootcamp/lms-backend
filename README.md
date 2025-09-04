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

1. Make sure you have Go installed (version 1.19 or later)
2. Clone the repository
3. Run the application:
   ```bash
   go run cmd/main.go
   ```

The server will start on port 8080 by default.

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