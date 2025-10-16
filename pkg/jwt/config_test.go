package jwt

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfigFromEnv(t *testing.T) {
	// Set up environment variables
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-with-sufficient-length-for-security")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-with-sufficient-length-for-security")
	os.Setenv("JWT_ACCESS_TTL", "30m")
	os.Setenv("JWT_REFRESH_TTL", "168h")
	os.Setenv("JWT_ISSUER", "test-issuer")

	defer func() {
		os.Unsetenv("JWT_ACCESS_SECRET")
		os.Unsetenv("JWT_REFRESH_SECRET")
		os.Unsetenv("JWT_ACCESS_TTL")
		os.Unsetenv("JWT_REFRESH_TTL")
		os.Unsetenv("JWT_ISSUER")
	}()

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.AccessTokenSecret != "test-access-secret-with-sufficient-length-for-security" {
		t.Errorf("Expected access secret to be set")
	}

	if config.RefreshTokenSecret != "test-refresh-secret-with-sufficient-length-for-security" {
		t.Errorf("Expected refresh secret to be set")
	}

	if config.AccessTokenTTL != 30*time.Minute {
		t.Errorf("Expected access TTL to be 30m, got %v", config.AccessTokenTTL)
	}

	if config.RefreshTokenTTL != 168*time.Hour {
		t.Errorf("Expected refresh TTL to be 168h, got %v", config.RefreshTokenTTL)
	}

	if config.Issuer != "test-issuer" {
		t.Errorf("Expected issuer to be 'test-issuer', got %s", config.Issuer)
	}
}

func TestLoadConfigFromEnvMissingAccessSecret(t *testing.T) {
	os.Unsetenv("JWT_ACCESS_SECRET")
	os.Unsetenv("JWT_REFRESH_SECRET")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Error("Expected error when JWT_ACCESS_SECRET is missing")
	}
}

func TestLoadConfigFromEnvMissingRefreshSecret(t *testing.T) {
	os.Setenv("JWT_ACCESS_SECRET", "test-secret")
	os.Unsetenv("JWT_REFRESH_SECRET")
	defer os.Unsetenv("JWT_ACCESS_SECRET")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Error("Expected error when JWT_REFRESH_SECRET is missing")
	}
}

func TestLoadConfigFromEnvInvalidTTL(t *testing.T) {
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret")
	os.Setenv("JWT_ACCESS_TTL", "invalid")
	defer func() {
		os.Unsetenv("JWT_ACCESS_SECRET")
		os.Unsetenv("JWT_REFRESH_SECRET")
		os.Unsetenv("JWT_ACCESS_TTL")
	}()

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Error("Expected error for invalid JWT_ACCESS_TTL format")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				AccessTokenSecret:  "this-is-a-very-long-secret-key-for-access-tokens",
				RefreshTokenSecret: "this-is-a-very-long-secret-key-for-refresh-tokens",
				AccessTokenTTL:     15 * time.Minute,
				RefreshTokenTTL:    7 * 24 * time.Hour,
				Issuer:             "test",
			},
			wantErr: false,
		},
		{
			name: "missing access secret",
			config: Config{
				RefreshTokenSecret: "test-refresh-secret-with-sufficient-length",
				AccessTokenTTL:     15 * time.Minute,
				RefreshTokenTTL:    7 * 24 * time.Hour,
			},
			wantErr: true,
		},
		{
			name: "missing refresh secret",
			config: Config{
				AccessTokenSecret: "test-access-secret-with-sufficient-length",
				AccessTokenTTL:    15 * time.Minute,
				RefreshTokenTTL:   7 * 24 * time.Hour,
			},
			wantErr: true,
		},
		{
			name: "same secrets",
			config: Config{
				AccessTokenSecret:  "same-secret-for-both-which-is-not-recommended",
				RefreshTokenSecret: "same-secret-for-both-which-is-not-recommended",
				AccessTokenTTL:     15 * time.Minute,
				RefreshTokenTTL:    7 * 24 * time.Hour,
			},
			wantErr: true,
		},
		{
			name: "negative access TTL",
			config: Config{
				AccessTokenSecret:  "test-access-secret-with-sufficient-length",
				RefreshTokenSecret: "test-refresh-secret-with-sufficient-length",
				AccessTokenTTL:     -1 * time.Minute,
				RefreshTokenTTL:    7 * 24 * time.Hour,
			},
			wantErr: true,
		},
		{
			name: "access TTL longer than refresh TTL",
			config: Config{
				AccessTokenSecret:  "test-access-secret-with-sufficient-length",
				RefreshTokenSecret: "test-refresh-secret-with-sufficient-length",
				AccessTokenTTL:     24 * time.Hour,
				RefreshTokenTTL:    1 * time.Hour,
			},
			wantErr: true,
		},
		{
			name: "short access secret",
			config: Config{
				AccessTokenSecret:  "short",
				RefreshTokenSecret: "test-refresh-secret-with-sufficient-length",
				AccessTokenTTL:     15 * time.Minute,
				RefreshTokenTTL:    7 * 24 * time.Hour,
			},
			wantErr: true,
		},
		{
			name: "short refresh secret",
			config: Config{
				AccessTokenSecret:  "test-access-secret-with-sufficient-length",
				RefreshTokenSecret: "short",
				AccessTokenTTL:     15 * time.Minute,
				RefreshTokenTTL:    7 * 24 * time.Hour,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
