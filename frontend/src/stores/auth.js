import { defineStore } from 'pinia'
import api from '../api'

const safeParse = (key) => {
  const item = localStorage.getItem(key)
  if (!item || item === 'undefined') return null
  try {
    return JSON.parse(item)
  } catch (e) {
    return null
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: safeParse('user'),
    token: localStorage.getItem('token') || null,
    isAuthenticated: !!localStorage.getItem('token')
  }),
  actions: {
    async login(email, password) {
      try {
        const response = await api.post('/login', { email, password })
        if (response.data.token) {
          this.token = response.data.token
          this.user = response.data.user
          this.isAuthenticated = true
          localStorage.setItem('token', this.token)
          localStorage.setItem('user', JSON.stringify(this.user))
        }
        return response.data
      } catch (error) {
        throw error
      }
    },
    async verifyOTP(userId, code) {
      try {
        const response = await api.post('/verify-otp', { user_id: Number(userId), code })
        if (response.data.token) {
          this.token = response.data.token
          this.user = response.data.user
          this.isAuthenticated = true
          localStorage.setItem('token', this.token)
          localStorage.setItem('user', JSON.stringify(this.user))
        }
        return response.data
      } catch (error) {
        throw error
      }
    },
    async resendOTP(userId) {
      try {
        const response = await api.post('/resend-otp', { user_id: Number(userId) })
        return response.data
      } catch (error) {
        throw error
      }
    },
    async verifyEmail(userId, code) {
      try {
        const response = await api.post('/verify-email', { user_id: Number(userId), code })
        return response.data
      } catch (error) {
        throw error
      }
    },
    async resendEmail(userId) {
      try {
        const response = await api.post('/resend-verification', { user_id: Number(userId) })
        return response.data
      } catch (error) {
        throw error
      }
    },
    logout() {
      this.user = null
      this.token = null
      this.isAuthenticated = false
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    },
    async fetchProfile() {
      try {
        const response = await api.get('/users/profile/me')
        this.user = { ...this.user, ...response.data.user }
        localStorage.setItem('user', JSON.stringify(this.user))
        return response.data.user
      } catch (error) {
        throw error
      }
    },
    async updateProfile(formData) {
      try {
        const response = await api.post('/users/users/acount/update', formData)
        this.user = { ...this.user, ...response.data }
        localStorage.setItem('user', JSON.stringify(this.user))
        return response.data
      } catch (error) {
        throw error
      }
    },
    async updateNotifications(data) {
      try {
        const response = await api.post('/users/users/notifications/settings', data)
        return response.data
      } catch (error) {
        throw error
      }
    }
  }
})
