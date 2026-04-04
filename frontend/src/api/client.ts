import { tokenStorage } from '@/utils/tokenStorage'
import axios from 'axios'

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  withCredentials: true,
  headers: { 'Content-Type': 'application/json' },
})

// リクエストインターセプター: Access Token を自動付与
apiClient.interceptors.request.use((config) => {
  const token = tokenStorage.getAccess()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// レスポンスインターセプター: 401 時に refresh して再試行
apiClient.interceptors.response.use(
  (res) => res,
  async (error) => {
    const original = error.config

    // 401 かつ未リトライの場合のみリフレッシュを試みる
    if (error.response?.status === 401 && !original._retry) {
      original._retry = true
      try {
        // refresh_token は HttpOnly Cookie で自動送信される
        const { data } = await axios.post(
          `${import.meta.env.VITE_API_BASE_URL}/api/auth/refresh`,
          {},
          { withCredentials: true }
        )

        // レスポンス形式が { data: { access_token: "..." } } か { data: "..." } かを判定して取得
        const newAccessToken =
          typeof data.data === 'string' ? data.data : data.data?.access_token

        if (!newAccessToken) {
          throw new Error('Access token not found in refresh response')
        }

        // 新しいトークンを保存してヘッダーを更新
        tokenStorage.setAccess(newAccessToken)
        original.headers.Authorization = `Bearer ${newAccessToken}`

        // 元のリクエストを再試行
        return apiClient(original)
      } catch (refreshError) {
        // リフレッシュ自体が失敗した場合は認証情報をクリア
        tokenStorage.clear()
        return Promise.reject(refreshError)
      }
    }

    return Promise.reject(error)
  }
)
