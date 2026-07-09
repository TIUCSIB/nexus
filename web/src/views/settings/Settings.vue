<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Separator } from '@/components/ui/separator'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { toast } from 'vue-sonner'
import { getSettings, updateSettings } from '@/api/settings'
import { getBackupInfo, downloadBackup } from '@/api/settings'
import { getSystemStatus } from '@/api/systemStatus'
import { useSettingsStore } from '@/stores/settings'
import { useRouter } from 'vue-router'
import { AlertCircle, RefreshCw, Download, Database, Server, Clock, HardDrive, Globe, Activity } from 'lucide-vue-next'

const settingsStore = useSettingsStore()
const router = useRouter()

const loading = ref(false)
const originalSubPath = ref('s')

const site = ref({ app_name: 'Nexus', admin_path: 'admin', auth_path: 'auth', user_path: 'user', sub_url: '', app_description: '', force_https: false })
const subscription = ref({
  sub_path: 's', sub_show_info: true, reset_traffic_method: '0',
  new_order_event_id: '', renew_order_event_id: '',
})
const nodeConfig = ref({ server_token: '', device_limit_mode: '0', node_pull_interval: '60', node_push_interval: '60', websocket_enabled: false, websocket_url: '' })
const backupInfo = ref<any>(null)
const backupLoading = ref(false)
const systemStatus = ref<any>(null)
const sysLoading = ref(false)

const subPathChanged = computed(() => {
  return subscription.value.sub_path !== originalSubPath.value
})

const subPathValid = computed(() => {
  const p = subscription.value.sub_path
  if (!p || p.length === 0) return false
  return /^[a-zA-Z0-9_-]+$/.test(p)
})

const subUrlPreview = computed(() => {
  const base = (site.value.sub_url || window.location.origin).split(',')[0].trim().replace(/\/+$/, '')
  const path = subscription.value.sub_path || 's'
  return `${base}/${path}/{token}`
})


const resetTrafficOptions = [
  { value: '0', label: '不重置' },
  { value: '1', label: '每月1号重置' },
  { value: '2', label: '按订阅周期重置' },
  { value: '3', label: '每年1月1号重置' },
]
const deviceLimitOptions = [
  { value: '0', label: '宽松模式' },
  { value: '1', label: '严格模式' },
]

function generateToken() {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let result = ''
  for (let i = 0; i < 22; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  nodeConfig.value.server_token = result
}

async function fetchSettings() {
  try {
    const res = await getSettings()
    if (res.code === 0 && res.data) {
      const d = res.data
      if (d.app_name) site.value.app_name = d.app_name
      if (d.admin_path) site.value.admin_path = d.admin_path
	      if (d.auth_path) site.value.auth_path = d.auth_path
	      if (d.user_path) site.value.user_path = d.user_path
      if (d.sub_url) site.value.sub_url = d.sub_url
      if (d.app_description) site.value.app_description = d.app_description
      if (d.force_https !== undefined) site.value.force_https = (d.force_https === 'true' || d.force_https === true || d.force_https === 1)
      if (d.sub_path) subscription.value.sub_path = d.sub_path
      if (d.sub_show_info !== undefined) subscription.value.sub_show_info = (d.sub_show_info === 'true' || d.sub_show_info === true || d.sub_show_info === 1)
      if (d.reset_traffic_method) subscription.value.reset_traffic_method = d.reset_traffic_method
      if (d.new_order_event_id) subscription.value.new_order_event_id = d.new_order_event_id
      if (d.renew_order_event_id) subscription.value.renew_order_event_id = d.renew_order_event_id
      if (d.server_token) nodeConfig.value.server_token = d.server_token
      if (d.device_limit_mode) nodeConfig.value.device_limit_mode = d.device_limit_mode
      if (d.node_pull_interval) nodeConfig.value.node_pull_interval = d.node_pull_interval
      if (d.node_push_interval) nodeConfig.value.node_push_interval = d.node_push_interval
      if (d.websocket_enabled !== undefined) nodeConfig.value.websocket_enabled = (d.websocket_enabled === 'true' || d.websocket_enabled === true || d.websocket_enabled === 1)
      if (d.websocket_url !== undefined) nodeConfig.value.websocket_url = d.websocket_url
      originalSubPath.value = d.sub_path || 's'
    }
  } catch { toast.error('获取设置失败') }
}

async function handleSave(tab: string) {
  loading.value = true
  try {
    let data: Record<string, string> = {}
    if (tab === 'site') {
data = {
	        app_name: site.value.app_name,
	        admin_path: site.value.admin_path,
	        auth_path: site.value.auth_path,
	        user_path: site.value.user_path,
	        sub_url: site.value.sub_url,
	        app_description: site.value.app_description,
	        force_https: site.value.force_https ? 'true' : 'false',
	      }
    } else if (tab === 'subscription') {
      if (!subPathValid.value) {
        toast.error('订阅路径格式不正确，只允许字母、数字、下划线和短横线')
        loading.value = false
        return
      }
      data = {
        sub_path: subscription.value.sub_path,
        sub_show_info: subscription.value.sub_show_info ? 'true' : 'false',
        reset_traffic_method: subscription.value.reset_traffic_method,
        new_order_event_id: subscription.value.new_order_event_id,
        renew_order_event_id: subscription.value.renew_order_event_id,
      }
    } else if (tab === 'node') {
      data = { server_token: nodeConfig.value.server_token, device_limit_mode: nodeConfig.value.device_limit_mode, node_pull_interval: nodeConfig.value.node_pull_interval, node_push_interval: nodeConfig.value.node_push_interval, websocket_enabled: nodeConfig.value.websocket_enabled ? 'true' : 'false', websocket_url: nodeConfig.value.websocket_url }
    }
    const res = await updateSettings(data)
    if (res.code === 0) {
      toast.success('保存成功')
      if (tab === 'site') {
        settingsStore.setAppName(site.value.app_name)
        settingsStore.setAppDescription(site.value.app_description)
        settingsStore.setAdminPath(site.value.admin_path)
	        localStorage.setItem('auth_path', site.value.auth_path)
	        localStorage.setItem('user_path', site.value.user_path)
        settingsStore.setSubUrl(site.value.sub_url)
        const newPath = '/' + site.value.admin_path + '/settings'
        if (router.currentRoute.value.path !== newPath) {
          router.push(newPath)
        }
      }
      if (tab === 'subscription') {
        originalSubPath.value = subscription.value.sub_path
        localStorage.setItem('sub_path', subscription.value.sub_path)
      }
    } else {
      toast.error(res.message || '保存失败')
    }
  } catch (e: any) { toast.error(e?.response?.data?.message || '保存失败') }
  finally { loading.value = false }
}

async function fetchBackupInfo() {
  try {
    const res = await getBackupInfo()
    if (res.code === 0) backupInfo.value = res.data
  } catch {}
}

async function loadSystemStatus() {
  sysLoading.value = true
  try {
    const res = await getSystemStatus()
    if (res.code === 0) systemStatus.value = res.data
  } catch { toast.error('获取系统状态失败') }
  finally { sysLoading.value = false }
}

async function handleBackup() {
  backupLoading.value = true
  try {
    await downloadBackup()
    toast.success('数据库备份下载成功')
    fetchBackupInfo()
  } catch (e: any) {
    toast.error(e?.response?.data?.message || '备份失败')
  } finally {
    backupLoading.value = false
  }
}

function formatBytes(bytes: number) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

onMounted(() => { fetchSettings(); fetchBackupInfo(); loadSystemStatus() })
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold">系统设置</h1>
    <Tabs default-value="site">
      <TabsList>
        <TabsTrigger value="site">站点设置</TabsTrigger>
        <TabsTrigger value="subscription">订阅设置</TabsTrigger>
        <TabsTrigger value="node">节点配置</TabsTrigger>
        <TabsTrigger value="system">系统状态</TabsTrigger>
      </TabsList>

      <TabsContent value="site">
        <Card>
          <CardHeader>
            <CardTitle>站点设置</CardTitle>
            <CardDescription>配置站点基本信息和访问路径</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="grid gap-2">
              <Label>站点名称</Label>
              <Input v-model="site.app_name" placeholder="Nexus" />
            </div>
            <div class="grid gap-2">
              <Label>后台路径</Label>
              <div class="flex items-center gap-2">
                <span class="text-sm text-muted-foreground">/</span>
                <Input v-model="site.admin_path" placeholder="admin" class="flex-1" />
              </div>
              <p class="text-sm text-muted-foreground">自定义后台访问路径，防止被扫描器发现。修改后需要重新登录</p>
            </div>
            <div class="grid gap-2">
              <Label>认证接口路径</Label>
              <div class="flex items-center gap-2">
                <span class="text-sm text-muted-foreground">/api/</span>
                <Input v-model="site.auth_path" placeholder="auth" class="flex-1" />
              </div>
              <p class="text-sm text-muted-foreground">登录/刷新Token接口路径，默认 auth</p>
            </div>
            <div class="grid gap-2">
              <Label>用户接口路径</Label>
              <div class="flex items-center gap-2">
                <span class="text-sm text-muted-foreground">/api/</span>
                <Input v-model="site.user_path" placeholder="user" class="flex-1" />
              </div>
              <p class="text-sm text-muted-foreground">用户信息接口路径，默认 user</p>
            </div>
            <div class="grid gap-2">
              <Label>订阅URL</Label>
              <Input v-model="site.sub_url" placeholder="https://example.com,https://example2.com" />
              <p class="text-sm text-muted-foreground">用于订阅所使用的，留空则为站点URL。多个地址用逗号分隔</p>
            </div>
            <div class="grid gap-2"><Label>站点描述</Label><Textarea v-model="site.app_description" rows="2"
                placeholder="一个简洁的代理面板" /></div>
            <div class="flex items-center justify-between">
              <div>
                <Label>强制HTTPS</Label>
                <p class="text-sm text-muted-foreground">所有请求自动跳转到HTTPS</p>
              </div>
              <Switch v-model="site.force_https" />
            </div>
            <div class="flex justify-end">
              <Button @click="handleSave('site')" :disabled="loading">{{ loading ? '保存中...' : '保存设置' }}</Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="subscription">
        <Card>
          <CardHeader>
            <CardTitle>订阅设置</CardTitle>
            <CardDescription>配置订阅和套餐相关选项</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="grid gap-2">
              <Label>订阅路径</Label>
              <div class="flex items-center gap-2">
                <Input v-model="subscription.sub_path" placeholder="s"
                  :class="{ 'border-destructive': subscription.sub_path && !subPathValid }" />
              </div>
              <p class="text-sm text-muted-foreground">订阅路径，只允许字母、数字、下划线和短横线。修改后立即生效，无需重启服务</p>
              <div v-if="subPathChanged"
                class="flex items-start gap-2 rounded-md border border-yellow-500/50 bg-yellow-500/10 p-3">
                <AlertCircle class="h-4 w-4 text-yellow-600 mt-0.5 shrink-0" />
                <div class="text-sm text-yellow-700">
                  订阅路径已修改，保存后原有的订阅链接将失效，需要将新的订阅链接重新导入到客户端中。
                </div>
              </div>
              <div v-if="subscription.sub_path && subPathValid" class="rounded-md bg-muted p-3">
                <p class="text-sm text-muted-foreground mb-1">预览订阅地址：</p>
                <code class="text-sm font-mono break-all">{{ subUrlPreview }}</code>
              </div>
            </div>
            <Separator />
            <div class="flex items-center justify-between">
              <div>
                <Label>在订阅中展示订阅信息</Label>
                <p class="text-sm text-muted-foreground">开启后将会在用户订阅节点时输出订阅信息</p>
              </div>
              <Switch v-model="subscription.sub_show_info" />
            </div>
            <div class="grid gap-2">
              <Label>流量重置方式</Label>
              <Select v-model="subscription.reset_traffic_method">
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="opt in resetTrafficOptions" :key="opt.value" :value="opt.value">{{ opt.label }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="flex justify-end">
              <Button @click="handleSave('subscription')" :disabled="loading || !subPathValid">{{ loading ? '保存中...' :
                '保存设置' }}</Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="node">
        <Card>
          <CardHeader>
            <CardTitle>节点配置</CardTitle>
            <CardDescription>配置节点通信和设备限制相关选项</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="grid gap-2">
              <Label>通讯密钥</Label>
              <div class="flex items-center gap-2">
                <Input v-model="nodeConfig.server_token" type="text" placeholder="用于节点注册的通信密钥"
                  class="flex-1 font-mono" />
                <Button variant="outline" size="icon" @click="generateToken" title="随机生成">
                  <RefreshCw class="h-4 w-4" />
                </Button>
              </div>
              <p class="text-sm text-muted-foreground">节点Agent注册时需要提供的密钥，留空则不验证</p>
            </div>
            <div class="grid gap-2">
              <Label>节点拉取轮询间隔（秒）</Label>
              <Input v-model="nodeConfig.node_pull_interval" type="number" placeholder="60" />
              <p class="text-sm text-muted-foreground">节点从面板获取配置数据的间隔频率，默认60秒</p>
            </div>
            <div class="grid gap-2">
              <Label>节点推送轮询间隔（秒）</Label>
              <Input v-model="nodeConfig.node_push_interval" type="number" placeholder="60" />
              <p class="text-sm text-muted-foreground">节点推送流量数据到面板的间隔频率，默认60秒</p>
            </div>
            <div class="grid gap-2">
              <Label>设备限制模式</Label>
              <Select v-model="nodeConfig.device_limit_mode">
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="opt in deviceLimitOptions" :key="opt.value" :value="opt.value">{{ opt.label }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <p class="text-sm text-muted-foreground">宽松模式：同一IP不重复计数；严格模式：严格按设备数限制</p>
            </div>
            <Separator />
            <div class="flex items-center justify-between">
              <div>
                <Label>启用 WebSocket 通信</Label>
                <p class="text-sm text-muted-foreground">开启后面板将通过 WebSocket 实时向节点下发重启、重载等指令</p>
              </div>
              <Switch v-model="nodeConfig.websocket_enabled" />
            </div>
            <div v-if="nodeConfig.websocket_enabled" class="grid gap-2">
              <Label>WebSocket 地址</Label>
              <Input v-model="nodeConfig.websocket_url" placeholder="wss://panel.example.com/api/internal/agent/ws" />
              <p class="text-sm text-muted-foreground">节点连接面板的 WebSocket 地址，留空则自动使用站点网址</p>
            </div>
            <div class="flex justify-end">
              <Button @click="handleSave('node')" :disabled="loading">{{ loading ? '保存中...' : '保存设置' }}</Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="system">
        <Card>
          <CardHeader>
            <CardTitle>系统状态</CardTitle>
            <CardDescription>面板运行状态和系统信息</CardDescription>
          </CardHeader>
          <CardContent>
            <div v-if="!systemStatus && !sysLoading" class="flex justify-center py-4">
              <Button @click="loadSystemStatus"><RefreshCw class="mr-2 h-4 w-4" />加载系统状态</Button>
            </div>
            <div v-if="sysLoading" class="text-center py-8 text-muted-foreground">加载中...</div>
            <div v-if="systemStatus" class="space-y-6">
              <!-- 版本信息 -->
              <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <div class="flex items-center gap-3 rounded-lg border p-4">
                  <Server class="h-8 w-8 text-blue-500" />
                  <div>
                    <p class="text-sm text-muted-foreground">面板版本</p>
                    <p class="text-lg font-bold">{{ systemStatus.version }}</p>
                  </div>
                </div>
                <div class="flex items-center gap-3 rounded-lg border p-4">
                  <Clock class="h-8 w-8 text-green-500" />
                  <div>
                    <p class="text-sm text-muted-foreground">运行时间</p>
                    <p class="text-lg font-bold">{{ systemStatus.uptime }}</p>
                  </div>
                </div>
                <div class="flex items-center gap-3 rounded-lg border p-4">
                  <HardDrive class="h-8 w-8 text-purple-500" />
                  <div>
                    <p class="text-sm text-muted-foreground">数据库大小</p>
                    <p class="text-lg font-bold">{{ systemStatus.db_size_human }}</p>
                  </div>
                </div>
                <div class="flex items-center gap-3 rounded-lg border p-4">
                  <Globe class="h-8 w-8 text-orange-500" />
                  <div>
                    <p class="text-sm text-muted-foreground">Go版本</p>
                    <p class="text-lg font-bold">{{ systemStatus.go_version }}</p>
                  </div>
                </div>
              </div>

              <!-- 统计信息 -->
              <div class="rounded-lg border p-4">
                <h3 class="text-sm font-medium text-muted-foreground mb-3">统计概览</h3>
                <div class="grid gap-4 md:grid-cols-3">
                  <div>
                    <p class="text-sm text-muted-foreground">用户</p>
                    <p class="text-xl font-bold">{{ systemStatus.active_users }} <span class="text-sm text-muted-foreground font-normal">/ {{ systemStatus.total_users }}</span></p>
                    <p class="text-xs text-muted-foreground">活跃 / 总计</p>
                  </div>
                  <div>
                    <p class="text-sm text-muted-foreground">节点</p>
                    <p class="text-xl font-bold">{{ systemStatus.online_nodes }} <span class="text-sm text-muted-foreground font-normal">/ {{ systemStatus.total_nodes }}</span></p>
                    <p class="text-xs text-muted-foreground">在线 / 总计</p>
                  </div>
                  <div>
                    <p class="text-sm text-muted-foreground">在线设备 / 用户</p>
                    <p class="text-xl font-bold">{{ systemStatus.online_devices }} / {{ systemStatus.online_users }}</p>
                  </div>
                </div>
              </div>

              <div class="rounded-lg border p-4">
                <div class="flex items-center gap-2">
                  <Activity class="h-4 w-4 text-muted-foreground" />
                  <span class="text-sm text-muted-foreground">今日流量</span>
                  <span class="text-lg font-bold ml-auto">{{ formatBytes(systemStatus.today_traffic) }}</span>
                </div>
              </div>

              <p class="text-xs text-muted-foreground">启动时间: {{ systemStatus.start_time }}</p>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>

    <!-- 数据库备份 -->
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <Database class="h-5 w-5" />
          数据库备份
        </CardTitle>
        <CardDescription>导出 SQLite 数据库文件作为备份，自动保留最近 10 份</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <Button @click="handleBackup" :disabled="backupLoading" variant="outline">
          <Download class="mr-2 h-4 w-4" />
          {{ backupLoading ? '备份中...' : '立即备份并下载' }}
        </Button>
        <div v-if="backupInfo?.backups?.length > 0" class="space-y-2">
          <p class="text-sm text-muted-foreground">已有 {{ backupInfo.backups.length }} 份备份：</p>
          <div class="text-xs text-muted-foreground space-y-1">
            <div v-for="b in backupInfo.backups.slice(-5).reverse()" :key="b.name" class="flex gap-4">
              <span class="font-mono">{{ b.name }}</span>
              <span>{{ formatBytes(b.size) }}</span>
              <span>{{ new Date(b.time).toLocaleString('zh-CN') }}</span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
