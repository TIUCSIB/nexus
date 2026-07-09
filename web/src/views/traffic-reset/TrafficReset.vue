<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Search, RotateCcw, RefreshCw, AlertTriangle, ClipboardList, Activity, Calendar } from 'lucide-vue-next'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter,
  DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { toast } from 'vue-sonner'
import { getTrafficResetUsers, manualTrafficReset, getTrafficResetStats } from '@/api/trafficReset'
import type { TrafficResetUser, TrafficResetStats, PageResult } from '@/types'

const users = ref<TrafficResetUser[]>([])
const stats = ref<TrafficResetStats>({ today_reset: 0, month_reset: 0, total_reset: 0, by_operator: [] })
const loading = ref(true)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const search = ref('')
const manualResetDialog = ref(false)
const resetting = ref(false)

function formatBytes(b: number) {
  if (b === 0) return '0 B'
  const k = 1024; const s = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(k))
  return (b / Math.pow(k, i)).toFixed(2) + ' ' + s[i]
}

function formatDate(d: string | null) {
  if (!d) return '-'
  return new Date(d).toLocaleString('zh-CN')
}

function resetTypeLabel(t: number): string {
  const labels: Record<number, string> = { 0: '不重置', 1: '每月', 2: '按周期', 3: '每年' }
  return labels[t] || '未知'
}

async function fetchData() {
  loading.value = true
  try {
    const [userRes, statsRes] = await Promise.all([
      getTrafficResetUsers({ page: page.value, page_size: pageSize.value, q: search.value }),
      getTrafficResetStats(),
    ])
    if (userRes.code === 0) {
      users.value = userRes.data.items
      total.value = userRes.data.total
    }
    if (statsRes.code === 0) stats.value = statsRes.data
  } catch { toast.error('获取数据失败') }
  finally { loading.value = false }
}

async function handleManualReset() {
  resetting.value = true
  try {
    const res = await manualTrafficReset()
    if (res.code === 0) {
      toast.success(res.data.message || '流量重置成功')
      manualResetDialog.value = false
      fetchData()
    } else { toast.error(res.message || '重置失败') }
  } catch { toast.error('操作失败') }
  finally { resetting.value = false }
}

onMounted(() => fetchData())
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">流量重置管理</h1>
      <Button variant="destructive" @click="manualResetDialog = true">
        <RefreshCw class="mr-2 h-4 w-4" />手动重置全部
      </Button>
    </div>

    <!-- Stats Cards -->
    <div class="grid gap-4 md:grid-cols-3">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">今日已重置</CardTitle>
          <Activity class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ stats.today_reset }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">本月已重置</CardTitle>
          <Calendar class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ stats.month_reset }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">累计重置</CardTitle>
          <ClipboardList class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ stats.total_reset }}</div>
        </CardContent>
      </Card>
    </div>

    <Card>
      <CardHeader class="pb-3">
        <div class="flex items-center gap-3">
          <div class="relative flex-1 max-w-sm">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input v-model="search" placeholder="搜索邮箱..." class="pl-9" @keyup.enter="fetchData" />
          </div>
          <Button variant="outline" @click="fetchData"><Search class="mr-2 h-4 w-4" />搜索</Button>
          <Button variant="ghost" size="icon" @click="search='';fetchData()"><RotateCcw class="h-4 w-4" /></Button>
        </div>
      </CardHeader>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>用户</TableHead>
              <TableHead>套餐</TableHead>
              <TableHead>重置方式</TableHead>
              <TableHead>已用流量</TableHead>
              <TableHead>总流量</TableHead>
              <TableHead>上次重置</TableHead>
              <TableHead>到期时间</TableHead>
              <TableHead>状态</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="u in users" :key="u.id">
              <TableCell class="font-medium">{{ u.email }}</TableCell>
              <TableCell>{{ u.plan_name || '—' }}</TableCell>
              <TableCell>
                <Badge variant="outline">{{ resetTypeLabel(u.plan_traffic_reset) }}</Badge>
              </TableCell>
              <TableCell class="font-mono text-sm">{{ formatBytes(u.traffic_used) }}</TableCell>
              <TableCell class="font-mono text-sm">{{ u.traffic_limit ? formatBytes(u.traffic_limit) : '不限' }}</TableCell>
              <TableCell class="text-xs">{{ formatDate(u.traffic_reset_at) }}</TableCell>
              <TableCell class="text-xs">{{ formatDate(u.expired_at) }}</TableCell>
              <TableCell>
                <Badge :variant="u.status === 1 ? 'default' : 'destructive'" class="text-xs">
                  {{ u.status === 1 ? '启用' : '禁用' }}
                </Badge>
              </TableCell>
            </TableRow>
            <TableRow v-if="!users.length && !loading">
              <TableCell colspan="8" class="text-center py-12 text-muted-foreground">暂无数据</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <!-- Manual Reset Dialog -->
    <Dialog v-model:open="manualResetDialog">
      <DialogContent class="max-w-sm">
        <DialogHeader>
          <DialogTitle class="flex items-center gap-2">
            <AlertTriangle class="h-5 w-5 text-red-500" />确认重置
          </DialogTitle>
          <DialogDescription>
            确定要手动重置所有启用用户的流量吗？此操作不可撤销，所有用户的流量计数器将归零。
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="manualResetDialog = false">取消</Button>
          <Button variant="destructive" @click="handleManualReset" :disabled="resetting">
            {{ resetting ? '重置中...' : '确认重置' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>