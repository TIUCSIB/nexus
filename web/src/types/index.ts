export interface NetworkSettings {
  path?: string
  host?: string
  headers?: Record<string, string>
  service_name?: string
  server_name?: string
  allow_insecure?: boolean
  reality_server_name?: string
  reality_port?: number
  reality_private_key?: string
  reality_public_key?: string
  reality_short_id?: string
  utls_enabled?: boolean
  utls_fingerprint?: string
  version?: number
  obfs_open?: boolean
  obfs_type?: string
  obfs_password?: string
  bandwidth_up?: number
  bandwidth_down?: number
  hop_interval?: number
  alpn?: string
  tls_server_name?: string
  tls_allow_insecure?: boolean
  ech_enabled?: boolean
  tuic_version?: number
  congestion_control?: string
  udp_relay_mode?: string
  zero_rtt?: boolean
  heartbeat?: string
}

export interface User {
  id: number
  uuid: string
  email: string
  balance: number
  plan_id: number | null
  group_id: number | null
  traffic_used: number
  traffic_limit: number
  traffic_reset_at: string | null
  expired_at: string | null
  is_admin: boolean
  token: string
  status: number
  device_limit: number
  speed_limit_up: number
  speed_limit_down: number
  upload_used: number
  download_used: number
  remarks: string
  created_at: string
  updated_at: string
}

export interface Plan {
  id: number
  name: string
  description: string
  group_id: number | null
  traffic_limit: number
  duration_days: number
  price: number
  speed_limit: number
  device_limit: number
  capacity_limit: number
  traffic_reset: number
  sort: number
  status: number
  created_at: string
}

export interface Node {
  id: number
  custom_id: string
  name: string
  address: string
  protocol: string
  port: number
  service_port: number
  group_id: number | null
  route_id: number | null
  rate: number
  dynamic_rate: boolean
  tags: string
  traffic_limit: number
  traffic_used: number
  online_count: number
  parent_id: number | null
  security: string
  transport: string
  flow_control: string
  vless_encryption: boolean
  config_mode: string
  network_settings: string
  config_json: string
  online: boolean
  last_heartbeat: string | null
  sort: number
  status: number
  created_at: string
  updated_at: string
}

export interface ServerGroup {
  id: number
  name: string
}

export interface RouteRule {
  id: number
  name: string
  match: string
  action: string
  action_value: string
  sort: number
  status: number
}

export interface SubscriptionInfo {
  plan_id: number | null
  plan_name?: string
  traffic_used: number
  traffic_limit: number
  expired_at: string | null
  token: string
  links: string[]
  clean_links?: string[]
  sub_path?: string
}

export interface StatsOverview {
  total_users: number
  total_nodes: number
  online_nodes: number
  total_traffic: number
}

export interface TrafficDay {
  date: string
  upload: number
  download: number
  total: number
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