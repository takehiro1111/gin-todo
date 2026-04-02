import { LoadingSpinner } from '@/components/ui/LoadingSpinner'
import { useAuth } from '@/contexts/AuthContext'
import { Navigate, Outlet } from 'react-router-dom'

/**
 * 管理者権限が必要なルートを保護するガードコンポーネント。
 * 非管理者はトップにリダイレクトする。
 */
export function AdminRoute() {
  const { user, isLoading } = useAuth()

  if (isLoading) return <LoadingSpinner />
  if (!user) return <Navigate to="/login" replace />
  // role_id === 1 が admin
  if (user.role_id !== 1) return <Navigate to="/" replace />

  return <Outlet />
}
