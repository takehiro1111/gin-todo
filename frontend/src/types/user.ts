export interface User {
  id: number
  name: string
  email: string
  role_id: number
  created_at: string
  updated_at: string
}

export interface JWTClaims {
  username: string
  user_id: number
  role: 'admin' | 'writer'
}
