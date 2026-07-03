import request from '@/utils/request'
import type { ApiResponse, PageResult, Node } from '@/types'

export const listNodes = (params?: { page?: number; page_size?: number; group_id?: string }) =>
  request.get('/api/admin/nodes', { params }) as Promise<ApiResponse<PageResult<Node>>>

export const getNode = (id: number) =>
  request.get(`/api/admin/nodes/${id}`) as Promise<ApiResponse<Node>>

export const createNode = (data: Partial<Node>) =>
  request.post('/api/admin/nodes', data) as Promise<ApiResponse>

export const updateNode = (id: number, data: Partial<Node>) =>
  request.put(`/api/admin/nodes/${id}`, data) as Promise<ApiResponse>

export const deleteNode = (id: number) =>
  request.delete(`/api/admin/nodes/${id}`) as Promise<ApiResponse>

export const restartNode = (id: number) =>
  request.post(`/api/admin/nodes/${id}/restart`) as Promise<ApiResponse>

export const resetNodeTraffic = (id: number) =>
  request.post(`/api/admin/nodes/${id}/reset-traffic`) as Promise<ApiResponse>

export const generateRealityKeys = () =>
  request.post('/api/admin/nodes/generate-reality-keys') as Promise<ApiResponse<{ private_key: string; public_key: string }>>
