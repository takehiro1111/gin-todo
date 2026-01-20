package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	appErr "github.com/takehiro1111/gin-todo/backend/internal/errors"
	"github.com/takehiro1111/gin-todo/backend/internal/services"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
	"github.com/takehiro1111/gin-todo/backend/internal/validators"
)

type AuthController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	RefreshToken(c *gin.Context)
	Logout(c *gin.Context)
	GetMe(c *gin.Context)
	ChangePassword(c *gin.Context)
	ForgotPassword(c *gin.Context)
	ResetPassword(c *gin.Context)
}

type AuthControllerImpl struct {
	authService  services.AuthService
	timeProvider *utils.RealTimeProvider
}

func NewAuthControllerImpl(authService services.AuthService, timeProvider *utils.RealTimeProvider) *AuthControllerImpl {
	return &AuthControllerImpl{
		authService:  authService,
		timeProvider: timeProvider,
	}
}

// 認証系はカスタムバリーデーションを実装しているためrequiredのみ設定。
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

type ResetPasswordRequest struct {
	ResetToken  string `json:"reset_token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// Register godoc
// @Summary      ユーザー登録
// @Description  新規ユーザーを登録する
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "登録情報"
// @Success      201 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /api/auth/register [post]
func (a *AuthControllerImpl) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// loggerを実装予定
		utils.ResponseError(c, http.StatusBadRequest,
			"invalid request",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	err := validators.NewAuthenticateValidator(
		validators.WithEmail(req.Email),
		validators.WithPassword(req.Password),
	)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest,
			"failed to validation",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	token, err := a.authService.Register(c, req.Name, req.Email, req.Password)

	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed register",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusCreated, "registered successfully", token, a.timeProvider)
}

// Login godoc
// @Summary      ログイン
// @Description  メールアドレスとパスワードでログインする
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "ログイン情報"
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      401 {object} utils.ErrorResponse
// @Router       /api/auth/login [post]
func (a *AuthControllerImpl) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// loggerを実装予定
		utils.ResponseError(c, http.StatusBadRequest,
			"invalid request",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	err := validators.NewAuthenticateValidator(
		validators.WithEmail(req.Email),
		validators.WithPassword(req.Password),
	)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest,
			"failed to validation",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	accessToken, refreshToken, err := a.authService.Login(c, req.Email, req.Password)
	if err != nil {
		utils.ResponseError(c, http.StatusUnauthorized,
			"failed login",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "login successfully", map[string]string{"access_token": accessToken, "refresh_token": refreshToken}, a.timeProvider)
}

// RefreshToken godoc
// @Summary      トークンリフレッシュ
// @Description  リフレッシュトークンを使用して新しいアクセストークンを取得する
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RefreshTokenRequest true "リフレッシュトークン"
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/auth/refresh [post]
func (a *AuthControllerImpl) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// loggerを実装予定
		utils.ResponseError(c, http.StatusBadRequest,
			"invalid request",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	token, err := a.authService.RefreshToken(c, req.RefreshToken)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed generate refresh token",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "generate refresh token successfully", token, a.timeProvider)
}

// Logout godoc
// @Summary      ログアウト
// @Description  ログアウトする
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200 {object} utils.SuccessResponse
// @Security     BearerAuth
// @Router       /api/auth/logout [post]
func (a *AuthControllerImpl) Logout(c *gin.Context) {
	utils.ResponseSuccess(c, http.StatusOK, "logout successfully", nil, a.timeProvider)
}

// GetMe godoc
// @Summary      ユーザー情報取得
// @Description  ログイン中のユーザー情報を取得する
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200 {object} utils.SuccessResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      404 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/auth/me [get]
func (a *AuthControllerImpl) GetMe(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		utils.ResponseError(c, http.StatusUnauthorized,
			"invalid user access",
			"userID not found in context",
			a.timeProvider,
		)
		return
	}

	user, err := a.authService.GetMe(c, userID.(string))
	if err != nil {
		if errors.Is(err, appErr.ErrUserNotFound) {
			utils.ResponseError(c, http.StatusNotFound,
				"user not found",
				err.Error(),
				a.timeProvider,
			)
			return
		}
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed get me by access token",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "get me successfully", user, a.timeProvider)
}

// ChangePassword godoc
// @Summary      パスワード変更
// @Description  ログイン中のユーザーのパスワードを変更する
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body ChangePasswordRequest true "パスワード変更情報"
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/auth/password [patch]
func (a *AuthControllerImpl) ChangePassword(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		utils.ResponseError(c, http.StatusUnauthorized,
			"invalid user access",
			"userID not found in context",
			a.timeProvider,
		)
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// loggerを実装予定
		utils.ResponseError(c, http.StatusBadRequest,
			"invalid request",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	fields := map[string]string{
		"OldPassword": req.OldPassword,
		"NewPassword": req.NewPassword,
	}

	for k, v := range fields {
		err := validators.NewAuthenticateValidator(
			validators.WithPassword(v),
		)
		if err != nil {
			utils.ResponseError(c, http.StatusBadRequest,
				fmt.Sprintf("failed to validation %s", k),
				err.Error(),
				a.timeProvider,
			)
			return
		}
	}

	if req.OldPassword == req.NewPassword {
		utils.ResponseError(c, http.StatusBadRequest,
			"invalid request",
			"new password must be different from old password",
			a.timeProvider,
		)
		return
	}

	err := a.authService.ChangePassword(c, userID.(string), req.OldPassword, req.NewPassword)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed change password",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "change password successfully", nil, a.timeProvider)
}

// ForgotPassword godoc
// @Summary      パスワードリセットのトークン生成
// @Description  パスワードリセットに使用するトークンの生成
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /api/auth/forgot [post]
func (a *AuthControllerImpl) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// loggerを実装予定
		utils.ResponseError(c, http.StatusBadRequest,
			"invalid request",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	err := validators.NewAuthenticateValidator(
		validators.WithEmail(req.Email),
	)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest,
			"failed to validation",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	// 後工程でリセットトークンをEmailでの検証を実装する。
	// その際にtokenを活用するよう修正予定。
	_, err = a.authService.ForgotPassword(c, req.Email)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed generate password reset token",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	// 後工程でメール送信を実装する
	utils.ResponseSuccess(c, http.StatusOK, "password reset email send successfully", nil, a.timeProvider)
}

// ResetPassword godoc
// @Summary      パスワードリセット
// @Description  トークン検証を行ってパスワードリセットを実行
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      201 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /api/auth/reset [post]
func (a *AuthControllerImpl) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// loggerを実装予定
		utils.ResponseError(c, http.StatusBadRequest,
			"invalid request",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	err := validators.NewAuthenticateValidator(
		validators.WithPassword(req.NewPassword),
	)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest,
			"failed to validation",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	err = a.authService.ResetPassword(c, req.ResetToken, req.NewPassword)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed reset password",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusCreated, "reset password successfully", nil, a.timeProvider)
}
