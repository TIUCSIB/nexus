<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSettingsStore } from '@/stores/settings'
import {
  SidebarProvider, Sidebar, SidebarContent, SidebarHeader,
  SidebarGroup, SidebarMenu, SidebarMenuItem, SidebarMenuButton, SidebarMenuSub, SidebarMenuSubItem, SidebarMenuSubButton,
  SidebarInset, SidebarTrigger
} from '@/components/ui/sidebar'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Separator } from '@/components/ui/separator'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { LayoutDashboard, Users, Package, Server, Shield, Route, Settings, LogOut, ChevronRight } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()
const settingsStore = useSettingsStore()
settingsStore.fetchSettings()
const sidebarOpen = ref(true)

const menuItems = computed(() => [
  { title: '仪表盘', icon: LayoutDashboard, path: settingsStore.adminRoute('dashboard') },
  { title: '用户管理', icon: Users, path: settingsStore.adminRoute('users') },
  { title: '套餐管理', icon: Package, path: settingsStore.adminRoute('plans') },
])

const nodeMenuItems = computed(() => [
  { title: '节点管理', icon: Server, path: settingsStore.adminRoute('nodes') },
  { title: '权限组管理', icon: Shield, path: settingsStore.adminRoute('groups') },
  { title: '路由管理', icon: Route, path: settingsStore.adminRoute('routes') },
])

const bottomMenuItems = computed(() => [
  { title: '系统设置', icon: Settings, path: settingsStore.adminRoute('settings') },
])

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
          {{ settingsStore.appName }}
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

        <SidebarGroup>
          <SidebarMenu>
            <Collapsible default-open>
              <SidebarMenuItem>
                <CollapsibleTrigger as-child>
                  <SidebarMenuButton>
                    <Server class="h-4 w-4" />
                    <span v-if="sidebarOpen">节点管理</span>
                    <ChevronRight v-if="sidebarOpen" class="ml-auto h-4 w-4 transition-transform group-data-[state=open]/collapsible:rotate-90" />
                  </SidebarMenuButton>
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <SidebarMenuSub>
                    <SidebarMenuSubItem v-for="item in nodeMenuItems" :key="item.path">
                      <SidebarMenuSubButton as-child>
                        <router-link :to="item.path" active-class="bg-accent text-accent-foreground">
                          <component :is="item.icon" class="h-4 w-4" />
                          <span>{{ item.title }}</span>
                        </router-link>
                      </SidebarMenuSubButton>
                    </SidebarMenuSubItem>
                  </SidebarMenuSub>
                </CollapsibleContent>
              </SidebarMenuItem>
            </Collapsible>
          </SidebarMenu>
        </SidebarGroup>

        <SidebarGroup>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in bottomMenuItems" :key="item.path">
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