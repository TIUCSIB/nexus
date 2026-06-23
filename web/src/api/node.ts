import request from '@/utils/request'
import type { ApiResponse, PageResult, Node } from '@/types'

export const listNodes = (params?: { page?: number; page_size?: number }) =>
  request.get('/admin/nodes', { params }) as Promise<ApiResponse<PageResult<Node>>>

export const createNode = (data: Partial<Node>) =>
  request.post('/admin/nodes', data) as Promise<ApiResponse>

export const updateNode = (id: number, data: Partial<Node>) =>
  request.put(`/admin/nodes/${id}`, data) as Promise<ApiResponse>

export const deleteNode = (id: number) =>
  request.delete(`/admin/nodes/${id}`) as Promise<ApiResponse>

export const restartNode = (id: number) =>
  request.post(`/admin/nodes/${id}/restart`) as Promise<ApiResponse>