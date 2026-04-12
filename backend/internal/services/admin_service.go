package services

import (
	"context"

	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"github.com/takehiro1111/gin-todo/backend/internal/repositories"
)

type AdminService interface {
	GetAllUsers(ctx context.Context) ([]models.User, error)
	GetAllUsersWithTasks(ctx context.Context) ([]models.User, error)
}

type AdminServiceImpl struct {
	userRepo repositories.UserRepository
}

func NewAdminService(userRepo repositories.UserRepository) AdminService {
	return &AdminServiceImpl{
		userRepo: userRepo,
	}
}

func (s *AdminServiceImpl) GetAllUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *AdminServiceImpl) GetAllUsersWithTasks(ctx context.Context) ([]models.User, error) {
	users, err := s.userRepo.FindAllWithTasks(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}
