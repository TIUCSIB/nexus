import request from '@/utils/request'
import type { ApiResponse, LoginResponse } from '@/types'

export function login(email: string, password: string) {
  return request.post<any, ApiResponse<LoginResponse>>('/auth/login', { email, password })
}

export function refreshToken(token: string) {
  return request.post<any, ApiResponse<LoginResponse>>('/auth/refresh', { token })
}
