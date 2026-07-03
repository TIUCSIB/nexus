import request from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface OnlineIP {
  id: number
  user_id: number
  user_email: string
  user_uuid: string
  node_id: number
  node_name: string
  ips: string
  updated_at: string
}

export interface TrafficLogEntry {
  id: number
  user_id: number
  user_email: string
  node_id: number
  node_name: string
  upload: number
  download: number
  recorded_at: string
}

export const listOnlineIPs = () =>
  request.get('/api/admin/online-ips') as Promise<ApiResponse<OnlineIP[]>>

export const listTrafficLogs = (params: { page?: number; page_size?: number }) =>
  request.get('/api/admin/traffic-logs', { params }) as Promise<ApiResponse<{ items: TrafficLogEntry[]; total: number; page: number; page_size: number; total_pages: number }>>