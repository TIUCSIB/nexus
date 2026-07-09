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
  online?: boolean
  last_seen?: string | null
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

export interface CertConfig {
  cert_mode: string      // none/file/content/self/http/dns
  domain: string
  email: string
  dns_provider: string
  dns_env: Record<string, string>
  http_port: number
  cert_file: string
  key_file: string
  cert_content: string
  key_content: string
  cert_dir: string
}

export interface CustomOutbound {
  id: number
  name: string
  tag: string
  protocol: string
  settings_json: string
  proxy_tag: string
  sort: number
  status: number
  created_at: string
  updated_at: string
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
	  group_ids: number[]
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
  cert_config: string
  kernel_type: string
  custom_outbounds: string
  online: boolean
  last_heartbeat: string | null
  sort: number
  status: number
  machine_id: number | null
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
  match_json: string
  action_json: string
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
  available?: boolean
  unavailable_reason?: string
}

export interface StatsOverview {
  total_users: number
  total_nodes: number
  online_nodes: number
  total_traffic: number
  total_upload: number
  total_download: number
  today_upload: number
  today_download: number
  today_traffic: number
  online_devices: number
  online_users: number
  monthly_traffic: number
  yesterday_ranking: YesterdayNodeRanking[]
}

export interface YesterdayNodeRanking {
  node_id: number
  name: string
  total: number
  upload: number
  download: number
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

export interface Machine {
  id: number
  name: string
  notes: string
  is_active: boolean
  last_seen_at: string | null
  load_status: LoadStatus | null
  servers_count: number
  created_at: string
  updated_at: string
}

export interface LoadStatus {
  cpu: number
  mem_total: number
  mem_used: number
  disk_total: number
  disk_used: number
  net_in_speed: number
  net_out_speed: number
}

export interface MachineLoadHistory {
  id: number
  machine_id: number
  cpu: number
  mem_total: number
  mem_used: number
  disk_total: number
  disk_used: number
  net_in_speed: number
  net_out_speed: number
  recorded_at: number
}

export interface MachineCreateResult {
  id: number
  name: string
  token: string
  notes: string
  is_active: boolean
  install_command: string
}

export interface UserStats {
  total_traffic: number
  total_upload: number
  total_download: number
  today_upload: number
  today_download: number
  monthly_upload: number
  monthly_download: number
  node_traffic: Array<{ node_id: number; node_name: string; upload: number; download: number }>
  daily_traffic: Array<{ date: string; upload: number; download: number }>
}

export interface TrafficResetUser {
  id: number
  email: string
  plan_id: number | null
  plan_name: string
  plan_traffic_reset: number
  traffic_used: number
  traffic_limit: number
  traffic_reset_at: string | null
  expired_at: string | null
  status: number
}

export interface TrafficResetStats {
  today_reset: number
  month_reset: number
  total_reset: number
  by_operator: Array<{ operator: string; count: number }>
}

export interface AuditLog {
  id: number
  user_id: number
  user_email: string
  action: string
  target: string
  detail: string
  ip: string
  created_at: string
}

// UserNode 用户端节点列表（安全字段，不含敏感配置）
export interface UserNode {
  id: number
  name: string
  address: string
  protocol: string
  port: number
  service_port: number
  rate: number
  tags: string
  online: boolean
  last_heartbeat: string | null
  online_count: number
  traffic_used: number
  traffic_limit: number
  sort: number
  status: number
}

export interface SystemStatus {
	version: string
	go_version: string
	uptime: string
	db_size: number
	db_size_human: string
	total_users: number
	active_users: number
	total_nodes: number
	online_nodes: number
	online_devices: number
	online_users: number
	today_traffic: number
	start_time: string
}

export interface NodeRankingItem {
	id: number
	name: string
	address: string
	protocol: string
	traffic_used: number
	traffic_limit: number
	online: boolean
	online_count: number
}

export interface UserRankingItem {
	id: number
	email: string
	uuid: string
	traffic_used: number
	traffic_limit: number
	upload_used: number
	download_used: number
	device_limit: number
}

export interface LoginResponseData {
  token: string
  refresh_token: string
  user: User
  admin_path: string
  auth_path: string
  user_path: string
  app_name: string
  app_description: string
  sub_path: string
}

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}