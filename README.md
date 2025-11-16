# Gin Todo API

タスク管理API（Go + Gin + React）のフルスタックアプリケーション

## 技術スタック

### Backend
- Go
- Gin Web Framework
- PostgreSQL
- JWT認証 (Access Token + Refresh Token)
- WebSocket (リアルタイム通知)
- Air (ホットリロード)

### Frontend
- React
- TypeScript
- Vite
- React Router
- Context API (状態管理)

## 主な機能

### 認証・認可
- ユーザー登録・ログイン・ログアウト
- JWT トークンベース認証 (Access 15分 / Refresh 7日)
- ロールベースアクセス制御 (user/admin)

### タスク管理 (CRUD)
- タスク作成・取得・更新・削除
- ステータス管理 (todo/in_progress/done)
- 優先度設定 (low/medium/high)
- ページネーション・ソート・フィルタリング
- CSV エクスポート

### 管理者機能
- 全ユーザー一覧
- 全タスク一覧

### リアルタイム通知
- WebSocket によるタスク変更通知

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
│   │   ├── websocket/      # WebSocket管理
│   │   │   ├── hub.go
│   │   │   ├── client.go
│   │   │   └── handler.go
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
- `POST /api/v1/auth/register` - ユーザー登録
- `POST /api/v1/auth/login` - ログイン
- `POST /api/v1/auth/refresh` - トークン更新
- `POST /api/v1/auth/logout` - ログアウト

### タスク (要認証)
- `GET /api/v1/tasks` - タスク一覧
- `POST /api/v1/tasks` - タスク作成
- `GET /api/v1/tasks/:id` - タスク詳細
- `PUT /api/v1/tasks/:id` - タスク更新
- `DELETE /api/v1/tasks/:id` - タスク削除
- `PATCH /api/v1/tasks/:id/status` - ステータス更新
- `GET /api/v1/tasks/export/csv` - CSV エクスポート

### 管理者 (要Admin権限)
- `GET /api/v1/admin/users` - 全ユーザー一覧
- `GET /api/v1/admin/tasks` - 全タスク一覧

### WebSocket
- `GET /api/v1/ws/notifications` - リアルタイム通知

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


## DBへの接続
```zsh
psql -U gin -h localhost -d gin-todo
```

## 実装順序
### フェーズ1: バックエンド基盤構築
- [x] 1. DB接続確認、ダミーデータの挿入
- [x] 2. モデル層（Model/Entity）の実装
  - [x] User, Task, TaskStatus, UserRole の構造体定義
- [x] 3. DB接続設定とリポジトリ基盤
  - [x] database.go（DB接続プール）
  - [ ] リポジトリのインターフェース定義
- [ ] 4. リポジトリ層（Repository/DAO）の実装
  - [ ] 基本的なCRUD操作の実装
  - [ ] ユニットテストも並行して作成
- [ ] 5. API層（Handler/Controller）の基本実装（認証なし）
  - [ ] 基本的なエンドポイント定義（GET /tasks, POST /tasks など）
  - [ ] リクエスト/レスポンス処理
  - [ ] Postmanで動作確認
- [ ] 6. 認証・認可の実装
  - [ ] JWT トークン生成/検証
  - [ ] パスワードハッシュ化
  - [ ] ミドルウェア（認証チェック、ロール権限チェック）
  - [ ] 既存のエンドポイントに認証を追加
- [ ] 7. サービス層（Business Logic）の実装
  - [ ] ビジネスロジックの実装
  - [ ] バリデーション
  - [ ] エラーハンドリング

### フェーズ2: フロントエンド実装
- [ ] フロントエンド開発
  - [ ] UI実装
  - [ ] APIとの統合

### フェーズ3: インフラ実装(並行して作る)
- [ ] EKSのpodへデプロイ

### フェーズ4: CI/CD
- [ ] GithubActions


## Swagger生成
cd backend
swag init -g cmd/api/main.go -o docs

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
