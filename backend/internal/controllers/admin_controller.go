package controllers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	"github.com/takehiro1111/gin-todo/backend/internal/middleware"
	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"github.com/takehiro1111/gin-todo/backend/internal/services"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

type AdminController interface {
	GetAllUsers(ctx context.Context, input *GetAllUsersInput) (*GetAllUsersOutput, error)
	GetAllUsersWithTasks(ctx context.Context, input *GetAllUsersWithTasksInput) (*GetAllUsersWithTasksOutput, error)
}

type AdminControllerImpl struct {
	adminService services.AdminService
	timeProvider *utils.RealTimeProvider
}

func NewAdminControllerImpl(adminService services.AdminService, timeProvider *utils.RealTimeProvider) *AdminControllerImpl {
	return &AdminControllerImpl{
		adminService: adminService,
		timeProvider: timeProvider,
	}
}

// --- Input / Output ---

type GetAllUsersInput struct{}

type GetAllUsersOutput struct {
	Body SuccessBody[[]models.User]
}

type GetAllUsersWithTasksInput struct{}

type GetAllUsersWithTasksOutput struct {
	Body SuccessBody[[]models.User]
}

// --- Handlers ---

func (a *AdminControllerImpl) GetAllUsers(ctx context.Context, input *GetAllUsersInput) (*GetAllUsersOutput, error) {
	_, ok := ctx.Value(middleware.UserIDKey).(uint)
	if !ok {
		return nil, huma.Error401Unauthorized("invalid user access")
	}

	users, err := a.adminService.GetAllUsers(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &GetAllUsersOutput{
		Body: NewSuccessBody[[]models.User]("success get all users", users, a.timeProvider),
	}, nil
}

func (a *AdminControllerImpl) GetAllUsersWithTasks(ctx context.Context, input *GetAllUsersWithTasksInput) (*GetAllUsersWithTasksOutput, error) {
	_, ok := ctx.Value(middleware.UserIDKey).(uint)
	if !ok {
		return nil, huma.Error401Unauthorized("invalid user access")
	}

	users, err := a.adminService.GetAllUsersWithTasks(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &GetAllUsersWithTasksOutput{
		Body: NewSuccessBody[[]models.User]("success get all users with tasks", users, a.timeProvider),
	}, nil
}
