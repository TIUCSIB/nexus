import request from '@/utils/request'
import type { ApiResponse, SystemStatus } from '@/types'

export const getSystemStatus = () =>
  request.get('/api/admin/stats/system') as Promise<ApiResponse<SystemStatus>>