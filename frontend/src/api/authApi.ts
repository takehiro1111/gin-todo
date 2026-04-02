import { apiClient } from '@/api/client'
import type { User } from '@/types/user'
import axios from 'axios'

interface LoginBody {
  email: string
  password: string
}

interface RegisterBody {
  name: string
  email: string
  password: string
}

interface AuthTokens {
  access_token: string
  refresh_token: string
}

/**
 * 認証関連 API
 */
export const authApi = {
  login: async (body: LoginBody): Promise<AuthTokens> => {
    const { data } = await axios.post<{ success: true; data: AuthTokens }>(
      `${import.meta.env.VITE_API_BASE_URL}/api/auth/login`,
      body
    )
    return data.data
  },

  // バックエンドの Register は access_token のみ返す (refresh_token なし)
  register: async (body: RegisterBody): Promise<string> => {
    const { data } = await axios.post<{ success: true; data: string }>(
      `${import.meta.env.VITE_API_BASE_URL}/api/auth/register`,
      body
    )
    return data.data
  },

  logout: async (): Promise<void> => {
    await apiClient.post('/api/auth/logout')
  },

  getMe: async (): Promise<User> => {
    const { data } = await apiClient.get('/api/auth/me')
    return data.data as User
  },

  forgotPassword: async (email: string): Promise<void> => {
    await axios.post(`${import.meta.env.VITE_API_BASE_URL}/api/auth/forgot`, { email })
  },

  resetPassword: async (resetToken: string, newPassword: string): Promise<void> => {
    await axios.post(`${import.meta.env.VITE_API_BASE_URL}/api/auth/reset`, {
      reset_token: resetToken,
      new_password: newPassword,
    })
  },

  changePassword: async (oldPassword: string, newPassword: string): Promise<void> => {
    await apiClient.patch('/api/auth/password', {
      old_password: oldPassword,
      new_password: newPassword,
    })
  },
}
