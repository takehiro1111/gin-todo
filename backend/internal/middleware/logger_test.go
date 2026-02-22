package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// テストヘルパー: MockUUIDGenerator
type MockUUIDGenerator struct {
	ID string
}

func (m *MockUUIDGenerator) Generate() string {
	return m.ID
}

// テストヘルパー: ログ出力をキャプチャする関数
func setupTestRouter(config *LoggerConfig) (*gin.Engine, *bytes.Buffer) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(LoggerMiddleware(config))
	return router, config.Writer.(*bytes.Buffer)
}

// 機密情報のマスキング
func TestMaskSensitiveData(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		shouldMask int
	}{
		{
			name:       "トップレベルのpasswordをマスキング",
			input:      `{"username":"john","password":"secret123"}`,
			shouldMask: len([]string{"password"}),
		},
		{
			name:       "ネストしたtokenをマスキング",
			input:      `{"user":{"name":"john","token":"abc-xyz"}}`,
			shouldMask: len([]string{"token"}),
		},
		{
			name:       "配列内のapi_keyをマスキング",
			input:      `{"items":[{"name":"item1","api_key":"key1"},{"name":"item2","api_key":"key2"}]}`,
			shouldMask: len([]string{"api_key"}),
		},
		{
			name:       "複数の機密フィールドをマスキング",
			input:      `{"username":"john","password":"pass","token":"tok","secret":"sec"}`,
			shouldMask: len([]string{"password", "token", "secret"}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			mockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
			config := NewLoggerConfig(
				"info",
				"text",
				WithWriter(&buf),
				WithEnableRequestBody(true),
				WithTimeFunc(func() time.Time { return mockTime }),
				WithUUIDGenerator(&MockUUIDGenerator{ID: "test-id"}),
			)

			router, _ := setupTestRouter(config)
			router.POST("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tc.input))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			logOutput := buf.String()

			if !strings.Contains(logOutput, "***MASKED***") {
				t.Errorf("機密情報がマスキングされていません\nログ: %s", logOutput)
			}
		})
	}
}

// スキップパス機能
func TestSkipPaths(t *testing.T) {
	tests := []struct {
		name        string
		skipPaths   []string
		requestPath string
		shouldLog   bool
	}{
		{
			name:        "スキップパスに一致する場合はログ出力しない",
			skipPaths:   []string{"/health", "/metrics"},
			requestPath: "/health",
			shouldLog:   false,
		},
		{
			name:        "スキップパスに一致しない場合はログ出力する",
			skipPaths:   []string{"/health", "/metrics"},
			requestPath: "/api/users",
			shouldLog:   true,
		},
		{
			name:        "スキップパスが空の場合は全てログ出力",
			skipPaths:   []string{},
			requestPath: "/health",
			shouldLog:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			config := NewLoggerConfig(
				"info",
				"text",
				WithWriter(&buf),
				WithSkipPaths(tc.skipPaths),
			)

			router, logBuf := setupTestRouter(config)
			router.GET(tc.requestPath, func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest("GET", tc.requestPath, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			logOutput := logBuf.String()
			hasLog := len(logOutput) > 0

			if hasLog != tc.shouldLog {
				t.Errorf("ログ出力の期待値が異なります: got=%v, want=%v\nログ: %s",
					hasLog, tc.shouldLog, logOutput)
			}
		})
	}
}

// ステータスコード別ログレベル
func TestLogLevel(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedLevel string
	}{
		{
			name:          "200番台はINFO",
			statusCode:    200,
			expectedLevel: "INFO",
		},
		{
			name:          "404はWARN",
			statusCode:    404,
			expectedLevel: "WARN",
		},
		{
			name:          "500番台はERROR",
			statusCode:    500,
			expectedLevel: "ERROR",
		},
		{
			name:          "503もERROR",
			statusCode:    503,
			expectedLevel: "ERROR",
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

			router, logBuf := setupTestRouter(config)
			router.GET("/test", func(c *gin.Context) {
				c.Status(tc.statusCode)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			logOutput := logBuf.String()

			if !bytes.Contains(logBuf.Bytes(), []byte("level="+tc.expectedLevel)) {
				t.Errorf("期待されるログレベルが含まれていません: got=%s, want=level=%s",
					logOutput, tc.expectedLevel)
			}
		})
	}
}

// レイテンシー計測
func TestLatencyMeasurement(t *testing.T) {
	tests := []struct {
		name            string
		mockLatencyMs   int
		expectedLatency string
	}{
		{
			name:            "100ms遅延",
			mockLatencyMs:   100,
			expectedLatency: "100.00ms",
		},
		{
			name:            "0ms遅延",
			mockLatencyMs:   0,
			expectedLatency: "0.00ms",
		},
		{
			name:            "50ms遅延",
			mockLatencyMs:   50,
			expectedLatency: "50.00ms",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			// Mock時刻関数
			callCount := 0
			mockTimeFunc := func() time.Time {
				baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
				result := baseTime.Add(time.Duration(callCount*tc.mockLatencyMs) * time.Millisecond)
				callCount++
				return result
			}

			config := NewLoggerConfig(
				"info",
				"text",
				WithWriter(&buf),
				WithTimeFunc(mockTimeFunc),
			)

			router, logBuf := setupTestRouter(config)
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			logOutput := logBuf.String()

			if !bytes.Contains(logBuf.Bytes(), []byte("latency="+tc.expectedLatency)) {
				t.Errorf("期待されるレイテンシーが含まれていません: got=%s, want=latency=%s",
					logOutput, tc.expectedLatency)
			}
		})
	}
}

// リクエストID
func TestRequestID(t *testing.T) {
	tests := []struct {
		name              string
		headerRequestID   string
		expectedRequestID string
	}{
		{
			name:              "ヘッダーにリクエストIDがある場合",
			headerRequestID:   "custom-id-12345",
			expectedRequestID: "custom-id-12345",
		},
		{
			name:              "ヘッダーにリクエストIDがない場合は自動生成",
			headerRequestID:   "",
			expectedRequestID: "mock-uuid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			config := NewLoggerConfig(
				"info",
				"text",
				WithWriter(&buf),
				WithUUIDGenerator(&MockUUIDGenerator{ID: "mock-uuid"}),
			)

			router, logBuf := setupTestRouter(config)
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tc.headerRequestID != "" {
				req.Header.Set("X-Request-ID", tc.headerRequestID)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			logOutput := logBuf.String()

			if !bytes.Contains(logBuf.Bytes(), []byte("request_id="+tc.expectedRequestID)) {
				t.Errorf("期待されるリクエストIDが含まれていません: got=%s, want=request_id=%s",
					logOutput, tc.expectedRequestID)
			}
		})
	}
}

// Content-Type判定
func TestContentTypeFiltering(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		shouldLog   bool
	}{
		{
			name:        "application/jsonはログ出力",
			contentType: "application/json",
			body:        `{"test":"value"}`,
			shouldLog:   true,
		},
		{
			name:        "text/plainはログ出力",
			contentType: "text/plain",
			body:        "plain text",
			shouldLog:   true,
		},
		{
			name:        "image/pngはログ出力しない",
			contentType: "image/png",
			body:        "binary data",
			shouldLog:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			config := NewLoggerConfig(
				"info",
				"text",
				WithWriter(&buf),
				WithEnableRequestBody(true),
			)

			router, logBuf := setupTestRouter(config)
			router.POST("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", tc.contentType)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			logOutput := logBuf.String()
			hasRequestBody := bytes.Contains(logBuf.Bytes(), []byte("request_body="))

			if hasRequestBody != tc.shouldLog {
				t.Errorf("リクエストボディのログ出力が期待と異なります: got=%v, want=%v\nログ: %s",
					hasRequestBody, tc.shouldLog, logOutput)
			}
		})
	}
}
