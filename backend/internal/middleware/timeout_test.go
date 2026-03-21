package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setupTimeoutRouter(timeout time.Duration, config *LoggerConfig) (*gin.Engine, *bytes.Buffer) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TimeoutMiddleware(timeout, config))
	return router, config.Writer.(*bytes.Buffer)
}

func TestTimeoutMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		timeout        time.Duration
		handlerDelay   time.Duration
		expectedStatus int
		expectedLogs   []string
		unexpectedLogs []string
	}{
		{
			name:           "タイムアウト内に完了した場合は正常レスポンス",
			timeout:        1 * time.Second,
			handlerDelay:   0,
			expectedStatus: http.StatusOK,
			unexpectedLogs: []string{
				"Request Timeout",
			},
		},
		{
			name:           "タイムアウト超過した場合は503を返す",
			timeout:        50 * time.Millisecond,
			handlerDelay:   200 * time.Millisecond,
			expectedStatus: http.StatusServiceUnavailable,
			expectedLogs: []string{
				"Request Timeout",
				"method",
				"path",
				"ip",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			config := NewLoggerConfig(
				"info",
				"text",
				WithWriter(&buf),
			)

			router, logBuf := setupTimeoutRouter(tc.timeout, config)
			router.GET("/test", func(c *gin.Context) {
				if tc.handlerDelay > 0 {
					select {
					case <-time.After(tc.handlerDelay):
					case <-c.Request.Context().Done():
						return
					}
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("ステータスコードが期待と異なります: got=%d, want=%d", w.Code, tc.expectedStatus)
			}

			logOutput := logBuf.String()

			for _, expected := range tc.expectedLogs {
				if !bytes.Contains(logBuf.Bytes(), []byte(expected)) {
					t.Errorf("期待されるログが含まれていません: want=%s\nログ: %s", expected, logOutput)
				}
			}

			for _, unexpected := range tc.unexpectedLogs {
				if bytes.Contains(logBuf.Bytes(), []byte(unexpected)) {
					t.Errorf("含まれるべきでないログが出力されています: unwanted=%s\nログ: %s", unexpected, logOutput)
				}
			}
		})
	}
}

func TestTimeoutMiddleware_WebSocketSkip(t *testing.T) {
	var buf bytes.Buffer

	config := NewLoggerConfig(
		"info",
		"text",
		WithWriter(&buf),
	)

	router, logBuf := setupTimeoutRouter(50*time.Millisecond, config)
	router.GET("/api/ws/chat", func(c *gin.Context) {
		// WebSocketはタイムアウト対象外なので、遅延があっても正常完了する
		time.Sleep(100 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/api/ws/chat", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("WebSocketパスはタイムアウト対象外のはず: got=%d, want=%d", w.Code, http.StatusOK)
	}

	logOutput := logBuf.String()
	if bytes.Contains(logBuf.Bytes(), []byte("Request Timeout")) {
		t.Errorf("WebSocketパスでタイムアウトログが出力されています\nログ: %s", logOutput)
	}
}

func TestTimeoutMiddleware_ContextPropagation(t *testing.T) {
	var buf bytes.Buffer

	config := NewLoggerConfig(
		"info",
		"text",
		WithWriter(&buf),
	)

	router, _ := setupTimeoutRouter(50*time.Millisecond, config)

	contextCancelled := false
	router.GET("/test", func(c *gin.Context) {
		select {
		case <-time.After(200 * time.Millisecond):
		case <-c.Request.Context().Done():
			contextCancelled = true
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if !contextCancelled {
		t.Error("タイムアウト時にコンテキストがキャンセルされていません")
	}
}
