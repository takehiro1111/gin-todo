---
paths:
  - "frontend/src/api/**/*.ts"
  - "frontend/src/hooks/**/*.ts"
---

# API Communication Rules

## API クライアント構成

`src/api/` に機能単位でファイルを分割する。

```
api/
├── client.ts     # fetch wrapper (共通ヘッダー、エラーハンドリング、リトライ)
├── authApi.ts
├── taskApi.ts
└── adminApi.ts
```

---

## Token 管理

| Token | 保存場所 | 有効期限 | 補足 |
|-------|---------|---------|------|
| Access Token | `localStorage` | 15分 | ボディで返却 |
| Refresh Token | `localStorage` | 7日 | ボディで返却 (Cookie ではない) |

```typescript
// トークンの保存・取得ユーティリティ
const TOKEN_KEY = 'accessToken'
const REFRESH_KEY = 'refreshToken'

export const tokenStorage = {
  getAccess: () => localStorage.getItem(TOKEN_KEY),
  setAccess: (t: string) => localStorage.setItem(TOKEN_KEY, t),
  getRefresh: () => localStorage.getItem(REFRESH_KEY),
  setRefresh: (t: string) => localStorage.setItem(REFRESH_KEY, t),
  clear: () => { localStorage.removeItem(TOKEN_KEY); localStorage.removeItem(REFRESH_KEY) },
}
```

---

## client.ts の基本構造

```typescript
const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message)
    this.name = 'ApiError'
  }
}

let isRefreshing = false
let refreshQueue: Array<(token: string) => void> = []

async function request<T>(path: string, options: RequestInit = {}, retry = true): Promise<T> {
  const token = tokenStorage.getAccess()

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    credentials: 'include',  // csrf_token Cookie の送受信に必須
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  })

  // 401: Access Token 期限切れ → Refresh して1回だけ再試行
  if (res.status === 401 && retry) {
    const newToken = await doRefresh()
    tokenStorage.setAccess(newToken)
    return request<T>(path, options, false)
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(res.status, body.message ?? 'Unknown error')
  }

  // data フィールドを unwrap して返す
  const body = await res.json()
  return body.data as T
}

async function doRefresh(): Promise<string> {
  const refreshToken = tokenStorage.getRefresh()
  if (!refreshToken) throw new ApiError(401, 'No refresh token')

  const res = await fetch(`${BASE_URL}/api/auth/refresh`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken }),
  })

  if (!res.ok) {
    tokenStorage.clear()
    throw new ApiError(401, 'Refresh failed')
  }

  const body = await res.json()
  return body.data  // 新しい access_token
}

export const apiClient = {
  get:    <T>(path: string, opts?: RequestInit) => request<T>(path, { method: 'GET', ...opts }),
  post:   <T>(path: string, data: unknown, opts?: RequestInit) =>
    request<T>(path, { method: 'POST', body: JSON.stringify(data), ...opts }),
  put:    <T>(path: string, data: unknown, opts?: RequestInit) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(data), ...opts }),
  patch:  <T>(path: string, data: unknown, opts?: RequestInit) =>
    request<T>(path, { method: 'PATCH', body: JSON.stringify(data), ...opts }),
  delete: <T>(path: string, opts?: RequestInit) => request<T>(path, { method: 'DELETE', ...opts }),
}
```

---

## CSRF Token の扱い

タスク操作 (POST / PUT / PATCH / DELETE) 前に CSRF トークンを取得する。
サーバーはレスポンス JSON と Cookie を同時にセットするため、
`credentials: 'include'` で受け取った後、ヘッダーに付けて使う。

```typescript
// hooks/useCsrfToken.ts
export function useCsrfToken() {
  const getCsrfToken = async (): Promise<string> => {
    const token = await apiClient.get<string>('/api/auth/csrf-token')
    return token
  }
  return { getCsrfToken }
}

// 使い方 (taskApi.ts)
export const taskApi = {
  create: async (payload: CreateTaskBody) => {
    const csrfToken = await apiClient.get<string>('/api/auth/csrf-token')
    return apiClient.post<Task>('/api/task/', payload, {
      headers: { 'X-CSRF-Token': csrfToken }
    })
  },
  // ...
}
```

---

## カスタムフックパターン

```typescript
// hooks/useTasks.ts
export function useTasks() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchTasks = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await taskApi.getAll()
      setTasks(data)
    } catch (e) {
      if (e instanceof ApiError && e.status === 429) {
        setError('しばらく待ってから再試行してください')
      } else {
        setError(e instanceof ApiError ? e.message : 'タスクの取得に失敗しました')
      }
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => { fetchTasks() }, [])

  return { tasks, isLoading, error, refetch: fetchTasks }
}
```

---

## WebSocket 接続

```typescript
// hooks/useChat.ts
export function useChat() {
  const wsRef = useRef<WebSocket | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])

  const connect = () => {
    const ws = new WebSocket('ws://localhost:8080/api/ws/chat')
    wsRef.current = ws

    ws.onopen = () => {
      // 接続直後、最初のメッセージで JWT を送信して認証
      const token = tokenStorage.getAccess()
      if (token) ws.send(token)
    }

    ws.onmessage = (e) => {
      const msg: ChatMessage = JSON.parse(e.data)
      setMessages(prev => [...prev, msg])
    }

    ws.onerror = (e) => console.error('WS error', e)
    ws.onclose = () => { wsRef.current = null }
  }

  const sendMessage = (text: string) => {
    wsRef.current?.send(text)  // テキストをそのまま送る
  }

  useEffect(() => {
    connect()
    return () => wsRef.current?.close()
  }, [])

  return { messages, sendMessage }
}
```

---

## エラーハンドリング方針

```typescript
try {
  await taskApi.create(payload)
} catch (e) {
  if (e instanceof ApiError) {
    switch (e.status) {
      case 403: // CSRF ミスマッチ → 再取得して再試行
      case 429: // Rate Limit → "しばらく待ってください" 表示
      case 408: // Timeout → "タイムアウトしました" 表示
      default:  // その他 → e.message を表示
    }
  }
}
```
