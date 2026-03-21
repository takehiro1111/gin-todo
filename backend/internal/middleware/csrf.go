package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

type CSRFUUIDProvider struct {
	uuidGenerator utils.UUIDGenerator
}

func NewCSRFUUIDProvider(uuidGenerator utils.UUIDGenerator) *CSRFUUIDProvider {
	return &CSRFUUIDProvider{uuidGenerator: uuidGenerator}
}

func (g *CSRFUUIDProvider) GenerateCSRFToken(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	limit := 24 * (60 * 60)
	randomToken := g.uuidGenerator.UUIDGenerate()
	c.SetCookie("csrf_token", randomToken, limit, "/", "localhost", true, true)

	c.JSON(http.StatusOK, gin.H{"token": randomToken})
}

func (g *CSRFUUIDProvider) VerifyCSRFToken(c *gin.Context) {
	tokenHeaderCSRF := c.Request.Header.Get("X-CSRF-Token")

	cookie, err := c.Cookie("csrf_token")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, utils.ErrorResponse{
			Success:   false,
			Message:   "forbidden",
			Error:     "failed get cookie",
			Timestamp: time.Now(),
		})
		return
	}

	if tokenHeaderCSRF != cookie {
		c.AbortWithStatusJSON(http.StatusForbidden, utils.ErrorResponse{
			Success:   false,
			Message:   "forbidden",
			Error:     "csrf token mismatch",
			Timestamp: time.Now(),
		})
		return
	}

	c.Next()
}
