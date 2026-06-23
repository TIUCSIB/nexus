<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { listUsers, createUser, updateUser, deleteUser } from '@/api/user'
import type { User } from '@/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'
import { Plus, Search, MoreHorizontal, Pencil, Trash2 } from '@lucide/vue'

const users = ref<User[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const searchQuery = ref('')
const loading = ref(false)

const showDialog = ref(false)
const showDeleteDialog = ref(false)
const editingUser = ref<User | null>(null)
const deletingUser = ref<User | null>(null)
const saving = ref(false)

const form = ref({
  email: '',
  password: '',
  traffic_limit: 0,
  speed_limit_up: 0,
  speed_limit_down: 0,
  device_limit: 0,
  expired_at: '',
  status: 1,
})

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

function formatTraffic(bytes: number): string {
  if (bytes >= 1073741824) {
    return (bytes / 1073741824).toFixed(2) + ' GB'
  }
  if (bytes >= 1048576) {
    return (bytes / 1048576).toFixed(2) + ' MB'
  }
  if (bytes === 0) return '0 B'
  return bytes + ' B'
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return '永久'
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

async function fetchUsers() {
  loading.value = true
  try {
    const res = await listUsers({ page: page.value, page_size: pageSize.value, q: searchQuery.value || undefined })
    if (res.code === 0 && res.data) {
      users.value = res.data.items
      total.value = res.data.total
    }
  } catch (err) {
    console.error('获取用户列表失败:', err)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingUser.value = null
  form.value = {
    email: '',
    password: '',
    traffic_limit: 0,
    speed_limit_up: 0,
    speed_limit_down: 0,
    device_limit: 0,
    expired_at: '',
    status: 1,
  }
  showDialog.value = true
}

function openEdit(user: User) {
  editingUser.value = user
  form.value = {
    email: user.email,
    password: '',
    traffic_limit: user.traffic_limit,
    speed_limit_up: user.speed_limit_up,
    speed_limit_down: user.speed_limit_down,
    device_limit: user.device_limit,
    expired_at: user.expired_at ? user.expired_at.slice(0, 10) : '',
    status: user.status,
  }
  showDialog.value = true
}

async function handleSave() {
  saving.value = true
  try {
    if (editingUser.value) {
      const data: Partial<User> = {
        email: form.value.email,
        traffic_limit: form.value.traffic_limit,
        speed_limit_up: form.value.speed_limit_up,
        speed_limit_down: form.value.speed_limit_down,
        device_limit: form.value.device_limit,
        expired_at: form.value.expired_at || null,
        status: form.value.status,
      }
      if (form.value.password) {
        (data as any).password = form.value.password
      }
      await updateUser(editingUser.value.id, data)
    } else {
      await createUser({
        email: form.value.email,
        password: form.value.password,
        traffic_limit: form.value.traffic_limit,
        speed_limit_up: form.value.speed_limit_up,
        speed_limit_down: form.value.speed_limit_down,
        device_limit: form.value.device_limit,
        expired_at: form.value.expired_at || null,
        status: form.value.status,
      })
    }
    showDialog.value = false
    await fetchUsers()
  } catch (err) {
    console.error('保存用户失败:', err)
  } finally {
    saving.value = false
  }
}

function confirmDelete(user: User) {
  deletingUser.value = user
  showDeleteDialog.value = true
}

async function handleDelete() {
  if (!deletingUser.value) return
  try {
    await deleteUser(deletingUser.value.id)
    showDeleteDialog.value = false
    deletingUser.value = null
    await fetchUsers()
  } catch (err) {
    console.error('删除用户失败:', err)
  }
}

function goToPage(p: number) {
  if (p >= 1 && p <= totalPages.value) {
    page.value = p
    fetchUsers()
  }
}

function handleSearch() {
  page.value = 1
  fetchUsers()
}

onMounted(fetchUsers)
</script>

<template>
  <div class="space-y-4">
    <!-- 顶部操作栏 -->
    <div class="flex items-center gap-2">
      <div class="relative flex-1 max-w-sm">
        <Search class="absolute left-2.5 top-2.5 size-4 text-muted-foreground" />
        <Input
          v-model="searchQuery"
          placeholder="搜索用户..."
          class="pl-8"
          @keyup.enter="handleSearch"
        />
      </div>
      <Button @click="openCreate">
        <Plus class="size-4" />
        创建用户
      </Button>
    </div>

    <!-- 用户表格 -->
    <div class="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead class="w-16">ID</TableHead>
            <TableHead>邮箱</TableHead>
            <TableHead>UUID</TableHead>
            <TableHead>流量</TableHead>
            <TableHead>限速</TableHead>
            <TableHead>状态</TableHead>
            <TableHead>到期时间</TableHead>
            <TableHead class="w-16">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-if="loading">
            <TableCell colspan="8" class="h-24 text-center text-muted-foreground">
              加载中...
            </TableCell>
          </TableRow>
          <TableRow v-else-if="users.length === 0">
            <TableCell colspan="8" class="h-24 text-center text-muted-foreground">
              暂无数据
            </TableCell>
          </TableRow>
          <TableRow v-for="user in users" :key="user.id">
            <TableCell class="font-medium">{{ user.id }}</TableCell>
            <TableCell>{{ user.email }}</TableCell>
            <TableCell class="font-mono text-xs">
              {{ user.uuid.slice(0, 8) }}...
            </TableCell>
            <TableCell>
              {{ formatTraffic(user.traffic_used) }} / {{ formatTraffic(user.traffic_limit) }}
            </TableCell>
            <TableCell>
              ↑{{ user.speed_limit_up || '∞' }} ↓{{ user.speed_limit_down || '∞' }}
            </TableCell>
            <TableCell>
              <Badge :variant="user.status === 1 ? 'default' : 'destructive'">
                {{ user.status === 1 ? '正常' : '禁用' }}
              </Badge>
            </TableCell>
            <TableCell>{{ formatDate(user.expired_at) }}</TableCell>
            <TableCell>
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon-sm">
                    <MoreHorizontal class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuItem @click="openEdit(user)">
                    <Pencil class="size-4" />
                    编辑
                  </DropdownMenuItem>
                  <DropdownMenuItem @click="confirmDelete(user)" class="text-destructive">
                    <Trash2 class="size-4" />
                    删除
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <!-- 分页 -->
    <div v-if="totalPages > 1" class="flex justify-center">
      <nav class="flex items-center gap-1">
        <Button variant="ghost" size="sm" :disabled="page <= 1" @click="goToPage(page - 1)">
          上一页
        </Button>
        <Button
          v-for="p in totalPages"
          :key="p"
          :variant="p === page ? 'outline' : 'ghost'"
          size="sm"
          @click="goToPage(p)"
        >
          {{ p }}
        </Button>
        <Button variant="ghost" size="sm" :disabled="page >= totalPages" @click="goToPage(page + 1)">
          下一页
        </Button>
      </nav>
    </div>

    <!-- 创建/编辑对话框 -->
    <Dialog v-model:open="showDialog">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{{ editingUser ? '编辑用户' : '创建用户' }}</DialogTitle>
          <DialogDescription>
            {{ editingUser ? '修改用户信息' : '填写新用户信息' }}
          </DialogDescription>
        </DialogHeader>
        <form class="grid gap-4 py-4" @submit.prevent="handleSave">
          <div class="grid gap-2">
            <Label for="form-email">邮箱</Label>
            <Input id="form-email" v-model="form.email" type="email" required />
          </div>
          <div class="grid gap-2">
            <Label for="form-password">{{ editingUser ? '新密码（留空不修改）' : '密码' }}</Label>
            <Input id="form-password" v-model="form.password" type="password" :required="!editingUser" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label for="form-traffic">流量限制 (字节)</Label>
              <Input id="form-traffic" v-model.number="form.traffic_limit" type="number" />
            </div>
            <div class="grid gap-2">
              <Label for="form-device">设备限制</Label>
              <Input id="form-device" v-model.number="form.device_limit" type="number" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label for="form-up">上行限速</Label>
              <Input id="form-up" v-model.number="form.speed_limit_up" type="number" />
            </div>
            <div class="grid gap-2">
              <Label for="form-down">下行限速</Label>
              <Input id="form-down" v-model.number="form.speed_limit_down" type="number" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label for="form-expired">到期时间</Label>
              <Input id="form-expired" v-model="form.expired_at" type="date" />
            </div>
            <div class="grid gap-2">
              <Label for="form-status">状态</Label>
              <select
                id="form-status"
                v-model.number="form.status"
                class="border-input bg-background h-8 rounded-md border px-3 text-sm"
              >
                <option :value="1">正常</option>
                <option :value="0">禁用</option>
              </select>
            </div>
          </div>
        </form>
        <DialogFooter>
          <Button variant="outline" @click="showDialog = false">取消</Button>
          <Button :disabled="saving" @click="handleSave">
            {{ saving ? '保存中...' : '保存' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 删除确认对话框 -->
    <Dialog v-model:open="showDeleteDialog">
      <DialogContent class="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>确认删除</DialogTitle>
          <DialogDescription>
            确定要删除用户 <strong>{{ deletingUser?.email }}</strong> 吗？此操作不可撤销。
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="showDeleteDialog = false">取消</Button>
          <Button variant="destructive" @click="handleDelete">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
