import request from '@/utils/request'
import type { ApiResponse, PageResult, User } from '@/types'

export function listUsers(params?: { page?: number; page_size?: number; keyword?: string; status?: string }) {
  return request.get<any, ApiResponse<PageResult<User>>>('/users', { params })
}

export function getUser(id: number) {
  return request.get<any, ApiResponse<User>>(`/users/${id}`)
}

export function createUser(data: Partial<User> & { password?: string }) {
  return request.post<any, ApiResponse<User>>('/users', data)
}

export function updateUser(id: number, data: Partial<User> & { password?: string }) {
  return request.put<any, ApiResponse<User>>(`/users/${id}`, data)
}

export function deleteUser(id: number) {
  return request.delete<any, ApiResponse<null>>(`/users/${id}`)
}
