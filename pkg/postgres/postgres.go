package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database represents a database connection using GORM
type Database struct {
	*gorm.DB
}

// New creates a new database connection using GORM
func New(cfg Config) (*Database, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if cfg.Host == "localhost" {
		gormLogger = logger.Default.LogMode(logger.Silent) // Reduce noise in development
	}

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		if !strings.Contains(err.Error(), "does not exist") {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
		defaultDsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode)
		defaultDb, err := gorm.Open(postgres.Open(defaultDsn), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to open default database: %w", err)
		}

		defaultDb.Exec(fmt.Sprintf("CREATE DATABASE %s;", cfg.Name))
		db, err = gorm.Open(postgres.Open(fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
	}
	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Database connection established with GORM",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Name)

	return &Database{db}, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	slog.Info("Closing database connection")
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// HealthCheck checks if the database is healthy
func (db *Database) HealthCheck() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Ping()
}

// BeginTx starts a new transaction
func (db *Database) BeginTx() (*gorm.DB, error) {
	return db.DB.Begin(), nil
}

// WithTimeout creates a context with timeout for database operations
func (db *Database) WithTimeout(timeout time.Duration) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// Note: GORM doesn't have a direct way to pass context to the DB instance
	// You should use this context when calling GORM methods
	_ = cancel // Prevent unused variable warning
	return db.DB.WithContext(ctx)
}

// GetDB returns the underlying GORM DB instance
func (db *Database) GetDB() *gorm.DB {
	return db.DB
}

// AutoMigrate runs auto migration for the given models
func (db *Database) AutoMigrate(dst ...interface{}) error {
	return db.DB.AutoMigrate(dst...)
}

// Transaction executes a function within a database transaction
func (db *Database) Transaction(fc func(tx *gorm.DB) error) error {
	return db.DB.Transaction(fc)
}
