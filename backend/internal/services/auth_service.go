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

type AuthService interface {
	Register(name, email, password string) (string, error)
	Login(email, password string) (string, error)
	RefreshToken(refreshToken string) (*TokenPair, error)
}

type AuthServiceImpl struct {
	userRepo        repositories.UserRepository
	passwordManager utils.PasswordManager
	jwtProvider     utils.JWTProvider
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type AuthenticateProvider struct {
	Email    string `validate:"emailCustom"`
	Password string `validate:"passwordCustom"`
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func NewTokenPair(accessToken, refreshToken string) *TokenPair {
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
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

	err := validate.Struct(AuthenticateProvider{Email: email, Password: password})
	if err != nil {
		return "", fmt.Errorf("failed to validation: %w", err)
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
		return "", fmt.Errorf("failed to hash password: %w", err)
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

func (s *AuthServiceImpl) Login(email, password string) (string, error) {
	validate := validator.New()
	validate.RegisterValidation("emailCustom", emailValidate)
	validate.RegisterValidation("passwordCustom", passwordValidate)

	err := validate.Struct(AuthenticateProvider{Email: email, Password: password})
	if err != nil {
		return "", fmt.Errorf("failed validation: %w", err)
	}

	// メールアドレス, パスワードの付け合わせ
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", fmt.Errorf("invalid credentials")
	}

	err = s.passwordManager.CheckPassword(user.PasswordHash, password)
	if err != nil {
		return "", fmt.Errorf("failed to check password: %w", err)
	}

	// JWT生成
	now := time.Now()
	exp := now.Add(s.accessTokenTTL)

	token, err := s.jwtProvider.GenerateAccessToken(
		user.Name,
		user.RoleName,
		exp, // 有効期限
		now, // 発行時刻
		now, // 有効開始時刻
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

func (s *AuthServiceImpl) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := s.jwtProvider.VerifyJWT(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify refresh token: %w", err)
	}

	user, err := s.userRepo.FindByName(claims.UserName)
	if err != nil {
		return nil, fmt.Errorf("invalid claim from verify refresh token: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	now := time.Now()
	accessTokenExp := now.Add(s.accessTokenTTL)
	refreshTokenExp := now.Add(s.refreshTokenTTL)

	newAccessToken, err := s.jwtProvider.GenerateAccessToken(
		user.Name,
		user.RoleName,
		accessTokenExp,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.jwtProvider.GenerateRefreshToken(
		user.Name,
		user.RoleName,
		refreshTokenExp,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	tokenPair := NewTokenPair(newAccessToken, newRefreshToken)
	return tokenPair, nil
}
