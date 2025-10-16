package redis

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
}

// LoadConfigFromEnv loads Redis configuration from environment variables
// Returns an error if required environment variables are missing
func LoadConfigFromEnv() (Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we don't return error if it fails to load
	}

	host := os.Getenv("REDIS_HOST")
	if host == "" {
		return Config{}, fmt.Errorf("REDIS_HOST environment variable is required")
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		return Config{}, fmt.Errorf("REDIS_PORT environment variable is required")
	}

	config := Config{
		Host:         host,
		Port:         port,
		Password:     "", // Default: no password
		DB:           0,  // Default: DB 0
		PoolSize:     10, // Default: 10 connections
		MinIdleConns: 2,  // Default: 2 idle connections
	}

	// Optional: Load password
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		config.Password = password
	}

	// Optional: Load DB number
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		db, err := strconv.Atoi(dbStr)
		if err != nil {
			return Config{}, fmt.Errorf("invalid REDIS_DB format: %w", err)
		}
		config.DB = db
	}

	// Optional: Load pool size
	if poolSizeStr := os.Getenv("REDIS_POOL_SIZE"); poolSizeStr != "" {
		poolSize, err := strconv.Atoi(poolSizeStr)
		if err != nil {
			return Config{}, fmt.Errorf("invalid REDIS_POOL_SIZE format: %w", err)
		}
		config.PoolSize = poolSize
	}

	// Optional: Load min idle connections
	if minIdleConnsStr := os.Getenv("REDIS_MIN_IDLE_CONNS"); minIdleConnsStr != "" {
		minIdleConns, err := strconv.Atoi(minIdleConnsStr)
		if err != nil {
			return Config{}, fmt.Errorf("invalid REDIS_MIN_IDLE_CONNS format: %w", err)
		}
		config.MinIdleConns = minIdleConns
	}

	return config, nil
}

// MustLoadConfigFromEnv loads Redis configuration from environment variables
// Panics if required environment variables are missing
func MustLoadConfigFromEnv() Config {
	config, err := LoadConfigFromEnv()
	if err != nil {
		panic(fmt.Sprintf("failed to load Redis config: %v", err))
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

	if config.DB < 0 || config.DB > 15 {
		return fmt.Errorf("DB must be between 0 and 15")
	}

	if config.PoolSize <= 0 {
		return fmt.Errorf("pool size must be positive")
	}

	if config.MinIdleConns < 0 {
		return fmt.Errorf("min idle connections cannot be negative")
	}

	if config.MinIdleConns > config.PoolSize {
		return fmt.Errorf("min idle connections cannot be greater than pool size")
	}

	return nil
}
