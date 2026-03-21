package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

func VerifyRoleAdmin(jwtProvider utils.JWTProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtToken := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

		claims, err := jwtProvider.VerifyJWT(jwtToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponse{
				Success:   false,
				Message:   "unauthorized",
				Error:     "invalid token",
				Timestamp: time.Now(),
			})
			return
		}

		if claims.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, utils.ErrorResponse{
				Success:   false,
				Message:   "forbidden",
				Error:     "invalid role",
				Timestamp: time.Now(),
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

func VerifyUser(jwtProvider utils.JWTProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtToken := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

		claims, err := jwtProvider.VerifyJWT(jwtToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponse{
				Success:   false,
				Message:   "unauthorized",
				Error:     "invalid token",
				Timestamp: time.Now(),
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
