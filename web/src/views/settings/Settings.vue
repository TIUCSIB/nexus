<template>
  <div class="settings-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>系统设置</span>
        </div>
      </template>
      <el-form
        ref="formRef"
        :model="form"
        label-width="140px"
        v-loading="loading"
        style="max-width: 640px"
      >
        <el-divider content-position="left">基本设置</el-divider>
        <el-form-item label="站点名称" prop="site_name">
          <el-input v-model="form.site_name" placeholder="Nexus Proxy" />
        </el-form-item>
        <el-form-item label="站点URL" prop="site_url">
          <el-input v-model="form.site_url" placeholder="https://example.com" />
        </el-form-item>
        <el-form-item label="管理员邮箱" prop="admin_email">
          <el-input v-model="form.admin_email" placeholder="admin@example.com" />
        </el-form-item>

        <el-divider content-position="left">注册设置</el-divider>
        <el-form-item label="开放注册" prop="register_enabled">
          <el-switch v-model="form.register_enabled" active-text="开启" inactive-text="关闭" />
        </el-form-item>
        <el-form-item label="仅邀请注册" prop="invite_only">
          <el-switch v-model="form.invite_only" active-text="是" inactive-text="否" />
        </el-form-item>

        <el-divider content-position="left">默认用户配置</el-divider>
        <el-form-item label="默认流量限制(字节)" prop="default_traffic_limit">
          <el-input-number v-model="form.default_traffic_limit" :min="0" :step="1073741824" style="width: 100%" />
        </el-form-item>
        <el-form-item label="默认套餐" prop="default_plan_id">
          <el-select v-model="form.default_plan_id" placeholder="无" clearable style="width: 100%">
            <el-option v-for="p in plans" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>

        <el-divider content-position="left">安全设置</el-divider>
        <el-form-item label="JWT密钥" prop="jwt_secret">
          <el-input v-model="form.jwt_secret" type="password" show-password placeholder="JWT Secret Key" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="saving" @click="handleSave">保存设置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import type { FormInstance } from 'element-plus'
import { ElMessage } from 'element-plus'
import { getSettings, updateSettings } from '@/api/settings'
import { listPlans } from '@/api/plan'
import type { Plan } from '@/types'

const formRef = ref<FormInstance>()
const loading = ref(false)
const saving = ref(false)
const plans = ref<Plan[]>([])

const form = reactive({
  site_name: '',
  site_url: '',
  admin_email: '',
  register_enabled: true,
  invite_only: false,
  default_traffic_limit: 10737418240,
  default_plan_id: null as number | null,
  jwt_secret: ''
})

async function loadSettings() {
  loading.value = true
  try {
    const res = await getSettings()
    Object.assign(form, res.data)
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

async function handleSave() {
  saving.value = true
  try {
    await updateSettings(form)
    ElMessage.success('设置保存成功')
  } catch {
    // handled by interceptor
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadSettings()
  loadPlans()
})
</script>

<style scoped>
.card-header {
  font-weight: 600;
  font-size: 16px;
}
</style>
