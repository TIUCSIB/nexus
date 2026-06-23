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
import { Plus, Pencil, Trash2, RotateCcw } from 'lucide-vue-next'
import { listNodes, createNode, updateNode, deleteNode, restartNode } from '@/api/node'
import type { Node } from '@/types'

const nodes = ref<Node[]>([])
const total = ref(0)
const page = ref(1)
const dialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const editing = ref<Partial<Node>>({})
const isEdit = ref(false)

async function fetchData() {
  const res = await listNodes({ page: page.value, page_size: 20 })
  if (res.code === 0) { nodes.value = res.data.items; total.value = res.data.total }
}

function openCreate() { editing.value = { name: '', address: '', protocol: 'vless', port: 443, config_mode: 'auto', status: 1 }; isEdit.value = false; dialogOpen.value = true }
function openEdit(n: Node) { editing.value = { ...n }; isEdit.value = true; dialogOpen.value = true }
async function handleSave() {
  isEdit.value ? await updateNode(editing.value.id!, editing.value) : await createNode(editing.value)
  dialogOpen.value = false; fetchData()
}
function confirmDelete(n: Node) { editing.value = n; deleteDialogOpen.value = true }
async function handleDelete() { await deleteNode(editing.value.id!); deleteDialogOpen.value = false; fetchData() }
async function handleRestart(id: number) { await restartNode(id) }

onMounted(fetchData)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">节点管理</h1>
      <Button @click="openCreate"><Plus class="mr-2 h-4 w-4" />创建节点</Button>
    </div>
    <Card>
      <CardHeader />
      <CardContent>
        <Table>
          <TableHeader><TableRow>
            <TableHead>ID</TableHead><TableHead>名称</TableHead><TableHead>地址</TableHead>
            <TableHead>协议</TableHead><TableHead>端口</TableHead><TableHead>模式</TableHead>
            <TableHead>在线</TableHead><TableHead>状态</TableHead><TableHead>操作</TableHead>
          </TableRow></TableHeader>
          <TableBody>
            <TableRow v-for="n in nodes" :key="n.id">
              <TableCell>{{ n.id }}</TableCell>
              <TableCell class="font-medium">{{ n.name }}</TableCell>
              <TableCell>{{ n.address }}</TableCell>
              <TableCell><Badge variant="outline">{{ n.protocol }}</Badge></TableCell>
              <TableCell>{{ n.port }}</TableCell>
              <TableCell>{{ n.config_mode === 'auto' ? '自动' : '手动' }}</TableCell>
              <TableCell><Badge :variant="n.online ? 'default' : 'destructive'">{{ n.online ? '在线' : '离线' }}</Badge></TableCell>
              <TableCell><Badge :variant="n.status === 1 ? 'default' : 'secondary'">{{ n.status === 1 ? '启用' : '禁用' }}</Badge></TableCell>
              <TableCell>
                <div class="flex gap-1">
                  <Button variant="ghost" size="sm" @click="handleRestart(n.id)"><RotateCcw class="h-4 w-4" /></Button>
                  <Button variant="ghost" size="sm" @click="openEdit(n)"><Pencil class="h-4 w-4" /></Button>
                  <Button variant="ghost" size="sm" @click="confirmDelete(n)"><Trash2 class="h-4 w-4 text-red-500" /></Button>
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
        <DialogHeader><DialogTitle>{{ isEdit ? '编辑节点' : '创建节点' }}</DialogTitle></DialogHeader>
        <div class="grid gap-4 py-4">
          <div class="grid gap-2"><Label>名称</Label><Input v-model="editing.name" placeholder="东京-01" /></div>
          <div class="grid gap-2"><Label>地址</Label><Input v-model="editing.address" placeholder="0.0.0.0" /></div>
          <div class="grid grid-cols-2 gap-2">
            <div class="grid gap-2">
              <Label>协议</Label>
              <Select v-model="editing.protocol"><SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="vless">VLESS</SelectItem>
                  <SelectItem value="hysteria2">Hysteria2</SelectItem>
                  <SelectItem value="tuic">TUIC</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="grid gap-2"><Label>端口</Label><Input v-model.number="editing.port" type="number" /></div>
          </div>
          <div class="grid gap-2">
            <Label>配置模式</Label>
            <Select v-model="editing.config_mode"><SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="auto">自动（表单配置）</SelectItem>
                <SelectItem value="manual">手动（JSON 配置）</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="grid gap-2"><Label>配置 JSON</Label><Input v-model="editing.config_json" placeholder='{}' /></div>
        </div>
        <DialogFooter><Button @click="handleSave">保存</Button></DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="deleteDialogOpen">
      <DialogContent>
        <DialogHeader><DialogTitle>确认删除</DialogTitle></DialogHeader>
        <DialogDescription>确定要删除节点 "{{ editing.name }}" 吗？</DialogDescription>
        <DialogFooter>
          <Button variant="outline" @click="deleteDialogOpen = false">取消</Button>
          <Button variant="destructive" @click="handleDelete">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>