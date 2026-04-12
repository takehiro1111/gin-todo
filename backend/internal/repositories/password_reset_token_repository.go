package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	appErr "github.com/takehiro1111/gin-todo/backend/internal/errors"
	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"gorm.io/gorm"
)

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, resetToken *models.PasswordResetToken) error
	FindByToken(ctx context.Context, newResetToken string) (*models.PasswordResetToken, error)
	UpdateUseAtByID(ctx context.Context, id uint, now time.Time) error
}

type passwordResetTokenRepositoryImpl struct {
	db *gorm.DB
}

func NewPasswordResetTokenRepository(db *gorm.DB) PasswordResetTokenRepository {
	return &passwordResetTokenRepositoryImpl{db: db}
}

func (r *passwordResetTokenRepositoryImpl) Create(ctx context.Context, resetToken *models.PasswordResetToken) error {
	err := gorm.G[models.PasswordResetToken](r.db).Create(ctx, resetToken)
	if err != nil {
		return fmt.Errorf("failed to create password reset token: %v", err)
	}

	return nil
}

func (r *passwordResetTokenRepositoryImpl) FindByToken(ctx context.Context, newResetToken string) (*models.PasswordResetToken, error) {
	token, err := gorm.G[models.PasswordResetToken](r.db).Where("reset_token = ?", newResetToken).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErr.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find password reset token: %v", err)
	}

	return &token, nil
}

func (r *passwordResetTokenRepositoryImpl) UpdateUseAtByID(ctx context.Context, id uint, now time.Time) error {
	_, err := gorm.G[models.PasswordResetToken](r.db).Where("id = ?", id).Update(ctx, "UsedAt", now)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErr.ErrNotFound
		}
		return fmt.Errorf("failed to find reset token by id: %v", err)
	}

	return nil
}
