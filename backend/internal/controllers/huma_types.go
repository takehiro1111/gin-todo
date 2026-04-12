package controllers

import (
	"time"

	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

// SuccessBody は既存の utils.SuccessResponse と同じ JSON 構造を維持する。
// 型パラメータ T により、OpenAPI スキーマに data フィールドの型が反映される。
type SuccessBody[T any] struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Data      T      `json:"data,omitempty"`
	Timestamp string `json:"timestamp"`
}

// ErrorBody は既存の utils.ErrorResponse と同じ JSON 構造を維持する。
type ErrorBody struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"`
	Timestamp string `json:"timestamp"`
}

func NewSuccessBody[T any](message string, data T, tp utils.TimeProvider) SuccessBody[T] {
	return SuccessBody[T]{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: utils.ToJSTLocalTime(tp.Now()).Format(time.RFC3339),
	}
}
