import request from '@/utils/request'
import type { ApiResponse, OverviewStats, TrafficStats } from '@/types'

export function getOverview() {
  return request.get<any, ApiResponse<OverviewStats>>('/stats/overview')
}

export function getTraffic(days: number = 7) {
  return request.get<any, ApiResponse<TrafficStats[]>>('/stats/traffic', { params: { days } })
}
