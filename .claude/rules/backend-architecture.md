---
paths:
  - "backend/**/*.go"
---

# Backend Architecture Rules

## レイヤー責務

| レイヤー | 責務 | やってはいけないこと |
|----------|------|---------------------|
| controllers | リクエストバインド・バリデーション・レスポンス返却 | DB 直接操作、ビジネスロジック |
| services | ビジネスロジック・トランザクション制御 | HTTP 依存 (`*gin.Context` のヘッダー操作等) |
| repositories | データアクセス (GORM) | ビジネスルールの判定 |
| models | 構造体定義・DB 初期化・マイグレーション | ロジック |

## DI・コンストラクタ

- すべての主要コンポーネントは **interface を定義** し、ファクトリ関数で実装を返す
- ファクトリ関数の戻り値は interface 型にする

```go
// Good
func NewTaskService(taskRepo repositories.TaskRepository) TaskService {
    return &TaskServiceImpl{taskRepo: taskRepo}
}

// Bad: 実装型を返す
func NewTaskService(taskRepo repositories.TaskRepository) *TaskServiceImpl { ... }
```

## エラーハンドリング

- repositories: `gorm.ErrRecordNotFound` を判別し、見つからない場合は `nil, nil` を返す
- services: `fmt.Errorf("コンテキスト: %w", err)` でラップして返す
- controllers: サービスのエラー内容に応じて HTTP ステータスを使い分ける (400/401/404/500)
- `internal/errors/` のカスタムエラーは `errors.Is()` で比較する

```go
// Good: サービス層
return nil, fmt.Errorf("failed to find user: %w", err)

// Good: コントローラー層
if errors.Is(err, appErr.ErrUserNotFound) {
    utils.ResponseError(c, http.StatusNotFound, ...)
    return
}
```

## レスポンス

共通ヘルパーを使う。直接 `c.JSON()` を呼ばない。

```go
utils.ResponseSuccess(c, http.StatusOK, "message", data, a.timeProvider)
utils.ResponseError(c, http.StatusBadRequest, "message", err.Error(), a.timeProvider)
```

## バリデーション

- コントローラー層: Gin の `binding:"required"` タグでリクエスト構造体を検証
- カスタムバリデーション: `internal/validators/` の Option パターンを使う

```go
err := validators.NewAuthenticateValidator(
    validators.WithEmail(req.Email),
    validators.WithPassword(req.Password),
)
```

## Cookie

`SetCookie` のパラメータ (path, domain, secure, httpOnly) は **セット時と削除時で完全一致させる**。不一致だとブラウザが Cookie を消せない。

```go
// セット
c.SetCookie("refresh_token", token, 60*60*24*7, "/", "", false, true)
// 削除 — 同じパラメータで MaxAge=-1
c.SetCookie("refresh_token", "",    -1,         "/", "", false, true)
```

## ミドルウェア

- `VerifyUser`: JWT 検証 → `c.Set("user_id", claims.UserID)` でコンテキストにセット
- `VerifyRoleAdmin`: JWT 検証 + `role == "admin"` チェック
- `VerifyCSRFToken`: タスクの書き込み操作 (POST/PUT/PATCH/DELETE) にのみ適用
- GET エンドポイントには CSRF 不要

## テスト

- テストファイルは `backend/tests/unit/`, `backend/tests/integration/` に配置
- `testify/assert` を使用
- DI によりモック差し替えが可能 (interface ベース)
- `TimeProvider`, `UUIDGenerator` 等のユーティリティもモック可能

## マイグレーション

- `backend/migrations/` に SQL ファイルを配置 (golang-migrate 形式)
- 命名: `NNN_description.up.sql` / `NNN_description.down.sql`
- 起動時に自動実行される (`models.RunMigration`)
