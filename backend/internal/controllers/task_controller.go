package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"github.com/takehiro1111/gin-todo/backend/internal/services"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

type TaskController interface {
	GetTasks(c *gin.Context)
	CreateTask(c *gin.Context)
	GetTaskByID(c *gin.Context)
	UpdateTask(c *gin.Context)
	UpdateTaskStatus(c *gin.Context)
	DeleteTask(c *gin.Context)
}

type TaskControllerImpl struct {
	taskService  services.TaskService
	timeProvider *utils.RealTimeProvider
}

type CreateRequest struct {
	Title       string     `json:"title" binding:"required,max=100"`
	Description string     `json:"description" binding:"max=300"`
	StatusID    int64      `json:"status_id"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateRequest struct {
	Title       string     `json:"title"  binding:"required,max=100"`
	Description string     `json:"description" binding:"max=100"`
	StatusID    int64      `json:"status_id"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateTaskStatusRequest struct {
	StatusID int64 `json:"status_id" binding:"required"`
}

func NewTaskController(taskService services.TaskService, timeProvider *utils.RealTimeProvider) *TaskControllerImpl {
	return &TaskControllerImpl{
		taskService:  taskService,
		timeProvider: timeProvider,
	}
}

// GetTasks godoc
// @Summary      タスク一覧取得
// @Description  ログイン中のユーザーのタスク一覧を取得する
// @Tags         task
// @Accept       json
// @Produce      json
// @Success      200 {object} utils.SuccessResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/task/ [get]
func (t *TaskControllerImpl) GetTasks(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		utils.ResponseError(c, http.StatusUnauthorized,
			"invalid user access",
			"userID not found in context",
			t.timeProvider,
		)
		return
	}

	tasks, err := t.taskService.GetAllTasks(c, userID.(uint))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed get tasks",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "success get all tasks", tasks, t.timeProvider)
}

// CreateTask godoc
// @Summary      タスク作成
// @Description  新しいタスクを作成する
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        request body CreateRequest true "タスク作成情報"
// @Success      201 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/task/ [post]
func (t *TaskControllerImpl) CreateTask(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// loggerを実装予定
		utils.ResponseError(c, http.StatusBadRequest,
			"invalid request",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	// middlewareでsetしているginのcontextから取得
	userID, exist := c.Get("user_id")
	if !exist {
		utils.ResponseError(c, http.StatusUnauthorized,
			"invalid user access",
			"userID not found in context",
			t.timeProvider,
		)
		return
	}

	task := &models.Task{
		UserID:      userID.(uint),
		Title:       req.Title,
		Description: req.Description,
		StatusID:    req.StatusID,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
	}

	data, err := t.taskService.CreateTask(c, task)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed create task",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusCreated, "success create task", data, t.timeProvider)
}

// GetTaskByID godoc
// @Summary      タスク取得
// @Description  指定されたIDのタスクを取得する
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        id path int true "タスクID"
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/task/{id} [get]
func (t *TaskControllerImpl) GetTaskByID(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest,
			"invalid id",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	userID, exist := c.Get("user_id")
	if !exist {
		utils.ResponseError(c, http.StatusUnauthorized,
			"invalid user access",
			"userID not found in context",
			t.timeProvider,
		)
		return
	}

	data, err := t.taskService.GetTaskByID(c, uint(idUint64), userID.(uint))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed get task by id",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "success get task by id", data, t.timeProvider)
}

// UpdateTask godoc
// @Summary      タスク更新
// @Description  指定されたIDのタスクを更新する
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        id path int true "タスクID"
// @Param        request body UpdateRequest true "タスク更新情報"
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/task/{id} [put]
func (t *TaskControllerImpl) UpdateTask(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest,
			"invalid id",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// loggerを実装予定
		utils.ResponseError(c, http.StatusBadRequest,
			"invalid request",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	userID, exist := c.Get("user_id")
	if !exist {
		utils.ResponseError(c, http.StatusUnauthorized,
			"invalid user access",
			"userID not found in context",
			t.timeProvider,
		)
		return
	}

	task := &models.Task{
		BaseModel:   models.BaseModel{ID: uint(idUint64)},
		UserID:      userID.(uint),
		Title:       req.Title,
		Description: req.Description,
		StatusID:    req.StatusID,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
	}

	data, err := t.taskService.UpdateTask(c, task)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed update task",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "success update task", data, t.timeProvider)
}

// UpdateTaskStatus godoc
// @Summary      タスクステータス更新
// @Description  指定されたIDのタスクのステータスを更新する
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        id path int true "タスクID"
// @Param        request body UpdateTaskStatusRequest true "ステータス更新情報"
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/task/{id} [patch]
func (t *TaskControllerImpl) UpdateTaskStatus(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest,
			"invalid id",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	userID, exist := c.Get("user_id")
	if !exist {
		utils.ResponseError(c, http.StatusUnauthorized,
			"invalid user access",
			"userID not found in context",
			t.timeProvider,
		)
		return
	}

	var req UpdateTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// loggerを実装予定
		utils.ResponseError(c, http.StatusBadRequest,
			"invalid request",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	err = t.taskService.UpdateTaskStatus(c, uint(idUint64), userID.(uint), req.StatusID)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed update task status id",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "success update task status id", nil, t.timeProvider)
}

// DeleteTask godoc
// @Summary      タスク削除
// @Description  指定されたIDのタスクを削除する
// @Tags         task
// @Accept       json
// @Produce      json
// @Param        id path int true "タスクID"
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/task/{id} [delete]
func (t *TaskControllerImpl) DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest,
			"invalid id",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	userID, exist := c.Get("user_id")
	if !exist {
		utils.ResponseError(c, http.StatusUnauthorized,
			"invalid user access",
			"userID not found in context",
			t.timeProvider,
		)
		return
	}

	err = t.taskService.DeleteTask(c, uint(idUint64), userID.(uint))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed delete task",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "success delete task", nil, t.timeProvider)
}
