package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"task-api/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockTx is a simple mock transaction for testing
type mockTx struct {
	committed   bool
	rolledBack  bool
}

func (m *mockTx) Commit() error {
	m.committed = true
	return nil
}

func (m *mockTx) Rollback() error {
	m.rolledBack = true
	return nil
}

func TestTaskUsecase_CreateTask(t *testing.T) {
	tests := []struct {
		name           string
		idempotencyKey string
		task           *domain.Task
		userID         string
		mockRepo       *MockTaskRepository
		expectedErr    error
		expectedTask   *domain.Task
	}{
		{
			name:           "success create task without idempotency",
			idempotencyKey: "",
			task:           &domain.Task{Title: "Test", Description: "Desc", Status: "pending"},
			userID:         "user1",
			mockRepo: &MockTaskRepository{
				CreateFunc: func(ctx context.Context, task *domain.Task) error {
					task.ID = "new-id"
					return nil
				},
			},
			expectedErr:  nil,
			expectedTask: &domain.Task{ID: "new-id", Title: "Test", Description: "Desc", Status: "pending", UserID: "user1"},
		},
		{
			name:           "idempotency key already exists",
			idempotencyKey: "key123",
			task:           &domain.Task{Title: "Test"},
			userID:         "user1",
			mockRepo: &MockTaskRepository{
				GetIdempotencyFunc: func(ctx context.Context, key string) ([]byte, int, error) {
					task := domain.Task{ID: "cached-id", Title: "Cached"}
					data, _ := json.Marshal(task)
					return data, 200, nil
				},
			},
			expectedErr:  nil,
			expectedTask: &domain.Task{ID: "cached-id", Title: "Cached"},
		},
		{
			name:           "create error without idempotency",
			idempotencyKey: "",
			task:           &domain.Task{Title: "Test"},
			userID:         "user1",
			mockRepo: &MockTaskRepository{
				CreateFunc: func(ctx context.Context, task *domain.Task) error {
					return errors.New("db error")
				},
			},
			expectedErr:  errors.New("db error"),
			expectedTask: nil,
		},
		{
			name:           "success with idempotency key (fallback path - tx nil)",
			idempotencyKey: "key123",
			task:           &domain.Task{Title: "Test", Description: "Desc", Status: "pending"},
			userID:         "user1",
			mockRepo: &MockTaskRepository{
				GetIdempotencyFunc: func(ctx context.Context, key string) ([]byte, int, error) {
					return nil, 0, nil
				},
				BeginTxFunc: func(ctx context.Context) (*sql.Tx, error) {
					return nil, nil // simulate tx unavailable (e.g., in testing)
				},
				CreateFunc: func(ctx context.Context, task *domain.Task) error {
					task.ID = "new-id"
					return nil
				},
				SaveIdempotencyFunc: func(ctx context.Context, key string, response []byte, code int) error {
					return nil
				},
			},
			expectedErr:  nil,
			expectedTask: &domain.Task{ID: "new-id", Title: "Test", Description: "Desc", Status: "pending", UserID: "user1"},
		},
		{
			name:           "create error with idempotency (fallback path)",
			idempotencyKey: "key123",
			task:           &domain.Task{Title: "Test"},
			userID:         "user1",
			mockRepo: &MockTaskRepository{
				GetIdempotencyFunc: func(ctx context.Context, key string) ([]byte, int, error) {
					return nil, 0, nil
				},
				BeginTxFunc: func(ctx context.Context) (*sql.Tx, error) {
					return nil, nil
				},
				CreateFunc: func(ctx context.Context, task *domain.Task) error {
					return errors.New("db error")
				},
			},
			expectedErr:  errors.New("db error"),
			expectedTask: nil,
		},
		{
			name:           "save idempotency error is ignored (fallback path) - task still created",
			idempotencyKey: "key123",
			task:           &domain.Task{Title: "Test", Description: "Desc", Status: "pending"},
			userID:         "user1",
			mockRepo: &MockTaskRepository{
				GetIdempotencyFunc: func(ctx context.Context, key string) ([]byte, int, error) {
					return nil, 0, nil
				},
				BeginTxFunc: func(ctx context.Context) (*sql.Tx, error) {
					return nil, nil
				},
				CreateFunc: func(ctx context.Context, task *domain.Task) error {
					task.ID = "new-id"
					return nil
				},
				SaveIdempotencyFunc: func(ctx context.Context, key string, response []byte, code int) error {
					return errors.New("idempotency save error") // error is ignored, task still created
				},
			},
			expectedErr:  nil,
			expectedTask: &domain.Task{ID: "new-id", Title: "Test", Description: "Desc", Status: "pending", UserID: "user1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := NewTaskUsecase(tt.mockRepo)
			task, err := usecase.CreateTask(context.Background(), tt.idempotencyKey, tt.task, tt.userID)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask.ID, task.ID)
				assert.Equal(t, tt.expectedTask.Title, task.Title)
				if tt.expectedTask.UserID != "" {
					assert.Equal(t, tt.expectedTask.UserID, task.UserID)
				}
			}
		})
	}
}

func TestTaskUsecase_GetTask(t *testing.T) {
	tests := []struct {
		name         string
		taskID       string
		userID       string
		mockRepo     *MockTaskRepository
		expectedErr  error
		expectedTask *domain.Task
	}{
		{
			name:   "success",
			taskID: "task1",
			userID: "user1",
			mockRepo: &MockTaskRepository{
				GetByIDForUserFunc: func(ctx context.Context, id, userID string) (*domain.Task, error) {
					return &domain.Task{ID: id, UserID: userID, Title: "Test"}, nil
				},
			},
			expectedErr:  nil,
			expectedTask: &domain.Task{ID: "task1", UserID: "user1", Title: "Test"},
		},
		{
			name:   "not found",
			taskID: "task1",
			userID: "user1",
			mockRepo: &MockTaskRepository{
				GetByIDForUserFunc: func(ctx context.Context, id, userID string) (*domain.Task, error) {
					return nil, errors.New("task not found or not owned by user")
				},
			},
			expectedErr:  errors.New("task not found or not owned by user"),
			expectedTask: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := NewTaskUsecase(tt.mockRepo)
			task, err := usecase.GetTask(context.Background(), tt.taskID, tt.userID)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask, task)
			}
		})
	}
}

func TestTaskUsecase_UpdateTask(t *testing.T) {
	tests := []struct {
		name        string
		task        *domain.Task
		userID      string
		mockRepo    *MockTaskRepository
		expectedErr error
	}{
		{
			name:   "success update",
			task:   &domain.Task{ID: "task1", Title: "Updated", Status: "completed"},
			userID: "user1",
			mockRepo: &MockTaskRepository{
				GetByIDForUserFunc: func(ctx context.Context, id, userID string) (*domain.Task, error) {
					return &domain.Task{ID: id, UserID: userID, Title: "Old", Status: "pending"}, nil
				},
				UpdateFunc: func(ctx context.Context, task *domain.Task) error {
					return nil
				},
			},
			expectedErr: nil,
		},
		{
			name:   "invalid status transition (completed -> pending)",
			task:   &domain.Task{ID: "task1", Status: "pending"},
			userID: "user1",
			mockRepo: &MockTaskRepository{
				GetByIDForUserFunc: func(ctx context.Context, id, userID string) (*domain.Task, error) {
					return &domain.Task{ID: id, UserID: userID, Status: "completed"}, nil
				},
			},
			expectedErr: errors.New("invalid status transition"),
		},
		{
			name:   "no field to update",
			task:   &domain.Task{ID: "task1"},
			userID: "user1",
			mockRepo: &MockTaskRepository{
				GetByIDForUserFunc: func(ctx context.Context, id, userID string) (*domain.Task, error) {
					return &domain.Task{ID: id, UserID: userID, Status: "pending"}, nil
				},
			},
			expectedErr: errors.New("at least one field must be updated"),
		},
		{
			name:   "not found",
			task:   &domain.Task{ID: "task1", Title: "Updated"},
			userID: "user1",
			mockRepo: &MockTaskRepository{
				GetByIDForUserFunc: func(ctx context.Context, id, userID string) (*domain.Task, error) {
					return nil, errors.New("task not found or not owned by user")
				},
			},
			expectedErr: errors.New("task not found or not owned by user"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := NewTaskUsecase(tt.mockRepo)
			err := usecase.UpdateTask(context.Background(), tt.task, tt.userID)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskUsecase_DeleteTask(t *testing.T) {
	tests := []struct {
		name        string
		taskID      string
		userID      string
		mockRepo    *MockTaskRepository
		expectedErr error
	}{
		{
			name:   "success",
			taskID: "task1",
			userID: "user1",
			mockRepo: &MockTaskRepository{
				DeleteFunc: func(ctx context.Context, id, userID string) error {
					return nil
				},
			},
			expectedErr: nil,
		},
		{
			name:   "not found",
			taskID: "task1",
			userID: "user1",
			mockRepo: &MockTaskRepository{
				DeleteFunc: func(ctx context.Context, id, userID string) error {
					return errors.New("task not found or not owned by user")
				},
			},
			expectedErr: errors.New("task not found or not owned by user"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := NewTaskUsecase(tt.mockRepo)
			err := usecase.DeleteTask(context.Background(), tt.taskID, tt.userID)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskUsecase_AssignTask(t *testing.T) {
	tests := []struct {
		name        string
		taskID      string
		assignerID  string
		assigneeID  string
		mockRepo    *MockTaskRepository
		expectedErr error
	}{
		{
			name:       "success",
			taskID:     "task1",
			assignerID: "user1",
			assigneeID: "user2",
			mockRepo: &MockTaskRepository{
				AssignFunc: func(ctx context.Context, taskID, assignerID, assigneeID string) error {
					return nil
				},
			},
			expectedErr: nil,
		},
		{
			name:        "self assign",
			taskID:      "task1",
			assignerID:  "user1",
			assigneeID:  "user1",
			mockRepo:    &MockTaskRepository{},
			expectedErr: errors.New("cannot assign task to yourself"),
		},
		{
			name:       "repository error",
			taskID:     "task1",
			assignerID: "user1",
			assigneeID: "user2",
			mockRepo: &MockTaskRepository{
				AssignFunc: func(ctx context.Context, taskID, assignerID, assigneeID string) error {
					return errors.New("task not found")
				},
			},
			expectedErr: errors.New("task not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := NewTaskUsecase(tt.mockRepo)
			err := usecase.AssignTask(context.Background(), tt.taskID, tt.assignerID, tt.assigneeID)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskUsecase_ListTasks(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		filter      domain.TaskFilter
		mockRepo    *MockTaskRepository
		expectedLen int
		expectedErr error
	}{
		{
			name:   "success with filter",
			userID: "user1",
			filter: domain.TaskFilter{Status: "pending", Page: 1, Limit: 10},
			mockRepo: &MockTaskRepository{
				ListFunc: func(ctx context.Context, userID string, filter domain.TaskFilter) ([]domain.Task, int, error) {
					return []domain.Task{{ID: "1", Title: "Task1"}}, 1, nil
				},
			},
			expectedLen: 1,
			expectedErr: nil,
		},
		{
			name:   "empty list",
			userID: "user1",
			filter: domain.TaskFilter{},
			mockRepo: &MockTaskRepository{
				ListFunc: func(ctx context.Context, userID string, filter domain.TaskFilter) ([]domain.Task, int, error) {
					return []domain.Task{}, 0, nil
				},
			},
			expectedLen: 0,
			expectedErr: nil,
		},
		{
			name:   "error from repo",
			userID: "user1",
			filter: domain.TaskFilter{},
			mockRepo: &MockTaskRepository{
				ListFunc: func(ctx context.Context, userID string, filter domain.TaskFilter) ([]domain.Task, int, error) {
					return nil, 0, errors.New("db error")
				},
			},
			expectedLen: 0,
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := NewTaskUsecase(tt.mockRepo)
			tasks, total, err := usecase.ListTasks(context.Background(), tt.userID, tt.filter)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, tasks, tt.expectedLen)
				assert.Equal(t, tt.expectedLen, total)
			}
		})
	}
}
