package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func CustomRecover(loggerConfig *LoggerConfig) gin.HandlerFunc {
	slogger := NewLogger(loggerConfig)

	return func(c *gin.Context) {
		defer func() {
			err := recover()
			if err != nil {
				slogger.Error("[Panic Recovery]",
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.String("ip", c.ClientIP()),
					slog.Any("error", err),
					slog.String("stack_trace", string(debug.Stack())),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
