package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"task_mng/domain/task"
	"task_mng/domain/task/aggregate"
	"task_mng/domain/task/entity"
	"task_mng/domain/user"
	"task_mng/pkg/metrics"
	"task_mng/pkg/redis"
	"task_mng/pkg/response"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Service struct {
	repository     task.Repository
	logger         *slog.Logger
	redis          redis.RedisClient
	userRepository user.Repository
}

func New(repository task.Repository, redis redis.RedisClient, userRepository user.Repository) *Service {
	s := &Service{repository: repository, logger: slog.Default(), redis: redis, userRepository: userRepository}
	// Initialize task count metrics on startup
	s.updateTaskMetrics()
	return s
}

// ********************* Create *********************

type CreateRequest struct {
	Summary     string           `json:"summary" valid:"required~summary_is_required" example:"Implement task management system"`
	Description string           `json:"description" example:"Implement task management system"`
	Assignee    string           `json:"assignee" valid:"required~assignee_is_required" example:"admin"`
	Priority    *entity.Priority `json:"priority" valid:"optional,in(lowest|low|medium|high|highest)~invalid_priority" example:"medium"`
	DueDate     *time.Time       `json:"due_date" example:"2025-01-01T00:00:00Z"`
}

func (s *Service) Create(req *CreateRequest) error {
	dueDate := time.Time{}
	if req.DueDate != nil {
		dueDate = *req.DueDate
	}

	priority := entity.PriorityMedium
	if req.Priority != nil {
		priority = *req.Priority
	}

	// check assignee if exists
	user, err := s.userRepository.FindByUsername(req.Assignee)
	if err != nil {
		s.logger.Error("error finding user", "error", err)
		return fmt.Errorf("can't find assignee user")
	}

	e := &entity.Task{
		Summary:     req.Summary,
		Description: req.Description,
		Assignee:    user.ID,
		Status:      entity.StatusTodo,
		Priority:    priority,
		DueDate:     dueDate,
	}

	err = s.repository.Create(e)
	if err != nil {
		return err
	}

	// Invalidate cache after creating a new task
	s.invalidateTasksCache()

	// Update task count metrics
	s.updateTaskMetrics()

	return nil
}

// ********************* Update *********************
type UpdateRequest struct {
	Summary     string          `json:"summary" valid:"required~summary_is_required" example:"Implement task management system"`
	Description string          `json:"description" example:"Implement task management system"`
	Assignee    string          `json:"assignee" valid:"required~assignee_is_required" example:"admin"`
	Priority    entity.Priority `json:"priority" valid:"optional,in(lowest|low|medium|high|highest)~invalid_priority" example:"medium"`
	DueDate     time.Time       `json:"due_date" example:"2025-01-01T00:00:00Z"`
}

func (s *Service) Update(req *UpdateRequest, id string) error {
	uintID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		s.logger.Error("error parsing id", "error", err)
		return fmt.Errorf("invalid_id")
	}

	task, err := s.repository.FindByID(uint(uintID))
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("error finding task", "error", err)
			return fmt.Errorf("internal_server_error")
		}
		s.logger.Error("task not found", "error", err)
		return fmt.Errorf("task_not_found")
	}

	user, err := s.userRepository.FindByUsername(req.Assignee)
	if err != nil {
		s.logger.Error("error finding user", "error", err)
		return fmt.Errorf("can't find assignee user")
	}

	task.Summary = req.Summary
	task.Description = req.Description
	task.Assignee = user.ID
	task.Priority = req.Priority
	task.DueDate = req.DueDate

	err = s.repository.Update(task)
	if err != nil {
		return err
	}

	// Invalidate cache after updating a task
	s.invalidateTasksCache()

	// Update task count metrics
	s.updateTaskMetrics()

	return nil
}

// ********************* Find By ID *********************
func (s *Service) FindByID(id string) (*aggregate.TaskResponse, error) {
	uintID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		s.logger.Error("error parsing id", "error", err)
		return nil, fmt.Errorf("invalid_id")
	}

	t, err := s.repository.FindByID(uint(uintID))
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("error finding task", "error", err)
			return nil, fmt.Errorf("internal_server_error")
		}
		s.logger.Error("task not found", "error", err)
		return nil, fmt.Errorf("task_not_found")
	}

	// Get assignee username
	assigneeUsername := ""
	user, err := s.userRepository.FindByID(t.Assignee)
	if err != nil {
		s.logger.Warn("assignee user not found", "assignee_id", t.Assignee, "error", err)
	} else {
		assigneeUsername = user.Username
	}

	return aggregate.NewTaskResponse(&t, assigneeUsername), nil
}

// ********************* Find All *********************
type FilterRequest struct {
	Assignee *string          `form:"assignee"`
	Status   *entity.Status   `form:"status"`
	Priority *entity.Priority `form:"priority"`
}

func (s *Service) FindAll(req *FilterRequest, page, limit int) (*aggregate.TaskListResponse, error) {
	ctx := context.Background()

	// Generate cache key based on filters, page, and limit
	cacheKey, err := s.generateCacheKey(ctx, req, page, limit)
	if err != nil {
		s.logger.Warn("Failed to generate cache key, proceeding without cache", "error", err)
	} else {
		cachedData, err := s.redis.Get(ctx, cacheKey)
		if err == nil {
			var cachedResponse cachedTasksResponse
			if err := json.Unmarshal([]byte(cachedData), &cachedResponse); err == nil {
				s.logger.Info("Cache hit for tasks list", "key", cacheKey)
				cachedResponse.Tasks.Meta = response.NewMeta(page, limit, int(cachedResponse.Count), "")
				return cachedResponse.Tasks, nil
			}
			s.logger.Warn("Failed to unmarshal cached data", "error", err)
		} else if !errors.Is(err, goredis.Nil) {
			s.logger.Warn("Failed to get from cache", "error", err)
		}
	}

	// Cache Miss or cache error - fetch from database
	s.logger.Info("Cache miss for tasks list, fetching from database")

	var assignee *uint
	if req.Assignee == nil {
		assignee = nil
	} else {
		user, err := s.userRepository.FindByUsername(*req.Assignee)
		if err == nil {
			assignee = &user.ID
		}
	}

	filter := &task.Filter{
		Assignee: assignee,
		Status:   req.Status,
		Priority: req.Priority,
	}

	tasks, count, err := s.repository.FindAll(filter, page, limit)
	if err != nil {
		s.logger.Error("error finding tasks", "error", err)
		return nil, err
	}

	assigneeIDs := make([]uint, 0)
	assigneeIDMap := make(map[uint]bool)
	for _, task := range tasks {
		if !assigneeIDMap[task.Assignee] {
			assigneeIDs = append(assigneeIDs, task.Assignee)
			assigneeIDMap[task.Assignee] = true
		}
	}

	assigneeUsernames := make(map[uint]string)
	if len(assigneeIDs) > 0 {
		users, err := s.userRepository.FindByIDs(assigneeIDs)
		if err != nil {
			s.logger.Warn("error finding assignee users", "error", err)
		} else {
			for _, user := range users {
				assigneeUsernames[user.ID] = user.Username
			}
		}
	}

	aggregatedTasks := aggregate.NewTaskListResponse(tasks, assigneeUsernames, page, limit, count, "")

	if cacheKey != "" {
		cachedResponse := cachedTasksResponse{
			Tasks: aggregatedTasks,
			Count: count,
		}

		cachedData, err := json.Marshal(cachedResponse)
		if err != nil {
			s.logger.Warn("Failed to marshal tasks for caching", "error", err)
		} else {
			err = s.redis.Set(ctx, cacheKey, cachedData, tasksCacheTTL)
			if err != nil {
				s.logger.Warn("Failed to store in cache", "error", err)
			} else {
				s.logger.Info("Tasks cached successfully", "key", cacheKey, "ttl", tasksCacheTTL)
			}
		}
	}

	return aggregatedTasks, nil
}

// ********************* Delete *********************
func (s *Service) Delete(id string) error {
	uintID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		s.logger.Error("error parsing id", "error", err)
		return fmt.Errorf("invalid_id")
	}

	t, err := s.repository.FindByID(uint(uintID))
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("error finding task", "error", err)
			return fmt.Errorf("internal_server_error")
		}
		s.logger.Error("task not found", "error", err)
		return fmt.Errorf("task_not_found")
	}

	err = s.repository.Delete(t)
	if err != nil {
		return err
	}

	// Invalidate cache after deleting a task
	s.invalidateTasksCache()

	// Update task count metrics
	s.updateTaskMetrics()

	return nil
}

// ********************* Assign *********************
type AssignRequest struct {
	TaskID   uint   `json:"task_id" valid:"required~task_id_is_required" example:"1"`
	Assignee string `json:"assignee" valid:"required~assignee_is_required" example:"admin"`
}

func (s *Service) Assign(req *AssignRequest) error {
	user, err := s.userRepository.FindByUsername(req.Assignee)
	if err != nil {
		s.logger.Error("error finding user", "error", err)
		return fmt.Errorf("can't find assignee user")
	}

	task, err := s.repository.FindByID(req.TaskID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("error finding task", "error", err)
			return fmt.Errorf("internal_server_error")
		}
		s.logger.Error("task not found", "error", err)
		return fmt.Errorf("task_not_found")
	}

	task.Assignee = user.ID
	err = s.repository.Update(task)
	if err != nil {
		return err
	}

	// Invalidate cache after assigning a task
	s.invalidateTasksCache()

	// Update task count metrics
	s.updateTaskMetrics()

	return nil
}

// ********************* Status Transition *********************
type StatusTransitionRequest struct {
	TaskID uint          `json:"task_id" valid:"required~task_id_is_required" example:"1"`
	Status entity.Status `json:"status" valid:"required~status_is_required,in(ToDo|InProgress|Done)~invalid_status" example:"InProgress"`
}

func (s *Service) StatusTransition(req *StatusTransitionRequest) error {
	task, err := s.repository.FindByID(req.TaskID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("error finding task", "error", err)
			return fmt.Errorf("internal_server_error")
		}
		s.logger.Error("task not found", "error", err)
		return fmt.Errorf("task_not_found")
	}

	task.Status = req.Status
	err = s.repository.Update(task)
	if err != nil {
		return err
	}

	// Invalidate cache after transitioning task status
	s.invalidateTasksCache()

	// Update task count metrics
	s.updateTaskMetrics()

	return nil
}

// ********************* Helper: Update Task Metrics *********************
func (s *Service) updateTaskMetrics() {
	counts, err := s.repository.CountByStatus()
	if err != nil {
		s.logger.Error("failed to get task counts for metrics", "error", err)
		return
	}

	for status, count := range counts {
		metrics.UpdateTasksCount(string(status), float64(count))
	}
}
