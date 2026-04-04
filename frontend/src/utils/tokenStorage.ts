// access token はメモリに保持する。localStorage に置かない（XSS 耐性向上）
// refresh token は HttpOnly Cookie でサーバーが管理するためここでは扱わない
let _accessToken: string | null = null

export const tokenStorage = {
  getAccess: () => _accessToken,
  setAccess: (t: string) => {
    _accessToken = t
  },
  clear: () => {
    _accessToken = null
  },
}
