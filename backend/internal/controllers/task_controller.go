package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/takehiro1111/gin-todo/backend/internal/services"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

type TaskController interface {
	GetTasks()
	CreateTask()
	GetTaskByID()
	UpdateTask()
	UpdateTaskStatus()
	DeleteTask()
}

type TaskControllerImpl struct {
	taskService  services.TaskService
	timeProvider *utils.RealTimeProvider
}

func NewTaskControlkler(taskService services.TaskService, timeProvider *utils.RealTimeProvider) *TaskControllerImpl {
	return &TaskControllerImpl{
		taskService:  taskService,
		timeProvider: timeProvider,
	}
}
