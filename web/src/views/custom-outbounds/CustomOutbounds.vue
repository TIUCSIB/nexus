<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { toast } from 'vue-sonner'
import { Plus, Pencil, Trash2, Link } from 'lucide-vue-next'
import { listCustomOutbounds, createCustomOutbound, updateCustomOutbound, deleteCustomOutbound } from '@/api/customOutbound'
import type { CustomOutbound } from '@/types'

const outbounds = ref<CustomOutbound[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const saving = ref(false)
const showDialog = ref(false)
const editing = ref<Partial<CustomOutbound>>({})
const isEdit = ref(false)

const protocols = [
  'vless', 'vmess', 'trojan', 'shadowsocks',
  'hysteria2', 'tuic', 'wireguard', 'http', 'socks', 'direct',
  'block', 'dns', 'selector', 'urltest',
]

function resetForm() {
  editing.value = {
    name: '', tag: '', protocol: 'vless',
    settings_json: '', proxy_tag: '',
    sort: 0, status: 1,
  }
  isEdit.value = false
}

function openCreate() {
  resetForm()
  showDialog.value = true
}

function openEdit(item: CustomOutbound) {
  editing.value = { ...item }
  isEdit.value = true
  showDialog.value = true
}

async function handleSave() {
  if (!editing.value.name || !editing.value.tag || !editing.value.protocol) {
    toast.error('请填写出站名称、标签和协议')
    return
  }
  saving.value = true
  try {
    if (isEdit.value && editing.value.id) {
      const res = await updateCustomOutbound(editing.value.id, editing.value)
      if (res.code === 0) {
        toast.success('自定义出站已更新')
        showDialog.value = false
        fetchData()
      } else {
        toast.error(res.message || '更新失败')
      }
    } else {
      const res = await createCustomOutbound(editing.value as any)
      if (res.code === 0) {
        toast.success('自定义出站已创建')
        showDialog.value = false
        fetchData()
      } else {
        toast.error(res.message || '创建失败')
      }
    }
  } catch (e: any) {
    toast.error(e?.message || '操作失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(item: CustomOutbound) {
  if (!confirm(`确定要删除出站「${item.name}」吗？`)) return
  try {
    const res = await deleteCustomOutbound(item.id)
    if (res.code === 0) {
      toast.success('已删除')
      fetchData()
    } else {
      toast.error(res.message || '删除失败')
    }
  } catch (e: any) {
    toast.error(e?.message || '删除失败')
  }
}

const statusBadge = (s: number) => s === 1 ? 'default' : 'secondary'
const statusLabel = (s: number) => s === 1 ? '启用' : '禁用'

async function fetchData() {
  loading.value = true
  try {
    const res = await listCustomOutbounds({ page: page.value, page_size: pageSize.value })
    if (res.code === 0) {
      outbounds.value = (res as any).data.items || (res as any).data || []
      total.value = (res as any).data.total || (res as any).data.length || 0
    }
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold flex items-center gap-2">
        <Link class="h-5 w-5" />自定义出站
      </h1>
      <Button @click="openCreate">
        <Plus class="mr-1 h-4 w-4" />新建出站
      </Button>
    </div>

    <Card>
      <CardHeader><CardTitle>出站列表（落地/链式代理）</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>名称</TableHead>
              <TableHead>标签</TableHead>
              <TableHead>协议</TableHead>
              <TableHead>代理链</TableHead>
              <TableHead>排序</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="item in outbounds" :key="item.id">
              <TableCell class="font-medium">{{ item.name }}</TableCell>
              <TableCell><Badge variant="outline" class="font-mono">{{ item.tag }}</Badge></TableCell>
              <TableCell><Badge variant="secondary">{{ item.protocol }}</Badge></TableCell>
              <TableCell>
                <span v-if="item.proxy_tag" class="text-sm font-mono text-muted-foreground">{{ item.proxy_tag }}</span>
                <span v-else class="text-sm text-muted-foreground">-</span>
              </TableCell>
              <TableCell class="text-sm">{{ item.sort }}</TableCell>
              <TableCell>
                <Badge :variant="statusBadge(item.status)">{{ statusLabel(item.status) }}</Badge>
              </TableCell>
              <TableCell>
                <div class="flex gap-1">
                  <Button variant="ghost" size="icon" @click="openEdit(item)">
                    <Pencil class="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="icon" class="text-destructive" @click="handleDelete(item)">
                    <Trash2 class="h-4 w-4" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <p v-if="!loading && outbounds.length === 0" class="text-center text-muted-foreground py-8">暂无自定义出站，点击右上角新建</p>
      </CardContent>
    </Card>

    <!-- Create/Edit Dialog -->
    <Dialog v-model:open="showDialog">
      <DialogContent class="max-w-lg max-h-[90vh] overflow-y-auto scrollbar-none">
        <DialogHeader>
          <DialogTitle>{{ isEdit ? '编辑' : '新建' }}自定义出站</DialogTitle>
        </DialogHeader>
        <div class="grid gap-4 py-4">
          <div class="grid gap-2">
            <Label>名称</Label>
            <Input v-model="editing.name" placeholder="例如：日本落地" />
          </div>
          <div class="grid gap-2">
            <Label>标签</Label>
            <Input v-model="editing.tag" placeholder="例如：jp-out" :disabled="isEdit" />
            <p class="text-xs text-muted-foreground">唯一标识，用于在路由规则中引用</p>
          </div>
          <div class="grid gap-2">
            <Label>协议</Label>
            <Select v-model="editing.protocol">
              <SelectTrigger><SelectValue placeholder="选择协议" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="p in protocols" :key="p" :value="p">{{ p }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="grid gap-2">
            <Label>配置参数 (JSON)</Label>
            <Input v-model="editing.settings_json" placeholder='{"server":"1.2.3.4","port":443,"uuid":"..."}' />
            <p class="text-xs text-muted-foreground">协议特定的配置参数，直接透传给内核</p>
          </div>
          <div class="grid gap-2">
            <Label>代理链标签</Label>
            <Input v-model="editing.proxy_tag" placeholder="可选，例如：jp-out" />
            <p class="text-xs text-muted-foreground">链式代理时指向下一个出站标签，留空则直连</p>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label>排序</Label>
              <Input v-model="editing.sort" type="number" placeholder="0" />
            </div>
            <div class="grid gap-2">
              <Label>状态</Label>
              <Select v-model="editing.status">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem :value="1">启用</SelectItem>
                  <SelectItem :value="0">禁用</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showDialog = false">取消</Button>
          <Button @click="handleSave" :disabled="saving">{{ saving ? '保存中...' : '保存' }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>