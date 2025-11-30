package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindAll() ([]*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
}

type userRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *userRepositoryImpl {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) FindByID(id uint) (*models.User, error) {
	ctx := context.Background()

	user, err := gorm.G[models.User](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found by ID")
		}
		return nil, fmt.Errorf("failed to find user by ID: %v", err)
	}

	return &user, nil
}

func (r *userRepositoryImpl) FindByEMail(email string) (*models.User, error) {
	ctx := context.Background()

	user, err := gorm.G[models.User](r.db).Where("email = ?", email).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found by Email")
		}
		return nil, fmt.Errorf("failed to find user by Email: %v", err)
	}

	return &user, nil
}
