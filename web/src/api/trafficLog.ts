import request from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface TrafficLogEntry {
  id: number
  user_id: number
  node_id: number
  upload: number
  download: number
  recorded_at: string
  node_name: string
}

export interface OnlineIPEntry {
  id: number
  user_id: number
  ip: string
  node_id: number
  node_name: string
  updated_at: string
}

export const getUserTrafficLogs = (userId: number) =>
  request.get(`/api/admin/users/${userId}/traffic-logs`) as Promise<ApiResponse<TrafficLogEntry[]>>

export const getUserOnlineIPs = (userId: number) =>
  request.get(`/api/admin/users/${userId}/online-ips`) as Promise<ApiResponse<OnlineIPEntry[]>>

export const listAllTrafficLogs = (params: { page?: number; page_size?: number }) =>
  request.get('/api/admin/traffic-logs', { params }) as Promise<ApiResponse<{ items: TrafficLogEntry[]; total: number }>>

export const listAllOnlineIPs = () =>
  request.get('/api/admin/online-ips') as Promise<ApiResponse<OnlineIPEntry[]>>
