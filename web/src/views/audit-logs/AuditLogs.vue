<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Search, RotateCcw, ScrollText, Filter } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { listAuditLogs } from '@/api/auditLog'
import type { AuditLog } from '@/types'

const logs = ref<AuditLog[]>([])
const loading = ref(true)
const page = ref(1)
const pageSize = ref(50)
const total = ref(0)
const totalPages = ref(1)
const actionFilter = ref('')
const startDate = ref('')
const endDate = ref('')

const actionLabels: Record<string, string> = {
  create_user: '创建用户',
  update_user: '更新用户',
  delete_user: '删除用户',
  reset_uuid: '重置UUID',
  reset_user_traffic: '重置用户流量',
  batch_ban: '批量封禁',
  batch_unban: '批量解封',
  batch_delete: '批量删除',
  batch_reset_traffic: '批量重置流量',
  batch_reset_uuid: '批量重置UUID',
  manual_traffic_reset: '手动重置全部流量',
  create_node: '创建节点',
  update_node: '更新节点',
  delete_node: '删除节点',
  restart_node: '重启节点',
  reset_node_traffic: '重置节点流量',
}

function formatDate(d: string) {
  return new Date(d).toLocaleString('zh-CN')
}

function getActionLabel(action: string): string {
  return actionLabels[action] || action
}

const visiblePages = computed(() => {
  const pages: number[] = []
  const current = page.value
  const totalPg = totalPages.value
  let start = Math.max(1, current - 2)
  const end = Math.min(totalPg, start + 4)
  if (end - start < 4) start = Math.max(1, end - 4)
  for (let i = start; i <= end; i++) pages.push(i)
  return pages
})

async function fetchData() {
  loading.value = true
  try {
    const res = await listAuditLogs({
      page: page.value,
      page_size: pageSize.value,
      action: actionFilter.value && actionFilter.value !== '__all__' ? actionFilter.value : undefined,
      start_date: startDate.value || undefined,
      end_date: endDate.value || undefined,
    })
    if (res.code === 0) {
      logs.value = res.data.items
      total.value = res.data.total
      totalPages.value = Math.max(1, Math.ceil(total.value / pageSize.value))
    }
  } catch { toast.error('获取审计日志失败') }
  finally { loading.value = false }
}

onMounted(() => fetchData())
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">操作审计日志</h1>
    </div>

    <Card>
      <CardHeader class="pb-3">
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex items-center gap-2">
            <Filter class="h-4 w-4 text-muted-foreground" />
            <Select v-model="actionFilter">
              <SelectTrigger class="w-40"><SelectValue placeholder="所有操作" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="__all__">所有操作</SelectItem>
                <SelectItem v-for="(label, key) in actionLabels" :key="key" :value="key">{{ label }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-sm text-muted-foreground">开始</span>
            <Input v-model="startDate" type="date" class="w-36" />
            <span class="text-sm text-muted-foreground">结束</span>
            <Input v-model="endDate" type="date" class="w-36" />
          </div>
          <Button variant="outline" @click="fetchData"><Search class="mr-2 h-4 w-4" />搜索</Button>
          <Button variant="ghost" size="icon" @click="actionFilter='__all__';startDate='';endDate='';fetchData()"><RotateCcw class="h-4 w-4" /></Button>
        </div>
      </CardHeader>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-44">时间</TableHead>
              <TableHead>操作人</TableHead>
              <TableHead>操作类型</TableHead>
              <TableHead>操作对象</TableHead>
              <TableHead>详情</TableHead>
              <TableHead class="w-32">IP</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="log in logs" :key="log.id">
              <TableCell class="text-xs whitespace-nowrap">{{ formatDate(log.created_at) }}</TableCell>
              <TableCell class="text-sm">{{ log.user_email }}</TableCell>
              <TableCell>
                <span class="inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium">
                  {{ getActionLabel(log.action) }}
                </span>
              </TableCell>
              <TableCell class="text-xs font-mono">{{ log.target }}</TableCell>
              <TableCell class="text-xs max-w-[200px] truncate" :title="log.detail">{{ log.detail }}</TableCell>
              <TableCell class="text-xs font-mono">{{ log.ip }}</TableCell>
            </TableRow>
            <TableRow v-if="!logs.length && !loading">
              <TableCell colspan="6" class="text-center py-12 text-muted-foreground">
                <div class="flex flex-col items-center gap-2">
                  <ScrollText class="h-8 w-8" />
                  <p>暂无审计日志</p>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>

        <div class="flex items-center justify-between px-4 py-3 border-t">
          <span class="text-sm text-muted-foreground">共 {{ total }} 条</span>
          <div class="flex gap-1" v-if="totalPages > 1">
            <Button variant="outline" size="sm" :disabled="page <= 1" @click="page--;fetchData()">上一页</Button>
            <span v-for="p in visiblePages" :key="p">
              <Button :variant="p === page ? 'default' : 'outline'" size="sm" @click="page=p;fetchData()">{{ p }}</Button>
            </span>
            <Button variant="outline" size="sm" :disabled="page >= totalPages" @click="page++;fetchData()">下一页</Button>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>