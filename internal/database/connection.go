package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/your-org/lms-backend/pkg/config"
	"github.com/your-org/lms-backend/pkg/logger"
)

// DB holds the database connection
var DB *sql.DB

// ConnectionConfig holds database connection configuration
type ConnectionConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewConnectionConfig creates a new connection config from app config
func NewConnectionConfig(cfg *config.Config) *ConnectionConfig {
	return &ConnectionConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}
}

// Connect establishes a connection to the database with retry logic
func Connect(cfg *ConnectionConfig) error {
	// Build connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(5)                  // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection
	db.SetConnMaxIdleTime(1 * time.Minute) // Maximum idle time of a connection

	// Retry connection with exponential backoff
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(); err != nil {
			if i == maxRetries-1 {
				return fmt.Errorf("failed to ping database after %d attempts: %w", maxRetries, err)
			}
			logger.Warn("Database connection attempt failed, retrying...", logger.WithFields(map[string]interface{}{
				"attempt": i + 1,
				"max_retries": maxRetries,
				"retry_in_seconds": (i + 1) * 2,
				"error": err.Error(),
			}).Data)
			time.Sleep(time.Duration((i+1)*2) * time.Second)
			continue
		}
		break
	}

	// Set global DB variable
	DB = db

	logger.Info("Successfully connected to database", logger.WithFields(map[string]interface{}{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"user":     cfg.User,
		"database": cfg.DBName,
	}).Data)
	
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return DB
}

// IsConnected checks if the database is connected
func IsConnected() bool {
	if DB == nil {
		return false
	}
	
	if err := DB.Ping(); err != nil {
		return false
	}
	
	return true
}

// HealthCheck performs a database health check
func HealthCheck() error {
	if !IsConnected() {
		return fmt.Errorf("database is not connected")
	}
	
	// Perform a simple query to check database health
	var result int
	err := DB.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	
	return nil
}