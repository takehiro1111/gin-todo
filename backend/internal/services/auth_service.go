package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	appErr "github.com/takehiro1111/gin-todo/backend/internal/errors"
	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"github.com/takehiro1111/gin-todo/backend/internal/repositories"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

type AuthService interface {
	Register(ctx context.Context, name, email, password string) (string, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	GetMe(ctx context.Context, userID string) (*models.User, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	ForgotPassword(ctx context.Context, email string) (string, error)
	ResetPassword(ctx context.Context, resetToken, newPassword string) error
}

type AuthServiceImpl struct {
	userRepo               repositories.UserRepository
	passwordResetTokenRepo repositories.PasswordResetTokenRepository
	passwordManager        utils.PasswordManager
	jwtProvider            utils.JWTProvider
	uuidGenerator          utils.UUIDGenerator
	accessTokenTTL         time.Duration
	refreshTokenTTL        time.Duration
	timeProvider           utils.TimeProvider
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

func NewAuthService(userRepo repositories.UserRepository, passwordResetTokenRepo repositories.PasswordResetTokenRepository, passwordManager utils.PasswordManager, jwtProvider utils.JWTProvider, accessTokenTTL, refreshTokenTTL time.Duration, uuidGenerator utils.UUIDGenerator, timeProvider utils.TimeProvider) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:               userRepo,
		passwordResetTokenRepo: passwordResetTokenRepo,
		passwordManager:        passwordManager,
		jwtProvider:            jwtProvider,
		accessTokenTTL:         accessTokenTTL,
		refreshTokenTTL:        refreshTokenTTL,
		uuidGenerator:          uuidGenerator,
		timeProvider:           timeProvider,
	}
}
func (s *AuthServiceImpl) Register(ctx context.Context, name, email, password string) (string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
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

	err = s.userRepo.Create(ctx, &newUser)
	if err != nil {
		return "", err
	}

	now := s.timeProvider.Now()
	accessTokenExp := now.Add(s.accessTokenTTL)

	token, err := s.jwtProvider.GenerateAccessToken(
		newUser.Name,
		models.GetRoleNameByID(newUser.RoleID),
		newUser.ID,
		accessTokenExp, // 有効期限
		now,            // 発行時刻
		now,            // 有効開始時刻
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

func (s *AuthServiceImpl) Login(ctx context.Context, email, password string) (string, string, error) {
	// メールアドレス, パスワードの付け合わせ
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}

	if user == nil {
		return "", "", fmt.Errorf("invalid credentials")
	}

	err = s.passwordManager.CheckPassword(user.PasswordHash, password)
	if err != nil {
		return "", "", fmt.Errorf("failed to check password: %w", err)
	}

	// JWT生成
	now := s.timeProvider.Now()
	accessTokenExp := now.Add(s.accessTokenTTL)
	refreshTokenExp := now.Add(s.refreshTokenTTL)

	roleName := models.GetRoleNameByID(user.RoleID)

	accessToken, err := s.jwtProvider.GenerateAccessToken(
		user.Name,
		roleName,
		user.ID,
		accessTokenExp, // 有効期限
		now,            // 発行時刻
		now,            // 有効開始時刻
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := s.jwtProvider.GenerateRefreshToken(
		user.Name,
		roleName,
		user.ID,
		refreshTokenExp,
		now,
		now,
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.jwtProvider.VerifyJWT(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify refresh token: %w", err)
	}

	user, err := s.userRepo.FindByName(ctx, claims.UserName)
	if err != nil {
		return nil, fmt.Errorf("invalid claim from verify refresh token: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	now := s.timeProvider.Now()
	accessTokenExp := now.Add(s.accessTokenTTL)
	refreshTokenExp := now.Add(s.refreshTokenTTL)

	roleName := models.GetRoleNameByID(user.RoleID)

	newAccessToken, err := s.jwtProvider.GenerateAccessToken(
		user.Name,
		roleName,
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
		roleName,
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

func (s *AuthServiceImpl) GetMe(ctx context.Context, userID string) (*models.User, error) {
	toUintUserID, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to cast userID  : %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, uint(toUintUserID))
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, appErr.ErrUserNotFound
	}

	return user, nil
}

func (s *AuthServiceImpl) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	toUintUserID, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to cast userID  : %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, uint(toUintUserID))
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

	err = s.userRepo.UpdatePassword(ctx, uint(toUintUserID), hashedNewPassword)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *AuthServiceImpl) ForgotPassword(ctx context.Context, email string) (string, error) {
	// 一致するユーザーが存在するかemail検証
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("failed to find user by email: %w", err)
	}

	if user == nil {
		return "", fmt.Errorf("user not found")
	}

	now := s.timeProvider.Now()
	resetTokenExp := now.Add(1 * time.Hour)

	// UUIDの生成
	uuid := s.uuidGenerator.UUIDGenerate()
	// DBに保存
	newToken := models.PasswordResetToken{
		ResetToken: uuid,
		UserID:     user.ID,
		ExpiresAt:  resetTokenExp,
	}

	err = s.passwordResetTokenRepo.Create(ctx, &newToken)
	if err != nil {
		return "", err
	}

	return newToken.ResetToken, nil
}

func (s *AuthServiceImpl) ResetPassword(ctx context.Context, resetToken, newPassword string) error {
	// リセットトークンの検証
	token, err := s.passwordResetTokenRepo.FindByToken(ctx, resetToken)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %w", err)
	}

	if token == nil {
		return fmt.Errorf("invalid reset token")
	}

	now := s.timeProvider.Now()

	// リセットトークンの期限をチェック
	if now.After(token.ExpiresAt) {
		return fmt.Errorf("reset token expired")
	}

	// 使用済みチェック
	if token.UsedAt != nil {
		return fmt.Errorf("reset token already used")
	}

	// パスワードのハッシュ化
	hashedNewPassword, err := s.passwordManager.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 新しいパスワードへ置き換える
	err = s.userRepo.UpdatePassword(ctx, token.UserID, hashedNewPassword)
	if err != nil {
		return err
	}

	// リセットトークンを使用済みにする
	err = s.passwordResetTokenRepo.UpdateUseAtByID(ctx, token.ID, now)
	if err != nil {
		return fmt.Errorf("failed to update used_at: %w", err)
	}

	return nil
}
