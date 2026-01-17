package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/takehiro1111/gin-todo/backend/docs"
	"github.com/takehiro1111/gin-todo/backend/internal/controllers"
	"github.com/takehiro1111/gin-todo/backend/internal/middleware"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
	"net/http"
)

// main関数をシンプルにするため。
func SetupRoutes(r *gin.Engine, adminCtrl controllers.AdminController, authCtrl controllers.AuthController, taskCtrl controllers.TaskController, jwtProvider utils.JWTProvider) {
	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		api.GET("/health", HealthCheck)
	}

	admin := r.Group("/api/admin")
	admin.Use(middleware.VerifyRoleAdmin(jwtProvider))
	{
		admin.GET("/users", adminCtrl.GetAllUsers)
		admin.GET("/tasks", adminCtrl.GetAllTasks)
	}

	// 認証不要（ログイン前）
	authPublic := r.Group("/api/auth")
	{
		authPublic.POST("/register", authCtrl.Register)
		authPublic.POST("/login", authCtrl.Login)
	}

	// 認証必須（ログイン後）
	authProtected := r.Group("/api/auth")
	authProtected.Use(middleware.VerifyUser(jwtProvider))
	{
		authProtected.POST("/logout", authCtrl.Logout)
		authProtected.POST("/refresh", authCtrl.RefreshToken)
		authProtected.GET("/me", authCtrl.GetMe)
		authProtected.PATCH("/password", authCtrl.ChangePassword)
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

// HealthCheck godoc
// @Summary      ヘルスチェック
// @Description  APIサーバーの稼働状態を確認
// @Tags         health
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /api/health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
