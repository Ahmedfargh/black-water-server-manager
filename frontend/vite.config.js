import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const backendTarget = env.VITE_BACKEND_URL || 'http://localhost:8080'
  const wsTarget = backendTarget.replace(/^http/, 'ws')
  const port = parseInt(env.VITE_PORT, 10) || 5173

  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      port,
      host: true,
      proxy: {
        // Proxy standard API requests, removing the /api prefix
        '/api': {
          target: backendTarget,
          changeOrigin: true,
          rewrite: (p) => p.replace(/^\/api/, ''),
        },
        // Proxy WebSocket requests and inject Authorization header from query param
        '/ws': {
          target: wsTarget,
          ws: true,
          changeOrigin: true,
          configure: (proxy, _options) => {
            proxy.on('proxyReqWs', (proxyReq, req, _res, _options) => {
              // Extract token from query parameter
              const url = new URL(req.url, `http://${req.headers.host}`)
              const token = url.searchParams.get('token')
              if (token) {
                // Inject the Authorization header for the backend's AuthMiddleware
                proxyReq.setHeader('Authorization', `Bearer ${token}`)
              }
            })
          },
        },
      },
    },
  }
})
