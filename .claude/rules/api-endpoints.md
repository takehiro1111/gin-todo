# API Endpoints

Base URL: `http://localhost:8080`
CORS 許可オリジン: `http://localhost:3000` (フロント dev server はこのポートにすること)

---

## レスポンス共通フォーマット

### 成功時

```typescript
interface SuccessResponse<T = unknown> {
  success: true
  message: string
  data?: T
  timestamp: string  // JST ISO 8601
}
```

### エラー時

```typescript
interface ErrorResponse {
  success: false
  message: string
  error?: string
  code?: string
  timestamp: string
}
```

**ほぼすべての** API レスポンスはこの形式。`data` フィールドにリソースが入る。

以下の 2 エンドポイントは例外 (独自形式を返す):

| エンドポイント | レスポンス形式 | 理由 |
|---------------|--------------|------|
| `GET /api/health` | `{ "status": "ok" }` | ヘルスチェック慣例形式 (LB・監視ツール向け) |
| `GET /api/auth/csrf-token` | `{ "token": "<uuid>" }` | CSRF トークン専用。`data` にネストしない設計 |

---

## Health

> **レスポンス形式の例外:** `{ "status": "ok" }` を直接返す (SuccessResponse 形式ではない)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/health` | - | ヘルスチェック |

---

## Auth (認証不要)

| Method | Path | Body | Description |
|--------|------|------|-------------|
| POST | `/api/auth/register` | `{ name, email, password }` | ユーザー登録 → `data: { access_token, refresh_token }` |
| POST | `/api/auth/login` | `{ email, password }` | ログイン → `data: { access_token, refresh_token }` |
| POST | `/api/auth/forgot` | `{ email }` | パスワードリセットメール送信 |
| POST | `/api/auth/reset` | `{ reset_token, new_password }` | パスワードリセット実行 |

**重要:** `access_token` も `refresh_token` も両方レスポンスボディで返る。
HttpOnly Cookie ではない → 両方 `localStorage` に保存する。

---

## Auth (要認証: Bearer)

| Method | Path | Body | Description |
|--------|------|------|-------------|
| POST | `/api/auth/logout` | - | ログアウト |
| POST | `/api/auth/refresh` | `{ refresh_token: string }` | Access Token 更新 → `data: string` (新 access_token)。**Bearer には refresh_token を使う** (下記参照) |
| GET | `/api/auth/me` | - | 自分のプロフィール取得 |
| PATCH | `/api/auth/password` | `{ old_password, new_password }` | パスワード変更 |
| GET | `/api/auth/csrf-token` | - | CSRF トークン取得 → `{ token: string }` を直接返す (共通 SuccessResponse 形式ではない) + Cookie セット |

### Refresh フロー (詳細)

`/api/auth/refresh` は `VerifyUser` ミドルウェア配下にあるため、`Authorization` ヘッダーに有効な JWT が必要。
Access Token が期限切れの状況でこのエンドポイントを呼ぶため、**Refresh Token を Bearer として送信する**。

```typescript
// axios インターセプターの例
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status !== 401) return Promise.reject(error)

    const refreshToken = localStorage.getItem('refresh_token')
    if (!refreshToken) {
      // refresh_token もなければログアウト
      logout()
      return Promise.reject(error)
    }

    try {
      const { data } = await axios.post(
        `${import.meta.env.VITE_API_BASE_URL}/api/auth/refresh`,
        { refresh_token: refreshToken },
        {
          // refresh_token を Bearer として送り VerifyUser を通過する
          headers: { Authorization: `Bearer ${refreshToken}` },
        }
      )
      const newAccessToken = data.data
      localStorage.setItem('access_token', newAccessToken)

      // 元のリクエストを新しい access_token で再試行
      error.config.headers.Authorization = `Bearer ${newAccessToken}`
      return apiClient.request(error.config)
    } catch {
      logout()
      return Promise.reject(error)
    }
  }
)
```

---

## Task (要認証 + CSRF)

| Method | Path | Body | Description |
|--------|------|------|-------------|
| GET | `/api/task/` | - | タスク一覧 |
| POST | `/api/task/` | CreateTaskBody | タスク作成 |
| GET | `/api/task/:id` | - | タスク取得 |
| PUT | `/api/task/:id` | UpdateTaskBody | タスク更新 |
| PATCH | `/api/task/:id` | `{ status_id: number }` | ステータスのみ更新 |
| DELETE | `/api/task/:id` | - | タスク削除 |

```typescript
interface CreateTaskBody {
  title: string        // required, max 100文字
  description?: string // max 300文字
  status_id?: number
  priority?: string    // "low" | "medium" | "high"
  due_date?: string    // ISO 8601
}

interface UpdateTaskBody {
  title: string        // required, max 100文字
  description?: string // max 100文字 (createと異なる)
  status_id?: number
  priority?: string
  due_date?: string
}
```

---

## Export (要認証)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/tasks/export/csv` | タスク CSV ダウンロード |

---

## Admin (要管理者権限)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/admin/users` | 全ユーザー一覧 |
| GET | `/api/admin/tasks` | 全タスク一覧 |

JWT の `role` クレームが `"admin"` のユーザーのみアクセス可能。

---

## WebSocket

| Protocol | Path | Description |
|----------|------|-------------|
| WS | `/api/ws/chat` | リアルタイムチャット |

**認証フロー (クエリパラメータ不使用):**
```typescript
const ws = new WebSocket('ws://localhost:8080/api/ws/chat')

ws.onopen = () => {
  // 接続直後、最初のメッセージとして JWT を送る
  ws.send(accessToken)
}

// 以降のメッセージは ChatMessage 形式で送受信
// 受信: { user: string, message: string }
// 送信: テキストそのまま (サーバーが ChatMessage に wrap して broadcast)
```

---

## マスターデータ (固定値)

### task_statuses

| id | name | description |
|----|------|-------------|
| 1 | `todo` | 未着手 |
| 2 | `in_progress` | 進行中 |
| 3 | `done` | 完了 |
| 4 | `archived` | アーカイブ |

### user_roles

| id | name |
|----|------|
| 1 | `admin` |
| 2 | `writer` |

---

## TypeScript 型定義

```typescript
interface Task {
  id: number
  user_id: number
  title: string
  description: string
  status_id: number
  priority: "low" | "medium" | "high"
  due_date: string | null
  created_at: string
  updated_at: string
}

interface User {
  id: number
  name: string
  email: string
  role_id: number
  created_at: string
  updated_at: string
}

// JWT に含まれるクレーム情報
interface JWTClaims {
  username: string
  user_id: number
  role: "admin" | "writer"
}
```

---

## CSRF フロー (詳細)

サーバーは `/api/auth/csrf-token` で以下を同時に行う:
1. レスポンス JSON に `{ token: string }` を返す
2. `csrf_token` Cookie をセット (Secure=**true**, HttpOnly=true, domain=localhost, 24時間)

フロント側は:
1. `credentials: 'include'` でリクエストし Cookie を受け取る
2. レスポンスの `token` を `X-CSRF-Token` ヘッダーに付けてタスク操作を行う
3. サーバーはヘッダー値と Cookie 値が一致するか検証する

### ローカル開発における Secure Cookie の注意事項

バックエンドは `Secure=true` をハードコードしており、環境による切り替えは現時点で**実装されていない**。

HTTP 上での `Secure` Cookie は通常ブラウザに保存されないが、**`localhost` は例外**として扱われる。

| 環境 | 動作 | 理由 |
|------|------|------|
| `localhost:3000` → `localhost:8080` (Chrome 89+, Firefox 75+) | **動作する** | ブラウザが `localhost` を「潜在的に信頼できるオリジン」とみなし、HTTP でも Secure Cookie を保存する |
| `127.0.0.1` や Docker カスタムドメイン (例: `app.local`) | **動作しない** | `localhost` 例外が適用されないため、HTTP 上で Secure Cookie が保存されない |
| curl / Postman など非ブラウザクライアント | **動作しない** | ブラウザの localhost 例外に依存しているため |

**フロント実装の前提:** 開発サーバーは必ず `http://localhost:3000` で動作させること。`127.0.0.1` や別ホスト名は使わない。

> **バックエンド改善メモ:** 将来的に Docker 環境や非ブラウザテストが必要になった場合は、`ENV` 環境変数を参照して `Secure=false` に切り替える実装が必要。

```typescript
// タスク操作の例
// /api/auth/csrf-token は SuccessResponse 形式ではなく { token: string } を直接返すため、
// apiClient のインターセプター (response.data.data アンラップ) を経由せず axios を直接使う
const response = await axios.get<{ token: string }>(
  `${import.meta.env.VITE_API_BASE_URL}/api/auth/csrf-token`,
  { withCredentials: true }
)
const csrfToken = response.data.token

await apiClient.post('/api/task/', payload, {
  headers: { 'X-CSRF-Token': csrfToken },
})
```

---

## エラーハンドリング

| ステータス | 対応 |
|-----------|------|
| 401 | Access Token 期限切れ → refresh して再試行、失敗したらログアウト |
| 403 | CSRF ミスマッチ or 権限不足 → csrf-token 再取得して再試行 |
| 429 | Rate Limit 超過 → しばらく待ってから再試行 |
| 408 / timeout | 5秒タイムアウト → ユーザーへ通知 |
