package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/take/gin-todo/docs"
)

// @title           Gin Todo API
// @version         1.0
// @description     タスク管理APIのフルスタックアプリケーション
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	r := gin.Default()

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/hello", HelloWorld)
	}

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.GET("/health", HealthCheck)

	r.Run(":8080")
}

// HelloWorld godoc
// @Summary      Hello World
// @Description  簡単なHello Worldエンドポイント
// @Tags         example
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /hello [get]
func HelloWorld(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello World!",
	})
}

// HealthCheck godoc
// @Summary      ヘルスチェック
// @Description  APIサーバーの稼働状態を確認
// @Tags         health
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
