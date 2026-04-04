export interface SuccessResponse<T = unknown> {
  success: true
  message: string
  data?: T
  timestamp: string
}

export interface ErrorResponse {
  success: false
  message: string
  error?: string
  code?: string
  timestamp: string
}
