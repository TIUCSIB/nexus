<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  SidebarProvider, Sidebar, SidebarContent, SidebarHeader,
  SidebarGroup, SidebarMenu, SidebarMenuItem, SidebarMenuButton,
  SidebarInset, SidebarTrigger
} from '@/components/ui/sidebar'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Separator } from '@/components/ui/separator'
import { LayoutDashboard, Link2, User, LogOut, Server, Wifi } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()
const sidebarOpen = ref(true)

const menuItems = [
  { title: '仪表盘', icon: LayoutDashboard, path: '/user/dashboard' },
  { title: '节点列表', icon: Wifi, path: '/user/nodes' },
  { title: '订阅管理', icon: Link2, path: '/user/subscription' },
  { title: '个人资料', icon: User, path: '/user/profile' },
]

function logout() {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <SidebarProvider v-model:open="sidebarOpen">
    <Sidebar collapsible="icon">
      <SidebarHeader class="border-b p-4">
        <div class="flex items-center gap-2 text-lg font-bold" v-if="sidebarOpen">
          <Server class="h-5 w-5" />
          Nexus
        </div>
        <Server v-else class="h-5 w-5 mx-auto" />
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in menuItems" :key="item.path">
              <SidebarMenuButton as-child>
                <router-link :to="item.path" active-class="bg-accent text-accent-foreground">
                  <component :is="item.icon" class="h-4 w-4" />
                  <span v-if="sidebarOpen">{{ item.title }}</span>
                </router-link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>

    <SidebarInset>
      <header class="flex h-14 items-center gap-2 border-b px-4">
        <SidebarTrigger />
        <Separator orientation="vertical" class="h-6" />
        <div class="flex-1" />
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button variant="ghost" size="sm">{{ authStore.email }}</Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem @click="logout">
              <LogOut class="mr-2 h-4 w-4" />
              退出登录
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </header>
      <main class="flex-1 p-6">
        <router-view />
      </main>
    </SidebarInset>
  </SidebarProvider>
</template>