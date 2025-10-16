package jwt

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// LoadConfigFromEnv loads JWT configuration from environment variables
// Returns an error if required environment variables are missing
func LoadConfigFromEnv() (Config, error) {

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we don't return error if it fails to load
	}

	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if accessSecret == "" {
		return Config{}, fmt.Errorf("JWT_ACCESS_SECRET environment variable is required")
	}

	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		return Config{}, fmt.Errorf("JWT_REFRESH_SECRET environment variable is required")
	}

	config := Config{
		AccessTokenSecret:  accessSecret,
		RefreshTokenSecret: refreshSecret,
		AccessTokenTTL:     15 * time.Minute,   // Default: 15 minutes
		RefreshTokenTTL:    7 * 24 * time.Hour, // Default: 7 days
		Issuer:             "offera",           // Default issuer
	}

	// Optional: Load custom TTL values
	if accessTTL := os.Getenv("JWT_ACCESS_TTL"); accessTTL != "" {
		duration, err := time.ParseDuration(accessTTL)
		if err != nil {
			return Config{}, fmt.Errorf("invalid JWT_ACCESS_TTL format: %w", err)
		}
		config.AccessTokenTTL = duration
	}

	if refreshTTL := os.Getenv("JWT_REFRESH_TTL"); refreshTTL != "" {
		duration, err := time.ParseDuration(refreshTTL)
		if err != nil {
			return Config{}, fmt.Errorf("invalid JWT_REFRESH_TTL format: %w", err)
		}
		config.RefreshTokenTTL = duration
	}

	// Optional: Load custom issuer
	if issuer := os.Getenv("JWT_ISSUER"); issuer != "" {
		config.Issuer = issuer
	}

	return config, nil
}

// MustLoadConfigFromEnv loads JWT configuration from environment variables
// Panics if required environment variables are missing
func MustLoadConfigFromEnv() Config {
	config, err := LoadConfigFromEnv()
	if err != nil {
		panic(fmt.Sprintf("failed to load JWT config: %v", err))
	}
	return config
}

// ValidateConfig checks if the configuration is valid
func ValidateConfig(config Config) error {
	if config.AccessTokenSecret == "" {
		return fmt.Errorf("access token secret is required")
	}

	if config.RefreshTokenSecret == "" {
		return fmt.Errorf("refresh token secret is required")
	}

	if config.AccessTokenSecret == config.RefreshTokenSecret {
		return fmt.Errorf("access and refresh token secrets should be different for security")
	}

	if config.AccessTokenTTL <= 0 {
		return fmt.Errorf("access token TTL must be positive")
	}

	if config.RefreshTokenTTL <= 0 {
		return fmt.Errorf("refresh token TTL must be positive")
	}

	if config.AccessTokenTTL >= config.RefreshTokenTTL {
		return fmt.Errorf("access token TTL should be shorter than refresh token TTL")
	}

	if len(config.AccessTokenSecret) < 32 {
		return fmt.Errorf("access token secret should be at least 32 characters for security")
	}

	if len(config.RefreshTokenSecret) < 32 {
		return fmt.Errorf("refresh token secret should be at least 32 characters for security")
	}

	return nil
}
