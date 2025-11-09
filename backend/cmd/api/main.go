package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
	"os"

	"github.com/takehiro1111/gin-todo/backend/internal/models"
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

	err := godotenv.Load("cmd/api/.env")
	if err != nil {
		log.Fatal("failed read .env")
	}

	// 後工程でSSMパラメータストアから取得する実装に変更予定
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB_NAME")
	dbPort := os.Getenv("POSTGRES_DB_PORT")
	dbHost := os.Getenv("POSTGRES_DB_HOST")
	sslMode := os.Getenv("POSTGRES_DB_SSL_MODE")
	tz := "Asia/Tokyo"

	// 後工程で戻り値を活用する
	_, err = models.DBInit(dbUser, dbPassword, dbHost, dbPort, dbName, sslMode, tz)
	if err != nil {
		log.Fatalf("db initialization failed: %v", err)
	}

	// 後工程で本番環境は実行しないよう修正する
	err = models.RunMigration(dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)
	// マイグレーションで変更のない場合はエラーにしない(migrate.ErrNoChange)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

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
	c.JSON(http.StatusOK, gin.H{
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
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
