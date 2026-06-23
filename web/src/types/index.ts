export interface User {
  id: number
  uuid: string
  email: string
  balance: number
  plan_id: number | null
  traffic_used: number
  traffic_limit: number
  expired_at: string | null
  is_admin: boolean
  token: string
  status: number
  device_limit: number
  speed_limit_up: number
  speed_limit_down: number
  created_at: string
  updated_at: string
}

export interface Plan {
  id: number
  name: string
  description: string
  traffic_limit: number
  duration_days: number
  price: number
  sort: number
  status: number
  created_at: string
}

export interface Node {
  id: number
  name: string
  address: string
  protocol: string
  port: number
  config_mode: string
  config_json: string
  online: boolean
  last_heartbeat: string | null
  sort: number
  status: number
  created_at: string
  updated_at: string
}

export interface PageResult<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}