package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"task-api/internal/domain"

	"github.com/google/uuid"
)

type taskUsecase struct {
	taskRepo domain.TaskRepository
}

func NewTaskUsecase(repo domain.TaskRepository) domain.TaskUsecase {
	return &taskUsecase{taskRepo: repo}
}

func (u *taskUsecase) CreateTask(ctx context.Context, idempotencyKey string, task *domain.Task, userID string) (*domain.Task, error) {
	if idempotencyKey != "" {
		cachedResp, _, err := u.taskRepo.GetIdempotency(ctx, idempotencyKey)
		if err == nil && cachedResp != nil {
			var task domain.Task
			if err := json.Unmarshal(cachedResp, &task); err == nil {
				return &task, nil
			}
		}
	}

	task.ID = uuid.New().String()
	task.UserID = userID
	if task.Status == "" {
		task.Status = "pending"
	}

	// Jika ada idempotency key, gunakan transaction untuk atomic operation
	if idempotencyKey != "" {
		tx, err := u.taskRepo.BeginTx(ctx)
		if err != nil {
			return nil, err
		}

		// Jika tx is nil (e.g., in testing), fallback ke non-transactional
		if tx != nil {
			defer tx.Rollback()

			// Create task dalam transaction
			if err := u.taskRepo.CreateTx(ctx, tx, task); err != nil {
				return nil, err
			}

			// Save idempotency dalam transaction yang sama
			respData, _ := json.Marshal(task)
			if err := u.taskRepo.SaveIdempotencyTx(ctx, tx, idempotencyKey, respData, 201); err != nil {
				return nil, err
			}

			// Commit transaction
			if err := tx.Commit(); err != nil {
				return nil, err
			}

			return task, nil
		}

		// Fallback: non-transactional create + save idempotency
		if err := u.taskRepo.Create(ctx, task); err != nil {
			return nil, err
		}
		respData, _ := json.Marshal(task)
		_ = u.taskRepo.SaveIdempotency(ctx, idempotencyKey, respData, 201)
		return task, nil
	}

	// Jika tidak ada idempotency key, create tanpa transaction
	err := u.taskRepo.Create(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (u *taskUsecase) GetTask(ctx context.Context, id, userID string) (*domain.Task, error) {
	return u.taskRepo.GetByIDForUser(ctx, id, userID)
}

func (u *taskUsecase) ListTasks(ctx context.Context, userID string, filter domain.TaskFilter) ([]domain.Task, int, error) {
	return u.taskRepo.List(ctx, userID, filter)
}

func (u *taskUsecase) UpdateTask(ctx context.Context, task *domain.Task, userID string) error {
	existing, err := u.taskRepo.GetByIDForUser(ctx, task.ID, userID)
	if err != nil {
		return err
	}

	if task.Title == "" && task.Description == "" && task.Status == "" {
		return errors.New("at least one field must be updated")
	}

	if existing.Status == "completed" && task.Status == "pending" {
		return errors.New("invalid status transition")
	}

	task.UserID = existing.UserID
	return u.taskRepo.Update(ctx, task)
}

func (u *taskUsecase) DeleteTask(ctx context.Context, id, userID string) error {
	return u.taskRepo.Delete(ctx, id, userID)
}

func (u *taskUsecase) AssignTask(ctx context.Context, taskID, assignerID, assigneeID string) error {
	if assignerID == assigneeID {
		return errors.New("cannot assign task to yourself")
	}
	return u.taskRepo.Assign(ctx, taskID, assignerID, assigneeID)
}
