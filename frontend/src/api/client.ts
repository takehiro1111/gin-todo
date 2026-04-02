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
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// レスポンスインターセプター: 401 時に refresh して再試行 (無限ループ防止)
apiClient.interceptors.response.use(
  (res) => res,
  async (error) => {
    const original = error.config

    // 401 かつ未リトライの場合のみリフレッシュを試みる
    if (error.response?.status === 401 && !original._retry) {
      original._retry = true
      const refreshToken = tokenStorage.getRefresh()
      if (!refreshToken) {
        // refresh_token がない場合はトークンを削除してエラーを返す。
        // リダイレクトは PrivateRoute に任せる (window.location.href は使わない)
        tokenStorage.clear()
        return Promise.reject(error)
      }
      try {
        // refresh_token を Bearer として送り VerifyUser を通過する
        const { data } = await axios.post(
          `${import.meta.env.VITE_API_BASE_URL}/api/auth/refresh`,
          { refresh_token: refreshToken },
          { headers: { Authorization: `Bearer ${refreshToken}` } }
        )
        const newAccessToken = data.data as string
        tokenStorage.setAccess(newAccessToken)
        original.headers.Authorization = `Bearer ${newAccessToken}`
        return apiClient(original)
      } catch {
        // リフレッシュ失敗もトークン削除のみ。リダイレクトは PrivateRoute に任せる
        tokenStorage.clear()
        return Promise.reject(error)
      }
    }

    return Promise.reject(error)
  }
)
