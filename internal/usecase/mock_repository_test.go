package usecase

import (
	"context"
	"database/sql"
	"task-api/internal/domain"
)

// MockTaskRepository adalah mock untuk domain.TaskRepository
type MockTaskRepository struct {
	CreateFunc           func(ctx context.Context, task *domain.Task) error
	CreateTxFunc         func(ctx context.Context, tx *sql.Tx, task *domain.Task) error
	BeginTxFunc          func(ctx context.Context) (*sql.Tx, error)
	GetByIDFunc          func(ctx context.Context, id string) (*domain.Task, error)
	GetByIDForUserFunc   func(ctx context.Context, id, userID string) (*domain.Task, error)
	ListFunc             func(ctx context.Context, userID string, filter domain.TaskFilter) ([]domain.Task, int, error)
	UpdateFunc           func(ctx context.Context, task *domain.Task) error
	DeleteFunc           func(ctx context.Context, id, userID string) error
	AssignFunc           func(ctx context.Context, taskID, assignerID, assigneeID string) error
	SaveIdempotencyFunc  func(ctx context.Context, key string, response []byte, code int) error
	SaveIdempotencyTxFunc func(ctx context.Context, tx *sql.Tx, key string, response []byte, code int) error
	GetIdempotencyFunc   func(ctx context.Context, key string) ([]byte, int, error)
}

func (m *MockTaskRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx)
	}
	return nil, nil
}

func (m *MockTaskRepository) CreateTx(ctx context.Context, tx *sql.Tx, task *domain.Task) error {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, task)
	}
	return nil
}

func (m *MockTaskRepository) SaveIdempotencyTx(ctx context.Context, tx *sql.Tx, key string, response []byte, code int) error {
	if m.SaveIdempotencyTxFunc != nil {
		return m.SaveIdempotencyTxFunc(ctx, tx, key, response, code)
	}
	return nil
}

func (m *MockTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, task)
	}
	return nil
}
func (m *MockTaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *MockTaskRepository) GetByIDForUser(ctx context.Context, id, userID string) (*domain.Task, error) {
	if m.GetByIDForUserFunc != nil {
		return m.GetByIDForUserFunc(ctx, id, userID)
	}
	return nil, nil
}
func (m *MockTaskRepository) List(ctx context.Context, userID string, filter domain.TaskFilter) ([]domain.Task, int, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, userID, filter)
	}
	return nil, 0, nil
}
func (m *MockTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, task)
	}
	return nil
}
func (m *MockTaskRepository) Delete(ctx context.Context, id, userID string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id, userID)
	}
	return nil
}
func (m *MockTaskRepository) Assign(ctx context.Context, taskID, assignerID, assigneeID string) error {
	if m.AssignFunc != nil {
		return m.AssignFunc(ctx, taskID, assignerID, assigneeID)
	}
	return nil
}
func (m *MockTaskRepository) SaveIdempotency(ctx context.Context, key string, response []byte, code int) error {
	if m.SaveIdempotencyFunc != nil {
		return m.SaveIdempotencyFunc(ctx, key, response, code)
	}
	return nil
}
func (m *MockTaskRepository) GetIdempotency(ctx context.Context, key string) ([]byte, int, error) {
	if m.GetIdempotencyFunc != nil {
		return m.GetIdempotencyFunc(ctx, key)
	}
	return nil, 0, nil
}

// MockUserRepository adalah mock untuk domain.UserRepository
type MockUserRepository struct {
	CreateFunc     func(ctx context.Context, user *domain.User) error
	GetByEmailFunc func(ctx context.Context, email string) (*domain.User, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil
}
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	return nil, nil
}
