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
const loading = ref(false)
const error = ref('')
const siteLoaded = ref(false)

onMounted(async () => {
  // 已登录用户直接跳转到对应控制台
  const token = localStorage.getItem('token')
  if (token) {
    const isAdmin = localStorage.getItem('is_admin') === 'true'
    const adminPath = localStorage.getItem('admin_path') || 'admin'
    router.replace(isAdmin ? '/' + adminPath + '/dashboard' : '/user/dashboard')
    return
  }

  try {
    const res = await getSiteInfo()
    if (res.code === 0 && res.data) {
      if (res.data.app_name) { settingsStore.setAppName(res.data.app_name) }
      if (res.data.app_description) { settingsStore.setAppDescription(res.data.app_description) }
      if (res.data.admin_path) { localStorage.setItem('admin_path', res.data.admin_path) }
      if (res.data.auth_path) { localStorage.setItem('auth_path', res.data.auth_path) }
      if (res.data.user_path) { localStorage.setItem('user_path', res.data.user_path) }
      if (res.data.sub_path) { localStorage.setItem('sub_path', res.data.sub_path) }
    }
  } catch { /* ignore */ }
  siteLoaded.value = true
})

const email = ref('')
const password = ref('')

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
