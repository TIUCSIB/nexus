<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Input } from '@/components/ui/input'
import { toast } from 'vue-sonner'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { getMachineDetail, getMachineLoadHistory, resetMachineToken } from '@/api/machine'
import { listNodes } from '@/api/node'
import { useSettingsStore } from '@/stores/settings'
import type { Machine, Node, LoadStatus, MachineLoadHistory } from '@/types'
import { Copy, ArrowLeft, RefreshCw, HardDrive, Cpu, Activity, Wifi } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const settingsStore = useSettingsStore()

const machine = ref<Machine | null>(null)
const nodes = ref<Node[]>([])
const loadHistory = ref<MachineLoadHistory[]>([])
const loading = ref(true)
const tokenDialogOpen = ref(false)
const newToken = ref('')
const serversCount = ref(0)

// Xboard-style color mapping
const cpuColors = {
  good: 'bg-green-500',
  warn: 'bg-yellow-500',
  danger: 'bg-red-500',
}

function getCpuColor(cpu: number) {
  if (cpu < 50) return cpuColors.good
  if (cpu < 80) return cpuColors.warn
  return cpuColors.danger
}

function getMemColor(pct: number) {
  if (pct < 50) return cpuColors.good
  if (pct < 80) return cpuColors.warn
  return cpuColors.danger
}

function formatDate(t: string | null) {
  if (!t) return '从未'
  return new Date(t).toLocaleString('zh-CN')
}

function formatTimestamp(ts: number) {
  return new Date(ts * 1000).toLocaleString('zh-CN')
}

function formatBytes(b: number) {
  if (b === 0) return '0 B'
  const k = 1024; const s = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(k))
  return (b / Math.pow(k, i)).toFixed(2) + ' ' + s[i]
}

function formatSpeed(s: number) {
  if (s === 0) return '0 b/s'
  if (s < 1000) return s.toFixed(1) + ' b/s'
  if (s < 1000000) return (s / 1000).toFixed(1) + ' Kb/s'
  return (s / 1000000).toFixed(2) + ' Mb/s'
}

const onlineNodes = computed(() => nodes.value.filter(n => n.online).length)

const isOnline = computed(() => {
  if (!machine.value?.last_seen_at) return false
  return Date.now() - new Date(machine.value.last_seen_at).getTime() < 180000
})

const memPercent = computed(() => {
  const ls = machine.value?.load_status
  if (!ls || ls.mem_total === 0) return 0
  return (ls.mem_used / ls.mem_total * 100)
})

const diskPercent = computed(() => {
  const ls = machine.value?.load_status
  if (!ls || ls.disk_total === 0) return 0
  return (ls.disk_used / ls.disk_total * 100)
})

function getNodeStatus(n: Node) {
  if (n.status !== 1) return { label: '停用', color: 'bg-gray-400' }
  if (n.online) return { label: '在线', color: 'bg-green-500' }
  return { label: '离线', color: 'bg-red-400' }
}

function copyText(text: string) {
  navigator.clipboard.writeText(text)
  toast.success('已复制到剪贴板')
}

async function handleResetToken() {
  if (!machine.value) return
  try {
    const res = await resetMachineToken(machine.value.id)
    if (res.code === 0 && res.data) {
      newToken.value = res.data.token
      tokenDialogOpen.value = true
    } else {
      toast.error(res.message || '重置失败')
    }
  } catch { toast.error('重置失败') }
}

onMounted(async () => {
  const id = Number(route.params.id)
  if (!id) { router.push(settingsStore.adminRoute('machines')); return }

  try {
    const [machineRes, nodeRes, historyRes] = await Promise.all([
      getMachineDetail(id),
      listNodes({ page: 1, page_size: 100 }),
      getMachineLoadHistory(id, { limit: 30, range_hours: 6 }),
    ])

    if (machineRes.code === 0 && machineRes.data) {
      machine.value = machineRes.data.machine
      serversCount.value = machineRes.data.servers_count
    } else {
      toast.error('机器不存在')
      router.push(settingsStore.adminRoute('machines'))
      return
    }

    if (nodeRes.code === 0) {
      nodes.value = (nodeRes.data.items || []).filter((n: Node) => n.machine_id === id)
    }

    if (historyRes.code === 0 && historyRes.data) {
      loadHistory.value = historyRes.data
    }
  } catch { toast.error('获取机器信息失败') }
  finally { loading.value = false }
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center gap-3">
      <Button variant="ghost" size="sm" @click="router.push(settingsStore.adminRoute('machines'))">
        <ArrowLeft class="h-4 w-4 mr-1" />返回
      </Button>
      <div v-if="machine" class="flex items-center gap-3 flex-1">
        <h1 class="text-2xl font-bold">{{ machine.name }}</h1>
        <Badge :variant="machine.is_active ? 'default' : 'secondary'">
          {{ machine.is_active ? '启用' : '停用' }}
        </Badge>
        <Badge variant="outline" class="gap-1">
          <span :class="isOnline ? 'bg-green-500' : 'bg-gray-400'" class="w-2 h-2 rounded-full" />
          {{ isOnline ? '在线' : '离线' }}
        </Badge>
        <Badge variant="outline">ID: {{ machine.id }}</Badge>
      </div>
    </div>

    <div v-if="loading" class="text-center py-12 text-muted-foreground">加载中...</div>

    <template v-if="machine">
      <!-- Overview Cards -->
      <div class="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader class="pb-2 flex flex-row items-center justify-between">
            <CardTitle class="text-sm font-medium">管理节点</CardTitle>
            <HardDrive class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold">{{ nodes.length }}</div>
            <p class="text-xs text-muted-foreground mt-1">台节点</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="pb-2 flex flex-row items-center justify-between">
            <CardTitle class="text-sm font-medium">在线节点</CardTitle>
            <Activity class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold">{{ onlineNodes }}</div>
            <p class="text-xs text-muted-foreground mt-1">
              {{ onlineNodes > 0 ? '正常运行中' : '无在线节点' }}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="pb-2 flex flex-row items-center justify-between">
            <CardTitle class="text-sm font-medium">最后活跃</CardTitle>
            <RefreshCw class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div class="text-sm font-medium">{{ formatDate(machine.last_seen_at) }}</div>
            <p class="text-xs text-muted-foreground mt-1">
              {{ isOnline ? '当前在线' : '已离线' }}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader class="pb-2 flex flex-row items-center justify-between">
            <CardTitle class="text-sm font-medium">备注</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="text-sm">{{ machine.notes || '无' }}</div>
          </CardContent>
        </Card>
      </div>

      <!-- System Load Section (Xboard-style) -->
      <div v-if="machine.load_status" class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <!-- CPU -->
        <Card>
          <CardHeader class="pb-2 flex flex-row items-center justify-between">
            <CardTitle class="text-sm font-medium">CPU 使用率</CardTitle>
            <Cpu class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div class="flex items-end gap-2">
              <span class="text-2xl font-bold">{{ machine.load_status.cpu.toFixed(1) }}%</span>
            </div>
            <div class="mt-2 h-2 rounded-full bg-muted overflow-hidden">
              <div :class="getCpuColor(machine.load_status.cpu)" class="h-full rounded-full transition-all" :style="{ width: Math.min(machine.load_status.cpu, 100) + '%' }" />
            </div>
            <p class="text-xs text-muted-foreground mt-2">
              {{ machine.load_status.cpu < 50 ? '负载正常' : machine.load_status.cpu < 80 ? '负载较高' : '负载过高' }}
            </p>
          </CardContent>
        </Card>

        <!-- Memory -->
        <Card>
          <CardHeader class="pb-2 flex flex-row items-center justify-between">
            <CardTitle class="text-sm font-medium">内存使用</CardTitle>
            <HardDrive class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div class="flex items-end gap-2">
              <span class="text-2xl font-bold">{{ memPercent.toFixed(1) }}%</span>
            </div>
            <div class="mt-2 h-2 rounded-full bg-muted overflow-hidden">
              <div :class="getMemColor(memPercent)" class="h-full rounded-full transition-all" :style="{ width: Math.min(memPercent, 100) + '%' }" />
            </div>
            <p class="text-xs text-muted-foreground mt-2">
              {{ formatBytes(machine.load_status.mem_used) }} / {{ formatBytes(machine.load_status.mem_total) }}
            </p>
          </CardContent>
        </Card>

        <!-- Disk -->
        <Card>
          <CardHeader class="pb-2 flex flex-row items-center justify-between">
            <CardTitle class="text-sm font-medium">磁盘使用</CardTitle>
            <HardDrive class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div class="flex items-end gap-2">
              <span class="text-2xl font-bold">{{ diskPercent.toFixed(1) }}%</span>
            </div>
            <div class="mt-2 h-2 rounded-full bg-muted overflow-hidden">
              <div :class="getMemColor(diskPercent)" class="h-full rounded-full transition-all" :style="{ width: Math.min(diskPercent, 100) + '%' }" />
            </div>
            <p class="text-xs text-muted-foreground mt-2">
              {{ formatBytes(machine.load_status.disk_used) }} / {{ formatBytes(machine.load_status.disk_total) }}
            </p>
          </CardContent>
        </Card>

        <!-- Network -->
        <Card>
          <CardHeader class="pb-2 flex flex-row items-center justify-between">
            <CardTitle class="text-sm font-medium">网络 IO</CardTitle>
            <Wifi class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div class="space-y-1">
              <div class="flex items-center justify-between text-xs">
                <span class="text-muted-foreground">入站</span>
                <span class="font-mono font-medium">{{ formatSpeed(machine.load_status.net_in_speed) }}</span>
              </div>
              <div class="flex items-center justify-between text-xs">
                <span class="text-muted-foreground">出站</span>
                <span class="font-mono font-medium">{{ formatSpeed(machine.load_status.net_out_speed) }}</span>
              </div>
            </div>
            <Separator class="my-2" />
            <p class="text-xs text-muted-foreground">实时速率</p>
          </CardContent>
        </Card>
      </div>

      <!-- Load History (mini chart) -->
      <Card v-if="loadHistory.length > 0">
        <CardHeader>
          <CardTitle class="text-sm">CPU 负载历史（最近 {{ loadHistory.length }} 条）</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="flex items-end gap-[2px] h-16">
            <div v-for="h in loadHistory" :key="h.id" class="flex-1 flex flex-col justify-end h-full">
              <div
                :class="getCpuColor(h.cpu)"
                class="rounded-t transition-all duration-300 hover:opacity-80 relative group"
                :style="{ height: Math.max(h.cpu, 2) + '%' }"
              >
                <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 hidden group-hover:block bg-popover text-popover-foreground text-xs rounded px-1.5 py-0.5 whitespace-nowrap z-10 shadow">
                  {{ h.cpu.toFixed(1) }}% @ {{ formatTimestamp(h.recorded_at) }}
                </div>
              </div>
            </div>
          </div>
          <div class="flex items-center justify-between mt-2 text-xs text-muted-foreground">
            <span>{{ formatTimestamp(loadHistory[0]?.recorded_at) }}</span>
            <span>{{ formatTimestamp(loadHistory[loadHistory.length - 1]?.recorded_at) }}</span>
          </div>
        </CardContent>
      </Card>

      <!-- Actions -->
      <div class="flex items-center gap-2">
        <Button variant="outline" @click="handleResetToken">
          <RefreshCw class="h-4 w-4 mr-1" />重置 Token
        </Button>
      </div>

      <!-- Node List -->
      <Card>
        <CardHeader>
          <CardTitle>管理的节点（{{ nodes.length }}）</CardTitle>
        </CardHeader>
        <CardContent class="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead class="w-16">ID</TableHead>
                <TableHead>节点名称</TableHead>
                <TableHead>地址</TableHead>
                <TableHead>协议</TableHead>
                <TableHead>状态</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="n in nodes" :key="n.id" class="cursor-pointer hover:bg-muted/50" @click="router.push(settingsStore.adminRoute('nodes'))">
                <TableCell class="font-mono text-xs">{{ n.id }}</TableCell>
                <TableCell class="font-medium">
                  <div class="flex items-center gap-2">
                    <span :class="getNodeStatus(n).color" class="w-2 h-2 rounded-full shrink-0" />
                    <span>{{ n.name }}</span>
                  </div>
                </TableCell>
                <TableCell class="font-mono text-sm">{{ n.address }}:{{ n.port }}</TableCell>
                <TableCell><Badge variant="secondary">{{ n.protocol }}</Badge></TableCell>
                <TableCell>
                  <Badge v-if="n.status === 1 && n.online" variant="default" class="bg-green-600">在线</Badge>
                  <Badge v-else-if="n.status === 1" variant="secondary">离线</Badge>
                  <Badge v-else variant="destructive">停用</Badge>
                </TableCell>
              </TableRow>
              <TableRow v-if="nodes.length === 0">
                <TableCell colspan="5" class="text-center py-8 text-muted-foreground">
                  暂无节点，请先在节点管理中绑定到此机器
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </template>

    <!-- Token Dialog -->
    <Dialog v-model:open="tokenDialogOpen">
      <DialogContent class="max-w-lg">
        <DialogHeader>
          <DialogTitle>新的 Token</DialogTitle>
          <DialogDescription>旧的 Token 将立即失效，请及时更新 Agent 配置</DialogDescription>
        </DialogHeader>
        <div class="py-4">
          <div class="flex items-center gap-2">
            <code class="flex-1 rounded-md bg-muted p-3 text-sm font-mono break-all">{{ newToken }}</code>
            <Button variant="outline" size="icon" @click="copyText(newToken)">
              <Copy class="h-4 w-4" />
            </Button>
          </div>
        </div>
        <DialogFooter>
          <Button @click="tokenDialogOpen = false">关闭</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>