import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    proxy: {
'/api': { target: 'http://localhost:9090', changeOrigin: true },
	  '/health': { target: 'http://localhost:8080', changeOrigin: true },
	  // 订阅路径（带 / 后缀防止匹配 /src 等 Vite 内部路径）
	  '/s/': { target: 'http://localhost:9090', changeOrigin: true },
	  '/hello/': { target: 'http://localhost:9090', changeOrigin: true },
    },
  },
})