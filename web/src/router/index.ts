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
      path: '/user',
      component: () => import('@/views/layout/UserLayout.vue'),
      children: [
        { path: '', redirect: '/user/dashboard' },
        { path: 'dashboard', name: 'UserDashboard', component: () => import('@/views/user/Dashboard.vue') },
        { path: 'subscription', name: 'UserSubscription', component: () => import('@/views/user/Subscription.vue') },
        { path: 'profile', name: 'UserProfile', component: () => import('@/views/user/Profile.vue') },
      ],
    },
    {
      path: '/:adminPrefix',
      component: () => import('@/views/layout/AdminLayout.vue'),
      meta: { requiresAdmin: true },
      children: [
        { path: '', redirect: (to) => '/' + to.params.adminPrefix + '/dashboard' },
        { path: 'dashboard', name: 'AdminDashboard', component: () => import('@/views/dashboard/Dashboard.vue') },
        { path: 'users', name: 'AdminUsers', component: () => import('@/views/users/Users.vue') },
        { path: 'users/:id', name: 'AdminUserDetail', component: () => import('@/views/users/UserDetail.vue') },
        { path: 'plans', name: 'AdminPlans', component: () => import('@/views/plans/Plans.vue') },
        { path: 'nodes', name: 'AdminNodes', component: () => import('@/views/nodes/Nodes.vue') },
        { path: 'groups', name: 'AdminGroups', component: () => import('@/views/groups/Groups.vue') },
        { path: 'routes', name: 'AdminRoutes', component: () => import('@/views/routes/Routes.vue') },
        { path: 'settings', name: 'AdminSettings', component: () => import('@/views/settings/Settings.vue') },
      ],
    },
    {
      path: '/',
      redirect: '/login',
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/not-found/NotFound.vue'),
    },
  ],
})

const adminRouteNames = ['AdminDashboard', 'AdminUsers', 'AdminUserDetail', 'AdminPlans', 'AdminNodes', 'AdminGroups', 'AdminRoutes', 'AdminSettings']

router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  const adminPath = localStorage.getItem('admin_path') || 'admin'

  if (to.name === 'Login') return

  if (!token) return { name: 'Login' }

  if (typeof to.name === 'string' && adminRouteNames.includes(to.name)) {
    const prefix = to.params.adminPrefix as string
    if (prefix !== adminPath) {
      return { name: 'NotFound' }
    }
  }
})

export default router