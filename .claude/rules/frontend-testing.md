---
paths:
  - "frontend/src/**/*.test.ts"
  - "frontend/src/**/*.test.tsx"
  - "frontend/src/**/*.spec.ts"
  - "frontend/src/**/*.spec.tsx"
---

# Testing Rules

## ツール

- **Vitest** — テストランナー
- **@testing-library/react** — コンポーネントテスト
- **@testing-library/user-event** — ユーザー操作シミュレーション
- **msw** (Mock Service Worker) — API モック

```bash
pnpm test           # ウォッチモード
pnpm test --run     # CI向け1回実行
pnpm coverage       # カバレッジ計測
```

## ファイル配置

テストファイルはテスト対象と同じディレクトリに置く。

```
components/task/
├── TaskCard.tsx
└── TaskCard.test.tsx
```

## 単体テスト (コンポーネント)

```tsx
import { render, screen } from '@testing-library/react'
import { TaskCard } from './TaskCard'

test('タイトルが表示される', () => {
  render(<TaskCard title="買い物" done={false} />)
  expect(screen.getByText('買い物')).toBeInTheDocument()
})

test('完了済みは視覚的に区別される', () => {
  render(<TaskCard title="買い物" done={true} />)
  expect(screen.getByRole('article')).toHaveClass('opacity-50')
})
```

## 結合テスト (hooks + API)

API 通信は msw でモックする。実装の内部ではなく **ユーザー視点の振る舞い** をテストする。
TanStack Query を使うフックは `QueryClientProvider` でラップする必要がある。

```tsx
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
import { useTasks } from './useTasks'

const server = setupServer(
  http.get('/api/task/', () => {
    return HttpResponse.json({ success: true, data: [{ id: 1, title: 'テスト' }] })
  })
)

beforeAll(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

// TanStack Query のラッパーを用意する
const wrapper = ({ children }: { children: React.ReactNode }) => (
  <QueryClientProvider client={new QueryClient()}>
    {children}
  </QueryClientProvider>
)

test('タスク一覧を取得できる', async () => {
  const { result } = renderHook(() => useTasks(), { wrapper })
  await waitFor(() => expect(result.current.isPending).toBe(false))
  // TanStack Query は data フィールドに結果を返す
  expect(result.current.data).toHaveLength(1)
})
```

## 方針

- `any` を使わない (テストコードも同様)
- 実装の内部詳細ではなく **ユーザーから見える振る舞い** をテストする
- モックは API 層 (msw) で止める。Context や hooks の内部はモックしない
