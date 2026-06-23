import request from '@/utils/request'
import type { ApiResponse, PageResult, Plan } from '@/types'

export const listPlans = (params?: { page?: number; page_size?: number }) =>
  request.get('/admin/plans', { params }) as Promise<ApiResponse<PageResult<Plan>>>

export const createPlan = (data: Partial<Plan>) =>
  request.post('/admin/plans', data) as Promise<ApiResponse>

export const updatePlan = (id: number, data: Partial<Plan>) =>
  request.put(`/admin/plans/${id}`, data) as Promise<ApiResponse>

export const deletePlan = (id: number) =>
  request.delete(`/admin/plans/${id}`) as Promise<ApiResponse>