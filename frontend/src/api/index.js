import axios from 'axios'
import { useAuthStore } from '../stores/auth'

const instance = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor for adding JWT token
instance.interceptors.request.use(
  (config) => {
    const authStore = useAuthStore()
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }
    
    // Debug logging for requests
    console.log(`[OUTGOING] ${config.method.toUpperCase()} ${config.url}`, config.data || '')
    
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for handling 401 Unauthorized
instance.interceptors.response.use(
  (response) => {
    // Debug logging for successful responses
    console.log(`[INCOMING] ${response.status} ${response.config.url}`, response.data)
    return response
  },
  (error) => {
    // Debug logging for errors
    if (error.response) {
      console.error(`[GRID ERROR] ${error.response.status} ${error.config.url}`, error.response.data)
    } else {
      console.error('[GRID ERROR] Network/Connection Failure', error.message)
    }

    if (error.response && error.response.status === 401) {
      const authStore = useAuthStore()
      authStore.logout()
    }
    return Promise.reject(error)
  }
)

export default instance
