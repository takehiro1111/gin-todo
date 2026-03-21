package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	gorillaWs "github.com/gorilla/websocket"

	"github.com/takehiro1111/gin-todo/backend/infrastructure/aws"
	"github.com/takehiro1111/gin-todo/backend/internal/controllers"
	"github.com/takehiro1111/gin-todo/backend/internal/middleware"
	"github.com/takehiro1111/gin-todo/backend/internal/models"
	"github.com/takehiro1111/gin-todo/backend/internal/repositories"
	"github.com/takehiro1111/gin-todo/backend/internal/routes"
	"github.com/takehiro1111/gin-todo/backend/internal/services"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
	"github.com/takehiro1111/gin-todo/backend/internal/websocket"
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
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	r := gin.New()

	r.Use(middleware.SetSecurityResponseHeader())
	loggerConfig := middleware.NewLoggerConfig(
		"info",
		"json",
		middleware.WithTimeFunc(time.Now),
		middleware.WithWriter(os.Stdout),
		middleware.WithEnableRequestBody(true),
	)
	r.Use(middleware.LoggerMiddleware(loggerConfig))
	r.Use(middleware.CustomRecover(loggerConfig))
	r.Use(middleware.TimeoutMiddleware(5*time.Second, loggerConfig))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPathsRegexs([]string{"/api/health"})))

	// フィールドを持たないのでファクトリーメソッドは実装してない
	uuidGenerator := &utils.UUIDGeneratorImpl{}

	csrfTokenProvider := middleware.NewCSRFUUIDProvider(uuidGenerator)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	ssmClient, err := aws.NewSSMClient()
	if err != nil {
		log.Fatalf("failed to generate ssmClient: %v", err)
	}

	params, err := services.GetSSMParameter(ssmClient, ctx)
	if err != nil {
		log.Fatalf("failed to get ssmParameters: %v", err)
	}

	env := os.Getenv("ENV")
	if env != "production" && env != "staging" {
		utils.EnvLoad(".env")
	}

	maxIdleConns := os.Getenv("DB_MAX_IDLE_CONNS")
	maxOpenConns := os.Getenv("DB_MAX_OPEN_CONNS")
	connMaxLifetime := os.Getenv("DB_CONN_MAX_LIFETIME")

	intMaxIdleConns, err := strconv.Atoi(maxIdleConns)
	if err != nil {
		log.Fatalf("failed read env:%v", err)
	}

	intMaxOpenConns, err := strconv.Atoi(maxOpenConns)
	if err != nil {
		log.Fatalf("failed read env:%v", err)
	}

	intConnMaxLifetime, err := strconv.Atoi(connMaxLifetime)
	if err != nil {
		log.Fatalf("failed read env:%v", err)
	}

	items := make(map[string]string)
	for _, param := range params.Parameters {
		items[*param.Name] = *param.Value
	}

	dbUser := items[services.PostgresUser]
	dbPassword := items[services.PostgresPassword]
	dbName := items[services.PostgresDBName]
	dbPort := items[services.PostgresDBPort]
	dbHost := items[services.PostgresDBHost]
	sslMode := items[services.PostgresSslMode]
	tz := "Asia/Tokyo"

	// 後工程で戻り値を活用する
	db, err := models.DBInit(dbUser, dbPassword, dbHost, dbPort, dbName, sslMode, tz, env, intMaxIdleConns, intMaxOpenConns, intConnMaxLifetime)
	if err != nil {
		log.Fatalf("db initialization failed: %v", err)
	}

	// 後工程で本番環境は実行しないよう修正する
	err = models.RunMigration(dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)
	// マイグレーションで変更のない場合はエラーにしない(migrate.ErrNoChange)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	passwordResetTokenRepo := repositories.NewPasswordResetTokenRepository(db)
	taskRepo := repositories.NewTaskRepository(db)

	accessTokenTTL := time.Minute * 15
	refreshTokenTTL := time.Hour * 24 * 7
	issuer := "gin-todo-api"

	jwtProvider := utils.NewDefaultJWTProvider(items[services.JwtSecretKey], issuer)
	timeProvider := utils.NewRealTimeProvider()
	passwordManager := utils.NewBcryptHasher()

	sesSDKCfg, err := aws.NewSESConfig(ctx, env)
	if err != nil {
		log.Fatalf("aws client initialization failed: %v", err)
	}
	sesClient := aws.NewMail(sesSDKCfg)

	emailFrom := os.Getenv("EMAIL_FROM")
	if emailFrom == "" {
		log.Fatal("required environment variable EMAIL_FROM is not set or empty")
	}

	passwordResetBaseURL := os.Getenv("PASSWORD_RESET_BASE_URL")
	if passwordResetBaseURL == "" {
		log.Fatal("required environment variable PASSWORD_RESET_BASE_URL is not set or empty")
	}

	emailCfg := services.NewEmailConfig(emailFrom, passwordResetBaseURL)

	adminService := services.NewAdminService(userRepo)
	emailService := services.NewEmailService(sesClient, emailCfg)
	authService := services.NewAuthService(userRepo, passwordResetTokenRepo, passwordManager, jwtProvider, accessTokenTTL, refreshTokenTTL, uuidGenerator, timeProvider, emailService)
	taskService := services.NewTaskService(taskRepo)

	adminCtrl := controllers.NewAdminControllerImpl(adminService, timeProvider)
	authCtrl := controllers.NewAuthControllerImpl(authService, timeProvider)
	taskCtrl := controllers.NewTaskControllerImpl(taskService, timeProvider)
	exportCtrl := controllers.NewExportControllerImpl(taskService, timeProvider)

	wsHub := websocket.NewHub(timeProvider, jwtProvider, gorillaWs.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")

			allowedOrigins := []string{
				"http://localhost:3000",
			}
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		},
	})
	go wsHub.Run()

	routes.SetupRoutes(r, adminCtrl, authCtrl, taskCtrl, exportCtrl, jwtProvider, wsHub, csrfTokenProvider)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// SIGINT or SIGTERM を待機して受信する
	<-ctx.Done()

	// シャットダウン猶予時間を設定
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		log.Fatal("server forced to shutdown: ", err)
	}

	// gracefulshutdownが完了した場合に表示
	log.Println("server exiting by gracefulshutdown")
}
