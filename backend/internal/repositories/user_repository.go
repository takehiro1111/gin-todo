package repositories

import (
	"context"
	"errors"
	"fmt"

	appErr "github.com/takehiro1111/gin-todo/backend/internal/errors"
	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uint) (*models.User, error)
	FindByName(ctx context.Context, ame string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindAll(ctx context.Context) ([]models.User, error)
	FindAllWithTasks(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, id uint, newHashedPassword string) error
	Delete(ctx context.Context, id uint) error
}

type userRepositoryImpl struct {
	db *gorm.DB
}

// UserRepositoryのinterfaceを型として返すことで実装の詳細を隠せる
// テストのMockへの切り替えやすさも考慮している。
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *models.User) error {
	err := gorm.G[models.User](r.db).Create(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.User, error) {
	user, err := gorm.G[models.User](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErr.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by ID: %v", err)
	}

	return &user, nil
}

func (r *userRepositoryImpl) FindByName(ctx context.Context, name string) (*models.User, error) {
	user, err := gorm.G[models.User](r.db).Where("name = ?", name).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErr.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by Name: %v", err)
	}

	return &user, nil
}

func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := gorm.G[models.User](r.db).Where("email = ?", email).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Registerで使用しているためerror時もnilで返す
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by Email: %v", err)
	}

	return &user, nil
}

func (r *userRepositoryImpl) FindAll(ctx context.Context) ([]models.User, error) {
	users, err := gorm.G[models.User](r.db).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find users: %v", err)
	}

	return users, nil
}

func (r *userRepositoryImpl) FindAllWithTasks(ctx context.Context) ([]models.User, error) {
	users, err := gorm.G[models.User](r.db).Preload("Tasks", nil).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find users: %v", err)
	}

	return users, nil
}

func (r *userRepositoryImpl) Update(ctx context.Context, user *models.User) error {
	_, err := gorm.G[models.User](r.db).Where("id = ?", user.ID).Updates(ctx, *user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErr.ErrUserNotFound
		}
		return fmt.Errorf("failed to find user by id: %v", err)
	}

	return nil
}

func (r *userRepositoryImpl) UpdatePassword(ctx context.Context, id uint, hashedNewPassword string) error {
	_, err := gorm.G[models.User](r.db).Where("id = ?", id).Update(ctx, "PasswordHash", hashedNewPassword)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErr.ErrUserNotFound
		}
		return fmt.Errorf("failed to find user by id: %v", err)
	}

	return nil
}

func (r *userRepositoryImpl) Delete(ctx context.Context, id uint) error {
	_, err := gorm.G[models.User](r.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErr.ErrUserNotFound
		}
		return fmt.Errorf("failed to find user by id: %v", err)
	}

	return nil
}
