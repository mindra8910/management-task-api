package http

import (
	"task-api/internal/domain"
	"task-api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUsecase domain.UserUsecase
}

func NewUserHandler(r *gin.Engine, us domain.UserUsecase) {
	handler := &UserHandler{userUsecase: us}
	r.POST("/register", handler.Register)
	r.POST("/login", handler.Login)
}

func (h *UserHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	user, err := h.userUsecase.Register(c.Request.Context(), &req)
	if err != nil {
		// Mapping error dari usecase ke response yang sesuai
		if err.Error() == "email already registered" {
			response.Conflict(c, response.CodeDuplicateEmail, "Email already registered")
			return
		}
		// Default internal server error
		response.InternalServerError(c, response.CodeInternalServer, "Failed to register user")
		return
	}

	response.Created(c, user, "User registered successfully")
}

func (h *UserHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := h.userUsecase.Login(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "invalid email or password" {
			response.Unauthorized(c, response.CodeInvalidCredential, "Invalid email or password")
			return
		}
		response.InternalServerError(c, response.CodeInternalServer, "Failed to login")
		return
	}

	response.Success(c, res, "Login successful")
}
