import { defineStore } from 'pinia'
import api from '../api'

export const useDockerStore = defineStore('docker', {
  state: () => ({
    containers: [],
    loading: false,
    selectedContainer: null,
    logs: [],
    volumes: [],
    containerStats: {} // Store real-time stats by container ID
  }),
  actions: {
    async fetchContainers() {
      this.loading = true
      try {
        const response = await api.get('/docker/containers')
        this.containers = response.data
      } catch (error) {
        console.error('Failed to fetch containers:', error)
      } finally {
        this.loading = false
      }
    },
    async fetchVolumes(id) {
      this.loading = true
      try {
        const response = await api.get(`/docker/container/${id}/get/volums`)
        this.volumes = response.data.volumns || []
        return this.volumes
      } catch (error) {
        console.error('Failed to fetch volumes:', error)
        throw error
      } finally {
        this.loading = false
      }
    },
    async performAction(id, action) {
      try {
        await api.post(`/docker/container/${id}/${action}`)
        await this.fetchContainers() // Refresh list
      } catch (error) {
        console.error(`Failed to ${action} container:`, error)
        throw error
      }
    },
    async pruneContainer(id) {
      this.loading = true
      try {
        await api.get(`/docker/image/${id}/prune`)
        await this.fetchContainers() // Refresh list
      } catch (error) {
        console.error('Failed to prune container:', error)
        throw error
      } finally {
        this.loading = false
      }
    },
    updateContainerStats(id, stats) {
      this.containerStats[id] = stats
      
      // Also update status in containers list if possible
      // Note: ContainerStatus might only return stats, but we can infer 'Up' if we get stats
      const container = this.containers.find(c => c.id === id)
      if (container) {
        // If we're getting stats, the container is likely running
        if (!container.status.includes('Up')) {
          container.status = 'Up (Live)'
        }
      }
    }
  }
})
