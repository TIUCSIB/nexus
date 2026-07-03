import request from '@/utils/request'
import type { ApiResponse, PageResult, User } from '@/types'

export const listUsers = (params: { page?: number; page_size?: number; q?: string }) =>
  request.get('/api/admin/users', { params }) as Promise<ApiResponse<PageResult<User>>>

export interface UserDetail {
  user: User
  plan_name: string
  group_name: string
  ip_count: number
  links: string[]
}

export const getUser = (id: number) =>
  request.get(`/api/admin/users/${id}`) as Promise<ApiResponse<UserDetail>>

export const createUser = (data: Partial<User> & { password: string }) =>
  request.post('/api/admin/users', data) as Promise<ApiResponse>

export const updateUser = (id: number, data: Partial<User>) =>
  request.put(`/api/admin/users/${id}`, data) as Promise<ApiResponse>

export const deleteUser = (id: number) =>
  request.delete(`/api/admin/users/${id}`) as Promise<ApiResponse>

export const resetUserUUID = (id: number) =>
  request.post(`/api/admin/users/${id}/reset-uuid`) as Promise<ApiResponse<{ uuid: string; token: string }>>

export const resetUserTraffic = (id: number) =>
  request.post(`/api/admin/users/${id}/reset-traffic`) as Promise<ApiResponse>

export const getUserTrafficLogs = (id: number, params?: { page?: number; page_size?: number }) =>
  request.get(`/api/admin/users/${id}/traffic-logs`, { params }) as Promise<ApiResponse<PageResult<TrafficLog>>>

export interface TrafficLog {
  id: number
  user_id: number
  node_id: number
  upload: number
  download: number
  recorded_at: string
}
