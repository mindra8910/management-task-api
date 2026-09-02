package usecase

import (
	"context"
	"errors"
	"task-api/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUserUsecase_Register(t *testing.T) {
	tests := []struct {
		name        string
		req         *domain.RegisterRequest
		mockRepo    *MockUserRepository
		expectedErr error
	}{
		{
			name: "success register",
			req: &domain.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockRepo: &MockUserRepository{
				GetByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
					return nil, errors.New("user not found")
				},
				CreateFunc: func(ctx context.Context, user *domain.User) error {
					return nil
				},
			},
			expectedErr: nil,
		},
		{
			name: "email already registered",
			req: &domain.RegisterRequest{
				Email: "existing@example.com",
			},
			mockRepo: &MockUserRepository{
				GetByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
					return &domain.User{Email: email}, nil
				},
			},
			expectedErr: errors.New("email already registered"),
		},
		{
			name: "create user error",
			req: &domain.RegisterRequest{
				Name:     "Jane",
				Email:    "jane@example.com",
				Password: "pass",
			},
			mockRepo: &MockUserRepository{
				GetByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
					return nil, errors.New("user not found")
				},
				CreateFunc: func(ctx context.Context, user *domain.User) error {
					return errors.New("db insert error")
				},
			},
			expectedErr: errors.New("db insert error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := NewUserUsecase(tt.mockRepo, "secret")
			user, err := usecase.Register(context.Background(), tt.req)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, user.ID)
				assert.Equal(t, tt.req.Name, user.Name)
				assert.Equal(t, tt.req.Email, user.Email)
				// Password should be hashed
				err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(tt.req.Password))
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserUsecase_Login(t *testing.T) {
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)

	tests := []struct {
		name        string
		req         *domain.LoginRequest
		mockRepo    *MockUserRepository
		expectedErr error
		expectToken bool
	}{
		{
			name: "success login",
			req: &domain.LoginRequest{
				Email:    "test@example.com",
				Password: "correctpass",
			},
			mockRepo: &MockUserRepository{
				GetByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
					return &domain.User{
						ID:       "user1",
						Email:    email,
						Password: string(hashedPass),
					}, nil
				},
			},
			expectedErr: nil,
			expectToken: true,
		},
		{
			name: "invalid password",
			req: &domain.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpass",
			},
			mockRepo: &MockUserRepository{
				GetByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
					return &domain.User{
						ID:       "user1",
						Email:    email,
						Password: string(hashedPass),
					}, nil
				},
			},
			expectedErr: errors.New("invalid email or password"),
			expectToken: false,
		},
		{
			name: "user not found",
			req: &domain.LoginRequest{
				Email:    "notfound@example.com",
				Password: "pass",
			},
			mockRepo: &MockUserRepository{
				GetByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
					return nil, errors.New("user not found")
				},
			},
			expectedErr: errors.New("invalid email or password"),
			expectToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := NewUserUsecase(tt.mockRepo, "secret")
			res, err := usecase.Login(context.Background(), tt.req)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.NotEmpty(t, res.Token)
				assert.Equal(t, tt.req.Email, res.User.Email)
			}
		})
	}
}
