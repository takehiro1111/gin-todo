package repositories

import (
	"github.com/takehiro1111/gin-todo/backend/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindAll() ([]*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
}
