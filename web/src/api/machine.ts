import request from '@/utils/request'
import type { ApiResponse, Machine, MachineCreateResult } from '@/types'

export function listMachines() {
  return request.get('/api/admin/machines') as Promise<ApiResponse<Machine[]>>
}

export function getMachineDetail(id: number) {
  return request.get(`/api/admin/machines/${id}`) as Promise<ApiResponse<{ machine: Machine; servers_count: number }>>
}

export function getMachineLoadHistory(id: number, params?: { limit?: number; range_hours?: number }) {
  return request.get(`/api/admin/machines/${id}/history`, { params }) as Promise<ApiResponse<any>>
}

export function createMachine(data: { name: string; notes?: string }) {
  return request.post('/api/admin/machines', data) as Promise<ApiResponse<MachineCreateResult>>
}

export function updateMachine(id: number, data: { name?: string; notes?: string; is_active?: boolean }) {
  return request.put(`/api/admin/machines/${id}`, data) as Promise<ApiResponse<boolean>>
}

export function deleteMachine(id: number) {
  return request.delete(`/api/admin/machines/${id}`) as Promise<ApiResponse<boolean>>
}

export function resetMachineToken(id: number) {
  return request.post(`/api/admin/machines/${id}/reset-token`) as Promise<ApiResponse<{ token: string }>>
}

export function getMachineInstallCommand(id: number) {
  return request.post(`/api/admin/machines/${id}/install-command`) as Promise<ApiResponse<{ command: string }>>
}
