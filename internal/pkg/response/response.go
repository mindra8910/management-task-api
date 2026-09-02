package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Response adalah struktur standar untuk semua respons API.
type Response struct {
	Status    string      `json:"status"`
	Code      string      `json:"code,omitempty"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// Success mengembalikan respons 200 OK dengan data.
func Success(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, Response{
		Status:    "success",
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// Created mengembalikan respons 201 Created.
func Created(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusCreated, Response{
		Status:    "success",
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// ValidationError mengembalikan respons 400 dengan detail field yang error
func ValidationError(c *gin.Context, err error) {
	var errorsMap map[string]string

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errorsMap = make(map[string]string)
		for _, ve := range validationErrors {
			field := ve.Field()
			tag := ve.Tag()
			param := ve.Param()
			message := getValidationMessage(field, tag, param)
			errorsMap[field] = message
		}
	} else {
		// Fallback jika bukan validation error
		errorsMap = map[string]string{"error": err.Error()}
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, Response{
		Status:    "error",
		Code:      "VALIDATION_ERROR",
		Message:   "Validation failed",
		Data:      errorsMap, // detail errors per field
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// helper untuk pesan validasi yang lebih ramah
func getValidationMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + param + " characters"
	case "max":
		return field + " must be at most " + param + " characters"
	default:
		return field + " failed validation on " + tag
	}
}

// Error mengembalikan respons error dengan status HTTP tertentu dan kode error.
func Error(c *gin.Context, statusCode int, code string, message string) {
	c.AbortWithStatusJSON(statusCode, Response{
		Status:    "error",
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// BadRequest = 400
func BadRequest(c *gin.Context, code, message string) {
	Error(c, http.StatusBadRequest, code, message)
}

// Unauthorized = 401
func Unauthorized(c *gin.Context, code, message string) {
	Error(c, http.StatusUnauthorized, code, message)
}

// Forbidden = 403
func Forbidden(c *gin.Context, code, message string) {
	Error(c, http.StatusForbidden, code, message)
}

// NotFound = 404
func NotFound(c *gin.Context, code, message string) {
	Error(c, http.StatusNotFound, code, message)
}

// Conflict = 409
func Conflict(c *gin.Context, code, message string) {
	Error(c, http.StatusConflict, code, message)
}

// InternalServerError = 500
func InternalServerError(c *gin.Context, code, message string) {
	Error(c, http.StatusInternalServerError, code, message)
}
