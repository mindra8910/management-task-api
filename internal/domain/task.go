package domain

import (
	"context"
	"database/sql"
)

type Task struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// TxManager interface untuk transaction management
type TxManager interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CreateTx(ctx context.Context, tx *sql.Tx, task *Task) error
	SaveIdempotencyTx(ctx context.Context, tx *sql.Tx, key string, response []byte, code int) error
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id string) (*Task, error)
	GetByIDForUser(ctx context.Context, id, userID string) (*Task, error)
	List(ctx context.Context, userID string, filter TaskFilter) ([]Task, int, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id, userID string) error
	Assign(ctx context.Context, taskID, assignerID, assigneeID string) error

	// Idempotency
	SaveIdempotency(ctx context.Context, key string, response []byte, code int) error
	GetIdempotency(ctx context.Context, key string) ([]byte, int, error)

	// Transaction support
	TxManager
}

type TaskFilter struct {
	Status string `form:"status"`
	Title  string `form:"title"`
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
}

type TaskUsecase interface {
	CreateTask(ctx context.Context, idempotencyKey string, task *Task, userID string) (*Task, error)
	GetTask(ctx context.Context, id, userID string) (*Task, error)
	ListTasks(ctx context.Context, userID string, filter TaskFilter) ([]Task, int, error)
	UpdateTask(ctx context.Context, task *Task, userID string) error
	DeleteTask(ctx context.Context, id, userID string) error
	AssignTask(ctx context.Context, taskID, assignerID, assigneeID string) error
}

// Request/Response DTOs
type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"omitempty,oneof=pending in_progress completed"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"omitempty,oneof=pending in_progress completed"`
}

type AssignTaskRequest struct {
	AssigneeID string `json:"assignee_id" binding:"required"`
}
