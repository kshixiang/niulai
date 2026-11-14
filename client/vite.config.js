import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig(({ mode }) => {
  // 代理目标可通过环境变量覆盖：默认连本机 go run .，
  // 联调真实服务器时在 shell 临时设置 VITE_PROXY_TARGET，不要把地址写进仓库。
  const env = loadEnv(mode, process.cwd(), '')
  const proxyTarget = env.VITE_PROXY_TARGET || 'http://localhost:8080'

  return {
    plugins: [vue()],
    resolve: {
      alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
    },
    server: {
      host: true,
      port: 5173,
      proxy: { '/api': { target: proxyTarget, changeOrigin: true } },
    },
  }
})
