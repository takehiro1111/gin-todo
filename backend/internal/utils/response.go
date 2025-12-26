package utils

import (
	"github.com/gin-gonic/gin"
	"time"
)

// ref: https://pkg.go.dev/github.com/Palguna1121/go-starter/template/template/libs/responses
type TimeProvider interface {
	Now() time.Time
}

type RealTimeProvider struct{}

type SuccessResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type ErrorResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Error     string    `json:"error,omitempty"`
	Code      string    `json:"code,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

type ValidationErrorResponse struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message"`
	Errors    []ValidationError `json:"errors"`
	Timestamp time.Time         `json:"timestamp"`
}

func (*RealTimeProvider) Now() time.Time {
	return time.Now()
}

func ResponseSuccess(c *gin.Context, statusCode int, message string, data interface{}, timeProvider TimeProvider) {
	c.JSON(statusCode, SuccessResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: timeProvider.Now(),
	})
}

func ResponseError(c *gin.Context, statusCode int, message string, err string, timeProvider TimeProvider) {
	c.JSON(statusCode, ErrorResponse{
		Success:   false,
		Message:   message,
		Error:     err,
		Timestamp: timeProvider.Now(),
	})
}

func ResponseValidationError(c *gin.Context, statusCode int, message string, errors []ValidationError, timeProvider TimeProvider) {
	c.JSON(statusCode, ValidationErrorResponse{
		Success:   false,
		Message:   message,
		Errors:    errors,
		Timestamp: timeProvider.Now(),
	})
}
