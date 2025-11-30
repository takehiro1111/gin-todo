package repositories

import (
	"github.com/takehiro1111/gin-todo/backend/internal/models"
)

type UserRoleRepository interface {
	FindAll() ([]*models.UserRole, error)
	FindByName(name string) (*models.UserRole, error)
}
