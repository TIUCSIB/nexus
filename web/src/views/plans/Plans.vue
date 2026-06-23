<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Plus, Pencil, Trash2 } from 'lucide-vue-next'
import { listPlans, createPlan, updatePlan, deletePlan } from '@/api/plan'
import type { Plan } from '@/types'

const plans = ref<Plan[]>([])
const total = ref(0)
const page = ref(1)
const dialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const editing = ref<Partial<Plan>>({})
const isEdit = ref(false)

function formatTraffic(b: number) {
  if (b === 0) return '不限'
  const k = 1024, s = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(k))
  return (b / Math.pow(k, i)).toFixed(1) + ' ' + s[i]
}

async function fetchData() {
  const res = await listPlans({ page: page.value, page_size: 20 })
  if (res.code === 0) { plans.value = res.data.items; total.value = res.data.total }
}

function openCreate() { editing.value = { name: '', description: '', traffic_limit: 0, duration_days: 30, price: 0, status: 1 }; isEdit.value = false; dialogOpen.value = true }
function openEdit(p: Plan) { editing.value = { ...p }; isEdit.value = true; dialogOpen.value = true }
async function handleSave() {
  isEdit.value ? await updatePlan(editing.value.id!, editing.value) : await createPlan(editing.value)
  dialogOpen.value = false; fetchData()
}
function confirmDelete(p: Plan) { editing.value = p; deleteDialogOpen.value = true }
async function handleDelete() { await deletePlan(editing.value.id!); deleteDialogOpen.value = false; fetchData() }

onMounted(fetchData)
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
            <TableHead>ID</TableHead><TableHead>名称</TableHead><TableHead>流量</TableHead>
            <TableHead>有效期</TableHead><TableHead>价格(分)</TableHead><TableHead>状态</TableHead><TableHead>操作</TableHead>
          </TableRow></TableHeader>
          <TableBody>
            <TableRow v-for="p in plans" :key="p.id">
              <TableCell>{{ p.id }}</TableCell>
              <TableCell class="font-medium">{{ p.name }}</TableCell>
              <TableCell>{{ formatTraffic(p.traffic_limit) }}</TableCell>
              <TableCell>{{ p.duration_days }}天</TableCell>
              <TableCell>{{ (p.price / 100).toFixed(2) }}</TableCell>
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
          <span class="text-sm text-muted-foreground">共 {{ total }} 条</span>
          <div class="flex gap-2">
            <Button variant="outline" size="sm" :disabled="page <= 1" @click="page--; fetchData()">上一页</Button>
            <span class="flex items-center text-sm">第 {{ page }} 页</span>
            <Button variant="outline" size="sm" :disabled="page * 20 >= total" @click="page++; fetchData()">下一页</Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <Dialog v-model:open="dialogOpen">
      <DialogContent>
        <DialogHeader><DialogTitle>{{ isEdit ? '编辑套餐' : '创建套餐' }}</DialogTitle></DialogHeader>
        <div class="grid gap-4 py-4">
          <div class="grid gap-2"><Label>名称</Label><Input v-model="editing.name" /></div>
          <div class="grid gap-2"><Label>描述</Label><Input v-model="editing.description" /></div>
          <div class="grid grid-cols-2 gap-2">
            <div class="grid gap-2"><Label>流量上限(字节)</Label><Input v-model.number="editing.traffic_limit" type="number" /></div>
            <div class="grid gap-2"><Label>有效天数</Label><Input v-model.number="editing.duration_days" type="number" /></div>
          </div>
          <div class="grid gap-2"><Label>价格(分)</Label><Input v-model.number="editing.price" type="number" /></div>
        </div>
        <DialogFooter><Button @click="handleSave">保存</Button></DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="deleteDialogOpen">
      <DialogContent>
        <DialogHeader><DialogTitle>确认删除</DialogTitle></DialogHeader>
        <DialogDescription>确定要删除套餐 "{{ editing.name }}" 吗？</DialogDescription>
        <DialogFooter>
          <Button variant="outline" @click="deleteDialogOpen = false">取消</Button>
          <Button variant="destructive" @click="handleDelete">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>