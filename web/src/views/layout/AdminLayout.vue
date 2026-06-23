<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  SidebarProvider,
  Sidebar,
  SidebarContent,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarFooter,
  SidebarInset,
  SidebarTrigger,
} from '@/components/ui/sidebar'
import { Separator } from '@/components/ui/separator'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  LayoutDashboard,
  Users,
  Package,
  Server,
  Settings,
  LogOut,
  ChevronsUpDown,
} from '@lucide/vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const navItems = [
  { name: '仪表盘', icon: LayoutDashboard, to: '/dashboard' },
  { name: '用户管理', icon: Users, to: '/users' },
  { name: '套餐管理', icon: Package, to: '/plans' },
  { name: '节点管理', icon: Server, to: '/nodes' },
  { name: '设置', icon: Settings, to: '/settings' },
]

function handleLogout() {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <SidebarProvider>
    <Sidebar variant="inset">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" as="div">
              <div class="bg-primary text-primary-foreground flex size-8 items-center justify-center rounded-lg">
                <Server class="size-4" />
              </div>
              <div class="grid flex-1 text-left text-sm leading-tight">
                <span class="truncate font-semibold">Nexus</span>
                <span class="truncate text-xs text-muted-foreground">管理后台</span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>导航菜单</SidebarGroupLabel>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in navItems" :key="item.to">
              <SidebarMenuButton
                as="a"
                :tooltip="item.name"
                :is-active="route.path === item.to"
                @click="router.push(item.to)"
              >
                <component :is="item.icon" />
                <span>{{ item.name }}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <SidebarMenuButton
                  size="lg"
                  class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                >
                  <Avatar class="h-8 w-8 rounded-lg">
                    <AvatarFallback class="rounded-lg">
                      {{ authStore.email.charAt(0).toUpperCase() }}
                    </AvatarFallback>
                  </Avatar>
                  <div class="grid flex-1 text-left text-sm leading-tight">
                    <span class="truncate font-semibold">{{ authStore.email }}</span>
                    <span class="truncate text-xs text-muted-foreground">管理员</span>
                  </div>
                  <ChevronsUpDown class="ml-auto size-4" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                side="top"
                class="w-(--reka-dropdown-menu-trigger-width)"
              >
                <DropdownMenuLabel>账号</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem @click="handleLogout">
                  <LogOut />
                  退出登录
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>

    <SidebarInset>
      <header class="flex h-12 shrink-0 items-center gap-2 border-b px-4">
        <SidebarTrigger class="-ml-1" />
        <Separator orientation="vertical" class="mr-2 h-4" />
        <h1 class="text-sm font-medium">
          {{ navItems.find(i => route.path === i.to)?.name || 'Nexus' }}
        </h1>
      </header>
      <main class="flex-1 p-4">
        <router-view />
      </main>
    </SidebarInset>
  </SidebarProvider>
</template>
