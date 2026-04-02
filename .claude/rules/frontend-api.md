---
paths:
  - "frontend/src/api/**/*.ts"
  - "frontend/src/hooks/**/*.ts"
---

# API Communication Rules

## 基本方針

- 非同期処理は **async/await** を使う (Promise チェーンは使わない)
- エラーハンドリングは **try/catch** で行う
- HTTP クライアントは **axios**
- サーバー状態のキャッシュ・取得は **TanStack Query**

---

## API クライアント構成

```
api/
├── client.ts     # axios インスタンス (共通設定・インターセプター)
├── authApi.ts
├── taskApi.ts
└── adminApi.ts
```

---

## client.ts (axios)

```typescript
import axios from 'axios'
import { tokenStorage } from '@/utils/tokenStorage'

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  withCredentials: true,  // csrf_token Cookie の送受信に必須
  headers: { 'Content-Type': 'application/json' },
})

// リクエストインターセプター: Access Token を自動付与
apiClient.interceptors.request.use((config) => {
  const token = tokenStorage.getAccess()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// レスポンスインターセプター: 401 時にリフレッシュして再試行
apiClient.interceptors.response.use(
  (res) => res,
  async (error) => {
    const original = error.config

    // 401 かつ未リトライの場合のみリフレッシュを試みる (無限ループ防止)
    if (error.response?.status === 401 && !original._retry) {
      original._retry = true
      try {
        // refresh_token は HttpOnly Cookie で自動送信される (withCredentials: true)
        const { data } = await axios.post(
          `${import.meta.env.VITE_API_BASE_URL}/api/auth/refresh`,
          {},
          { withCredentials: true }
        )
        const newAccessToken = data.data as string
        tokenStorage.setAccess(newAccessToken)
        original.headers.Authorization = `Bearer ${newAccessToken}`
        return apiClient(original)
      } catch {
        // リフレッシュ失敗もトークン削除のみ。リダイレクトは PrivateRoute に任せる
        tokenStorage.clear()
        return Promise.reject(error)
      }
    }

    return Promise.reject(error)
  }
)
```

---

## Token 管理

| Token | 保存場所 | 有効期限 | 備考 |
|-------|---------|---------|------|
| Access Token | メモリ (モジュール変数) | 15分 | XSS で盗まれても短命 |
| Refresh Token | HttpOnly Cookie (サーバー管理) | 7日 | JS からアクセス不可 |

```typescript
// utils/tokenStorage.ts
// access_token はメモリ管理。localStorage には保存しない
let _accessToken: string | null = null

export const tokenStorage = {
  getAccess: () => _accessToken,
  setAccess: (t: string) => { _accessToken = t },
  clear: () => { _accessToken = null },
}
```

**ページリロード時の復元:** メモリはリロードで消えるため、`AuthContext` マウント時に `POST /api/auth/refresh` を呼んで access_token を復元する。Cookie が自動送信されるため追加の認証情報は不要。

---

## API 関数 (async/await + try/catch)

```typescript
// api/taskApi.ts
import { apiClient } from '@/api/client'
import type { Task, CreateTaskBody, UpdateTaskBody } from '@/types/task'

export const taskApi = {
  getAll: async (): Promise<Task[]> => {
    const { data } = await apiClient.get('/api/task/')
    return data.data
  },

  getById: async (id: number): Promise<Task> => {
    const { data } = await apiClient.get(`/api/task/${id}`)
    return data.data
  },

  create: async (payload: CreateTaskBody): Promise<Task> => {
    // タスク操作前に CSRF トークンを取得してヘッダーに付与する
    const { data: csrf } = await apiClient.get('/api/auth/csrf-token')
    const { data } = await apiClient.post('/api/task/', payload, {
      headers: { 'X-CSRF-Token': csrf.data.token },
    })
    return data.data
  },

  update: async (id: number, payload: UpdateTaskBody): Promise<Task> => {
    const { data: csrf } = await apiClient.get('/api/auth/csrf-token')
    const { data } = await apiClient.put(`/api/task/${id}`, payload, {
      headers: { 'X-CSRF-Token': csrf.data.token },
    })
    return data.data
  },

  updateStatus: async (id: number, statusId: number): Promise<void> => {
    const { data: csrf } = await apiClient.get('/api/auth/csrf-token')
    await apiClient.patch(`/api/task/${id}`, { status_id: statusId }, {
      headers: { 'X-CSRF-Token': csrf.data.token },
    })
  },

  delete: async (id: number): Promise<void> => {
    const { data: csrf } = await apiClient.get('/api/auth/csrf-token')
    await apiClient.delete(`/api/task/${id}`, {
      headers: { 'X-CSRF-Token': csrf.data.token },
    })
  },
}
```

---

## TanStack Query

`QueryClient` は `main.tsx` のルートに1つだけ設置する。

```tsx
// main.tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
const queryClient = new QueryClient()

<QueryClientProvider client={queryClient}>
  <App />
</QueryClientProvider>
```

### データ取得 (useQuery)

```typescript
// hooks/useTasks.ts
import { useQuery } from '@tanstack/react-query'
import { taskApi } from '@/api/taskApi'

export function useTasks() {
  return useQuery({
    queryKey: ['tasks'],
    queryFn: taskApi.getAll,
  })
}

// 使い方
const { data: tasks, isLoading, error } = useTasks()
```

### データ更新 (useMutation)

```typescript
// hooks/useCreateTask.ts
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { taskApi } from '@/api/taskApi'
import toast from 'react-hot-toast'

export function useCreateTask() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: taskApi.create,
    onSuccess: async () => {
      // 作成成功後にキャッシュを破棄して一覧を再取得する
      await queryClient.invalidateQueries({ queryKey: ['tasks'] })
      toast.success('タスクを作成しました')
    },
    onError: (error) => {
      toast.error('タスクの作成に失敗しました')
    },
  })
}
```

### try/catch が必要な場面

TanStack Query の外で async 処理を書く場合は try/catch を使う。

```typescript
const handleSubmit = async (data: LoginFields) => {
  try {
    const { access_token, refresh_token } = await authApi.login(data)
    tokenStorage.setAccess(access_token)
    tokenStorage.setRefresh(refresh_token)
    navigate('/')
  } catch (error) {
    if (axios.isAxiosError(error)) {
      toast.error(error.response?.data?.message ?? 'ログインに失敗しました')
    }
  }
}
```

---

## 並列 API 呼び出し

独立したリクエストは `Promise.all` で並列実行する。
依存関係がない限りシーケンシャルにしない。

```typescript
// Bad: 順番に待つ必要がないのに直列になっている
const user = await authApi.getMe()
const tasks = await taskApi.getAll()

// Good: 並列で取得してパフォーマンスを上げる
const [user, tasks] = await Promise.all([
  authApi.getMe(),
  taskApi.getAll(),
])
```

一部失敗しても他の結果を使いたい場合は `Promise.allSettled` を使う。

```typescript
const results = await Promise.allSettled([taskApi.getAll(), adminApi.getStats()])
results.forEach(r => {
  if (r.status === 'fulfilled') { ... }
})
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

    ws.onclose = () => { wsRef.current = null }
  }

  const sendMessage = (text: string) => {
    wsRef.current?.send(text)
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
  await someApi()
} catch (error) {
  if (axios.isAxiosError(error)) {
    switch (error.response?.status) {
      case 403:  // CSRF ミスマッチ → csrf-token を再取得して再試行
      case 429:  // Rate Limit → "しばらく待ってください" 表示
      case 408:  // Timeout → "タイムアウトしました" 表示
      default:
        toast.error(error.response?.data?.message ?? 'エラーが発生しました')
    }
  }
}
```
