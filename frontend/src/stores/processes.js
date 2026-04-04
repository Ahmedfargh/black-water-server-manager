import { defineStore } from 'pinia'
import api from '../api'

export const useProcessStore = defineStore('processes', {
  state: () => ({
    processes: [],
    loading: false,
    searchTerm: ''
  }),
  actions: {
    async fetchProcesses() {
      this.loading = true
      try {
        const response = await api.get('/processes')
        this.processes = this.mapProcesses(response.data)
      } catch (error) {
        console.error('Failed to fetch processes:', error)
      } finally {
        this.loading = false
      }
    },
    mapProcesses(data) {
      if (!Array.isArray(data)) return []
      return data.map(p => ({
        pid: p.PID,
        // Clean up null characters from /proc/cmdline and join with spaces
        name: (p.Name || '').replace(/\u0000/g, ' ').trim(),
        status: p.Status,
        username: 'SYSTEM', // Backend doesn't provide user yet
        cpu_percent: 0,
        memory_usage: 0,
        num_threads: 0
      }))
    },
    async killProcess(pid) {
      try {
        await api.delete(`/process/kill/${pid}`)
        this.processes = this.processes.filter(p => p.pid !== pid)
      } catch (error) {
        console.error('Failed to kill process:', error)
        throw error
      }
    },
    updateProcesses(newProcesses) {
      this.processes = this.mapProcesses(newProcesses)
    }
  },
  getters: {
    filteredProcesses: (state) => {
      const term = state.searchTerm.toLowerCase()
      return state.processes.filter(p => 
        p.name?.toLowerCase().includes(term) || 
        p.pid?.toString().includes(term) ||
        p.username?.toLowerCase().includes(term)
      )
    }
  }
})
