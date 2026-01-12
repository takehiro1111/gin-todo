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

	token, err := a.authService.Register(req.Name, req.Email, req.Password)

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

	token, err := a.authService.Login(req.Email, req.Password)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed login",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "login successfully", token, a.timeProvider)
}

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

	token, err := a.authService.RefreshToken(req.RefreshToken)
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

func (a *AuthControllerImpl) Logout(c *gin.Context) {
	utils.ResponseSuccess(c, http.StatusOK, "logout successfully", nil, a.timeProvider)
}

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

	user, err := a.authService.GetMe(userID.(string))
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

	err := a.authService.ChangePassword(userID.(string), req.OldPassword, req.NewPassword)
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
