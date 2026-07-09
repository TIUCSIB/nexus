import request from '@/utils/request'
import type { ApiResponse, PageResult, CustomOutbound } from '@/types'

export const listCustomOutbounds = (params?: { page?: number; page_size?: number }) =>
  request.get('/api/admin/custom-outbounds', { params }) as Promise<ApiResponse<PageResult<CustomOutbound>>>

export const createCustomOutbound = (data: Partial<CustomOutbound>) =>
  request.post('/api/admin/custom-outbounds', data) as Promise<ApiResponse<CustomOutbound>>

export const updateCustomOutbound = (id: number, data: Partial<CustomOutbound>) =>
  request.put(`/api/admin/custom-outbounds/${id}`, data) as Promise<ApiResponse<CustomOutbound>>

export const deleteCustomOutbound = (id: number) =>
  request.delete(`/api/admin/custom-outbounds/${id}`) as Promise<ApiResponse>

export const getNodeOutbounds = (nodeId: number) =>
  request.get(`/api/admin/nodes/${nodeId}/outbounds`) as Promise<ApiResponse<{ outbound_ids: number[]; outbounds: CustomOutbound[] }>>

export const updateNodeOutbounds = (nodeId: number, outboundIds: number[]) =>
  request.put(`/api/admin/nodes/${nodeId}/outbounds`, { outbound_ids: outboundIds }) as Promise<ApiResponse>
