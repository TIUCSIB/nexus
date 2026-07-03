<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useSettingsStore } from '@/stores/settings'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const router = useRouter()
const settingsStore = useSettingsStore()
</script>

<template>
  <div class="flex min-h-svh items-center justify-center bg-muted/40 p-4">
    <Card class="w-full max-w-sm text-center">
      <CardHeader>
        <CardTitle class="text-2xl">404</CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-muted-foreground mb-4">页面不存在</p>
        <Button @click="router.push('/')">返回首页</Button>
      </CardContent>
    </Card>
  </div>
</template>