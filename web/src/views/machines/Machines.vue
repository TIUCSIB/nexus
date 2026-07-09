<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSettingsStore } from '@/stores/settings'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { toast } from 'vue-sonner'
import { MoreHorizontal, Plus, Pencil, Trash2, Copy, RefreshCw, Terminal, Eye } from 'lucide-vue-next'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { listMachines, createMachine, updateMachine, deleteMachine, resetMachineToken, getMachineInstallCommand } from '@/api/machine'
import type { ApiResponse, Machine, MachineCreateResult } from '@/types'
	
	const router = useRouter()
	const settingsStore = useSettingsStore()

const machines = ref<Machine[]>([])
const loading = ref(false)
const dialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const tokenDialogOpen = ref(false)
const commandDialogOpen = ref(false)
const editing = ref<Partial<Machine>>({})
const isEdit = ref(false)
const saving = ref(false)
const selectedId = ref(0)
const newToken = ref('')
const installCommand = ref('')

// Create/Edit dialog form
const form = ref({ name: '', notes: '', is_active: true })

async function fetchMachines() {
  try {
    const res = await listMachines()
    if (res.code === 0 && res.data) {
      machines.value = res.data
    }
  } catch { toast.error('获取机器列表失败') }
}

function openCreate() {
  isEdit.value = false
  form.value = { name: '', notes: '', is_active: true }
  dialogOpen.value = true
}

function openEdit(m: Machine) {
  isEdit.value = true
  form.value = { name: m.name, notes: m.notes || '', is_active: m.is_active }
  editing.value = { ...m }
  dialogOpen.value = true
}

async function handleSave() {
  if (!form.value.name) {
    toast.error('请输入机器名称')
    return
  }
  saving.value = true
  try {
    if (isEdit.value && editing.value.id) {
      const res = await updateMachine(editing.value.id, { name: form.value.name, notes: form.value.notes, is_active: form.value.is_active })
      if (res.code === 0) {
        toast.success('保存成功')
        dialogOpen.value = false
        fetchMachines()
      } else {
        toast.error(res.message || '保存失败')
      }
    } else {
      const res = await createMachine({ name: form.value.name, notes: form.value.notes })
      if (res.code === 0 && res.data) {
        toast.success('创建成功')
        dialogOpen.value = false
        // Show token
        newToken.value = res.data.token
        installCommand.value = res.data.install_command
        tokenDialogOpen.value = true
        fetchMachines()
      } else {
        toast.error(res.message || '创建失败')
      }
    }
  } catch (e: any) { toast.error(e?.response?.data?.message || '操作失败') }
  finally { saving.value = false }
}

function confirmDelete(m: Machine) {
  selectedId.value = m.id
  editing.value = { ...m }
  deleteDialogOpen.value = true
}

async function handleDelete() {
  try {
    const res = await deleteMachine(selectedId.value)
    if (res.code === 0) {
      toast.success('删除成功')
      deleteDialogOpen.value = false
      fetchMachines()
    } else {
      toast.error(res.message || '删除失败')
    }
  } catch (e: any) { toast.error(e?.response?.data?.message || '删除失败') }
}

async function handleResetToken(id: number) {
  try {
    const res = await resetMachineToken(id)
    if (res.code === 0 && res.data) {
      newToken.value = res.data.token
      tokenDialogOpen.value = true
    } else {
      toast.error(res.message || '重置失败')
    }
  } catch (e: any) { toast.error(e?.response?.data?.message || '重置失败') }
}

async function handleShowCommand(id: number) {
  try {
    const res = await getMachineInstallCommand(id)
    if (res.code === 0 && res.data) {
      installCommand.value = res.data.command
      commandDialogOpen.value = true
    } else {
      toast.error(res.message || '获取命令失败')
    }
  } catch (e: any) { toast.error(e?.response?.data?.message || '获取命令失败') }
}

function copyText(text: string) {
  navigator.clipboard.writeText(text)
  toast.success('已复制到剪贴板')
}

function formatDate(t: string | null) {
  if (!t) return '从未'
  return new Date(t).toLocaleString('zh-CN')
}

function formatBytes(b: number) {
  if (b === 0) return '0 B'
  const k = 1024; const s = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(k))
  return (b / Math.pow(k, i)).toFixed(1) + ' ' + s[i]
}

function getOnlineStatus(t: string | null) {
  if (!t) return { label: '从未连接', online: false }
  const elapsed = Date.now() - new Date(t).getTime()
  if (elapsed < 180000) return { label: '在线', online: true }
  const mins = Math.floor(elapsed / 60000)
  return { label: `${mins}分钟前离线`, online: false }
}

function cpuColor(cpu: number) {
  if (cpu < 50) return 'bg-green-500'
  if (cpu < 80) return 'bg-yellow-500'
  return 'bg-red-500'
}

onMounted(fetchMachines)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">机器管理</h1>
      <Button @click="openCreate">
        <Plus class="h-4 w-4 mr-2" />创建机器
      </Button>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>机器列表</CardTitle>
        <CardDescription>管理部署节点 Agent 的物理机器。机器模式下，一台机器通过一个 WebSocket 连接管理多个节点</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-16">ID</TableHead>
              <TableHead>名称</TableHead>
              <TableHead class="w-20">节点数</TableHead>
              <TableHead class="w-20">启用</TableHead>
              <TableHead class="w-28">负载</TableHead>
              <TableHead class="w-36">最后在线</TableHead>
              <TableHead class="w-36">创建时间</TableHead>
              <TableHead class="w-[80px]">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="m in machines" :key="m.id" class="cursor-pointer hover:bg-muted/50" @click="router.push(settingsStore.adminRoute('machines/' + m.id))">
              <TableCell class="font-mono text-xs">{{ m.id }}</TableCell>
              <TableCell class="font-medium">{{ m.name }}</TableCell>
              <TableCell>
                <Badge variant="secondary">{{ m.servers_count }}</Badge>
              </TableCell>
              <TableCell>
                <Badge :variant="m.is_active ? 'default' : 'secondary'">
                  {{ m.is_active ? '启用' : '停用' }}
                </Badge>
              </TableCell>
              <TableCell>
                <div v-if="m.load_status" class="flex flex-col gap-0.5 min-w-[100px]">
                  <div class="flex items-center gap-1.5 text-xs">
                    <span class="w-8 text-muted-foreground">CPU</span>
                    <div class="flex-1 h-1.5 rounded-full bg-muted overflow-hidden">
                      <div :class="cpuColor(m.load_status.cpu)" class="h-full rounded-full" :style="{ width: Math.min(m.load_status.cpu, 100) + '%' }" />
                    </div>
                    <span class="w-9 text-right font-mono text-xs">{{ m.load_status.cpu.toFixed(1) }}%</span>
                  </div>
                  <div class="flex items-center gap-1.5 text-xs">
                    <span class="w-8 text-muted-foreground">MEM</span>
                    <div class="flex-1 h-1.5 rounded-full bg-muted overflow-hidden">
                      <div class="bg-blue-500 h-full rounded-full" :style="{ width: (m.load_status.mem_total > 0 ? (m.load_status.mem_used / m.load_status.mem_total * 100) : 0) + '%' }" />
                    </div>
                    <span class="w-9 text-right font-mono text-xs">{{ (m.load_status.mem_total > 0 ? (m.load_status.mem_used / m.load_status.mem_total * 100).toFixed(0) : '0') }}%</span>
                  </div>
                </div>
                <span v-else class="text-xs text-muted-foreground">-</span>
              </TableCell>
              <TableCell class="text-sm">
                <div class="flex items-center gap-1.5">
                  <span :class="getOnlineStatus(m.last_seen_at).online ? 'bg-green-500' : 'bg-gray-400'" class="w-2 h-2 rounded-full shrink-0" />
                  <span class="text-muted-foreground text-xs">{{ formatDate(m.last_seen_at) }}</span>
                </div>
              </TableCell>
              <TableCell class="text-sm text-muted-foreground">{{ formatDate(m.created_at) }}</TableCell>
              <TableCell @click.stop>
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon">
                      <MoreHorizontal class="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem @click="router.push(settingsStore.adminRoute('machines/' + m.id))">
                      <Eye class="h-4 w-4 mr-2" />详情
                    </DropdownMenuItem>
                    <DropdownMenuItem @click="openEdit(m)">
                      <Pencil class="h-4 w-4 mr-2" />编辑
                    </DropdownMenuItem>
                    <DropdownMenuItem @click="handleShowCommand(m.id)">
                      <Terminal class="h-4 w-4 mr-2" />安装命令
                    </DropdownMenuItem>
                    <DropdownMenuItem @click="handleResetToken(m.id)">
                      <RefreshCw class="h-4 w-4 mr-2" />重置Token
                    </DropdownMenuItem>
                    <DropdownMenuItem @click="confirmDelete(m)">
                      <Trash2 class="h-4 w-4 mr-2" />删除
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
            <TableRow v-if="machines.length === 0">
              <TableCell colspan="8" class="text-center text-muted-foreground py-8">暂无机器</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <!-- Create/Edit Dialog -->
    <Dialog v-model:open="dialogOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ isEdit ? '编辑机器' : '创建机器' }}</DialogTitle>
          <DialogDescription>{{ isEdit ? '修改机器名称和备注' : '创建新的机器，创建后将显示 Token 和安装命令' }}</DialogDescription>
        </DialogHeader>
        <div class="space-y-4 py-4">
          <div class="grid gap-2">
            <Label>机器名称</Label>
            <Input v-model="form.name" placeholder="例如：东京VPS" />
          </div>
          <div class="grid gap-2">
            <Label>备注</Label>
            <Textarea v-model="form.notes" placeholder="可选备注信息" rows="2" />
          </div>
          <div class="flex items-center justify-between rounded-lg border p-3">
            <div>
              <p class="text-sm font-medium">启用服务器</p>
              <p class="text-xs text-muted-foreground">禁用后 nexus-agent 将不再使用此服务器</p>
            </div>
            <Switch v-model="form.is_active" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="dialogOpen = false">取消</Button>
          <Button @click="handleSave" :disabled="saving">{{ saving ? '保存中...' : '保存' }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Delete Dialog -->
    <Dialog v-model:open="deleteDialogOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>确认删除</DialogTitle>
          <DialogDescription>确定要删除机器「{{ editing.name }}」吗？删除后关联的节点将解除机器绑定</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="deleteDialogOpen = false">取消</Button>
          <Button variant="destructive" @click="handleDelete">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Token Display Dialog -->
    <Dialog v-model:open="tokenDialogOpen">
      <DialogContent class="max-w-lg">
        <DialogHeader>
          <DialogTitle>机器 Token</DialogTitle>
          <DialogDescription>请妥善保管此 Token，关闭后将无法再次查看。如需重置请使用「重置Token」功能</DialogDescription>
        </DialogHeader>
        <div class="space-y-4 py-4">
          <div class="grid gap-2">
            <Label>Token</Label>
            <div class="flex items-center gap-2">
              <code class="flex-1 rounded-md bg-muted p-3 text-sm font-mono break-all">{{ newToken }}</code>
              <Button variant="outline" size="icon" @click="copyText(newToken)">
                <Copy class="h-4 w-4" />
              </Button>
            </div>
          </div>
          <div class="grid gap-2">
            <Label>一键安装命令</Label>
            <div class="flex items-start gap-2">
              <code class="flex-1 rounded-md bg-muted p-3 text-sm font-mono text-xs break-all whitespace-pre-wrap">{{ installCommand }}</code>
              <Button variant="outline" size="icon" @click="copyText(installCommand)">
                <Copy class="h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button @click="tokenDialogOpen = false">关闭</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Install Command Dialog -->
    <Dialog v-model:open="commandDialogOpen">
      <DialogContent class="max-w-lg">
        <DialogHeader>
          <DialogTitle>一键安装命令</DialogTitle>
          <DialogDescription>在目标服务器上执行此命令，即可自动安装并注册机器</DialogDescription>
        </DialogHeader>
        <div class="py-4">
          <div class="flex items-start gap-2">
            <code class="flex-1 rounded-md bg-muted p-3 text-sm font-mono text-xs break-all whitespace-pre-wrap">{{ installCommand }}</code>
            <Button variant="outline" size="icon" @click="copyText(installCommand)">
              <Copy class="h-4 w-4" />
            </Button>
          </div>
        </div>
        <DialogFooter>
          <Button @click="commandDialogOpen = false">关闭</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>