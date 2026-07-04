<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Separator } from '@/components/ui/separator'
import { Plus, MoreHorizontal, Pencil, Trash2, RotateCcw, Settings, Copy } from 'lucide-vue-next'
import { RefreshCw, SlidersHorizontal } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { listNodes, createNode, updateNode, deleteNode, restartNode, resetNodeTraffic } from '@/api/node'
import { generateRealityKeys } from '@/api/node'
import { listGroups } from '@/api/group'
import { listRoutes } from '@/api/route'
import type { Node } from '@/types'
import type { NetworkSettings } from '@/types'
import type { ServerGroup } from '@/api/group'
import type { RouteRule } from '@/api/route'

const nodes = ref<Node[]>([])
const groups = ref<ServerGroup[]>([])
const routes = ref<RouteRule[]>([])

const page = ref(1)
const dialogOpen = ref(false)
const total = ref(0)
const deleteDialogOpen = ref(false)
const resetTrafficDialogOpen = ref(false)
const editing = ref<Partial<Node>>({})
const isEdit = ref(false)
const saving = ref(false)

/* ── 高级设置弹窗 ── */
const advancedOpen = ref(false)
const advancedTab = ref('tls')
const advancedTls = ref({
  cert_mode: 'none',
  domain: '',
  email: '',
  http_port: '80',
  dns_provider: '',
  dns_env: '',
  cert_content: '',
  key_content: '',
})
const advancedOutbounds = ref('')
const advancedRoutes = ref('')
const multiplex = ref({
  enabled: false,
  protocol: 'smux',
  padding: false,
  max_connections: 16,
  min_streams: 4,
  brutal_enabled: false,
  brutal_up_mbps: 100,
  brutal_down_mbps: 500,
})

function openAdvanced() {
  try {
    const raw = editing.value.network_settings || '{}'
    const parsed = JSON.parse(raw)
    advancedTls.value = {
      cert_mode: parsed.tls_cert_mode || 'none',
      domain: parsed.tls_domain || '',
      email: parsed.tls_email || '',
      http_port: parsed.tls_http_port || '80',
      dns_provider: parsed.tls_dns_provider || '',
      dns_env: parsed.tls_dns_env || '',
      cert_content: parsed.tls_cert_content || '',
      key_content: parsed.tls_key_content || '',
    }
    advancedOutbounds.value = parsed.custom_outbounds
      ? (typeof parsed.custom_outbounds === 'string' ? parsed.custom_outbounds : JSON.stringify(parsed.custom_outbounds, null, 2))
      : ''
    advancedRoutes.value = parsed.custom_routes
      ? (typeof parsed.custom_routes === 'string' ? parsed.custom_routes : JSON.stringify(parsed.custom_routes, null, 2))
      : ''
    multiplex.value = {
      enabled: parsed.multiplex_enabled || false,
      protocol: parsed.multiplex_protocol || 'smux',
      padding: parsed.multiplex_padding || false,
      max_connections: parsed.multiplex_max_connections || 16,
      min_streams: parsed.multiplex_min_streams || 4,
      brutal_enabled: parsed.multiplex_brutal_enabled || false,
      brutal_up_mbps: parsed.multiplex_brutal_up_mbps || 100,
      brutal_down_mbps: parsed.multiplex_brutal_down_mbps || 500,
    }
  } catch {
    advancedTls.value = { cert_mode: 'none', domain: '', email: '', http_port: '80', dns_provider: '', dns_env: '', cert_content: '', key_content: '' }
    advancedOutbounds.value = ''
    advancedRoutes.value = ''
    multiplex.value = { enabled: false, protocol: 'smux', padding: false, max_connections: 16, min_streams: 4, brutal_enabled: false, brutal_up_mbps: 100, brutal_down_mbps: 500 }
  }
  advancedTab.value = 'tls'
  advancedOpen.value = true
}

function saveAdvanced() {
  try {
    const settings = JSON.parse(editing.value.network_settings || '{}')
    settings.tls_cert_mode = advancedTls.value.cert_mode
    if (advancedTls.value.cert_mode !== 'none') {
      settings.tls_domain = advancedTls.value.domain
    }
    if (advancedTls.value.cert_mode === 'http-01' || advancedTls.value.cert_mode === 'dns-01') {
      settings.tls_email = advancedTls.value.email
    }
    if (advancedTls.value.cert_mode === 'http-01' || advancedTls.value.cert_mode === 'dns-01') {
      settings.tls_http_port = advancedTls.value.http_port
    }
    if (advancedTls.value.cert_mode === 'dns-01') {
      settings.tls_dns_provider = advancedTls.value.dns_provider
      settings.tls_dns_env = advancedTls.value.dns_env
    }
    if (advancedTls.value.cert_mode === 'manual') {
      settings.tls_cert_content = advancedTls.value.cert_content
      settings.tls_key_content = advancedTls.value.key_content
    }
    settings.multiplex_enabled = multiplex.value.enabled
    if (multiplex.value.enabled) {
      settings.multiplex_protocol = multiplex.value.protocol
      settings.multiplex_padding = multiplex.value.padding
      settings.multiplex_max_connections = multiplex.value.max_connections
      settings.multiplex_min_streams = multiplex.value.min_streams
      settings.multiplex_brutal_enabled = multiplex.value.brutal_enabled
      if (multiplex.value.brutal_enabled) {
        settings.multiplex_brutal_up_mbps = multiplex.value.brutal_up_mbps
        settings.multiplex_brutal_down_mbps = multiplex.value.brutal_down_mbps
      }
    }
    if (advancedOutbounds.value.trim()) {
      try { settings.custom_outbounds = JSON.stringify(JSON.parse(advancedOutbounds.value)) } catch { settings.custom_outbounds = advancedOutbounds.value }
    } else {
      delete settings.custom_outbounds
    }
    if (advancedRoutes.value.trim()) {
      try { settings.custom_routes = JSON.stringify(JSON.parse(advancedRoutes.value)) } catch { settings.custom_routes = advancedRoutes.value }
    } else {
      delete settings.custom_routes
    }
    editing.value.network_settings = JSON.stringify(settings)
    advancedOpen.value = false
    toast.success('高级设置已保存')
  } catch {
    toast.error('保存失败')
  }
}

/* ── 传输协议配置 ── */
const vlessTransports = [
  { value: 'tcp', label: 'TCP' },
  { value: 'ws', label: 'WebSocket' },
  { value: 'grpc', label: 'gRPC' },
  { value: 'http', label: 'HTTP' },
  { value: 'h2', label: 'HTTP/2' },
  { value: 'httpupgrade', label: 'HTTPUpgrade' },
  { value: 'xhttp', label: 'XHTTP' },
]

/* ── Reality 密钥对生成 ── */

/* ── 编辑协议弹窗 ── */
const protocolEditOpen = ref(false)
const protocolJson = ref('')

const transportTemplates: Record<string, { label: string; json: string }[]> = {
  tcp: [
    { label: '使用TCP模板', json: '{"header":{"type":"none"}}' },
    { label: '使用TCP + HTTP模板', json: '{"header":{"type":"http","request":{"path":["/"],"headers":{"Host":["www.example.com"]}}}}' },
  ],
  ws: [
    { label: '使用WebSocket模板', json: '{"path":"/ws","headers":{"Host":"www.example.com"}}' },
  ],
  grpc: [
    { label: '使用gRPC模板', json: '{"serviceName":"grpc-service"}' },
  ],
  http: [
    { label: '使用HTTP模板', json: '{"path":"/","host":"www.example.com"}' },
  ],
  h2: [
    { label: '使用HTTP/2模板', json: '{"path":"/","host":"www.example.com"}' },
  ],
  httpupgrade: [
    { label: '使用HTTPUpgrade模板', json: '{"path":"/httpupgrade","host":"www.example.com","headers":{}}' },
  ],
  xhttp: [
    { label: '使用XHTTP模板', json: '{"path":"/xhttp","host":"www.example.com"}' },
  ],
  hysteria2: [
    { label: '使用Hysteria2模板', json: '{"version":2,"bandwidth":{"up":100,"down":500}}' },
    { label: '使用Hysteria2 + 混淆', json: '{"version":2,"bandwidth":{"up":100,"down":500},"obfs":{"open":true,"type":"salamander","password":"changeme"}}' },
  ],
  tuic: [
    { label: '使用TUIC模板', json: '{"congestion_control":"cubic","udp_relay_mode":"native"}' },
  ],
}

function openProtocolEdit() {
  try {
    const raw = editing.value.network_settings || '{}'
    const parsed = JSON.parse(raw)
    protocolJson.value = JSON.stringify(parsed, null, 2)
  } catch {
    protocolJson.value = editing.value.network_settings || '{}'
  }
  protocolEditOpen.value = true
}

function applyTransportTemplate(templateJson: string) {
  protocolJson.value = templateJson
}

function saveProtocolEdit() {
  try {
    JSON.parse(protocolJson.value)
    editing.value.network_settings = protocolJson.value
    protocolEditOpen.value = false
    toast.success('协议配置已更新')
  } catch {
    toast.error('JSON 格式错误，请检查后重试')
  }
}

async function handleGenerateRealityKeys() {
  try {
    const res = await generateRealityKeys()
    if (res.code === 0) {
      netSettings.value.reality_private_key = res.data.private_key
      netSettings.value.reality_public_key = res.data.public_key
      toast.success('密钥对已生成')
    } else {
      toast.error(res.message || '生成失败')
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.message || '生成密钥对失败')
  }
}

const netSettings = ref<NetworkSettings>({})
const tuicAlpnList = ref<string[]>([])
const tuicAlpnAdd = ref('')

function syncAlpnFromSettings() {
  const alpn = netSettings.value.alpn
  if (typeof alpn === 'string' && alpn) {
    tuicAlpnList.value = alpn.split(',').map(s => s.trim()).filter(Boolean)
  } else if (Array.isArray(alpn)) {
    tuicAlpnList.value = [...alpn]
  } else {
    tuicAlpnList.value = []
  }
}

function addTuicAlpn(val: any) {
  if (val && typeof val === 'string' && !tuicAlpnList.value.includes(val)) {
    tuicAlpnList.value.push(val)
    netSettings.value.alpn = tuicAlpnList.value.join(',')
  }
  tuicAlpnAdd.value = ''
}

function removeTuicAlpn(idx: number) {
  tuicAlpnList.value.splice(idx, 1)
  netSettings.value.alpn = tuicAlpnList.value.join(',')
}

function parseNetSettings(raw: string | undefined): NetworkSettings {
  if (!raw) return {}
  try { return JSON.parse(raw) } catch { return {} }
}

function buildNetSettingsString(): string {
  const clean: Record<string, any> = {}
  for (const [k, v] of Object.entries(netSettings.value)) {
    if (v !== undefined && v !== null && v !== '') clean[k] = v
  }
  return Object.keys(clean).length ? JSON.stringify(clean) : ''
}

/* 根据传输协议自动设置 host 默认值 */
watch(() => editing.value.protocol, (proto) => {
  if (proto === 'vless' && editing.value.address && !netSettings.value.host) {
    netSettings.value.host = editing.value.address
  }
})

/* ── 工具函数 ── */
function formatBytes(b: number | undefined | null) {
  if (!b || b === 0) return '0 B'
  const k = 1024; const s = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(k))
  return (b / Math.pow(k, i)).toFixed(2) + ' ' + s[i]
}

function parseTags(tags: string | undefined | null) {
  if (!tags) return []
  return tags.split(',').map(t => t.trim()).filter(Boolean)
}

function getGroupName(id: number | null | undefined) {
  if (!id) return ''
  const g = groups.value.find(g => g.id === id)
  return g ? g.name : ''
}

function getRouteName(id: number | null | undefined) {
  if (!id) return ''
  const r = routes.value.find(r => r.id === id)
  return r ? r.name : ''
}

/* 节点状态：配置异常→红色，离线→红色，无人使用→黄色，运行正常→绿色 */
function isNodeConfigValid(n: Node): boolean {
  if (!n.address) return false
  if (!n.port || n.port < 1 || n.port > 65535) return false
  return true
}

function getNodeStatus(n: Node): { color: string; label: string } {
  if (!isNodeConfigValid(n)) return { color: 'bg-red-500', label: '配置异常' }
  if (!n.online) return { color: 'bg-red-500', label: '未运行' }
  if (n.online_count > 0) return { color: 'bg-green-500', label: '运行正常' }
  return { color: 'bg-yellow-500', label: '无人使用' }
}

/* ── 数据加载 ── */
async function fetchData() {
  try {
    const res = await listNodes({ page: page.value, page_size: 20 })
    if (res.code === 0) { nodes.value = res.data.items; total.value = res.data.total }
  } catch { toast.error('获取节点列表失败') }
}

async function fetchOptions() {
  try {
    const [groupRes, routeRes] = await Promise.all([
      listGroups(), listRoutes({ page: 1, page_size: 100 })
    ])
    if (groupRes.code === 0) groups.value = groupRes.data || []
    if (routeRes.code === 0) routes.value = routeRes.data.items || []
  } catch {}
}

/* ── 弹窗操作 ── */
function openCreate() {
  editing.value = {
    custom_id: '', name: '', address: '', protocol: 'vless', port: 443,
    service_port: 0, rate: 1, dynamic_rate: false, tags: '',
    traffic_limit: 0, group_id: null, route_id: null, parent_id: null,
    security: 'none', transport: 'tcp', flow_control: 'none',
    config_mode: 'auto', config_json: '', network_settings: '', status: 1,
  }
  netSettings.value = { host: '', reality_port: 443 }
  isEdit.value = false
  dialogOpen.value = true
}

function openEdit(n: Node) {
  editing.value = { ...n }
  netSettings.value = parseNetSettings(n.network_settings)
  syncAlpnFromSettings()
  isEdit.value = true
  dialogOpen.value = true
}

function openCopy(n: Node) {
  editing.value = { ...n }
  delete editing.value.id
  delete editing.value.created_at
  delete editing.value.updated_at
  editing.value.name = n.name + ' - 副本'
  editing.value.custom_id = ''
  netSettings.value = parseNetSettings(n.network_settings)
  syncAlpnFromSettings()
  isEdit.value = false
  dialogOpen.value = true
  toast.info('已复制节点配置，请修改后提交')
}

async function handleSave() {
  saving.value = true
  try {
    editing.value.network_settings = buildNetSettingsString()
    if (isEdit.value) {
      const res = await updateNode(editing.value.id!, editing.value)
      if (res.code === 0) { toast.success('节点已更新'); dialogOpen.value = false; fetchData() }
      else { toast.error(res.message || '更新失败') }
    } else {
      const res = await createNode(editing.value)
      if (res.code === 0) { toast.success('节点已创建'); dialogOpen.value = false; fetchData() }
      else { toast.error(res.message || '创建失败') }
    }
  } catch (e: any) { toast.error(e?.response?.data?.message || '操作失败') }
  finally { saving.value = false }
}

function confirmDelete(n: Node) { editing.value = n; deleteDialogOpen.value = true }

async function handleDelete() {
  try {
    const res = await deleteNode(editing.value.id!)
    if (res.code === 0) { toast.success('节点已删除'); deleteDialogOpen.value = false; fetchData() }
    else { toast.error(res.message || '删除失败') }
  } catch (e: any) { toast.error(e?.response?.data?.message || '删除失败') }
}

function confirmResetTraffic(n: Node) { editing.value = n; resetTrafficDialogOpen.value = true }

async function handleResetTraffic() {
  try {
    const res = await resetNodeTraffic(editing.value.id!)
    if (res.code === 0) { toast.success('流量已重置'); resetTrafficDialogOpen.value = false; fetchData() }
    else { toast.error(res.message || '重置失败') }
  } catch (e: any) { toast.error(e?.response?.data?.message || '重置失败') }
}

async function handleRestart(id: number) {
  try {
    const res = await restartNode(id)
    if (res.code === 0) toast.success('重启指令已发送')
    else { toast.error(res.message || '重启失败') }
  } catch (e: any) { toast.error(e?.response?.data?.message || '重启失败') }
}

onMounted(() => { fetchData(); fetchOptions() })
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold">节点管理</h1>
        <p class="text-muted-foreground mt-1">管理所有节点，包括添加、删除、编辑等操作</p>
      </div>
      <Button @click="openCreate"><Plus class="mr-2 h-4 w-4" />创建节点</Button>
    </div>

    <!-- 节点表格 (Xboard 风格) -->
    <Card>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-20">节点ID</TableHead>
              <TableHead class="w-12">显隐</TableHead>
              <TableHead>节点</TableHead>
              <TableHead>地址</TableHead>
              <TableHead class="w-24 text-center">在线人数</TableHead>
              <TableHead class="w-20">倍率</TableHead>
              <TableHead class="w-24">权限组</TableHead>
              <TableHead class="w-28">流量使用</TableHead>
              <TableHead class="w-16 text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="n in nodes" :key="n.id">
              <TableCell>
                <Badge variant="outline" class="font-mono">{{ n.custom_id || n.id }}</Badge>
              </TableCell>
              <TableCell>
                <Switch
                  :model-value="n.status === 1"
                  @update:model-value="async (val) => { await updateNode(n.id, { status: val ? 1 : 0 }); fetchData() }"
                />
              </TableCell>
              <TableCell>
                <div class="flex items-center gap-2">
                  <span :class="getNodeStatus(n).color" class="w-2.5 h-2.5 rounded-full shrink-0" :title="getNodeStatus(n).label" />
                  <span class="font-medium">{{ n.name }}</span>
                  <Badge v-for="tag in parseTags(n.tags)" :key="tag" variant="secondary" class="text-xs">{{ tag }}</Badge>
                </div>
              </TableCell>
              <TableCell>
                <span class="font-mono text-sm">{{ n.address }}:{{ n.port }}</span>
              </TableCell>
              <TableCell class="text-center">
                <div class="flex items-center justify-center gap-1 text-sm">
                  <span class="text-muted-foreground">👤</span>
                  <span>{{ n.online_count || 0 }}</span>
                </div>
              </TableCell>
              <TableCell>
                <Badge variant="secondary">{{ n.rate || 1 }}x</Badge>
              </TableCell>
              <TableCell>
                <Badge v-if="n.group_id" variant="outline">{{ getGroupName(n.group_id) }}</Badge>
                <span v-else class="text-muted-foreground text-sm">-</span>
              </TableCell>
              <TableCell>
                <span class="text-sm">{{ formatBytes(n.traffic_used) }}</span>
                <span v-if="n.traffic_limit > 0" class="text-muted-foreground text-xs"> / {{ formatBytes(n.traffic_limit) }}</span>
              </TableCell>
              <TableCell class="text-right" @click.stop>
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="sm" class="h-8 w-8 p-0"><MoreHorizontal class="h-4 w-4" /></Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem @click="openEdit(n)"><Pencil class="mr-2 h-4 w-4" />编辑</DropdownMenuItem>
                    <DropdownMenuItem @click="openCopy(n)"><Copy class="mr-2 h-4 w-4" />复制</DropdownMenuItem>
                    <DropdownMenuItem @click="confirmResetTraffic(n)"><RotateCcw class="mr-2 h-4 w-4" />重置流量</DropdownMenuItem>
                    <DropdownMenuItem class="text-red-500" @click="confirmDelete(n)"><Trash2 class="mr-2 h-4 w-4" />删除</DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
            <TableRow v-if="!nodes.length">
              <TableCell colspan="9" class="text-center py-12 text-muted-foreground">暂无节点数据</TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <div class="flex items-center justify-between px-4 py-3 border-t">
          <span class="text-sm text-muted-foreground">共 {{ total }} 条</span>
          <div class="flex gap-2">
            <Button variant="outline" size="sm" :disabled="page <= 1" @click="page--; fetchData()">上一页</Button>
            <span class="flex items-center text-sm">第 {{ page }} 页</span>
            <Button variant="outline" size="sm" :disabled="page * 20 >= total" @click="page++; fetchData()">下一页</Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- 创建/编辑弹窗 -->
    <Dialog v-model:open="dialogOpen">
      <DialogContent class="max-w-3xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <div class="flex items-center justify-between">
            <div>
              <DialogTitle>{{ isEdit ? '编辑节点' : '创建节点' }}</DialogTitle>
              <DialogDescription class="mt-1">配置节点的连接参数和传输协议</DialogDescription>
            </div>
            <Select v-model="editing.protocol">
              <SelectTrigger class="w-[160px]"><SelectValue placeholder="选择协议类型" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="vless">VLESS</SelectItem>
                <SelectItem value="hysteria2">Hysteria2</SelectItem>
                <SelectItem value="tuic">TUIC</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </DialogHeader>

        <div class="grid gap-6 py-2">
          <!-- 基础信息 -->
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label>节点名称</Label>
              <Input v-model="editing.name" placeholder="请输入节点名称" />
            </div>
            <div class="grid gap-2">
              <Label>基础倍率</Label>
              <div class="relative">
                <Input v-model.number="editing.rate" type="number" step="0.1" min="0" class="pr-8" placeholder="1" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">x</span>
              </div>
            </div>
          </div>

          <!-- 动态倍率 -->
          <div class="flex items-center justify-between border rounded-lg p-4">
            <div>
              <Label>启用动态倍率</Label>
              <p class="text-sm text-muted-foreground">根据时间段设置不同的倍率乘数</p>
            </div>
            <Switch v-model="editing.dynamic_rate" />
          </div>

          <!-- 流量限制 + 自定义节点ID -->
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label>流量限制（GB）</Label>
              <Input v-model.number="editing.traffic_limit" type="number" min="0" placeholder="0 表示不限制" />
            </div>
            <div class="grid gap-2">
              <Label>自定义节点ID（选填）</Label>
              <Input v-model="editing.custom_id" placeholder="请输入自定义节点ID" />
            </div>
          </div>

          <!-- 节点标签 -->
          <div class="grid gap-2">
            <Label>节点标签</Label>
            <Input v-model="editing.tags" placeholder="输入后回车添加标签，多个用逗号分隔" />
          </div>

          <!-- 权限组 -->
          <div class="grid gap-2">
            <Label>权限组</Label>
            <Select v-model="editing.group_id">
              <SelectTrigger><SelectValue placeholder="请选择权限组" /></SelectTrigger>
              <SelectContent>
                <SelectItem :value="null">不分组</SelectItem>
                <SelectItem v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- 节点地址 -->
          <div class="grid gap-2">
            <Label>节点地址</Label>
            <Input v-model="editing.address" placeholder="请输入节点域名或者IP" />
          </div>

          <!-- 连接端口 + 服务端口 -->
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label>连接端口</Label>
              <Input v-model.number="editing.port" type="number" placeholder="用户连接端口" />
            </div>
            <div class="grid gap-2">
              <Label>服务端口</Label>
              <Input v-model.number="editing.service_port" type="number" placeholder="请输入服务端口" />
            </div>
          </div>

          <!-- ========== VLESS 协议参数 ========== -->
          <template v-if="editing.protocol === 'vless'">
            <Separator />
            <div class="flex items-center gap-2">
              <Settings class="h-4 w-4" />
              <Label class="text-base font-semibold">VLESS 协议参数</Label>
            </div>

            <!-- 安全性 + 流控 -->
            <div class="grid grid-cols-2 gap-4">
              <div class="grid gap-2">
                <Label>安全性</Label>
                <Select v-model="editing.security">
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="none">无</SelectItem>
                    <SelectItem value="tls">TLS</SelectItem>
                    <SelectItem value="reality">Reality</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div class="grid gap-2">
                <Label>流控</Label>
                <Select v-model="editing.flow_control">
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="none">无</SelectItem>
                    <SelectItem value="xtls-rprx-direct">xtls-rprx-direct</SelectItem>
                    <SelectItem value="xtls-rprx-splice">xtls-rprx-splice</SelectItem>
                    <SelectItem value="xtls-rprx-vision">xtls-rprx-vision</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <!-- TLS 设置 -->
            <div v-if="editing.security === 'tls'" class="grid gap-3 border rounded-lg p-4 bg-muted/30">
              <Label class="font-medium">TLS 设置</Label>
              <div class="grid grid-cols-2 gap-4">
                <div class="grid gap-2">
                  <Label>SNI / 服务器名称</Label>
                  <Input v-model="netSettings.server_name" placeholder="例如 www.microsoft.com" />
                </div>
                <div class="flex items-center gap-2 pt-6">
                  <Switch v-model="netSettings.allow_insecure" />
                  <Label class="text-sm">允许不安全连接</Label>
                </div>
              </div>
            </div>

            <!-- Reality 设置 -->
            <div v-if="editing.security === 'reality'" class="grid gap-4 border rounded-lg p-4 bg-muted/30">
              <!-- 伪装站点 + 端口 + 允许不安全 -->
              <div class="grid grid-cols-[1fr_120px_auto] gap-4 items-end">
                <div class="grid gap-2">
                  <Label>伪装站点(dest)</Label>
                  <Input v-model="netSettings.reality_server_name" placeholder="例如: example.com" />
                </div>
                <div class="grid gap-2">
                  <Label>端口(port)</Label>
                  <Input v-model.number="netSettings.reality_port" type="number" placeholder="443" />
                </div>
                <div class="flex items-center gap-2 pb-1">
                  <Switch v-model="netSettings.allow_insecure" />
                  <Label class="text-sm whitespace-nowrap">允许不安全?</Label>
                </div>
              </div>

              <!-- 私钥 -->
              <div class="grid gap-2">
                <Label>私钥(Private key)</Label>
                <div class="relative">
                  <Input v-model="netSettings.reality_private_key" placeholder="点击右侧按钮生成密钥对" class="pr-10" />
                  <Button variant="ghost" size="sm" class="absolute right-0 top-0 h-full px-3" @click="handleGenerateRealityKeys" title="生成密钥对">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/></svg>
                  </Button>
                </div>
              </div>

              <!-- 公钥 + Short ID -->
              <div class="grid grid-cols-2 gap-4">
                <div class="grid gap-2">
                  <Label>公钥(Public key)</Label>
                  <Input v-model="netSettings.reality_public_key" placeholder="自动填入" />
                </div>
                <div class="grid gap-2">
                  <Label>Short ID</Label>
                  <Input v-model="netSettings.reality_short_id" placeholder="留空自动生成" />
                </div>
              </div>

              <!-- uTLS -->
              <div class="flex items-center justify-between border rounded-lg p-4 bg-background">
                <div>
                  <Label class="font-medium">uTLS</Label>
                  <p class="text-sm text-muted-foreground">客户端伪装指纹，用于降低被识别风险</p>
                </div>
                <Switch v-model="netSettings.utls_enabled" />
              </div>
            </div>

            <!-- 传输协议选择 -->
            <div class="grid gap-2">
              <div class="flex items-center gap-2"><Label>传输协议</Label><Button variant="link" size="sm" class="h-auto p-0 text-sm" @click="openProtocolEdit">编辑协议</Button></div>
              <Select v-model="editing.transport">
                <SelectTrigger><SelectValue placeholder="选择传输协议" /></SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="t in vlessTransports" :key="t.value" :value="t.value">{{ t.label }}</SelectItem>
                </SelectContent>
              </Select>
            </div>

          </template>

          <!-- ========== Hysteria2 协议参数 ========== -->
          <template v-else-if="editing.protocol === 'hysteria2'">
            <Separator />
            <div class="flex items-center gap-2 mb-2">
              <Settings class="h-4 w-4" />
              <Label class="text-base font-semibold">Hysteria2 协议参数</Label>
            </div>

            <!-- 协议版本 + ALPN -->
            <div class="grid grid-cols-2 gap-4">
              <div class="grid gap-2">
                <Label>协议版本</Label>
                <Select v-model="netSettings.version">
                  <SelectTrigger><SelectValue placeholder="V2" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem :value="1">V1</SelectItem>
                    <SelectItem :value="2">V2</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div v-if="netSettings.version === 1" class="grid gap-2">
                <Label>ALPN</Label>
                <Select v-model="netSettings.alpn">
                  <SelectTrigger><SelectValue placeholder="hysteria" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="hysteria">hysteria</SelectItem>
                    <SelectItem value="http/1.1">http/1.1</SelectItem>
                    <SelectItem value="h2">h2</SelectItem>
                    <SelectItem value="h3">h3</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <!-- 混淆 -->
            <div class="border rounded-lg p-4">
              <div class="flex items-center gap-4 mb-3">
                <Label>混淆</Label>
                <Switch v-model="netSettings.obfs_open" />
                <Label class="text-sm text-muted-foreground">混淆实现</Label>
              </div>
              <div v-if="netSettings.obfs_open" class="grid grid-cols-2 gap-4">
                <div class="grid gap-2">
                  <Label>混淆实现</Label>
                  <Select v-model="netSettings.obfs_type">
                    <SelectTrigger><SelectValue placeholder="Salamander" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="salamander">Salamander</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div class="grid gap-2">
                  <Label>混淆密码</Label>
                  <div class="relative">
                    <Input v-model="netSettings.obfs_password" placeholder="点击右侧按钮生成密码" class="pr-10" />
                    <Button variant="ghost" size="sm" class="absolute right-0 top-0 h-full px-3" @click="netSettings.obfs_password = Math.random().toString(36).substring(2, 18)" title="生成密码">
                      <RefreshCw class="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </div>
            </div>

            <!-- SNI + 允许不安全 -->
            <div class="grid gap-3 border rounded-lg p-4 bg-muted/30">
              <div class="grid gap-2">
                <Label>服务器名称指示(SNI)</Label>
                <Input v-model="netSettings.tls_server_name" placeholder="当节点地址与证书不一致时用于证书验证" />
              </div>
              <div class="flex items-center justify-between">
                <Label class="text-sm text-muted-foreground">允许不安全?</Label>
                <Switch v-model="netSettings.tls_allow_insecure" />
              </div>
            </div>

            <!-- ECH -->
            <div class="flex items-center justify-between border rounded-lg p-4">
              <div>
                <Label class="font-medium">ECH</Label>
                <p class="text-sm text-muted-foreground">为支持的 TLS 客户端启用 Encrypted Client Hello。留空配置时会尝试通过 DNS 查询。</p>
              </div>
              <Switch v-model="netSettings.ech_enabled" />
            </div>

            <!-- 上行带宽 -->
            <div class="grid gap-2">
              <Label>上行带宽</Label>
              <div class="relative">
                <Input v-model.number="netSettings.bandwidth_up" type="number" placeholder="请输入上行带宽，留空则使用BBR" class="pr-12" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">Mbps</span>
              </div>
            </div>

            <!-- 下行带宽 -->
            <div class="grid gap-2">
              <Label>下行带宽</Label>
              <div class="relative">
                <Input v-model.number="netSettings.bandwidth_down" type="number" placeholder="请输入下行带宽，留空则使用BBR" class="pr-12" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">Mbps</span>
              </div>
            </div>

            <!-- Hop 间隔 -->
            <div class="grid gap-2">
              <Label>Hop 间隔（秒）</Label>
              <Input v-model.number="netSettings.hop_interval" type="number" placeholder="例如：30" />
              <p class="text-xs text-muted-foreground">Hop 间隔时间，单位为秒</p>
            </div>
          </template>

          <!-- ========== TUIC 协议参数 ========== -->
          <template v-else-if="editing.protocol === 'tuic'">
            <Separator />
            <div class="flex items-center gap-2 mb-2">
              <Settings class="h-4 w-4" />
              <Label class="text-base font-semibold">TUIC 协议参数</Label>
            </div>

            <!-- 版本 + 拥塞控制 -->
            <div class="grid grid-cols-2 gap-4">
              <div class="grid gap-2">
                <Label>协议版本</Label>
                <Select v-model="netSettings.tuic_version">
                  <SelectTrigger><SelectValue placeholder="5" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem :value="4">V4</SelectItem>
                    <SelectItem :value="5">V5</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div class="grid gap-2">
                <Label>拥塞控制</Label>
                <Select v-model="netSettings.congestion_control">
                  <SelectTrigger><SelectValue placeholder="cubic" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="cubic">Cubic</SelectItem>
                    <SelectItem value="bbr">BBR</SelectItem>
                    <SelectItem value="new_reno">New Reno</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <!-- SNI + 允许不安全 -->
            <div class="grid gap-3 border rounded-lg p-4 bg-muted/30">
              <div class="grid gap-2">
                <Label>服务器名称指示(SNI)</Label>
                <Input v-model="netSettings.tls_server_name" placeholder="当节点地址与证书不一致时用于证书验证" />
              </div>
              <div class="flex items-center justify-between">
                <Label class="text-sm text-muted-foreground">允许不安全?</Label>
                <Switch v-model="netSettings.tls_allow_insecure" />
              </div>
            </div>

            <!-- ECH -->
            <div class="flex items-center justify-between border rounded-lg p-4">
              <div>
                <Label class="font-medium">ECH</Label>
                <p class="text-sm text-muted-foreground">为支持的 TLS 客户端启用 Encrypted Client Hello。留空配置时会尝试通过 DNS 查询。</p>
              </div>
              <Switch v-model="netSettings.ech_enabled" />
            </div>

            <!-- ALPN (多选标签) -->
            <div class="grid gap-2">
              <Label>ALPN</Label>
              <div class="border rounded-md p-2 flex flex-wrap gap-2 min-h-[38px] items-center">
                <Badge v-for="(a, i) in tuicAlpnList" :key="i" variant="secondary" class="gap-1">
                  {{ a }}
                  <button class="ml-1 hover:text-destructive" @click="removeTuicAlpn(i)">×</button>
                </Badge>
                <Select v-model="tuicAlpnAdd" @update:modelValue="addTuicAlpn">
                  <SelectTrigger class="w-auto border-0 h-6 p-0 shadow-none"><SelectValue placeholder="选择ALPN协议" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="h3">HTTP/3</SelectItem>
                    <SelectItem value="h2">HTTP/2</SelectItem>
                    <SelectItem value="http/1.1">HTTP/1.1</SelectItem>
                    <SelectItem value="spdy/3.1">SPDY/3.1</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <!-- UDP中继模式 -->
            <div class="grid gap-2">
              <Label>UDP中继模式</Label>
              <Select v-model="netSettings.udp_relay_mode">
                <SelectTrigger><SelectValue placeholder="native" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="native">Native</SelectItem>
                  <SelectItem value="quic">QUIC</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <!-- Zero RTT + Heartbeat -->
            <div class="grid grid-cols-2 gap-4">
              <div class="flex items-center justify-between border rounded-lg p-4">
                <div>
                  <Label class="font-medium">Zero RTT</Label>
                  <p class="text-sm text-muted-foreground">启用零往返握手</p>
                </div>
                <Switch v-model="netSettings.zero_rtt" />
              </div>
              <div class="grid gap-2">
                <Label>Heartbeat</Label>
                <Input v-model="netSettings.heartbeat" placeholder="10s" />
              </div>
            </div>
          </template>

          <!-- 未选择协议时的提示 -->
          <div v-else class="border rounded-lg p-6 text-center text-muted-foreground">
            <div class="flex flex-col items-center gap-2">
              <div class="w-8 h-8 rounded-full bg-muted flex items-center justify-center text-lg">i</div>
              <p>请先选择协议类型</p>
            </div>
          </div>

          <!-- 父级节点 + 路由组 -->
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label>父级节点</Label>
              <Select v-model="editing.parent_id">
                <SelectTrigger><SelectValue placeholder="无" /></SelectTrigger>
                <SelectContent>
                  <SelectItem :value="null">无</SelectItem>
                  <SelectItem v-for="n in nodes" :key="n.id" :value="n.id">{{ n.name }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="grid gap-2">
              <Label>路由组</Label>
              <Select v-model="editing.route_id">
                <SelectTrigger><SelectValue placeholder="选择路由组" /></SelectTrigger>
                <SelectContent>
                  <SelectItem :value="null">不绑定</SelectItem>
                  <SelectItem v-for="r in routes" :key="r.id" :value="r.id">{{ r.name }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

        </div>

        <DialogFooter class="flex items-center justify-between pt-2 border-t">
          <Button variant="ghost" size="sm" @click="openAdvanced" class="gap-2 text-muted-foreground">
            <SlidersHorizontal class="h-4 w-4" />
            高级设置
          </Button>
          <div class="flex gap-2">
            <Button variant="outline" @click="dialogOpen = false">取消</Button>
            <Button @click="handleSave" :disabled="saving">{{ saving ? '保存中...' : '提交' }}</Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 高级协议配置弹窗 -->
    <Dialog v-model:open="advancedOpen">
      <DialogContent class="!w-full !max-w-[800px] !max-h-[85vh] overflow-y-auto overflow-x-hidden">
        <DialogHeader>
          <DialogTitle>高级协议配置</DialogTitle>
          <DialogDescription>配置 TLS 证书、自定义 Outbounds 和 Routes</DialogDescription>
        </DialogHeader>
        <Tabs v-model:value="advancedTab" class="w-full">
          <TabsList class="flex w-full gap-1 shrink-0">
            <TabsTrigger value="tls">TLS</TabsTrigger>
            <TabsTrigger v-if="editing.protocol === 'vless'" value="multiplex">多路复用</TabsTrigger>
            <TabsTrigger value="outbounds">自定义 Outbounds</TabsTrigger>
            <TabsTrigger value="routes">自定义 Routes</TabsTrigger>
          </TabsList>

          <!-- TLS 证书 Tab -->
          <TabsContent value="tls" class="grid gap-4 py-4">
            <div class="grid gap-2">
              <Label>证书模式</Label>
              <p class="text-xs text-muted-foreground">选择证书申请方式，仅部分后端节点支持</p>
              <Select v-model="advancedTls.cert_mode">
                <SelectTrigger class="w-full truncate"><SelectValue placeholder="选择证书模式" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">none - 未启用 TLS 证书配置</SelectItem>
                  <SelectItem value="self">自签名模式 - 证书由节点后端自动生成（10年有效期）</SelectItem>
                  <SelectItem value="http-01">HTTP-01 模式 - 需要 80 端口可正常访问以完成认证</SelectItem>
                  <SelectItem value="dns-01">DNS-01 模式 - 通过 DNS 解析记录认证，支持申请泛域名证书</SelectItem>
                  <SelectItem value="manual">内容推送模式 - 直接将证书内容下发至节点</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <template v-if="advancedTls.cert_mode !== 'none'">
              <div class="grid gap-2">
                <Label>证书域名</Label>
                <Input v-model="advancedTls.domain" placeholder="example.com" />
              </div>
            </template>

            <template v-if="advancedTls.cert_mode === 'http-01' || advancedTls.cert_mode === 'dns-01'">
              <div class="grid gap-2">
                <Label>通知邮箱</Label>
                <Input v-model="advancedTls.email" placeholder="admin@example.com" />
              </div>
            </template>

            <template v-if="advancedTls.cert_mode === 'http-01'">
              <div class="grid gap-2">
                <Label>认证端口</Label>
                <Input v-model="advancedTls.http_port" type="number" placeholder="80" />
                <p class="text-xs text-muted-foreground">ACME 认证端口，默认 80</p>
              </div>
            </template>

            <template v-if="advancedTls.cert_mode === 'dns-01'">
              <div class="grid gap-2">
                <Label>DNS 提供商</Label>
                <Input v-model="advancedTls.dns_provider" placeholder="选择 DNS 提供商" />
                <p class="text-xs text-muted-foreground">查看 DNS 提供商配置指南</p>
              </div>
              <div class="grid gap-2">
                <Label>环境变量 (API 密钥)</Label>
                <Textarea v-model="advancedTls.dns_env" rows="3" placeholder="KEY=VALUE" class="font-mono text-sm" />
                <p class="text-xs text-muted-foreground">每行一个 KEY=VALUE 配置</p>
              </div>
            </template>

            <template v-if="advancedTls.cert_mode === 'manual'">
              <div class="grid gap-2">
                <Label>证书内容 (Public Key)</Label>
                <Textarea v-model="advancedTls.cert_content" rows="4" placeholder="-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----" class="font-mono text-sm" />
              </div>
              <div class="grid gap-2">
                <Label>密钥内容 (Private Key)</Label>
                <Textarea v-model="advancedTls.key_content" rows="4" placeholder="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----" class="font-mono text-sm" />
              </div>
            </template>
          </TabsContent>

          <!-- 多路复用 Tab (VLESS only) -->
          <TabsContent v-if="editing.protocol === 'vless'" value="multiplex" class="grid gap-4 py-4">
            <div class="flex items-center justify-between border rounded-lg p-4 bg-muted/30">
              <div>
                <Label class="font-medium">多路复用 (Multiplex)</Label>
                <p class="text-sm text-muted-foreground">通过单条 TCP 连接传输多个流，降低握手延迟</p>
              </div>
              <Switch v-model="multiplex.enabled" />
            </div>
            <template v-if="multiplex.enabled">
              <div class="grid grid-cols-3 gap-4">
                <div class="grid gap-2">
                  <Label>复用协议</Label>
                  <Select v-model="multiplex.protocol">
                    <SelectTrigger><SelectValue placeholder="选择复用协议" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="smux">smux</SelectItem>
                      <SelectItem value="yamux">yamux</SelectItem>
                      <SelectItem value="h2mux">h2mux</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div class="flex items-center gap-3 pt-6">
                  <Switch v-model="multiplex.padding" />
                  <Label class="text-sm whitespace-nowrap">启用填充</Label>
                </div>
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div class="grid gap-2">
                  <Label>最大连接数</Label>
                  <Input v-model.number="multiplex.max_connections" type="number" min="1" placeholder="16" />
                </div>
                <div class="grid gap-2">
                  <Label>最小流数</Label>
                  <Input v-model.number="multiplex.min_streams" type="number" min="1" placeholder="4" />
                </div>
              </div>
              <!-- TCP Brutal -->
              <div class="border rounded-lg p-4 bg-muted/30">
                <div class="flex items-center justify-between mb-3">
                  <div>
                    <Label class="font-medium">TCP Brutal (激进拥塞控制)</Label>
                    <p class="text-xs text-muted-foreground">双边加速算法，建议带宽设为机器实际带宽的 80%-90%，开启后 BBR 将失效</p>
                  </div>
                  <Switch v-model="multiplex.brutal_enabled" />
                </div>
                <div v-if="multiplex.brutal_enabled" class="grid grid-cols-2 gap-4">
                  <div class="grid gap-2">
                    <Label>上行带宽 (Mbps)</Label>
                    <Input v-model.number="multiplex.brutal_up_mbps" type="number" min="1" placeholder="请输入上行带宽" />
                  </div>
                  <div class="grid gap-2">
                    <Label>下行带宽 (Mbps)</Label>
                    <Input v-model.number="multiplex.brutal_down_mbps" type="number" min="1" placeholder="请输入下行带宽" />
                  </div>
                </div>
              </div>
            </template>
          </TabsContent>

          <!-- 自定义 Outbounds Tab -->
          <TabsContent value="outbounds" class="grid gap-4 py-4">
            <div class="border rounded-lg p-4 bg-muted/30">
              <div class="flex items-center justify-between mb-2">
                <div>
                  <Label class="font-medium">自定义 Outbounds</Label>
                  <p class="text-xs text-muted-foreground mt-1">配置自定义出站规则，内容会合并到 sing-box 的 outbounds 配置中</p>
                </div>
                <Button variant="outline" size="sm" @click="advancedOutbounds = JSON.stringify(JSON.parse(advancedOutbounds || '[]'), null, 2)" :disabled="!advancedOutbounds.trim()">JSON 格式化</Button>
              </div>
              <Textarea v-model="advancedOutbounds" rows="8" class="font-mono text-sm bg-background" placeholder='[
  {
    "type": "direct",
    "tag": "direct-out"
  }
]' />
            </div>
          </TabsContent>

          <!-- 自定义 Routes Tab -->
          <TabsContent value="routes" class="grid gap-4 py-4">
            <div class="border rounded-lg p-4 bg-muted/30">
              <div class="flex items-center justify-between mb-2">
                <div>
                  <Label class="font-medium">自定义 Routes</Label>
                  <p class="text-xs text-muted-foreground mt-1">配置自定义路由规则，内容会合并到 sing-box 的 route 配置中</p>
                </div>
                <Button variant="outline" size="sm" @click="advancedRoutes = JSON.stringify(JSON.parse(advancedRoutes || '[]'), null, 2)" :disabled="!advancedRoutes.trim()">JSON 格式化</Button>
              </div>
              <Textarea v-model="advancedRoutes" rows="8" class="font-mono text-sm bg-background" placeholder='[
  {
    "type": "field",
    "outbound": "any",
    "domain": ["geosite:cn"]
  }
]' />
            </div>
          </TabsContent>
        </Tabs>

        <DialogFooter class="flex items-center justify-end gap-2 pt-2 border-t shrink-0">
          <Button variant="outline" @click="advancedOpen = false">取消</Button>
          <Button @click="saveAdvanced">保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 编辑协议配置弹窗 -->
    <Dialog v-model:open="protocolEditOpen">
      <DialogContent class="max-w-lg">
        <DialogHeader>
          <DialogTitle>编辑协议配置</DialogTitle>
        </DialogHeader>
        <div class="grid gap-4 py-2">
          <div class="flex flex-wrap gap-2">
            <Button
              v-for="t in (transportTemplates[editing.transport || 'tcp'] || [])"
              :key="t.label" variant="outline" size="sm"
              @click="applyTransportTemplate(t.json)">
              {{ t.label }}
            </Button>
          </div>
          <Textarea v-model="protocolJson" rows="12" placeholder="请输入JSON配置或选择上方模板" class="font-mono text-sm" />
        </div>
        <DialogFooter>
          <Button variant="outline" @click="protocolEditOpen = false">取消</Button>
          <Button @click="saveProtocolEdit">确定</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 重置流量确认弹窗 -->
    <Dialog v-model:open="resetTrafficDialogOpen">
      <DialogContent>
        <DialogHeader><DialogTitle>确认重置流量</DialogTitle></DialogHeader>
        <DialogDescription>确定要将节点「{{ editing.name }}」的流量统计重置为零吗？此操作不可撤销。</DialogDescription>
        <DialogFooter>
          <Button variant="outline" @click="resetTrafficDialogOpen = false">取消</Button>
          <Button variant="destructive" @click="handleResetTraffic">重置</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 删除确认弹窗 -->
    <Dialog v-model:open="deleteDialogOpen">
      <DialogContent>
        <DialogHeader><DialogTitle>确认删除</DialogTitle></DialogHeader>
        <DialogDescription>确定要删除节点「{{ editing.name }}」吗？此操作不可撤销。</DialogDescription>
        <DialogFooter>
          <Button variant="outline" @click="deleteDialogOpen = false">取消</Button>
          <Button variant="destructive" @click="handleDelete">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
