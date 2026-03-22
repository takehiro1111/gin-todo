# Frontend Project Setup Rules

## Biome (Linter + Formatter)

ESLint + Prettier の代わりに Biome を使う。高速かつ設定がシンプル。

```bash
pnpm add -D @biomejs/biome
pnpm biome init
```

### biome.json

```json
{
  "$schema": "https://biomejs.dev/schemas/1.9.0/schema.json",
  "organizeImports": { "enabled": true },
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true,
      "suspicious": {
        "noExplicitAny": "error"
      },
      "style": {
        "noParameterAssign": "error"
      },
      "correctness": {
        "useExhaustiveDependencies": "warn"
      }
    }
  },
  "formatter": {
    "enabled": true,
    "indentStyle": "space",
    "indentWidth": 2,
    "lineWidth": 100
  },
  "javascript": {
    "formatter": {
      "quoteStyle": "single",
      "trailingCommas": "es5",
      "semicolons": "asNeeded"
    }
  }
}
```

### コマンド

```bash
pnpm biome check src/          # lint + format チェック
pnpm biome check --write src/  # 自動修正
```

## husky + lint-staged (コミット前自動チェック)

```bash
pnpm add -D husky lint-staged
pnpm exec husky init
```

```json
// package.json
{
  "lint-staged": {
    "src/**/*.{ts,tsx}": ["biome check --write"]
  }
}
```

コミット前に Biome が自動実行され、エラーがあればコミットをブロックする。

## 型管理 (src/types/)

API レスポンスの型は `src/types/` に集約する。
フォームの型は `z.infer` で生成するため手書きしない。

```
src/types/
├── task.ts      # Task, CreateTaskBody, UpdateTaskBody
├── user.ts      # User, JWTClaims
└── common.ts    # SuccessResponse, ErrorResponse 等の共通型
```

```typescript
// types/task.ts
export interface Task {
  id: number
  user_id: number
  title: string
  description: string
  status_id: number
  priority: 'low' | 'medium' | 'high'
  due_date: string | null
  created_at: string
  updated_at: string
}

// API レスポンスの data フィールドの型
export interface SuccessResponse<T> {
  success: true
  message: string
  data?: T
  timestamp: string
}
```

## コード分割 (React.lazy + Suspense)

ページコンポーネントは遅延読み込みにして初期バンドルサイズを削減する。

```tsx
// App.tsx
import { lazy, Suspense } from 'react'
import NProgress from 'nprogress'

const DashboardPage  = lazy(() => import('@/pages/DashboardPage'))
const TaskDetailPage = lazy(() => import('@/pages/TaskDetailPage'))
const AdminPage      = lazy(() => import('@/pages/AdminPage'))

// ページ遷移中は nprogress でプログレスバーを表示
<Suspense fallback={<PageLoader />}>
  <Routes>
    <Route path="/"          element={<DashboardPage />} />
    <Route path="/tasks/:id" element={<TaskDetailPage />} />
    <Route path="/admin"     element={<AdminPage />} />
  </Routes>
</Suspense>
```

## XSS 対策

```tsx
// Bad: ユーザー入力をそのまま HTML として埋め込む
<div dangerouslySetInnerHTML={{ __html: userInput }} />

// Good: テキストとして表示する (React がエスケープする)
<div>{userInput}</div>
```

- `dangerouslySetInnerHTML` は原則禁止
- 外部から取得した HTML を表示する必要がある場合は `dompurify` でサニタイズする

## レスポンシブ (Tailwind mobile-first)

Tailwind のブレークポイントは mobile-first で記述する。
まずモバイルのスタイルを書き、`sm:` `md:` `lg:` で上書きする。

| prefix | 幅 | 用途 |
|--------|----|------|
| (なし) | 0px〜 | モバイル |
| `sm:` | 640px〜 | タブレット |
| `md:` | 768px〜 | 小型PC |
| `lg:` | 1024px〜 | PC |

```tsx
// Bad: デスクトップ基準で書いてモバイルで上書き
<div className="grid grid-cols-3 sm:grid-cols-1">

// Good: モバイル基準で書いてデスクトップで上書き
<div className="grid grid-cols-1 md:grid-cols-3">
```

## DevTools (開発時のみ)

```tsx
// main.tsx - 開発時のみ DevTools を表示する
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'

<QueryClientProvider client={queryClient}>
  <App />
  {import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
</QueryClientProvider>
```

TanStack Query DevTools でキャッシュの状態・クエリの実行履歴を確認できる。

## React.memo / useMemo / useCallback

パフォーマンス最適化は計測してから行う。推測で使わない。

```
React.memo    → 親の再レンダリングで不要な再描画が起きている場合のみ
useMemo       → 重い計算処理 (ソート・フィルタ等) に使う。単純な値には使わない
useCallback   → useEffect / React.memo の依存関係に関数を渡す場合のみ
```

```tsx
// Bad: 不要な最適化
const value = useMemo(() => count * 2, [count])

// Good: フィルタ・ソートなど配列操作に使う
const filtered = useMemo(
  () => tasks.filter(t => t.status_id === statusId).sort(...),
  [tasks, statusId]
)
```
