<script setup>
import { onMounted, ref } from 'vue'
import { 
  Box, 
  Play, 
  Square, 
  RotateCw, 
  Terminal, 
  Activity, 
  Trash2,
  X
} from 'lucide-vue-next'
import { useDockerStore } from '../stores/docker'
import { useAuthStore } from '../stores/auth'

const dockerStore = useDockerStore()
const showLogs = ref(false)
const selectedContainer = ref(null)
const logsData = ref([])
let wsLogs

const fetchContainers = () => dockerStore.fetchContainers()

onMounted(() => {
  fetchContainers()
})

const handleAction = async (id, action) => {
  try {
    await dockerStore.performAction(id, action)
  } catch (err) {
    alert(`ACTION FAILED: ${action.toUpperCase()} operation failed on node ${id}.`)
  }
}

const openLogs = (container) => {
  selectedContainer.value = container
  showLogs.value = true
  logsData.value = []
  connectLogs(container.id)
}

const closeLogs = () => {
  showLogs.value = false
  if (wsLogs) wsLogs.close()
  selectedContainer.value = null
}

const connectLogs = (id) => {
  const authStore = useAuthStore()
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws/docker/${id}/logs?token=${authStore.token}`
  
  wsLogs = new WebSocket(wsUrl)
  wsLogs.onmessage = (event) => {
    logsData.value.push(event.data)
    // Keep last 100 lines
    if (logsData.value.length > 100) logsData.value.shift()
  }
}
</script>

<template>
  <div class="docker-view">
    <div class="header-actions">
      <h2 class="glow-cyan">CONTAINER GRID</h2>
      <button @click="fetchContainers" class="refresh-btn tron-btn">
        <RotateCw :size="18" />
        RESCAN NODES
      </button>
    </div>

    <!-- Container Grid -->
    <div class="container-grid">
      <div 
        v-for="container in dockerStore.containers" 
        :key="container.id"
        class="tron-card container-card"
        :class="{ 'running': container.status?.includes('Up'), 'stopped': !container.status?.includes('Up') }"
      >
        <div class="card-header">
           <Box :size="24" class="node-icon" />
           <div class="node-info">
             <span class="node-name">{{ container.names?.[0] || 'UNKNOWN NODE' }}</span>
             <span class="node-id">{{ container.id.substring(0, 12) }}</span>
           </div>
           <div class="status-badge" :class="{ 'pulse': container.status?.includes('Up') }">
             {{ container.status?.includes('Up') ? 'ACTIVE' : 'OFFLINE' }}
           </div>
        </div>

        <div class="card-body">
           <div class="detail-row">
             <span class="label">IMAGE</span>
             <span class="value">{{ container.image }}</span>
           </div>
           <div class="detail-row">
             <span class="label">UPTIME</span>
             <span class="value">{{ container.status }}</span>
           </div>
        </div>

        <div class="card-actions">
           <button 
             v-if="!container.status?.includes('Up')" 
             @click="handleAction(container.id, 'start')"
             class="action-btn glow-cyan"
             title="START"
           >
             <Play :size="18" />
           </button>
           <button 
             v-if="container.status?.includes('Up')" 
             @click="handleAction(container.id, 'stop')"
             class="action-btn glow-orange"
             title="STOP"
           >
             <Square :size="18" />
           </button>
           <button 
             @click="handleAction(container.id, 'restart')"
             class="action-btn"
             title="RESTART"
           >
             <RotateCw :size="18" />
           </button>
           <button 
             @click="openLogs(container)"
             class="action-btn"
             title="VIEW LOGS"
           >
             <Terminal :size="18" />
           </button>
        </div>
      </div>
    </div>

    <!-- Logs Modal -->
    <transition name="modal">
      <div v-if="showLogs" class="modal-overlay">
        <div class="tron-card modal-container">
          <div class="modal-header">
            <h3>LOG STREAM: {{ selectedContainer?.name }}</h3>
            <button @click="closeLogs" class="close-btn"><X /></button>
          </div>
          <div class="log-viewport">
            <div v-for="(log, i) in logsData" :key="i" class="log-line">
              <span class="line-num">[{{ i + 1 }}]</span>
              <span class="line-text">{{ log }}</span>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.docker-view {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.header-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.container-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
}

.container-card {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  border-left: 4px solid transparent;
}

.container-card.running {
  border-left-color: var(--neon-cyan);
}

.container-card.stopped {
  border-left-color: var(--neon-orange);
  opacity: 0.8;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.node-icon {
  color: var(--text-secondary);
}

.node-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.node-name {
  font-weight: 700;
  font-size: 1.1rem;
  letter-spacing: 1px;
}

.node-id {
  font-family: var(--font-data);
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.status-badge {
  font-size: 0.7rem;
  padding: 0.2rem 0.6rem;
  border: 1px solid currentColor;
  border-radius: 2px;
  letter-spacing: 1px;
}

.container-card.running .status-badge { color: var(--neon-cyan); }
.container-card.stopped .status-badge { color: var(--neon-orange); }

.card-body {
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.85rem;
}

.detail-row .label {
  color: var(--text-secondary);
}

.detail-row .value {
  font-family: var(--font-data);
}

.card-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  border-top: 1px solid rgba(0, 242, 255, 0.1);
  padding-top: 1rem;
}

.action-btn {
  background: transparent;
  border: 1px solid rgba(224, 250, 255, 0.1);
  color: var(--text-primary);
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: rgba(0, 242, 255, 0.05);
  border-color: var(--neon-cyan);
}

/* Modal Styling */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(5px);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.modal-container {
  width: 100%;
  max-width: 900px;
  height: 80vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  padding: 1.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid rgba(0, 242, 255, 0.1);
}

.log-viewport {
  flex: 1;
  background: #000;
  padding: 1.5rem;
  overflow-y: auto;
  font-family: var(--font-data);
  font-size: 0.9rem;
}

.log-line {
  display: flex;
  gap: 1rem;
  margin-bottom: 0.3rem;
}

.line-num {
  color: var(--text-secondary);
  user-select: none;
}

.line-text {
  color: var(--neon-cyan);
  white-space: pre-wrap;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
}

.close-btn:hover { color: #fff; }
</style>
