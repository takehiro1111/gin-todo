---
paths:
  - "frontend/src/**/*.tsx"
  - "frontend/src/**/*.ts"
---

# Routing & Auth Guard Rules

## ルート設計

```typescript
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
</Routes>
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
const location = useLocation()
const from = (location.state as any)?.from?.pathname ?? '/'
navigate(from, { replace: true })
```
