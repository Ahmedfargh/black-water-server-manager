import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      // Proxy standard API requests, removing the /api prefix
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
      // Proxy WebSocket requests and inject Authorization header from query param
      '/ws': {
        target: 'ws://localhost:8080',
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
})
