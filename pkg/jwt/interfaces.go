package jwt

// JWTManager defines the interface for JWT operations
type JWTManager interface {
	GenerateTokenPair(userID, email, username, role string) (*TokenPair, error)
	GenerateNewTokenPair(refreshToken string) (*TokenPair, error)
	ValidateAccessToken(tokenString string) (*Claims, error)
	ValidateRefreshToken(tokenString string) (*Claims, error)
	RefreshTokens(refreshToken string) (*TokenPair, error)
	ExtractClaims(tokenString string) (*Claims, error)
}
