import { defineStore } from 'pinia'
import api from '../api'

export const useAuditStore = defineStore('audit', {
  state: () => ({
    logs: [],
    loading: false,
    page: 1,
    limit: 20,
    type: ''
  }),
  actions: {
    async fetchLogs() {
      this.loading = true
      try {
        const response = await api.get('/audit/list', {
          params: {
            page: this.page,
            limit: this.limit,
            type: this.type
          }
        })
        this.logs = response.data.logs || []
      } catch (error) {
        console.error('Failed to fetch audit logs:', error)
      } finally {
        this.loading = false
      }
    },
    setFilters(filters) {
      if (filters.page) this.page = filters.page
      if (filters.limit) this.limit = filters.limit
      if (filters.type !== undefined) this.type = filters.type
      this.fetchLogs()
    }
  }
})
