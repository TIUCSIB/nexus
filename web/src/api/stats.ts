import request from '@/utils/request'
import type { ApiResponse } from '@/types'

export const getOverview = () =>
  request.get('/admin/stats/overview') as Promise<ApiResponse>

export const getTraffic = (days = 7) =>
  request.get('/admin/stats/traffic', { params: { days } }) as Promise<ApiResponse>