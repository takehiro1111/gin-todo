export const tokenStorage = {
  getAccess: () => localStorage.getItem('accessToken'),
  setAccess: (t: string) => localStorage.setItem('accessToken', t),
  getRefresh: () => localStorage.getItem('refreshToken'),
  setRefresh: (t: string) => localStorage.setItem('refreshToken', t),
  clear: () => {
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
  },
}
