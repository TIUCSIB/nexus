<template>
  <div class="users-page">
    <el-card shadow="never">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索邮箱或UUID"
            style="width: 240px"
            clearable
            @clear="loadUsers"
            @keyup.enter="loadUsers"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-select v-model="searchStatus" placeholder="状态筛选" clearable style="width: 140px; margin-left: 10px" @change="loadUsers">
            <el-option label="正常" value="active" />
            <el-option label="禁用" value="disabled" />
            <el-option label="过期" value="expired" />
          </el-select>
          <el-button type="primary" style="margin-left: 10px" @click="loadUsers">搜索</el-button>
        </div>
        <el-button type="success" @click="openDialog('create')">添加用户</el-button>
      </div>

      <el-table :data="users" v-loading="loading" stripe style="margin-top: 16px">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="email" label="邮箱" min-width="180" />
        <el-table-column prop="uuid" label="UUID" min-width="200" show-overflow-tooltip />
        <el-table-column label="流量" min-width="160">
          <template #default="{ row }">
            <span>{{ formatTraffic(row.traffic_used) }} / {{ formatTraffic(row.traffic_limit) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="到期时间" width="170">
          <template #default="{ row }">
            {{ row.expire_at ? formatDate(row.expire_at) : '永不过期' }}
          </template>
        </el-table-column>
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
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadUsers"
          @current-change="loadUsers"
        />
      </div>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '添加用户' : '编辑用户'"
      width="560px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="密码" :prop="dialogMode === 'create' ? 'password' : ''">
          <el-input v-model="form.password" type="password" :placeholder="dialogMode === 'edit' ? '留空则不修改' : '请输入密码'" show-password />
        </el-form-item>
        <el-form-item label="套餐" prop="plan_id">
          <el-select v-model="form.plan_id" placeholder="选择套餐" clearable style="width: 100%">
            <el-option v-for="p in plans" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="流量限制(字节)" prop="traffic_limit">
          <el-input-number v-model="form.traffic_limit" :min="0" :step="1073741824" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="正常" value="active" />
            <el-option label="禁用" value="disabled" />
            <el-option label="过期" value="expired" />
          </el-select>
        </el-form-item>
        <el-form-item label="到期时间" prop="expire_at">
          <el-date-picker v-model="form.expire_at" type="datetime" placeholder="选择到期时间" style="width: 100%" value-format="YYYY-MM-DDTHH:mm:ssZ" />
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
import { listUsers, createUser, updateUser, deleteUser } from '@/api/user'
import { listPlans } from '@/api/plan'
import type { User, Plan } from '@/types'

const users = ref<User[]>([])
const plans = ref<Plan[]>([])
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
  email: '',
  password: '',
  plan_id: null as number | null,
  traffic_limit: 10737418240,
  status: 'active' as 'active' | 'disabled' | 'expired',
  expire_at: ''
})

const formRules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ]
}

function formatTraffic(bytes: number): string {
  if (!bytes || bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(2) + ' ' + units[i]
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString('zh-CN')
}

function statusTagType(status: string) {
  const map: Record<string, string> = { active: 'success', disabled: 'danger', expired: 'warning' }
  return (map[status] || 'info') as any
}

function statusLabel(status: string) {
  const map: Record<string, string> = { active: '正常', disabled: '禁用', expired: '过期' }
  return map[status] || status
}

async function loadUsers() {
  loading.value = true
  try {
    const res = await listUsers({ page: page.value, page_size: pageSize.value, keyword: searchKeyword.value, status: searchStatus.value || undefined })
    users.value = res.data.list || []
    total.value = res.data.total || 0
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

async function loadPlans() {
  try {
    const res = await listPlans({ page: 1, page_size: 100 })
    plans.value = res.data.list || []
  } catch {
    // ignore
  }
}

function openDialog(mode: 'create' | 'edit', row?: User) {
  dialogMode.value = mode
  dialogVisible.value = true
  if (mode === 'edit' && row) {
    editingId.value = row.id
    form.email = row.email
    form.password = ''
    form.plan_id = row.plan_id
    form.traffic_limit = row.traffic_limit
    form.status = row.status
    form.expire_at = row.expire_at || ''
  } else {
    editingId.value = 0
    form.email = ''
    form.password = ''
    form.plan_id = null
    form.traffic_limit = 10737418240
    form.status = 'active'
    form.expire_at = ''
  }
}

async function handleSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const data: any = { ...form }
    if (dialogMode.value === 'edit' && !data.password) {
      delete data.password
    }
    if (dialogMode.value === 'create') {
      await createUser(data)
      ElMessage.success('用户创建成功')
    } else {
      await updateUser(editingId.value, data)
      ElMessage.success('用户更新成功')
    }
    dialogVisible.value = false
    loadUsers()
  } catch {
    // handled by interceptor
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: User) {
  try {
    await ElMessageBox.confirm(`确定要删除用户 ${row.email} 吗？此操作不可撤销。`, '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteUser(row.id)
    ElMessage.success('用户已删除')
    loadUsers()
  } catch {
    // cancelled or error
  }
}

onMounted(() => {
  loadUsers()
  loadPlans()
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
</style>
