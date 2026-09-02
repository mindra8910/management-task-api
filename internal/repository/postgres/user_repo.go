package postgres

import (
	"context"
	"database/sql"
	"errors"

	"task-api/internal/domain"
)

type postgresUserRepo struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) domain.UserRepository {
	return &postgresUserRepo{db}
}

func (m *postgresUserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, name, email, password) VALUES ($1, $2, $3, $4)`
	_, err := m.db.ExecContext(ctx, query, user.ID, user.Name, user.Email, user.Password)
	return err
}

func (m *postgresUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, name, email, password, created_at FROM users WHERE email = $1`
	
	user := &domain.User{}
	err := m.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	
	return user, nil
}
