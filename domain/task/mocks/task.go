package mocks

import (
	"task_mng/domain/task"
	"task_mng/domain/task/entity"

	"github.com/stretchr/testify/mock"
)

type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(e *entity.Task) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockTaskRepository) Update(e entity.Task) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockTaskRepository) FindByID(id uint) (entity.Task, error) {
	args := m.Called(id)
	return args.Get(0).(entity.Task), args.Error(1)
}

func (m *MockTaskRepository) FindAll(filter *task.Filter, page, limit int) ([]entity.Task, int64, error) {
	args := m.Called(filter, page, limit)
	return args.Get(0).([]entity.Task), args.Get(1).(int64), args.Error(2)
}

func (m *MockTaskRepository) Delete(e entity.Task) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockTaskRepository) CountByStatus() (map[entity.Status]int64, error) {
	args := m.Called()
	return args.Get(0).(map[entity.Status]int64), args.Error(1)
}
