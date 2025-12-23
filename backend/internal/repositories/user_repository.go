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
	FindByName(name string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindAll() ([]models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
}

type userRepositoryImpl struct {
	db *gorm.DB
}

// UserRepositoryのinterfaceを型として返すことで実装の詳細を隠せる
// テストのMockへの切り替えやすさも考慮している。
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) Create(user *models.User) error {
	ctx := context.Background()

	err := gorm.G[models.User](r.db).Create(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
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

func (r *userRepositoryImpl) FindByName(name string) (*models.User, error) {
	ctx := context.Background()

	user, err := gorm.G[models.User](r.db).Where("name = ?", name).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found by Name")
		}
		return nil, fmt.Errorf("failed to find user by Name: %v", err)
	}

	return &user, nil
}

func (r *userRepositoryImpl) FindByEmail(email string) (*models.User, error) {
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

func (r *userRepositoryImpl) FindAll() ([]models.User, error) {
	ctx := context.Background()

	users, err := gorm.G[models.User](r.db).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find users: %v", err)
	}

	return users, nil
}

func (r *userRepositoryImpl) Update(user *models.User) error {
	ctx := context.Background()

	_, err := gorm.G[models.User](r.db).Where("id = ?", user.ID).Updates(ctx, *user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found by id")
		}
		return fmt.Errorf("failed to find user by id: %v", err)
	}

	return nil
}

func (r *userRepositoryImpl) Delete(id uint) error {
	ctx := context.Background()

	_, err := gorm.G[models.User](r.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found by id")
		}
		return fmt.Errorf("failed to find user by id: %v", err)
	}

	return nil
}
