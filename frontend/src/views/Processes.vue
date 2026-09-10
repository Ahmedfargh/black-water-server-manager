<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
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

const { t } = useI18n()
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
  if (confirm(t('proc.confirm_kill', { pid }) || `INITIATE DISCONNECT: Are you sure you want to terminate process ${pid}?`)) {
    try {
      await processStore.killProcess(pid)
    } catch (err) {
      alert(t('proc.kill_failed') || `DISCONNECT FAILED: Unable to terminate process ${pid}.`)
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
          :placeholder="t('proc.search_placeholder') || 'SEARCH PROCESS... (PID, NAME, USER)'"
        />
      </div>
      
      <div class="toggles">
        <label class="tron-switch">
          <input type="checkbox" v-model="isAutoRefresh">
          <span class="slider"></span>
          <span class="label">{{ $t('proc.live_stream') }}</span>
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
            <th>{{ $t('proc.pid') }}</th>
            <th>{{ $t('proc.name') }}</th>
            <th>{{ $t('proc.user') }}</th>
            <th>{{ $t('proc.cpu') || 'CPU %' }}</th>
            <th>{{ $t('proc.mem') }}</th>
            <th>{{ $t('proc.threads') }}</th>
            <th class="actions-col">{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in processStore.filteredProcesses" :key="p.pid" class="process-row">
            <td class="pid-cell font-data glow-cyan">{{ p.pid }}</td>
            <td class="name-cell">
              <div class="process-name-wrap">
                <Terminal :size="14" class="icon" />
                <span class="truncate-name" :title="p.name">{{ p.name || ($t('proc.kernel') || '[SYSTEM KERNEL]') }}</span>
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
                :title="t('proc.terminate') || 'TERMINATE PROCESS'"
              >
                <XOctagon :size="18" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      
      <div v-if="processStore.loading" class="loading-overlay">
         <Activity class="pulse" :size="48" />
         <span>{{ $t('proc.syncing') || 'SYNCHRONIZING WITH GRID...' }}</span>
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
  gap: 1.5rem;
}

.search-bar {
  flex: 1;
  max-width: 500px;
  position: relative;
  display: flex;
  align-items: center;
  background: #11131a;
  border: 1px solid var(--border-subtle);
  border-radius: 3px;
  padding: 0 0.85rem;
  transition: border-color 0.15s ease;
}

.search-bar:focus-within {
  border-color: var(--rdr-crimson);
}

.search-icon {
  color: var(--text-muted);
}

.search-bar input {
  background: transparent;
  border: none;
  padding: 0.65rem 0.85rem;
  color: var(--text-primary);
  font-family: var(--font-data);
  font-size: 0.85rem;
  flex: 1;
  outline: none;
}

.toggles {
  display: flex;
  align-items: center;
  gap: 1.2rem;
}

.tron-switch {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  cursor: pointer;
  font-size: 0.78rem;
  color: var(--text-secondary);
}

.slider {
  width: 32px;
  height: 16px;
  background-color: #11131a;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  position: relative;
  transition: 0.2s;
}

.slider:before {
  content: "";
  position: absolute;
  height: 10px;
  width: 10px;
  left: 2px;
  bottom: 2px;
  background-color: var(--text-muted);
  border-radius: 50%;
  transition: 0.2s;
}

input:checked + .slider {
  border-color: var(--rdr-crimson);
  background-color: rgba(220, 38, 38, 0.2);
}

input:checked + .slider:before {
  transform: translateX(16px);
  background-color: var(--rdr-crimson);
}

input { display: none; }

.refresh-btn {
  background: #161822;
  border: 1px solid var(--border-subtle);
  color: var(--text-secondary);
  border-radius: 3px;
  padding: 0.45rem;
  cursor: pointer;
  transition: all 0.15s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.refresh-btn:hover { 
  color: var(--rdr-crimson); 
  border-color: var(--rdr-crimson); 
}

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

[dir="rtl"] .process-table {
  text-align: right;
}

.process-table thead th {
  padding: 0.85rem 1rem;
  font-size: 0.75rem;
  color: var(--text-muted);
  font-family: var(--font-data);
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--border-subtle);
  position: sticky;
  top: 0;
  background: #101218;
  z-index: 10;
}

.process-table tbody {
  height: 100%;
  overflow-y: auto;
}

.process-row {
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  transition: background 0.15s ease;
}

.process-row:hover {
  background: rgba(220, 38, 38, 0.03);
}

.process-table td {
  padding: 0.65rem 1rem;
  font-size: 0.84rem;
  font-family: var(--font-data);
}

.pid-cell { 
  font-weight: 600; 
  color: var(--rdr-amber);
}

.process-name-wrap, .user-wrap {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.truncate-name {
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-primary);
}

.status-tag {
  font-size: 0.7rem;
  padding: 0.15rem 0.45rem;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border-subtle);
  color: var(--text-muted);
  border-radius: 2px;
  white-space: nowrap;
  font-family: var(--font-data);
}

.status-tag[title*="sleeping"], .status-tag[title*="S"] {
  color: var(--text-muted);
  border-color: var(--border-subtle);
}

.status-tag[title*="running"], .status-tag[title*="R"] {
  color: var(--linux-green);
  border-color: rgba(16, 185, 129, 0.35);
  background: rgba(16, 185, 129, 0.1);
}

.icon { opacity: 0.5; }

.cpu-cell .mini-bar {
  width: 90px;
  height: 5px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 2px;
  position: relative;
  overflow: visible;
}

.bar-fill {
  height: 100%;
  background: var(--rdr-crimson);
  border-radius: 2px;
}

.cpu-cell .val {
  position: absolute;
  right: -42px;
  top: -6px;
  font-size: 0.72rem;
  font-family: var(--font-data);
  color: var(--text-secondary);
}

[dir="rtl"] .cpu-cell .val {
  right: auto;
  left: -42px;
}

.actions-col { text-align: right; }
[dir="rtl"] .actions-col { text-align: left; }

.kill-btn {
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.3rem;
  border-radius: 3px;
  transition: all 0.15s ease;
}

.kill-btn:hover {
  color: #ef4444;
  background: rgba(220, 38, 38, 0.12);
  border-color: rgba(220, 38, 38, 0.35);
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
