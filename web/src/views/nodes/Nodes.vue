<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { listNodes, createNode, updateNode, deleteNode, restartNode } from '@/api/node'
import type { Node } from '@/types'
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'
import { Plus, MoreHorizontal, Pencil, Trash2, RotateCw } from '@lucide/vue'

const nodes = ref<Node[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const restartingId = ref<number | null>(null)

const showDialog = ref(false)
const showDeleteDialog = ref(false)
const editingNode = ref<Node | null>(null)
const deletingNode = ref<Node | null>(null)
const saving = ref(false)

const form = ref({
  name: '',
  address: '',
  protocol: 'vless',
  port: 443,
  config_mode: 'auto',
  config_json: '',
  sort: 0,
  status: 1,
})

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

function formatDate(dateStr: string | null): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

async function fetchNodes() {
  loading.value = true
  try {
    const res = await listNodes({ page: page.value, page_size: pageSize.value })
    if (res.code === 0 && res.data) {
      nodes.value = res.data.items
      total.value = res.data.total
    }
  } catch (err) {
    console.error('获取节点列表失败:', err)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingNode.value = null
  form.value = {
    name: '',
    address: '',
    protocol: 'vless',
    port: 443,
    config_mode: 'auto',
    config_json: '',
    sort: 0,
    status: 1,
  }
  showDialog.value = true
}

function openEdit(node: Node) {
  editingNode.value = node
  form.value = {
    name: node.name,
    address: node.address,
    protocol: node.protocol,
    port: node.port,
    config_mode: node.config_mode,
    config_json: node.config_json,
    sort: node.sort,
    status: node.status,
  }
  showDialog.value = true
}

async function handleSave() {
  saving.value = true
  try {
    if (editingNode.value) {
      await updateNode(editingNode.value.id, { ...form.value })
    } else {
      await createNode({ ...form.value })
    }
    showDialog.value = false
    await fetchNodes()
  } catch (err) {
    console.error('保存节点失败:', err)
  } finally {
    saving.value = false
  }
}

function confirmDelete(node: Node) {
  deletingNode.value = node
  showDeleteDialog.value = true
}

async function handleDelete() {
  if (!deletingNode.value) return
  try {
    await deleteNode(deletingNode.value.id)
    showDeleteDialog.value = false
    deletingNode.value = null
    await fetchNodes()
  } catch (err) {
    console.error('删除节点失败:', err)
  }
}

async function handleRestart(id: number) {
  restartingId.value = id
  try {
    await restartNode(id)
  } catch (err) {
    console.error('重启节点失败:', err)
  } finally {
    restartingId.value = null
  }
}

function goToPage(p: number) {
  if (p >= 1 && p <= totalPages.value) {
    page.value = p
    fetchNodes()
  }
}

onMounted(fetchNodes)
</script>

<template>
  <div class="space-y-4">
    <!-- 顶部操作栏 -->
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold">节点管理</h2>
      <Button @click="openCreate">
        <Plus class="size-4" />
        创建节点
      </Button>
    </div>

    <!-- 节点表格 -->
    <div class="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead class="w-16">ID</TableHead>
            <TableHead>名称</TableHead>
            <TableHead>地址</TableHead>
            <TableHead>协议</TableHead>
            <TableHead>端口</TableHead>
            <TableHead>配置模式</TableHead>
            <TableHead>在线</TableHead>
            <TableHead>状态</TableHead>
            <TableHead class="w-24">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-if="loading">
            <TableCell colspan="9" class="h-24 text-center text-muted-foreground">
              加载中...
            </TableCell>
          </TableRow>
          <TableRow v-else-if="nodes.length === 0">
            <TableCell colspan="9" class="h-24 text-center text-muted-foreground">
              暂无数据
            </TableCell>
          </TableRow>
          <TableRow v-for="node in nodes" :key="node.id">
            <TableCell class="font-medium">{{ node.id }}</TableCell>
            <TableCell>{{ node.name }}</TableCell>
            <TableCell class="font-mono text-xs">{{ node.address }}</TableCell>
            <TableCell>
              <Badge variant="outline">{{ node.protocol }}</Badge>
            </TableCell>
            <TableCell>{{ node.port }}</TableCell>
            <TableCell>{{ node.config_mode }}</TableCell>
            <TableCell>
              <Badge :variant="node.online ? 'default' : 'destructive'" class="gap-1">
                <span :class="['size-1.5 rounded-full', node.online ? 'bg-green-500' : 'bg-red-500']" />
                {{ node.online ? '在线' : '离线' }}
              </Badge>
            </TableCell>
            <TableCell>
              <Badge :variant="node.status === 1 ? 'default' : 'destructive'">
                {{ node.status === 1 ? '启用' : '禁用' }}
              </Badge>
            </TableCell>
            <TableCell>
              <div class="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon-sm"
                  :disabled="restartingId === node.id"
                  @click="handleRestart(node.id)"
                >
                  <RotateCw :class="['size-4', restartingId === node.id && 'animate-spin']" />
                </Button>
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon-sm">
                      <MoreHorizontal class="size-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent>
                    <DropdownMenuItem @click="openEdit(node)">
                      <Pencil class="size-4" />
                      编辑
                    </DropdownMenuItem>
                    <DropdownMenuItem @click="confirmDelete(node)" class="text-destructive">
                      <Trash2 class="size-4" />
                      删除
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
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
      <DialogContent class="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{{ editingNode ? '编辑节点' : '创建节点' }}</DialogTitle>
          <DialogDescription>
            {{ editingNode ? '修改节点信息' : '填写新节点信息' }}
          </DialogDescription>
        </DialogHeader>
        <form class="grid gap-4 py-4" @submit.prevent="handleSave">
          <div class="grid gap-2">
            <Label for="node-name">名称</Label>
            <Input id="node-name" v-model="form.name" required />
          </div>
          <div class="grid gap-2">
            <Label for="node-address">地址</Label>
            <Input id="node-address" v-model="form.address" placeholder="example.com" required />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label>协议</Label>
              <Select v-model="form.protocol">
                <SelectTrigger class="w-full">
                  <SelectValue placeholder="选择协议" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="vless">vless</SelectItem>
                  <SelectItem value="hysteria2">hysteria2</SelectItem>
                  <SelectItem value="tuic">tuic</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="grid gap-2">
              <Label for="node-port">端口</Label>
              <Input id="node-port" v-model.number="form.port" type="number" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label>配置模式</Label>
              <Select v-model="form.config_mode">
                <SelectTrigger class="w-full">
                  <SelectValue placeholder="选择模式" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="auto">自动</SelectItem>
                  <SelectItem value="manual">手动</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="grid gap-2">
              <Label for="node-sort">排序</Label>
              <Input id="node-sort" v-model.number="form.sort" type="number" />
            </div>
          </div>
          <div class="grid gap-2">
            <Label for="node-config">配置 JSON</Label>
            <textarea
              id="node-config"
              v-model="form.config_json"
              rows="4"
              class="border-input bg-background placeholder:text-muted-foreground rounded-md border px-3 py-2 text-sm"
              placeholder='{"key": "value"}'
            />
          </div>
          <div class="grid gap-2">
            <Label for="node-status">状态</Label>
            <select
              id="node-status"
              v-model.number="form.status"
              class="border-input bg-background h-8 rounded-md border px-3 text-sm"
            >
              <option :value="1">启用</option>
              <option :value="0">禁用</option>
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
            确定要删除节点 <strong>{{ deletingNode?.name }}</strong> 吗？此操作不可撤销。
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
