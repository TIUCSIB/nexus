<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Plus, Pencil, Trash2 } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { listPlans, createPlan, updatePlan, deletePlan } from '@/api/plan'
import { listGroups } from '@/api/group'
import type { Plan } from '@/types'
import type { ServerGroup } from '@/api/group'

const plans = ref<Plan[]>([])
const groups = ref<ServerGroup[]>([])
const total = ref(0)
const page = ref(1)
const dialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const editing = ref<Partial<Plan>>({})
const isEdit = ref(false)
const saving = ref(false)

function formatTraffic(b: number) {
  if (b === 0) return '不限'
  const k = 1024; const s = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(k))
  return (b / Math.pow(k, i)).toFixed(1) + ' ' + s[i]
}

function toPlanGB(bytes: number | null | undefined): number | null {
  if (!bytes || bytes === 0) return null
  return Math.round((bytes / (1024 * 1024 * 1024)) * 100) / 100
}

function toPlanBytes(gb: number | null | undefined): number {
  if (!gb || gb === 0) return 0
  return Math.round(gb * 1024 * 1024 * 1024)
}

const trafficUnit = ref('GB')

async function fetchData() {
  try {
    const res = await listPlans({ page: page.value, page_size: 20 })
    if (res.code === 0) { plans.value = res.data.items; total.value = res.data.total }
  } catch { toast.error('获取套餐列表失败') }
}

async function fetchGroups() {
  try { const res = await listGroups(); if (res.code === 0) groups.value = res.data || [] } catch {}
}

function getGroupName(id: number | null) {
  if (!id) return '未分组'
  const g = groups.value.find(g => g.id === id)
  return g ? g.name : '未知'
}

function openCreate() {
  editing.value = { name: '', description: '', group_id: null, traffic_limit: null, duration_days: 30, price: 0, speed_limit: null, device_limit: null, capacity_limit: null, traffic_reset: 0, status: 1 }
  trafficUnit.value = 'GB'
  isEdit.value = false
  dialogOpen.value = true
}

function openEdit(p: Plan) {
  editing.value = { ...p, traffic_limit: toPlanGB(p.traffic_limit) }
  trafficUnit.value = 'GB'
  isEdit.value = true
  dialogOpen.value = true
}

async function handleSave() {
  saving.value = true
  try {
    const data: any = { ...editing.value }
    data.traffic_limit = toPlanBytes(data.traffic_limit)
    data.speed_limit = data.speed_limit ? Number(data.speed_limit) : 0
    data.device_limit = data.device_limit ? Number(data.device_limit) : 0
    data.capacity_limit = data.capacity_limit ? Number(data.capacity_limit) : 0
    if (isEdit.value) {
      const res = await updatePlan(editing.value.id!, data)
      if (res.code === 0) { toast.success('套餐已更新'); dialogOpen.value = false; fetchData() }
      else { toast.error(res.message || '更新失败') }
    } else {
      const res = await createPlan(data)
      if (res.code === 0) { toast.success('套餐已创建'); dialogOpen.value = false; fetchData() }
      else { toast.error(res.message || '创建失败') }
    }
  } catch (e: any) { toast.error(e?.response?.data?.message || '操作失败，请重试') }
  finally { saving.value = false }
}

function confirmDelete(p: Plan) { editing.value = p; deleteDialogOpen.value = true }

async function handleDelete() {
  try {
    const res = await deletePlan(editing.value.id!)
    if (res.code === 0) { toast.success('套餐已删除'); deleteDialogOpen.value = false; fetchData() }
    else { toast.error(res.message || '删除失败') }
  } catch (e: any) { toast.error(e?.response?.data?.message || '删除失败') }
}

onMounted(() => { fetchData(); fetchGroups() })
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">套餐管理</h1>
      <Button @click="openCreate"><Plus class="mr-2 h-4 w-4" />创建套餐</Button>
    </div>
    <Card>
      <CardHeader />
      <CardContent>
        <Table>
          <TableHeader><TableRow>
            <TableHead>ID</TableHead><TableHead>名称</TableHead><TableHead>权限组</TableHead>
            <TableHead>流量</TableHead><TableHead>有效期</TableHead><TableHead>价格(元)</TableHead>
            <TableHead>限速</TableHead><TableHead>设备</TableHead><TableHead>状态</TableHead><TableHead>操作</TableHead>
          </TableRow></TableHeader>
          <TableBody>
            <TableRow v-for="p in plans" :key="p.id">
              <TableCell>{{ p.id }}</TableCell>
              <TableCell class="font-medium">{{ p.name }}</TableCell>
              <TableCell><Badge variant="outline">{{ getGroupName(p.group_id) }}</Badge></TableCell>
              <TableCell>{{ formatTraffic(p.traffic_limit) }}</TableCell>
              <TableCell>{{ p.duration_days }}天</TableCell>
              <TableCell>{{ (p.price / 100).toFixed(2) }}</TableCell>
              <TableCell>{{ p.speed_limit ? p.speed_limit + ' Mbps' : '不限' }}</TableCell>
              <TableCell>{{ p.device_limit ? p.device_limit + ' 台' : '不限' }}</TableCell>
              <TableCell><Badge :variant="p.status === 1 ? 'default' : 'secondary'">{{ p.status === 1 ? '上架' : '下架' }}</Badge></TableCell>
              <TableCell>
                <div class="flex gap-1">
                  <Button variant="ghost" size="sm" @click="openEdit(p)"><Pencil class="h-4 w-4" /></Button>
                  <Button variant="ghost" size="sm" @click="confirmDelete(p)"><Trash2 class="h-4 w-4 text-red-500" /></Button>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <div class="flex items-center justify-between mt-4">
          <span class="text-sm text-muted-foreground">共{{ total }}条</span>
          <div class="flex gap-2">
            <Button variant="outline" size="sm" :disabled="page <= 1" @click="page--; fetchData()">上一页</Button>
            <span class="flex items-center text-sm">第{{ page }}页</span>
            <Button variant="outline" size="sm" :disabled="page * 20 >= total" @click="page++; fetchData()">下一页</Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <Dialog v-model:open="dialogOpen">
      <DialogContent class="max-w-2xl">
        <DialogHeader><DialogTitle>{{ isEdit ? '编辑套餐' : '创建套餐' }}</DialogTitle></DialogHeader>
        <div class="grid gap-4 py-4 max-h-[60vh] overflow-y-auto">
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2"><Label>套餐名称</Label><Input v-model="editing.name" /></div>
            <div class="grid gap-2">
              <Label>服务器分组</Label>
              <Select :model-value="editing.group_id?.toString() || ''" @update:model-value="v => editing.group_id = v ? Number(v) : null">
                <SelectTrigger><SelectValue placeholder="请选择分组" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">不分组</SelectItem>
                  <SelectItem v-for="g in groups" :key="g.id" :value="g.id.toString()">{{ g.name }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div class="grid gap-2"><Label>描述</Label><Input v-model="editing.description" /></div>
          <div class="grid grid-cols-3 gap-4">
            <div class="grid gap-2"><Label>流量</Label>
              <div class="relative">
                <Input v-model.number="editing.traffic_limit" type="number" min="0" class="pr-12" placeholder="留空则不限" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">GB</span>
              </div>
            </div>
            <div class="grid gap-2"><Label>有效天数</Label><Input v-model.number="editing.duration_days" type="number" /></div>
            <div class="grid gap-2"><Label>价格(分)</Label><Input v-model.number="editing.price" type="number" /></div>
          </div>
          <div class="grid grid-cols-3 gap-4">
            <div class="grid gap-2"><Label>限速</Label>
              <div class="relative">
                <Input v-model.number="editing.speed_limit" type="number" min="0" class="pr-16" placeholder="留空则不限" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">Mbps</span>
              </div>
            </div>
            <div class="grid gap-2"><Label>设备限制</Label>
              <div class="relative">
                <Input v-model.number="editing.device_limit" type="number" min="0" class="pr-10" placeholder="留空则不限" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">台</span>
              </div>
            </div>
            <div class="grid gap-2"><Label>容量限制</Label>
              <div class="relative">
                <Input v-model.number="editing.capacity_limit" type="number" min="0" class="pr-10" placeholder="留空则不限" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">人</span>
              </div>
            </div>
          </div>
          <div class="grid gap-2">
            <Label>流量重置方式</Label>
            <Select :model-value="(editing.traffic_reset || 0).toString()" @update:model-value="v => editing.traffic_reset = Number(v)">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="0">跟随系统设置</SelectItem>
                <SelectItem value="1">每月1号重置</SelectItem>
                <SelectItem value="2">不重置</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="dialogOpen = false">取消</Button>
          <Button @click="handleSave" :disabled="saving">{{ saving ? '保存中...' : '保存' }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="deleteDialogOpen">
      <DialogContent>
        <DialogHeader><DialogTitle>确认删除</DialogTitle></DialogHeader>
        <DialogDescription>确定要删除套餐「{{ editing.name }}」吗？</DialogDescription>
        <DialogFooter>
          <Button variant="outline" @click="deleteDialogOpen = false">取消</Button>
          <Button variant="destructive" @click="handleDelete">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>