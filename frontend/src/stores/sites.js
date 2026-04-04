import { defineStore } from 'pinia'
import api from '../api'

export const useSiteStore = defineStore('sites', {
  state: () => ({
    sites: [],
    loading: false
  }),
  actions: {
    async fetchSites() {
      this.loading = true
      try {
        const response = await api.get('/site/list')
        this.sites = response.data.sites || []
      } catch (error) {
        console.error('Failed to fetch sites:', error)
      } finally {
        this.loading = false
      }
    },
    async addSite(formData) {
      try {
        await api.post('/site/create', formData)
        await this.fetchSites()
      } catch (error) {
        console.error('Failed to create site:', error)
        throw error
      }
    },
    async updateSite(id, formData) {
      try {
        await api.put(`/site/update/${id}`, formData)
        await this.fetchSites()
      } catch (error) {
        console.error('Failed to update site:', error)
        throw error
      }
    },
    async triggerCheckup() {
      try {
        const response = await api.get('/site/full-checkup')
        const results = response.data.results
        
        // Update sites with latest checkup results
        if (results) {
          this.sites = this.sites.map(site => {
            const checkups = results[site.id]
            if (checkups && checkups.length > 0) {
              const latest = checkups[0]
              return {
                ...site,
                status: latest.status,
                last_checked: latest.time
              }
            }
            return site
          })
        }
      } catch (error) {
        console.error('Failed to trigger checkup:', error)
        throw error
      }
    }
  }
})
