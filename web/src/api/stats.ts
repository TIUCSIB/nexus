import request from '@/utils/request'
import type { ApiResponse, StatsOverview, TrafficDay } from '@/types'

export const getOverview = () =>
  request.get('/api/admin/stats/overview') as Promise<ApiResponse<StatsOverview>>

export const getTraffic = (days = 7) =>
  request.get('/api/admin/stats/traffic', { params: { days } }) as Promise<ApiResponse<{ days: number; records: TrafficDay[] }>>