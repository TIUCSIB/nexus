import request from '@/utils/request'
import type { ApiResponse, StatsOverview, TrafficDay, NodeRankingItem, UserRankingItem } from '@/types'

export const getOverview = () =>
  request.get('/api/admin/stats/overview') as Promise<ApiResponse<StatsOverview>>

export const getTraffic = (days = 7) =>
  request.get('/api/admin/stats/traffic', { params: { days } }) as Promise<ApiResponse<{ days: number; records: TrafficDay[] }>>

export const getNodeRanking = () =>
  request.get('/api/admin/stats/node-ranking') as Promise<ApiResponse<{ nodes: NodeRankingItem[] }>>

export const getUserRanking = (limit = 20) =>
  request.get('/api/admin/stats/user-ranking', { params: { limit } }) as Promise<ApiResponse<{ users: UserRankingItem[] }>>