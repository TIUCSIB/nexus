<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Plus, Pencil, Trash2, MoreHorizontal, Users, Server } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { listGroups, createGroup, updateGroup, deleteGroup } from '@/api/group'
import type { ServerGroup } from '@/api/group'

const groups = ref<ServerGroup[]>([])
const dialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const editing = ref<Partial<ServerGroup>>({})
const isEdit = ref(false)
const saving = ref(false)

async function fetchData() {
  try {
    const res = await listGroups()
    if (res.code === 0) groups.value = res.data || []
  } catch { toast.error('获取权限组列表失败') }
}

function openCreate() {
  editing.value = { name: '' }
  isEdit.value = false
  dialogOpen.value = true
}

function openEdit(g: ServerGroup) {
  editing.value = { ...g }
  isEdit.value = true
  dialogOpen.value = true
}

async function handleSave() {
  if (!editing.value.name) { toast.error('请输入权限组名称'); return }
  saving.value = true
  try {
    if (isEdit.value) {
      const res = await updateGroup(editing.value.id!, { name: editing.value.name })
      if (res.code === 0) { toast.success('权限组已更新'); dialogOpen.value = false; fetchData() }
      else { toast.error(res.message || '更新失败') }
    } else {
      const res = await createGroup({ name: editing.value.name })
      if (res.code === 0) { toast.success('权限组已创建'); dialogOpen.value = false; fetchData() }
      else { toast.error(res.message || '创建失败') }
    }
  } catch (e: any) { toast.error(e?.response?.data?.message || '操作失败，请重试') }
  finally { saving.value = false }
}

function confirmDelete(g: ServerGroup) { editing.value = g; deleteDialogOpen.value = true }

async function handleDelete() {
  try {
    const res = await deleteGroup(editing.value.id!)
    if (res.code === 0) { toast.success('权限组已删除'); deleteDialogOpen.value = false; fetchData() }
    else { toast.error(res.message || '删除失败') }
  } catch (e: any) { toast.error(e?.response?.data?.message || '删除失败，请重试') }
}

onMounted(fetchData)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold">权限组管理</h1>
        <p class="text-muted-foreground mt-1">管理所有权限组，包括添加、删除、编辑等操作</p>
      </div>
      <Button @click="openCreate"><Plus class="mr-2 h-4 w-4" />创建权限组</Button>
    </div>
    <Card>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-16">组ID</TableHead>
              <TableHead>组名称</TableHead>
              <TableHead class="w-32">用户数量</TableHead>
              <TableHead class="w-32">节点数量</TableHead>
              <TableHead class="w-16 text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="g in groups" :key="g.id">
              <TableCell><Badge variant="outline">{{ g.id }}</Badge></TableCell>
              <TableCell class="font-medium">{{ g.name }}</TableCell>
              <TableCell>
                <div class="flex items-center gap-2 text-sm">
                  <Users class="h-4 w-4 text-muted-foreground" />
                  <span>{{ g.user_count ?? 0 }}</span>
                </div>
              </TableCell>
              <TableCell>
                <div class="flex items-center gap-2 text-sm">
                  <Server class="h-4 w-4 text-muted-foreground" />
                  <span>{{ g.node_count ?? 0 }}</span>
                </div>
              </TableCell>
              <TableCell class="text-right" @click.stop>
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="sm" class="h-8 w-8 p-0">
                      <MoreHorizontal class="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem @click="openEdit(g)"><Pencil class="mr-2 h-4 w-4" />编辑</DropdownMenuItem>
                    <DropdownMenuItem class="text-red-500" @click="confirmDelete(g)"><Trash2 class="mr-2 h-4 w-4" />删除</DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
            <TableRow v-if="!groups.length">
              <TableCell colspan="5" class="text-center py-12 text-muted-foreground">
                暂无权限组数据
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <Dialog v-model:open="dialogOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ isEdit ? '编辑权限组' : '创建权限组' }}</DialogTitle>
        </DialogHeader>
        <div class="grid gap-4 py-2">
          <div class="grid gap-2">
            <Label>组名称</Label>
            <Input v-model="editing.name" placeholder="例如：学习、自用" />
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
        <DialogHeader>
          <DialogTitle>确认删除</DialogTitle>
          <DialogDescription>
            确定要删除权限组「{{ editing.name }}」吗？此操作不可撤销。
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="deleteDialogOpen = false">取消</Button>
          <Button variant="destructive" @click="handleDelete">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
