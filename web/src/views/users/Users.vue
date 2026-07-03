<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter,
  DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
// pagination uses custom buttons
import { Plus, Pencil, Trash2, Eye, Search, RotateCcw, CalendarIcon, MoreHorizontal, Copy, RefreshCw, Activity, BarChart3 } from 'lucide-vue-next'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Calendar } from '@/components/ui/calendar'
import { getLocalTimeZone, today, parseDate, type DateValue } from '@internationalized/date'
import { toast } from 'vue-sonner'
import { useSettingsStore } from '@/stores/settings'
import { listUsers, createUser, updateUser, deleteUser, resetUserUUID, resetUserTraffic, getUserTrafficLogs } from '@/api/user'
import { listPlans } from '@/api/plan'
import { listGroups } from '@/api/group'
import type { User, Plan, PageResult } from '@/types'
import type { ServerGroup } from '@/api/group'

const router = useRouter()
const settingsStore = useSettingsStore()
const users = ref<User[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const search = ref('')
const loading = ref(false)
const totalPages = ref(1)

const plans = ref<Plan[]>([])
const groups = ref<ServerGroup[]>([])
const loadedOptions = ref(false)

const dialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const editingUser = ref<Partial<User> & { password?: string }>({})
const selectedPlanId = ref<string>('__none__')
const selectedGroupId = ref<string>('__none__')
const isAdmin = ref(false)
const emailPrefix = ref('')
const emailDomain = ref('')
const expiredAt = ref('')
const datePickerOpen = ref(false)

function formatDateValue(d: DateValue | undefined): string {
  if (!d) return ''
  const y = d.year
  const m = String(d.month).padStart(2, '0')
  const day = String(d.day).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function toGB(bytes: number | null | undefined): number | null {
  if (!bytes || bytes === 0) return null
  return Math.round((bytes / (1024 * 1024 * 1024)) * 100) / 100
}

function toBytes(gb: number | null | undefined): number {
  if (!gb || gb === 0) return 0
  return Math.round(gb * 1024 * 1024 * 1024)
}

const datePickerModel = computed({
  get(): DateValue | undefined {
    if (!expiredAt.value) return undefined
    try { return parseDate(expiredAt.value) } catch { return undefined }
  },
  set(val: DateValue | undefined) {
    expiredAt.value = val ? formatDateValue(val) : ''
    datePickerOpen.value = false
  }
})
const isEdit = ref(false)
const saving = ref(false)

function formatBytes(b: number) {
  if (b === 0) return '0 B'
  const k = 1024; const s = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(k))
  return (b / Math.pow(k, i)).toFixed(2) + ' ' + s[i]
}

function formatExpiry(d: string | null) {
  if (!d) return '永久'
  const date = new Date(d)
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function resolveGroupName(u: User): string {
  if (u.group_id) {
    const g = groups.value.find(g => g.id === u.group_id)
    return g ? g.name : ''
  }
  if (u.plan_id) {
    const plan = plans.value.find(p => p.id === u.plan_id)
    if (plan && plan.group_id) {
      const g = groups.value.find(g => g.id === plan.group_id)
      return g ? g.name : ''
    }
  }
  return ''
}

function isExpired(d: string | null): boolean {
  if (!d) return false
  return new Date(d) < new Date()
}

async function loadOptions() {
  if (loadedOptions.value) return
  try {
    const [planRes, groupRes] = await Promise.all([
      listPlans({ page: 1, page_size: 100 }),
      listGroups(),
    ])
    if (planRes.code === 0) plans.value = planRes.data.items || []
    if (groupRes.code === 0) groups.value = groupRes.data || []
    loadedOptions.value = true
  } catch { /* ignore */ }
}

async function fetchData() {
  loading.value = true
  try {
    const res = await listUsers({ page: page.value, page_size: pageSize.value, q: search.value })
    if (res.code === 0) {
      users.value = res.data.items
      total.value = res.data.total
      totalPages.value = Math.max(1, Math.ceil(total.value / pageSize.value))
    }
  } catch { toast.error('获取用户列表失败') }
  finally { loading.value = false }
}

function openCreate() {
  editingUser.value = {
    email: '', password: '', traffic_limit: null as any,
    speed_limit_up: null as any, speed_limit_down: null as any, device_limit: null as any,
    status: 1, balance: 0, expired_at: null,
  }
  emailPrefix.value = ''
  emailDomain.value = ''
  expiredAt.value = ''
  selectedPlanId.value = "__none__"
  selectedGroupId.value = "__none__"
  isAdmin.value = false
  isEdit.value = false
  loadOptions()
  dialogOpen.value = true
}

function openEdit(u: User) {
  editingUser.value = {
    ...u, password: '',
    upload_used: toGB(u.upload_used),
    download_used: toGB(u.download_used),
    traffic_limit: toGB(u.traffic_limit),
    speed_limit_up: u.speed_limit_up || u.speed_limit_down || null,
    speed_limit_down: null,
    device_limit: u.device_limit || null,
    remarks: u.remarks || '',
  }
  if (u.email && u.email.includes('@')) {
    const parts = u.email.split('@')
    emailPrefix.value = parts[0]
    emailDomain.value = parts[1]
  }
  expiredAt.value = u.expired_at ? u.expired_at.substring(0, 10) : ''
  selectedPlanId.value = u.plan_id ? String(u.plan_id) : "__none__"
  selectedGroupId.value = u.group_id ? String(u.group_id) : "__none__"
  isAdmin.value = !!u.is_admin
  isEdit.value = true
  loadOptions()
  dialogOpen.value = true
}

async function handleSave() {
  const composedEmail = (emailPrefix.value + '@' + emailDomain.value).trim()
  if (!composedEmail || !emailPrefix.value || !emailDomain.value) {
    toast.error('请输入完整的邮箱地址'); return
  }
  if (!isEdit.value && !editingUser.value.password) {
    editingUser.value.password = emailPrefix.value
  }

  saving.value = true
  try {
    const data: any = { ...editingUser.value }
    data.email = composedEmail
    if (!data.password) delete data.password
    if (data.password === '') delete data.password
    data.plan_id = selectedPlanId.value && selectedPlanId.value !== '__none__' ? Number(selectedPlanId.value) : null
    data.group_id = selectedGroupId.value && selectedGroupId.value !== '__none__' ? Number(selectedGroupId.value) : null
    data.expired_at = expiredAt.value || null
    data.is_admin = isAdmin.value
    data.upload_used = toBytes(data.upload_used)
    data.download_used = toBytes(data.download_used)
    data.traffic_limit = toBytes(data.traffic_limit)
    data.speed_limit_up = data.speed_limit_up ? Number(data.speed_limit_up) : 0
    data.speed_limit_down = data.speed_limit_up
    data.device_limit = data.device_limit ? Number(data.device_limit) : 0

    if (isEdit.value) {
      const res = await updateUser(editingUser.value.id!, data)
      if (res.code === 0) { toast.success('用户已更新'); dialogOpen.value = false; fetchData() }
      else { toast.error(res.message || '更新失败') }
    } else {
      const res = await createUser(data)
      if (res.code === 0) { toast.success('用户已创建'); dialogOpen.value = false; fetchData() }
      else { toast.error(res.message || '创建失败') }
    }
  } catch (e: any) { toast.error(e?.response?.data?.message || '操作失败，请重试') }
  finally { saving.value = false }
}


async function handleCopySubUrl(u: User) {
  // 从订阅URL列表中随机选一个
  const subUrlSetting = settingsStore.subUrl
  let baseUrl = window.location.origin
  if (subUrlSetting) {
    const urls = subUrlSetting.split(',').map((s: string) => s.trim().replace(/\/$/, '')).filter(Boolean)
    if (urls.length > 0) {
      baseUrl = urls[Math.floor(Math.random() * urls.length)]
    }
  }
  const url = baseUrl + '/' + settingsStore.subPath + '/' + u.token
  try {
    await navigator.clipboard.writeText(url)
    toast.success('订阅URL已复制')
  } catch {
    const textarea = document.createElement('textarea')
    textarea.value = url
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    toast.success('订阅URL已复制')
  }
}

async function handleResetUUID(u: User) {
  try {
    const res = await resetUserUUID(u.id)
    if (res.code === 0) {
      toast.success('UUID已重置')
      fetchData()
    } else {
      toast.error(res.message || '重置失败')
    }
  } catch (e: any) { toast.error(e?.response?.data?.message || '重置失败') }
}

async function handleResetTraffic(u: User) {
  try {
    const res = await resetUserTraffic(u.id)
    if (res.code === 0) {
      toast.success('流量已重置')
      fetchData()
    } else {
      toast.error(res.message || '重置失败')
    }
  } catch (e: any) { toast.error(e?.response?.data?.message || '重置失败') }
}

function confirmDelete(u: User) { editingUser.value = u; deleteDialogOpen.value = true }

async function handleDelete() {
  try {
    const res = await deleteUser(editingUser.value.id!)
    if (res.code === 0) { toast.success('用户已删除'); deleteDialogOpen.value = false; fetchData() }
    else { toast.error(res.message || '删除失败') }
  } catch (e: any) { toast.error(e?.response?.data?.message || '删除失败，请重试') }
}

function goDetail(id: number) { router.push(`/admin/users/${id}`) }

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

onMounted(() => { fetchData(); loadOptions() })
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">用户管理</h1>
      <Button @click="openCreate"><Plus class="mr-2 h-4 w-4" />创建用户</Button>
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
              <TableHead class="w-12">ID</TableHead>
              <TableHead>邮箱</TableHead>
              <TableHead>权限组</TableHead>
              <TableHead>已用流量</TableHead>
              <TableHead>总流量</TableHead>
              <TableHead>设备</TableHead>
              <TableHead>到期时间</TableHead>
              <TableHead>状态</TableHead>
              <TableHead class="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="u in users" :key="u.id" class="cursor-pointer hover:bg-muted/50" @click="goDetail(u.id)">
              <TableCell class="font-mono text-xs">{{ u.id }}</TableCell>
              <TableCell class="font-medium max-w-[180px]">
                <div class="flex items-center gap-2">
                  <span :class="u.online ? 'bg-green-500' : 'bg-gray-400'" class="w-2 h-2 rounded-full shrink-0" />
                  <span class="truncate">{{ u.email }}</span>
                </div>
              </TableCell>
              <TableCell>
                <span class="text-sm">{{ resolveGroupName(u) || '—' }}</span>
              </TableCell>
              <TableCell class="min-w-[140px]">
                <div class="flex items-center gap-1.5">
                  <span class="font-mono text-sm">{{ formatBytes(u.traffic_used) }}</span>
                  <span class="text-muted-foreground text-xs" v-if="u.traffic_limit">{{ ((u.traffic_used / u.traffic_limit) * 100).toFixed(1) }}%</span>
                </div>
                <div class="h-1.5 bg-muted rounded-full overflow-hidden mt-1">
                  <div class="h-full rounded-full transition-all" :class="(u.traffic_limit && (u.traffic_used / u.traffic_limit) > 0.8) ? 'bg-red-500' : 'bg-primary'" :style="{ width: u.traffic_limit ? Math.min(100, (u.traffic_used / u.traffic_limit) * 100) + '%' : '0%' }"></div>
                </div>
              </TableCell>
              <TableCell class="min-w-[100px]">
                <span class="font-mono text-sm">{{ u.traffic_limit ? formatBytes(u.traffic_limit) : '不限' }}</span>
              </TableCell>
              <TableCell class="text-sm">{{ u.device_limit || '不限' }}</TableCell>
              <TableCell class="text-xs whitespace-nowrap">
                <span :class="isExpired(u.expired_at) ? 'inline-flex items-center rounded-md border border-red-200 bg-red-50 px-2 py-0.5 text-red-600 font-medium dark:bg-red-950 dark:text-red-400 dark:border-red-800' : ''">
                  {{ formatExpiry(u.expired_at) }}
                </span>
              </TableCell>
              <TableCell>
                <Badge :variant="u.status === 1 ? 'default' : 'destructive'" class="text-xs">
                  {{ u.status === 1 ? '启用' : '禁用' }}
                </Badge>
              </TableCell>
              <TableCell class="text-right" @click.stop>
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="sm" class="h-8 w-8 p-0">
                      <MoreHorizontal class="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" class="w-48">
                    <DropdownMenuItem @click="goDetail(u.id)"><Eye class="mr-2 h-4 w-4" />详情</DropdownMenuItem>
                    <DropdownMenuItem @click="openEdit(u)"><Pencil class="mr-2 h-4 w-4" />编辑</DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem @click="handleCopySubUrl(u)"><Copy class="mr-2 h-4 w-4" />复制订阅URL</DropdownMenuItem>
                    <DropdownMenuItem @click="handleResetUUID(u)"><RefreshCw class="mr-2 h-4 w-4" />重置UUID</DropdownMenuItem>
                    <DropdownMenuItem @click="handleResetTraffic(u)"><BarChart3 class="mr-2 h-4 w-4" />重置流量</DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem class="text-red-500" @click="confirmDelete(u)"><Trash2 class="mr-2 h-4 w-4" />删除</DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
            <TableRow v-if="!users.length && !loading">
              <TableCell colspan="11" class="text-center py-12 text-muted-foreground">
                暂无用户数据
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>

        <div class="flex items-center justify-between px-4 py-3 border-t">
          <span class="text-sm text-muted-foreground">共 {{ total }} 条，第 {{ page }} / {{ totalPages }} 页</span>
          <Pagination v-if="totalPages > 1">
            <PaginationContent>
              <PaginationItem>
                <PaginationPrevious @click="page = Math.max(1, page - 1); fetchData()" :disabled="page <= 1" />
              </PaginationItem>
              <PaginationItem v-for="p in visiblePages" :key="p">
                <PaginationLink :isActive="p === page" @click="page = p; fetchData()">{{ p }}</PaginationLink>
              </PaginationItem>
              <PaginationItem>
                <PaginationNext @click="page = Math.min(totalPages, page + 1); fetchData()" :disabled="page >= totalPages" />
              </PaginationItem>
            </PaginationContent>
          </Pagination>
        </div>
      </CardContent>
    </Card>

    <!-- Create/Edit Dialog -->
    <Dialog v-model:open="dialogOpen">
      <DialogContent class="max-w-2xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{{ isEdit ? '编辑用户' : '创建用户' }}</DialogTitle>
        </DialogHeader>
        <div class="grid gap-4 py-2">
          <div class="grid gap-2">
            <Label>邮箱 <span class="text-red-500">*</span></Label>
            <div class="flex items-center gap-0">
              <Input v-model="emailPrefix" placeholder="帐号（批量生成请留空）" class="rounded-r-none" />
              <span class="flex items-center justify-center px-3 h-9 border border-l-0 border-input bg-muted text-muted-foreground text-sm">@</span>
              <Input v-model="emailDomain" placeholder="域" class="rounded-l-none" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label>密码</Label>
              <Input v-model="editingUser.password" type="password" :placeholder="isEdit ? '留空则不修改' : '留空则密码与邮件相同'" />
            </div>
            <div class="grid gap-2">
              <Label>账号状态</Label>
              <Select v-model="editingUser.status">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem :value="1">启用</SelectItem>
                  <SelectItem :value="0">禁用</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label>到期时间</Label>
              <Popover v-model:open="datePickerOpen">
                <PopoverTrigger as-child>
                  <Button variant="outline" class="justify-start text-left font-normal h-9">
                    <CalendarIcon class="mr-2 h-4 w-4 shrink-0" />
                    <span :class="!expiredAt && 'text-muted-foreground'">{{ expiredAt || '留空为长期有效' }}</span>
                  </Button>
                </PopoverTrigger>
                <PopoverContent class="w-auto p-0" align="start">
                  <Calendar v-model="datePickerModel" locale="zh-CN" />
                </PopoverContent>
              </Popover>
            </div>
            <div class="grid gap-2">
              <Label>订阅计划</Label>
              <Select v-model="selectedPlanId">
                <SelectTrigger><SelectValue placeholder="无" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="__none__">无</SelectItem>
                  <SelectItem v-for="p in plans" :key="p.id" :value="String(p.id)">{{ p.name }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div class="grid grid-cols-3 gap-4">
            <div class="grid gap-2">
              <Label>已用上行</Label>
              <div class="relative">
                <Input v-model.number="editingUser.upload_used" type="number" min="0" class="pr-12" placeholder="0" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">GB</span>
              </div>
            </div>
            <div class="grid gap-2">
              <Label>已用下行</Label>
              <div class="relative">
                <Input v-model.number="editingUser.download_used" type="number" min="0" class="pr-12" placeholder="0" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">GB</span>
              </div>
            </div>
            <div class="grid gap-2">
              <Label>流量</Label>
              <div class="relative">
                <Input v-model.number="editingUser.traffic_limit" type="number" min="0" class="pr-12" placeholder="0=不限" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">GB</span>
              </div>
            </div>
          </div>
          <div class="grid gap-2">
            <Label>限速</Label>
            <div class="relative">
              <Input v-model.number="editingUser.speed_limit_up" type="number" min="0" class="pr-16" placeholder="留空则不限速" />
              <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">Mbps</span>
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label>设备限制</Label>
              <div class="relative">
                <Input v-model.number="editingUser.device_limit" type="number" min="0" class="pr-10" placeholder="留空则不限制" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">台</span>
              </div>
            </div>
            <div class="grid gap-2">
              <Label>权限组</Label>
              <Select v-model="selectedGroupId">
                <SelectTrigger><SelectValue placeholder="无" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="__none__">无</SelectItem>
                  <SelectItem v-for="g in groups" :key="g.id" :value="String(g.id)">{{ g.name }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div class="flex items-center justify-between">
            <div>
              <Label>是否管理员</Label>
              <p class="text-sm text-muted-foreground">管理员拥有后台全部权限</p>
            </div>
            <Switch v-model="isAdmin" />
          </div>
          <div class="grid gap-2">
            <Label>备注</Label>
            <Textarea v-model="editingUser.remarks" rows="2" placeholder="管理员备注信息" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="dialogOpen = false">取消</Button>
          <Button @click="handleSave" :disabled="saving">{{ saving ? '保存中...' : '保存' }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Delete Confirmation -->
    <Dialog v-model:open="deleteDialogOpen">
      <DialogContent class="max-w-sm">
        <DialogHeader>
          <DialogTitle>确认删除</DialogTitle>
          <DialogDescription>
            确定要删除用户 <strong>{{ editingUser.email }}</strong> 吗？此操作不可撤销，该用户的所有数据将被永久删除。
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="deleteDialogOpen = false">取消</Button>
          <Button variant="destructive" @click="handleDelete">确认删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
