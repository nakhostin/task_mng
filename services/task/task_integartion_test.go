package task_test

import (
	"fmt"
	taskR "task_mng/domain/task"
	"task_mng/domain/task/entity"
	userR "task_mng/domain/user"
	userEntity "task_mng/domain/user/entity"
	"task_mng/pkg/postgres"
	"task_mng/pkg/redis"
	"task_mng/services/task"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDatabase(t *testing.T) (*postgres.Database, func()) {
	db, err := postgres.New(postgres.Config{
		Host:     "0.0.0.0",
		Port:     "5432",
		User:     "admin",
		Password: "admin1234",
		Name:     "xdr",
	})

	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return db, func() {
		db.Close()
	}
}

func setupTestRedis(t *testing.T) (*redis.Redis, func()) {
	redis, err := redis.New(redis.Config{
		Host:     "0.0.0.0",
		Port:     "6379",
		Password: "admin1234",
	})

	if err != nil {
		t.Fatalf("Failed to create test Redis: %v", err)
	}

	return redis, func() {
		redis.Close()
	}
}

func cleanupDatabase(t *testing.T, db *postgres.Database) {
	err := db.GetDB().Exec("DELETE FROM tasks").Error
	require.NoError(t, err)

	err = db.GetDB().Exec("DELETE FROM users").Error
	require.NoError(t, err)
}

func createTestUser(t *testing.T, db *postgres.Database, username string) {
	user := &userEntity.User{
		Username: username,
		FullName: "Test User",
		Email:    username + "@example.com",
		Password: "hashedpassword",
	}

	userRepo := userR.New(db)
	err := userRepo.Create(user)
	require.NoError(t, err)
}

func setupTestService(t *testing.T) (*task.Service, *postgres.Database, func()) {
	db, dbCleanup := setupTestDatabase(t)
	redis, redisCleanup := setupTestRedis(t)

	cleanupDatabase(t, db)

	createTestUser(t, db, "admin")

	taskRepo := taskR.New(db)
	userRepo := userR.New(db)

	service := task.New(taskRepo, redis, userRepo)

	cleanup := func() {
		dbCleanup()
		redisCleanup()
	}

	return service, db, cleanup
}

func TestTaskIntegration_Success_CreateTask(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	priority := entity.PriorityMedium
	dueDate := time.Now().Add(time.Hour * 24)
	req := &task.CreateRequest{
		Summary:     "Test Task",
		Description: "Test Description",
		Assignee:    "admin",
		Priority:    &priority,
		DueDate:     &dueDate,
	}

	err := service.Create(req)

	assert.NoError(t, err)
}

func TestTaskIntegration_CreateTask_WithNonexistentAssignee(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	priority := entity.PriorityMedium
	dueDate := time.Now().Add(time.Hour * 24)

	req := &task.CreateRequest{
		Summary:     "Test Task",
		Description: "Test Description",
		Assignee:    "nonexistent",
		Priority:    &priority,
		DueDate:     &dueDate,
	}

	err := service.Create(req)

	assert.Error(t, err)
	assert.Equal(t, "can't find assignee user", err.Error())
}

func TestTaskIntegration_UpdateTask(t *testing.T) {
	service, db, cleanup := setupTestService(t)
	defer cleanup()

	priority := entity.PriorityLow
	dueDate := time.Now().Add(time.Hour * 24)
	createReq := &task.CreateRequest{
		Summary:     "Original Task",
		Description: "Original Description",
		Assignee:    "admin",
		Priority:    &priority,
		DueDate:     &dueDate,
	}

	err := service.Create(createReq)
	assert.NoError(t, err)

	var createdTask entity.Task
	err = db.GetDB().Where("summary = ?", "Original Task").First(&createdTask).Error
	require.NoError(t, err)

	updateReq := &task.UpdateRequest{
		Summary:     "Updated Task",
		Description: "Updated Description",
		Assignee:    "admin",
		Priority:    entity.PriorityHigh,
		DueDate:     time.Now().Add(time.Hour * 48),
	}

	err = service.Update(updateReq, fmt.Sprintf("%d", createdTask.ID))
	assert.NoError(t, err)

	var updatedTask entity.Task
	err = db.GetDB().First(&updatedTask, createdTask.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "Updated Task", updatedTask.Summary)
	assert.Equal(t, entity.PriorityHigh, updatedTask.Priority)
}

func TestTaskIntegration_FindAll(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	for i := 1; i <= 3; i++ {
		priority := entity.PriorityMedium
		dueDate := time.Now().Add(time.Hour * 24)
		req := &task.CreateRequest{
			Summary:     fmt.Sprintf("Task %d", i),
			Description: fmt.Sprintf("Description %d", i),
			Assignee:    "admin",
			Priority:    &priority,
			DueDate:     &dueDate,
		}

		err := service.Create(req)
		assert.NoError(t, err)
	}

	filter := &task.FilterRequest{}
	result, err := service.FindAll(filter, 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Tasks), 3)
}

func TestTaskIntegration_StatusTransition(t *testing.T) {
	service, db, cleanup := setupTestService(t)
	defer cleanup()

	priority := entity.PriorityMedium
	dueDate := time.Now().Add(time.Hour * 24)
	createReq := &task.CreateRequest{
		Summary:     "Task for Status Change",
		Description: "Testing status transition",
		Assignee:    "admin",
		Priority:    &priority,
		DueDate:     &dueDate,
	}

	err := service.Create(createReq)
	assert.NoError(t, err)

	var createdTask entity.Task
	err = db.GetDB().Where("summary = ?", "Task for Status Change").First(&createdTask).Error
	require.NoError(t, err)

	statusReq := &task.StatusTransitionRequest{
		TaskID: createdTask.ID,
		Status: entity.StatusInProgress,
	}

	err = service.StatusTransition(statusReq)
	assert.NoError(t, err)

	var updatedTask entity.Task
	err = db.GetDB().First(&updatedTask, createdTask.ID).Error
	require.NoError(t, err)
	assert.Equal(t, entity.StatusInProgress, updatedTask.Status)
}

func TestTaskIntegration_DeleteTask(t *testing.T) {
	service, db, cleanup := setupTestService(t)
	defer cleanup()

	priority := entity.PriorityMedium
	dueDate := time.Now().Add(time.Hour * 24)
	createReq := &task.CreateRequest{
		Summary:     "Task to Delete",
		Description: "This task will be deleted",
		Assignee:    "admin",
		Priority:    &priority,
		DueDate:     &dueDate,
	}

	err := service.Create(createReq)
	assert.NoError(t, err)

	var createdTask entity.Task
	err = db.GetDB().Where("summary = ?", "Task to Delete").First(&createdTask).Error
	require.NoError(t, err)

	err = service.Delete(fmt.Sprintf("%d", createdTask.ID))
	assert.NoError(t, err)

	var deletedTask entity.Task
	err = db.GetDB().Unscoped().First(&deletedTask, createdTask.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, deletedTask.DeletedAt)
}

func TestTaskIntegration_AssignTask(t *testing.T) {
	service, db, cleanup := setupTestService(t)
	defer cleanup()

	createTestUser(t, db, "bob")

	priority := entity.PriorityMedium
	dueDate := time.Now().Add(time.Hour * 24)
	createReq := &task.CreateRequest{
		Summary:     "Task to Reassign",
		Description: "This task will be reassigned",
		Assignee:    "admin",
		Priority:    &priority,
		DueDate:     &dueDate,
	}

	err := service.Create(createReq)
	assert.NoError(t, err)

	var createdTask entity.Task
	err = db.GetDB().Where("summary = ?", "Task to Reassign").First(&createdTask).Error
	require.NoError(t, err)

	var bobUser userEntity.User
	err = db.GetDB().Where("username = ?", "bob").First(&bobUser).Error
	require.NoError(t, err)

	assignReq := &task.AssignRequest{
		TaskID:   createdTask.ID,
		Assignee: "bob",
	}

	err = service.Assign(assignReq)
	assert.NoError(t, err)

	var reassignedTask entity.Task
	err = db.GetDB().First(&reassignedTask, createdTask.ID).Error
	require.NoError(t, err)
	assert.Equal(t, bobUser.ID, reassignedTask.Assignee)
}

func TestTaskIntegration_FindByID(t *testing.T) {
	service, db, cleanup := setupTestService(t)
	defer cleanup()

	priority := entity.PriorityHigh
	dueDate := time.Now().Add(time.Hour * 48)
	createReq := &task.CreateRequest{
		Summary:     "Find Me Task",
		Description: "Testing FindByID",
		Assignee:    "admin",
		Priority:    &priority,
		DueDate:     &dueDate,
	}

	err := service.Create(createReq)
	assert.NoError(t, err)

	var createdTask entity.Task
	err = db.GetDB().Where("summary = ?", "Find Me Task").First(&createdTask).Error
	require.NoError(t, err)

	result, err := service.FindByID(fmt.Sprintf("%d", createdTask.ID))
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Find Me Task", result.Summary)
	assert.Equal(t, entity.PriorityHigh, result.Priority)
}

func TestTaskIntegration_FindByID_NotFound(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	result, err := service.FindByID("99999")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "task_not_found", err.Error())
}

func TestTaskIntegration_FindAll_WithFilters(t *testing.T) {
	service, db, cleanup := setupTestService(t)
	defer cleanup()

	createTestUser(t, db, "alice")

	tasks := []struct {
		Summary  string
		Assignee string
		Priority entity.Priority
		Status   entity.Status
	}{
		{"High Priority Task", "admin", entity.PriorityHigh, entity.StatusTodo},
		{"Medium Priority Task", "alice", entity.PriorityMedium, entity.StatusInProgress},
		{"Low Priority Task", "admin", entity.PriorityLow, entity.StatusTodo},
		{"Done Task", "alice", entity.PriorityMedium, entity.StatusDone},
	}

	for _, tc := range tasks {
		dueDate := time.Now().Add(time.Hour * 24)
		priority := tc.Priority
		req := &task.CreateRequest{
			Summary:     tc.Summary,
			Description: "Test",
			Assignee:    tc.Assignee,
			Priority:    &priority,
			DueDate:     &dueDate,
		}
		err := service.Create(req)
		require.NoError(t, err)

		if tc.Status != entity.StatusTodo {
			var createdTask entity.Task
			err = db.GetDB().Where("summary = ?", tc.Summary).First(&createdTask).Error
			require.NoError(t, err)

			statusReq := &task.StatusTransitionRequest{
				TaskID: createdTask.ID,
				Status: tc.Status,
			}
			err = service.StatusTransition(statusReq)
			require.NoError(t, err)
		}
	}

	todoStatus := entity.StatusTodo
	filter := &task.FilterRequest{Status: &todoStatus}
	result, err := service.FindAll(filter, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result.Tasks))

	highPriority := entity.PriorityHigh
	filter = &task.FilterRequest{Priority: &highPriority}
	result, err = service.FindAll(filter, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Tasks))

	assignee := "alice"
	filter = &task.FilterRequest{Assignee: &assignee}
	result, err = service.FindAll(filter, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result.Tasks))
}

func TestTaskIntegration_Pagination(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	for i := 1; i <= 15; i++ {
		priority := entity.PriorityMedium
		dueDate := time.Now().Add(time.Hour * 24)
		req := &task.CreateRequest{
			Summary:     fmt.Sprintf("Pagination Task %d", i),
			Description: "Testing pagination",
			Assignee:    "admin",
			Priority:    &priority,
			DueDate:     &dueDate,
		}
		err := service.Create(req)
		require.NoError(t, err)
	}

	filter := &task.FilterRequest{}

	page1, err := service.FindAll(filter, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(page1.Tasks))
	assert.Equal(t, 15, page1.Meta.Total)
	assert.Equal(t, 1, page1.Meta.Page)

	page2, err := service.FindAll(filter, 2, 10)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(page2.Tasks))
	assert.Equal(t, 15, page2.Meta.Total)
	assert.Equal(t, 2, page2.Meta.Page)
}

func TestTaskIntegration_CacheBehavior(t *testing.T) {
	service, db, cleanup := setupTestService(t)
	defer cleanup()

	priority := entity.PriorityMedium
	dueDate := time.Now().Add(time.Hour * 24)

	createReq := &task.CreateRequest{
		Summary:     "Cache Test Task 1",
		Description: "Testing cache invalidation",
		Assignee:    "admin",
		Priority:    &priority,
		DueDate:     &dueDate,
	}
	err := service.Create(createReq)
	require.NoError(t, err)

	filter := &task.FilterRequest{}
	result1, err := service.FindAll(filter, 1, 10)
	assert.NoError(t, err)
	initialCount := len(result1.Tasks)

	createReq.Summary = "Cache Test Task 2"
	err = service.Create(createReq)
	require.NoError(t, err)

	result2, err := service.FindAll(filter, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, initialCount+1, len(result2.Tasks), "Cache should be invalidated after create")

	var createdTask entity.Task
	err = db.GetDB().Where("summary = ?", "Cache Test Task 2").First(&createdTask).Error
	require.NoError(t, err)

	err = service.Delete(fmt.Sprintf("%d", createdTask.ID))
	assert.NoError(t, err)

	result3, err := service.FindAll(filter, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, initialCount, len(result3.Tasks), "Cache should be invalidated after delete")
}

func TestTaskIntegration_MultipleStatusTransitions(t *testing.T) {
	service, db, cleanup := setupTestService(t)
	defer cleanup()

	priority := entity.PriorityMedium
	dueDate := time.Now().Add(time.Hour * 24)
	createReq := &task.CreateRequest{
		Summary:     "Workflow Task",
		Description: "Testing complete workflow",
		Assignee:    "admin",
		Priority:    &priority,
		DueDate:     &dueDate,
	}

	err := service.Create(createReq)
	assert.NoError(t, err)

	var createdTask entity.Task
	err = db.GetDB().Where("summary = ?", "Workflow Task").First(&createdTask).Error
	require.NoError(t, err)
	assert.Equal(t, entity.StatusTodo, createdTask.Status)

	statusReq := &task.StatusTransitionRequest{
		TaskID: createdTask.ID,
		Status: entity.StatusInProgress,
	}
	err = service.StatusTransition(statusReq)
	assert.NoError(t, err)

	err = db.GetDB().First(&createdTask, createdTask.ID).Error
	require.NoError(t, err)
	assert.Equal(t, entity.StatusInProgress, createdTask.Status)

	statusReq.Status = entity.StatusDone
	err = service.StatusTransition(statusReq)
	assert.NoError(t, err)

	err = db.GetDB().First(&createdTask, createdTask.ID).Error
	require.NoError(t, err)
	assert.Equal(t, entity.StatusDone, createdTask.Status)
}

func TestTaskIntegration_UpdateNonexistentTask(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	updateReq := &task.UpdateRequest{
		Summary:     "Updated Task",
		Description: "This should fail",
		Assignee:    "admin",
		Priority:    entity.PriorityHigh,
		DueDate:     time.Now().Add(time.Hour * 48),
	}

	err := service.Update(updateReq, "99999")
	assert.Error(t, err)
	assert.Equal(t, "task_not_found", err.Error())
}

func TestTaskIntegration_AssignToNonexistentUser(t *testing.T) {
	service, db, cleanup := setupTestService(t)
	defer cleanup()

	priority := entity.PriorityMedium
	dueDate := time.Now().Add(time.Hour * 24)
	createReq := &task.CreateRequest{
		Summary:     "Task for Bad Assign",
		Description: "Testing assign to nonexistent user",
		Assignee:    "admin",
		Priority:    &priority,
		DueDate:     &dueDate,
	}

	err := service.Create(createReq)
	assert.NoError(t, err)

	var createdTask entity.Task
	err = db.GetDB().Where("summary = ?", "Task for Bad Assign").First(&createdTask).Error
	require.NoError(t, err)

	assignReq := &task.AssignRequest{
		TaskID:   createdTask.ID,
		Assignee: "nonexistent_user",
	}

	err = service.Assign(assignReq)
	assert.Error(t, err)
	assert.Equal(t, "can't find assignee user", err.Error())
}
