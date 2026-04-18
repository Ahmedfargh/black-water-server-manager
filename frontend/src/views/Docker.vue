<script setup>
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { 
  Box, 
  Play, 
  Square, 
  RotateCw, 
  Terminal, 
  Activity, 
  Trash2,
  X,
  HardDrive
} from 'lucide-vue-next'
import { useDockerStore } from '../stores/docker'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const dockerStore = useDockerStore()
const showLogs = ref(false)
const selectedContainer = ref(null)
const logsData = ref([])
const showVolumes = ref(false)
const selectedVolumesContainer = ref(null)
let wsLogs

const fetchContainers = () => dockerStore.fetchContainers()

onMounted(() => {
  fetchContainers()
})

const handleAction = async (id, action) => {
  try {
    await dockerStore.performAction(id, action)
  } catch (err) {
    alert(`${t('docker.action_failed')}: ${action.toUpperCase()} operation failed on node ${id}.`)
  }
}

const handlePrune = async (id) => {
  if (confirm(t('docker.confirm_prune', { id }))) {
    try {
      await dockerStore.pruneContainer(id)
    } catch (err) {
      alert(`${t('docker.prune_failed')}: Critical error while attempting to purge node ${id}.`)
    }
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

const openVolumes = async (container) => {
  selectedVolumesContainer.value = container
  try {
    await dockerStore.fetchVolumes(container.id)
    showVolumes.value = true
  } catch (err) {
    alert(`FETCH FAILED: Could not retrieve volumes for node ${container.id}.`)
  }
}

const closeVolumes = () => {
  showVolumes.value = false
  selectedVolumesContainer.value = null
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
      <h2 class="glow-cyan">{{ $t('docker.container_grid') }}</h2>
      <button @click="fetchContainers" class="refresh-btn tron-btn">
        <RotateCw :size="18" />
        {{ $t('docker.rescan_nodes') }}
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
             {{ container.status?.includes('Up') ? $t('common.active') : $t('common.offline') }}
           </div>
        </div>

        <div class="card-body">
           <div class="detail-row">
             <span class="label">{{ $t('docker.image') }}</span>
             <span class="value">{{ container.image }}</span>
           </div>
           <div class="detail-row">
             <span class="label">{{ $t('docker.uptime') }}</span>
             <span class="value">{{ container.status }}</span>
           </div>
        </div>

        <div class="card-actions">
           <button 
             v-if="!container.status?.includes('Up')" 
             @click="handleAction(container.id, 'start')"
             class="action-btn glow-cyan"
             :title="$t('docker.start')"
           >
             <Play :size="18" />
           </button>
           <button 
             v-if="container.status?.includes('Up')" 
             @click="handleAction(container.id, 'stop')"
             class="action-btn glow-orange"
             :title="$t('docker.stop')"
           >
             <Square :size="18" />
           </button>
           <button 
             @click="handleAction(container.id, 'restart')"
             class="action-btn"
             :title="$t('docker.restart')"
           >
             <RotateCw :size="18" />
           </button>
           <button 
             @click="openLogs(container)"
             class="action-btn"
             :title="$t('docker.view_logs')"
           >
             <Terminal :size="18" />
           </button>
           <button 
             @click="openVolumes(container)"
             class="action-btn"
             :title="$t('docker.view_volumes')"
           >
             <HardDrive :size="18" />
           </button>
           <button 
             @click="handlePrune(container.id)"
             class="action-btn glow-red"
             :title="$t('docker.prune')"
           >
             <Trash2 :size="18" />
           </button>
        </div>
      </div>
    </div>

    <!-- Logs Modal -->
    <transition name="modal">
      <div v-if="showLogs" class="modal-overlay">
        <div class="tron-card modal-container">
          <div class="modal-header">
            <h3>{{ $t('docker.log_stream') }}: {{ selectedContainer?.names?.[0] || selectedContainer?.name }}</h3>
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

    <!-- Volumes Modal -->
    <transition name="modal">
      <div v-if="showVolumes" class="modal-overlay">
        <div class="tron-card modal-container volumes-modal">
          <div class="modal-header">
            <h3>{{ $t('docker.volume_mappings') }}: {{ selectedVolumesContainer?.names?.[0] || 'NODE' }}</h3>
            <button @click="closeVolumes" class="close-btn"><X /></button>
          </div>
          <div class="volumes-content">
            <div v-if="dockerStore.volumes.length === 0" class="no-data">
              {{ $t('docker.no_volumes') }}
            </div>
            <table v-else class="tron-table">
              <thead>
                <tr>
                  <th>{{ $t('docker.type') }}</th>
                  <th>{{ $t('docker.source') }}</th>
                  <th>{{ $t('docker.destination') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(vol, i) in dockerStore.volumes" :key="i">
                  <td><span class="type-tag" :class="vol.Type.toLowerCase()">{{ vol.Type }}</span></td>
                  <td class="path-cell" :title="vol.Source"><code>{{ vol.Source }}</code></td>
                  <td class="path-cell" :title="vol.Destination"><code>{{ vol.Destination }}</code></td>
                </tr>
              </tbody>
            </table>
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

/* Volumes Table Styling */
.volumes-modal {
  height: auto;
  max-height: 80vh;
}

.volumes-content {
  padding: 1.5rem;
  overflow-y: auto;
}

.no-data {
  text-align: center;
  padding: 3rem;
  color: var(--text-secondary);
  font-family: var(--font-data);
  letter-spacing: 2px;
}

.tron-table {
  width: 100%;
  border-collapse: collapse;
  font-family: var(--font-data);
}

.tron-table th {
  text-align: left;
  padding: 1rem;
  color: var(--neon-cyan);
  border-bottom: 2px solid rgba(0, 242, 255, 0.2);
  font-size: 0.8rem;
  letter-spacing: 1px;
}

.tron-table td {
  padding: 1rem;
  border-bottom: 1px solid rgba(0, 242, 255, 0.05);
  font-size: 0.85rem;
}

.type-tag {
  padding: 0.1rem 0.4rem;
  border-radius: 3px;
  font-size: 0.7rem;
  text-transform: uppercase;
  border: 1px solid currentColor;
}

.type-tag.bind { color: var(--neon-cyan); }
.type-tag.volume { color: var(--neon-purple, #bc13fe); }

.path-cell {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.path-cell code {
  color: var(--text-secondary);
}

.path-cell:hover code {
  color: var(--text-primary);
}
</style>
