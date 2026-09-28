import { defineStore } from 'pinia'
import api from '../api'

export const useSSHStore = defineStore('ssh', {
  state: () => ({
    keys: [],
    total: 0,
    page: 1,
    limit: 10,
    loading: false,
    generating: false,
    importing: false,
    error: null
  }),

  getters: {
    authorizedKeysCount: (state) => state.keys.filter(k => k.added_to_authorized_keys).length,
    ed25519Count: (state) => state.keys.filter(k => k.key_type?.toLowerCase().includes('ed25519')).length,
    rsaCount: (state) => state.keys.filter(k => k.key_type?.toLowerCase().includes('rsa')).length
  },

  actions: {
    async fetchKeys(page = 1, limit = 50) {
      this.loading = true
      this.error = null
      try {
        const response = await api.get(`/ssh/keys/list?page=${page}&limit=${limit}`)
        this.keys = response.data.keys || []
        this.total = response.data.total || 0
        this.page = response.data.page || page
        this.limit = response.data.limit || limit
        return response.data
      } catch (err) {
        this.error = err.response?.data?.error || err.message
        console.error('Failed to fetch SSH keys:', err)
        throw err
      } finally {
        this.loading = false
      }
    },

    async generateKey({ name, key_type, comment, add_to_authorized_keys, rsa_bits }) {
      this.generating = true
      this.error = null
      try {
        const response = await api.post('/ssh/keys/generate', {
          name,
          key_type,
          comment,
          add_to_authorized_keys,
          rsa_bits
        })
        await this.fetchKeys(this.page, this.limit)
        return response.data
      } catch (err) {
        this.error = err.response?.data?.error || err.message
        console.error('Failed to generate SSH key:', err)
        throw err
      } finally {
        this.generating = false
      }
    },

    async importKey({ name, public_key, comment, add_to_authorized_keys }) {
      this.importing = true
      this.error = null
      try {
        const response = await api.post('/ssh/keys/import', {
          name,
          public_key,
          comment,
          add_to_authorized_keys
        })
        await this.fetchKeys(this.page, this.limit)
        return response.data
      } catch (err) {
        this.error = err.response?.data?.error || err.message
        console.error('Failed to import SSH key:', err)
        throw err
      } finally {
        this.importing = false
      }
    },

    async toggleAuth(id) {
      try {
        const response = await api.post(`/ssh/keys/${id}/toggle-auth`)
        const idx = this.keys.findIndex(k => k.id === id)
        if (idx !== -1) {
          this.keys[idx].added_to_authorized_keys = response.data.added_to_authorized_keys
        }
        return response.data
      } catch (err) {
        console.error('Failed to toggle SSH key authorization:', err)
        throw err
      }
    },

    async deleteKey(id) {
      try {
        const response = await api.delete(`/ssh/keys/${id}`)
        this.keys = this.keys.filter(k => k.id !== id)
        this.total = Math.max(0, this.total - 1)
        return response.data
      } catch (err) {
        console.error('Failed to delete SSH key:', err)
        throw err
      }
    }
  }
})
