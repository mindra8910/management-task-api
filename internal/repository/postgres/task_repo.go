package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"task-api/internal/domain"
)

type postgresTaskRepo struct {
	db *sql.DB
}

func NewPostgresTaskRepository(db *sql.DB) domain.TaskRepository {
	return &postgresTaskRepo{db}
}

// BeginTx - mulai transaction
func (r *postgresTaskRepo) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// CreateTx - insert task dalam transaction
func (r *postgresTaskRepo) CreateTx(ctx context.Context, tx *sql.Tx, task *domain.Task) error {
	query := `INSERT INTO tasks (id, user_id, title, description, status, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`
	_, err := tx.ExecContext(ctx, query, task.ID, task.UserID, task.Title, task.Description, task.Status)
	return err
}

// SaveIdempotencyTx - save idempotency key dalam transaction (24 hour window)
func (r *postgresTaskRepo) SaveIdempotencyTx(ctx context.Context, tx *sql.Tx, key string, response []byte, code int) error {
	query := `INSERT INTO idempotency_keys (idempotency_key, response_body, status_code, created_at, expire_at)
	          VALUES ($1, $2, $3, NOW(), NOW() + INTERVAL '24 hours')
	          ON CONFLICT (idempotency_key) DO UPDATE SET
	          response_body = EXCLUDED.response_body,
	          status_code = EXCLUDED.status_code`
	_, err := tx.ExecContext(ctx, query, key, response, code)
	return err
}

// Create task (tanpa transaction - untuk backward compatibility)
func (r *postgresTaskRepo) Create(ctx context.Context, task *domain.Task) error {
	query := `INSERT INTO tasks (id, user_id, title, description, status, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, task.ID, task.UserID, task.Title, task.Description, task.Status)
	return err
}

// GetByID - get task by ID (without user check)
func (r *postgresTaskRepo) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `SELECT id, user_id, title, description, status, created_at, updated_at
	          FROM tasks WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var task domain.Task
	err := row.Scan(&task.ID, &task.UserID, &task.Title, &task.Description,
		&task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("task not found")
		}
		return nil, err
	}
	return &task, nil
}

// GetByIDForUser - get task by ID and ensure it belongs to user (for authorization)
func (r *postgresTaskRepo) GetByIDForUser(ctx context.Context, id, userID string) (*domain.Task, error) {
	query := `SELECT id, user_id, title, description, status, created_at, updated_at
	          FROM tasks WHERE id = $1 AND user_id = $2`
	row := r.db.QueryRowContext(ctx, query, id, userID)

	var task domain.Task
	err := row.Scan(&task.ID, &task.UserID, &task.Title, &task.Description,
		&task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("task not found or not owned by user")
		}
		return nil, err
	}
	return &task, nil
}

// List tasks with pagination and filters
func (r *postgresTaskRepo) List(ctx context.Context, userID string, filter domain.TaskFilter) ([]domain.Task, int, error) {
	// Build WHERE clause
	where := "user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	if filter.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}
	if filter.Title != "" {
		where += fmt.Sprintf(" AND title ILIKE $%d", argIndex)
		args = append(args, "%"+filter.Title+"%")
		argIndex++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM tasks WHERE " + where
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}
	offset := (filter.Page - 1) * filter.Limit

	query := fmt.Sprintf(`SELECT id, user_id, title, description, status, created_at, updated_at
	                      FROM tasks WHERE %s
	                      ORDER BY created_at DESC
	                      LIMIT $%d OFFSET $%d`,
		where, argIndex, argIndex+1)
	args = append(args, filter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := []domain.Task{}
	for rows.Next() {
		var task domain.Task
		err := rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description,
			&task.Status, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, task)
	}
	return tasks, total, nil
}

// Update task
func (r *postgresTaskRepo) Update(ctx context.Context, task *domain.Task) error {
	query := `UPDATE tasks SET title = $1, description = $2, status = $3, updated_at = NOW()
	          WHERE id = $4 AND user_id = $5`
	result, err := r.db.ExecContext(ctx, query, task.Title, task.Description, task.Status, task.ID, task.UserID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("task not found or not owned by user")
	}
	return nil
}

// Delete task
func (r *postgresTaskRepo) Delete(ctx context.Context, id, userID string) error {
	query := `DELETE FROM tasks WHERE id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("task not found or not owned by user")
	}
	return nil
}

// Assign task - with transaction (already implemented)
func (r *postgresTaskRepo) Assign(ctx context.Context, taskID, assignerID, assigneeID string) error {
	// 3. Database Transaction & Integrity
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if task exists and belongs to assigner
	var currentUserID string
	err = tx.QueryRowContext(ctx, "SELECT user_id FROM tasks WHERE id = $1", taskID).Scan(&currentUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("task not found")
		}
		return err
	}
	if currentUserID != assignerID {
		return errors.New("you can only assign your own tasks")
	}

	// Update task user_id
	_, err = tx.ExecContext(ctx, "UPDATE tasks SET user_id = $1, updated_at = NOW() WHERE id = $2", assigneeID, taskID)
	if err != nil {
		return err
	}

	// Insert log
	_, err = tx.ExecContext(ctx, "INSERT INTO task_logs (task_id, action) VALUES ($1, $2)", taskID, "assigned to new user")
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// Idempotency methods
func (r *postgresTaskRepo) SaveIdempotency(ctx context.Context, key string, response []byte, code int) error {
	query := `INSERT INTO idempotency_keys (idempotency_key, response_body, status_code, created_at, expire_at)
	          VALUES ($1, $2, $3, NOW(), NOW() + INTERVAL '24 hours')
	          ON CONFLICT (idempotency_key) DO UPDATE SET
	          response_body = EXCLUDED.response_body,
	          status_code = EXCLUDED.status_code`
	_, err := r.db.ExecContext(ctx, query, key, response, code)
	return err
}

func (r *postgresTaskRepo) GetIdempotency(ctx context.Context, key string) ([]byte, int, error) {
	query := `SELECT response_body, status_code FROM idempotency_keys
	          WHERE idempotency_key = $1 AND expire_at > NOW()`
	var response []byte
	var code int
	err := r.db.QueryRowContext(ctx, query, key).Scan(&response, &code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, nil // not found or expired
		}
		return nil, 0, err
	}
	return response, code, nil
}
