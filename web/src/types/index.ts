export interface User {
  id: number
  email: string
  uuid: string
  plan_id: number | null
  plan_name?: string
  traffic_used: number
  traffic_limit: number
  status: 'active' | 'disabled' | 'expired'
  expire_at: string | null
  created_at: string
  updated_at: string
}

export interface Plan {
  id: number
  name: string
  traffic_limit: number
  duration_days: number
  price: number
  description: string
  node_ids: number[]
  created_at: string
  updated_at: string
}

export interface Node {
  id: number
  name: string
  address: string
  port: number
  protocol: 'vmess' | 'vless' | 'trojan' | 'shadowsocks'
  status: 'online' | 'offline' | 'maintenance'
  config_mode: string
  traffic_rate: number
  sort_order: number
  created_at: string
  updated_at: string
}

export interface TrafficLog {
  id: number
  user_id: number
  node_id: number
  upload: number
  download: number
  recorded_at: string
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export interface LoginResponse {
  token: string
  user: User
}

export interface OverviewStats {
  total_users: number
  active_users: number
  total_traffic: number
  total_nodes: number
  online_nodes: number
  today_income: number
  month_income: number
}

export interface TrafficStats {
  date: string
  upload: number
  download: number
}

export interface SystemSettings {
  site_name: string
  site_url: string
  admin_email: string
  register_enabled: boolean
  default_traffic_limit: number
  default_plan_id: number | null
  jwt_secret: string
  invite_only: boolean
}
