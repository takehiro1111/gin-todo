package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"gorm.io/gorm"
)

type UserRoleRepository interface {
	FindAll() ([]models.UserRole, error)
	FindByRoleName(name string) (*models.UserRole, error)
}

type userRoleRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRoleRepository(db *gorm.DB) UserRoleRepository {
	return &userRoleRepositoryImpl{db: db}
}

func (r *userRoleRepositoryImpl) FindAll() ([]models.UserRole, error) {
	ctx := context.Background()

	userRoles, err := gorm.G[models.UserRole](r.db).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user roles: %v", err)
	}

	return userRoles, nil
}

func (r *userRoleRepositoryImpl) FindByRoleName(name string) (*models.UserRole, error) {
	ctx := context.Background()

	userRole, err := gorm.G[models.UserRole](r.db).Where("name = ?", name).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user role not found by name")
		}
		return nil, fmt.Errorf("failed to find user role by name: %v", err)
	}

	return &userRole, nil
}
