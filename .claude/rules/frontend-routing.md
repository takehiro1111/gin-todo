---
paths:
  - "frontend/src/**/*.tsx"
  - "frontend/src/**/*.ts"
---

# Routing & Auth Guard Rules

## ルート設計

```tsx
// App.tsx
<Routes>
  {/* 認証不要 */}
  <Route path="/login" element={<LoginPage />} />
  <Route path="/register" element={<RegisterPage />} />
  <Route path="/forgot-password" element={<ForgotPasswordPage />} />
  <Route path="/reset-password" element={<ResetPasswordPage />} />

  {/* 認証必須 */}
  <Route element={<PrivateRoute />}>
    <Route path="/" element={<DashboardPage />} />
    <Route path="/tasks/:id" element={<TaskDetailPage />} />
  </Route>

  {/* 管理者のみ */}
  <Route element={<AdminRoute />}>
    <Route path="/admin" element={<AdminPage />} />
  </Route>

  {/* 404 */}
  <Route path="*" element={<NotFoundPage />} />
</Routes>
```

## 404 ページ

存在しないパスへのアクセスは `NotFoundPage` で受け取る。

```tsx
// pages/NotFoundPage.tsx
import { Link } from 'react-router-dom'

/**
 * 存在しないパスへのアクセス時に表示する404ページ。
 */
export function NotFoundPage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-4">
      <h1 className="text-4xl font-bold">404</h1>
      <p className="text-gray-500">ページが見つかりません</p>
      <Link to="/" className="text-blue-500 hover:underline">
        トップへ戻る
      </Link>
    </div>
  )
}
```

## Error Boundary (レンダリングエラー)

JavaScript の例外がレンダリング中に発生すると画面が真っ白になる。
ページ単位で `ErrorBoundary` を設置してフォールバック UI を表示する。

```tsx
// components/ErrorBoundary.tsx
import { Component, type ReactNode } from 'react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false }

  static getDerivedStateFromError(): State {
    return { hasError: true }
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback ?? <ErrorPage />
    }
    return this.props.children
  }
}
```

```tsx
// App.tsx でページ単位に設置する
<Route
  path="/"
  element={
    <ErrorBoundary>
      <DashboardPage />
    </ErrorBoundary>
  }
/>
```

## エラーページ (API・予期せぬエラー)

API エラーや予期せぬ例外が発生した場合に表示する汎用エラーページ。

```tsx
// pages/ErrorPage.tsx
import { useNavigate } from 'react-router-dom'

/**
 * 予期せぬエラー発生時に表示する汎用エラーページ。
 */
export function ErrorPage({ message }: { message?: string }) {
  const navigate = useNavigate()

  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-4">
      <h1 className="text-4xl font-bold">エラーが発生しました</h1>
      <p className="text-gray-500">{message ?? '予期せぬエラーが発生しました'}</p>
      <button
        onClick={() => navigate('/')}
        className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
      >
        トップへ戻る
      </button>
    </div>
  )
}
```

## PrivateRoute パターン

```typescript
// components/PrivateRoute.tsx
export function PrivateRoute() {
  const { user, isLoading } = useAuth()

  if (isLoading) return <LoadingSpinner />
  if (!user) return <Navigate to="/login" replace />

  return <Outlet />
}
```

## 未ログイン時のリダイレクト

- 認証が必要なページへの直接アクセス → `/login` へリダイレクト
- ログイン済みで `/login` にアクセス → `/` へリダイレクト
- ログイン後は元のページに戻る (`state.from` を利用)

```typescript
// LoginPage.tsx
interface LocationState { from?: { pathname: string } }
const location = useLocation()
const from = (location.state as LocationState)?.from?.pathname ?? '/'
navigate(from, { replace: true })
```
