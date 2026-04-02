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

/**
 * 認証関連 API
 */
export const authApi = {
  login: async (body: LoginBody): Promise<string> => {
    const { data } = await axios.post<{ success: true; data: { access_token: string } }>(
      `${import.meta.env.VITE_API_BASE_URL}/api/auth/login`,
      body,
      { withCredentials: true } // refresh_token Cookie を受け取るために必須
    )
    return data.data.access_token
  },

  // refresh_token Cookie を使って新しい access_token を取得する
  refresh: async (): Promise<string> => {
    const { data } = await axios.post<{ success: true; data: string }>(
      `${import.meta.env.VITE_API_BASE_URL}/api/auth/refresh`,
      {},
      { withCredentials: true }
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
