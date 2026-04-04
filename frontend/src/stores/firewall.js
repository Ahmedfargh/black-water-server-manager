import { defineStore } from 'pinia'
import api from '../api'

export const useFirewallStore = defineStore('firewall', {
  state: () => ({
    status: 'unknown',
    rules: '', // Store as string since backend returns raw terminal output
    loading: false
  }),
  actions: {
    async fetchStatus() {
      try {
        const response = await api.get('/firewall/status')
        this.status = response.data.message || 'unknown'
      } catch (error) {
        console.error('Failed to fetch firewall status:', error)
      }
    },
    async fetchRules() {
      this.loading = true
      try {
        const response = await api.get('/firewall/rules')
        this.rules = response.data.message || ''
      } catch (error) {
        console.error('Failed to fetch firewall rules:', error)
      } finally {
        this.loading = false
      }
    },
    async toggleFirewall(enable) {
      try {
        const endpoint = enable ? '/firewall/enable' : '/firewall/disable'
        await api.get(endpoint)
        await this.fetchStatus()
      } catch (error) {
        console.error('Failed to toggle firewall:', error)
        throw error
      }
    }
  }
})
