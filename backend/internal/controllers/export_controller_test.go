package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

// TaskFetcher のmock
type mockTaskFetcher struct {
	tasks []models.Task
	err   error
}

func (m *mockTaskFetcher) GetAllTasks(ctx context.Context, userID uint) ([]models.Task, error) {
	return m.tasks, m.err
}

func setupExportTest(mock *mockTaskFetcher) (*ExportControllerImpl, *httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/tasks/export/csv", nil)

	timeProvider := utils.NewRealTimeProvider()
	ctrl := NewExportControllerImpl(mock, timeProvider)

	return ctrl, w, c
}

func TestExportTasksCSV(t *testing.T) {
	tests := []struct {
		name       string
		setUserID  bool
		userID     uint
		mockTasks  []models.Task
		mockErr    error
		wantStatus int
		wantCSV    bool
	}{
		{
			name:      "[正常系] タスクが複数件ある場合",
			setUserID: true,
			userID:    1,
			mockTasks: []models.Task{
				{BaseModel: models.BaseModel{ID: 10}, Title: "タスク1"},
				{BaseModel: models.BaseModel{ID: 25}, Title: "タスク2"},
				{BaseModel: models.BaseModel{ID: 99}, Title: "タスク3"},
			},
			wantStatus: http.StatusOK,
			wantCSV:    true,
		},
		{
			name:      "[正常系] タスクが1件の場合",
			setUserID: true,
			userID:    1,
			mockTasks: []models.Task{
				{BaseModel: models.BaseModel{ID: 5}, Title: "テストタスク"},
			},
			wantStatus: http.StatusOK,
			wantCSV:    true,
		},
		{
			name:       "[異常系] user_idがcontextにない場合",
			setUserID:  false,
			wantStatus: http.StatusUnauthorized,
			wantCSV:    false,
		},
		{
			name:       "[異常系] GetAllTasksがエラーを返す場合",
			setUserID:  true,
			userID:     1,
			mockErr:    errors.New("db connection error"),
			wantStatus: http.StatusInternalServerError,
			wantCSV:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockTaskFetcher{tasks: tt.mockTasks, err: tt.mockErr}
			ctrl, w, c := setupExportTest(mock)

			if tt.setUserID {
				c.Set("user_id", tt.userID)
			}

			ctrl.ExportTasksCSV(c)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantCSV {
				// CSVレスポンスのヘッダー検証
				assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
				assert.Equal(t, "attachment;filename=export_all_task.csv", w.Header().Get("Content-Disposition"))

				// ヘッダー行の検証
				body := w.Body.String()
				assert.Contains(t, body, "ID,Title")
			} else {
				// エラーレスポンスのJSON検証
				var errResp utils.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				require.NoError(t, err)
				assert.False(t, errResp.Success)
			}
		})
	}
}

// DB IDではなくループインデックス(1始まり)がCSVのIDになることを検証
func TestExportTasksCSV_IDIsLoopIndex(t *testing.T) {
	mock := &mockTaskFetcher{
		tasks: []models.Task{
			{BaseModel: models.BaseModel{ID: 10}, Title: "タスクA"},
			{BaseModel: models.BaseModel{ID: 25}, Title: "タスクB"},
			{BaseModel: models.BaseModel{ID: 99}, Title: "タスクC"},
		},
	}
	ctrl, w, c := setupExportTest(mock)
	c.Set("user_id", uint(1))

	ctrl.ExportTasksCSV(c)

	body := w.Body.String()

	// DB ID(10,25,99)ではなく連番(1,2,3)が使われること
	assert.Contains(t, body, "1,タスクA")
	assert.Contains(t, body, "2,タスクB")
	assert.Contains(t, body, "3,タスクC")
	assert.NotContains(t, body, "10,")
	assert.NotContains(t, body, "25,")
	assert.NotContains(t, body, "99,")
}

// タスク0件でもヘッダー行が出力されることを検証
func TestExportTasksCSV_EmptyCSVHasHeaderOnly(t *testing.T) {
	mock := &mockTaskFetcher{tasks: []models.Task{}}
	ctrl, w, c := setupExportTest(mock)
	c.Set("user_id", uint(1))

	ctrl.ExportTasksCSV(c)

	body := w.Body.String()
	assert.Equal(t, "ID,Title\n", body)
}
