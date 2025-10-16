package handlers

import (
	"task_mng/services/user"
)

type Handlers struct {
	User *UserHandler
}

func New(
	userService *user.Service,
) *Handlers {
	return &Handlers{
		User: NewUserHandler(userService),
	}
}
