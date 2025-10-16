package jwt

import (
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	config := Config{
		AccessTokenSecret:  "test-access-secret",
		RefreshTokenSecret: "test-refresh-secret",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}

	manager := NewManager(config)

	if manager.accessTokenSecret != config.AccessTokenSecret {
		t.Errorf("Expected access token secret %s, got %s", config.AccessTokenSecret, manager.accessTokenSecret)
	}

	if manager.refreshTokenSecret != config.RefreshTokenSecret {
		t.Errorf("Expected refresh token secret %s, got %s", config.RefreshTokenSecret, manager.refreshTokenSecret)
	}

	if manager.accessTokenTTL != config.AccessTokenTTL {
		t.Errorf("Expected access token TTL %v, got %v", config.AccessTokenTTL, manager.accessTokenTTL)
	}
}

func TestGenerateTokenPair(t *testing.T) {
	config := Config{
		AccessTokenSecret:  "test-access-secret",
		RefreshTokenSecret: "test-refresh-secret",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}

	manager := NewManager(config)

	userID := "user123"
	email := "test@example.com"
	username := "testuser"
	role := "admin"

	tokenPair, err := manager.GenerateTokenPair(userID, email, username, role)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	if tokenPair.AccessToken == "" {
		t.Error("Access token is empty")
	}

	if tokenPair.RefreshToken == "" {
		t.Error("Refresh token is empty")
	}

	if tokenPair.TokenType != "Bearer" {
		t.Errorf("Expected token type Bearer, got %s", tokenPair.TokenType)
	}

	if tokenPair.ExpiresAt.IsZero() {
		t.Error("ExpiresAt is zero")
	}
}

func TestValidateAccessToken(t *testing.T) {
	config := Config{
		AccessTokenSecret:  "test-access-secret",
		RefreshTokenSecret: "test-refresh-secret",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}

	manager := NewManager(config)

	userID := "user123"
	email := "test@example.com"
	username := "testuser"
	role := "admin"

	tokenPair, err := manager.GenerateTokenPair(userID, email, username, role)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	// Validate access token
	claims, err := manager.ValidateAccessToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}

	if claims.Username != username {
		t.Errorf("Expected username %s, got %s", username, claims.Username)
	}

	if claims.Role != role {
		t.Errorf("Expected role %s, got %s", role, claims.Role)
	}
}

func TestValidateRefreshToken(t *testing.T) {
	config := Config{
		AccessTokenSecret:  "test-access-secret",
		RefreshTokenSecret: "test-refresh-secret",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}

	manager := NewManager(config)

	userID := "user123"
	email := "test@example.com"
	username := "testuser"
	role := "admin"

	tokenPair, err := manager.GenerateTokenPair(userID, email, username, role)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	// Validate refresh token
	claims, err := manager.ValidateRefreshToken(tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}
}

func TestInvalidToken(t *testing.T) {
	config := Config{
		AccessTokenSecret:  "test-access-secret",
		RefreshTokenSecret: "test-refresh-secret",
	}

	manager := NewManager(config)

	// Test with invalid token
	_, err := manager.ValidateAccessToken("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}

	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestExpiredToken(t *testing.T) {
	config := Config{
		AccessTokenSecret:  "test-access-secret",
		RefreshTokenSecret: "test-refresh-secret",
		AccessTokenTTL:     1 * time.Millisecond, // Very short TTL
		RefreshTokenTTL:    1 * time.Millisecond,
		Issuer:             "test-issuer",
	}

	manager := NewManager(config)

	userID := "user123"
	email := "test@example.com"
	username := "testuser"
	role := "admin"

	tokenPair, err := manager.GenerateTokenPair(userID, email, username, role)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Try to validate expired token
	_, err = manager.ValidateAccessToken(tokenPair.AccessToken)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}

	if err != ErrExpiredToken {
		t.Errorf("Expected ErrExpiredToken, got %v", err)
	}
}

func TestRefreshTokens(t *testing.T) {
	config := Config{
		AccessTokenSecret:  "test-access-secret",
		RefreshTokenSecret: "test-refresh-secret",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}

	manager := NewManager(config)

	userID := "user123"
	email := "test@example.com"
	username := "testuser"
	role := "admin"

	// Generate initial token pair
	originalTokenPair, err := manager.GenerateTokenPair(userID, email, username, role)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	// Refresh tokens
	newTokenPair, err := manager.RefreshTokens(originalTokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("Failed to refresh tokens: %v", err)
	}

	if newTokenPair.AccessToken == "" {
		t.Error("New access token is empty")
	}

	if newTokenPair.RefreshToken == "" {
		t.Error("New refresh token is empty")
	}

	// Validate new access token
	claims, err := manager.ValidateAccessToken(newTokenPair.AccessToken)
	if err != nil {
		t.Fatalf("Failed to validate new access token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}
}

func TestClaimsIsExpired(t *testing.T) {
	// Test with a token that has a reasonable lifetime
	config := Config{
		AccessTokenSecret:  "test-access-secret",
		RefreshTokenSecret: "test-refresh-secret",
		AccessTokenTTL:     2 * time.Second,
		Issuer:             "test-issuer",
	}

	manager := NewManager(config)

	tokenPair, err := manager.GenerateTokenPair("user123", "test@example.com", "testuser", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	claims, err := manager.ValidateAccessToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	// Check if not expired immediately
	if claims.IsExpired() {
		t.Errorf("Token should not be expired immediately. ExpiresAt: %v, Now: %v", claims.GetTokenExpiry(), time.Now())
	}

	// Verify time until expiry is positive
	timeUntil := claims.TimeUntilExpiry()
	if timeUntil <= 0 {
		t.Errorf("Time until expiry should be positive, got %v", timeUntil)
	}

	// Wait for token to expire
	time.Sleep(2100 * time.Millisecond)

	// Check if expired
	if !claims.IsExpired() {
		t.Errorf("Token should be expired. ExpiresAt: %v, Now: %v", claims.GetTokenExpiry(), time.Now())
	}
}

func TestTimeUntilExpiry(t *testing.T) {
	config := Config{
		AccessTokenSecret: "test-access-secret",
		AccessTokenTTL:    15 * time.Minute,
		Issuer:            "test-issuer",
	}

	manager := NewManager(config)

	tokenPair, err := manager.GenerateTokenPair("user123", "test@example.com", "testuser", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	claims, err := manager.ExtractClaims(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("Failed to extract claims: %v", err)
	}

	timeUntilExpiry := claims.TimeUntilExpiry()
	if timeUntilExpiry <= 0 {
		t.Error("Time until expiry should be positive")
	}

	if timeUntilExpiry > 15*time.Minute {
		t.Errorf("Time until expiry should be less than or equal to 15 minutes, got %v", timeUntilExpiry)
	}
}
