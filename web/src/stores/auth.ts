import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as loginApi } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const email = ref(localStorage.getItem('email') || '')

  async function login(emailAddr: string, password: string) {
    const res = await loginApi({ email: emailAddr, password })
    if (res.code === 0) {
      token.value = res.data.token
      email.value = emailAddr
      localStorage.setItem('token', res.data.token)
      localStorage.setItem('email', emailAddr)
      return true
    }
    return false
  }

  function logout() {
    token.value = ''
    email.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('email')
  }

  return { token, email, login, logout }
})