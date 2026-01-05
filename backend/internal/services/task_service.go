package services

import (
	"context"

	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"github.com/takehiro1111/gin-todo/backend/internal/repositories"
)

type TaskService interface {
	CreateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	GetTaskByID(ctx context.Context, id, userID uint) (*models.Task, error)
	GetTasksByUserID(ctx context.Context, userID uint) ([]models.Task, error)
	GetAllTasks(ctx context.Context, userID uint) ([]models.Task, error)
	UpdateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	UpdateTaskStatus(ctx context.Context, id, userID uint, statusID int64) error
	DeleteTask(ctx context.Context, id, userID uint) error
}

type TaskServiceImpl struct {
	taskRepository repositories.TaskRepository
}

func NewTaskService(repo repositories.TaskRepository) TaskService {
	return &TaskServiceImpl{taskRepository: repo}
}

func (s *TaskServiceImpl) CreateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	// バリデーションの実装はあと工程で考える
	// カスタムバリデーションを差し込みたい

	err := s.taskRepository.Create(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskServiceImpl) GetTaskByID(ctx context.Context, id, userID uint) (*models.Task, error) {
	// バリデーションの実装はあと工程で考える
	// カスタムバリデーションを差し込みたい

	task, err := s.taskRepository.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskServiceImpl) GetTasksByUserID(ctx context.Context, userID uint) ([]models.Task, error) {
	// バリデーションの実装はあと工程で考える
	// カスタムバリデーションを差し込みたい

	tasks, err := s.taskRepository.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *TaskServiceImpl) GetAllTasks(ctx context.Context, userID uint) ([]models.Task, error) {
	// バリデーションの実装はあと工程で考える
	// カスタムバリデーションを差し込みたい

	tasks, err := s.taskRepository.FindAll(ctx, userID)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *TaskServiceImpl) UpdateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	// バリデーションの実装はあと工程で考える
	// カスタムバリデーションを差し込みたい

	err := s.taskRepository.Update(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskServiceImpl) UpdateTaskStatus(ctx context.Context, id, userID uint, statusID int64) error {
	// バリデーションの実装はあと工程で考える
	// カスタムバリデーションを差し込みたい

	err := s.taskRepository.UpdateStatus(ctx, id, userID, statusID)
	if err != nil {
		return err
	}

	return nil
}

func (s *TaskServiceImpl) DeleteTask(ctx context.Context, id, userID uint) error {
	// バリデーションの実装はあと工程で考える
	// カスタムバリデーションを差し込みたい

	err := s.taskRepository.Delete(ctx, id, userID)
	if err != nil {
		return err
	}

	return nil
}
