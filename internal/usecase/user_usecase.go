package usecase

import (
	"context"
	"errors"
	"time"

	"task-api/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepo  domain.UserRepository
	jwtSecret []byte
}

func NewUserUsecase(repo domain.UserRepository, secret string) domain.UserUsecase {
	return &userUsecase{
		userRepo:  repo,
		jwtSecret: []byte(secret),
	}
}

func (u *userUsecase) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error) {
	// Check if user exists
	existingUser, _ := u.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create new user
	user := &domain.User{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	err = u.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	// Get user by email
	user, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(u.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &domain.LoginResponse{
		Token: tokenString,
		User:  user,
	}, nil
}
