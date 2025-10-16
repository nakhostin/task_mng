package task

import (
	"context"
	"fmt"
	"task_mng/domain/task/entity"
	"task_mng/domain/task/mocks"
	userMocks "task_mng/domain/user/mocks"
	redisMocks "task_mng/pkg/redis/mocks"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestGetCacheVersion_FirstTime(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	setCalled := false
	redisMock := &redisMocks.MockRedisClient{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == tasksCacheVersionKey {
				return "", goredis.Nil
			}
			return "", fmt.Errorf("unexpected key")
		},
		SetFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
			if key == tasksCacheVersionKey {
				setCalled = true
				assert.Equal(t, "1", value)
				return nil
			}
			return fmt.Errorf("unexpected key")
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	version, err := service.getCacheVersion(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "1", version)
	assert.True(t, setCalled, "Set should be called to initialize version")
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestGetCacheVersion_ExistingVersion(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == tasksCacheVersionKey {
				return "5", nil
			}
			return "", fmt.Errorf("unexpected key")
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	version, err := service.getCacheVersion(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "5", version)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestGetCacheVersion_RedisError(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			return "", fmt.Errorf("redis connection error")
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	version, err := service.getCacheVersion(context.Background())

	assert.Error(t, err)
	assert.Equal(t, "", version)
	assert.Equal(t, "redis connection error", err.Error())
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestGenerateCacheKey_AllFiltersNil(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == tasksCacheVersionKey {
				return "1", nil
			}
			return "", fmt.Errorf("unexpected key")
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	filter := &FilterRequest{}
	key, err := service.generateCacheKey(context.Background(), filter, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, "tasks:list:v1:assignee:nil:status:nil:priority:nil:page:1:limit:10", key)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestGenerateCacheKey_WithFilters(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == tasksCacheVersionKey {
				return "2", nil
			}
			return "", fmt.Errorf("unexpected key")
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	assignee := "john.doe"
	status := entity.StatusInProgress
	priority := entity.PriorityHigh

	filter := &FilterRequest{
		Assignee: &assignee,
		Status:   &status,
		Priority: &priority,
	}
	key, err := service.generateCacheKey(context.Background(), filter, 2, 20)

	assert.NoError(t, err)
	assert.Equal(t, "tasks:list:v2:assignee:john.doe:status:InProgress:priority:high:page:2:limit:20", key)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestGenerateCacheKey_PartialFilters(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == tasksCacheVersionKey {
				return "3", nil
			}
			return "", fmt.Errorf("unexpected key")
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	status := entity.StatusDone

	filter := &FilterRequest{
		Status: &status,
	}
	key, err := service.generateCacheKey(context.Background(), filter, 1, 15)

	assert.NoError(t, err)
	assert.Equal(t, "tasks:list:v3:assignee:nil:status:Done:priority:nil:page:1:limit:15", key)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestGenerateCacheKey_RedisError(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			return "", fmt.Errorf("redis error")
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	filter := &FilterRequest{}
	key, err := service.generateCacheKey(context.Background(), filter, 1, 10)

	assert.Error(t, err)
	assert.Equal(t, "", key)
	assert.Equal(t, "redis error", err.Error())
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestInvalidateTasksCache_Success(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	incrCalled := false
	redisMock := &redisMocks.MockRedisClient{
		IncrFunc: func(ctx context.Context, key string) error {
			if key == tasksCacheVersionKey {
				incrCalled = true
				return nil
			}
			return fmt.Errorf("unexpected key")
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	service.invalidateTasksCache()

	assert.True(t, incrCalled, "Incr should be called")
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestInvalidateTasksCache_RedisError(t *testing.T) {
	mockRepo := new(mocks.MockTaskRepository)
	redisMock := &redisMocks.MockRedisClient{
		IncrFunc: func(ctx context.Context, key string) error {
			return fmt.Errorf("redis connection error")
		},
	}
	mockUserRepo := new(userMocks.MockUserRepository)

	mockRepo.On("CountByStatus").Return(map[entity.Status]int64{}, nil)

	service := New(mockRepo, redisMock, mockUserRepo)

	service.invalidateTasksCache()

	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}
