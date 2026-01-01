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
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	StatusID    *int64     `json:"status_id"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"due_date"`
}

func NewTaskControlkler(taskService services.TaskService, timeProvider *utils.RealTimeProvider) *TaskControllerImpl {
	return &TaskControllerImpl{
		taskService:  taskService,
		timeProvider: timeProvider,
	}
}

func (t *TaskControllerImpl) GetTasks(c *gin.Context) {
	tasks, err := t.taskService.GetAllTasks(c)
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

	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
	}

	// StatusIDの入力が見られない場合のバリデーション
	if req.StatusID != nil {
		task.StatusID = *req.StatusID
	} else {
		task.StatusID = 1 // 一旦、便宜上1として後から正しいデフォルト値を
	}

	result, err := t.taskService.CreateTask(c, task)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed create task",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "success create task", result, t.timeProvider)
}

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

	task, err := t.taskService.GetTaskByID(c, uint(idUint64))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed get task by id",
			err.Error(),
			t.timeProvider,
		)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "success get task by id", task, t.timeProvider)
}

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

	err = t.taskService.DeleteTask(c, uint(idUint64))
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
