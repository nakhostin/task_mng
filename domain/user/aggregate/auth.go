package aggregate

import (
	"task_mng/pkg/jwt"
)

type AuthResponse struct {
	User   *UserResponse  `json:"user,omitempty"`
	Tokens *jwt.TokenPair `json:"tokens,omitempty"`
}

func NewAuthResponse(user *UserResponse, tokens *jwt.TokenPair) *AuthResponse {
	return &AuthResponse{
		User:   user,
		Tokens: tokens,
	}
}
