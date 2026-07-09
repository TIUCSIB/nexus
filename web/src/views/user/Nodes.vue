<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Wifi, Server, Globe, Activity } from 'lucide-vue-next'
import { listUserNodes } from '@/api/userSelf'
import type { UserNode } from '@/types'

const nodes = ref<UserNode[]>([])
const loading = ref(true)

function formatBytes(b: number) {
  if (b === 0) return '0 B'
  const k = 1024; const s = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(k))
  return (b / Math.pow(k, i)).toFixed(2) + ' ' + s[i]
}

function getProtocolColor(protocol: string): string {
  const colors: Record<string, string> = {
    vless: 'text-blue-500',
    vmess: 'text-green-500',
    trojan: 'text-purple-500',
    shadowsocks: 'text-orange-500',
    hysteria2: 'text-pink-500',
    tuic: 'text-cyan-500',
  }
  return colors[protocol.toLowerCase()] || 'text-gray-500'
}

onMounted(async () => {
  try {
    const res = await listUserNodes()
    if (res.code === 0) nodes.value = res.data
  } catch {
    // ignore
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold">节点列表</h1>

    <Card>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>节点名称</TableHead>
              <TableHead>协议</TableHead>
              <TableHead>地址</TableHead>
              <TableHead>端口</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>流量</TableHead>
              <TableHead>在线数</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="n in nodes" :key="n.id">
              <TableCell class="font-medium">
                <div class="flex items-center gap-2">
                  <Server class="h-4 w-4 text-muted-foreground" />
                  {{ n.name }}
                </div>
              </TableCell>
              <TableCell>
                <span :class="getProtocolColor(n.protocol)" class="font-mono text-sm font-medium">{{ n.protocol.toUpperCase() }}</span>
              </TableCell>
              <TableCell class="font-mono text-sm">{{ n.address }}</TableCell>
              <TableCell class="font-mono text-sm">{{ n.port }}</TableCell>
              <TableCell>
                <div class="flex items-center gap-1.5">
                  <span :class="n.online ? 'bg-green-500' : 'bg-gray-400'" class="w-2 h-2 rounded-full" />
                  <Badge :variant="n.online ? 'default' : 'secondary'" class="text-xs">
                    {{ n.online ? '在线' : '离线' }}
                  </Badge>
                </div>
              </TableCell>
              <TableCell class="font-mono text-sm">{{ formatBytes(n.traffic_used) }}</TableCell>
              <TableCell class="text-sm">{{ n.online_count }}</TableCell>
            </TableRow>
            <TableRow v-if="!nodes.length && !loading">
              <TableCell colspan="7" class="text-center py-12 text-muted-foreground">
                <div class="flex flex-col items-center gap-2">
                  <Wifi class="h-8 w-8" />
                  <p>暂无可用节点</p>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>