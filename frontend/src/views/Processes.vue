<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { 
  Terminal, 
  Search, 
  XOctagon, 
  RefreshCcw, 
  User, 
  Cpu, 
  Zap,
  Activity
} from 'lucide-vue-next'
import { useProcessStore } from '../stores/processes'
import { useAuthStore } from '../stores/auth'

const processStore = useProcessStore()
const isAutoRefresh = ref(true)
let wsProcesses

const connectWebSockets = () => {
  const authStore = useAuthStore()
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws/processes?token=${authStore.token}`
  
  wsProcesses = new WebSocket(wsUrl)
  wsProcesses.onmessage = (event) => {
    if (isAutoRefresh.value) {
      try {
        const data = JSON.parse(event.data)
        processStore.updateProcesses(data)
      } catch (e) {
        console.error('Failed to parse process data:', e)
      }
    }
  }
}

onMounted(() => {
  processStore.fetchProcesses()
  connectWebSockets()
})

onUnmounted(() => {
  if (wsProcesses) wsProcesses.close()
})

const handleKill = async (pid) => {
  if (confirm(`INITIATE DISCONNECT: Are you sure you want to terminate process ${pid}?`)) {
    try {
      await processStore.killProcess(pid)
    } catch (err) {
      alert(`DISCONNECT FAILED: Unable to terminate process ${pid}.`)
    }
  }
}

const formatMemory = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<template>
  <div class="processes-view">
    <div class="controls-row">
      <div class="search-bar">
        <Search :size="18" class="search-icon" />
        <input 
          v-model="processStore.searchTerm" 
          type="text" 
          placeholder="SEARCH PROCESS... (PID, NAME, USER)"
        />
      </div>
      
      <div class="toggles">
        <label class="tron-switch">
          <input type="checkbox" v-model="isAutoRefresh">
          <span class="slider"></span>
          <span class="label">LIVE STREAM</span>
        </label>
        <button @click="processStore.fetchProcesses" class="refresh-btn">
          <RefreshCcw :size="18" />
        </button>
      </div>
    </div>

    <div class="tron-card process-table-container">
      <table class="process-table">
        <thead>
          <tr>
            <th>PID</th>
            <th>PROCESS NAME</th>
            <th>USER</th>
            <th>CPU %</th>
            <th>MEMORY</th>
            <th>THREADS</th>
            <th class="actions-col">ACTIONS</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in processStore.filteredProcesses" :key="p.pid" class="process-row">
            <td class="pid-cell font-data glow-cyan">{{ p.pid }}</td>
            <td class="name-cell">
              <div class="process-name-wrap">
                <Terminal :size="14" class="icon" />
                <span class="truncate-name" :title="p.name">{{ p.name || '[SYSTEM KERNEL]' }}</span>
              </div>
            </td>
            <td class="user-cell">
               <div class="user-wrap">
                 <User :size="14" class="icon" />
                 <span>{{ p.username }}</span>
               </div>
            </td>
            <td class="status-cell">
               <span class="status-tag" :title="p.status">{{ p.status }}</span>
            </td>
            <td class="cpu-cell">
               <div class="mini-bar">
                 <div class="bar-fill" :style="{ width: (p.cpu_percent || 0) + '%' }"></div>
                 <span class="val font-data">{{ p.cpu_percent?.toFixed(1) || '0.0' }}%</span>
               </div>
            </td>
            <td class="mem-cell font-data">{{ p.memory_usage ? formatMemory(p.memory_usage) : '0 B' }}</td>
            <td class="threads-cell font-data">{{ p.num_threads || '--' }}</td>
            <td class="actions-col">
              <button 
                @click="handleKill(p.pid)" 
                class="kill-btn" 
                title="TERMINATE PROCESS"
              >
                <XOctagon :size="18" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      
      <div v-if="processStore.loading" class="loading-overlay">
         <Activity class="pulse" :size="48" />
         <span>SYNCHRONIZING WITH GRID...</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.processes-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  height: 100%;
}

.controls-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 2rem;
}

.search-bar {
  flex: 1;
  max-width: 500px;
  position: relative;
  display: flex;
  align-items: center;
  background: var(--bg-card);
  border: 1px solid rgba(0, 242, 255, 0.2);
  padding: 0 1rem;
}

.search-icon {
  color: var(--text-secondary);
}

.search-bar input {
  background: transparent;
  border: none;
  padding: 0.8rem;
  color: var(--text-primary);
  font-family: var(--font-header);
  flex: 1;
  outline: none;
  letter-spacing: 1px;
}

.toggles {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.tron-switch {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  cursor: pointer;
  font-size: 0.8rem;
  letter-spacing: 1px;
}

.slider {
  width: 34px;
  height: 14px;
  background-color: var(--bg-black);
  border: 1px solid var(--text-secondary);
  position: relative;
  transition: .4s;
}

.slider:before {
  content: "";
  position: absolute;
  height: 18px;
  width: 10px;
  left: -2px;
  bottom: -3px;
  background-color: var(--text-secondary);
  transition: .4s;
  box-shadow: 0 0 5px rgba(0, 0, 0, 0.5);
}

input:checked + .slider {
  border-color: var(--neon-cyan);
}

input:checked + .slider:before {
  transform: translateX(28px);
  background-color: var(--neon-cyan);
  box-shadow: 0 0 10px var(--neon-cyan-glow);
}

input { display: none; }

.refresh-btn {
  background: transparent;
  border: 1px solid rgba(0, 242, 255, 0.2);
  color: var(--text-secondary);
  padding: 0.5rem;
  cursor: pointer;
}

.refresh-btn:hover { color: var(--neon-cyan); border-color: var(--neon-cyan); }

/* Table styles */
.process-table-container {
  flex: 1;
  position: relative;
  min-height: 400px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.process-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.process-table thead th {
  padding: 1.2rem;
  font-size: 0.8rem;
  color: var(--text-secondary);
  letter-spacing: 1.5px;
  border-bottom: 1px solid rgba(0, 242, 255, 0.1);
  position: sticky;
  top: 0;
  background: var(--bg-black);
  z-index: 10;
}

.process-table tbody {
  height: 100%;
  overflow-y: auto;
}

.process-row {
  border-bottom: 1px solid rgba(0, 242, 255, 0.05);
  transition: background 0.2s ease;
}

.process-row:hover {
  background: rgba(0, 242, 255, 0.03);
}

.process-table td {
  padding: 0.8rem 1.2rem;
  font-size: 0.9rem;
}

.pid-cell { font-weight: 600; }

.process-name-wrap, .user-wrap {
  display: flex;
  align-items: center;
  gap: 0.8rem;
}

.truncate-name {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-tag {
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-secondary);
  white-space: nowrap;
}

.status-tag[title*="sleeping"], .status-tag[title*="S"] {
  color: var(--neon-cyan);
  border-color: rgba(0, 242, 255, 0.2);
}

.status-tag[title*="running"], .status-tag[title*="R"] {
  color: #00ff00;
  border-color: rgba(0, 255, 0, 0.2);
  box-shadow: 0 0 5px rgba(0, 255, 0, 0.2);
}

.icon { opacity: 0.5; }

.cpu-cell .mini-bar {
  width: 100px;
  height: 6px;
  background: rgba(0, 242, 255, 0.05);
  position: relative;
  overflow: visible;
}

.bar-fill {
  height: 100%;
  background: var(--neon-cyan);
  box-shadow: 0 0 8px var(--neon-cyan-glow);
}

.cpu-cell .val {
  position: absolute;
  right: -45px;
  top: -5px;
  font-size: 0.75rem;
}

.actions-col { text-align: right; }

.kill-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.kill-btn:hover {
  color: var(--neon-orange);
  filter: drop-shadow(0 0 5px var(--neon-orange-glow));
  transform: scale(1.1);
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1.5rem;
  z-index: 20;
}
</style>
