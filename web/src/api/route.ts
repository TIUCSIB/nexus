import request from '@/utils/request'
import type { ApiResponse, PageResult } from '@/types'

export interface RouteRule {
  id: number
  name: string
  match: string
  action: string
  action_value: string
  match_json: string
  action_json: string
  sort: number
  status: number
  created_at: string
}

export const listRoutes = (params?: { page?: number; page_size?: number }) =>
  request.get('/api/admin/routes', { params }) as Promise<ApiResponse<PageResult<RouteRule>>>

export const createRoute = (data: Partial<RouteRule>) =>
  request.post('/api/admin/routes', data) as Promise<ApiResponse>

export const updateRoute = (id: number, data: Partial<RouteRule>) =>
  request.put(`/api/admin/routes/${id}`, data) as Promise<ApiResponse>

export const deleteRoute = (id: number) =>
  request.delete(`/api/admin/routes/${id}`) as Promise<ApiResponse>