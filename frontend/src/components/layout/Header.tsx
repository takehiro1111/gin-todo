import { useAuth } from '@/contexts/AuthContext'
import toast from 'react-hot-toast'
import { Link, useNavigate } from 'react-router-dom'

/**
 * アプリケーションのヘッダーナビゲーション。
 */
export function Header() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = async () => {
    try {
      await logout()
      navigate('/login')
      toast.success('ログアウトしました')
    } catch {
      toast.error('ログアウトに失敗しました')
    }
  }

  return (
    <header className="bg-white border-b border-gray-200 sticky top-0 z-10">
      <div className="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between">
        <Link to="/" className="text-xl font-bold text-blue-600">
          Todo
        </Link>
        <nav className="flex items-center gap-4">
          {user?.role_id === 1 && (
            <Link to="/admin" className="text-sm text-gray-600 hover:text-gray-900">
              管理
            </Link>
          )}
          <span className="text-sm text-gray-600">{user?.name}</span>
          <button
            type="button"
            onClick={handleLogout}
            className="text-sm text-red-600 hover:text-red-800"
          >
            ログアウト
          </button>
        </nav>
      </div>
    </header>
  )
}
