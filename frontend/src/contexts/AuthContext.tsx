import { authApi } from '@/api/authApi'
import type { User } from '@/types/user'
import { tokenStorage } from '@/utils/tokenStorage'
import axios from 'axios'
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
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  // マウント時にトークンがあれば自分のプロフィールを取得する
  useEffect(() => {
    const token = tokenStorage.getAccess()
    if (!token) {
      setIsLoading(false)
      return
    }
    authApi
      .getMe()
      .then(setUser)
      .catch((error) => {
        // 401 のみトークンを削除する。ネットワーク障害や 5xx ではトークンを保持する
        if (axios.isAxiosError(error) && error.response?.status === 401) {
          tokenStorage.clear()
        }
      })
      .finally(() => setIsLoading(false))
  }, [])

  const login = async (email: string, password: string) => {
    const { access_token, refresh_token } = await authApi.login({ email, password })
    tokenStorage.setAccess(access_token)
    tokenStorage.setRefresh(refresh_token)
    const me = await authApi.getMe()
    setUser(me)
  }

  const register = async (name: string, email: string, password: string) => {
    // バックエンドの Register は access_token のみ返すため、
    // 続けて login を呼んで refresh_token も取得する
    await authApi.register({ name, email, password })
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
