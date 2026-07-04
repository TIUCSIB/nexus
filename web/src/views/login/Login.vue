<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSettingsStore } from '@/stores/settings'
import { getSiteInfo } from '@/api/settings'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { onMounted } from 'vue'

const router = useRouter()
const authStore = useAuthStore()
const settingsStore = useSettingsStore()
getSiteInfo().then((res: any) => {
  if (res.code === 0 && res.data) {
    if (res.data.app_name) { settingsStore.setAppName(res.data.app_name) }
    if (res.data.app_description) { settingsStore.setAppDescription(res.data.app_description) }
  }
}).catch(() => { })

const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  if (!email.value || !password.value) {
    error.value = '请填写邮箱和密码'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const success = await authStore.login(email.value, password.value)
    if (success) {
      if (authStore.isAdmin) {
        router.push(settingsStore.adminRoute('dashboard'))
      } else {
        router.push('/user/dashboard')
      }
    } else {
      error.value = '登录失败，请检查邮箱和密码'
    }
  } catch {
    error.value = '登录失败，请检查网络连接'
  } finally {
    loading.value = false
  }
}
onMounted(() => {
  document.title = settingsStore.appName + ' - 登录'
})
</script>

<template>
  <div class="flex min-h-svh items-center justify-center bg-muted/40 p-4">
    <Card class="w-full max-w-sm">
      <CardHeader class="text-center">
        <CardTitle class="text-2xl">{{ settingsStore.appName }}</CardTitle>
        <CardDescription v-if="settingsStore.appDescription">{{ settingsStore.appDescription }}</CardDescription>
        <CardDescription>请使用账号登录</CardDescription>
      </CardHeader>
      <CardContent>
        <form class="grid gap-4" @submit.prevent="handleLogin">
          <div class="grid gap-2">
            <Label for="email">邮箱</Label>
            <Input id="email" v-model="email" type="email" placeholder="admin@example.com" autocomplete="email" />
          </div>
          <div class="grid gap-2">
            <Label for="password">密码</Label>
            <Input id="password" v-model="password" type="password" placeholder="请输入密码"
              autocomplete="current-password" />
          </div>
          <p v-if="error" class="text-sm text-red-500">{{ error }}</p>
          <Button type="submit" class="w-full" :disabled="loading">
            {{ loading ? '登录中...' : '登录' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
