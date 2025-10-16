package aggregate_test

import (
	"task_mng/domain/user/aggregate"
	"task_mng/domain/user/entity"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNewUserResponse(t *testing.T) {
	user := &entity.User{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
		},
		Username: "n.nakhostin",
		FullName: "Nima Nakhostin",
		Email:    "nakhostin.nima1998@gmail.com",
	}

	resp := aggregate.NewUserResponse(user)

	assert.Equal(t, user.ID, resp.ID)
	assert.Equal(t, user.Username, resp.Username)
	assert.Equal(t, user.FullName, resp.FullName)
	assert.Equal(t, user.Email, resp.Email)
	assert.Equal(t, user.CreatedAt, resp.RegisteredAt)
}
