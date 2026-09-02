package http

import (
	"task-api/internal/domain"
	"task-api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskUsecase domain.TaskUsecase
}

func NewTaskHandler(r *gin.Engine, tu domain.TaskUsecase, authMiddleware gin.HandlerFunc) {
	handler := &TaskHandler{taskUsecase: tu}

	// Protected routes
	taskGroup := r.Group("/tasks")
	taskGroup.Use(authMiddleware)
	{
		taskGroup.POST("/", handler.CreateTask)
		taskGroup.GET("/", handler.ListTasks)
		taskGroup.GET("/:id", handler.GetTask)
		taskGroup.PUT("/:id", handler.UpdateTask)
		taskGroup.DELETE("/:id", handler.DeleteTask)
		taskGroup.POST("/:id/assign", handler.AssignTask)
	}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, response.CodeInvalidCredential, "User not authenticated")
		return
	}

	var req domain.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.CodeValidation, "Invalid request payload: "+err.Error())
		return
	}

	// Idempotency-Key header
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		response.BadRequest(c, response.CodeValidation, "Idempotency-Key header required")
		return
	}

	task := &domain.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	}

	created, err := h.taskUsecase.CreateTask(c.Request.Context(), idempotencyKey, task, userID)
	if err != nil {
		// Handle duplicate idempotency key (MySQL unique constraint)
		if err.Error() == "duplicate key" {
			response.Conflict(c, response.CodeIdempotencyFailed, "Duplicate idempotency key")
			return
		}
		response.InternalServerError(c, response.CodeInternalServer, "Failed to create task")
		return
	}

	response.Created(c, created, "Task created successfully")
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	task, err := h.taskUsecase.GetTask(c.Request.Context(), id, userID)
	if err != nil {
		if err.Error() == "task not found or not owned by user" {
			response.NotFound(c, response.CodeNotFound, "Task not found")
			return
		}
		response.InternalServerError(c, response.CodeInternalServer, "Failed to get task")
		return
	}

	response.Success(c, task, "Task retrieved successfully")
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	userID := c.GetString("user_id")

	var filter domain.TaskFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, response.CodeValidation, "Invalid query parameters")
		return
	}

	tasks, total, err := h.taskUsecase.ListTasks(c.Request.Context(), userID, filter)
	if err != nil {
		response.InternalServerError(c, response.CodeInternalServer, "Failed to list tasks")
		return
	}

	response.Success(c, gin.H{
		"tasks": tasks,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	}, "Tasks retrieved successfully")
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var req domain.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.CodeValidation, "Invalid request payload: "+err.Error())
		return
	}

	task := &domain.Task{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	}

	err := h.taskUsecase.UpdateTask(c.Request.Context(), task, userID)
	if err != nil {
		if err.Error() == "task not found or not owned by user" {
			response.NotFound(c, response.CodeNotFound, "Task not found")
			return
		}
		response.InternalServerError(c, response.CodeInternalServer, "Failed to update task")
		return
	}

	// Fetch updated task to return
	updated, _ := h.taskUsecase.GetTask(c.Request.Context(), id, userID)
	response.Success(c, updated, "Task updated successfully")
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	err := h.taskUsecase.DeleteTask(c.Request.Context(), id, userID)
	if err != nil {
		if err.Error() == "task not found or not owned by user" {
			response.NotFound(c, response.CodeNotFound, "Task not found")
			return
		}
		response.InternalServerError(c, response.CodeInternalServer, "Failed to delete task")
		return
	}

	response.Success(c, nil, "Task deleted successfully")
}

func (h *TaskHandler) AssignTask(c *gin.Context) {
	userID := c.GetString("user_id")
	taskID := c.Param("id")

	var req domain.AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.CodeValidation, "Invalid request payload: "+err.Error())
		return
	}

	err := h.taskUsecase.AssignTask(c.Request.Context(), taskID, userID, req.AssigneeID)
	if err != nil {
		switch err.Error() {
		case "task not found":
			response.NotFound(c, response.CodeNotFound, "Task not found")
		case "you can only assign your own tasks":
			response.Forbidden(c, response.CodeForbidden, "You can only assign your own tasks")
		case "cannot assign task to yourself":
			response.BadRequest(c, response.CodeValidation, "Cannot assign task to yourself")
		default:
			response.InternalServerError(c, response.CodeInternalServer, "Failed to assign task")
		}
		return
	}

	response.Success(c, nil, "Task assigned successfully")
}
