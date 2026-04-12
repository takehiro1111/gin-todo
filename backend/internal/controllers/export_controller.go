package controllers

import (
	"context"
	"encoding/csv"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

// ExportController はCSVエクスポート機能のインターフェース
type ExportController interface {
	ExportTasksCSV(c *gin.Context)
}

// TaskFetcher はExportControllerが必要とするメソッドのみを定義したインターフェース
type TaskFetcher interface {
	GetAllTasks(ctx context.Context, userID uint) ([]models.Task, error)
}

type ExportControllerImpl struct {
	taskService  TaskFetcher
	timeProvider *utils.RealTimeProvider
}

func NewExportControllerImpl(taskService TaskFetcher, timeProvider *utils.RealTimeProvider) *ExportControllerImpl {
	return &ExportControllerImpl{
		taskService:  taskService,
		timeProvider: timeProvider,
	}
}

// ExportTasksCSV はログインユーザーのタスク一覧を CSV 形式でダウンロードする。
func (e *ExportControllerImpl) ExportTasksCSV(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		utils.ResponseError(c, http.StatusUnauthorized,
			"invalid user access",
			"userID not found in context",
			e.timeProvider,
		)
		return
	}

	tasks, err := e.taskService.GetAllTasks(c, userID.(uint))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError,
			"failed get tasks",
			err.Error(),
			e.timeProvider,
		)
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment;filename=export_all_task.csv")
	c.Header("Transfer-Encoding", "chunked")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	err = writer.Write([]string{"ID", "Title"})
	if err != nil {
		// CSV出力された後にエラーになった場合はログ用に記録だけ
		c.Error(err)
		return
	}

	for idx, task := range tasks {
		// CSVでID表示する際はユーザーのタスクのみ加味したいため、idxで処理。
		err := writer.Write([]string{strconv.Itoa(idx + 1), task.Title})
		if err != nil {
			c.Error(err)
			return
		}
	}
}
