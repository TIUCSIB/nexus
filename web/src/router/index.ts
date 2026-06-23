import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/Login.vue'),
    },
    {
      path: '/',
      component: () => import('@/views/layout/AdminLayout.vue'),
      redirect: '/dashboard',
      children: [
        { path: 'dashboard', name: 'Dashboard', component: () => import('@/views/dashboard/Dashboard.vue') },
        { path: 'users', name: 'Users', component: () => import('@/views/users/Users.vue') },
        { path: 'plans', name: 'Plans', component: () => import('@/views/plans/Plans.vue') },
        { path: 'nodes', name: 'Nodes', component: () => import('@/views/nodes/Nodes.vue') },
        { path: 'settings', name: 'Settings', component: () => import('@/views/settings/Settings.vue') },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  if (to.name !== 'Login' && !token) {
    return { name: 'Login' }
  }
})

export default router