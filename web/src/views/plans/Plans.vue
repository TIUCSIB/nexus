<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { listPlans, createPlan, updatePlan, deletePlan } from '@/api/plan'
import type { Plan } from '@/types'
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
import { Plus, MoreHorizontal, Pencil, Trash2 } from '@lucide/vue'

const plans = ref<Plan[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const showDialog = ref(false)
const showDeleteDialog = ref(false)
const editingPlan = ref<Plan | null>(null)
const deletingPlan = ref<Plan | null>(null)
const saving = ref(false)

const form = ref({
  name: '',
  description: '',
  traffic_limit: 0,
  duration_days: 30,
  price: 0,
  sort: 0,
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
  if (bytes === 0) return '不限'
  return bytes + ' B'
}

function formatPrice(price: number): string {
  return '￥' + (price / 100).toFixed(2)
}

async function fetchPlans() {
  loading.value = true
  try {
    const res = await listPlans({ page: page.value, page_size: pageSize.value })
    if (res.code === 0 && res.data) {
      plans.value = res.data.items
      total.value = res.data.total
    }
  } catch (err) {
    console.error('获取套餐列表失败:', err)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingPlan.value = null
  form.value = {
    name: '',
    description: '',
    traffic_limit: 0,
    duration_days: 30,
    price: 0,
    sort: 0,
    status: 1,
  }
  showDialog.value = true
}

function openEdit(plan: Plan) {
  editingPlan.value = plan
  form.value = {
    name: plan.name,
    description: plan.description,
    traffic_limit: plan.traffic_limit,
    duration_days: plan.duration_days,
    price: plan.price,
    sort: plan.sort,
    status: plan.status,
  }
  showDialog.value = true
}

async function handleSave() {
  saving.value = true
  try {
    if (editingPlan.value) {
      await updatePlan(editingPlan.value.id, { ...form.value })
    } else {
      await createPlan({ ...form.value })
    }
    showDialog.value = false
    await fetchPlans()
  } catch (err) {
    console.error('保存套餐失败:', err)
  } finally {
    saving.value = false
  }
}

function confirmDelete(plan: Plan) {
  deletingPlan.value = plan
  showDeleteDialog.value = true
}

async function handleDelete() {
  if (!deletingPlan.value) return
  try {
    await deletePlan(deletingPlan.value.id)
    showDeleteDialog.value = false
    deletingPlan.value = null
    await fetchPlans()
  } catch (err) {
    console.error('删除套餐失败:', err)
  }
}

function goToPage(p: number) {
  if (p >= 1 && p <= totalPages.value) {
    page.value = p
    fetchPlans()
  }
}

onMounted(fetchPlans)
</script>

<template>
  <div class="space-y-4">
    <!-- 顶部操作栏 -->
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold">套餐管理</h2>
      <Button @click="openCreate">
        <Plus class="size-4" />
        创建套餐
      </Button>
    </div>

    <!-- 套餐表格 -->
    <div class="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead class="w-16">ID</TableHead>
            <TableHead>名称</TableHead>
            <TableHead>流量限制</TableHead>
            <TableHead>时长（天）</TableHead>
            <TableHead>价格</TableHead>
            <TableHead>排序</TableHead>
            <TableHead>状态</TableHead>
            <TableHead class="w-16">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-if="loading">
            <TableCell colspan="8" class="h-24 text-center text-muted-foreground">
              加载中...
            </TableCell>
          </TableRow>
          <TableRow v-else-if="plans.length === 0">
            <TableCell colspan="8" class="h-24 text-center text-muted-foreground">
              暂无数据
            </TableCell>
          </TableRow>
          <TableRow v-for="plan in plans" :key="plan.id">
            <TableCell class="font-medium">{{ plan.id }}</TableCell>
            <TableCell>{{ plan.name }}</TableCell>
            <TableCell>{{ formatTraffic(plan.traffic_limit) }}</TableCell>
            <TableCell>{{ plan.duration_days }}</TableCell>
            <TableCell>{{ formatPrice(plan.price) }}</TableCell>
            <TableCell>{{ plan.sort }}</TableCell>
            <TableCell>
              <Badge :variant="plan.status === 1 ? 'default' : 'destructive'">
                {{ plan.status === 1 ? '上架' : '下架' }}
              </Badge>
            </TableCell>
            <TableCell>
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon-sm">
                    <MoreHorizontal class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuItem @click="openEdit(plan)">
                    <Pencil class="size-4" />
                    编辑
                  </DropdownMenuItem>
                  <DropdownMenuItem @click="confirmDelete(plan)" class="text-destructive">
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
    <Pagination v-if="totalPages > 1" :total="total" :items-per-page="pageSize" :page="page" @update:page="goToPage">
      <PaginationContent>
        <PaginationItem>
          <PaginationPrevious @click="goToPage(page - 1)" />
        </PaginationItem>
        <PaginationItem v-for="p in totalPages" :key="p">
          <PaginationLink :is-active="p === page" @click="goToPage(p)">
            {{ p }}
          </PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationNext @click="goToPage(page + 1)" />
        </PaginationItem>
      </PaginationContent>
    </Pagination>

    <!-- 创建/编辑对话框 -->
    <Dialog v-model:open="showDialog">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{{ editingPlan ? '编辑套餐' : '创建套餐' }}</DialogTitle>
          <DialogDescription>
            {{ editingPlan ? '修改套餐信息' : '填写新套餐信息' }}
          </DialogDescription>
        </DialogHeader>
        <form class="grid gap-4 py-4" @submit.prevent="handleSave">
          <div class="grid gap-2">
            <Label for="plan-name">名称</Label>
            <Input id="plan-name" v-model="form.name" required />
          </div>
          <div class="grid gap-2">
            <Label for="plan-desc">描述</Label>
            <Input id="plan-desc" v-model="form.description" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label for="plan-traffic">流量限制 (字节)</Label>
              <Input id="plan-traffic" v-model.number="form.traffic_limit" type="number" />
            </div>
            <div class="grid gap-2">
              <Label for="plan-duration">时长（天）</Label>
              <Input id="plan-duration" v-model.number="form.duration_days" type="number" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label for="plan-price">价格（分）</Label>
              <Input id="plan-price" v-model.number="form.price" type="number" />
            </div>
            <div class="grid gap-2">
              <Label for="plan-sort">排序</Label>
              <Input id="plan-sort" v-model.number="form.sort" type="number" />
            </div>
          </div>
          <div class="grid gap-2">
            <Label for="plan-status">状态</Label>
            <select
              id="plan-status"
              v-model.number="form.status"
              class="border-input bg-background h-8 rounded-md border px-3 text-sm"
            >
              <option :value="1">上架</option>
              <option :value="0">下架</option>
            </select>
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
            确定要删除套餐 <strong>{{ deletingPlan?.name }}</strong> 吗？此操作不可撤销。
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
