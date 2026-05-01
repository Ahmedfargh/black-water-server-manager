<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { 
  Folder, 
  File, 
  ChevronLeft, 
  RefreshCw, 
  Search, 
  ArrowUp,
  HardDrive
} from 'lucide-vue-next'
import api from '../api'
import { useToastStore } from '../stores/toast'

const { t } = useI18n()
const toast = useToastStore()

const currentPath = ref('/home/ahmed')
const files = ref([])
const isLoading = ref(false)
const searchQuery = ref('')
const showHidden = ref(false)

const fetchFiles = async (path = currentPath.value) => {
  isLoading.value = true
  try {
    const response = await api.get(`/filesystem/browse`, {
      params: { path }
    })
    files.value = response.data.files || []
    currentPath.value = path
  } catch (error) {
    console.error('Failed to fetch files:', error)
    toast.error(t('common.action_failed'))
  } finally {
    isLoading.value = false
  }
}

const navigateTo = (file) => {
  if (file.is_dir) {
    const separator = currentPath.value.endsWith('/') ? '' : '/'
    const newPath = `${currentPath.value}${separator}${file.name}`
    fetchFiles(newPath)
  }
}

const goBack = () => {
  if (currentPath.value === '/' || currentPath.value === '') return
  const parts = currentPath.value.split('/')
  parts.pop()
  const parentPath = parts.join('/') || '/'
  fetchFiles(parentPath)
}

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatMode = (mode) => {
  // Simple conversion of decimal mode to octal/perm string if possible
  // But for now just show the decimal or a simplified version
  return mode.toString(8).slice(-3)
}

const sortedFiles = computed(() => {
  let result = [...files.value]
  
  // Filter hidden files
  if (!showHidden.value) {
    result = result.filter(file => !file.name.startsWith('.'))
  }

  // Filter by search query
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(file => file.name.toLowerCase().includes(query))
  }

  // Sort: Directories first, then alphabetical
  return result.sort((a, b) => {
    if (a.is_dir && !b.is_dir) return -1
    if (!a.is_dir && b.is_dir) return 1
    return a.name.localeCompare(b.name)
  })
})

onMounted(() => {
  fetchFiles()
})

const breadcrumbs = computed(() => {
  const parts = currentPath.value.split('/').filter(p => p !== '')
  const crumbs = [{ name: 'Root', path: '/' }]
  let current = ''
  parts.forEach(part => {
    current += `/${part}`
    crumbs.push({ name: part, path: current })
  })
  return crumbs
})

const navigateToPath = (path) => {
  fetchFiles(path)
}
</script>

<template>
  <div class="file-manager">
    <div class="glass-header">
      <div class="header-left">
        <HardDrive class="glow-cyan" :size="24" />
        <h1 class="glow-text uppercase letter-spacing-2">{{ $t('files.title') }}</h1>
      </div>
      <div class="header-right">
        <div class="toggle-hidden">
          <label class="switch-container">
            <input type="checkbox" v-model="showHidden">
            <span class="switch-label">Hidden</span>
          </label>
        </div>
        <div class="search-box">
          <Search :size="18" class="search-icon" />
          <input 
            type="text" 
            v-model="searchQuery" 
            :placeholder="$t('proc.search_placeholder')"
            class="glass-input"
          />
        </div>
        <button @click="fetchFiles()" class="refresh-btn" :class="{ 'spinning': isLoading }">
          <RefreshCw :size="18" />
        </button>
      </div>
    </div>

    <div class="breadcrumb-container glass-card">
      <button @click="goBack" class="back-btn" :disabled="currentPath === '/'">
        <ChevronLeft :size="20" />
        <span>{{ $t('files.back') }}</span>
      </button>
      <div class="breadcrumbs">
        <template v-for="(crumb, index) in breadcrumbs" :key="index">
          <span v-if="index > 0" class="crumb-separator">/</span>
          <button 
            @click="navigateToPath(crumb.path)" 
            class="crumb-link"
            :class="{ 'active': index === breadcrumbs.length - 1 }"
          >
            {{ crumb.name }}
          </button>
        </template>
      </div>
    </div>

    <div class="files-container glass-card">
      <div v-if="isLoading" class="loading-overlay">
        <div class="glitch-text">{{ $t('files.loading') }}</div>
      </div>

      <div class="table-scroll">
        <table class="files-table">
          <thead>
            <tr>
              <th>{{ $t('files.name') }}</th>
              <th class="hidden-mobile">{{ $t('files.size') }}</th>
              <th class="hidden-mobile">Mode</th>
              <th class="hidden-mobile">{{ $t('files.type') }}</th>
              <th>{{ $t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="sortedFiles.length === 0 && !isLoading">
              <td colspan="5" class="empty-msg">{{ $t('files.empty') }}</td>
            </tr>
            <tr 
              v-for="file in sortedFiles" 
              :key="file.name"
              class="file-row"
              :class="{ 'dir-row': file.is_dir }"
              @dblclick="navigateTo(file)"
            >
              <td class="name-cell" @click="navigateTo(file)">
                <Folder v-if="file.is_dir" :size="18" class="folder-icon" />
                <File v-else :size="18" class="file-icon" />
                <span class="file-name">{{ file.name }}</span>
              </td>
              <td class="hidden-mobile size-cell">
                {{ file.is_dir ? '--' : formatSize(file.size) }}
              </td>
              <td class="hidden-mobile mode-cell">
                <code>{{ formatMode(file.mode) }}</code>
              </td>
              <td class="hidden-mobile type-cell">
                <span :class="['type-badge', file.is_dir ? 'badge-dir' : 'badge-file']">
                  {{ file.is_dir ? $t('files.dir') : $t('files.file') }}
                </span>
              </td>
              <td class="actions-cell">
                <button 
                  v-if="file.is_dir" 
                  @click="navigateTo(file)" 
                  class="action-btn view-btn"
                  :title="$t('common.view')"
                >
                  <ArrowUp :size="16" class="rotate-90" />
                </button>
                <button 
                  v-else 
                  class="action-btn download-btn"
                  :title="$t('common.download')"
                  disabled
                >
                  <File :size="16" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Existing styles ... Adding new styles for improvements */

.table-scroll {
  flex: 1;
  overflow-y: auto;
}

.switch-container {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  color: var(--text-secondary);
  font-size: 0.8rem;
  font-weight: 600;
  text-transform: uppercase;
}

.switch-container input {
  cursor: pointer;
  accent-color: var(--neon-cyan);
}

.mode-cell code {
  background: rgba(255, 255, 255, 0.05);
  padding: 0.1rem 0.3rem;
  border-radius: 3px;
  color: var(--neon-cyan);
  font-size: 0.8rem;
}

.file-manager {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  height: 100%;
}

.glass-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  background: rgba(0, 242, 255, 0.03);
  border-left: 4px solid var(--neon-cyan);
  backdrop-filter: blur(10px);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.glow-text {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
  text-shadow: 0 0 10px rgba(0, 242, 255, 0.5);
}

.search-box {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  color: var(--text-secondary);
}

.glass-input {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(0, 242, 255, 0.2);
  color: var(--text-primary);
  padding: 0.5rem 1rem 0.5rem 2.5rem;
  border-radius: 4px;
  width: 250px;
  transition: all 0.3s ease;
  font-family: var(--font-data);
}

.glass-input:focus {
  outline: none;
  border-color: var(--neon-cyan);
  box-shadow: 0 0 15px rgba(0, 242, 255, 0.2);
}

.refresh-btn {
  background: transparent;
  border: 1px solid rgba(0, 242, 255, 0.3);
  color: var(--neon-cyan);
  padding: 0.5rem;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.refresh-btn:hover {
  background: rgba(0, 242, 255, 0.1);
  box-shadow: 0 0 10px rgba(0, 242, 255, 0.3);
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.breadcrumb-container {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.8rem 1.5rem;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: rgba(0, 242, 255, 0.1);
  border: 1px solid rgba(0, 242, 255, 0.3);
  color: var(--neon-cyan);
  padding: 0.4rem 0.8rem;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s ease;
}

.back-btn:hover:not(:disabled) {
  background: rgba(0, 242, 255, 0.2);
  transform: translateX(-2px);
}

.back-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.breadcrumbs {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  overflow-x: auto;
  white-space: nowrap;
  flex: 1;
}

.crumb-link {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.crumb-link:hover {
  color: var(--neon-cyan);
}

.crumb-link.active {
  color: var(--neon-cyan);
  font-weight: 700;
  cursor: default;
}

.crumb-separator {
  color: rgba(255, 255, 255, 0.2);
}

.files-container {
  flex: 1;
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.files-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.files-table th {
  padding: 1rem 1.5rem;
  font-size: 0.8rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 1px;
  border-bottom: 1px solid rgba(0, 242, 255, 0.1);
  background: rgba(255, 255, 255, 0.02);
}

.files-table td {
  padding: 0.8rem 1.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  font-family: var(--font-data);
  font-size: 0.9rem;
}

.file-row {
  transition: all 0.2s ease;
  cursor: pointer;
}

.file-row:hover {
  background: rgba(0, 242, 255, 0.05);
}

.name-cell {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  color: var(--text-primary);
}

.folder-icon {
  color: var(--neon-cyan);
}

.file-icon {
  color: var(--text-secondary);
}

.dir-row .file-name {
  color: var(--neon-cyan);
  font-weight: 600;
}

.type-badge {
  padding: 0.2rem 0.6rem;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 1px;
}

.badge-dir {
  background: rgba(0, 242, 255, 0.1);
  color: var(--neon-cyan);
  border: 1px solid rgba(0, 242, 255, 0.3);
}

.badge-file {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-secondary);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.empty-msg {
  text-align: center;
  padding: 4rem !important;
  color: var(--text-secondary);
  font-style: italic;
  letter-spacing: 2px;
}

.loading-overlay {
  position: absolute;
  inset: 0;
  background: rgba(5, 7, 10, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  backdrop-filter: blur(5px);
}

.actions-cell {
  display: flex;
  gap: 0.5rem;
}

.action-btn {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-secondary);
  padding: 0.4rem;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.view-btn:hover {
  color: var(--neon-cyan);
  border-color: var(--neon-cyan);
  background: rgba(0, 242, 255, 0.1);
}

.download-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.rotate-90 {
  transform: rotate(90deg);
}

[dir="rtl"] .rotate-90 {
  transform: rotate(-90deg);
}

/* Glass Card Utility - copied from common styles if needed, but defined here for safety */
.glass-card {
  background: var(--bg-card);
  backdrop-filter: blur(20px);
  border: 1px solid var(--neon-cyan-glow);
  border-radius: 4px;
  box-shadow: var(--card-shadow);
}

@media (max-width: 768px) {
  .hidden-mobile {
    display: none;
  }
  
  .glass-input {
    width: 150px;
  }
}

.letter-spacing-2 {
  letter-spacing: 2px;
}
</style>
