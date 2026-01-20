package repositories

import (
	"context"
	"fmt"

	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"gorm.io/gorm"
)

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, resetToken *models.PasswordResetToken) error
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
