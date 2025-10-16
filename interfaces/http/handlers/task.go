package handlers

import (
	"task_mng/pkg/response"
	"task_mng/services/task"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskService *task.Service
}

func NewTaskHandler(taskService *task.Service) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// Create godoc
// @Summary Create a new task
// @Description Create a new task in the system
// @Tags Tasks
// @Accept json
// @Produce json
// @Param request body task.CreateRequest true "Task creation data"
// @Success 201 {object} response.Response "created"
// @Failure 400 {object} response.Response "Bad request"
// @Security BearerAuth
// @Router /tasks [post]
func (h *TaskHandler) Create(c *gin.Context) {
	req, err := response.Parse[task.CreateRequest](c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.taskService.Create(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, nil)
}

// Update godoc
// @Summary Update a task
// @Description Update an existing task by ID
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body task.UpdateRequest true "Task update data"
// @Success 200 {object} response.Response "Task updated successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Security BearerAuth
// @Router /tasks/{id} [put]
func (h *TaskHandler) Update(c *gin.Context) {
	id := c.Param("id")

	req, err := response.Parse[task.UpdateRequest](c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.taskService.Update(req, id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "Task updated successfully", nil, nil)
}

// FindByID godoc
// @Summary Get a task by ID
// @Description Get detailed information about a specific task
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} response.Response{data=aggregate.TaskResponse} "Task fetched successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Security BearerAuth
// @Router /tasks/{id} [get]
func (h *TaskHandler) FindByID(c *gin.Context) {
	id := c.Param("id")

	task, err := h.taskService.FindByID(id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, "Task fetched successfully", task, nil)
}

// FindAll godoc
// @Summary Get all tasks
// @Description Get a list of all tasks with optional filters and pagination
// @Tags Tasks
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param assignee query string false "Filter by assignee username"
// @Param status query string false "Filter by status (ToDo, InProgress, Done)" Enums(ToDo, InProgress, Done)
// @Param priority query string false "Filter by priority (lowest, low, medium, high, highest)" Enums(lowest, low, medium, high, highest)
// @Success 200 {object} response.Response{data=aggregate.TaskListResponse} "Tasks fetched successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Security BearerAuth
// @Router /tasks [get]
func (h *TaskHandler) FindAll(c *gin.Context) {
	pag := response.NewPagination(c)

	req, err := response.ParseQuery[task.FilterRequest](c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.taskService.FindAll(req, pag.Page, pag.Limit)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "Tasks fetched successfully", result.Tasks, result.Meta)
}

// Delete godoc
// @Summary Delete a task
// @Description Delete a task by ID
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} response.Response "Task deleted successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Security BearerAuth
// @Router /tasks/{id} [delete]
func (h *TaskHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.taskService.Delete(id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "Task deleted successfully", nil, nil)
}

// Assign godoc
// @Summary Assign a task to a user
// @Description Assign or reassign a task to a specific user
// @Tags Tasks
// @Accept json
// @Produce json
// @Param request body task.AssignRequest true "Task assignment data"
// @Success 200 {object} response.Response "Task assigned successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Security BearerAuth
// @Router /tasks/assign [put]
func (h *TaskHandler) Assign(c *gin.Context) {
	req, err := response.Parse[task.AssignRequest](c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.taskService.Assign(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "Task assigned successfully", nil, nil)
}

// Transition godoc
// @Summary Transition task status
// @Description Change the status of a task (e.g., from ToDo to InProgress)
// @Tags Tasks
// @Accept json
// @Produce json
// @Param request body task.StatusTransitionRequest true "Task status transition data"
// @Success 200 {object} response.Response "Task status transitioned successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Security BearerAuth
// @Router /tasks/transition [put]
func (h *TaskHandler) Transition(c *gin.Context) {
	req, err := response.Parse[task.StatusTransitionRequest](c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.taskService.StatusTransition(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "Task status transitioned successfully", nil, nil)
}
