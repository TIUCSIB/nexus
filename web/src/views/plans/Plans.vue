<template>
  <div class="plans-page">
    <el-card shadow="never">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索套餐名称"
            style="width: 240px"
            clearable
            @clear="loadPlans"
            @keyup.enter="loadPlans"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-button type="primary" style="margin-left: 10px" @click="loadPlans">搜索</el-button>
        </div>
        <el-button type="success" @click="openDialog('create')">添加套餐</el-button>
      </div>

      <el-table :data="plans" v-loading="loading" stripe style="margin-top: 16px">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="套餐名称" min-width="150" />
        <el-table-column label="流量限制" min-width="120">
          <template #default="{ row }">
            {{ formatTraffic(row.traffic_limit) }}
          </template>
        </el-table-column>
        <el-table-column label="有效天数" width="100">
          <template #default="{ row }">
            {{ row.duration_days }} 天
          </template>
        </el-table-column>
        <el-table-column label="价格" width="100">
          <template #default="{ row }">
            &yen;{{ row.price.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
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
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadPlans"
          @current-change="loadPlans"
        />
      </div>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '添加套餐' : '编辑套餐'"
      width="520px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="套餐名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入套餐名称" />
        </el-form-item>
        <el-form-item label="流量限制(字节)" prop="traffic_limit">
          <el-input-number v-model="form.traffic_limit" :min="0" :step="1073741824" style="width: 100%" />
        </el-form-item>
        <el-form-item label="有效天数" prop="duration_days">
          <el-input-number v-model="form.duration_days" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="价格" prop="price">
          <el-input-number v-model="form.price" :min="0" :precision="2" :step="10" style="width: 100%" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入套餐描述" />
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
import { listPlans, createPlan, updatePlan, deletePlan } from '@/api/plan'
import type { Plan } from '@/types'

const plans = ref<Plan[]>([])
const loading = ref(false)
const submitting = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const searchKeyword = ref('')
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingId = ref(0)
const formRef = ref<FormInstance>()

const form = reactive({
  name: '',
  traffic_limit: 107374182400,
  duration_days: 30,
  price: 0,
  description: ''
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入套餐名称', trigger: 'blur' }],
  traffic_limit: [{ required: true, message: '请输入流量限制', trigger: 'blur' }],
  duration_days: [{ required: true, message: '请输入有效天数', trigger: 'blur' }],
  price: [{ required: true, message: '请输入价格', trigger: 'blur' }]
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

async function loadPlans() {
  loading.value = true
  try {
    const res = await listPlans({ page: page.value, page_size: pageSize.value, keyword: searchKeyword.value })
    plans.value = res.data.list || []
    total.value = res.data.total || 0
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

function openDialog(mode: 'create' | 'edit', row?: Plan) {
  dialogMode.value = mode
  dialogVisible.value = true
  if (mode === 'edit' && row) {
    editingId.value = row.id
    form.name = row.name
    form.traffic_limit = row.traffic_limit
    form.duration_days = row.duration_days
    form.price = row.price
    form.description = row.description || ''
  } else {
    editingId.value = 0
    form.name = ''
    form.traffic_limit = 107374182400
    form.duration_days = 30
    form.price = 0
    form.description = ''
  }
}

async function handleSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (dialogMode.value === 'create') {
      await createPlan(form)
      ElMessage.success('套餐创建成功')
    } else {
      await updatePlan(editingId.value, form)
      ElMessage.success('套餐更新成功')
    }
    dialogVisible.value = false
    loadPlans()
  } catch {
    // handled by interceptor
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: Plan) {
  try {
    await ElMessageBox.confirm(`确定要删除套餐「${row.name}」吗？`, '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deletePlan(row.id)
    ElMessage.success('套餐已删除')
    loadPlans()
  } catch {
    // cancelled or error
  }
}

onMounted(() => {
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
