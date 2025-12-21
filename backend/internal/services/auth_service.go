package services

import (
	"github.com/go-playground/validator/v10"

	"fmt"
	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"github.com/takehiro1111/gin-todo/backend/internal/repositories"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
	"regexp"
)

type AuthServiceImpl struct {
	userRepo        repositories.UserRepository
	passwordManager utils.PasswordManager
	jwtProvider     utils.JWTProvider
}

type RegisterProvider struct {
	Email    string `validate:"emailCustom"`
	Password string `validate:"Password"`
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
	emailValidator := validator.New()
	err := emailValidator.RegisterValidation("emailCustom", emailValidate)
	if err != nil {
		return "", fmt.Errorf("failed register validation for email: %w", err)
	}

	err = emailValidator.Struct(RegisterProvider{Email: email})
	if err != nil {
		return "", fmt.Errorf("failed validation email: %w", err)
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", err
	}

	if user != nil {
		return "", fmt.Errorf("This email address is already registered")
	}

	passwordValidator := validator.New()
	err = passwordValidator.RegisterValidation("passwordCustom", passwordValidate)
	if err != nil {
		return "", fmt.Errorf("failed register validation for password: %w", err)
	}

	err = emailValidator.Struct(RegisterProvider{Password: password})
	if err != nil {
		return "", fmt.Errorf("failed validation password: %w", err)
	}

	hashedPassword, err := s.passwordManager.HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed HashPassword: %w", err)
	}

	err = s.passwordManager.CheckPassword(user.PasswordHash, hashedPassword)
	if err != nil {
		return "", fmt.Errorf("failed CheckPassword: %w", err)
	}

	newUser := models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hashedPassword,
		RoleID:       models.RoleUser,
	}

	err = s.userRepo.Create(&newUser)
	if err != nil {
		return "", err
	}

	// 登録後にログイン状態にするためJWTトークンの生成
	// 引数の渡し方をどうするか。。
	token, err := s.jwtProvider.GenerateAccessToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}
