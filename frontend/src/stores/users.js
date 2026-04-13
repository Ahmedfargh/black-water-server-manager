import { defineStore } from 'pinia'
import api from '../api'

export const useUserStore = defineStore('users', {
  state: () => ({
    users: [],
    roles: [],
    loading: false
  }),
  actions: {
    async fetchUsers() {
      this.loading = true
      try {
        const response = await api.get('/users/crud/users/list')
        // Handle pagination data format 'data: [...]' or flat list
        this.users = response.data.data || response.data.users || response.data || []
      } catch (error) {
        console.error('Failed to fetch users:', error)
      } finally {
        this.loading = false
      }
    },
    async fetchRoles() {
      try {
        const response = await api.get('/users/roles')
        this.roles = response.data.roles || response.data || []
      } catch (error) {
        console.error('Failed to fetch roles:', error)
      }
    },
    async addUser(formData) {
      try {
        await api.post('/users/crud/users/', formData)
        await this.fetchUsers()
      } catch (error) {
        console.error('Failed to create user:', error)
        throw error
      }
    },
    async updateUser(id, formData) {
      try {
        await api.put(`/users/crud/users/${id}`, formData)
        await this.fetchUsers()
      } catch (error) {
        console.error('Failed to update user:', error)
        throw error
      }
    },
    async deleteUser(id) {
      try {
        await api.delete(`/users/crud/users/${id}`)
        await this.fetchUsers()
      } catch (error) {
        console.error('Failed to delete user:', error)
        throw error
      }
    }
  }
})
