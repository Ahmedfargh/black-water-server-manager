<script setup>
import { onMounted, ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { 
  Filter, 
  ChevronLeft, 
  ChevronRight,
  Shield,
  Box,
  Terminal,
  Activity,
  User as UserIcon,
  Eye,
  X
} from 'lucide-vue-next'
import { useAuditStore } from '../stores/audit'

const { t, locale } = useI18n()
const auditStore = useAuditStore()
const filterType = ref('')
const isModalOpen = ref(false)
const modalContent = ref('')
const modalTitle = ref('')
const selectedLog = ref(null)

onMounted(() => {
  auditStore.fetchLogs()
})

const handleFilter = () => {
  auditStore.setFilters({ type: filterType.value, page: 1 })
}

const changePage = (offset) => {
  const newPage = auditStore.page + offset
  if (newPage > 0 && newPage <= totalPages.value) {
    auditStore.setFilters({ page: newPage })
  }
}

const totalPages = computed(() => {
  if (auditStore.total === 0) return 1
  return Math.ceil(auditStore.total / auditStore.limit)
})

const getLogIcon = (type) => {
  if (!type) return Activity
  const lowType = type.toLowerCase()
  if (lowType.includes('firewall')) return Shield
  if (lowType.includes('docker')) return Box
  if (lowType.includes('process')) return Terminal
  return Activity
}

const formatDate = (dateStr) => {
  if (!dateStr) return 'N/A'
  const date = new Date(dateStr)
  return date.toLocaleString(locale.value === 'ar' ? 'ar-EG' : 'en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  }).toUpperCase()
}

const isJSON = (str) => {
  if (!str || typeof str !== 'string') return false
  try {
    const obj = JSON.parse(str)
    return !!obj && typeof obj === 'object'
  } catch (e) {
    return false
  }
}

const formatJSON = (str) => {
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch (e) {
    return str
  }
}

const openDetailModal = (log) => {
  selectedLog.value = log
  if (isJSON(log.action)) {
    modalTitle.value = t('audit.action_details') || 'ACTION DETAILS'
    modalContent.value = formatJSON(log.action)
  } else if (isJSON(log.results)) {
    modalTitle.value = t('audit.execution_results') || 'EXECUTION RESULTS'
    modalContent.value = formatJSON(log.results)
  } else {
    modalTitle.value = t('audit.record_details') || 'RECORD DETAILS'
    modalContent.value = log.results || log.action
  }
  isModalOpen.value = true
}

const closeModal = () => {
  isModalOpen.value = false
}
</script>

<template>
  <div class="audit-view">
    <div class="header-row">
      <h2 class="glow-cyan">{{ $t('audit.logs_grid') }}</h2>
      <div class="filter-bar">
        <Filter :size="18" class="icon" />
        <select v-model="filterType" @change="handleFilter" class="filter-select">
          <option value="">{{ $t('audit.all_events') }}</option>
          <option value="firewall">{{ $t('audit.firewall') }}</option>
          <option value="docker">{{ $t('audit.docker') }}</option>
          <option value="process">{{ $t('audit.process') }}</option>
        </select>
      </div>
    </div>

    <div class="tron-card logs-container">
      <div class="table-wrapper">
        <table class="tron-table">
          <thead>
            <tr>
              <th class="w-icon"></th>
              <th>{{ $t('audit.type') }}</th>
              <th>{{ $t('audit.action_msg') || 'ACTION / MESSAGE' }}</th>
              <th>{{ $t('audit.actor') }}</th>
              <th class="text-right">{{ $t('audit.timestamp') }}</th>
              <th class="w-action"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in auditStore.logs" :key="log.ID">
              <td class="w-icon">
                <component :is="getLogIcon(log.service_type)" :size="18" class="neon-cyan" />
              </td>
              <td>
                <span class="type-badge">{{ log.service_type?.toUpperCase() }}</span>
              </td>
              <td class="log-message-cell">
                {{ log.action }}
                <div class="service-id" v-if="log.service_id">ID: {{ log.service_id }}</div>
              </td>
              <td>
                <div class="actor-info">
                  <UserIcon :size="14" class="dim" />
                  <span>{{ (log.user_id && log.User?.username) ? log.User.username : ($t('common.system') || 'SYSTEM') }}</span>
                </div>
              </td>
              <td class="text-right font-data dim">
                {{ formatDate(log.CreatedAt) }}
              </td>
              <td class="w-action">
                <button 
                  v-if="isJSON(log.action) || (log.results && log.results !== 'null')" 
                  @click="openDetailModal(log)" 
                  class="action-btn"
                  :title="t('audit.view_details') || 'View Details'"
                >
                  <Eye :size="16" />
                </button>
              </td>
            </tr>
            <tr v-if="auditStore.logs.length === 0">
              <td colspan="6" class="empty-cell">
                {{ $t('audit.no_records') || 'NO AUDIT RECORDS FOUND' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="pagination">
        <div class="page-stats dim">
          {{ $t('audit.showing_of', { count: auditStore.logs.length, total: auditStore.total }) || `SHOWING ${auditStore.logs.length} OF ${auditStore.total} ENTRIES` }}
        </div>
        <div class="page-controls">
          <button @click="changePage(-1)" :disabled="auditStore.page <= 1" class="page-btn">
            <ChevronLeft :size="20" />
          </button>
          <span class="page-info">{{ $t('audit.sector') || 'SECTOR' }} {{ auditStore.page }} / {{ totalPages }}</span>
          <button @click="changePage(1)" :disabled="auditStore.page >= totalPages" class="page-btn">
            <ChevronRight :size="20" />
          </button>
        </div>
      </div>

      <div v-if="auditStore.loading" class="loading-overlay">
        <Activity class="pulse" :size="48" />
        <span>{{ $t('audit.retrieving') || 'RETRIVING SECURE ARCHIVES...' }}</span>
      </div>
    </div>

    <!-- Detail Modal -->
    <Transition name="modal-fade">
      <div v-if="isModalOpen" class="modal-backdrop" @click.self="closeModal">
        <div class="modal-container tron-card">
          <div class="modal-header">
            <h3 class="glow-cyan">{{ modalTitle }}</h3>
            <button @click="closeModal" class="close-btn">
              <X :size="20" />
            </button>
          </div>
          <div class="modal-body">
            <div class="meta-info" v-if="selectedLog">
              <div class="meta-item">
                <span class="label">{{ $t('audit.service') || 'SERVICE' }}:</span>
                <span class="value cyan-text">{{ selectedLog.service_type?.toUpperCase() }}</span>
              </div>
              <div class="meta-item">
                <span class="label">{{ $t('audit.timestamp') || 'TIME' }}:</span>
                <span class="value">{{ formatDate(selectedLog.CreatedAt) }}</span>
              </div>
            </div>
            <div class="code-block-wrapper">
              <pre class="code-block font-data">{{ modalContent }}</pre>
            </div>
          </div>
          <div class="modal-footer">
            <button @click="closeModal" class="tron-btn secondary">{{ $t('common.close') || 'CLOSE' }}</button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.audit-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  height: 100%;
}

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 1rem;
  background: var(--bg-card);
  padding: 0.5rem 1rem;
  border: 1px solid rgba(0, 242, 255, 0.2);
}

.filter-select {
  background: transparent;
  border: none;
  color: var(--text-primary);
  font-family: var(--font-header);
  font-weight: 600;
  outline: none;
  cursor: pointer;
}

.filter-select option {
  background: var(--bg-black);
}

/* Table Styles */
.logs-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  position: relative;
  overflow: hidden;
}

.table-wrapper {
  flex: 1;
  overflow-y: auto;
}

.tron-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.tron-table th {
  position: sticky;
  top: 0;
  background: var(--bg-card);
  padding: 1rem;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 2px;
  color: var(--neon-cyan);
  border-bottom: 2px solid rgba(0, 242, 255, 0.2);
  z-index: 10;
}

.tron-table td {
  padding: 1rem;
  border-bottom: 1px solid rgba(0, 242, 255, 0.05);
  font-size: 0.9rem;
}

.tron-table tr:hover td {
  background: rgba(0, 242, 255, 0.02);
}

.w-icon {
  width: 50px;
  text-align: center;
}

.type-badge {
  font-size: 0.7rem;
  font-weight: 800;
  color: var(--neon-cyan);
  background: rgba(0, 242, 255, 0.05);
  padding: 0.2rem 0.5rem;
  border: 1px solid rgba(0, 242, 255, 0.2);
}

.log-message-cell {
  max-width: 400px;
}

.service-id {
  font-size: 0.7rem;
  color: var(--text-secondary);
  margin-top: 0.2rem;
  font-family: var(--font-data);
}

.actor-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.dim {
  opacity: 0.6;
}

.text-right {
  text-align: right;
}

/* Pagination */
.pagination {
  padding: 1rem 1.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid rgba(0, 242, 255, 0.1);
  background: rgba(0, 0, 0, 0.2);
}

.page-stats {
  font-size: 0.75rem;
  letter-spacing: 1px;
}

.page-controls {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.page-btn {
  background: transparent;
  border: 1px solid rgba(0, 242, 255, 0.2);
  color: var(--text-secondary);
  padding: 0.3rem;
  cursor: pointer;
  display: flex;
}

.page-btn:hover:not(:disabled) {
  color: var(--neon-cyan);
  border-color: var(--neon-cyan);
}

.page-btn:disabled {
  opacity: 0.2;
  cursor: not-allowed;
}

.page-info {
  font-size: 0.85rem;
  letter-spacing: 2px;
}

.w-action {
  width: 60px;
  text-align: center;
}

.action-btn {
  background: transparent;
  border: 1px solid rgba(0, 242, 255, 0.2);
  color: var(--text-secondary);
  padding: 0.4rem;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
}

.action-btn:hover {
  background: rgba(0, 242, 255, 0.1);
  color: var(--neon-cyan);
  border-color: var(--neon-cyan);
  box-shadow: 0 0 10px rgba(0, 242, 255, 0.2);
}

.empty-cell {
  text-align: center;
  padding: 5rem;
  color: var(--text-secondary);
  font-style: italic;
  letter-spacing: 2px;
}

/* Modal Styles */
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.85);
  backdrop-filter: blur(5px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 2rem;
}

.modal-container {
  width: 100%;
  max-width: 800px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 0 30px rgba(0, 0, 0, 0.5), 0 0 10px rgba(0, 242, 255, 0.2);
}

.modal-header {
  padding: 1.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid rgba(0, 242, 255, 0.1);
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0.5rem;
  display: flex;
  transition: color 0.2s;
}

.close-btn:hover {
  color: var(--neon-red);
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  flex: 1;
}

.meta-info {
  display: flex;
  gap: 2rem;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid rgba(224, 250, 255, 0.05);
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.meta-item .label {
  font-size: 0.7rem;
  color: var(--text-secondary);
  letter-spacing: 1px;
}

.meta-item .value {
  font-size: 0.9rem;
  font-weight: 600;
}

.cyan-text {
  color: var(--neon-cyan);
}

.code-block-wrapper {
  background: rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(0, 242, 255, 0.1);
  padding: 1rem;
  border-radius: 4px;
}

.code-block {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 0.85rem;
  color: #a5f3fc;
  line-height: 1.5;
}

.modal-footer {
  padding: 1.5rem;
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid rgba(0, 242, 255, 0.1);
  background: rgba(0, 0, 0, 0.2);
}

/* Animations */
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.3s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1.5rem;
  z-index: 100;
}

.pulse {
  animation: pulse 2s infinite;
  color: var(--neon-cyan);
}

@keyframes pulse {
  0% { transform: scale(0.95); opacity: 0.5; }
  50% { transform: scale(1.05); opacity: 1; }
  100% { transform: scale(0.95); opacity: 0.5; }
}
</style>
