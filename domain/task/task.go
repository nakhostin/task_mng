package task

import (
	"task_mng/domain/task/entity"
	"task_mng/pkg/postgres"

	"gorm.io/gorm"
)

type repository struct {
	db *postgres.Database
}

func New(db *postgres.Database) Repository {
	return &repository{db: db}
}

func (r *repository) Create(e *entity.Task) error {
	return r.db.Create(e).Error
}

func (r *repository) FindByID(id uint) (entity.Task, error) {
	var task entity.Task
	err := r.db.Where("id = ?", id).First(&task).Error
	return task, err
}

func (r *repository) FindAll(filter *Filter, page, limit int) ([]entity.Task, int64, error) {
	var tasks []entity.Task
	var count int64

	offset := (page - 1) * limit

	query := r.buildQuery(filter)

	err := query.Count(&count).Error
	if err != nil {
		return tasks, count, err
	}

	err = query.Offset(offset).Limit(limit).Find(&tasks).Error
	return tasks, count, err
}

func (r *repository) Update(e entity.Task) error {
	return r.db.Save(&e).Error
}

func (r *repository) Delete(e entity.Task) error {
	return r.db.Delete(&e).Error
}

func (r *repository) CountByStatus() (map[entity.Status]int64, error) {
	counts := make(map[entity.Status]int64)

	// Count for each status
	statuses := []entity.Status{entity.StatusTodo, entity.StatusInProgress, entity.StatusDone}

	for _, status := range statuses {
		var count int64
		err := r.db.Model(&entity.Task{}).Where("status = ?", status).Count(&count).Error
		if err != nil {
			return nil, err
		}
		counts[status] = count
	}

	return counts, nil
}

// Helper functions
func (r *repository) buildQuery(filter *Filter) *gorm.DB {
	query := r.db.Model(&entity.Task{})

	if filter.Assignee != nil {
		query = query.Where("assignee = ?", *filter.Assignee)
	}

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if filter.Priority != nil {
		query = query.Where("priority = ?", *filter.Priority)
	}

	return query
}
