package services

import (
	"fmt"
	"strconv"
	"time"

	appErr "github.com/takehiro1111/gin-todo/backend/internal/errors"
	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"github.com/takehiro1111/gin-todo/backend/internal/repositories"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

type AuthService interface {
	Register(name, email, password string) (string, error)
	Login(email, password string) (string, error)
	RefreshToken(refreshToken string) (*TokenPair, error)
	GetMe(userID string) (*models.User, error)
	ChangePassword(userID, oldPassword, newPassword string) error
}

type AuthServiceImpl struct {
	userRepo        repositories.UserRepository
	passwordManager utils.PasswordManager
	jwtProvider     utils.JWTProvider
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
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

func NewAuthService(userRepo repositories.UserRepository, passwordManager utils.PasswordManager, jwtProvider utils.JWTProvider, accessTokenTTL, refreshTokenTTL time.Duration) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:        userRepo,
		passwordManager: passwordManager,
		jwtProvider:     jwtProvider,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}
func (s *AuthServiceImpl) Register(name, email, password string) (string, error) {
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
		RoleID:       1,
	}

	err = s.userRepo.Create(&newUser)
	if err != nil {
		return "", err
	}

	now := time.Now()
	exp := now.Add(s.accessTokenTTL)

	token, err := s.jwtProvider.GenerateAccessToken(
		newUser.Name,
		models.RoleWriter,
		newUser.ID,
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
		models.RoleWriter,
		user.ID,
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
		models.RoleWriter,
		user.ID,
		accessTokenExp,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.jwtProvider.GenerateRefreshToken(
		user.Name,
		models.RoleWriter,
		user.ID,
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

func (s *AuthServiceImpl) GetMe(userID string) (*models.User, error) {
	toUintUserID, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to cast userID  : %w", err)
	}

	user, err := s.userRepo.FindByID(uint(toUintUserID))
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, appErr.ErrUserNotFound
	}

	return user, nil
}

func (s *AuthServiceImpl) ChangePassword(userID, oldPassword, newPassword string) error {
	toUintUserID, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to cast userID  : %w", err)
	}

	user, err := s.userRepo.FindByID(uint(toUintUserID))
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	err = s.passwordManager.CheckPassword(user.PasswordHash, oldPassword)
	if err != nil {
		return fmt.Errorf("failed to check password: %w", err)
	}

	hashedNewPassword, err := s.passwordManager.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = s.userRepo.UpdatePassword(uint(toUintUserID), hashedNewPassword)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
