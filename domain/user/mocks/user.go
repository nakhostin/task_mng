package mocks

import (
	"task_mng/domain/user/entity"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of user.Repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(e *entity.User) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(email string) (entity.User, error) {
	args := m.Called(email)
	return args.Get(0).(entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(username string) (entity.User, error) {
	args := m.Called(username)
	return args.Get(0).(entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id uint) (entity.User, error) {
	args := m.Called(id)
	return args.Get(0).(entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByIDs(ids []uint) ([]entity.User, error) {
	args := m.Called(ids)
	return args.Get(0).([]entity.User), args.Error(1)
}

func (m *MockUserRepository) FindAll(page, limit int) ([]entity.User, int64, error) {
	args := m.Called(page, limit)
	return args.Get(0).([]entity.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) Update(e entity.User) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}
