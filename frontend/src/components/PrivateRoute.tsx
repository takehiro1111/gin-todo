import { LoadingSpinner } from '@/components/ui/LoadingSpinner'
import { useAuth } from '@/contexts/AuthContext'
import { Navigate, Outlet, useLocation } from 'react-router-dom'

/**
 * 認証が必要なルートを保護するガードコンポーネント。
 * 未認証時は /login にリダイレクトする。
 */
export function PrivateRoute() {
  const { user, isLoading } = useAuth()
  const location = useLocation()

  if (isLoading) return <LoadingSpinner />
  if (!user) return <Navigate to="/login" state={{ from: location }} replace />

  return <Outlet />
}
