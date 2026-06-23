<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSidebar } from '@/components/ui/sidebar'
import {
  SidebarProvider, Sidebar, SidebarContent, SidebarHeader, SidebarFooter,
  SidebarGroup, SidebarGroupLabel, SidebarMenu, SidebarMenuItem, SidebarMenuButton,
  SidebarInset, SidebarTrigger
} from '@/components/ui/sidebar'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Separator } from '@/components/ui/separator'
import { LayoutDashboard, Users, Package, Server, Settings, LogOut, PanelLeft } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()
const { state } = useSidebar()

const menuItems = [
  { title: '仪表盘', icon: LayoutDashboard, path: '/dashboard' },
  { title: '用户管理', icon: Users, path: '/users' },
  { title: '套餐管理', icon: Package, path: '/plans' },
  { title: '节点管理', icon: Server, path: '/nodes' },
  { title: '系统设置', icon: Settings, path: '/settings' },
]

function logout() {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <SidebarProvider>
    <Sidebar collapsible="icon">
      <SidebarHeader class="border-b p-4">
        <div class="flex items-center gap-2 text-lg font-bold" v-if="state === 'expanded'">
          <Server class="h-5 w-5" />
          Nexus
        </div>
        <Server v-else class="h-5 w-5 mx-auto" />
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel v-if="state === 'expanded'">导航菜单</SidebarGroupLabel>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in menuItems" :key="item.path">
              <SidebarMenuButton as-child>
                <router-link :to="item.path" active-class="bg-accent text-accent-foreground">
                  <component :is="item.icon" class="h-4 w-4" />
                  <span v-if="state === 'expanded'">{{ item.title }}</span>
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