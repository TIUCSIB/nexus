<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Copy, Check, Link2, Zap } from 'lucide-vue-next'
import { getSubscription } from '@/api/userSelf'
import { toast } from 'vue-sonner'
import type { SubscriptionInfo } from '@/types'

const sub = ref<SubscriptionInfo | null>(null)
const loading = ref(true)
const copied = ref(false)
const copiedClean = ref(false)
const selectedFormat = ref('clash')

const formats = [
  { value: 'clash', label: 'Clash / Clash.Meta', index: 1 },
  { value: 'singbox', label: 'sing-box', index: 0 },
  { value: 'surge', label: 'Surge', index: 2 },
  { value: 'surfboard', label: 'Surfboard', index: 3 },
  { value: 'shadowrocket', label: 'Shadowrocket', index: 4 },
  { value: 'v2rayn', label: 'V2RayN', index: 5 },
]

const currentLink = computed(() => {
  if (!sub.value || !sub.value.links) return ''
  const fmt = formats.find(f => f.value === selectedFormat.value)
  if (!fmt) return ''
  return sub.value.links[fmt.index] || ''
})

const cleanLink = computed(() => {
  if (!sub.value || !sub.value.clean_links) return ''
  return sub.value.clean_links[0] || ''
})

async function copyLink() {
  if (!currentLink.value) return
  try {
    await navigator.clipboard.writeText(currentLink.value)
    copied.value = true
    toast.success('订阅链接已复制到剪贴板')
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    toast.error('复制失败，请手动复制')
  }
}

async function copyCleanLink() {
  if (!cleanLink.value) return
  try {
    await navigator.clipboard.writeText(cleanLink.value)
    copiedClean.value = true
    toast.success('订阅链接已复制到剪贴板')
    setTimeout(() => { copiedClean.value = false }, 2000)
  } catch {
    toast.error('复制失败，请手动复制')
  }
}

onMounted(async () => {
  try {
    const res = await getSubscription()
    if (res.code === 0) sub.value = res.data
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold">订阅管理</h1>

    <Card v-if="cleanLink">
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <Zap class="h-5 w-5" />
          快捷订阅链接
        </CardTitle>
        <CardDescription>
          将下方链接直接导入 sing-box 客户端即可使用，无需手动选择格式
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div class="flex gap-2">
          <Input
            :model-value="cleanLink"
            readonly
            class="flex-1 font-mono text-sm"
          />
          <Button @click="copyCleanLink" variant="outline" size="icon">
            <Check v-if="copiedClean" class="h-4 w-4 text-green-600" />
            <Copy v-else class="h-4 w-4" />
          </Button>
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <Link2 class="h-5 w-5" />
          订阅链接
        </CardTitle>
        <CardDescription>选择你的客户端类型，复制对应的订阅链接导入即可使用</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <Tabs v-model="selectedFormat">
          <TabsList class="grid w-full grid-cols-3 lg:grid-cols-6">
            <TabsTrigger v-for="fmt in formats" :key="fmt.value" :value="fmt.value">
              {{ fmt.label }}
            </TabsTrigger>
          </TabsList>
        </Tabs>

        <div class="flex gap-2">
          <Input
            :model-value="currentLink"
            readonly
            class="flex-1 font-mono text-sm"
            placeholder="暂无订阅链接"
          />
          <Button @click="copyLink" variant="outline" size="icon" :disabled="!currentLink">
            <Check v-if="copied" class="h-4 w-4 text-green-600" />
            <Copy v-else class="h-4 w-4" />
          </Button>
        </div>

        <div v-if="sub" class="flex flex-wrap gap-2 pt-2">
          <Badge variant="outline">套餐：{{ sub.plan_name || '未订阅' }}</Badge>
          <Badge variant="outline">到期：{{ sub.expired_at ? new Date(sub.expired_at).toLocaleDateString('zh-CN') : '永久' }}</Badge>
        </div>
      </CardContent>
    </Card>
  </div>
</template>