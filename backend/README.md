
```zsh
# 1. Go モジュール初期化
go mod init github.com/take/gin-todo

# 2. 基本パッケージ
go get -u github.com/gin-gonic/gin \
  gorm.io/gorm \
  gorm.io/driver/postgres \
  github.com/golang-jwt/jwt/v5 \
  github.com/joho/godotenv \
  github.com/go-playground/validator/v10 \
  github.com/gorilla/websocket \
  github.com/swaggo/swag \
  github.com/swaggo/gin-swagger \
  github.com/swaggo/files

# 3. Swagger
go get -u github.com/swaggo/swag \
  github.com/swaggo/gin-swagger \
  github.com/swaggo/files

# 4. swagコマンドインストール
go install github.com/swaggo/swag/cmd/swag@latest

# 5. 依存関係整理
go mod tidy
```

## 主な機能

### 認証・認可
- ユーザー登録・ログイン・ログアウト
- JWT トークンベース認証 (Access 15分 / Refresh 7日)
- ロールベースアクセス制御 (user/admin)

#### バリデーションルール
| 項目 | ルール |
|------|--------|
| メールアドレス | `xxx@xxx.xxx` 形式（@と.を含む） |
| パスワード | 8〜24文字、英大文字(A-Z)・英小文字(a-z)・数字(0-9)・記号(.?/!-)のみ使用可 |

### タスク管理 (CRUD)
- タスク作成・取得・更新・削除
- ステータス管理 (todo/in_progress/done)
- 優先度設定 (low/medium/high)
- ページネーション・ソート・フィルタリング
- CSV エクスポート

### 管理者機能
- 全ユーザー一覧
- 全タスク一覧

### WebSocket チャット
- JWT認証は First Message方式で最初のメッセージからtokenを取得して認証する。
- Hub パターンによるメッセージブロードキャスト
- Ping/Pong による接続死活監視
- オリジン検証（CheckOrigin）

### 非機能要件
- リクエストログ (リクエストID付き)
- CORS 対応
- CSRF 保護
- レート制限 (100req/min per IP)
- GZIP 圧縮
- セキュリティヘッダー
- Graceful Shutdown
- ヘルスチェック

## ディレクトリ構成
### ディレクトリ構成イメージ
```
gin-todo/
├── .git/
├── .github/
│   └── workflows/          # CI/CD設定
├── .gitignore
├── README.md
│
├── backend/                # Go (Gin) API
│   ├── cmd/
│   │   └── api/
│   │       └── main.go     # エントリーポイント
│   │
│   ├── internal/           # プライベートコード
│   │   ├── config/         # 設定管理
│   │   │   └── config.go
│   │   │
│   │   ├── models/         # データモデル (Model)
│   │   │   ├── user.go
│   │   │   ├── task.go
│   │   │   └── db.go
│   │   │
│   │   ├── controllers/    # ハンドラー (Controller)
│   │   │   ├── auth_controller.go
│   │   │   ├── task_controller.go
│   │   │   ├── admin_controller.go
│   │   │   └── export_controller.go
│   │   │
│   │   ├── services/       # ビジネスロジック
│   │   │   ├── auth_service.go
│   │   │   ├── task_service.go
│   │   │   ├── user_service.go
│   │   │   └── notification_service.go
│   │   │
│   │   ├── repositories/   # データアクセス層
│   │   │   ├── user_repository.go
│   │   │   └── task_repository.go
│   │   │
│   │   ├── middleware/     # ミドルウェア
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── csrf.go
│   │   │   ├── rate_limit.go
│   │   │   ├── logger.go
│   │   │   ├── recovery.go
│   │   │   ├── security_headers.go
│   │   │   └── timeout.go
│   │   │
│   │   ├── routes/         # ルーティング設定
│   │   │   └── routes.go
│   │   │
│   │   ├── validators/     # バリデーション
│   │   │   ├── user_validator.go
│   │   │   └── task_validator.go
│   │   │
│   │   ├── websocket/      # WebSocketチャット
│   │   │   ├── chat.go     # Hub構造体 + ChatServer + readAndBroadcast + Run
│   │   │   └── client.go   # Client構造体 + writePump + 定数定義
│   │   │
│   │   └── utils/          # ユーティリティ
│   │       ├── jwt.go
│   │       ├── response.go
│   │       └── password.go
│   │
│   ├── migrations/         # DBマイグレーション
│   │   ├── 001_create_users.sql
│   │   └── 002_create_tasks.sql
│   │
│   ├── tests/              # テストコード
│   │   ├── integration/
│   │   └── unit/
│   │
│   ├── .air.toml           # ホットリロード設定
│   ├── .env.example
│   ├── .env
│   ├── go.mod
│   └── go.sum
│
└── frontend/               # React SPA
    ├── public/
    │   └── index.html
    │
    ├── src/
    │   ├── api/            # APIクライアント
    │   │   ├── auth.ts
    │   │   ├── tasks.ts
    │   │   └── client.ts
    │   │
    │   ├── components/     # 再利用可能なコンポーネント (View)
    │   │   ├── common/
    │   │   │   ├── Button.tsx
    │   │   │   ├── Input.tsx
    │   │   │   └── Modal.tsx
    │   │   ├── tasks/
    │   │   │   ├── TaskList.tsx
    │   │   │   ├── TaskItem.tsx
    │   │   │   └── TaskForm.tsx
    │   │   └── auth/
    │   │       ├── LoginForm.tsx
    │   │       └── RegisterForm.tsx
    │   │
    │   ├── pages/          # ページコンポーネント
    │   │   ├── Login.tsx
    │   │   ├── Register.tsx
    │   │   ├── TaskList.tsx
    │   │   ├── TaskDetail.tsx
    │   │   └── AdminDashboard.tsx
    │   │
    │   ├── hooks/          # カスタムフック
    │   │   ├── useAuth.ts
    │   │   ├── useTasks.ts
    │   │   └── useWebSocket.ts
    │   │
    │   ├── context/        # 状態管理 (Model)
    │   │   ├── AuthContext.tsx
    │   │   └── TaskContext.tsx
    │   │
    │   ├── types/          # TypeScript型定義
    │   │   ├── user.ts
    │   │   └── task.ts
    │   │
    │   ├── utils/          # ユーティリティ
    │   │   ├── formatDate.ts
    │   │   └── validation.ts
    │   │
    │   ├── routes/         # ルーティング
    │   │   └── AppRoutes.tsx
    │   │
    │   ├── App.tsx
    │   ├── index.tsx
    │   └── index.css
    │
    ├── .env.example
    ├── .env
    ├── package.json
    ├── tsconfig.json
    └── vite.config.ts      # or webpack.config.js
```

## セットアップ

### 前提条件
- Go
- Node.js
- PostgreSQL
- Air (ホットリロード用)

### Backend セットアップ

```bash
cd backend

# 依存関係インストール
go mod download

# Air インストール (未インストールの場合)
go install github.com/cosmtrek/air@latest

# 環境変数設定
cp .env.example .env
# .env を編集してDB接続情報やJWT秘密鍵を設定

# 開発サーバー起動 (ホットリロード)
air

# または通常起動
go run cmd/api/main.go

# 本番で実行する場合
GIN_MODE=release go run cmd/api/main.go
```

### Frontend セットアップ

```bash
cd frontend

# 依存関係インストール
npm install

# 環境変数設定
cp .env.example .env
# .env を編集してAPI URLを設定

# 開発サーバー起動
npm run dev

# ビルド
npm run build
```

## API エンドポイント

### 認証
- `POST /api/auth/register` - ユーザー登録
- `POST /api/auth/login` - ログイン
- `POST /api/auth/refresh` - トークン更新
- `POST /api/auth/logout` - ログアウト
- `GET /api/auth/me` - ログインユーザー情報取得（要認証）
- `PATCH /api/auth/password` - パスワード変更（要認証）

### タスク (要認証)
- `GET /api/tasks` - タスク一覧
- `POST /api/tasks` - タスク作成
- `GET /api/tasks/:id` - タスク詳細
- `PUT /api/tasks/:id` - タスク更新
- `DELETE /api/tasks/:id` - タスク削除
- `PATCH /api/tasks/:id/status` - ステータス更新
- `GET /api/tasks/export/csv` - CSV エクスポート

### 管理者 (要Admin権限)
- `GET /api/admin/users` - 全ユーザー一覧
- `GET /api/admin/tasks` - 全タスク一覧

### WebSocket (JWT認証: First Message方式)
- `GET /api/ws/chat` - リアルタイムチャット（接続後、最初のメッセージでJWTトークンを送信して認証）

> [!NOTE]
> リリース後にPhase2以降でメッセージのDB保存や履歴取得機能を実装したい。

### その他
- `GET /health` - ヘルスチェック
- `GET /metrics` - メトリクス (プロキシ)

## 開発ワークフロー

### pre-commit フック
```bash
pre-commit install
```

### テスト実行
```bash
# Backend
cd backend
go test ./...

# Frontend
cd frontend
npm test
```

### コードフォーマット
```bash
# Backend
gofmt -w .

# Frontend
npm run lint
npm run format
```

## データモデル

### User
- `id` (bigserial, PK)
- `email` (varchar, unique)
- `password_hash` (varchar)
- `name` (varchar)
- `role` (varchar: admin/writer/viewer)
- `created_at` (timestamp)
- `updated_at` (timestamp)

### Task
- `id` (bigserial, PK)
- `user_id` (bigint, FK)
- `title` (varchar)
- `description` (text)
- `status` (varchar: todo/in_progress/done)
- `priority` (varchar: low/medium/high)
- `due_date` (timestamp)
- `created_at` (timestamp)
- `updated_at` (timestamp)

## Reference
https://gin-gonic.com/ja/docs/

# 検証コマンド
```zsh
# ユーザー登録
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"test-taro","email":"test@gmail.com","password":"password123"}'

# ログイン
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@gmail.com","password":"password123"}'

# ユーザー一覧の取得
curl -X GET http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer <admin_access_token>"

# ユーザーごとのタスク一覧の取得
curl -X GET http://localhost:8080/api/admin/tasks \
  -H "Authorization: Bearer <admin_access_token>"

# リセットトークンの発行
curl -X POST http://localhost:8080/api/auth/forgot \
  -H "Content-Type: application/json" \
  -d '{"email":"test@gmail.com"}'

# パスワードをリセット
curl -X POST http://localhost:8080/api/auth/reset \
  -H "Content-Type: application/json" \
  -d '{"reset_token":"xxx", "new_password": "test4583"}'

# CSVエクスポート
curl -v -o export.csv \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  http://localhost:8080/api/tasks/export/csv

# WebSocketチャット接続（wscat使用）
# 接続後、最初のメッセージでJWTトークンを送信して認証する
wscat -c "ws://localhost:8080/api/ws/chat" --origin http://localhost:3000
```


## DBへの接続
```zsh
psql -U gin -h localhost -d gin-todo
```

## 実装順序
- 業務時間外で対応するため、挫折や塩漬けで手をつけなくなることを極力回避するため予め細かい粒度で実装順序を記述している。
  - ファイル名、内容、順序については開発しながら必要に応じて都度修正する。
### フェーズ1: バックエンド基盤構築
#### 1. DB接続確認、ダミーデータの挿入
- [x] PostgreSQL起動確認
- [x] psqlでDB接続確認
- [x] テーブル作成（users, tasks）
- [x] ダミーデータINSERT

#### 2. モデル層（Model/Entity）の実装
- [x] `internal/models/user.go` - User構造体定義
- [x] `internal/models/user_role.go` - UserRole型定義
- [x] `internal/models/task.go` - Task構造体定義
- [x] `internal/models/task_status.go` - TaskStatus型定義

#### 3. DB接続設定
- [x] `internal/models/database.go` - InitDB()関数実装（GORM接続）
- [x] `internal/models/database.go` - AutoMigrate実行

#### 4. リポジトリ層
- [x] `internal/repositories/user_repository.go` - UserRepository（インターフェース + 実装）
- [x] `internal/repositories/task_repository.go` - TaskRepository（インターフェース + 実装）
- [x] `internal/repositories/user_role_repository.go` - UserRoleRepository（インターフェース + 実装）
- [x] `internal/repositories/task_status_repository.go` - TaskStatusRepository（インターフェース + 実装）

---

#### 5. サービス層の実装

##### 5-1. TaskService（ビジネスロジック）
- [x] `internal/services/task_service.go` - TaskService 構造体定義
- [x] `internal/services/task_service.go` - CreateTask() メソッド実装
- [x] `internal/services/task_service.go` - GetTaskByID() メソッド実装（権限チェック含む）
- [x] `internal/services/task_service.go` - GetTasksByUserID() メソッド実装
- [x] `internal/services/task_service.go` - UpdateTask() メソッド実装（権限チェック含む）
- [x] `internal/services/task_service.go` - UpdateTaskStatus() メソッド実装
- [x] `internal/services/task_service.go` - DeleteTask() メソッド実装（権限チェック含む）

---

#### 6. 認証基盤の実装

##### 6-1. パスワードハッシュ化
- [x] `internal/utils/password.go` - HashPassword() 関数実装
- [x] `internal/utils/password.go` - CheckPassword() 関数実装

##### 6-2. JWT トークン生成/検証
- [x] `internal/utils/jwt.go` - Claims 構造体定義
- [x] `internal/utils/jwt.go` - GenerateAccessToken() 実装
- [x] `internal/utils/jwt.go` - GenerateRefreshToken() 実装
- [x] `internal/utils/jwt.go` - ValidateToken() 実装

##### 6-3. 認証ミドルウェア
- ~~ [ ] `internal/middleware/auth.go` - AuthMiddleware() 実装（JWT検証）~~
  - `6-2. JWT トークン生成/検証`と処理が重複するため。
- [x] `internal/middleware/auth.go` - RoleMiddleware() 実装（Admin権限チェック）

##### 6-4. AuthService（認証ロジック）
- [x] `internal/services/auth_service.go` - AuthService 構造体定義
- [x] `internal/services/auth_service.go` - Register() メソッド実装
- [x] `internal/services/auth_service.go` - Login() メソッド実装
- [x] `internal/services/auth_service.go` - RefreshToken() メソッド実装
- ~~ [ ] `internal/services/auth_service.go` - Logout() メソッド実装 ~~
  - JWTはステートレスのためサーバーでトークンを削除するのではなく、フロント側でローカルストレージやCookieから削除する
- [x] `internal/services/auth_service.go` - GetMe() メソッド実装（ユーザー情報取得）
- [x] `internal/services/auth_service.go` - ChangePassword() メソッド実装（現パスワード照合、新パスワードハッシュ化）

> [!NOTE]
> - ユーザーのCRUD操作は主に認証か管理者機能に紐づく
>   - 管理者ではない限り、他ユーザーを操作する必要はないため。

---

#### 7. コントローラー層（Handler）の実装

##### 7-1. レスポンスヘルパー
- [x] `internal/utils/response.go` - SuccessResponse() 関数実装
- [x] `internal/utils/response.go` - ErrorResponse() 関数実装
- [x] `internal/utils/response.go` - ValidationErrorResponse() 関数実装

##### 7-2. 認証コントローラー
- [x] `internal/controllers/auth_controller.go` - AuthController 構造体定義
- [x] `internal/controllers/auth_controller.go` - NewAuthController() コンストラクタ
- [x] `internal/controllers/auth_controller.go` - Register() ハンドラ実装
- [x] `internal/controllers/auth_controller.go` - Login() ハンドラ実装（Cookie設定含む）
- [x] `internal/controllers/auth_controller.go` - RefreshToken() ハンドラ実装
- [x] `internal/controllers/auth_controller.go` - Logout() ハンドラ実装（Cookie削除）
- [x] `internal/controllers/auth_controller.go` - GetMe() ハンドラ実装（JWTからuser_id取得、Service呼び出し）
- [x] `internal/controllers/auth_controller.go` - ChangePassword() ハンドラ実装（current_password, new_passwordのバインド）
  - ログイン中のユーザーが現在のパスワードを知っている状態で変更

> [!IMPORTANT]
> Logoutの処理はひとまずJWTで実装不要だが、後からCookieの処理を実装する際にロジックを付け足す。

##### 7-3. タスクコントローラー
- [x] `internal/controllers/task_controller.go` - TaskController 構造体定義
- [x] `internal/controllers/task_controller.go` - NewTaskController() コンストラクタ
- [x] `internal/controllers/task_controller.go` - GetTasks() ハンドラ実装（クエリパラメータ処理）
- [x] `internal/controllers/task_controller.go` - CreateTask() ハンドラ実装
- [x] `internal/controllers/task_controller.go` - GetTaskByID() ハンドラ実装
- [x] `internal/controllers/task_controller.go` - UpdateTask() ハンドラ実装
- [x] `internal/controllers/task_controller.go` - UpdateTaskStatus() ハンドラ実装
- [x] `internal/controllers/task_controller.go` - DeleteTask() ハンドラ実装

---

#### 8. ルーティング設定

- [x] `internal/routes/routes.go` - SetupRoutes() 関数実装
- [x] `internal/routes/routes.go` - 認証なしエンドポイント設定（/auth/register, /auth/login）
- [x] `internal/routes/routes.go` - 認証必須エンドポイント設定（/tasks/*）
- [x] `internal/routes/routes.go` - ヘルスチェック（/health）設定
- [x] `internal/routes/routes.go` - GET /auth/me 設定（要認証）
- [x] `internal/routes/routes.go` - PATCH /auth/password 設定（要認証）

---

#### 9. main.go の実装

- [x] `cmd/api/main.go` - 環境変数読み込み
- [x] `cmd/api/main.go` - DB接続初期化
- [x] `cmd/api/main.go` - Repository 初期化
- [x] `cmd/api/main.go` - Service 初期化
- [x] `cmd/api/main.go` - Controller 初期化
- [x] `cmd/api/main.go` - Router 設定
- [x] `cmd/api/main.go` - Graceful Shutdown 実装
- [x] `cmd/api/main.go` - サーバー起動

---

#### 10. 動作確認（Swagger）

- [x] POST /api/auth/register - ユーザー登録テスト
- [x] POST /api/auth/login - ログインテスト（Cookie確認）
- [x] POST /api/tasks - タスク作成テスト（要認証）
- [x] GET /api/tasks - タスク一覧テスト
- [x] GET /api/tasks/:id - タスク詳細テスト
- [x] PUT /api/tasks/:id - タスク更新テスト
- [x] PATCH /api/tasks/:id/status - ステータス更新テスト
- [x] DELETE /api/tasks/:id - タスク削除テスト

---

#### 11. バリデーション実装

##### 11-1. リクエスト構造体（コントローラー内に定義済み）
- [x] `internal/controllers/auth_controller.go` - RegisterRequest, LoginRequest, ChangePasswordRequest, RefreshTokenRequest
- [x] `internal/controllers/task_controller.go` - CreateRequest, UpdateRequest, UpdateTaskStatusRequest

##### 11-2. カスタムバリデーション
- [x] `internal/validators/auth.go` - AuthValidator 実装（Functional Options パターン）
- [x] `internal/validators/auth.go` - emailCustom バリデーション（`.+@.+\..+` 形式）
- [x] `internal/validators/auth.go` - passwordCustom バリデーション（8〜24文字、英数字+記号`.?/!-`）
- [x] `internal/validators/auth_test.go` - バリデーションテスト

##### 11-3. 細かいバリデーション
- [x] bindingタグでバリデーションのチェック
---

#### 12. 管理者機能実装

- [x] `internal/services/admin_service.go` - AdminService 実装
- [x] `internal/controllers/admin_controller.go` - GetAllUsers() 実装
- [x] `internal/controllers/admin_controller.go` - GetAllTasks() 実装
- [x]  `internal/routes/routes.go` - ルーティング追加 - GET /api/admin/users
- [x]  `internal/routes/routes.go` - ルーティング追加 - GET /api/admin/tasks

---

#### 13.パスワード再設定処理
- [x] `internal/services/auth_service.go` - ForgotPassword() メソッド実装（リセットトークン生成、トークンのDB保存）
- [x] `internal/services/auth_service.go` - ResetPassword() メソッド実装（トークン検証、パスワード更新）
- [x] `internal/controllers/auth_controller.go` - ForgotPassword() ハンドラ実装（emailのバインド）
- [x] `internal/controllers/auth_controller.go` - ResetPassword() ハンドラ実装（token, new_passwordのバインド）
- [x] `internal/routes/routes.go` - POST /auth/forgot 設定（認証不要）
- [x] `internal/routes/routes.go` - POST /auth/reset 設定（認証不要）
- [x] `internal/controllers/email_service.go` - SendPasswordResetEmail()メソッド実装(メール送信処理)

---

#### 14. エクスポート機能実装

- [x] `internal/controllers/export_controller.go` - ExportController 実装
- [x] `internal/controllers/export_controller.go` - ExportTasksCSV() 実装（ストリーミング）
- [x] ルーティング追加 - GET /api/tasks/export/csv

---

#### 15. WebSocketチャット実装

- [x] `internal/websocket/chat.go` - Hub 構造体定義（クライアント管理）
- [x] `internal/websocket/chat.go` - NewHub() コンストラクタ（timeProvider, jwtProvider, upgrader のDI）
- [x] `internal/websocket/chat.go` - Run() メソッド実装（register/unregister/broadcast イベントループ）
- [x] `internal/websocket/chat.go` - ChatServer() ハンドラ実装（JWT認証 + WebSocketアップグレード）
- [x] `internal/websocket/chat.go` - readAndBroadcast() 実装（メッセージ読み取り + ユーザー名付きJSON配信）
- [x] `internal/websocket/chat.go` - CheckOrigin によるオリジン検証
- [x] `internal/websocket/client.go` - Client 構造体定義
- [x] `internal/websocket/client.go` - writePump() 実装（メッセージ送信 + Ping/Pong 死活監視）
- [x] `internal/routes/routes.go` - ルーティング追加 - GET /api/ws/chat
- [x] `cmd/api/main.go` - Hub初期化 + `go wsHub.Run()` 起動
---

#### 16. 非機能要件（ミドルウェア）実装

##### 15-1. ロギング
- [x] `internal/middleware/logger.go` - RequestLogger() 実装（リクエストID生成）

##### 15-2. CORS
- [x] `internal/middleware/cors.go` - CORS() 実装

##### 15-3. CSRF
- [x] `internal/middleware/csrf.go` - CSRF() 実装

##### 15-4. Rate Limiting
- [ ] `internal/middleware/rate_limit.go` - RateLimit() 実装（IP単位）

##### 15-5. セキュリティヘッダー
- [ ] `internal/middleware/security_headers.go` - SecurityHeaders() 実装

##### 15-6. GZIP圧縮
- [ ] `internal/middleware/gzip.go` - GZIP() 実装（またはgin-gzip使用）

##### 15-7. Recovery
- [x] `internal/middleware/recovery.go` - Recovery() 実装

##### 15-8. タイムアウト
- [x] `internal/middleware/timeout.go` - Timeout() 実装

---

#### 16. エラーハンドリング強化
- [ ] 403ハンドラ設定、改善（他ユーザーの権限エラー時など）
- [ ] 404ハンドラ設定
- [ ] 405ハンドラ設定
- [ ] 統一エラーレスポンス形式確認

---

#### 17.書けてないテスト実装
- Unit
- API Integration

# Swagger参考
https://github.com/swaggo/swag#mime-types

## Swagger生成
```zsh
cd backend
swag init -g cmd/api/main.go -o docs

# JSONのみ
swag init --outputTypes json

# YAMLのみ
swag init --outputTypes yaml
```

## migrationで以下のエラーになった場合
```zsh

2025/11/08 19:38:26 db initialization failed: migration failed: Dirty database version 2. Fix and force version.
exit status 1

# 以下コマンドを実行
# dirtyフラグを強制的にfalseに書き換えている。
# migration実行時にdrop tableしているのでこれでも良い。
# そもそも本番環境はmigration使わない方が良い。
$psql -U gin -d gin-todo -h localhost -p 5432 -c "UPDATE schema_migrations SET dirty = false;"
Password for user gin:
UPDATE 1
```
## ローカルでDBへの接続
```zsh
docker compose exec gin-todo-postgres psql -U gin -d gin-todo
```

## Swaggerのエンドポイント
http://localhost:8080/swagger/index.html


## API エンドポイント

### ヘルスチェック

```bash
curl -X GET http://localhost:8080/api/health
```

### 認証 (Auth)

#### ユーザー登録

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "test-user",
    "email": "test@example.com",
    "password": "Password123!"
  }'
```

#### ログイン

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "test@example.com",
    "password": "Password123!"
  }'
```

#### トークンリフレッシュ

```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -d '{
    "refresh_token": "<REFRESH_TOKEN>"
  }'
```

#### ログアウト

```bash
curl -X POST http://localhost:8080/api/auth/logout \
  -H 'Authorization: Bearer <ACCESS_TOKEN>'
```

#### ユーザー情報取得

```bash
curl -X GET http://localhost:8080/api/auth/me \
  -H 'Authorization: Bearer <ACCESS_TOKEN>'
```

#### パスワード変更

```bash
curl -X PATCH http://localhost:8080/api/auth/password \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -d '{
    "old_password": "Password123!",
    "new_password": "NewPassword456!"
  }'
```

### タスク (Task)

> 全てのタスクAPIは認証が必要です。`Authorization: Bearer <ACCESS_TOKEN>` ヘッダーを付与してください。

#### タスク一覧取得

```bash
curl -X GET http://localhost:8080/api/task/ \
  -H 'Authorization: Bearer <ACCESS_TOKEN>'
```

#### タスク作成

```bash
curl -X POST http://localhost:8080/api/task/ \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -d '{
    "title": "新しいタスク",
    "description": "タスクの説明",
    "status_id": 1,
    "priority": "high",
    "due_date": "2026-01-31T23:59:59+09:00"
  }'
```

#### タスク取得（ID指定）

```bash
curl -X GET http://localhost:8080/api/task/1 \
  -H 'Authorization: Bearer <ACCESS_TOKEN>'
```

#### タスク更新

```bash
curl -X PUT http://localhost:8080/api/task/1 \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -d '{
    "title": "更新後のタスク",
    "description": "更新後の説明",
    "status_id": 2,
    "priority": "medium",
    "due_date": "2026-02-15T23:59:59+09:00"
  }'
```

#### タスクステータス更新

```bash
curl -X PATCH http://localhost:8080/api/task/1 \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -d '{
    "status_id": 3
  }'
```

#### タスク削除

```bash
curl -X DELETE http://localhost:8080/api/task/1 \
  -H 'Authorization: Bearer <ACCESS_TOKEN>'
```
