package services

import (
	"context"
	"fmt"

	"github.com/takehiro1111/gin-todo/backend/infrastructure/aws"
)

type EmailService interface {
	SendPasswordResetEmail(ctx context.Context, to, token string) error
}

type EmailConfig struct {
	From                 string
	PasswordResetBaseURL string
}

func NewEmailConfig(from, passwordResetBaseURL string) *EmailConfig {
	return &EmailConfig{
		From:                 from,
		PasswordResetBaseURL: passwordResetBaseURL,
	}
}

type emailServiceImpl struct {
	mail     aws.Mail
	emailCfg EmailConfig
}

func NewEmailService(mail aws.Mail, emailCfg *EmailConfig) EmailService {
	return &emailServiceImpl{
		mail:     mail,
		emailCfg: *emailCfg,
	}
}

func (e *emailServiceImpl) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	resetURL := fmt.Sprintf("%s/reset?token=%s", e.emailCfg.PasswordResetBaseURL, token)

	title := "パスワードリセットのご案内"
	content := fmt.Sprintf("以下のリンクからパスワードを再設定してください:\n%s", resetURL)

	return e.mail.SendEmail(ctx, title, content, to, e.emailCfg.From)
}
