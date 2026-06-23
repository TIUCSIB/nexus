<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  if (!email.value || !password.value) {
    error.value = '请填写所有字段'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const success = await authStore.login(email.value, password.value)
    if (success) {
      router.push('/dashboard')
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
        <CardTitle class="text-2xl">Nexus 代理面板</CardTitle>
        <CardDescription>请使用管理员账号登录</CardDescription>
      </CardHeader>
      <CardContent>
        <form class="grid gap-4" @submit.prevent="handleLogin">
          <div class="grid gap-2">
            <Label for="email">邮箱</Label>
            <Input id="email" v-model="email" type="email" placeholder="admin@example.com" autocomplete="email" />
          </div>
          <div class="grid gap-2">
            <Label for="password">密码</Label>
            <Input id="password" v-model="password" type="password" placeholder="请输入密码" autocomplete="current-password" />
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