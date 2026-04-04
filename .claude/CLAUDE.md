# Gin Todo - Full Stack Development Guide

## Project Overview

Go + Gin バックエンド と React フロントエンド によるタスク管理アプリケーション。

## Repository Structure

```
gin-todo/
├── backend/          # Go + Gin API server (port 8080)
├── frontend/         # React + TypeScript SPA (port 3000)
├── infra/cdk/        # AWS CDK (TypeScript) — SSM Parameter Store 等
└── compose.yaml      # ローカル開発: PostgreSQL, Adminer, SES local
```

## Commands

### ローカル環境の起動

```bash
docker compose up -d            # PostgreSQL (5432), Adminer (9090), SES local (8005)
```

### Backend (`backend/` で実行)

```bash
go run cmd/api/main.go          # サーバー起動
air                             # ホットリロード付き起動 (.air.toml 使用)
make swagger.init               # Swagger ドキュメント再生成 (swaggo)
go test ./...                   # 全テスト実行
go test ./internal/services/... # パッケージ指定テスト
go vet ./...                    # 静的解析
```

### Frontend (`frontend/` で実行)

```bash
pnpm dev                        # 開発サーバー (http://localhost:3000)
pnpm build                      # プロダクションビルド
pnpm biome check src/           # Lint + Format チェック
pnpm biome check --write src/   # 自動修正
pnpm typecheck                  # TypeScript 型チェック
pnpm test --run                 # テスト (1回実行)
```

### Infrastructure (`infra/cdk/` で実行)

```bash
npx cdk synth                   # CloudFormation テンプレート生成
npx cdk deploy                  # デプロイ (ap-northeast-1)
```

---

## Tech Stack

### Backend
- Go + Gin Web Framework
- PostgreSQL + GORM (ORM) + golang-migrate (マイグレーション)
- JWT 認証 (Access Token 15分 / Refresh Token 7日)
- CSRF ダブルサブミット Cookie パターン
- WebSocket (Gorilla WebSocket — リアルタイムチャット)
- AWS SDK v2 (SSM Parameter Store / SES v2)
- Swagger / OpenAPI (swaggo)
- Air (ホットリロード)
- golang.org/x/time/rate (レートリミット)

### Frontend
- React 18 + TypeScript + Vite
- React Router v6 / Context API (状態管理)
- Tailwind CSS + clsx (スタイリング)
- axios + TanStack Query (API通信・キャッシュ)
- react-hook-form + zod (フォーム・バリデーション)
- Biome (Linter + Formatter) / husky + lint-staged
- Vitest + Testing Library (テスト)

### Infrastructure
- AWS CDK (TypeScript) — SSM Parameter Store でシークレット管理
- デプロイ先: ap-northeast-1 (東京)

---

## Backend Architecture

### レイヤー構成

```
cmd/api/main.go  (エントリポイント・DI 配線)
  ↓
routes/          → ルート登録・ミドルウェアのグループ化
controllers/     → HTTP ハンドラー (リクエストバインド・レスポンス返却)
services/        → ビジネスロジック
repositories/    → GORM によるデータアクセス
models/          → GORM モデル定義・DB 初期化・マイグレーション
```

### DI パターン

すべての主要コンポーネント (Repository, Service, Controller, JWTProvider 等) は **interface で定義** し、ファクトリ関数で注入する。`main.go` で配線される。

```
Repositories → Services → Controllers → Routes
(+ Utils: JWTProvider, PasswordManager, TimeProvider, UUIDGenerator)
(+ AWS: SSMClient, SESClient)
```

### レスポンス共通形式

```go
// 成功
utils.ResponseSuccess(c, statusCode, message, data, timeProvider)
// → { "success": true, "message": "...", "data": ..., "timestamp": "..." }

// エラー
utils.ResponseError(c, statusCode, message, errMsg, timeProvider)
// → { "success": false, "message": "...", "error": "...", "timestamp": "..." }
```

例外: `GET /api/health` → `{ "status": "ok" }` / `GET /api/auth/csrf-token` → `{ "token": "..." }`

### ミドルウェアパイプライン

```
CORS → SecurityHeaders → Logger → Recover → Timeout(5s) → RateLimit → Gzip
  → ルート別: VerifyUser / VerifyRoleAdmin / VerifyCSRFToken
```

### エラーハンドリング

- **repositories**: `gorm.ErrRecordNotFound` を判別し、見つからない場合は `nil` を返す
- **services**: `fmt.Errorf("context: %w", err)` でラップ
- **controllers**: HTTP ステータスコード (400/401/403/404/500) を使い分けて返却
- **カスタムエラー** (`internal/errors/`): `ErrNotFound`, `ErrUserNotFound`, `ErrUnauthorized`, `ErrInvalidInput`, `ErrAlreadyExists`

### 設定・シークレット管理

| 環境 | シークレット取得元 | .env ロード |
|------|------------------|-------------|
| production / staging | AWS SSM Parameter Store | しない |
| ローカル開発 | `.env` ファイル | する |

SSM パラメータパス: `/gin-todo/db/postgres-*`, `/gin-todo/db/jwt-secret-key`

### バリデーション

- **コントローラー層**: Gin の `binding:"required"` タグ
- **カスタムバリデーター** (`internal/validators/`): Option パターンで Email/Password を検証
  - Email: `.+@.+\..+`
  - Password: `^[a-zA-Z0-9.?/!-]{8,24}$`

### Cookie 設定の注意

`SetCookie` のパラメータ (path, domain, secure, httpOnly) は **セット時と削除時で一致させる必要がある**。不一致だとブラウザが Cookie を削除できない。

---

## Frontend Architecture

### API 通信の仕組み

- `api/client.ts`: 共通 axios インスタンス (Bearer トークン自動付与 + 401 時の自動リフレッシュ)
- `api/authApi.ts`: 公開認証エンドポイント (login/register/refresh) は生の `axios` を使用
- `api/taskApi.ts`, `api/adminApi.ts`: 認証付きエンドポイントは `apiClient` を使用

### 認証フロー

1. `POST /api/auth/login` → `{ access_token }` をメモリに保存 / `refresh_token` は HttpOnly Cookie
2. リクエストに `Authorization: Bearer <token>` を自動付与 (interceptor)
3. リロード時 → `POST /api/auth/refresh` で access_token を復元
4. 401 応答時 → interceptor が refresh を自動試行 → 成功すれば元リクエストを再試行
5. タスク書き込み前 → `GET /api/auth/csrf-token` で CSRF トークン取得 → `X-CSRF-Token` ヘッダーに付与

### ルーティング・認証ガード

- `PrivateRoute` — 未認証なら `/login` へリダイレクト
- `AdminRoute` — admin ロール以外は拒否
- ページコンポーネントは `React.lazy` で遅延読み込み

---

## ローカル開発の注意事項

- フロントエンド dev server は **必ず `http://localhost:3000`** で起動すること (`127.0.0.1` 不可)
  - CORS と Secure Cookie の localhost 例外に依存しているため
- Swagger UI: `http://localhost:8080/swagger/index.html` (`make swagger.init` 後)
- Adminer (DB ブラウザ): `http://localhost:9090`
- SES ローカル: `http://localhost:8005` (パスワードリセットメール)

## Key Rules

- 詳細なコーディング規約は `.claude/rules/` 参照
- API エンドポイント一覧は `.claude/rules/api-endpoints.md` 参照
