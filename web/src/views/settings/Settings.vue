<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { toast } from 'vue-sonner'
import { getSettings, updateSettings } from '@/api/settings'

const settings = ref<Record<string, string>>({})
const loading = ref(false)

onMounted(async () => {
  const res = await getSettings()
  if (res.code === 0) settings.value = res.data
})

async function handleSave() {
  loading.value = true
  try {
    const res = await updateSettings(settings.value)
    if (res.code === 0) toast.success('保存成功')
    else toast.error('保存失败: ' + res.message)
  } finally { loading.value = false }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">系统设置</h1>
      <Button @click="handleSave" :disabled="loading">{{ loading ? '保存中...' : '保存设置' }}</Button>
    </div>
    <Card>
      <CardHeader><CardTitle>基本设置</CardTitle></CardHeader>
      <CardContent class="space-y-4">
        <div class="grid gap-2" v-for="(_, key) in settings" :key="key">
          <Label>{{ key }}</Label>
          <Input v-model="settings[key]" />
        </div>
        <p v-if="Object.keys(settings).length === 0" class="text-muted-foreground">暂无系统设置</p>
      </CardContent>
    </Card>
  </div>
</template>