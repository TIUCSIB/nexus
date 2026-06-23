import request from '@/utils/request'
import type { ApiResponse, PageResult, User } from '@/types'

export const listUsers = (params: { page?: number; page_size?: number; q?: string }) =>
  request.get('/admin/users', { params }) as Promise<ApiResponse<PageResult<User>>>

export const getUser = (id: number) =>
  request.get(`/admin/users/${id}`) as Promise<ApiResponse<User>>

export const createUser = (data: Partial<User> & { password: string }) =>
  request.post('/admin/users', data) as Promise<ApiResponse>

export const updateUser = (id: number, data: Partial<User>) =>
  request.put(`/admin/users/${id}`, data) as Promise<ApiResponse>

export const deleteUser = (id: number) =>
  request.delete(`/admin/users/${id}`) as Promise<ApiResponse>