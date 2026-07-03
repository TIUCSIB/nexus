<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { toast } from 'vue-sonner'
import { getUser } from '@/api/user'
import { getUserTrafficLogs, getUserOnlineIPs } from '@/api/trafficLog'
import type { UserDetail } from '@/api/user'
import type { TrafficLogEntry, OnlineIPEntry } from '@/api/trafficLog'

const route = useRoute()
const router = useRouter()
const detail = ref<UserDetail | null>(null)
const loading = ref(true)
const activeTab = ref('info')

const trafficLogs = ref<TrafficLogEntry[]>([])
const onlineIPs = ref<OnlineIPEntry[]>([])
const logsLoading = ref(false)
const ipsLoading = ref(false)

function formatBytes(b: number) {
  if (b === 0) return '0 B'
  const k = 1024; const s = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(k))
  return (b / Math.pow(k, i)).toFixed(2) + ' ' + s[i]
}

const trafficPercent = computed(() => {
  if (!detail.value) return 0
  const u = detail.value.user
  if (!u.traffic_limit || u.traffic_limit === 0) return 0
  return Math.min(100, Math.round((u.traffic_used / u.traffic_limit) * 100))
})

const statusText = computed(() => {
  if (!detail.value) return ''
  return detail.value.user.status === 1 ? '启用' : '禁用'
})

const statusVariant = computed(() => {
  return detail.value?.user.status === 1 ? 'default' : 'destructive'
})

function formatExpiry(d: string | null | undefined) {
  if (!d) return '永久'
  const date = new Date(d)
  const now = new Date()
  const diff = Math.ceil((date.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  return date.toLocaleDateString('zh-CN') + (diff <= 0 ? '（已过期）' : `（剩余${diff}天）`)
}

function copyLink(link: string) {
  const el = document.createElement('textarea')
  el.value = link
  document.body.appendChild(el)
  el.select()
  document.execCommand('copy')
  document.body.removeChild(el)
  toast.success('已复制')
}

function formatDate(d: string | null | undefined) {
  if (!d) return '-'
  return new Date(d).toLocaleString('zh-CN')
}

function formatDateShort(d: string) {
  if (!d) return '-'
  return new Date(d).toLocaleDateString('zh-CN')
}

async function loadTrafficLogs(userId: number) {
  logsLoading.value = true
  try {
    const res = await getUserTrafficLogs(userId)
    if (res.code === 0) trafficLogs.value = res.data
  } catch { toast.error('获取流量记录失败') }
  finally { logsLoading.value = false }
}

async function loadOnlineIPs(userId: number) {
  ipsLoading.value = true
  try {
    const res = await getUserOnlineIPs(userId)
    if (res.code === 0) onlineIPs.value = res.data
  } catch { toast.error('获取在线IP失败') }
  finally { ipsLoading.value = false }
}

function onTabChange(tab: any) {
  activeTab.value = tab
  if (!detail.value) return
  if (tab === 'traffic' && !trafficLogs.value.length) loadTrafficLogs(detail.value.user.id)
  if (tab === 'online-ips' && !onlineIPs.value.length) loadOnlineIPs(detail.value.user.id)
}

const formatLabels = ['sing-box', 'Clash', 'Surge', 'Surfboard', 'Shadowrocket', 'V2RayN']

onMounted(async () => {
  const id = Number(route.params.id)
  if (!id) { router.push('/admin/users'); return }
  try {
    const res = await getUser(id)
    if (res.code === 0) detail.value = res.data
    else toast.error(res.message || '获取用户信息失败')
  } finally { loading.value = false }
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center gap-3">
      <Button variant="ghost" size="sm" @click="router.push('/admin/users')">&larr; 返回</Button>
      <h1 class="text-2xl font-bold" v-if="detail">用户详情 &mdash; {{ detail.user.email }}</h1>
      <div v-if="detail" class="ml-auto flex items-center gap-2">
        <Badge :variant="statusVariant">{{ statusText }}</Badge>
        <Badge variant="outline">ID: {{ detail.user.id }}</Badge>
      </div>
    </div>

    <div v-if="loading" class="text-center py-12 text-muted-foreground">加载中...</div>

    <template v-if="detail">
      <div class="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader class="pb-2 flex flex-row items-center justify-between space-y-0">
            <CardTitle class="text-sm font-medium">套餐</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="text-lg font-bold">{{ detail.plan_name || '未订阅' }}</div>
            <p class="text-xs text-muted-foreground mt-1">{{ formatExpiry(detail.user.expired_at) }}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="pb-2">
            <CardTitle class="text-sm font-medium">流量使用</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="text-sm mb-2">
              {{ formatBytes(detail.user.traffic_used) }}
              <span v-if="detail.user.traffic_limit"> / {{ formatBytes(detail.user.traffic_limit) }}</span>
            </div>
            <div class="h-2 bg-muted rounded-full overflow-hidden" v-if="detail.user.traffic_limit">
              <div class="h-full bg-primary rounded-full transition-all" :style="{ width: trafficPercent + '%' }"></div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="pb-2">
            <CardTitle class="text-sm font-medium">在线设备</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="text-lg font-bold">{{ detail.ip_count }}</div>
            <p class="text-xs text-muted-foreground mt-1">
              设备限制: {{ detail.user.device_limit || '不限' }}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="pb-2 flex flex-row items-center justify-between space-y-0">
            <CardTitle class="text-sm font-medium">余额 / 权限组</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="text-lg font-bold">{{ (detail.user.balance / 100).toFixed(2) }} 元</div>
            <p class="text-xs text-muted-foreground mt-1">
              <Badge variant="outline" v-if="detail.group_name">{{ detail.group_name }}</Badge>
              <span v-else>无权限组</span>
            </p>
          </CardContent>
        </Card>
      </div>

      <Tabs default-value="info" @update:modelValue="onTabChange">
        <TabsList class="grid grid-cols-5 w-full max-w-2xl">
          <TabsTrigger value="info">基本信息</TabsTrigger>
          <TabsTrigger value="subscription">订阅链接</TabsTrigger>
          <TabsTrigger value="traffic">流量记录</TabsTrigger>
          <TabsTrigger value="online-ips">在线IP</TabsTrigger>
          <TabsTrigger value="node-traffic">节点流量</TabsTrigger>
        </TabsList>

        <TabsContent value="info" class="space-y-4">
          <Card>
            <CardContent class="py-6">
              <div class="grid gap-4 md:grid-cols-2">
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">ID</span>
                  <p class="font-mono text-sm">{{ detail.user.id }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">UUID</span>
                  <p class="font-mono text-xs break-all">{{ detail.user.uuid }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">邮箱</span>
                  <p class="text-sm">{{ detail.user.email }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">订阅令牌</span>
                  <p class="font-mono text-xs break-all">{{ detail.user.token }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">套餐</span>
                  <p class="text-sm">{{ detail.plan_name || '无' }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">权限组</span>
                  <p class="text-sm">{{ detail.group_name || '无' }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">余额</span>
                  <p class="text-sm">{{ (detail.user.balance / 100).toFixed(2) }} 元</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">状态</span>
                  <p><Badge :variant="statusVariant" class="text-xs">{{ statusText }}</Badge></p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">限速</span>
                  <p class="text-sm">{{ detail.user.speed_limit_up || '不限' }} Mbps</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">已用上行</span>
                  <p class="text-sm">{{ formatBytes(detail.user.upload_used || 0) }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">已用下行</span>
                  <p class="text-sm">{{ formatBytes(detail.user.download_used || 0) }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">设备限制</span>
                  <p class="text-sm">{{ detail.user.device_limit || '不限' }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">已用流量</span>
                  <p class="text-sm">{{ formatBytes(detail.user.traffic_used) }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">流量上限</span>
                  <p class="text-sm">{{ detail.user.traffic_limit ? formatBytes(detail.user.traffic_limit) : '不限' }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">到期时间</span>
                  <p class="text-sm">{{ formatExpiry(detail.user.expired_at) }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">上次流量重置</span>
                  <p class="text-sm">{{ formatDate(detail.user.traffic_reset_at) }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">创建时间</span>
                  <p class="text-sm">{{ formatDate(detail.user.created_at) }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">最后更新</span>
                  <p class="text-sm">{{ formatDate(detail.user.updated_at) }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">管理员</span>
                  <p class="text-sm">{{ detail.user.is_admin ? '是' : '否' }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">备注</span>
                  <p class="text-sm">{{ detail.user.remarks || '无' }}</p>
                </div>
                <div class="space-y-1">
                  <span class="text-xs text-muted-foreground uppercase tracking-wider">在线IP数</span>
                  <p class="text-sm">{{ detail.ip_count }}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="subscription" class="space-y-3">
          <Card>
            <CardHeader>
              <CardTitle class="text-sm">订阅链接</CardTitle>
            </CardHeader>
            <CardContent class="space-y-3">
              <div v-for="(link, i) in detail.links" :key="i" class="flex items-center gap-2">
                <Badge variant="outline" class="w-24 shrink-0 justify-center">{{ formatLabels[i] }}</Badge>
                <Input :model-value="link" readonly class="font-mono text-xs flex-1" />
                <Button variant="outline" size="sm" class="shrink-0" @click="copyLink(link)">
                  复制
                </Button>
              </div>
              <p v-if="!detail.links.length" class="text-muted-foreground text-center py-8">
                暂无订阅链接
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="traffic">
          <Card>
            <CardHeader>
              <CardTitle class="text-sm">流量记录（最近100条）</CardTitle>
            </CardHeader>
            <CardContent class="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>节点</TableHead>
                    <TableHead class="text-right">上传</TableHead>
                    <TableHead class="text-right">下载</TableHead>
                    <TableHead class="text-right">合计</TableHead>
                    <TableHead class="text-right">时间</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-for="log in trafficLogs" :key="log.id">
                    <TableCell>{{ log.node_name || '节点#' + log.node_id }}</TableCell>
                    <TableCell class="text-right font-mono text-xs">{{ formatBytes(log.upload) }}</TableCell>
                    <TableCell class="text-right font-mono text-xs">{{ formatBytes(log.download) }}</TableCell>
                    <TableCell class="text-right font-mono text-xs">{{ formatBytes(log.upload + log.download) }}</TableCell>
                    <TableCell class="text-right text-xs">{{ formatDate(log.recorded_at) }}</TableCell>
                  </TableRow>
                  <TableRow v-if="!trafficLogs.length && !logsLoading">
                    <TableCell colspan="5" class="text-center py-8 text-muted-foreground">暂无流量记录</TableCell>
                  </TableRow>
                </TableBody>
              </Table>
              <div v-if="logsLoading" class="text-center py-4 text-muted-foreground text-sm">加载中...</div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="online-ips">
          <Card>
            <CardHeader>
              <CardTitle class="text-sm">在线IP（最近100条）</CardTitle>
            </CardHeader>
            <CardContent class="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>IP地址</TableHead>
                    <TableHead>节点</TableHead>
                    <TableHead class="text-right">最后活跃</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-for="ip in onlineIPs" :key="ip.id">
                    <TableCell class="font-mono text-sm">{{ ip.ip }}</TableCell>
                    <TableCell>{{ ip.node_name || '节点#' + ip.node_id }}</TableCell>
                    <TableCell class="text-right text-xs">{{ formatDate(ip.updated_at) }}</TableCell>
                  </TableRow>
                  <TableRow v-if="!onlineIPs.length && !ipsLoading">
                    <TableCell colspan="3" class="text-center py-8 text-muted-foreground">暂无在线IP</TableCell>
                  </TableRow>
                </TableBody>
              </Table>
              <div v-if="ipsLoading" class="text-center py-4 text-muted-foreground text-sm">加载中...</div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="node-traffic">
          <Card>
            <CardHeader>
              <CardTitle class="text-sm">节点流量汇总</CardTitle>
            </CardHeader>
            <CardContent>
              <p class="text-muted-foreground text-center py-8">
                节点流量汇总功能开发中，将在后续版本提供
              </p>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </template>
  </div>
</template>
