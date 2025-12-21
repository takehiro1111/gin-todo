package services

import (
	"github.com/go-playground/validator/v10"

	"fmt"
	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"github.com/takehiro1111/gin-todo/backend/internal/repositories"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
	"regexp"
	"time"
)

type AuthServiceImpl struct {
	userRepo        repositories.UserRepository
	passwordManager utils.PasswordManager
	jwtProvider     utils.JWTProvider
	accessTokenTTL  time.Duration
}

type RegisterProvider struct {
	Email    string `validate:"emailCustom"`
	Password string `validate:"passwordCustom"`
}

func emailValidate(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	if email == "" {
		return false
	}
	if match, _ := regexp.MatchString(".+@.+\\..+", email); !match {
		return false
	}
	return true
}

func passwordValidate(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if password == "" {
		return false
	}
	if match, _ := regexp.MatchString("^[a-zA-Z0-9.?/-!]{8,24}$", password); !match {
		return false
	}
	return true
}

func (s *AuthServiceImpl) Register(name, email, password string) (string, error) {
	validate := validator.New()
	validate.RegisterValidation("emailCustom", emailValidate)
	validate.RegisterValidation("passwordCustom", passwordValidate)

	err := validate.Struct(RegisterProvider{Email: email, Password: password})
	if err != nil {
		return "", fmt.Errorf("failed validation: %w", err)
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", err
	}

	if user != nil {
		return "", fmt.Errorf("This email address is already registered")
	}

	hashedPassword, err := s.passwordManager.HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed HashPassword: %w", err)
	}

	newUser := models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hashedPassword,
		RoleName:     models.RoleUser,
	}

	err = s.userRepo.Create(&newUser)
	if err != nil {
		return "", err
	}

	now := time.Now()
	exp := now.Add(s.accessTokenTTL)

	token, err := s.jwtProvider.GenerateAccessToken(
		newUser.Name,
		models.RoleUser,
		exp, // 有効期限
		now, // 発行時刻
		now, // 有効開始時刻
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}
