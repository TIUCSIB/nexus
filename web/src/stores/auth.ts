import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const email = ref(localStorage.getItem('email') || '')
  const isAdmin = ref(localStorage.getItem('is_admin') === 'true')

  async function login(emailAddr: string, password: string) {
    const res = await loginApi({ email: emailAddr, password })
    if (res.code === 0) {
      token.value = res.data.token
      email.value = emailAddr
      isAdmin.value = !!res.data.user?.is_admin
      localStorage.setItem('token', res.data.token)
      localStorage.setItem('email', emailAddr)
      localStorage.setItem('is_admin', String(isAdmin.value))

      // 从登录响应中直接获取路径配置，无需等待 getSiteInfo
      if (res.data.admin_path) localStorage.setItem('admin_path', res.data.admin_path)
      if (res.data.auth_path) localStorage.setItem('auth_path', res.data.auth_path)
      if (res.data.user_path) localStorage.setItem('user_path', res.data.user_path)
      if (res.data.app_name) { localStorage.setItem('app_name', res.data.app_name); document.title = res.data.app_name }
      if (res.data.app_description) localStorage.setItem('app_description', res.data.app_description)
      if (res.data.sub_path) localStorage.setItem('sub_path', res.data.sub_path)

      return true
    }
    return false
  }

  function logout() {
    token.value = ''
    email.value = ''
    isAdmin.value = false
    localStorage.removeItem('token')
    localStorage.removeItem('email')
    localStorage.removeItem('is_admin')
  }

  return { token, email, isAdmin, login, logout }
})