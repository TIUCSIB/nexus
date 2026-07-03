import request from '@/utils/request'
import type { ApiResponse } from '@/types'

export const getSettings = () =>
  request.get('/api/admin/settings') as Promise<ApiResponse>

export const updateSettings = (data: Record<string, string>) =>
  request.put('/api/admin/settings', { settings: data }) as Promise<ApiResponse>

export const getSiteInfo = () =>
  request.get('/api/site/info') as Promise<ApiResponse>