import request from '@/utils/request'
import type { ApiResponse, TrafficResetUser, TrafficResetStats, PageResult } from '@/types'

export const getTrafficResetUsers = (params: { page?: number; page_size?: number; q?: string }) =>
  request.get('/api/admin/traffic-reset/users', { params }) as Promise<ApiResponse<PageResult<TrafficResetUser>>>

export const manualTrafficReset = () =>
  request.post('/api/admin/traffic-reset/manual') as Promise<ApiResponse<{ message: string; reset_count: number }>>

export const getTrafficResetStats = () =>
  request.get('/api/admin/traffic-reset/stats') as Promise<ApiResponse<TrafficResetStats>>