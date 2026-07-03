import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getSettings } from '@/api/settings'

export const useSettingsStore = defineStore('settings', () => {
  const appName = ref(localStorage.getItem('app_name') || 'Nexus')
  const appDescription = ref(localStorage.getItem('app_description') || '')
  const adminPath = ref(localStorage.getItem('admin_path') || 'admin')
  const subUrl = ref(localStorage.getItem('sub_url') || '')
  const subPath = ref(localStorage.getItem('sub_path') || 's')

  const adminBase = computed(() => '/' + adminPath.value)

  async function fetchSettings() {
    try {
      const res = await getSettings()
      if (res.code === 0 && res.data) {
        if (res.data.app_name) {
          appName.value = res.data.app_name
          localStorage.setItem('app_name', res.data.app_name)
          document.title = res.data.app_name
        }
        if (res.data.app_description) {
          appDescription.value = res.data.app_description
        }
        if (res.data.admin_path) {
          adminPath.value = res.data.admin_path
          localStorage.setItem('admin_path', res.data.admin_path)
        }
        if (res.data.sub_url) {
          subUrl.value = res.data.sub_url
          localStorage.setItem('sub_url', res.data.sub_url)
        }
        if (res.data.sub_path) {
          subPath.value = res.data.sub_path
          localStorage.setItem('sub_path', res.data.sub_path)
        }
      }
    } catch {
      // ignore
    }
  }

  function setAppName(name: string) {
    appName.value = name
    localStorage.setItem('app_name', name)
    document.title = name
  }

  function setAppDescription(desc: string) {
    appDescription.value = desc
    localStorage.setItem('app_description', desc)
  }

  function setAdminPath(path: string) {
    adminPath.value = path || 'admin'
    localStorage.setItem('admin_path', path || 'admin')
  }

  function setSubUrl(url: string) {
    subUrl.value = url
    localStorage.setItem('sub_url', url)
  }

  function adminRoute(sub: string) {
    return '/' + adminPath.value + '/' + sub
  }

  function getRandomSubUrl(): string {
    if (!subUrl.value) return window.location.origin
    const urls = subUrl.value.split(',').map((s: string) => s.trim()).filter(Boolean)
    if (urls.length === 0) return window.location.origin
    return urls[Math.floor(Math.random() * urls.length)]
  }

  function buildSubUrl(token: string): string {
    const base = getRandomSubUrl()
    return base + '/' + subPath.value + '/' + token
  }

  return { appName, appDescription, adminPath, adminBase, subUrl, subPath, fetchSettings, setAppName, setAppDescription, setAdminPath, setSubUrl, adminRoute, getRandomSubUrl, buildSubUrl }
})