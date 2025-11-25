package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(task *models.Task) error
	FindByID(id uint) (*models.Task, error)
	FindByEmail(email string) (*models.User, error)
	FindAll() ([]*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
}

type taskRepositoryImpl struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepositoryImpl{db: db}
}

func (r *taskRepositoryImpl) FindByID(id uint) (*models.Task, error) {
	ctx := context.Background()

	task, err := gorm.G[models.Task](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found by ID")
		}
		return nil, fmt.Errorf("failed to find task by ID: %v", err)
	}

	return &task, nil
}

func (r *taskRepositoryImpl) FindByEMail(email string) (*models.Task, error) {
	ctx := context.Background()

	task, err := gorm.G[models.Task](r.db).Where("email = ?", email).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found by Email")
		}
		return nil, fmt.Errorf("failed to find task by Email: %v", err)
	}

	return &task, nil
}
