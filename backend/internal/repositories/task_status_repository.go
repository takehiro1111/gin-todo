package repositories

import (
	"github.com/takehiro1111/gin-todo/backend/internal/models"
)

type TaskStatusRepository interface {
	FindAll() ([]*models.TaskStatus, error)
	FindByName(name string) (*models.TaskStatus, error)
}
