package aggregate

import (
	"task_mng/domain/user/entity"
	"task_mng/pkg/response"
	"time"
)

type UserResponse struct {
	ID           uint      `json:"id"`
	FullName     string    `json:"full_name"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	RegisteredAt time.Time `json:"registered_at"`
}

type UserListResponse struct {
	Users []*UserResponse `json:"users"`
	Meta  *response.Meta  `json:"meta"`
}

func NewUserResponse(user *entity.User) *UserResponse {
	return &UserResponse{
		ID:           user.ID,
		FullName:     user.FullName,
		Username:     user.Username,
		Email:        user.Email,
		RegisteredAt: user.CreatedAt,
	}
}

func NewUserListResponse(users []entity.User, page, limit int, count int64, sort string) *UserListResponse {
	userResponses := make([]*UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = NewUserResponse(&user)
	}
	return &UserListResponse{
		Users: userResponses,
		Meta:  response.NewMeta(page, limit, int(count), sort),
	}
}
