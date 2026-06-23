import request from '@/utils/request'
import type { ApiResponse, PageResult, Node } from '@/types'

export function listNodes(params?: { page?: number; page_size?: number; keyword?: string; status?: string }) {
  return request.get<any, ApiResponse<PageResult<Node>>>('/nodes', { params })
}

export function createNode(data: Partial<Node>) {
  return request.post<any, ApiResponse<Node>>('/nodes', data)
}

export function updateNode(id: number, data: Partial<Node>) {
  return request.put<any, ApiResponse<Node>>(`/nodes/${id}`, data)
}

export function deleteNode(id: number) {
  return request.delete<any, ApiResponse<null>>(`/nodes/${id}`)
}
