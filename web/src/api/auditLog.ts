import request from '@/utils/request'
import type { ApiResponse, AuditLog, PageResult } from '@/types'

export const listAuditLogs = (params: { page?: number; page_size?: number; action?: string; start_date?: string; end_date?: string }) =>
  request.get('/api/admin/audit-logs', { params }) as Promise<ApiResponse<PageResult<AuditLog>>>