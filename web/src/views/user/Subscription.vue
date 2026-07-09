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
const loadError = ref('')

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

// 到期预警：7天内到期
const expiryWarning = computed(() => {
  if (!sub.value?.expired_at) return null
  const expireDate = new Date(sub.value.expired_at)
  const now = new Date()
  const daysLeft = Math.ceil((expireDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  if (daysLeft < 0) return { type: 'expired', message: '套餐已过期', days: daysLeft }
  if (daysLeft <= 3) return { type: 'critical', message: `套餐将在 ${daysLeft} 天后到期`, days: daysLeft }
  if (daysLeft <= 7) return { type: 'warning', message: `套餐将在 ${daysLeft} 天后到期`, days: daysLeft }
  return null
})

// 流量预警：超过 80%
const trafficWarning = computed(() => {
  if (!sub.value || sub.value.traffic_limit <= 0) return null
  const ratio = sub.value.traffic_used / sub.value.traffic_limit
  const percent = Math.round(ratio * 100)
  const remaining = sub.value.traffic_limit - sub.value.traffic_used
  if (ratio >= 1) return { type: 'exhausted', message: '流量已用完', percent, remaining }
  if (ratio >= 0.9) return { type: 'critical', message: `流量已使用 ${percent}%，剩余 ${formatBytes(remaining)}`, percent, remaining }
  if (ratio >= 0.8) return { type: 'warning', message: `流量已使用 ${percent}%，剩余 ${formatBytes(remaining)}`, percent, remaining }
  return null
})

const trafficPercent = computed(() => {
  if (!sub.value || sub.value.traffic_limit <= 0) return 0
  return Math.min(100, Math.round((sub.value.traffic_used / sub.value.traffic_limit) * 100))
})

function formatBytes(bytes: number) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(Math.abs(bytes)) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

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
    if (res.code === 0) {
      sub.value = res.data
    } else {
      loadError.value = res.message || '获取订阅信息失败'
      toast.error(loadError.value)
    }
  } catch (error: any) {
    loadError.value = error?.response?.data?.message || error?.response?.data?.error || '获取订阅信息失败'
    toast.error(loadError.value)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold">订阅管理</h1>

    <Card v-if="sub?.available === false" class="border-orange-200 bg-orange-50">
      <CardContent class="pt-6 text-sm text-orange-700">
        当前订阅不可用：{{ sub.unavailable_reason || '订阅不可用' }}
      </CardContent>
    </Card>

    <Card v-else-if="loadError" class="border-red-200 bg-red-50">
      <CardContent class="pt-6 text-sm text-red-700">
        {{ loadError }}
      </CardContent>
    </Card>

    <!-- 到期预警 -->
    <Card v-if="expiryWarning" :class="expiryWarning.type === 'expired' ? 'border-red-300 bg-red-50' : expiryWarning.type === 'critical' ? 'border-red-200 bg-red-50' : 'border-yellow-200 bg-yellow-50'">
      <CardContent class="pt-6 text-sm" :class="expiryWarning.type === 'expired' ? 'text-red-700' : expiryWarning.type === 'critical' ? 'text-red-600' : 'text-yellow-700'">
        ⚠️ {{ expiryWarning.message }}
      </CardContent>
    </Card>

    <!-- 流量预警 -->
    <Card v-if="trafficWarning" :class="trafficWarning.type === 'exhausted' ? 'border-red-300 bg-red-50' : trafficWarning.type === 'critical' ? 'border-red-200 bg-red-50' : 'border-yellow-200 bg-yellow-50'">
      <CardContent class="pt-6 text-sm" :class="trafficWarning.type === 'exhausted' ? 'text-red-700' : trafficWarning.type === 'critical' ? 'text-red-600' : 'text-yellow-700'">
        ⚠️ {{ trafficWarning.message }}
      </CardContent>
    </Card>

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

        <div v-if="sub" class="space-y-3 pt-2">
          <div class="flex flex-wrap gap-2">
            <Badge variant="outline">套餐：{{ sub.plan_name || '未订阅' }}</Badge>
            <Badge variant="outline">到期：{{ sub.expired_at ? new Date(sub.expired_at).toLocaleDateString('zh-CN') : '永久' }}</Badge>
          </div>
          <div v-if="sub.traffic_limit > 0" class="space-y-1">
            <div class="flex justify-between text-sm text-muted-foreground">
              <span>流量使用</span>
              <span>{{ formatBytes(sub.traffic_used) }} / {{ formatBytes(sub.traffic_limit) }}</span>
            </div>
            <div class="w-full bg-muted rounded-full h-2">
              <div
                class="h-2 rounded-full transition-all"
                :class="trafficPercent >= 90 ? 'bg-red-500' : trafficPercent >= 80 ? 'bg-yellow-500' : 'bg-green-500'"
                :style="{ width: trafficPercent + '%' }"
              />
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>