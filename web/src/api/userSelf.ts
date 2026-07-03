import request from '@/utils/request'
import type { ApiResponse, User, SubscriptionInfo } from '@/types'

export const getProfile = () =>
  request.get('/api/user/profile') as Promise<ApiResponse<User>>

export const updateProfile = (data: { email?: string; password?: string }) =>
  request.put('/api/user/profile', data) as Promise<ApiResponse<User>>

export const getSubscription = () =>
  request.get('/api/user/subscription') as Promise<ApiResponse<SubscriptionInfo>>