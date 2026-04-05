# Gin Todo API
- タスク管理API（Go + Gin + React）のフルスタックアプリケーション

## 目的
- Go言語の基礎的な実装力の向上

## 技術スタック

### Backend(実装済み)
- Go
- Gin Web Framework
- PostgreSQL
- GORM (ORM)
- golang-migrate (DBマイグレーション)
- JWT認証 (Access Token + Refresh Token)
- WebSocket (リアルタイム通知)
- AWS SDK v2 (SSM ParameterStore / SES)
- Swagger / OpenAPI (swaggo)
- Air (ホットリロード)

### Frontend(未実装 / 使用予定の技術スタック)
- React
- TypeScript
- Vite
- React Router
- Context API (状態管理)
- Tailwind CSS

### Infra(未実装 / バックエンドに必要なリソースを一部設定)
- AWS
  - VPC
  - Subnet
  - Internet Gateway
  - NAT Instance
  - Route53
  - CloudFront
    - VPC Origin
  - ALB
  - S3
  - ECS
  - RDS for PostgreSQL
  - SSM Parameter Store
  - CDK with TypeScript

### CI/CD(未実装)
- Github Actions

---

## アーキテクチャ / リクエスト処理フロー

Router 層で共通ミドルウェアを通し、Controller → Service → Repository → DB の順に処理を委譲する。
外部サービス呼び出しは Service 層から AWS SDK 経由で行う。

```mermaid
sequenceDiagram
    participant C as Client
    participant MW as Middleware
    participant Ctrl as Controller
    participant Svc as Service
    participant Repo as Repository
    participant DB as PostgreSQL
    participant AWS as AWS SSM SES

    C->>MW: HTTP Request
    MW->>MW: CORS Logger RateLimit Gzip
    MW->>MW: JWT 検証 認証ルートのみ
    MW->>MW: CSRF 検証 書き込み時のみ
    MW->>Ctrl: gin.Context user_id セット済み
    Ctrl->>Ctrl: バインド バリデーション
    Ctrl->>Svc: ビジネスロジック呼び出し
    Svc->>Repo: データアクセス
    Repo->>DB: SQL GORM
    DB-->>Repo: 結果
    Svc->>AWS: SSM SES 呼び出し
    AWS-->>Svc: レスポンス
    Svc-->>Ctrl: データ or error
    Ctrl-->>C: JSON レスポンス
```

---

## 認証・認可

### 認証フロー総合図 (ログイン / Refresh / CSRF / 401 自動リトライ)

- Access Token はメモリ保存 15分、Refresh Token は HttpOnly Cookie 7日
- タスク書き込み前に CSRF トークンを取得しヘッダーに付与 (ダブルサブミット)
- 401 は axios interceptor が自動で refresh して再試行

```mermaid
sequenceDiagram
    participant F as Frontend
    participant API as Backend API

    Note over F,API: 1. ログイン
    F->>API: POST /api/auth/login
    API-->>F: 200 access_token<br/>Set-Cookie refresh_token

    Note over F,API: 2. リロード時は refresh でセッション復元
    F->>API: POST /api/auth/refresh
    API-->>F: 200 新 access_token

    Note over F,API: 3. タスク書き込み前は CSRF トークン取得
    F->>API: GET /api/auth/csrf-token
    API-->>F: token<br/>Set-Cookie csrf_token
    F->>API: POST /api/task/<br/>X-CSRF-Token ヘッダー付与
    API-->>F: 201 task

    Note over F,API: 4. 401 は interceptor が自動リトライ
    F->>API: GET /api/task/ 期限切れ token
    API-->>F: 401 Unauthorized
    F->>API: POST /api/auth/refresh
    API-->>F: 200 新 access_token
    F->>API: GET /api/task/ 新 token で再試行
    API-->>F: 200 tasks
```

### 認可 (ロールベースアクセス制御)

```mermaid
flowchart TD
    Req["リクエスト"] --> JWT{"JWT 検証"}
    JWT -->|無効| R401["401 Unauthorized"]
    JWT -->|有効| Role{"ロール判定"}
    Role -->|VerifyUser| OK1["Controller へ"]
    Role -->|VerifyRoleAdmin| Admin{"role admin"}
    Admin -->|Yes| OK2["Controller へ"]
    Admin -->|No| R403["403 Forbidden"]
```

### トークン管理まとめ

| トークン | 保存場所 | 有効期限 | 用途 |
|----------|----------|----------|------|
| Access Token | フロントのメモリ (変数) | 15分 | API リクエストの認証 |
| Refresh Token | HttpOnly Cookie | 7日 | Access Token の更新 |
| CSRF Token | ヘッダー + Cookie (ダブルサブミット) | 24時間 | 書き込み操作の保護 |

---

## ローカル開発

```bash
# 1. DB / メールサーバー起動
docker compose up -d

# 2. バックエンド起動
cd backend
air                    # ホットリロード

# 3. フロントエンド起動
cd frontend
pnpm dev               # http://localhost:3000
```

| サービス | URL |
|----------|-----|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| Adminer (DB) | http://localhost:9090 |
| SES Local | http://localhost:8005 |
