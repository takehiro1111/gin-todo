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

## アーキテクチャ

### バックエンドのレイヤー構成

```mermaid
graph LR
    Client[クライアント] --> Router[Router / Middleware]
    Router --> Controller[Controller]
    Controller --> Service[Service]
    Service --> Repository[Repository]
    Repository --> DB[(PostgreSQL)]
    Service --> AWS[AWS SDK<br/>SSM / SES]
```

### リクエスト処理フロー

```mermaid
sequenceDiagram
    participant C as Client
    participant M as Middleware
    participant Ctrl as Controller
    participant Svc as Service
    participant Repo as Repository
    participant DB as PostgreSQL

    C->>M: HTTP Request
    M->>M: CORS / Logger / RateLimit / Gzip
    M->>M: JWT 検証 (認証必須ルートのみ)
    M->>M: CSRF 検証 (書き込み操作のみ)
    M->>Ctrl: gin.Context (user_id セット済み)
    Ctrl->>Ctrl: リクエストバインド & バリデーション
    Ctrl->>Svc: ビジネスロジック呼び出し
    Svc->>Repo: データアクセス
    Repo->>DB: SQL (GORM)
    DB-->>Repo: 結果
    Repo-->>Svc: モデル or nil
    Svc-->>Ctrl: データ or error
    Ctrl-->>C: JSON レスポンス (SuccessResponse / ErrorResponse)
```

---

## 認証・認可

### ログインフロー

```mermaid
sequenceDiagram
    participant B as ブラウザ
    participant F as Frontend
    participant API as Backend API

    B->>F: メールアドレス・パスワード入力
    F->>API: POST /api/auth/login
    API->>API: パスワード照合 bcrypt
    API->>API: Access Token 生成 JWT 15分
    API->>API: Refresh Token 生成 JWT 7日
    API-->>F: 200 access_token
    Note over API,F: Set-Cookie refresh_token<br/>HttpOnly SameSite=Lax
    F->>F: access_token をメモリに保存
    F->>API: GET /api/auth/me<br/>Authorization Bearer token
    API-->>F: 200 user情報
    F->>B: ダッシュボード表示
```

### トークンリフレッシュ (ページリロード時 / 401 発生時)

```mermaid
sequenceDiagram
    participant F as Frontend
    participant API as Backend API

    Note over F: リロード時はメモリの access_token が消える
    F->>API: POST /api/auth/refresh<br/>Cookie refresh_token は自動送信
    API->>API: refresh_token を検証
    API->>API: 新しい Access Token 生成
    API->>API: 新しい Refresh Token 生成
    API-->>F: 200 新 access_token
    Note over API,F: Set-Cookie refresh_token=新token<br/>トークンローテーション
    F->>F: 新 access_token をメモリに保存
    F->>API: GET /api/auth/me Bearer 新token
    API-->>F: 200 user情報
    Note over F: セッション復元完了
```

### 401 発生時の自動リトライ (axios interceptor)

```mermaid
sequenceDiagram
    participant F as Frontend
    participant Int as axios interceptor
    participant API as Backend API

    F->>API: GET /api/task/ 期限切れ access_token
    API-->>Int: 401 Unauthorized
    Int->>API: POST /api/auth/refresh<br/>Cookie 自動送信
    API-->>Int: 200 新 access_token
    Int->>Int: メモリに保存
    Int->>API: GET /api/task/ 新 access_token で再試行
    API-->>F: 200 tasks
    Note over F: ユーザーはエラーに気づかない
```

### CSRF トークンフロー (タスク書き込み操作)

```mermaid
sequenceDiagram
    participant F as Frontend
    participant API as Backend API

    F->>API: GET /api/auth/csrf-token<br/>Authorization Bearer token
    API->>API: UUID でトークン生成
    API-->>F: 200 token uuid
    Note over API,F: Set-Cookie csrf_token=uuid<br/>HttpOnly Secure
    F->>API: POST /api/task/<br/>Authorization Bearer token<br/>X-CSRF-Token uuid<br/>Cookie csrf_token も自動送信
    API->>API: ヘッダーの token と<br/>Cookie の token が一致するか検証
    API-->>F: 201 task
```

### 認可 (ロールベースアクセス制御)

```mermaid
graph TD
    Req[リクエスト] --> JWT{JWT 検証}
    JWT -->|無効| R401[401 Unauthorized]
    JWT -->|有効| Role{ロール判定}

    Role -->|一般ルート<br/>VerifyUser| OK1[user_id を Context にセット<br/>→ Controller へ]
    Role -->|管理者ルート<br/>VerifyRoleAdmin| Admin{role == admin?}
    Admin -->|Yes| OK2[Controller へ]
    Admin -->|No| R403[403 Forbidden]

    style R401 fill:#fee,stroke:#c00
    style R403 fill:#fee,stroke:#c00
    style OK1 fill:#efe,stroke:#0a0
    style OK2 fill:#efe,stroke:#0a0
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
