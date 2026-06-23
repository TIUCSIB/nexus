<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getOverview } from '@/api/stats'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Users, Server, Wifi, Activity } from '@lucide/vue'

const overview = ref({
  total_users: 0,
  total_nodes: 0,
  online_nodes: 0,
  total_traffic: 0,
})
const loading = ref(true)

function formatTraffic(bytes: number): string {
  if (bytes >= 1073741824) {
    return (bytes / 1073741824).toFixed(2) + ' GB'
  }
  if (bytes >= 1048576) {
    return (bytes / 1048576).toFixed(2) + ' MB'
  }
  return bytes + ' B'
}

onMounted(async () => {
  try {
    const res = await getOverview()
    if (res.code === 0 && res.data) {
      overview.value = res.data
    }
  } catch (err) {
    console.error('获取概览数据失败:', err)
  } finally {
    loading.value = false
  }
})

const statCards = [
  { title: '总用户数', key: 'total_users' as const, icon: Users, color: 'text-blue-500' },
  { title: '总节点数', key: 'total_nodes' as const, icon: Server, color: 'text-green-500' },
  { title: '在线节点', key: 'online_nodes' as const, icon: Wifi, color: 'text-emerald-500' },
  { title: '总流量', key: 'total_traffic' as const, icon: Activity, color: 'text-orange-500', format: true },
]
</script>

<template>
  <div class="space-y-6">
    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card v-for="card in statCards" :key="card.key">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">{{ card.title }}</CardTitle>
          <component :is="card.icon" :class="['size-4 text-muted-foreground', card.color]" />
        </CardHeader>
        <CardContent>
          <div v-if="loading" class="h-8 w-20 animate-pulse rounded bg-muted" />
          <div v-else class="text-2xl font-bold">
            {{ card.format ? formatTraffic(overview[card.key]) : overview[card.key] }}
          </div>
        </CardContent>
      </Card>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>流量统计</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="flex h-64 items-center justify-center text-muted-foreground">
          流量统计图表开发中...
        </div>
      </CardContent>
    </Card>
  </div>
</template>
