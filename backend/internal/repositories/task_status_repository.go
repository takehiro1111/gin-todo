package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"gorm.io/gorm"
)

type TaskStatusRepository interface {
	FindAll() ([]models.TaskStatus, error)
	FindByName(name string) (*models.TaskStatus, error)
}

type taskStatusRepositoryImpl struct {
	db *gorm.DB
}

func NewTaskStatusRepository(db *gorm.DB) TaskStatusRepository {
	return &taskStatusRepositoryImpl{db: db}
}

func (r *taskStatusRepositoryImpl) FindAll() ([]models.TaskStatus, error) {
	ctx := context.Background()

	taskStatus, err := gorm.G[models.TaskStatus](r.db).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find task statuses: %v", err)
	}

	return taskStatus, nil
}

func (r *taskStatusRepositoryImpl) FindByName(name string) (*models.TaskStatus, error) {
	ctx := context.Background()

	taskStatus, err := gorm.G[models.TaskStatus](r.db).Where("name = ?", name).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task status not found by name")
		}
		return nil, fmt.Errorf("failed to find task status by name: %v", err)
	}

	return &taskStatus, nil
}
