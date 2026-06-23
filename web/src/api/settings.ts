import request from '@/utils/request'
import type { ApiResponse } from '@/types'

export const getSettings = () =>
  request.get('/admin/settings') as Promise<ApiResponse>

export const updateSettings = (data: Record<string, string>) =>
  request.put('/admin/settings', data) as Promise<ApiResponse>