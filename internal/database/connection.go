package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/your-org/lms-backend/internal/repository"
	"github.com/your-org/lms-backend/internal/repository/postgres"
	"github.com/your-org/lms-backend/pkg/config"
	"github.com/your-org/lms-backend/pkg/logger"
)

// DB holds the database connection
var DB *sql.DB

// RepoManager holds the repository manager
var RepoManager repository.RepositoryManager

// ConnectionConfig holds database connection configuration
type ConnectionConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// NewConnectionConfig creates a new connection config from app config
func NewConnectionConfig(cfg *config.Config) *ConnectionConfig {
	return &ConnectionConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}
}

// Connect establishes a connection to the database with retry logic
func Connect(cfg *ConnectionConfig) error {
	// Convert port string to int
	port := 5432 // default PostgreSQL port
	if cfg.Port != "" {
		if p, err := fmt.Sscanf(cfg.Port, "%d", &port); err != nil || p != 1 {
			logger.Warn("Invalid port, using default 5432", logger.WithFields(map[string]interface{}{
				"provided_port": cfg.Port,
				"default_port": 5432,
			}).Data)
		}
	}

	// Create PostgreSQL database config
	dbConfig := postgres.DatabaseConfig{
		Host:            cfg.Host,
		Port:            port,
		User:            cfg.User,
		Password:        cfg.Password,
		DBName:          cfg.DBName,
		SSLMode:         cfg.SSLMode,
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
	}

	// Create database connection with pooling
	db, err := postgres.NewDatabaseConnection(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}

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

	// Create repository manager
	RepoManager = postgres.NewRepositoryManager(db)

	logger.Info("Successfully connected to database with repository manager", logger.WithFields(map[string]interface{}{
		"host":            cfg.Host,
		"port":            port,
		"user":            cfg.User,
		"database":        cfg.DBName,
		"max_open_conns":  cfg.MaxOpenConns,
		"max_idle_conns":  cfg.MaxIdleConns,
		"conn_max_lifetime": cfg.ConnMaxLifetime,
		"conn_max_idle_time": cfg.ConnMaxIdleTime,
	}).Data)
	
	return nil
}

// Close closes the database connection
func Close() error {
	if RepoManager != nil {
		if err := RepoManager.Close(); err != nil {
			return fmt.Errorf("failed to close repository manager: %w", err)
		}
	}
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return DB
}

// GetRepoManager returns the repository manager
func GetRepoManager() repository.RepositoryManager {
	return RepoManager
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