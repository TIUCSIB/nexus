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
import { useSettingsStore } from '@/stores/settings'
import { useRouter } from 'vue-router'
import { AlertCircle, RefreshCw } from 'lucide-vue-next'

const settingsStore = useSettingsStore()
const router = useRouter()

const loading = ref(false)
const originalSubPath = ref('s')

const site = ref({ app_name: 'Nexus', admin_path: 'admin', sub_url: '', app_description: '', force_https: false })
const subscription = ref({
  sub_path: 's', sub_show_info: true, reset_traffic_method: '0',
  new_order_event_id: '', renew_order_event_id: '',
})
const nodeConfig = ref({ server_token: '', device_limit_mode: '0', node_pull_interval: '60', node_push_interval: '60' })

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
      data = { server_token: nodeConfig.value.server_token, device_limit_mode: nodeConfig.value.device_limit_mode, node_pull_interval: nodeConfig.value.node_pull_interval, node_push_interval: nodeConfig.value.node_push_interval }
    }
    const res = await updateSettings(data)
    if (res.code === 0) {
      toast.success('保存成功')
      if (tab === 'site') {
        settingsStore.setAppName(site.value.app_name)
        settingsStore.setAppDescription(site.value.app_description)
        settingsStore.setAdminPath(site.value.admin_path)
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

onMounted(fetchSettings)
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold">系统设置</h1>
    <Tabs default-value="site">
      <TabsList>
        <TabsTrigger value="site">站点设置</TabsTrigger>
        <TabsTrigger value="subscription">订阅设置</TabsTrigger>
        <TabsTrigger value="node">节点配置</TabsTrigger>
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
            <div class="flex justify-end">
              <Button @click="handleSave('node')" :disabled="loading">{{ loading ? '保存中...' : '保存设置' }}</Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  </div>
</template>
