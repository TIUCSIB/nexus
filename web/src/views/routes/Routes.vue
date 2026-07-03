<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Plus, Pencil, Trash2 } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { listRoutes, createRoute, updateRoute, deleteRoute } from '@/api/route'
import type { RouteRule } from '@/api/route'

const routes = ref<RouteRule[]>([])
const total = ref(0)
const page = ref(1)
const dialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const editing = ref<Partial<RouteRule>>({})
const isEdit = ref(false)
const saving = ref(false)

const actionOptions = [
  { value: 'block', label: '禁止访问', color: 'destructive' as const },
  { value: 'direct', label: '直连', color: 'default' as const },
  { value: 'forward', label: '转发', color: 'secondary' as const },
  { value: 'dns', label: '指定DNS解析', color: 'outline' as const },
]

function getActionLabel(val: string) {
  return actionOptions.find(a => a.value === val)?.label || val
}

function getActionColor(val: string) {
  return actionOptions.find(a => a.value === val)?.color || 'default'
}

function countRules(match: string) {
  if (!match) return 0
  return match.split('\n').filter(l => l.trim()).length
}

async function fetchData() {
  try {
    const res = await listRoutes({ page: page.value, page_size: 20 })
    if (res.code === 0) { routes.value = res.data.items; total.value = res.data.total }
  } catch { toast.error('获取路由列表失败') }
}

function openCreate() {
  editing.value = { name: '', match: '', action: 'block', action_value: '' }
  isEdit.value = false
  dialogOpen.value = true
}

function openEdit(r: RouteRule) {
  editing.value = { ...r }
  isEdit.value = true
  dialogOpen.value = true
}

async function handleSave() {
  if (!editing.value.name) { toast.error('请输入备注'); return }
  if (!editing.value.match) { toast.error('请输入匹配规则'); return }
  saving.value = true
  try {
    if (isEdit.value) {
      const res = await updateRoute(editing.value.id!, editing.value)
      if (res.code === 0) { toast.success('路由已更新'); dialogOpen.value = false; fetchData() }
      else { toast.error(res.message || '更新失败') }
    } else {
      const res = await createRoute(editing.value)
      if (res.code === 0) { toast.success('路由已创建'); dialogOpen.value = false; fetchData() }
      else { toast.error(res.message || '创建失败') }
    }
  } catch (e: any) { toast.error(e?.response?.data?.message || '操作失败') }
  finally { saving.value = false }
}

function confirmDelete(r: RouteRule) { editing.value = r; deleteDialogOpen.value = true }

async function handleDelete() {
  try {
    const res = await deleteRoute(editing.value.id!)
    if (res.code === 0) { toast.success('路由已删除'); deleteDialogOpen.value = false; fetchData() }
    else { toast.error(res.message || '删除失败') }
  } catch (e: any) { toast.error(e?.response?.data?.message || '删除失败') }
}

onMounted(fetchData)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold">路由管理</h1>
      <p class="text-muted-foreground mt-1">管理所有路由规则，包括添加、删除、编辑等操作。</p>
    </div>
    <div class="flex items-center gap-2">
      <Button @click="openCreate"><Plus class="mr-2 h-4 w-4" />添加路由</Button>
    </div>
    <Card>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-20">组ID</TableHead>
              <TableHead class="w-32">备注</TableHead>
              <TableHead>规则数量</TableHead>
              <TableHead class="w-32">动作</TableHead>
              <TableHead class="w-24 text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="r in routes" :key="r.id">
              <TableCell><Badge variant="outline">{{ r.id }}</Badge></TableCell>
              <TableCell class="font-medium">{{ r.name }}</TableCell>
              <TableCell><span class="text-sm text-muted-foreground">匹配{{ countRules(r.match) }}条规则</span></TableCell>
              <TableCell><Badge :variant="getActionColor(r.action)">{{ getActionLabel(r.action) }}</Badge></TableCell>
              <TableCell class="text-right">
                <div class="flex gap-1 justify-end">
                  <Button variant="ghost" size="sm" @click="openEdit(r)"><Pencil class="h-4 w-4" /></Button>
                  <Button variant="ghost" size="sm" @click="confirmDelete(r)"><Trash2 class="h-4 w-4 text-red-500" /></Button>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <div class="flex items-center justify-between p-4">
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
      <DialogContent>
        <DialogHeader><DialogTitle>{{ isEdit ? '编辑路由' : '创建路由' }}</DialogTitle></DialogHeader>
        <div class="grid gap-4 py-4">
          <div class="grid gap-2">
            <Label>备注 <span class="text-red-500">*</span></Label>
            <Input v-model="editing.name" placeholder="请输入备注" />
          </div>
          <div class="grid gap-2">
            <Label>匹配规则 <span class="text-red-500">*</span></Label>
            <Textarea v-model="editing.match" rows="6" placeholder="每行一条，例如：&#10;example.com&#10;*.example.com&#10;google.com" />
          </div>
          <div class="grid gap-2">
            <Label>动作 <span class="text-red-500">*</span></Label>
            <Select v-model="editing.action">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="block">禁止访问</SelectItem>
                <SelectItem value="direct">直连</SelectItem>
                <SelectItem value="forward">转发</SelectItem>
                <SelectItem value="dns">指定DNS解析</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="grid gap-2" v-if="editing.action === 'dns'">
            <Label>DNS 服务器</Label>
            <Input v-model="editing.action_value" placeholder="例如：8.8.8.8" />
          </div>
          <div class="grid gap-2" v-if="editing.action === 'forward'">
            <Label>转发目标</Label>
            <Input v-model="editing.action_value" placeholder="例如：proxy-group-name" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="dialogOpen = false">取消</Button>
          <Button @click="handleSave" :disabled="saving">{{ saving ? '确认中...' : '确认' }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="deleteDialogOpen">
      <DialogContent>
        <DialogHeader><DialogTitle>确认删除</DialogTitle></DialogHeader>
        <DialogDescription>确定要删除路由「{{ editing.name }}」吗？</DialogDescription>
        <DialogFooter>
          <Button variant="outline" @click="deleteDialogOpen = false">取消</Button>
          <Button variant="destructive" @click="handleDelete">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>