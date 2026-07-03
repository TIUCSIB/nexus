<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { User, Save } from 'lucide-vue-next'
import { getProfile, updateProfile } from '@/api/userSelf'
import { toast } from 'vue-sonner'
import type { User as UserType } from '@/types'

const user = ref<UserType | null>(null)
const loading = ref(true)
const saving = ref(false)

const form = ref({
  email: '',
  password: '',
  confirmPassword: '',
})

onMounted(async () => {
  try {
    const res = await getProfile()
    if (res.code === 0) {
      user.value = res.data
      form.value.email = res.data.email
    }
  } finally {
    loading.value = false
  }
})

async function handleSave() {
  if (!form.value.email) {
    toast.error('邮箱不能为空')
    return
  }

  if (form.value.password) {
    if (form.value.password.length < 8) {
      toast.error('密码长度不能少于8位')
      return
    }
    if (form.value.password !== form.value.confirmPassword) {
      toast.error('两次输入的密码不一致')
      return
    }
  }

  saving.value = true
  try {
    const data: { email?: string; password?: string } = {}
    if (form.value.email !== user.value?.email) {
      data.email = form.value.email
    }
    if (form.value.password) {
      data.password = form.value.password
    }

    if (Object.keys(data).length === 0) {
      toast.info('没有需要修改的内容')
      return
    }

    const res = await updateProfile(data)
    if (res.code === 0) {
      user.value = res.data
      form.value.password = ''
      form.value.confirmPassword = ''
      toast.success('资料更新成功')
    } else {
      toast.error(res.message || '更新失败')
    }
  } catch {
    toast.error('网络错误，请稍后重试')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold">个人资料</h1>

    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <User class="h-5 w-5" />
          账号设置
        </CardTitle>
        <CardDescription>修改你的登录邮箱和密码</CardDescription>
      </CardHeader>
      <CardContent>
        <form class="grid gap-4 max-w-md" @submit.prevent="handleSave">
          <div class="grid gap-2">
            <Label for="email">邮箱</Label>
            <Input id="email" v-model="form.email" type="email" placeholder="your@email.com" autocomplete="email" />
          </div>

          <Separator />

          <div class="grid gap-2">
            <Label for="password">新密码（留空则不修改）</Label>
            <Input id="password" v-model="form.password" type="password" placeholder="至少8位" autocomplete="new-password" />
          </div>
          <div class="grid gap-2">
            <Label for="confirmPassword">确认新密码</Label>
            <Input id="confirmPassword" v-model="form.confirmPassword" type="password" placeholder="再次输入新密码" autocomplete="new-password" />
          </div>

          <Button type="submit" class="w-fit" :disabled="saving">
            <Save class="mr-2 h-4 w-4" />
            {{ saving ? '保存中...' : '保存修改' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>