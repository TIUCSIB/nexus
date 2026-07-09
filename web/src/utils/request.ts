import axios from 'axios'

const request = axios.create({
  baseURL: '',
  timeout: 10000,
})

request.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }

  // 路径伪装：根据 localStorage 中的配置替换 API 路径前缀
  if (config.url) {
    const adminPath = localStorage.getItem('admin_path') || 'admin'
    const authPath = localStorage.getItem('auth_path') || 'auth'
    const userPath = localStorage.getItem('user_path') || 'user'
    config.url = config.url
      .replace('/api/admin/', `/api/${adminPath}/`)
      .replace('/api/auth/', `/api/${authPath}/`)
      .replace('/api/user/', `/api/${userPath}/`)
  }

  return config
})

request.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401 && !window.location.pathname.startsWith('/login')) {
      localStorage.removeItem('token')
      localStorage.removeItem('email')
      localStorage.removeItem('is_admin')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default request
