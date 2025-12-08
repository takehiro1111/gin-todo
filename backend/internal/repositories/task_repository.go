package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	FindByID(ctx context.Context, id uint) (*models.Task, error)
	FindByUserID(ctx context.Context, userID uint) ([]models.Task, error)
	FindAll(ctx context.Context) ([]models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	UpdateStatus(ctx context.Context, id uint, statusID int64) error
	Delete(ctx context.Context, id uint) error
}

type taskRepositoryImpl struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepositoryImpl{db: db}
}

func (r *taskRepositoryImpl) Create(ctx context.Context, task *models.Task) error {
	err := gorm.G[models.Task](r.db).Create(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %v", err)
	}

	return nil
}

func (r *taskRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.Task, error) {
	task, err := gorm.G[models.Task](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found by id")
		}
		return nil, fmt.Errorf("failed to find task by id: %v", err)
	}

	return &task, nil
}

func (r *taskRepositoryImpl) FindByUserID(ctx context.Context, userID uint) ([]models.Task, error) {
	tasks, err := gorm.G[models.Task](r.db).Where("user_id = ?", userID).Find(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found by userID")
		}
		return nil, fmt.Errorf("failed to find task by userID: %v", err)
	}

	return tasks, nil
}

func (r *taskRepositoryImpl) FindAll(ctx context.Context) ([]models.Task, error) {
	tasks, err := gorm.G[models.Task](r.db).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %v", err)
	}

	return tasks, nil
}

func (r *taskRepositoryImpl) Update(ctx context.Context, task *models.Task) error {
	_, err := gorm.G[models.Task](r.db).Where("id = ?", task.ID).Updates(ctx, *task)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("task not found by id")
		}
		return fmt.Errorf("failed to find task by id: %v", err)
	}

	return nil
}

func (r *taskRepositoryImpl) UpdateStatus(ctx context.Context, id uint, statusID int64) error {
	_, err := gorm.G[models.Task](r.db).Where("id = ?", id).Update(ctx, "status_id", statusID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("status not found by id")
		}
		return fmt.Errorf("failed to find status by id: %v", err)
	}

	return nil
}

func (r *taskRepositoryImpl) Delete(ctx context.Context, id uint) error {
	_, err := gorm.G[models.Task](r.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("task not found by id")
		}
		return fmt.Errorf("failed to find task by id: %v", err)
	}

	return nil
}
