import request from '@/utils/request'
import type { ApiResponse } from '@/types'

export const login = (data: { email: string; password: string }) =>
  request.post('/auth/login', data) as Promise<ApiResponse>

export const refreshToken = (data: { refresh_token: string }) =>
  request.post('/auth/refresh', data) as Promise<ApiResponse>