package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/takehiro1111/gin-todo/backend/internal/services"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

type AdminController interface {
	GetAllUsers(c *gin.Context)
	GetAllTasks(c *gin.Context)
}

type AdminControllerImpl struct {
	adminService services.AdminService
	timeProvider *utils.RealTimeProvider
}

func NewAdminControllerImpl(adminService services.AdminService, timeProvider *utils.RealTimeProvider) *AdminControllerImpl {
	return &AdminControllerImpl{
		adminService: adminService,
		timeProvider: timeProvider,
	}
}

// GetAllUsers godoc
// @Summary      全ユーザー一覧取得
// @Description  管理者権限で全ユーザーの一覧を取得する
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200 {object} utils.SuccessResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      403 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/users [get]
func (a *AdminControllerImpl) GetAllUsers(c *gin.Context) {
	_, exist := c.Get("user_id")
	if !exist {
		utils.ResponseError(c, http.StatusUnauthorized,
			"invalid user access",
			"userID not found in context",
			a.timeProvider,
		)
		return
	}

	users, err := a.adminService.GetAllUsers(c)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed get users",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "success get all users", users, a.timeProvider)
}

// GetAllTasks godoc
// @Summary      全タスク一覧取得
// @Description  管理者権限で全ユーザーのタスク一覧を取得する
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200 {object} utils.SuccessResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      403 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/tasks [get]
func (a *AdminControllerImpl) GetAllTasks(c *gin.Context) {
	_, exist := c.Get("user_id")
	if !exist {
		utils.ResponseError(c, http.StatusUnauthorized,
			"invalid user access",
			"userID not found in context",
			a.timeProvider,
		)
		return
	}

	users, err := a.adminService.GetAllTasks(c)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed get tasks",
			err.Error(),
			a.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "success get tasks", users, a.timeProvider)
}
