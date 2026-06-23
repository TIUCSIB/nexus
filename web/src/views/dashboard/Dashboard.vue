<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Users, Server, Activity, HardDrive } from 'lucide-vue-next'
import { getOverview } from '@/api/stats'

const stats = ref({ total_users: 0, total_nodes: 0, online_nodes: 0, total_traffic: 0 })
const loading = ref(true)

function formatBytes(bytes: number) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

onMounted(async () => {
  try {
    const res = await getOverview()
    if (res.code === 0) stats.value = res.data
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
          <CardTitle class="text-sm font-medium">总流量</CardTitle>
          <HardDrive class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ formatBytes(stats.total_traffic) }}</div>
        </CardContent>
      </Card>
    </div>
    <Card>
      <CardHeader><CardTitle>流量趋势</CardTitle></CardHeader>
      <CardContent>
        <p class="text-muted-foreground">流量趋势图表开发中...</p>
      </CardContent>
    </Card>
  </div>
</template>