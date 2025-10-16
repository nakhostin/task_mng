package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidClaims    = errors.New("invalid token claims")
	ErrTokenNotYetValid = errors.New("token not yet valid")
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// Manager handles JWT token operations
type Manager struct {
	accessTokenSecret  string
	refreshTokenSecret string
	accessTokenTTL     time.Duration
	refreshTokenTTL    time.Duration
	issuer             string
}

// Config holds JWT manager configuration
type Config struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	Issuer             string
}

// NewManager creates a new JWT manager instance
func NewManager(config Config) *Manager {
	// Set defaults if not provided
	if config.AccessTokenTTL == 0 {
		config.AccessTokenTTL = 15 * time.Minute
	}
	if config.RefreshTokenTTL == 0 {
		config.RefreshTokenTTL = 7 * 24 * time.Hour // 7 days
	}
	if config.Issuer == "" {
		config.Issuer = "offera"
	}

	return &Manager{
		accessTokenSecret:  config.AccessTokenSecret,
		refreshTokenSecret: config.RefreshTokenSecret,
		accessTokenTTL:     config.AccessTokenTTL,
		refreshTokenTTL:    config.RefreshTokenTTL,
		issuer:             config.Issuer,
	}
}

// GenerateTokenPair creates both access and refresh tokens
func (m *Manager) GenerateTokenPair(userID, email, username, role string) (*TokenPair, error) {
	now := time.Now()
	expiresAt := now.Add(m.accessTokenTTL)

	// Generate access token
	accessToken, err := m.generateToken(userID, email, username, role, m.accessTokenSecret, expiresAt)
	if err != nil {
		return nil, err
	}

	// Generate refresh token with longer expiry
	refreshExpiresAt := now.Add(m.refreshTokenTTL)
	refreshToken, err := m.generateToken(userID, email, username, role, m.refreshTokenSecret, refreshExpiresAt)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}

func (m *Manager) GenerateNewTokenPair(refreshToken string) (*TokenPair, error) {
	claims, err := m.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	expiresAt := now.Add(m.accessTokenTTL)

	// Generate access token
	accessToken, err := m.generateToken(claims.UserID, claims.Email, claims.Username, claims.Role, m.accessTokenSecret, expiresAt)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: "",
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}

// generateToken creates a JWT token with the given claims
func (m *Manager) generateToken(userID, email, username, role, secret string, expiresAt time.Time) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:   userID,
		Email:    email,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ValidateAccessToken validates an access token and returns the claims
func (m *Manager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, m.accessTokenSecret)
}

// ValidateRefreshToken validates a refresh token and returns the claims
func (m *Manager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, m.refreshTokenSecret)
}

// validateToken validates a JWT token with the given secret
func (m *Manager) validateToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotYetValid
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// RefreshTokens generates a new token pair using a valid refresh token
func (m *Manager) RefreshTokens(refreshToken string) (*TokenPair, error) {
	claims, err := m.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Generate new token pair
	return m.GenerateTokenPair(claims.UserID, claims.Email, claims.Username, claims.Role)
}

// ExtractClaims extracts claims from a token without validation
// Use with caution - only for debugging or when validation is not required
func (m *Manager) ExtractClaims(tokenString string) (*Claims, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// GetTokenExpiry returns the expiry time from token claims
func (c *Claims) GetTokenExpiry() time.Time {
	if c.ExpiresAt != nil {
		return c.ExpiresAt.Time
	}
	return time.Time{}
}

// IsExpired checks if the token is expired
func (c *Claims) IsExpired() bool {
	if c.ExpiresAt == nil {
		return true
	}
	return time.Now().After(c.ExpiresAt.Time)
}

// TimeUntilExpiry returns the duration until token expiry
func (c *Claims) TimeUntilExpiry() time.Duration {
	if c.ExpiresAt == nil {
		return 0
	}
	return time.Until(c.ExpiresAt.Time)
}
