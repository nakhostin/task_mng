package aggregate_test

import (
	"task_mng/domain/user/aggregate"
	"task_mng/pkg/jwt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Unit tests for AuthResponse

func TestNewAuthResponse(t *testing.T) {
	user := &aggregate.UserResponse{
		ID:       1,
		Username: "n.nakhostin",
		FullName: "Nima Nakhostin",
		Email:    "nakhostin.nima1998@gmail.com",
	}

	tokenPair := &jwt.TokenPair{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
	}

	resp := aggregate.NewAuthResponse(user, tokenPair)

	assert.Equal(t, user, resp.User)
	assert.Equal(t, tokenPair, resp.Tokens)
}
