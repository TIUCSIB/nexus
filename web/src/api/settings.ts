import request from '@/utils/request'
import type { ApiResponse } from '@/types'

export const getSettings = () =>
  request.get('/api/admin/settings') as Promise<ApiResponse>

export const getSubscriptionTemplateDefaults = () =>
  request.get('/api/admin/settings/subscription-template-defaults') as Promise<ApiResponse>

export const updateSettings = (data: Record<string, string>) =>
  request.put('/api/admin/settings', { settings: data }) as Promise<ApiResponse>

export const getSiteInfo = () =>
  request.get('/api/site/info') as Promise<ApiResponse>

export const getBackupInfo = () =>
  request.get('/api/admin/settings/backup-info') as Promise<ApiResponse>

export const downloadBackup = async () => {
  const token = localStorage.getItem('token')
  const adminPath = localStorage.getItem('admin_path') || 'admin'
  const res = await fetch(`/api/${adminPath}/settings/backup`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: '备份失败' }))
    throw { response: { data: err } }
  }
  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `nexus_backup_${new Date().toISOString().slice(0, 10)}.db`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}
