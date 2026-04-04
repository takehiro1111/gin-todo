import { LoadingSpinner } from '@/components/ui/LoadingSpinner'
import { useAuth } from '@/contexts/AuthContext'
import { Navigate, Outlet, useLocation } from 'react-router-dom'

/**
 * 管理者権限が必要なルートを保護するガードコンポーネント。
 * 未ログイン時は /login へ遷移し、ログイン後に元のページへ戻れるよう state を渡す。
 * 非管理者はトップにリダイレクトする。
 */
export function AdminRoute() {
  const { user, isLoading } = useAuth()
  const location = useLocation()

  if (isLoading) return <LoadingSpinner />
  if (!user) return <Navigate to="/login" state={{ from: location }} replace />
  // role_id === 1 が admin
  if (user.role_id !== 1) return <Navigate to="/" replace />

  return <Outlet />
}
