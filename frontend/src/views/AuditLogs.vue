<script setup>
import { onMounted, ref } from 'vue'
import { 
  History, 
  Search, 
  Filter, 
  ChevronLeft, 
  ChevronRight,
  Shield,
  Box,
  Terminal,
  Activity
} from 'lucide-vue-next'
import { useAuditStore } from '../stores/audit'

const auditStore = useAuditStore()
const filterType = ref('')

onMounted(() => {
  auditStore.fetchLogs()
})

const handleFilter = () => {
  auditStore.setFilters({ type: filterType.value, page: 1 })
}

const changePage = (offset) => {
  const newPage = auditStore.page + offset
  if (newPage > 0) {
    auditStore.setFilters({ page: newPage })
  }
}

const getLogIcon = (type) => {
  if (type.includes('firewall')) return Shield
  if (type.includes('docker')) return Box
  if (type.includes('process')) return Terminal
  return Activity
}

const formatDate = (dateStr) => {
  const date = new Date(dateStr)
  return date.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  }).toUpperCase()
}
</script>

<template>
  <div class="audit-view">
    <div class="header-row">
      <h2 class="glow-cyan">SYSTEM AUDIT LOGS</h2>
      <div class="filter-bar">
        <Filter :size="18" class="icon" />
        <select v-model="filterType" @change="handleFilter" class="filter-select">
          <option value="">ALL EVENTS</option>
          <option value="firewall">FIREWALL</option>
          <option value="docker">DOCKER</option>
          <option value="process">PROCESS</option>
        </select>
      </div>
    </div>

    <div class="tron-card logs-container">
      <div class="logs-list">
        <div v-for="log in auditStore.logs" :key="log.id" class="log-entry">
          <div class="log-icon-wrap">
            <component :is="getLogIcon(log.type)" :size="20" />
          </div>
          <div class="log-content">
            <div class="log-header">
              <span class="log-type">{{ log.type?.toUpperCase() }}</span>
              <span class="log-time font-data">{{ formatDate(log.created_at) }}</span>
            </div>
            <p class="log-message">{{ log.message }}</p>
            <div class="log-meta">
              <span class="user-tag">ACTOR: {{ log.user_id || 'SYSTEM' }}</span>
              <span class="ip-tag" v-if="log.ip_address">IP: {{ log.ip_address }}</span>
            </div>
          </div>
        </div>

        <div v-if="auditStore.logs.length === 0" class="empty-msg">
          NO AUDIT RECORDS FOUND IN THE CURRENT SECTOR.
        </div>
      </div>

      <div class="pagination">
        <button @click="changePage(-1)" :disabled="auditStore.page <= 1" class="page-btn">
          <ChevronLeft :size="20" />
        </button>
        <span class="page-info">SECTOR {{ auditStore.page }}</span>
        <button @click="changePage(1)" :disabled="auditStore.logs.length < auditStore.limit" class="page-btn">
          <ChevronRight :size="20" />
        </button>
      </div>

      <div v-if="auditStore.loading" class="loading-overlay">
        <Activity class="pulse" :size="48" />
        <span>RETRIVING SECURE ARCHIVES...</span>
      </div>
    </div>
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

/* Logs List */
.logs-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  position: relative;
  min-height: 500px;
}

.logs-list {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
}

.log-entry {
  display: flex;
  gap: 1.5rem;
  padding: 1.5rem;
  border-bottom: 1px solid rgba(0, 242, 255, 0.05);
  transition: all 0.2s ease;
}

.log-entry:hover {
  background: rgba(0, 242, 255, 0.02);
}

.log-icon-wrap {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 242, 255, 0.05);
  border: 1px solid rgba(0, 242, 255, 0.2);
  color: var(--neon-cyan);
  flex-shrink: 0;
}

.log-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.log-type {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 2px;
  color: var(--neon-cyan);
}

.log-time {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.log-message {
  font-size: 0.95rem;
  line-height: 1.4;
}

.log-meta {
  display: flex;
  gap: 1.5rem;
  font-size: 0.7rem;
  color: var(--text-secondary);
  font-family: var(--font-data);
}

.pagination {
  padding: 1.5rem;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 2rem;
  border-top: 1px solid rgba(0, 242, 255, 0.1);
}

.page-btn {
  background: transparent;
  border: 1px solid rgba(0, 242, 255, 0.2);
  color: var(--text-secondary);
  padding: 0.4rem;
  cursor: pointer;
  display: flex;
}

.page-btn:hover:not(:disabled) {
  color: var(--neon-cyan);
  border-color: var(--neon-cyan);
}

.page-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.page-info {
  font-size: 0.9rem;
  letter-spacing: 2px;
}

.empty-msg {
  text-align: center;
  padding: 5rem;
  color: var(--text-secondary);
  font-style: italic;
  letter-spacing: 2px;
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
</style>
