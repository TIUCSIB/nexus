<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Globe } from 'lucide-vue-next'
import { listOnlineIPs } from '@/api/monitor'
import type { OnlineIP } from '@/api/monitor'

const ips = ref<OnlineIP[]>([])
const loading = ref(false)

async function fetchData() {
  loading.value = true
  try {
    const res = await listOnlineIPs()
    if (res.code === 0) ips.value = res.data || []
  } finally { loading.value = false }
}

function parseIPs(ipStr: string): string[] {
  try { return JSON.parse(ipStr) } catch { return [] }
}

onMounted(fetchData)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold flex items-center gap-2">
        <Globe class="h-5 w-5" />在线IP
      </h1>
    </div>
    <Card>
      <CardHeader><CardTitle>当前在线设备（近120秒活跃）</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>用户</TableHead>
              <TableHead>节点</TableHead>
              <TableHead>IP地址</TableHead>
              <TableHead>设备数</TableHead>
              <TableHead>最后活跃</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="ip in ips" :key="ip.id">
              <TableCell class="font-medium">{{ ip.user_email }}</TableCell>
              <TableCell><Badge variant="outline">{{ ip.node_name }}</Badge></TableCell>
              <TableCell>
                <div class="flex flex-wrap gap-1">
                  <Badge v-for="(addr, i) in parseIPs(ip.ips)" :key="i" variant="secondary" class="font-mono text-xs">
                    {{ addr }}
                  </Badge>
                </div>
              </TableCell>
              <TableCell>{{ parseIPs(ip.ips).length }}</TableCell>
              <TableCell class="text-sm text-muted-foreground">{{ ip.updated_at }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <p v-if="!loading && ips.length === 0" class="text-center text-muted-foreground py-8">暂无在线设备</p>
      </CardContent>
    </Card>
  </div>
</template>