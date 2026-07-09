<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Activity, Clock, HardDrive, CreditCard, TrendingUp, BarChart3, Wifi } from 'lucide-vue-next'
import { getProfile, getSubscription, getUserStats } from '@/api/userSelf'
import type { User, SubscriptionInfo, UserStats } from '@/types'
import { Bar } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend)

const user = ref<User | null>(null)
const sub = ref<SubscriptionInfo | null>(null)
const stats = ref<UserStats | null>(null)
const loading = ref(true)
const chartData = ref<any>(null)
const chartOptions = ref<any>(null)

function formatBytes(bytes: number) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const trafficPercent = computed(() => {
  if (!sub.value || !sub.value.traffic_limit || sub.value.traffic_limit === 0) return 0
  return Math.min(100, Math.round((sub.value.traffic_used / sub.value.traffic_limit) * 100))
})

const trafficText = computed(() => {
  if (!sub.value) return '--'
  if (!sub.value.traffic_limit || sub.value.traffic_limit === 0) {
    return formatBytes(sub.value.traffic_used) + ' / 不限'
  }
  return formatBytes(sub.value.traffic_used) + ' / ' + formatBytes(sub.value.traffic_limit)
})

const expireText = computed(() => {
  if (!sub.value || !sub.value.expired_at) return '永久有效'
  const d = new Date(sub.value.expired_at)
  const now = new Date()
  const diff = Math.ceil((d.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  if (diff <= 0) return '已过期'
  return d.toLocaleDateString('zh-CN') + '（剩余 ' + diff + ' 天）'
})

const expireStatus = computed(() => {
  if (!sub.value || !sub.value.expired_at) return 'default'
  const d = new Date(sub.value.expired_at)
  const now = new Date()
  const diff = Math.ceil((d.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  if (diff <= 0) return 'destructive'
  if (diff <= 7) return 'destructive'
  if (diff <= 30) return 'secondary'
  return 'default'
})

onMounted(async () => {
  try {
    const [profileRes, subRes, statsRes] = await Promise.all([getProfile(), getSubscription(), getUserStats()])
    if (profileRes.code === 0) user.value = profileRes.data
    if (subRes.code === 0) sub.value = subRes.data
    if (statsRes.code === 0) {
      stats.value = statsRes.data
      if (statsRes.data.daily_traffic?.length > 0) {
        chartData.value = {
          labels: statsRes.data.daily_traffic.map((r: any) => String(r.date).slice(5)),
          datasets: [
            { label: '上传', backgroundColor: '#3b82f6', data: statsRes.data.daily_traffic.map((r: any) => Math.round(r.upload / (1024 * 1024))), borderRadius: 4 },
            { label: '下载', backgroundColor: '#22c55e', data: statsRes.data.daily_traffic.map((r: any) => Math.round(r.download / (1024 * 1024))), borderRadius: 4 },
          ],
        }
        chartOptions.value = {
          responsive: true, maintainAspectRatio: false,
          plugins: { legend: { position: 'top' }, tooltip: { callbacks: { label: (ctx: any) => ctx.dataset.label + ': ' + formatBytes(ctx.raw * 1024 * 1024) } } },
          scales: { x: { grid: { display: false } }, y: { beginAtZero: true, ticks: { callback: (v: any) => formatBytes(v * 1024 * 1024) } } },
        }
      }
    }
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold">仪表盘</h1>

    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">当前套餐</CardTitle>
          <CreditCard class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ sub?.plan_name || '未订购' }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">到期时间</CardTitle>
          <Clock class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-lg font-bold">{{ expireText }}</div>
          <Badge v-if="expireStatus !== 'default'" :variant="expireStatus as any" class="mt-1">
            {{ expireStatus === 'destructive' ? '即将到期' : '正常' }}
          </Badge>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">流量使用</CardTitle>
          <HardDrive class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-sm text-muted-foreground mb-2">{{ trafficText }}</div>
          <Progress :model-value="trafficPercent" class="h-2" />
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">连接状态</CardTitle>
          <Activity class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">
            <Badge :variant="user?.status === 1 ? 'default' : 'destructive'">
              {{ user?.status === 1 ? '正常' : '已禁用' }}
            </Badge>
          </div>
        </CardContent>
      </Card>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>账号信息</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="grid gap-3 md:grid-cols-2">
          <div>
            <span class="text-sm text-muted-foreground">邮箱</span>
            <p class="font-medium">{{ user?.email || '--' }}</p>
          </div>
          <div>
            <span class="text-sm text-muted-foreground">注册时间</span>
            <p class="font-medium">{{ user?.created_at ? new Date(user.created_at).toLocaleDateString('zh-CN') : '--' }}</p>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Traffic Stats -->
    <div v-if="stats">
      <h2 class="text-xl font-bold mb-4">流量统计</h2>
      <div class="grid gap-4 md:grid-cols-3 mb-4">
        <Card>
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardTitle class="text-sm font-medium">今日流量</CardTitle>
            <TrendingUp class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold">{{ formatBytes(stats.today_upload + stats.today_download) }}</div>
            <p class="text-xs text-muted-foreground">上传 {{ formatBytes(stats.today_upload) }} / 下载 {{ formatBytes(stats.today_download) }}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardTitle class="text-sm font-medium">本月流量</CardTitle>
            <BarChart3 class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold">{{ formatBytes(stats.monthly_upload + stats.monthly_download) }}</div>
            <p class="text-xs text-muted-foreground">上传 {{ formatBytes(stats.monthly_upload) }} / 下载 {{ formatBytes(stats.monthly_download) }}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardTitle class="text-sm font-medium">累计流量</CardTitle>
            <HardDrive class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold">{{ formatBytes(stats.total_traffic) }}</div>
            <p class="text-xs text-muted-foreground">上传 {{ formatBytes(stats.total_upload) }} / 下载 {{ formatBytes(stats.total_download) }}</p>
          </CardContent>
        </Card>
      </div>

      <!-- Per-node traffic -->
      <Card v-if="stats.node_traffic?.length" class="mb-4">
        <CardHeader><CardTitle>节点流量分布</CardTitle></CardHeader>
        <CardContent class="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>节点</TableHead>
                <TableHead>上传</TableHead>
                <TableHead>下载</TableHead>
                <TableHead>合计</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="nt in stats.node_traffic" :key="nt.node_id">
                <TableCell class="font-medium">{{ nt.node_name }}</TableCell>
                <TableCell class="font-mono text-sm">{{ formatBytes(nt.upload) }}</TableCell>
                <TableCell class="font-mono text-sm">{{ formatBytes(nt.download) }}</TableCell>
                <TableCell class="font-mono text-sm">{{ formatBytes(nt.upload + nt.download) }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <!-- Daily traffic chart -->
      <Card>
        <CardHeader><CardTitle>流量趋势（近30天）</CardTitle></CardHeader>
        <CardContent>
          <div v-if="chartData" class="h-72">
            <Bar :data="chartData" :options="chartOptions" />
          </div>
          <p v-else class="text-muted-foreground">暂无流量数据</p>
        </CardContent>
      </Card>
    </div>
  </div>
</template>