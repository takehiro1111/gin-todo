package repositories

import (
	// 実装時にコメントイン予定
	// "context"
	// "errors"
	// "fmt"

	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(task *models.Task) error
	FindByUserID(id uint) (*models.Task, error)
	FindAll() ([]*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
}

type taskRepositoryImpl struct {
	db *gorm.DB
}
