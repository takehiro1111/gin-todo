---
paths:
  - "frontend/src/components/**/*.tsx"
  - "frontend/src/pages/**/*.tsx"
---

# Frontend Component Rules

## コンポーネント基本方針

- 関数コンポーネント + hooks のみ使用 (クラスコンポーネント禁止 ※ErrorBoundary は例外)
- `export default` ではなく named export を使う
- 1ファイル 1コンポーネント
- Props は関数の引数に直接インラインで型を書く (`interface Props` は使わない)

```typescript
// Good
export function TaskCard({ title, onDelete }: { title: string; onDelete: (id: number) => void }) {
  return <div>{title}</div>
}

// Bad
export default function TaskCard(props: any) { ... }
// Bad (interface Props は使わない)
interface Props { title: string }
export function TaskCard({ title }: Props) { ... }
```

## ディレクトリ構成

```
components/
├── ui/          # Button, Input, Modal 等の汎用UI
├── layout/      # Header, Sidebar, Layout 等
├── task/        # TaskCard, TaskForm, TaskList 等
└── auth/        # LoginForm, RegisterForm 等

pages/
├── LoginPage.tsx
├── RegisterPage.tsx
├── DashboardPage.tsx
├── TaskDetailPage.tsx
└── AdminPage.tsx
```

## ファイル命名

- コンポーネント: PascalCase (`TaskCard.tsx`)
- hooks: camelCase with `use` prefix (`useAuth.ts`)
- 型定義: camelCase (`task.ts`)
- API: camelCase (`taskApi.ts`)

## カスタムフックへの切り出し

コンポーネントから切り出せるロジックは積極的にカスタムフックに分離する。
コンポーネントは「表示」に集中させ、「振る舞い」はフックに持たせる。

```
// 切り出しの目安
- useState + useEffect の組み合わせ → カスタムフックに切り出す
- 複数コンポーネントで同じロジックが出てきた → 即カスタムフックに切り出す
- コンポーネントのロジック行数が増えてきた → カスタムフックに切り出す
```

```tsx
// Bad: ロジックがコンポーネントに直書き、Promise チェーンを使っている
export function TaskList() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [isLoading, setIsLoading] = useState(false)
  useEffect(() => {
    setIsLoading(true)
    taskApi.getAll().then(setTasks).catch(console.error).finally(() => setIsLoading(false))
  }, [])
  return ...
}

// Good: ロジックをカスタムフックに切り出してコンポーネントをシンプルに保つ
export function TaskList() {
  const { tasks, isLoading, error } = useTasks()
  return ...
}
```

カスタムフックは `src/hooks/` に配置する。

## State 管理方針

- サーバー状態 (API データ) → TanStack Query (`useQuery` / `useMutation`)
- クライアントのグローバル状態 (認証情報等) → Context API (`contexts/`)
- ローカル状態 → `useState`
- 副作用・データ取得ロジック → カスタムフック (`hooks/`)
- Redux 等の外部状態管理ライブラリは使わない

### Context に入れる判断基準

「複数のコンポーネントをまたいで使う状態」は Context に持ち上げて共通化する。
props のバケツリレーが2段以上になりそうなら Context を検討する。

| Context | 持たせる状態 |
|---------|------------|
| `AuthContext` | `user`, `login()`, `logout()`, `isLoading` |

タスクデータはサーバー状態のため TanStack Query (`useTasks`, `useCreateTask` 等) で管理する。
`TaskContext` は不要。

```typescript
// Bad: props バケツリレー
<DashboardPage user={user}>
  <TaskList user={user}>
    <TaskCard user={user} />

// Good: Context から直接取得
export function TaskCard({ taskId }: { taskId: number }) {
  const { user } = useAuth()
  const { tasks } = useTask()
  ...
}
```

### Context に入れない状態

- フォームの入力値 → `useState` (そのコンポーネント内で完結)
- モーダルの開閉 → `useState` (親1段で管理できる)
- API のローディング・エラー → カスタムフック内の `useState`

## Context の使い方

```typescript
// contexts/AuthContext.tsx
interface AuthContextValue {
  user: User | null
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
```

## React 実装の注意点

**useEffect は必要最低限にする**
useEffect は副作用を伴うため、使用箇所を最小限に抑える。
以下の場合のみ使用を検討する。
- データ取得 (マウント時の API コール)
- WebSocket・タイマー・イベントリスナーの登録
- 外部ライブラリとの連携 (nprogress 等)

イベントハンドラで済む処理・派生する値の計算に useEffect を使わない。
```typescript
// Bad: 値の派生に useEffect を使っている
const [filtered, setFiltered] = useState<Task[]>([])
useEffect(() => {
  setFiltered(tasks.filter(t => !t.done))
}, [tasks])

// Good: useMemo で計算する
const filtered = useMemo(() => tasks.filter(t => !t.done), [tasks])
```

**useEffect のクリーンアップ**
WebSocket・タイマー・イベントリスナーはアンマウント時に必ず解除する。
```typescript
useEffect(() => {
  const ws = new WebSocket(...)
  return () => ws.close() // アンマウント時に切断
}, [])
```

**axios のリクエストキャンセル**
ページ遷移時に未完了のリクエストをキャンセルする。
アンマウント済みコンポーネントへの state セットを防ぐ。
TanStack Query を使う場合は自動でキャンセルされるため不要。
直接 axios を呼ぶ場合は `AbortController` を使う。
```typescript
useEffect(() => {
  const controller = new AbortController()
  apiClient.get('/api/task/', { signal: controller.signal }).then(...)
  return () => controller.abort()
}, [])
```

**list の key に index を使わない**
```tsx
// Bad  : 並び替え・追加・削除時に意図しない再レンダリングが起きる
tasks.map((t, i) => <TaskCard key={i} ... />)
// Good : ユニークな ID を使う
tasks.map(t => <TaskCard key={t.id} ... />)
```

**Context は用途ごとに分割する**
AuthContext と TaskContext を1つにまとめると、タスク更新のたびに
認証情報を参照するコンポーネントも再レンダリングされる。

**Error Boundary を設置する**
レンダリングエラーをキャッチしないと画面が真っ白になる。
ページ単位で `<ErrorBoundary>` を設置する。

**StrictMode を有効にする**
`main.tsx` で `<StrictMode>` で囲む。
useEffect が意図的に2回実行され、クリーンアップ漏れを早期発見できる。

## コメント (JSDoc)

コンポーネント・hooks・API関数にはJSDocを書く。型は TypeScript で表現するため `@param` の型注釈は省略する。

```tsx
/**
 * タスクの概要を表示するカード。
 * 完了済みの場合は半透明で表示される。
 */
export function TaskCard({ title, done }: { title: string; done: boolean }) { ... }

/**
 * タスク一覧を取得するカスタムフック。
 * @returns tasks, isLoading, error, refetch
 */
export function useTasks() { ... }

/**
 * タスクを新規作成する。
 * CSRF トークンを自動取得してリクエストに付与する。
 */
export async function createTask(payload: CreateTaskBody): Promise<Task> { ... }
```

- ロジックが自明な場合はコメント不要
- `@param` の型注釈は省略 (TypeScript の型定義で十分)
- 実装が複雑な箇所・重要な箇所は行単位でインラインコメントを書く

```typescript
// 401 かつ未リトライの場合のみリフレッシュを試みる (無限ループ防止のため _retry フラグを使う)
if (error.response?.status === 401 && !original._retry) {
  original._retry = true
  const { data } = await axios.post('/api/auth/refresh', { refresh_token: refreshToken })
  // 新しいトークンで元のリクエストを1回だけ再試行
  tokenStorage.setAccess(data.data)
  return apiClient(original)
}

// Cookie と X-CSRF-Token ヘッダーの両方が一致しないとサーバーが 403 を返す
// withCredentials: true がないと csrf_token Cookie が送られず必ず失敗する
const { data } = await apiClient.post('/api/task/', payload, {
  headers: { 'X-CSRF-Token': csrfToken },
})
```

## TypeScript 方針

- `tsconfig.json` の `strict: true` を必ず有効にする
- `any` は使わない。型が不明な場合は `unknown` を使って絞り込む

```typescript
// Bad
const data: any = await fetchSomething()

// Good
const data: unknown = await fetchSomething()
if (typeof data === 'object' && data !== null && 'id' in data) { ... }
```

## フォーム (react-hook-form + zod)

バリデーションは zod スキーマで定義し、`@hookform/resolvers/zod` で react-hook-form に渡す。
バックエンドの制約と合わせてスキーマを定義する。

```tsx
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const loginSchema = z.object({
  email: z.string().email('正しいメールアドレスを入力してください'),
  password: z.string().min(8, '8文字以上で入力してください'),
})

// スキーマから型を生成する (手書き不要)
type LoginFields = z.infer<typeof loginSchema>

export function LoginForm() {
  const { register, handleSubmit, formState: { errors } } = useForm<LoginFields>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = async (data: LoginFields) => { ... }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input {...register('email')} />
      {errors.email && <p>{errors.email.message}</p>}
    </form>
  )
}
```

### タスクフォームのスキーマ例 (バックエンドの制約に合わせる)

```typescript
const createTaskSchema = z.object({
  title:       z.string().min(1, '必須').max(100, '100文字以内'),
  description: z.string().max(300, '300文字以内').optional(),
  priority:    z.enum(['low', 'medium', 'high']).default('medium'),
  due_date:    z.string().optional(),
})

const updateTaskSchema = z.object({
  title:       z.string().min(1, '必須').max(100, '100文字以内'),
  description: z.string().max(100, '100文字以内').optional(), // updateは100文字
  priority:    z.enum(['low', 'medium', 'high']),
  due_date:    z.string().optional(),
})
```

## 日付 (date-fns)

```typescript
import { format, isPast } from 'date-fns'
import { ja } from 'date-fns/locale'

format(new Date(task.due_date), 'yyyy/MM/dd', { locale: ja })
isPast(new Date(task.due_date))  // 期限切れチェック
```

## Toast 通知 (react-hot-toast)

API 操作の成功・失敗は必ず Toast で通知する。

```typescript
import toast from 'react-hot-toast'

// 成功
toast.success('タスクを作成しました')

// エラー
toast.error('タスクの作成に失敗しました')

// ローディング付き
await toast.promise(taskApi.create(payload), {
  loading: '作成中...',
  success: 'タスクを作成しました',
  error: '作成に失敗しました',
})
```

`<Toaster />` は `App.tsx` のルートに1つだけ置く。

## スタイリング

- Tailwind CSS のユーティリティクラスを使う
- `style` 属性・CSS Modules・CSS-in-JS は使わない
- 条件付きクラスは `clsx` を使う

```tsx
import clsx from 'clsx'

export function TaskCard({ title, done }: { title: string; done: boolean }) {
  return (
    <div className={clsx('rounded p-4 border', done ? 'opacity-50' : 'bg-white')}>
      {title}
    </div>
  )
}
```

## ローディング表示

用途によって使い分ける。

| 用途 | ライブラリ | 使いどころ |
|------|-----------|-----------|
| コンテンツ読み込み中 | `react-loading-skeleton` | タスク一覧・詳細など |
| ページ遷移 | `nprogress` | ルート切り替え時 |
| ボタン操作中 | Tailwind の `animate-spin` | 送信ボタンなど |

```tsx
import Skeleton from 'react-loading-skeleton'
import 'react-loading-skeleton/dist/skeleton.css'

export function TaskList({ tasks, isLoading }: { tasks: Task[]; isLoading: boolean }) {
  if (isLoading) {
    // コンテンツと同じ形のスケルトンを表示する
    return <>{[...Array(5)].map((_, i) => <Skeleton key={i} height={60} className="mb-2" />)}</>
  }
  return <>{tasks.map(t => <TaskCard key={t.id} task={t} />)}</>
}
```

```typescript
// nprogress はページ遷移時に自動でプログレスバーを表示する
// App.tsx の Router 内で useEffect を使って設定する
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'

// ルート変更時に start/done を呼ぶ
NProgress.start()
NProgress.done()
```

## Error / Loading の扱い

- API コールは try/catch で囲む
- ローディング中は `isLoading` フラグで表示制御し、スケルトンを表示する
- エラーはユーザーに分かりやすいメッセージを表示 (生のエラーメッセージをそのまま表示しない)
