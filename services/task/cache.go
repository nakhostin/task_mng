package task

import (
	"context"
	"errors"
	"fmt"
	"task_mng/domain/task/aggregate"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	tasksCacheVersionKey = "tasks:cache:version"
	tasksCacheTTL        = 1 * time.Hour
)

type cachedTasksResponse struct {
	Tasks *aggregate.TaskListResponse `json:"tasks"`
	Count int64                       `json:"count"`
}

// getCacheVersion gets the current cache version from Redis
func (s *Service) getCacheVersion(ctx context.Context) (string, error) {
	version, err := s.redis.Get(ctx, tasksCacheVersionKey)
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			version = "1"
			_ = s.redis.Set(ctx, tasksCacheVersionKey, version, 0)
		} else {
			return "", err
		}
	}
	return version, nil
}

// generateCacheKey generates a unique cache key for tasks list based on filters and cache version
func (s *Service) generateCacheKey(ctx context.Context, filter *FilterRequest, page, limit int) (string, error) {
	version, err := s.getCacheVersion(ctx)
	if err != nil {
		return "", err
	}

	assignee := "nil"
	if filter.Assignee != nil {
		assignee = *filter.Assignee
	}

	status := "nil"
	if filter.Status != nil {
		status = string(*filter.Status)
	}

	priority := "nil"
	if filter.Priority != nil {
		priority = string(*filter.Priority)
	}

	return fmt.Sprintf("tasks:list:v%s:assignee:%s:status:%s:priority:%s:page:%d:limit:%d",
		version, assignee, status, priority, page, limit), nil
}

// invalidateTasksCache invalidates all tasks cache entries by incrementing the cache version
func (s *Service) invalidateTasksCache() {
	ctx := context.Background()

	err := s.redis.Incr(ctx, tasksCacheVersionKey)
	if err != nil {
		s.logger.Error("Failed to invalidate tasks cache", "error", err)
		return
	}

	s.logger.Info("Tasks cache invalidated successfully")
}
