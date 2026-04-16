import { defineStore } from 'pinia'
import api from '../api'

export const useSystemStore = defineStore('system', {
  state: () => ({
    cpu: { usage: 0, model: '', cores: 0, temp: 0 },
    ram: { total: 0, used: 0, free: 0, usage: 0 },
    disk: { total: 0, used: 0, free: 0, usage: 0 },
    network: { bytesSent: 0, bytesRecv: 0 },
    history: {
      cpu: [],
      ram: [],
      network: []
    },
    reports: [],
    averages: {
      cpu: 0,
      memory: 0,
      disk: 0
    }
  }),
  actions: {
    async fetchAllStats() {
      try {
        const [cpuRes, ramRes, diskRes, netRes, reportRes] = await Promise.all([
          api.get('/cpu'),
          api.get('/ram'),
          api.get('/disk'),
          api.get('/network'),
          api.get('/report')
        ])

        // Update real usages from report
        const report = reportRes.data.report
        if (report) {
          this.cpu.usage = parseFloat(report.cpu_usage.toFixed(1))
          this.ram.usage = parseFloat(report.memory_usage.toFixed(1))
          this.disk.usage = parseFloat(report.disk_usage.toFixed(1))
        }

        // CPU specs
        this.cpu = { 
          ...this.cpu, 
          model: cpuRes.data.Cpu_Hard_Ware_Info[0]?.model || 'Generic CPU',
          cores: cpuRes.data.Logical_core 
        }

        // RAM specs
        const ramData = ramRes.data.Vertiual_info
        if (ramData) {
          this.ram = {
            ...this.ram,
            total: ramData.Total_memory * 1024 * 1024,
            used: ramData.Used_memory * 1024 * 1024,
            free: ramData.Free_memory * 1024 * 1024
          }
        }

        // Disk specs
        const primaryDisk = diskRes.data.Disks?.[0]
        if (primaryDisk) {
          this.disk = {
            ...this.disk,
            total: primaryDisk.TotalGB * 1024 * 1024 * 1024,
            used: primaryDisk.UsedGB * 1024 * 1024 * 1024,
            free: primaryDisk.FreeGB * 1024 * 1024 * 1024
          }
        }

        // Network
        const netData = netRes.data.network?.[0]
        if (netData) {
          this.network = {
            bytesSent: netData.bytesSent,
            bytesRecv: netData.bytesRecv
          }
        }

        this.updateHistory()
      } catch (error) {
        console.error('Failed to fetch system stats:', error)
      }
    },
    updateHistory() {
      const now = new Date().toLocaleTimeString()
      
      this.history.cpu.push({ time: now, value: this.cpu.usage })
      this.history.ram.push({ time: now, value: this.ram.usage })
      
      // Keep only last 20 points
      if (this.history.cpu.length > 20) {
        this.history.cpu.shift()
        this.history.ram.shift()
      }
    },
    updateCpuTemp(temp) {
      this.cpu.temp = temp
    },
    async fetchHistoryReports(startTime, endTime) {
      try {
        const res = await api.post('/hardware-report/by-time-range', {
          start: this.formatDate(startTime),
          end: this.formatDate(endTime)
        })
        this.reports = res.data.reports || []
        return this.reports
      } catch (error) {
        console.error('Failed to fetch history reports:', error)
        return []
      }
    },
    async fetchAverageReports(startTime, endTime) {
      try {
        const res = await api.post('/hardware-report/average-usage-by-time-range', {
          start: this.formatDate(startTime),
          end: this.formatDate(endTime)
        })
        const data = res.data
        this.averages = {
          cpu: data.average_cpu_usage || 0,
          memory: data.average_memory_usage || 0,
          disk: data.average_disk_usage || 0
        }
        return this.averages
      } catch (error) {
        console.error('Failed to fetch average reports:', error)
        return null
      }
    },
    formatDate(dateStr) {
      const date = new Date(dateStr)
      const pad = (n) => n.toString().padStart(2, '0')
      const ms = date.getMilliseconds().toString().padStart(3, '0')
      return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}.${ms}`
    }
  }
})
