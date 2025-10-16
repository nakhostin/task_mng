package task

import "task_mng/domain/task/entity"

type Filter struct {
	Assignee *uint            `json:"assignee,omitempty"`
	Status   *entity.Status   `json:"status,omitempty"`
	Priority *entity.Priority `json:"priority,omitempty"`
}

type Repository interface {
	Create(e *entity.Task) error
	Update(e entity.Task) error
	FindByID(id uint) (entity.Task, error)
	FindAll(filter *Filter, page, limit int) ([]entity.Task, int64, error)
	Delete(e entity.Task) error
	CountByStatus() (map[entity.Status]int64, error)
}
