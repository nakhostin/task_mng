package handlers

import (
	"task_mng/services/task"
	"task_mng/services/user"
)

type Handlers struct {
	User *UserHandler
	Task *TaskHandler
}

func New(
	userService *user.Service,
	taskService *task.Service,
) *Handlers {
	return &Handlers{
		User: NewUserHandler(userService),
		Task: NewTaskHandler(taskService),
	}
}
