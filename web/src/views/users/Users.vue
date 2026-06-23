<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Plus, Pencil, Trash2 } from 'lucide-vue-next'
import { listUsers, createUser, updateUser, deleteUser } from '@/api/user'
import type { User } from '@/types'

const users = ref<User[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const search = ref('')
const loading = ref(false)

const dialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const editingUser = ref<Partial<User> & { password?: string }>({})
const isEdit = ref(false)

function formatBytes(b: number) {
  if (b === 0) return '0 B'
  const k = 1024, s = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(k))
  return (b / Math.pow(k, i)).toFixed(2) + ' ' + s[i]
}

async function fetchData() {
  loading.value = true
  try {
    const res = await listUsers({ page: page.value, page_size: pageSize.value, q: search.value })
    if (res.code === 0) { users.value = res.data.items; total.value = res.data.total }
  } finally { loading.value = false }
}

function openCreate() {
  editingUser.value = { email: '', password: '', traffic_limit: 0, speed_limit_up: 0, speed_limit_down: 0, device_limit: 0, status: 1 }
  isEdit.value = false
  dialogOpen.value = true
}

function openEdit(u: User) {
  editingUser.value = { ...u, password: '' }
  isEdit.value = true
  dialogOpen.value = true
}

async function handleSave() {
  if (isEdit.value) {
    const data: any = { ...editingUser.value }
    if (!data.password) delete data.password
    await updateUser(editingUser.value.id!, data)
  } else {
    await createUser(editingUser.value as any)
  }
  dialogOpen.value = false
  fetchData()
}

function confirmDelete(u: User) {
  editingUser.value = u
  deleteDialogOpen.value = true
}

async function handleDelete() {
  await deleteUser(editingUser.value.id!)
  deleteDialogOpen.value = false
  fetchData()
}

onMounted(fetchData)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">用户管理</h1>
      <Button @click="openCreate"><Plus class="mr-2 h-4 w-4" />创建用户</Button>
    </div>
    <Card>
      <CardHeader>
        <div class="flex items-center gap-2">
          <Input v-model="search" placeholder="搜索邮箱..." class="w-64" @keyup.enter="fetchData" />
          <Button variant="outline" @click="fetchData">搜索</Button>
        </div>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>邮箱</TableHead>
              <TableHead>UUID</TableHead>
              <TableHead>流量</TableHead>
              <TableHead>限速</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>到期时间</TableHead>
              <TableHead>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="u in users" :key="u.id">
              <TableCell>{{ u.id }}</TableCell>
              <TableCell>{{ u.email }}</TableCell>
              <TableCell class="font-mono text-xs">{{ u.uuid.slice(0, 8) }}...</TableCell>
              <TableCell>{{ formatBytes(u.traffic_used) }} / {{ u.traffic_limit ? formatBytes(u.traffic_limit) : '不限' }}</TableCell>
              <TableCell>{{ u.speed_limit_up || 0 }}/{{ u.speed_limit_down || 0 }} Mbps</TableCell>
              <TableCell><Badge :variant="u.status === 1 ? 'default' : 'destructive'">{{ u.status === 1 ? '启用' : '禁用' }}</Badge></TableCell>
              <TableCell>{{ u.expired_at || '永久' }}</TableCell>
              <TableCell>
                <div class="flex gap-1">
                  <Button variant="ghost" size="sm" @click="openEdit(u)"><Pencil class="h-4 w-4" /></Button>
                  <Button variant="ghost" size="sm" @click="confirmDelete(u)"><Trash2 class="h-4 w-4 text-red-500" /></Button>
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
            <Button variant="outline" size="sm" :disabled="page * pageSize >= total" @click="page++; fetchData()">下一页</Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <Dialog v-model:open="dialogOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ isEdit ? '编辑用户' : '创建用户' }}</DialogTitle>
          <DialogDescription>{{ isEdit ? '修改用户信息' : '填写用户信息创建新账号' }}</DialogDescription>
        </DialogHeader>
        <div class="grid gap-4 py-4">
          <div class="grid gap-2"><Label>邮箱</Label><Input v-model="editingUser.email" /></div>
          <div class="grid gap-2" v-if="!isEdit"><Label>密码</Label><Input v-model="editingUser.password" type="password" /></div>
          <div class="grid gap-2"><Label>流量上限 (字节, 0=不限)</Label><Input v-model.number="editingUser.traffic_limit" type="number" /></div>
          <div class="grid grid-cols-2 gap-2">
            <div class="grid gap-2"><Label>上行限速 (Mbps)</Label><Input v-model.number="editingUser.speed_limit_up" type="number" /></div>
            <div class="grid gap-2"><Label>下行限速 (Mbps)</Label><Input v-model.number="editingUser.speed_limit_down" type="number" /></div>
          </div>
          <div class="grid gap-2"><Label>设备限制 (0=不限)</Label><Input v-model.number="editingUser.device_limit" type="number" /></div>
        </div>
        <DialogFooter><Button @click="handleSave">保存</Button></DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="deleteDialogOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>确认删除</DialogTitle>
          <DialogDescription>确定要删除用户 {{ editingUser.email }} 吗？此操作不可撤销。</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="deleteDialogOpen = false">取消</Button>
          <Button variant="destructive" @click="handleDelete">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>