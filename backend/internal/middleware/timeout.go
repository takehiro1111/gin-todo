package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(timeout time.Duration, loggerConfig *LoggerConfig) gin.HandlerFunc {
	slogger := NewLogger(loggerConfig)

	return func(c *gin.Context) {
		// WebSocketはタイムアウト対象外
		if strings.HasPrefix(c.Request.URL.Path, "/api/ws/") {
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		// 処理完了を通知するチャネル
		done := make(chan struct{}, 1)

		go func() {
			c.Next()
			done <- struct{}{}
		}()

		select {
		case <-done:
			// 正常完了
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				slogger.Warn("[Request Timeout]",
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.String("ip", c.ClientIP()),
					slog.Duration("timeout", timeout),
				)
				c.AbortWithStatus(http.StatusServiceUnavailable)
			}
		}
	}
}
