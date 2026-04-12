package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	appErr "github.com/takehiro1111/gin-todo/backend/internal/errors"
	"github.com/takehiro1111/gin-todo/backend/internal/middleware"
	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"github.com/takehiro1111/gin-todo/backend/internal/services"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
	"github.com/takehiro1111/gin-todo/backend/internal/validators"
)

type AuthController interface {
	Register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error)
	Login(ctx context.Context, input *LoginInput) (*LoginOutput, error)
	RefreshToken(ctx context.Context, input *RefreshTokenInput) (*RefreshTokenOutput, error)
	Logout(ctx context.Context, input *LogoutInput) (*LogoutOutput, error)
	GetMe(ctx context.Context, input *GetMeInput) (*GetMeOutput, error)
	ChangePassword(ctx context.Context, input *ChangePasswordInput) (*ChangePasswordOutput, error)
	ForgotPassword(ctx context.Context, input *ForgotPasswordInput) (*ForgotPasswordOutput, error)
	ResetPassword(ctx context.Context, input *ResetPasswordInput) (*ResetPasswordOutput, error)
	GenerateCSRFToken(ctx context.Context, input *GenerateCSRFTokenInput) (*GenerateCSRFTokenOutput, error)
}

type AuthControllerImpl struct {
	authService   services.AuthService
	timeProvider  *utils.RealTimeProvider
	uuidGenerator utils.UUIDGenerator
}

func NewAuthControllerImpl(authService services.AuthService, timeProvider *utils.RealTimeProvider, uuidGenerator utils.UUIDGenerator) *AuthControllerImpl {
	return &AuthControllerImpl{
		authService:   authService,
		timeProvider:  timeProvider,
		uuidGenerator: uuidGenerator,
	}
}

// --- Request DTOs ---

type RegisterRequest struct {
	Name     string `json:"name" required:"true" doc:"ユーザー名"`
	Email    string `json:"email" required:"true" doc:"メールアドレス"`
	Password string `json:"password" required:"true" doc:"パスワード"`
}

type LoginRequest struct {
	Email    string `json:"email" required:"true" doc:"メールアドレス"`
	Password string `json:"password" required:"true" doc:"パスワード"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" required:"true" doc:"現在のパスワード"`
	NewPassword string `json:"new_password" required:"true" doc:"新しいパスワード"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" required:"true" doc:"メールアドレス"`
}

type ResetPasswordRequest struct {
	ResetToken  string `json:"reset_token" required:"true" doc:"パスワードリセットトークン"`
	NewPassword string `json:"new_password" required:"true" doc:"新しいパスワード"`
}

// --- Input / Output ---

type RegisterInput struct {
	Body RegisterRequest
}
type RegisterOutput struct {
	Body SuccessBody[string]
}

type LoginInput struct {
	Body LoginRequest
}
type LoginOutput struct {
	SetCookie string `header:"Set-Cookie"`
	Body      SuccessBody[map[string]string]
}

type RefreshTokenInput struct {
	RefreshToken string `cookie:"refresh_token" doc:"リフレッシュトークン"`
}
type RefreshTokenOutput struct {
	SetCookie string `header:"Set-Cookie"`
	Body      SuccessBody[string]
}

type LogoutInput struct{}
type LogoutOutput struct {
	SetCookie string `header:"Set-Cookie"`
	Body      SuccessBody[*struct{}]
}

type GetMeInput struct{}
type GetMeOutput struct {
	Body SuccessBody[*models.User]
}

type ChangePasswordInput struct {
	Body ChangePasswordRequest
}
type ChangePasswordOutput struct {
	Body SuccessBody[*struct{}]
}

type ForgotPasswordInput struct {
	Body ForgotPasswordRequest
}
type ForgotPasswordOutput struct {
	Body SuccessBody[*struct{}]
}

type ResetPasswordInput struct {
	Body ResetPasswordRequest
}
type ResetPasswordOutput struct {
	Body SuccessBody[*struct{}]
}

// --- CSRF ---

type GenerateCSRFTokenInput struct{}
type CSRFTokenBody struct {
	Token string `json:"token" doc:"CSRFトークン"`
}
type GenerateCSRFTokenOutput struct {
	SetCookie string `header:"Set-Cookie"`
	Body      CSRFTokenBody
}

// --- Cookie helpers ---

func refreshTokenCookie(value string, maxAge int) string {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	return cookie.String()
}

// --- Handlers ---

func (a *AuthControllerImpl) Register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error) {
	err := validators.NewAuthenticateValidator(
		validators.WithEmail(input.Body.Email),
		validators.WithPassword(input.Body.Password),
	)
	if err != nil {
		return nil, huma.Error400BadRequest("failed to validation", err)
	}

	token, err := a.authService.Register(ctx, input.Body.Name, input.Body.Email, input.Body.Password)
	if err != nil {
		if errors.Is(err, appErr.ErrAlreadyExists) {
			return nil, huma.Error409Conflict("user already exists")
		}
		return nil, huma.Error500InternalServerError("failed register")
	}

	return &RegisterOutput{
		Body: NewSuccessBody("registered successfully", token, a.timeProvider),
	}, nil
}

func (a *AuthControllerImpl) Login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	err := validators.NewAuthenticateValidator(
		validators.WithEmail(input.Body.Email),
		validators.WithPassword(input.Body.Password),
	)
	if err != nil {
		return nil, huma.Error400BadRequest("failed to validation", err)
	}

	accessToken, refreshToken, err := a.authService.Login(ctx, input.Body.Email, input.Body.Password)
	if err != nil {
		if errors.Is(err, appErr.ErrUnauthorized) || errors.Is(err, appErr.ErrUserNotFound) {
			return nil, huma.Error401Unauthorized("invalid email or password")
		}
		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &LoginOutput{
		SetCookie: refreshTokenCookie(refreshToken, 60*60*24*7),
		Body:      NewSuccessBody("login successfully", map[string]string{"access_token": accessToken}, a.timeProvider),
	}, nil
}

func (a *AuthControllerImpl) RefreshToken(ctx context.Context, input *RefreshTokenInput) (*RefreshTokenOutput, error) {
	if input.RefreshToken == "" {
		return nil, huma.Error401Unauthorized("missing refresh token")
	}

	token, err := a.authService.RefreshToken(ctx, input.RefreshToken)
	if err != nil {
		return nil, huma.Error401Unauthorized("invalid or expired refresh token")
	}

	return &RefreshTokenOutput{
		SetCookie: refreshTokenCookie(token.RefreshToken, 60*60*24*7),
		Body:      NewSuccessBody("generate refresh token successfully", token.AccessToken, a.timeProvider),
	}, nil
}

func (a *AuthControllerImpl) Logout(ctx context.Context, input *LogoutInput) (*LogoutOutput, error) {
	return &LogoutOutput{
		SetCookie: refreshTokenCookie("", -1),
		Body:      NewSuccessBody[*struct{}]("logout successfully", nil, a.timeProvider),
	}, nil
}

func (a *AuthControllerImpl) GetMe(ctx context.Context, input *GetMeInput) (*GetMeOutput, error) {
	userID, ok := ctx.Value(middleware.UserIDKey).(uint)
	if !ok {
		return nil, huma.Error401Unauthorized("invalid user access")
	}

	user, err := a.authService.GetMe(ctx, fmt.Sprintf("%d", userID))
	if err != nil {
		if errors.Is(err, appErr.ErrUserNotFound) {
			return nil, huma.Error404NotFound("user not found")
		}
		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &GetMeOutput{
		Body: NewSuccessBody("get me successfully", user, a.timeProvider),
	}, nil
}

func (a *AuthControllerImpl) ChangePassword(ctx context.Context, input *ChangePasswordInput) (*ChangePasswordOutput, error) {
	userID, ok := ctx.Value(middleware.UserIDKey).(uint)
	if !ok {
		return nil, huma.Error401Unauthorized("invalid user access")
	}

	fields := map[string]string{
		"OldPassword": input.Body.OldPassword,
		"NewPassword": input.Body.NewPassword,
	}

	for k, v := range fields {
		err := validators.NewAuthenticateValidator(
			validators.WithPassword(v),
		)
		if err != nil {
			return nil, huma.Error400BadRequest(fmt.Sprintf("failed to validation %s", k), err)
		}
	}

	if input.Body.OldPassword == input.Body.NewPassword {
		return nil, huma.Error400BadRequest("new password must be different from old password")
	}

	err := a.authService.ChangePassword(ctx, fmt.Sprintf("%d", userID), input.Body.OldPassword, input.Body.NewPassword)
	if err != nil {
		if errors.Is(err, appErr.ErrUnauthorized) {
			return nil, huma.Error401Unauthorized("invalid old password")
		}
		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &ChangePasswordOutput{
		Body: NewSuccessBody[*struct{}]("change password successfully", nil, a.timeProvider),
	}, nil
}

func (a *AuthControllerImpl) ForgotPassword(ctx context.Context, input *ForgotPasswordInput) (*ForgotPasswordOutput, error) {
	err := validators.NewAuthenticateValidator(
		validators.WithEmail(input.Body.Email),
	)
	if err != nil {
		return nil, huma.Error400BadRequest("failed to validation", err)
	}

	// ユーザー列挙攻撃を防ぐため、メールの存在有無に関わらず同一レスポンスを返す
	_ = a.authService.ForgotPassword(ctx, input.Body.Email)

	return &ForgotPasswordOutput{
		Body: NewSuccessBody[*struct{}]("if the email exists, a reset link has been sent", nil, a.timeProvider),
	}, nil
}

func (a *AuthControllerImpl) ResetPassword(ctx context.Context, input *ResetPasswordInput) (*ResetPasswordOutput, error) {
	err := validators.NewAuthenticateValidator(
		validators.WithPassword(input.Body.NewPassword),
	)
	if err != nil {
		return nil, huma.Error400BadRequest("failed to validation", err)
	}

	err = a.authService.ResetPassword(ctx, input.Body.ResetToken, input.Body.NewPassword)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) || errors.Is(err, appErr.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("invalid or expired reset token")
		}
		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &ResetPasswordOutput{
		Body: NewSuccessBody[*struct{}]("reset password successfully", nil, a.timeProvider),
	}, nil
}

func (a *AuthControllerImpl) GenerateCSRFToken(ctx context.Context, input *GenerateCSRFTokenInput) (*GenerateCSRFTokenOutput, error) {
	token := a.uuidGenerator.UUIDGenerate()
	cookie := &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		MaxAge:   24 * 60 * 60,
		Path:     "/",
		Domain:   "localhost",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	return &GenerateCSRFTokenOutput{
		SetCookie: cookie.String(),
		Body:      CSRFTokenBody{Token: token},
	}, nil
}
