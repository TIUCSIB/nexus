<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getUserRanking } from '@/api/stats'
import { useSettingsStore } from '@/stores/settings'
import type { UserRankingItem } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ArrowUpDown, Users } from 'lucide-vue-next'

const settingsStore = useSettingsStore()
const users = ref<UserRankingItem[]>([])
const loading = ref(true)
const limit = ref(20)

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

async function fetchData() {
  loading.value = true
  try {
    const res = await getUserRanking(limit.value)
    users.value = res.data.users || []
  } catch (e) {
    console.error('Failed to load user ranking:', e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold">用户流量排行</h1>
        <p class="text-sm text-muted-foreground mt-1">活跃用户按总流量降序排列</p>
      </div>
      <div class="flex items-center gap-2">
        <span class="text-sm text-muted-foreground">显示</span>
        <Select v-model="limit" @update:model-value="fetchData">
          <SelectTrigger class="w-20">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="10">10</SelectItem>
            <SelectItem value="20">20</SelectItem>
            <SelectItem value="50">50</SelectItem>
            <SelectItem value="100">100</SelectItem>
          </SelectContent>
        </Select>
        <span class="text-sm text-muted-foreground">条</span>
      </div>
    </div>

    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <Users class="h-4 w-4" />
          用户排行
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="flex items-center justify-center py-12">
          <span class="text-muted-foreground">加载中...</span>
        </div>
        <div v-else-if="users.length === 0" class="flex items-center justify-center py-12">
          <span class="text-muted-foreground">暂无数据</span>
        </div>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead class="w-12">#</TableHead>
              <TableHead>邮箱</TableHead>
              <TableHead>UUID</TableHead>
              <TableHead class="text-right">上传</TableHead>
              <TableHead class="text-right">下载</TableHead>
              <TableHead class="text-right">总计</TableHead>
              <TableHead class="text-right">流量上限</TableHead>
              <TableHead class="text-right">使用率</TableHead>
              <TableHead class="text-right">设备限制</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="(user, index) in users" :key="user.id">
              <TableCell class="font-mono text-muted-foreground">{{ index + 1 }}</TableCell>
              <TableCell class="font-medium">{{ user.email }}</TableCell>
              <TableCell class="font-mono text-xs text-muted-foreground max-w-[160px] truncate" :title="user.uuid">
                {{ user.uuid }}
              </TableCell>
              <TableCell class="text-right font-mono">{{ formatBytes(user.upload_used) }}</TableCell>
              <TableCell class="text-right font-mono">{{ formatBytes(user.download_used) }}</TableCell>
              <TableCell class="text-right font-mono font-medium">{{ formatBytes(user.traffic_used) }}</TableCell>
              <TableCell class="text-right font-mono">
                {{ user.traffic_limit > 0 ? formatBytes(user.traffic_limit) : '无限制' }}
              </TableCell>
              <TableCell class="text-right">
                <span v-if="user.traffic_limit > 0" class="font-mono">
                  {{ Math.round((user.traffic_used / user.traffic_limit) * 100) }}%
                </span>
                <span v-else class="text-muted-foreground">-</span>
              </TableCell>
              <TableCell class="text-right">{{ user.device_limit > 0 ? user.device_limit : '不限' }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>