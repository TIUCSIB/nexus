<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Activity, Clock, HardDrive, CreditCard } from 'lucide-vue-next'
import { getProfile, getSubscription } from '@/api/userSelf'
import type { User, SubscriptionInfo } from '@/types'

const user = ref<User | null>(null)
const sub = ref<SubscriptionInfo | null>(null)
const loading = ref(true)

function formatBytes(bytes: number) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const trafficPercent = computed(() => {
  if (!sub.value || !sub.value.traffic_limit || sub.value.traffic_limit === 0) return 0
  return Math.min(100, Math.round((sub.value.traffic_used / sub.value.traffic_limit) * 100))
})

const trafficText = computed(() => {
  if (!sub.value) return '--'
  if (!sub.value.traffic_limit || sub.value.traffic_limit === 0) {
    return formatBytes(sub.value.traffic_used) + ' / 不限'
  }
  return formatBytes(sub.value.traffic_used) + ' / ' + formatBytes(sub.value.traffic_limit)
})

const expireText = computed(() => {
  if (!sub.value || !sub.value.expired_at) return '永久有效'
  const d = new Date(sub.value.expired_at)
  const now = new Date()
  const diff = Math.ceil((d.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  if (diff <= 0) return '已过期'
  return d.toLocaleDateString('zh-CN') + '（剩余 ' + diff + ' 天）'
})

const expireStatus = computed(() => {
  if (!sub.value || !sub.value.expired_at) return 'default'
  const d = new Date(sub.value.expired_at)
  const now = new Date()
  const diff = Math.ceil((d.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  if (diff <= 0) return 'destructive'
  if (diff <= 7) return 'destructive'
  if (diff <= 30) return 'secondary'
  return 'default'
})

onMounted(async () => {
  try {
    const [profileRes, subRes] = await Promise.all([getProfile(), getSubscription()])
    if (profileRes.code === 0) user.value = profileRes.data
    if (subRes.code === 0) sub.value = subRes.data
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold">仪表盘</h1>

    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">当前套餐</CardTitle>
          <CreditCard class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ sub?.plan_name || '未订购' }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">到期时间</CardTitle>
          <Clock class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-lg font-bold">{{ expireText }}</div>
          <Badge v-if="expireStatus !== 'default'" :variant="expireStatus as any" class="mt-1">
            {{ expireStatus === 'destructive' ? '即将到期' : '正常' }}
          </Badge>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">流量使用</CardTitle>
          <HardDrive class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-sm text-muted-foreground mb-2">{{ trafficText }}</div>
          <Progress :model-value="trafficPercent" class="h-2" />
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">连接状态</CardTitle>
          <Activity class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">
            <Badge :variant="user?.status === 1 ? 'default' : 'destructive'">
              {{ user?.status === 1 ? '正常' : '已禁用' }}
            </Badge>
          </div>
        </CardContent>
      </Card>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>账号信息</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="grid gap-3 md:grid-cols-2">
          <div>
            <span class="text-sm text-muted-foreground">邮箱</span>
            <p class="font-medium">{{ user?.email || '--' }}</p>
          </div>
          <div>
            <span class="text-sm text-muted-foreground">注册时间</span>
            <p class="font-medium">{{ user?.created_at ? new Date(user.created_at).toLocaleDateString('zh-CN') : '--' }}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>