<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getSettings, updateSettings } from '@/api/settings'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Save, Loader2 } from '@lucide/vue'

const settings = ref<Record<string, string>>({})
const loading = ref(true)
const saving = ref(false)

const settingGroups = [
  {
    title: '站点设置',
    description: '配置站点基本信息',
    fields: [
      { key: 'site_name', label: '站点名称', placeholder: 'Nexus' },
      { key: 'site_description', label: '站点描述', placeholder: 'Nexus 管理系统' },
      { key: 'site_url', label: '站点 URL', placeholder: 'https://example.com' },
    ],
  },
  {
    title: '邮件设置',
    description: '配置 SMTP 邮件服务',
    fields: [
      { key: 'smtp_host', label: 'SMTP 主机', placeholder: 'smtp.example.com' },
      { key: 'smtp_port', label: 'SMTP 端口', placeholder: '587' },
      { key: 'smtp_user', label: 'SMTP 用户名', placeholder: 'user@example.com' },
      { key: 'smtp_password', label: 'SMTP 密码', placeholder: '??????', type: 'password' },
      { key: 'smtp_from', label: '发件人地址', placeholder: 'noreply@example.com' },
    ],
  },
  {
    title: '订阅设置',
    description: '配置用户订阅相关参数',
    fields: [
      { key: 'subscribe_path', label: '订阅路径', placeholder: '/api/v1/subscribe' },
      { key: 'default_traffic_limit', label: '默认流量限制 (字节)', placeholder: '10737418240' },
      { key: 'default_device_limit', label: '默认设备限制', placeholder: '3' },
    ],
  },
]

async function fetchSettings() {
  loading.value = true
  try {
    const res = await getSettings()
    if (res.code === 0 && res.data) {
      settings.value = res.data
    }
  } catch (err) {
    console.error('获取设置失败:', err)
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    await updateSettings(settings.value)
  } catch (err) {
    console.error('保存设置失败:', err)
  } finally {
    saving.value = false
  }
}

onMounted(fetchSettings)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-lg font-semibold">系统设置</h2>
        <p class="text-sm text-muted-foreground">管理您的系统配置</p>
      </div>
      <Button :disabled="saving || loading" @click="handleSave">
        <Loader2 v-if="saving" class="size-4 animate-spin" />
        <Save v-else class="size-4" />
        {{ saving ? '保存中...' : '保存设置' }}
      </Button>
    </div>

    <template v-if="loading">
      <Card v-for="i in 3" :key="i">
        <CardContent class="p-6">
          <div class="space-y-4">
            <div class="h-4 w-32 animate-pulse rounded bg-muted" />
            <div class="h-4 w-48 animate-pulse rounded bg-muted" />
            <div class="h-8 w-full animate-pulse rounded bg-muted" />
          </div>
        </CardContent>
      </Card>
    </template>

    <template v-else>
      <Card v-for="group in settingGroups" :key="group.title">
        <CardHeader>
          <CardTitle>{{ group.title }}</CardTitle>
          <CardDescription>{{ group.description }}</CardDescription>
        </CardHeader>
        <CardContent>
          <div class="grid gap-4">
            <div v-for="field in group.fields" :key="field.key" class="grid gap-2">
              <Label :for="field.key">{{ field.label }}</Label>
              <Input
                :id="field.key"
                v-model="settings[field.key]"
                :type="field.type || 'text'"
                :placeholder="field.placeholder"
              />
            </div>
          </div>
        </CardContent>
      </Card>
    </template>
  </div>
</template>
