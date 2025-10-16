package mocks

import (
	"task_mng/pkg/jwt"
)

// MockJWTManager is a mock implementation of jwt.JWTManager
type MockJWTManager struct {
	GenerateTokenPairFunc    func(userID, email, username, role string) (*jwt.TokenPair, error)
	GenerateNewTokenPairFunc func(refreshToken string) (*jwt.TokenPair, error)
	ValidateAccessTokenFunc  func(token string) (*jwt.Claims, error)
	ValidateRefreshTokenFunc func(token string) (*jwt.Claims, error)
	RefreshTokensFunc        func(refreshToken string) (*jwt.TokenPair, error)
	ExtractClaimsFunc        func(tokenString string) (*jwt.Claims, error)
}

func (m *MockJWTManager) GenerateTokenPair(userID, email, username, role string) (*jwt.TokenPair, error) {
	if m.GenerateTokenPairFunc != nil {
		return m.GenerateTokenPairFunc(userID, email, username, role)
	}
	return &jwt.TokenPair{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
	}, nil
}

func (m *MockJWTManager) GenerateNewTokenPair(refreshToken string) (*jwt.TokenPair, error) {
	if m.GenerateNewTokenPairFunc != nil {
		return m.GenerateNewTokenPairFunc(refreshToken)
	}
	return &jwt.TokenPair{
		AccessToken:  "mock_new_access_token",
		RefreshToken: "mock_new_refresh_token",
	}, nil
}

func (m *MockJWTManager) ValidateAccessToken(token string) (*jwt.Claims, error) {
	if m.ValidateAccessTokenFunc != nil {
		return m.ValidateAccessTokenFunc(token)
	}
	return &jwt.Claims{
		UserID:   "1",
		Email:    "test@example.com",
		Username: "testuser",
		Role:     "",
	}, nil
}

func (m *MockJWTManager) ValidateRefreshToken(token string) (*jwt.Claims, error) {
	if m.ValidateRefreshTokenFunc != nil {
		return m.ValidateRefreshTokenFunc(token)
	}
	return &jwt.Claims{
		UserID:   "1",
		Email:    "test@example.com",
		Username: "testuser",
		Role:     "",
	}, nil
}

func (m *MockJWTManager) RefreshTokens(refreshToken string) (*jwt.TokenPair, error) {
	if m.RefreshTokensFunc != nil {
		return m.RefreshTokensFunc(refreshToken)
	}
	return &jwt.TokenPair{
		AccessToken:  "mock_refreshed_access_token",
		RefreshToken: "mock_refreshed_refresh_token",
	}, nil
}

func (m *MockJWTManager) ExtractClaims(tokenString string) (*jwt.Claims, error) {
	if m.ExtractClaimsFunc != nil {
		return m.ExtractClaimsFunc(tokenString)
	}
	return &jwt.Claims{
		UserID:   "1",
		Email:    "test@example.com",
		Username: "testuser",
		Role:     "",
	}, nil
}
