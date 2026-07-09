import request from '@/utils/request'
import type { ApiResponse, User, SubscriptionInfo, UserStats, UserNode } from '@/types'

export const getProfile = () =>
  request.get('/api/user/profile') as Promise<ApiResponse<User>>

export const updateProfile = (data: { email?: string; password?: string }) =>
  request.put('/api/user/profile', data) as Promise<ApiResponse<User>>

export const getSubscription = () =>
  request.get('/api/user/subscription') as Promise<ApiResponse<SubscriptionInfo>>

export const getUserStats = () =>
  request.get('/api/user/stats') as Promise<ApiResponse<UserStats>>

export const listUserNodes = () =>
  request.get('/api/nodes') as Promise<ApiResponse<UserNode[]>>