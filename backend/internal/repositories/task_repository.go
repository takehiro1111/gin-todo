package repositories

import (
	"github.com/takehiro1111/gin-todo/backend/internal/models"
)

type TaskRepository interface {
	Create(task *models.Task) error
	FindByID(id uint) (*models.Task, error)
	FindByEmail(email string) (*models.User, error)
	FindAll() ([]*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
}
