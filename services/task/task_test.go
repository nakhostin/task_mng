package task

import (
	"context"
	"fmt"
	"task_mng/domain/task"
	"task_mng/domain/task/entity"
	"task_mng/domain/task/mocks"
	userEntity "task_mng/domain/user/entity"
	userMocks "task_mng/domain/user/mocks"
	redisMocks "task_mng/pkg/redis/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestCreateTask_Success(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := new(redisMocks.MockRedisClient)
	mockUserRepo := new(userMocks.MockUserRepository)

	// Mock CountByStatus for metrics initialization in New() and after Create()
	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil).Twice()

	service := New(mockRepo, redisMock, mockUserRepo)

	priority := entity.PriorityMedium
	dueDate := time.Now().Add(time.Hour * 24)

	req := &CreateRequest{
		Summary:     "Test Task",
		Description: "Test Description",
		Assignee:    "test.user",
		Priority:    &priority,
		DueDate:     &dueDate,
	}

	// Mock user repository to return a valid user
	mockUserRepo.On("FindByUsername", req.Assignee).Return(userEntity.User{
		Model:    gorm.Model{ID: 1},
		Username: req.Assignee,
	}, nil)

	mockRepo.On("Create", mock.MatchedBy(func(t *entity.Task) bool {
		return t.Summary == req.Summary &&
			t.Description == req.Description &&
			t.Priority == *req.Priority &&
			t.Assignee == 1 &&
			t.Status == entity.StatusTodo &&
			t.DueDate.Equal(*req.DueDate)
	})).Return(nil)

	err := service.Create(req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestCreateTask_UserNotFound(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := new(redisMocks.MockRedisClient)
	mockUserRepo := new(userMocks.MockUserRepository)

	// Mock CountByStatus for metrics initialization in New()
	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	priority := entity.PriorityMedium
	dueDate := time.Now().Add(time.Hour * 24)

	req := &CreateRequest{
		Summary:     "Test Task",
		Description: "Test Description",
		Assignee:    "test.user",
		Priority:    &priority,
		DueDate:     &dueDate,
	}

	// Mock user repository to return user not found error
	mockUserRepo.On("FindByUsername", req.Assignee).Return(userEntity.User{}, gorm.ErrRecordNotFound)

	err := service.Create(req)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "can't find assignee user")
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUpdateTask_Success(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		IncrFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	taskID := uint(1)
	priority := entity.PriorityMedium
	dueDate := time.Now().Add(time.Hour * 24)

	req := &UpdateRequest{
		Summary:     "Test Task",
		Description: "Test Description",
		Assignee:    "test.user",
		Priority:    priority,
		DueDate:     dueDate,
	}

	// Mock user repository to return a valid user
	mockUserRepo.On("FindByUsername", req.Assignee).Return(userEntity.User{
		Model:    gorm.Model{ID: 1},
		Username: req.Assignee,
	}, nil)

	mockRepo.On("FindByID", taskID).Return(entity.Task{
		Model:       gorm.Model{ID: taskID},
		Summary:     req.Summary,
		Description: req.Description,
		Priority:    priority,
		Assignee:    1,
		Status:      entity.StatusTodo,
		DueDate:     dueDate,
	}, nil)

	mockRepo.On("Update", mock.MatchedBy(func(t entity.Task) bool {
		return t.Summary == req.Summary &&
			t.Description == req.Description &&
			t.Priority == priority &&
			t.Assignee == 1 &&
			t.Status == entity.StatusTodo &&
			t.DueDate.Equal(dueDate)
	})).Return(nil)

	err := service.Update(req, "1")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUpdateTask_UserNotFound(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		IncrFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	taskID := uint(1)
	priority := entity.PriorityMedium
	dueDate := time.Now().Add(time.Hour * 24)

	req := &UpdateRequest{
		Summary:     "Test Task",
		Description: "Test Description",
		Assignee:    "nonexistent.user",
		Priority:    priority,
		DueDate:     dueDate,
	}

	mockRepo.On("FindByID", taskID).Return(entity.Task{
		Model:       gorm.Model{ID: taskID},
		Summary:     "Old Summary",
		Description: "Old Description",
		Priority:    priority,
		Assignee:    1,
		Status:      entity.StatusTodo,
		DueDate:     dueDate,
	}, nil)

	// Mock user repository to return error (user not found)
	mockUserRepo.On("FindByUsername", req.Assignee).Return(userEntity.User{}, fmt.Errorf("record not found"))

	err := service.Update(req, "1")

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "can't find assignee user")
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUpdateTask_TaskNotFound(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		IncrFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	taskID := uint(1)
	priority := entity.PriorityMedium
	dueDate := time.Now().Add(time.Hour * 24)

	req := &UpdateRequest{
		Summary:     "Test Task",
		Description: "Test Description",
		Assignee:    "test.user",
		Priority:    priority,
		DueDate:     dueDate,
	}

	// Mock FindByID to return error (task not found)
	mockRepo.On("FindByID", taskID).Return(entity.Task{}, gorm.ErrRecordNotFound)

	err := service.Update(req, "1")

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "task_not_found")
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestFindByID_Success(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := new(redisMocks.MockRedisClient)
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	taskID := uint(1)
	assigneeID := uint(1)

	dueDate := time.Now().Add(time.Hour * 24)
	mockRepo.On("FindByID", taskID).Return(entity.Task{
		Model:       gorm.Model{ID: taskID},
		Summary:     "Summary",
		Description: "Description",
		Priority:    entity.PriorityMedium,
		Assignee:    assigneeID,
		Status:      entity.StatusTodo,
		DueDate:     dueDate,
	}, nil)

	// Mock user repository to return assignee user
	mockUserRepo.On("FindByID", assigneeID).Return(userEntity.User{
		Model:    gorm.Model{ID: assigneeID},
		Username: "test.user",
	}, nil)

	task, err := service.FindByID("1")

	assert.NoError(t, err)
	assert.Equal(t, taskID, task.ID)
	assert.Equal(t, "Summary", task.Summary)
	assert.Equal(t, "Description", task.Description)
	assert.Equal(t, entity.PriorityMedium, task.Priority)
	assert.Equal(t, uint(1), task.Assignee.ID)
	assert.Equal(t, "test.user", task.Assignee.Username)
	assert.Equal(t, entity.StatusTodo, task.Status)
	assert.Equal(t, dueDate, task.DueDate)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestFindByID_TaskNotFound(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := new(redisMocks.MockRedisClient)
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	taskID := uint(1)

	mockRepo.On("FindByID", taskID).Return(entity.Task{}, gorm.ErrRecordNotFound)

	_, err := service.FindByID("1")

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "task_not_found")
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestFindAll_Success(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == "tasks:cache:version" {
				return "1", nil
			}
			return "", fmt.Errorf("cache miss")
		},
		SetFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
			return nil
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	dueDate := time.Now().Add(time.Hour * 24)
	mockRepo.On("FindAll", mock.MatchedBy(func(filter *task.Filter) bool {
		return filter.Assignee == nil && filter.Status == nil && filter.Priority == nil
	}), 1, 10).Return([]entity.Task{
		{
			Model:       gorm.Model{ID: 1},
			Summary:     "Test Task",
			Description: "Test Description",
			Priority:    entity.PriorityMedium,
			Assignee:    1,
			Status:      entity.StatusTodo,
			DueDate:     dueDate,
		},
	}, int64(1), nil)

	// Mock user repository to return users for batch fetch
	mockUserRepo.On("FindByIDs", []uint{1}).Return([]userEntity.User{
		{
			Model:    gorm.Model{ID: 1},
			Username: "test.user",
		},
	}, nil)

	req := &FilterRequest{}
	taskList, err := service.FindAll(req, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(taskList.Tasks))
	assert.Equal(t, "Test Task", taskList.Tasks[0].Summary)
	assert.Equal(t, "test.user", taskList.Tasks[0].Assignee.Username)
	assert.Equal(t, uint(1), taskList.Tasks[0].Assignee.ID)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestFindAll_EmptyResult(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == "tasks:cache:version" {
				return "1", nil
			}
			return "", fmt.Errorf("cache miss")
		},
		SetFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
			return nil
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	mockRepo.On("FindAll", mock.MatchedBy(func(filter *task.Filter) bool {
		return filter.Assignee == nil && filter.Status == nil && filter.Priority == nil
	}), 1, 10).Return([]entity.Task{}, int64(0), nil)

	// No need to mock FindByIDs since there are no tasks (empty array)

	req := &FilterRequest{}
	taskList, err := service.FindAll(req, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(taskList.Tasks))
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestDeleteTask_Success(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		IncrFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	taskID := uint(1)
	dueDate := time.Now().Add(time.Hour * 24)

	taskToDelete := entity.Task{
		Model:       gorm.Model{ID: taskID},
		Summary:     "Task to Delete",
		Description: "Description",
		Priority:    entity.PriorityMedium,
		Assignee:    1,
		Status:      entity.StatusTodo,
		DueDate:     dueDate,
	}

	mockRepo.On("FindByID", taskID).Return(taskToDelete, nil)
	mockRepo.On("Delete", mock.MatchedBy(func(t entity.Task) bool {
		return t.ID == taskID
	})).Return(nil)

	err := service.Delete("1")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestDeleteTask_TaskNotFound(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		IncrFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	taskID := uint(1)

	mockRepo.On("FindByID", taskID).Return(entity.Task{}, gorm.ErrRecordNotFound)

	err := service.Delete("1")

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "task_not_found")
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestAssignTask_Success(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		IncrFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	taskID := uint(1)
	assigneeID := uint(2)

	mockRepo.On("FindByID", taskID).Return(entity.Task{
		Model:       gorm.Model{ID: taskID},
		Summary:     "Task to Assign",
		Description: "Description",
		Priority:    entity.PriorityMedium,
		Assignee:    1,
		Status:      entity.StatusTodo,
		DueDate:     time.Now().Add(time.Hour * 24),
	}, nil)

	mockUserRepo.On("FindByUsername", "assignee").Return(userEntity.User{
		Model:    gorm.Model{ID: assigneeID},
		Username: "assignee",
	}, nil)

	mockRepo.On("Update", mock.MatchedBy(func(t entity.Task) bool {
		return t.ID == taskID && t.Assignee == assigneeID
	})).Return(nil)

	err := service.Assign(&AssignRequest{
		TaskID:   taskID,
		Assignee: "assignee",
	})

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestAssignTask_TaskNotFound(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		IncrFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}

	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	taskID := uint(1)
	assigneeID := uint(2)

	// Mock user repository to return a valid user (checked first)
	mockUserRepo.On("FindByUsername", "assignee").Return(userEntity.User{
		Model:    gorm.Model{ID: assigneeID},
		Username: "assignee",
	}, nil)

	// Mock task repository to return error (task not found)
	mockRepo.On("FindByID", taskID).Return(entity.Task{}, gorm.ErrRecordNotFound)

	err := service.Assign(&AssignRequest{
		TaskID:   taskID,
		Assignee: "assignee",
	})

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "task_not_found")
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestAssignTask_UserNotFound(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		IncrFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	taskID := uint(1)
	assignee := "nonexistent.user"

	// Mock user repository to return error (user not found - checked first)
	mockUserRepo.On("FindByUsername", assignee).Return(userEntity.User{}, fmt.Errorf("record not found"))

	err := service.Assign(&AssignRequest{
		TaskID:   taskID,
		Assignee: assignee,
	})

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "can't find assignee user")
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestStatusTransition_Success(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		IncrFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	taskID := uint(1)

	mockRepo.On("FindByID", taskID).Return(entity.Task{
		Model:       gorm.Model{ID: taskID},
		Summary:     "Task to Transition",
		Description: "Description",
		Priority:    entity.PriorityMedium,
		Assignee:    1,
		Status:      entity.StatusTodo,
		DueDate:     time.Now().Add(time.Hour * 24),
	}, nil)

	mockRepo.On("Update", mock.MatchedBy(func(t entity.Task) bool {
		return t.ID == taskID && t.Status == entity.StatusInProgress
	})).Return(nil)

	err := service.StatusTransition(&StatusTransitionRequest{
		TaskID: taskID,
		Status: entity.StatusInProgress,
	})

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestStatusTransition_TaskNotFound(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		IncrFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	taskID := uint(1)

	mockRepo.On("FindByID", taskID).Return(entity.Task{}, gorm.ErrRecordNotFound)

	err := service.StatusTransition(&StatusTransitionRequest{
		TaskID: taskID,
		Status: entity.StatusInProgress,
	})

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "task_not_found")
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}
