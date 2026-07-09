<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getNodeRanking } from '@/api/stats'
import { useSettingsStore } from '@/stores/settings'
import type { NodeRankingItem } from '@/types'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { ArrowUpDown, Wifi, WifiOff } from 'lucide-vue-next'

const settingsStore = useSettingsStore()
const nodes = ref<NodeRankingItem[]>([])
const loading = ref(true)

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function protocolBadge(protocol: string): string {
  const map: Record<string, string> = {
    vless: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200',
    vmess: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
    trojan: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
    shadowsocks: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
    hysteria2: 'bg-pink-100 text-pink-800 dark:bg-pink-900 dark:text-pink-200',
    tuic: 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200',
  }
  return map[protocol] || 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
}

onMounted(async () => {
  try {
    const res = await getNodeRanking()
    nodes.value = res.data.nodes || []
  } catch (e) {
    console.error('Failed to load node ranking:', e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold">节点流量排行</h1>
        <p class="text-sm text-muted-foreground mt-1">所有节点按总流量降序排列</p>
      </div>
    </div>

    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <ArrowUpDown class="h-4 w-4" />
          节点排行
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="flex items-center justify-center py-12">
          <span class="text-muted-foreground">加载中...</span>
        </div>
        <div v-else-if="nodes.length === 0" class="flex items-center justify-center py-12">
          <span class="text-muted-foreground">暂无数据</span>
        </div>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead class="w-12">#</TableHead>
              <TableHead>节点名称</TableHead>
              <TableHead>地址</TableHead>
              <TableHead>协议</TableHead>
              <TableHead>状态</TableHead>
              <TableHead class="text-right">已用流量</TableHead>
              <TableHead class="text-right">流量上限</TableHead>
              <TableHead class="text-right">使用率</TableHead>
              <TableHead class="text-right">在线数</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="(node, index) in nodes" :key="node.id">
              <TableCell class="font-mono text-muted-foreground">{{ index + 1 }}</TableCell>
              <TableCell class="font-medium">{{ node.name }}</TableCell>
              <TableCell class="font-mono text-sm">{{ node.address }}</TableCell>
              <TableCell>
                <Badge variant="secondary" :class="protocolBadge(node.protocol)">
                  {{ node.protocol.toUpperCase() }}
                </Badge>
              </TableCell>
              <TableCell>
                <Badge :variant="node.online ? 'default' : 'secondary'">
                  <component :is="node.online ? Wifi : WifiOff" class="h-3 w-3 mr-1" />
                  {{ node.online ? '在线' : '离线' }}
                </Badge>
              </TableCell>
              <TableCell class="text-right font-mono">{{ formatBytes(node.traffic_used) }}</TableCell>
              <TableCell class="text-right font-mono">
                {{ node.traffic_limit > 0 ? formatBytes(node.traffic_limit) : '无限制' }}
              </TableCell>
              <TableCell class="text-right">
                <span v-if="node.traffic_limit > 0" class="font-mono">
                  {{ Math.round((node.traffic_used / node.traffic_limit) * 100) }}%
                </span>
                <span v-else class="text-muted-foreground">-</span>
              </TableCell>
              <TableCell class="text-right">{{ node.online_count }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>