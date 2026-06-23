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
    error.value = 'ÇëÌîĞ´ÓÊÏäºÍÃÜÂë'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const success = await authStore.login(email.value, password.value)
    if (success) {
      router.push('/dashboard')
    } else {
      error.value = 'µÇÂ¼Ê§°Ü£¬Çë¼ì²éÓÊÏäºÍÃÜÂë'
    }
  } catch {
    error.value = 'µÇÂ¼Ê§°Ü£¬Çë¼ì²éÍøÂçÁ¬½Ó'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex min-h-svh items-center justify-center bg-muted/40 p-4">
    <Card class="w-full max-w-sm">
      <CardHeader class="text-center">
        <CardTitle class="text-2xl">Nexus ¹ÜÀíºóÌ¨</CardTitle>
        <CardDescription>ÇëÊ¹ÓÃ¹ÜÀíÔ±ÕËºÅµÇÂ¼</CardDescription>
      </CardHeader>
      <CardContent>
        <form class="grid gap-4" @submit.prevent="handleLogin">
          <div class="grid gap-2">
            <Label for="email">ÓÊÏä</Label>
            <Input
              id="email"
              v-model="email"
              type="email"
              placeholder="admin@example.com"
              autocomplete="email"
            />
          </div>
          <div class="grid gap-2">
            <Label for="password">ÃÜÂë</Label>
            <Input
              id="password"
              v-model="password"
              type="password"
              placeholder="ÇëÊäÈëÃÜÂë"
              autocomplete="current-password"
            />
          </div>
          <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
          <Button type="submit" class="w-full" :disabled="loading">
            {{ loading ? 'µÇÂ¼ÖĞ...' : 'µÇÂ¼' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
