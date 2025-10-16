package postgres

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// LoadConfigFromEnv loads PostgreSQL configuration from environment variables
// Returns an error if required environment variables are missing
func LoadConfigFromEnv() (Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we don't return error if it fails to load
	}

	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		return Config{}, fmt.Errorf("POSTGRES_HOST environment variable is required")
	}

	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		return Config{}, fmt.Errorf("POSTGRES_PORT environment variable is required")
	}

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		return Config{}, fmt.Errorf("POSTGRES_USER environment variable is required")
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		return Config{}, fmt.Errorf("POSTGRES_PASSWORD environment variable is required")
	}

	name := os.Getenv("POSTGRES_NAME")
	if name == "" {
		return Config{}, fmt.Errorf("POSTGRES_NAME environment variable is required")
	}

	config := Config{
		Host:            host,
		Port:            port,
		User:            user,
		Password:        password,
		Name:            name,
		SSLMode:         "disable",       // Default: disable
		MaxOpenConns:    25,              // Default: 25 connections
		MaxIdleConns:    5,               // Default: 5 idle connections
		ConnMaxLifetime: 5 * time.Minute, // Default: 5 minutes
	}

	// Optional: Load SSL mode
	if sslMode := os.Getenv("POSTGRES_SSL_MODE"); sslMode != "" {
		config.SSLMode = sslMode
	}

	// Optional: Load max open connections
	if maxOpenConns := os.Getenv("POSTGRES_MAX_OPEN_CONNS"); maxOpenConns != "" {
		value, err := strconv.Atoi(maxOpenConns)
		if err != nil {
			return Config{}, fmt.Errorf("invalid POSTGRES_MAX_OPEN_CONNS format: %w", err)
		}
		config.MaxOpenConns = value
	}

	// Optional: Load max idle connections
	if maxIdleConns := os.Getenv("POSTGRES_MAX_IDLE_CONNS"); maxIdleConns != "" {
		value, err := strconv.Atoi(maxIdleConns)
		if err != nil {
			return Config{}, fmt.Errorf("invalid POSTGRES_MAX_IDLE_CONNS format: %w", err)
		}
		config.MaxIdleConns = value
	}

	// Optional: Load connection max lifetime
	if connMaxLifetime := os.Getenv("POSTGRES_CONN_MAX_LIFETIME"); connMaxLifetime != "" {
		duration, err := time.ParseDuration(connMaxLifetime)
		if err != nil {
			return Config{}, fmt.Errorf("invalid POSTGRES_CONN_MAX_LIFETIME format: %w", err)
		}
		config.ConnMaxLifetime = duration
	}

	return config, nil
}

// MustLoadConfigFromEnv loads PostgreSQL configuration from environment variables
// Panics if required environment variables are missing
func MustLoadConfigFromEnv() Config {
	config, err := LoadConfigFromEnv()
	if err != nil {
		panic(fmt.Sprintf("failed to load PostgreSQL config: %v", err))
	}
	return config
}

// ValidateConfig checks if the configuration is valid
func ValidateConfig(config Config) error {
	if config.Host == "" {
		return fmt.Errorf("host is required")
	}

	if config.Port == "" {
		return fmt.Errorf("port is required")
	}

	if config.User == "" {
		return fmt.Errorf("user is required")
	}

	if config.Password == "" {
		return fmt.Errorf("password is required")
	}

	if config.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if config.SSLMode == "" {
		return fmt.Errorf("SSL mode is required")
	}

	validSSLModes := map[string]bool{
		"disable":     true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if !validSSLModes[config.SSLMode] {
		return fmt.Errorf("invalid SSL mode: %s (valid options: disable, require, verify-ca, verify-full)", config.SSLMode)
	}

	if config.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be positive")
	}

	if config.MaxIdleConns <= 0 {
		return fmt.Errorf("max idle connections must be positive")
	}

	if config.MaxIdleConns > config.MaxOpenConns {
		return fmt.Errorf("max idle connections cannot be greater than max open connections")
	}

	if config.ConnMaxLifetime <= 0 {
		return fmt.Errorf("connection max lifetime must be positive")
	}

	return nil
}
