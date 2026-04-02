import { authApi } from '@/api/authApi'
import type { User } from '@/types/user'
import { tokenStorage } from '@/utils/tokenStorage'
import { type ReactNode, createContext, useContext, useEffect, useState } from 'react'

interface AuthContextValue {
  user: User | null
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (name: string, email: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)

/**
 * 認証状態を管理するプロバイダー。
 * - access token はメモリ管理 (tokenStorage)
 * - refresh token は HttpOnly Cookie でサーバーが管理
 * - マウント時に /refresh を叩いて access token を復元する
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  // マウント時に Cookie の refresh_token で access_token を復元する
  useEffect(() => {
    authApi
      .refresh()
      .then((accessToken) => {
        tokenStorage.setAccess(accessToken)
        return authApi.getMe()
      })
      .then(setUser)
      .catch(() => {
        // refresh_token Cookie がない or 期限切れ = 未ログイン。エラーは無視
      })
      .finally(() => setIsLoading(false))
  }, [])

  const login = async (email: string, password: string) => {
    const accessToken = await authApi.login({ email, password })
    tokenStorage.setAccess(accessToken)
    const me = await authApi.getMe()
    setUser(me)
  }

  const register = async (name: string, email: string, password: string) => {
    await authApi.register({ name, email, password })
    // register 後に login を呼ぶことで refresh_token Cookie がセットされる
    await login(email, password)
  }

  const logout = async () => {
    try {
      await authApi.logout()
    } finally {
      tokenStorage.clear()
      setUser(null)
    }
  }

  return (
    <AuthContext.Provider value={{ user, isLoading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

/**
 * 認証コンテキストを取得するフック。
 */
export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
