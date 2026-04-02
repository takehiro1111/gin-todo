import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import NProgress from 'nprogress'
import { Suspense, lazy, useEffect } from 'react'
import { Toaster } from 'react-hot-toast'
import { BrowserRouter, Route, Routes, useLocation } from 'react-router-dom'
import 'nprogress/nprogress.css'
import { AdminRoute } from '@/components/AdminRoute'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import { PrivateRoute } from '@/components/PrivateRoute'
import { LoadingSpinner } from '@/components/ui/LoadingSpinner'
import { AuthProvider } from '@/contexts/AuthContext'
import { ForgotPasswordPage } from '@/pages/ForgotPasswordPage'
import { LoginPage } from '@/pages/LoginPage'
import { NotFoundPage } from '@/pages/NotFoundPage'
import { RegisterPage } from '@/pages/RegisterPage'
import { ResetPasswordPage } from '@/pages/ResetPasswordPage'

// ページコンポーネントを遅延読み込みして初期バンドルサイズを削減
const DashboardPage = lazy(() =>
  import('@/pages/DashboardPage').then((m) => ({ default: m.DashboardPage }))
)
const TaskDetailPage = lazy(() =>
  import('@/pages/TaskDetailPage').then((m) => ({ default: m.TaskDetailPage }))
)
const AdminPage = lazy(() => import('@/pages/AdminPage').then((m) => ({ default: m.AdminPage })))

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 30, // 30秒
      retry: 1,
    },
  },
})

// ルート変更時に NProgress を表示する
function NProgressHandler() {
  const location = useLocation()

  // location オブジェクト全体を依存配列に入れてパス変更を検知する
  // biome-ignore lint/correctness/useExhaustiveDependencies: location.pathname の変化を検知するために意図的に使用
  useEffect(() => {
    NProgress.start()
    NProgress.done()
  }, [location.pathname])

  return null
}

function AppRoutes() {
  return (
    <>
      <NProgressHandler />
      <Routes>
        {/* 認証不要 */}
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />
        <Route path="/reset-password" element={<ResetPasswordPage />} />

        {/* 認証必須 */}
        <Route element={<PrivateRoute />}>
          <Route
            path="/"
            element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner />}>
                  <DashboardPage />
                </Suspense>
              </ErrorBoundary>
            }
          />
          <Route
            path="/tasks/:id"
            element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner />}>
                  <TaskDetailPage />
                </Suspense>
              </ErrorBoundary>
            }
          />
        </Route>

        {/* 管理者のみ */}
        <Route element={<AdminRoute />}>
          <Route
            path="/admin"
            element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner />}>
                  <AdminPage />
                </Suspense>
              </ErrorBoundary>
            }
          />
        </Route>

        {/* 404 */}
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </>
  )
}

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <AppRoutes />
          <Toaster position="top-right" />
        </BrowserRouter>
      </AuthProvider>
      {import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
    </QueryClientProvider>
  )
}
