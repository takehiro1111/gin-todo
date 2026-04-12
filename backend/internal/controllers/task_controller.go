package controllers

import (
	"context"
	"errors"
	"time"

	"github.com/danielgtaylor/huma/v2"

	appErr "github.com/takehiro1111/gin-todo/backend/internal/errors"
	"github.com/takehiro1111/gin-todo/backend/internal/middleware"
	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"github.com/takehiro1111/gin-todo/backend/internal/services"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

type TaskController interface {
	GetTasks(ctx context.Context, input *GetTasksInput) (*GetTasksOutput, error)
	CreateTask(ctx context.Context, input *CreateTaskInput) (*CreateTaskOutput, error)
	GetTaskByID(ctx context.Context, input *GetTaskByIDInput) (*GetTaskByIDOutput, error)
	UpdateTask(ctx context.Context, input *UpdateTaskInput) (*UpdateTaskOutput, error)
	UpdateTaskStatus(ctx context.Context, input *UpdateTaskStatusInput) (*UpdateTaskStatusOutput, error)
	DeleteTask(ctx context.Context, input *DeleteTaskInput) (*DeleteTaskOutput, error)
}

type TaskControllerImpl struct {
	taskService  services.TaskService
	timeProvider *utils.RealTimeProvider
}

func NewTaskControllerImpl(taskService services.TaskService, timeProvider *utils.RealTimeProvider) *TaskControllerImpl {
	return &TaskControllerImpl{
		taskService:  taskService,
		timeProvider: timeProvider,
	}
}

// --- Request DTOs ---

type CreateRequest struct {
	Title       string     `json:"title" required:"true" maxLength:"100" doc:"タスクタイトル"`
	Description string     `json:"description" maxLength:"300" doc:"タスク説明"`
	StatusID    int64      `json:"status_id" doc:"ステータスID"`
	Priority    string     `json:"priority" doc:"優先度"`
	DueDate     *time.Time `json:"due_date,omitempty" doc:"期限"`
}

type UpdateRequest struct {
	Title       string     `json:"title" required:"true" maxLength:"100" doc:"タスクタイトル"`
	Description string     `json:"description" maxLength:"100" doc:"タスク説明"`
	StatusID    int64      `json:"status_id" doc:"ステータスID"`
	Priority    string     `json:"priority" doc:"優先度"`
	DueDate     *time.Time `json:"due_date,omitempty" doc:"期限"`
}

type UpdateTaskStatusRequest struct {
	StatusID int64 `json:"status_id" required:"true" doc:"新しいステータスID"`
}

// --- Input / Output ---

type GetTasksInput struct{}
type GetTasksOutput struct {
	Body SuccessBody[[]models.Task]
}

type CreateTaskInput struct {
	Body CreateRequest
}
type CreateTaskOutput struct {
	Body SuccessBody[*models.Task]
}

type GetTaskByIDInput struct {
	ID uint `path:"id" doc:"タスクID"`
}
type GetTaskByIDOutput struct {
	Body SuccessBody[*models.Task]
}

type UpdateTaskInput struct {
	ID   uint `path:"id" doc:"タスクID"`
	Body UpdateRequest
}
type UpdateTaskOutput struct {
	Body SuccessBody[*models.Task]
}

type UpdateTaskStatusInput struct {
	ID   uint `path:"id" doc:"タスクID"`
	Body UpdateTaskStatusRequest
}
type UpdateTaskStatusOutput struct {
	Body SuccessBody[*struct{}]
}

type DeleteTaskInput struct {
	ID uint `path:"id" doc:"タスクID"`
}
type DeleteTaskOutput struct {
	Body SuccessBody[*struct{}]
}

// --- Handlers ---

func (t *TaskControllerImpl) GetTasks(ctx context.Context, input *GetTasksInput) (*GetTasksOutput, error) {
	userID, ok := ctx.Value(middleware.UserIDKey).(uint)
	if !ok {
		return nil, huma.Error401Unauthorized("invalid user access")
	}

	tasks, err := t.taskService.GetAllTasks(ctx, userID)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &GetTasksOutput{
		Body: NewSuccessBody("success get all tasks", tasks, t.timeProvider),
	}, nil
}

func (t *TaskControllerImpl) CreateTask(ctx context.Context, input *CreateTaskInput) (*CreateTaskOutput, error) {
	userID, ok := ctx.Value(middleware.UserIDKey).(uint)
	if !ok {
		return nil, huma.Error401Unauthorized("invalid user access")
	}

	task := &models.Task{
		UserID:      userID,
		Title:       input.Body.Title,
		Description: input.Body.Description,
		StatusID:    input.Body.StatusID,
		Priority:    input.Body.Priority,
		DueDate:     input.Body.DueDate,
	}

	data, err := t.taskService.CreateTask(ctx, task)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &CreateTaskOutput{
		Body: NewSuccessBody("success create task", data, t.timeProvider),
	}, nil
}

func (t *TaskControllerImpl) GetTaskByID(ctx context.Context, input *GetTaskByIDInput) (*GetTaskByIDOutput, error) {
	userID, ok := ctx.Value(middleware.UserIDKey).(uint)
	if !ok {
		return nil, huma.Error401Unauthorized("invalid user access")
	}

	data, err := t.taskService.GetTaskByID(ctx, input.ID, userID)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			return nil, huma.Error404NotFound("task not found")
		}
		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &GetTaskByIDOutput{
		Body: NewSuccessBody("success get task by id", data, t.timeProvider),
	}, nil
}

func (t *TaskControllerImpl) UpdateTask(ctx context.Context, input *UpdateTaskInput) (*UpdateTaskOutput, error) {
	userID, ok := ctx.Value(middleware.UserIDKey).(uint)
	if !ok {
		return nil, huma.Error401Unauthorized("invalid user access")
	}

	task := &models.Task{
		BaseModel:   models.BaseModel{ID: input.ID},
		UserID:      userID,
		Title:       input.Body.Title,
		Description: input.Body.Description,
		StatusID:    input.Body.StatusID,
		Priority:    input.Body.Priority,
		DueDate:     input.Body.DueDate,
	}

	data, err := t.taskService.UpdateTask(ctx, task)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			return nil, huma.Error404NotFound("task not found")
		}
		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &UpdateTaskOutput{
		Body: NewSuccessBody("success update task", data, t.timeProvider),
	}, nil
}

func (t *TaskControllerImpl) UpdateTaskStatus(ctx context.Context, input *UpdateTaskStatusInput) (*UpdateTaskStatusOutput, error) {
	userID, ok := ctx.Value(middleware.UserIDKey).(uint)
	if !ok {
		return nil, huma.Error401Unauthorized("invalid user access")
	}

	err := t.taskService.UpdateTaskStatus(ctx, input.ID, userID, input.Body.StatusID)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			return nil, huma.Error404NotFound("task not found")
		}
		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &UpdateTaskStatusOutput{
		Body: NewSuccessBody[*struct{}]("success update task status id", nil, t.timeProvider),
	}, nil
}

func (t *TaskControllerImpl) DeleteTask(ctx context.Context, input *DeleteTaskInput) (*DeleteTaskOutput, error) {
	userID, ok := ctx.Value(middleware.UserIDKey).(uint)
	if !ok {
		return nil, huma.Error401Unauthorized("invalid user access")
	}

	err := t.taskService.DeleteTask(ctx, input.ID, userID)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			return nil, huma.Error404NotFound("task not found")
		}
		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &DeleteTaskOutput{
		Body: NewSuccessBody[*struct{}]("success delete task", nil, t.timeProvider),
	}, nil
}
