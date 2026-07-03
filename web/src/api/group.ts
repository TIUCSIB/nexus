import request from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface ServerGroup {
  id: number
  name: string
  user_count?: number
  node_count?: number
  created_at: string
}

export const listGroups = () =>
  request.get('/api/admin/groups') as Promise<ApiResponse<ServerGroup[]>>

export const createGroup = (data: { name: string }) =>
  request.post('/api/admin/groups', data) as Promise<ApiResponse>

export const updateGroup = (id: number, data: { name: string }) =>
  request.put(`/api/admin/groups/${id}`, data) as Promise<ApiResponse>

export const deleteGroup = (id: number) =>
  request.delete(`/api/admin/groups/${id}`) as Promise<ApiResponse>