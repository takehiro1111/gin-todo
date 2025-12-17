package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

// routeを作成次第で適用
func VerifyRoleAdmin(jwtProvider utils.JWTProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtToken := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

		claims, err := jwtProvider.VerifyJWT(jwtToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
		}

		if claims.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "invalid role",
			})
		}

		c.Next()
	}
}
