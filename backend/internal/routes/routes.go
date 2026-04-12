package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"

	"github.com/takehiro1111/gin-todo/backend/internal/controllers"
	"github.com/takehiro1111/gin-todo/backend/internal/middleware"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
	"github.com/takehiro1111/gin-todo/backend/internal/websocket"
)

// --- Huma middleware wrappers ---

func abortWithError(ctx huma.Context, status int, message string) {
	ctx.SetStatus(status)
	ctx.SetHeader("Content-Type", "application/json")
	_ = json.NewEncoder(ctx.BodyWriter()).Encode(utils.ErrorResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now(),
	})
}

func humaVerifyUser(jwtProvider utils.JWTProvider) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		jwtToken := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")
		claims, err := jwtProvider.VerifyJWT(jwtToken)
		if err != nil {
			abortWithError(ctx, http.StatusUnauthorized, "unauthorized")
			return
		}
		gc := humagin.Unwrap(ctx)
		gc.Set("user_id", claims.UserID)
		reqCtx := context.WithValue(gc.Request.Context(), middleware.UserIDKey, claims.UserID)
		gc.Request = gc.Request.WithContext(reqCtx)
		next(ctx)
	}
}

func humaVerifyRoleAdmin(jwtProvider utils.JWTProvider) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		jwtToken := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")
		claims, err := jwtProvider.VerifyJWT(jwtToken)
		if err != nil {
			abortWithError(ctx, http.StatusUnauthorized, "unauthorized")
			return
		}
		if claims.Role != "admin" {
			abortWithError(ctx, http.StatusForbidden, "forbidden")
			return
		}
		gc := humagin.Unwrap(ctx)
		gc.Set("user_id", claims.UserID)
		reqCtx := context.WithValue(gc.Request.Context(), middleware.UserIDKey, claims.UserID)
		gc.Request = gc.Request.WithContext(reqCtx)
		next(ctx)
	}
}

func humaVerifyCSRFToken(ctx huma.Context, next func(huma.Context)) {
	gc := humagin.Unwrap(ctx)
	tokenHeader := gc.Request.Header.Get("X-CSRF-Token")
	cookie, err := gc.Cookie("csrf_token")
	if err != nil {
		abortWithError(ctx, http.StatusForbidden, "failed get cookie")
		return
	}
	if tokenHeader != cookie {
		abortWithError(ctx, http.StatusForbidden, "csrf token mismatch")
		return
	}
	next(ctx)
}

// --- Route setup ---

func SetupRoutes(r *gin.Engine, adminCtrl controllers.AdminController, authCtrl controllers.AuthController, taskCtrl controllers.TaskController, exportCtrl controllers.ExportController, jwtProvider utils.JWTProvider, wsHub *websocket.Hub) {
	config := huma.DefaultConfig("Todo API", "1.0.0")
	config.Info.Description = "タスク管理APIのフルスタックアプリケーション"
	config.Info.Contact = &huma.Contact{
		Name:  "API Support",
		Email: "support@example.com",
	}
	config.Info.License = &huma.License{
		Name: "MIT",
		URL:  "https://opensource.org/licenses/MIT",
	}
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"BearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "Type Bearer followed by a space and JWT token.",
		},
	}

	api := humagin.New(r, config)

	bearerSecurity := []map[string][]string{{"BearerAuth": {}}}
	mwAuth := huma.Middlewares{humaVerifyUser(jwtProvider)}
	mwAdmin := huma.Middlewares{humaVerifyRoleAdmin(jwtProvider)}
	mwAuthCSRF := huma.Middlewares{humaVerifyUser(jwtProvider), humaVerifyCSRFToken}

	// =====================
	// Health
	// =====================
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/api/health",
		Summary:     "ヘルスチェック",
		Description: "APIサーバーの稼働状態を確認",
		Tags:        []string{"health"},
	}, func(ctx context.Context, input *struct{}) (*struct{ Body map[string]string }, error) {
		return &struct{ Body map[string]string }{Body: map[string]string{"status": "ok"}}, nil
	})

	// =====================
	// Admin (管理者権限)
	// =====================
	huma.Register(api, huma.Operation{
		OperationID: "admin-get-all-users",
		Method:      http.MethodGet,
		Path:        "/api/admin/users",
		Summary:     "全ユーザー一覧取得",
		Description: "管理者権限で全ユーザーの一覧を取得する",
		Tags:        []string{"admin"},
		Security:    bearerSecurity,
		Middlewares: mwAdmin,
	}, adminCtrl.GetAllUsers)

	huma.Register(api, huma.Operation{
		OperationID: "admin-get-all-users-with-tasks",
		Method:      http.MethodGet,
		Path:        "/api/admin/tasks",
		Summary:     "全ユーザー（タスク付き）一覧取得",
		Description: "管理者権限で全ユーザーとそのタスク一覧を取得する",
		Tags:        []string{"admin"},
		Security:    bearerSecurity,
		Middlewares: mwAdmin,
	}, adminCtrl.GetAllUsersWithTasks)

	// =====================
	// Auth (公開)
	// =====================
	huma.Register(api, huma.Operation{
		OperationID: "auth-register",
		Method:      http.MethodPost,
		Path:        "/api/auth/register",
		Summary:     "ユーザー登録",
		Description: "新規ユーザーを登録する",
		Tags:        []string{"auth"},
	}, authCtrl.Register)

	huma.Register(api, huma.Operation{
		OperationID: "auth-login",
		Method:      http.MethodPost,
		Path:        "/api/auth/login",
		Summary:     "ログイン",
		Description: "メールアドレスとパスワードでログインする",
		Tags:        []string{"auth"},
	}, authCtrl.Login)

	huma.Register(api, huma.Operation{
		OperationID: "auth-forgot-password",
		Method:      http.MethodPost,
		Path:        "/api/auth/forgot",
		Summary:     "パスワードリセットのトークン生成",
		Description: "パスワードリセットに使用するトークンの生成",
		Tags:        []string{"auth"},
	}, authCtrl.ForgotPassword)

	huma.Register(api, huma.Operation{
		OperationID: "auth-reset-password",
		Method:      http.MethodPost,
		Path:        "/api/auth/reset",
		Summary:     "パスワードリセット",
		Description: "トークン検証を行ってパスワードリセットを実行",
		Tags:        []string{"auth"},
	}, authCtrl.ResetPassword)

	huma.Register(api, huma.Operation{
		OperationID: "auth-refresh-token",
		Method:      http.MethodPost,
		Path:        "/api/auth/refresh",
		Summary:     "トークンリフレッシュ",
		Description: "リフレッシュトークンを使用して新しいアクセストークンを取得する",
		Tags:        []string{"auth"},
	}, authCtrl.RefreshToken)

	// =====================
	// Auth (認証必須)
	// =====================
	huma.Register(api, huma.Operation{
		OperationID: "auth-logout",
		Method:      http.MethodPost,
		Path:        "/api/auth/logout",
		Summary:     "ログアウト",
		Description: "ログアウトする",
		Tags:        []string{"auth"},
		Security:    bearerSecurity,
		Middlewares: mwAuth,
	}, authCtrl.Logout)

	huma.Register(api, huma.Operation{
		OperationID: "auth-get-me",
		Method:      http.MethodGet,
		Path:        "/api/auth/me",
		Summary:     "ユーザー情報取得",
		Description: "ログイン中のユーザー情報を取得する",
		Tags:        []string{"auth"},
		Security:    bearerSecurity,
		Middlewares: mwAuth,
	}, authCtrl.GetMe)

	huma.Register(api, huma.Operation{
		OperationID: "auth-change-password",
		Method:      http.MethodPatch,
		Path:        "/api/auth/password",
		Summary:     "パスワード変更",
		Description: "ログイン中のユーザーのパスワードを変更する",
		Tags:        []string{"auth"},
		Security:    bearerSecurity,
		Middlewares: mwAuth,
	}, authCtrl.ChangePassword)

	huma.Register(api, huma.Operation{
		OperationID: "auth-csrf-token",
		Method:      http.MethodGet,
		Path:        "/api/auth/csrf-token",
		Summary:     "CSRFトークン取得",
		Description: "CSRFトークンを生成し、Cookieとレスポンスで返す",
		Tags:        []string{"auth"},
		Security:    bearerSecurity,
		Middlewares: mwAuth,
	}, authCtrl.GenerateCSRFToken)

	// =====================
	// Task (認証必須 / GET は CSRF 不要)
	// =====================
	huma.Register(api, huma.Operation{
		OperationID: "task-get-all",
		Method:      http.MethodGet,
		Path:        "/api/task/",
		Summary:     "タスク一覧取得",
		Description: "ログイン中のユーザーのタスク一覧を取得する",
		Tags:        []string{"task"},
		Security:    bearerSecurity,
		Middlewares: mwAuth,
	}, taskCtrl.GetTasks)

	huma.Register(api, huma.Operation{
		OperationID: "task-get-by-id",
		Method:      http.MethodGet,
		Path:        "/api/task/{id}",
		Summary:     "タスク取得",
		Description: "指定されたIDのタスクを取得する",
		Tags:        []string{"task"},
		Security:    bearerSecurity,
		Middlewares: mwAuth,
	}, taskCtrl.GetTaskByID)

	huma.Register(api, huma.Operation{
		OperationID: "task-create",
		Method:      http.MethodPost,
		Path:        "/api/task/",
		Summary:     "タスク作成",
		Description: "新しいタスクを作成する",
		Tags:        []string{"task"},
		Security:    bearerSecurity,
		Middlewares: mwAuthCSRF,
	}, taskCtrl.CreateTask)

	huma.Register(api, huma.Operation{
		OperationID: "task-update",
		Method:      http.MethodPut,
		Path:        "/api/task/{id}",
		Summary:     "タスク更新",
		Description: "指定されたIDのタスクを更新する",
		Tags:        []string{"task"},
		Security:    bearerSecurity,
		Middlewares: mwAuthCSRF,
	}, taskCtrl.UpdateTask)

	huma.Register(api, huma.Operation{
		OperationID: "task-update-status",
		Method:      http.MethodPatch,
		Path:        "/api/task/{id}",
		Summary:     "タスクステータス更新",
		Description: "指定されたIDのタスクのステータスを更新する",
		Tags:        []string{"task"},
		Security:    bearerSecurity,
		Middlewares: mwAuthCSRF,
	}, taskCtrl.UpdateTaskStatus)

	huma.Register(api, huma.Operation{
		OperationID: "task-delete",
		Method:      http.MethodDelete,
		Path:        "/api/task/{id}",
		Summary:     "タスク削除",
		Description: "指定されたIDのタスクを削除する",
		Tags:        []string{"task"},
		Security:    bearerSecurity,
		Middlewares: mwAuthCSRF,
	}, taskCtrl.DeleteTask)

	// =====================
	// Export (認証必須 / Gin ハンドラーのまま: バイナリストリーミング)
	// =====================
	apiTasks := r.Group("/api/tasks")
	apiTasks.Use(middleware.VerifyUser(jwtProvider))
	{
		apiTasks.GET("/export/csv", exportCtrl.ExportTasksCSV)
	}

	// =====================
	// WebSocket (Gin ハンドラーのまま)
	// =====================
	ws := r.Group("/api/ws")
	{
		ws.GET("/chat", wsHub.ChatServer)
	}

	// =====================
	// Fallback
	// =====================
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Success:   false,
			Message:   "endpoint not found",
			Timestamp: time.Now(),
		})
	})

	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, utils.ErrorResponse{
			Success:   false,
			Message:   "method not allowed",
			Timestamp: time.Now(),
		})
	})
}
