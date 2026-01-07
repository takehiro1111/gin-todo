package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/takehiro1111/gin-todo/backend/internal/controllers"
	"github.com/takehiro1111/gin-todo/backend/internal/middleware"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
	"net/http"
)

// main関数をシンプルにするため。
func SetupRoutes(r *gin.Engine, authCtrl controllers.AuthController, taskCtrl controllers.TaskController, jwtProvider utils.JWTProvider) {
	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		api.GET("/health", HealthCheck)
	}

	apiAuth := r.Group("/api/auth")
	{
		apiAuth.POST("/register", authCtrl.Register)
		apiAuth.POST("/login", authCtrl.Login)
		apiAuth.POST("/logout", authCtrl.Logout)
		apiAuth.POST("/refresh", authCtrl.RefreshToken)
		// /auth/me  , /auth/refresh , /auth/password , /auth/forgot は追加予定
	}

	apiTask := r.Group("/api/task")
	apiTask.Use(middleware.VerifyUser(jwtProvider))
	{
		apiTask.GET("/", taskCtrl.GetTasks)
		apiTask.POST("/", taskCtrl.CreateTask)
		apiTask.GET("/:id", taskCtrl.GetTaskByID)
		apiTask.PUT("/:id", taskCtrl.UpdateTask)
		apiTask.PATCH("/:id", taskCtrl.UpdateTaskStatus)
		apiTask.DELETE("/:id", taskCtrl.DeleteTask)
	}

	// 管理者機能を後で追加する
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
