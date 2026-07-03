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