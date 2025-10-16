package aggregate

import (
	"task_mng/domain/task/entity"
	"task_mng/pkg/response"
	"time"
)

// AssigneeInfo represents the assignee user information in task responses
type AssigneeInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

type TaskResponse struct {
	ID          uint            `json:"id"`
	Summary     string          `json:"summary"`
	Description string          `json:"description"`
	Assignee    AssigneeInfo    `json:"assignee"`
	Status      entity.Status   `json:"status"`
	Priority    entity.Priority `json:"priority"`
	DueDate     time.Time       `json:"due_date"`
	CreatedAt   time.Time       `json:"created_at"`
}

func NewTaskResponse(task *entity.Task, assigneeUsername string) *TaskResponse {
	return &TaskResponse{
		ID:          task.ID,
		Summary:     task.Summary,
		Description: task.Description,
		Assignee: AssigneeInfo{
			ID:       task.Assignee,
			Username: assigneeUsername,
		},
		Status:    task.Status,
		Priority:  task.Priority,
		DueDate:   task.DueDate,
		CreatedAt: task.CreatedAt,
	}
}

type TaskListResponse struct {
	Tasks []*TaskResponse `json:"tasks"`
	Meta  *response.Meta  `json:"-"`
}

func NewTaskListResponse(tasks []entity.Task, assigneeUsernames map[uint]string, page, limit int, count int64, sort string) *TaskListResponse {
	taskResponses := make([]*TaskResponse, len(tasks))
	for i, task := range tasks {
		username := assigneeUsernames[task.Assignee]
		taskResponses[i] = NewTaskResponse(&task, username)
	}
	return &TaskListResponse{
		Tasks: taskResponses,
		Meta:  response.NewMeta(page, limit, int(count), sort),
	}
}
