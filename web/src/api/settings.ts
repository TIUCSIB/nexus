import request from '@/utils/request'
import type { ApiResponse, SystemSettings } from '@/types'

export function getSettings() {
  return request.get<any, ApiResponse<SystemSettings>>('/settings')
}

export function updateSettings(data: Partial<SystemSettings>) {
  return request.put<any, ApiResponse<SystemSettings>>('/settings', data)
}
