<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Activity } from 'lucide-vue-next'
import { listTrafficLogs } from '@/api/monitor'
import type { TrafficLogEntry } from '@/api/monitor'

const logs = ref<TrafficLogEntry[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

function formatBytes(b: number) {
  if (b === 0) return '0 B'
  const k = 1024; const s = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(k))
  return (b / Math.pow(k, i)).toFixed(2) + ' ' + s[i]
}

async function fetchData() {
  loading.value = true
  try {
    const res = await listTrafficLogs({ page: page.value, page_size: pageSize.value })
    if (res.code === 0) { logs.value = res.data.items || []; total.value = res.data.total || 0 }
  } finally { loading.value = false }
}

onMounted(fetchData)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold flex items-center gap-2">
        <Activity class="h-5 w-5" />流量日志
      </h1>
    </div>
    <Card>
      <CardHeader><CardTitle>节点流量记录（按时间倒序）</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>时间</TableHead>
              <TableHead>用户</TableHead>
              <TableHead>节点</TableHead>
              <TableHead>上传</TableHead>
              <TableHead>下载</TableHead>
              <TableHead>合计</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="log in logs" :key="log.id">
              <TableCell class="text-sm">{{ log.recorded_at }}</TableCell>
              <TableCell class="font-medium">{{ log.user_email || '--' }}</TableCell>
              <TableCell><Badge variant="outline">{{ log.node_name || '--' }}</Badge></TableCell>
              <TableCell class="text-blue-600">{{ formatBytes(log.upload) }}</TableCell>
              <TableCell class="text-green-600">{{ formatBytes(log.download) }}</TableCell>
              <TableCell>{{ formatBytes(log.upload + log.download) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <p v-if="!loading && logs.length === 0" class="text-center text-muted-foreground py-8">暂无流量记录</p>
        <div v-if="total > 0" class="flex items-center justify-between mt-4">
          <span class="text-sm text-muted-foreground">共{{ total }}条</span>
          <div class="flex gap-2">
            <Button variant="outline" size="sm" :disabled="page <= 1" @click="page--; fetchData()">上一页</Button>
            <span class="flex items-center text-sm">第{{ page }}页</span>
            <Button variant="outline" size="sm" :disabled="page * pageSize >= total" @click="page++; fetchData()">下一页</Button>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>