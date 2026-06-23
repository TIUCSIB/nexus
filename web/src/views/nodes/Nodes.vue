<template>
  <div class="nodes-page">
    <el-card shadow="never">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索节点名称或地址"
            style="width: 240px"
            clearable
            @clear="loadNodes"
            @keyup.enter="loadNodes"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-select v-model="searchStatus" placeholder="状态筛选" clearable style="width: 140px; margin-left: 10px" @change="loadNodes">
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="维护中" value="maintenance" />
          </el-select>
          <el-button type="primary" style="margin-left: 10px" @click="loadNodes">搜索</el-button>
        </div>
        <el-button type="success" @click="openDialog('create')">添加节点</el-button>
      </div>

      <el-table :data="nodes" v-loading="loading" stripe style="margin-top: 16px">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="节点名称" min-width="150" />
        <el-table-column prop="address" label="地址" min-width="150" />
        <el-table-column prop="port" label="端口" width="80" />
        <el-table-column label="协议" width="110">
          <template #default="{ row }">
            <el-tag type="info">{{ row.protocol.toUpperCase() }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="nodeStatusType(row.status)" effect="dark">
              <span class="status-dot" :class="'dot-' + row.status"></span>
              {{ nodeStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="config_mode" label="配置模式" width="100" />
        <el-table-column label="倍率" width="80">
          <template #default="{ row }">
            {{ row.traffic_rate }}x
          </template>
        </el-table-column>
        <el-table-column prop="sort_order" label="排序" width="70" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog('edit', row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadNodes"
          @current-change="loadNodes"
        />
      </div>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '添加节点' : '编辑节点'"
      width="560px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="节点名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入节点名称" />
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input v-model="form.address" placeholder="IP或域名" />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input-number v-model="form.port" :min="1" :max="65535" style="width: 100%" />
        </el-form-item>
        <el-form-item label="协议" prop="protocol">
          <el-select v-model="form.protocol" style="width: 100%">
            <el-option label="VMess" value="vmess" />
            <el-option label="VLESS" value="vless" />
            <el-option label="Trojan" value="trojan" />
            <el-option label="Shadowsocks" value="shadowsocks" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="维护中" value="maintenance" />
          </el-select>
        </el-form-item>
        <el-form-item label="配置模式" prop="config_mode">
          <el-input v-model="form.config_mode" placeholder="如: default, cdn, ws" />
        </el-form-item>
        <el-form-item label="流量倍率" prop="traffic_rate">
          <el-input-number v-model="form.traffic_rate" :min="0" :max="100" :precision="1" :step="0.1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="排序" prop="sort_order">
          <el-input-number v-model="form.sort_order" :min="0" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listNodes, createNode, updateNode, deleteNode } from '@/api/node'
import type { Node } from '@/types'

const nodes = ref<Node[]>([])
const loading = ref(false)
const submitting = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const searchKeyword = ref('')
const searchStatus = ref('')
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingId = ref(0)
const formRef = ref<FormInstance>()

const form = reactive({
  name: '',
  address: '',
  port: 443,
  protocol: 'vmess' as 'vmess' | 'vless' | 'trojan' | 'shadowsocks',
  status: 'online' as 'online' | 'offline' | 'maintenance',
  config_mode: 'default',
  traffic_rate: 1.0,
  sort_order: 0
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入节点名称', trigger: 'blur' }],
  address: [{ required: true, message: '请输入地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  protocol: [{ required: true, message: '请选择协议', trigger: 'change' }]
}

function nodeStatusType(status: string) {
  const map: Record<string, string> = { online: 'success', offline: 'danger', maintenance: 'warning' }
  return (map[status] || 'info') as any
}

function nodeStatusLabel(status: string) {
  const map: Record<string, string> = { online: '在线', offline: '离线', maintenance: '维护中' }
  return map[status] || status
}

async function loadNodes() {
  loading.value = true
  try {
    const res = await listNodes({ page: page.value, page_size: pageSize.value, keyword: searchKeyword.value, status: searchStatus.value || undefined })
    nodes.value = res.data.list || []
    total.value = res.data.total || 0
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

function openDialog(mode: 'create' | 'edit', row?: Node) {
  dialogMode.value = mode
  dialogVisible.value = true
  if (mode === 'edit' && row) {
    editingId.value = row.id
    form.name = row.name
    form.address = row.address
    form.port = row.port
    form.protocol = row.protocol
    form.status = row.status
    form.config_mode = row.config_mode || 'default'
    form.traffic_rate = row.traffic_rate || 1.0
    form.sort_order = row.sort_order || 0
  } else {
    editingId.value = 0
    form.name = ''
    form.address = ''
    form.port = 443
    form.protocol = 'vmess'
    form.status = 'online'
    form.config_mode = 'default'
    form.traffic_rate = 1.0
    form.sort_order = 0
  }
}

async function handleSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (dialogMode.value === 'create') {
      await createNode(form)
      ElMessage.success('节点创建成功')
    } else {
      await updateNode(editingId.value, form)
      ElMessage.success('节点更新成功')
    }
    dialogVisible.value = false
    loadNodes()
  } catch {
    // handled by interceptor
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: Node) {
  try {
    await ElMessageBox.confirm(`确定要删除节点「${row.name}」吗？`, '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteNode(row.id)
    ElMessage.success('节点已删除')
    loadNodes()
  } catch {
    // cancelled or error
  }
}

onMounted(() => {
  loadNodes()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
}

.dot-online {
  background-color: #67c23a;
}

.dot-offline {
  background-color: #f56c6c;
}

.dot-maintenance {
  background-color: #e6a23c;
}
</style>
