<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Users, Server, Activity, HardDrive, TrendingUp, Wifi, UserCheck, Calendar, ArrowUpDown } from 'lucide-vue-next'
import { useSettingsStore } from '@/stores/settings'
import { getOverview, getTraffic } from '@/api/stats'
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

const router = useRouter()
const settingsStore = useSettingsStore()
const stats = ref<any>({
  total_users: 0, total_nodes: 0, online_nodes: 0, total_traffic: 0, total_upload: 0, total_download: 0,
  today_upload: 0, today_download: 0, today_traffic: 0,
  online_devices: 0, online_users: 0, monthly_traffic: 0, yesterday_ranking: [],
})
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

onMounted(async () => {
  try {
    const [overviewRes, trafficRes] = await Promise.all([getOverview(), getTraffic(14)])
    if (overviewRes.code === 0) stats.value = overviewRes.data
    
    if (trafficRes.code === 0 && trafficRes.data.records?.length > 0) {
      const records = trafficRes.data.records
      chartData.value = {
        labels: records.map((r: any) => r.date.slice(5)),
        datasets: [
          {
            label: '上传',
            backgroundColor: '#3b82f6',
            data: records.map((r: any) => Math.round(r.upload / (1024 * 1024))),
            borderRadius: 4,
          },
          {
            label: '下载',
            backgroundColor: '#22c55e',
            data: records.map((r: any) => Math.round(r.download / (1024 * 1024))),
            borderRadius: 4,
          },
        ],
      }
      chartOptions.value = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { position: 'top' },
          tooltip: {
            callbacks: {
              label: (ctx: any) => ctx.dataset.label + ': ' + formatBytes(ctx.raw * 1024 * 1024),
            },
          },
        },
        scales: {
          x: { grid: { display: false } },
          y: {
            beginAtZero: true,
            ticks: { callback: (v: any) => formatBytes(v * 1024 * 1024) },
          },
        },
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
    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">用户总数</CardTitle>
          <Users class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ stats.total_users }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">节点总数</CardTitle>
          <Server class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ stats.total_nodes }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">在线节点</CardTitle>
          <Activity class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-green-600">{{ stats.online_nodes }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">在线用户</CardTitle>
          <UserCheck class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ stats.online_users }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">在线设备</CardTitle>
          <Wifi class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ stats.online_devices }}</div>
        </CardContent>
      </Card>
    </div>
    <div class="grid gap-4 md:grid-cols-3">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">今日流量</CardTitle>
          <TrendingUp class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ formatBytes(stats.today_traffic) }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">月度流量</CardTitle>
          <Calendar class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ formatBytes(stats.monthly_traffic) }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">总流量</CardTitle>
          <HardDrive class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ formatBytes(stats.total_traffic) }}</div>
        </CardContent>
      </Card>
    </div>
    <Card>
      <CardHeader><CardTitle>流量趋势（近14天）</CardTitle></CardHeader>
      <CardContent>
        <div v-if="chartData" class="h-80">
          <Bar :data="chartData" :options="chartOptions" />
        </div>
        <p v-else class="text-muted-foreground">暂无流量数据</p>
      </CardContent>
    </Card>

    <!-- Yesterday Node Ranking -->
    <Card v-if="stats.yesterday_ranking?.length">
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <ArrowUpDown class="h-4 w-4" />昨日节点排行
        </CardTitle>
      </CardHeader>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-12">排名</TableHead>
              <TableHead>节点</TableHead>
              <TableHead>上传</TableHead>
              <TableHead>下载</TableHead>
              <TableHead>合计</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="(n, i) in stats.yesterday_ranking" :key="n.node_id" class="cursor-pointer hover:bg-muted/50" @click="router.push(settingsStore.adminRoute('nodes'))">
              <TableCell>
                <span class="inline-flex items-center justify-center w-6 h-6 rounded-full text-xs font-bold" :class="i === 0 ? 'bg-yellow-100 text-yellow-700' : i === 1 ? 'bg-gray-100 text-gray-600' : i === 2 ? 'bg-orange-100 text-orange-700' : 'bg-muted text-muted-foreground'">
                  {{ i + 1 }}
                </span>
              </TableCell>
              <TableCell class="font-medium">{{ n.name }}</TableCell>
              <TableCell class="font-mono text-sm">{{ formatBytes(n.upload) }}</TableCell>
              <TableCell class="font-mono text-sm">{{ formatBytes(n.download) }}</TableCell>
              <TableCell class="font-mono text-sm font-medium">{{ formatBytes(n.total) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>