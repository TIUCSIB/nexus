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
        { path: 'nodes', name: 'UserNodes', component: () => import('@/views/user/Nodes.vue') },
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
        { path: 'custom-outbounds', name: 'AdminCustomOutbounds', component: () => import('@/views/custom-outbounds/CustomOutbounds.vue') },
        { path: 'traffic-logs', name: 'AdminTrafficLogs', component: () => import('@/views/TrafficLogs.vue') },
        { path: 'online-ips', name: 'AdminOnlineIPs', component: () => import('@/views/OnlineIPs.vue') },
        { path: 'machines', name: 'AdminMachines', component: () => import('@/views/machines/Machines.vue') },
        { path: 'machines/:id', name: 'AdminMachineDetail', component: () => import('@/views/machines/MachineDetail.vue') },
        { path: 'settings', name: 'AdminSettings', component: () => import('@/views/settings/Settings.vue') },
        { path: 'traffic-reset', name: 'AdminTrafficReset', component: () => import('@/views/traffic-reset/TrafficReset.vue') },
        { path: 'audit-logs', name: 'AdminAuditLogs', component: () => import('@/views/audit-logs/AuditLogs.vue') },
        { path: 'node-ranking', name: 'AdminNodeRanking', component: () => import('@/views/stats/NodeRanking.vue') },
        { path: 'user-ranking', name: 'AdminUserRanking', component: () => import('@/views/stats/UserRanking.vue') },
      ],
    },
    {
      path: '/',
      name: 'Landing',
      meta: { public: true },
      component: () => import('@/views/landing/Landing.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      meta: { public: true },
      component: () => import('@/views/not-found/NotFound.vue'),
    },
  ],
})

const adminRouteNames = ['AdminDashboard', 'AdminUsers', 'AdminUserDetail', 'AdminPlans', 'AdminNodes', 'AdminGroups', 'AdminRoutes', 'AdminCustomOutbounds', 'AdminTrafficLogs', 'AdminOnlineIPs', 'AdminMachines', 'AdminMachineDetail', 'AdminSettings', 'AdminTrafficReset', 'AdminAuditLogs', 'AdminNodeRanking', 'AdminUserRanking']

router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  const adminPath = localStorage.getItem('admin_path') || 'admin'
  const isAdmin = localStorage.getItem('is_admin') === 'true'

  if (to.name === 'Login') return

  // 公开页面（落地页、404）无需登录
  if (to.meta.public) return

  // 未登录 → 跳转登录页
  if (!token) return { name: 'Login' }

  // 管理员访问用户页面 → 跳转后台
  if (isAdmin && to.path.startsWith('/user/')) {
    return '/' + adminPath + '/dashboard'
  }

  // 非管理员访问后台 → 跳转用户页面
  if (!isAdmin && typeof to.name === 'string' && adminRouteNames.includes(to.name)) {
    return '/user/dashboard'
  }

  // 验证后台路径前缀
  if (typeof to.name === 'string' && adminRouteNames.includes(to.name)) {
    const prefix = to.params.adminPrefix as string
    if (prefix !== adminPath) {
      return { name: 'NotFound' }
    }
  }
})

export default router