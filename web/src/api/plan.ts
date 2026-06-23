import request from '@/utils/request'
import type { ApiResponse, PageResult, Plan } from '@/types'

export function listPlans(params?: { page?: number; page_size?: number; keyword?: string }) {
  return request.get<any, ApiResponse<PageResult<Plan>>>('/plans', { params })
}

export function createPlan(data: Partial<Plan>) {
  return request.post<any, ApiResponse<Plan>>('/plans', data)
}

export function updatePlan(id: number, data: Partial<Plan>) {
  return request.put<any, ApiResponse<Plan>>(`/plans/${id}`, data)
}

export function deletePlan(id: number) {
  return request.delete<any, ApiResponse<null>>(`/plans/${id}`)
}
