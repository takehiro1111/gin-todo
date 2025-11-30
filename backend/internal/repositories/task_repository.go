package repositories

import (
	"context"
	"fmt"

	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(task *models.Task) error
	FindByID(id uint) (*models.Task, error)
	FindByUserID(userID uint) ([]models.Task, error)
	FindAll() ([]models.Task, error)
	Update(task *models.Task) error
	UpdateStatus(id uint, statusID int64) error
	Delete(id uint) error
}

type taskRepositoryImpl struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepositoryImpl{db: db}
}

func (r *taskRepositoryImpl) Create(task *models.Task) error {
	ctx := context.Background()

	err := gorm.G[models.Task](r.db).Create(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %v", err)
	}

	return nil
}

func (r *taskRepositoryImpl) FindByID(id uint) (*models.Task, error) {
	ctx := context.Background()

	task, err := gorm.G[models.Task](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find task by id: %v", err)
	}

	return &task, nil
}

func (r *taskRepositoryImpl) FindByUserID(userID uint) ([]models.Task, error) {
	ctx := context.Background()

	tasks, err := gorm.G[models.Task](r.db).Where("user_id = ?", userID).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by user id: %v", err)
	}

	return tasks, nil
}

func (r *taskRepositoryImpl) FindAll() ([]models.Task, error) {
	ctx := context.Background()

	tasks, err := gorm.G[models.Task](r.db).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %v", err)
	}

	return tasks, nil
}

func (r *taskRepositoryImpl) Update(task *models.Task) error {
	ctx := context.Background()

	_, err := gorm.G[models.Task](r.db).Where("id = ?", task.ID).Updates(ctx, *task)
	if err != nil {
		return fmt.Errorf("failed to update task by id: %v", err)
	}

	return nil
}

func (r *taskRepositoryImpl) UpdateStatus(id uint, statusID int64) error {
	ctx := context.Background()

	_, err := gorm.G[models.Task](r.db).Where("id = ?", id).Update(ctx, "status_id", statusID)
	if err != nil {
		return fmt.Errorf("failed to update task status by task id: %v", err)
	}

	return nil
}

func (r *taskRepositoryImpl) Delete(id uint) error {
	ctx := context.Background()

	_, err := gorm.G[models.Task](r.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete task by id: %v", err)
	}

	return nil
}
