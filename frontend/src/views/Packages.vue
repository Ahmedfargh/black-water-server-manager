<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../api'
import { 
  Package, 
  RefreshCw, 
  Search, 
  Layers, 
  CheckCircle2, 
  AlertCircle, 
  ArrowUpCircle, 
  HardDrive,
  Cpu,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Terminal,
  Server
} from 'lucide-vue-next'

const { t } = useI18n()

const loadingOverview = ref(true)
const loadingPackages = ref(false)
const loadingUpdates = ref(false)
const errorMsg = ref('')

const overview = ref(null)
const selectedManager = ref('')
const activeTab = ref('installed') // 'installed' | 'updates'

// Installed packages state & pagination
const packagesList = ref([])
const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = ref(50)
const totalPackages = ref(0)
const jumpPage = ref(1)
let searchTimeout = null

// Updates state & pagination
const updatesList = ref([])
const updatesSearch = ref('')
const updatesPage = ref(1)
const updatesPageSize = ref(25)

const fetchOverview = async (forceRefresh = false) => {
  loadingOverview.value = true
  errorMsg.value = ''
  try {
    const res = await api.get(`/packages/overview${forceRefresh ? '?refresh=true' : ''}`)
    overview.value = res.data.data
    
    // Default selected manager to primary or first available
    if (!selectedManager.value && overview.value?.managers?.length > 0) {
      const primary = overview.value.managers.find(m => m.is_primary)
      selectedManager.value = primary ? primary.name : overview.value.managers[0].name
    }
  } catch (err) {
    errorMsg.value = err.response?.data?.message || 'Failed to load package managers'
  } finally {
    loadingOverview.value = false
  }
}

const fetchPackages = async () => {
  if (!selectedManager.value) return
  loadingPackages.value = true
  try {
    const res = await api.get(`/packages/${selectedManager.value}/list`, {
      params: {
        query: searchQuery.value,
        page: currentPage.value,
        limit: pageSize.value
      }
    })
    packagesList.value = res.data.data || []
    totalPackages.value = res.data.total || 0
    jumpPage.value = currentPage.value
  } catch (err) {
    packagesList.value = []
  } finally {
    loadingPackages.value = false
  }
}

const fetchUpdates = async () => {
  if (!selectedManager.value) return
  loadingUpdates.value = true
  try {
    const res = await api.get(`/packages/${selectedManager.value}/updates`)
    updatesList.value = res.data.data || []
    updatesPage.value = 1
  } catch (err) {
    updatesList.value = []
  } finally {
    loadingUpdates.value = false
  }
}

const handleSearchInput = () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    fetchPackages()
  }, 350)
}

const selectManager = (name) => {
  selectedManager.value = name
  currentPage.value = 1
  jumpPage.value = 1
  searchQuery.value = ''
  updatesSearch.value = ''
  updatesPage.value = 1
  if (activeTab.value === 'installed') {
    fetchPackages()
  } else {
    fetchUpdates()
  }
}

const switchTab = (tab) => {
  activeTab.value = tab
  if (tab === 'installed') {
    fetchPackages()
  } else {
    fetchUpdates()
  }
}

// Pagination computations
const totalPages = computed(() => {
  return Math.ceil(totalPackages.value / pageSize.value) || 1
})

const visiblePages = computed(() => {
  const current = currentPage.value
  const total = totalPages.value
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }

  const pages = []
  pages.push(1)

  if (current > 4) {
    pages.push('...')
  }

  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)

  for (let i = start; i <= end; i++) {
    pages.push(i)
  }

  if (current < total - 3) {
    pages.push('...')
  }

  pages.push(total)
  return pages
})

const goToPage = (p) => {
  if (p === '...' || p < 1 || p > totalPages.value || p === currentPage.value) return
  currentPage.value = p
  jumpPage.value = p
  fetchPackages()
  scrollWorkspace()
}

const handleJump = () => {
  const p = parseInt(jumpPage.value, 10)
  if (!isNaN(p) && p >= 1 && p <= totalPages.value) {
    goToPage(p)
  } else {
    jumpPage.value = currentPage.value
  }
}

const handlePageSizeChange = () => {
  currentPage.value = 1
  jumpPage.value = 1
  fetchPackages()
}

const scrollWorkspace = () => {
  const el = document.querySelector('.workspace-card')
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

// Updates pagination
const filteredUpdates = computed(() => {
  if (!updatesSearch.value) return updatesList.value
  const q = updatesSearch.value.toLowerCase()
  return updatesList.value.filter(u => 
    u.name.toLowerCase().includes(q) || 
    (u.repository && u.repository.toLowerCase().includes(q))
  )
})

const totalUpdatesPages = computed(() => {
  return Math.ceil(filteredUpdates.value.length / updatesPageSize.value) || 1
})

const paginatedUpdates = computed(() => {
  const start = (updatesPage.value - 1) * updatesPageSize.value
  return filteredUpdates.value.slice(start, start + updatesPageSize.value)
})

const goToUpdatesPage = (p) => {
  if (p < 1 || p > totalUpdatesPages.value || p === updatesPage.value) return
  updatesPage.value = p
}

watch(selectedManager, () => {
  if (activeTab.value === 'installed') {
    fetchPackages()
  } else {
    fetchUpdates()
  }
})

onMounted(async () => {
  await fetchOverview()
  if (selectedManager.value) {
    fetchPackages()
  }
})
</script>

<template>
  <div class="packages-container">
    <!-- Top System Banner -->
    <div class="tron-card os-banner" v-if="overview">
      <div class="os-info">
        <div class="os-avatar">
          <Server class="glow-cyan" :size="32" />
        </div>
        <div>
          <div class="os-label">{{ $t('packages.host_distribution') || 'HOST DISTRIBUTION' }}</div>
          <h2 class="os-title">{{ overview.os.pretty_name || overview.os.name }}</h2>
          <div class="os-badges">
            <span class="cyber-badge cyan">ID: {{ overview.os.id }}</span>
            <span v-if="overview.os.version_id" class="cyber-badge muted">v{{ overview.os.version_id }}</span>
            <span class="cyber-badge primary-tag" v-if="overview.primary_manager">
              {{ $t('packages.primary_engine') || 'PRIMARY' }}: {{ overview.primary_manager.toUpperCase() }}
            </span>
          </div>
        </div>
      </div>

      <div class="banner-actions">
        <button 
          @click="fetchOverview(true)" 
          class="cyber-btn"
          :disabled="loadingOverview"
        >
          <RefreshCw :size="16" :class="{ 'spin': loadingOverview }" />
          <span>{{ $t('packages.rescan') || 'RESCAN SYSTEM' }}</span>
        </button>
      </div>
    </div>

    <!-- Manager Selector Cards -->
    <div class="managers-grid" v-if="overview?.managers?.length">
      <div 
        v-for="mgr in overview.managers" 
        :key="mgr.name"
        @click="selectManager(mgr.name)"
        class="tron-card manager-card"
        :class="{ 'active-card': selectedManager === mgr.name }"
      >
        <div class="card-header">
          <div class="mgr-title-box">
            <Package class="mgr-icon" :size="20" />
            <span class="mgr-name">{{ mgr.name.toUpperCase() }}</span>
          </div>
          <span 
            class="mgr-badge" 
            :class="mgr.is_primary ? 'badge-primary' : 'badge-universal'"
          >
            {{ mgr.is_primary ? ($t('packages.primary') || 'PRIMARY') : mgr.category.toUpperCase() }}
          </span>
        </div>

        <div class="card-metrics">
          <div class="metric-row">
            <span class="metric-label">{{ $t('packages.version') || 'VERSION' }}:</span>
            <span class="metric-val mono">{{ mgr.version || 'Unknown' }}</span>
          </div>
          <div class="metric-row">
            <span class="metric-label">{{ $t('packages.packages') || 'PACKAGES' }}:</span>
            <span class="metric-val highlight mono">{{ mgr.total_packages.toLocaleString() }}</span>
          </div>
          <div class="metric-row path-row">
            <span class="metric-label">{{ $t('packages.path') || 'PATH' }}:</span>
            <span class="metric-val mono text-dim">{{ mgr.executable_path }}</span>
          </div>
        </div>

        <div class="card-glow-bar" :class="{ 'bar-active': selectedManager === mgr.name }"></div>
      </div>
    </div>

    <!-- Main Workspace Section -->
    <div class="tron-card workspace-card" v-if="selectedManager">
      <div class="workspace-header">
        <!-- Tab Controls -->
        <div class="tab-controls">
          <button 
            @click="switchTab('installed')"
            class="tab-btn"
            :class="{ active: activeTab === 'installed' }"
          >
            <Layers :size="16" />
            <span>{{ $t('packages.installed_packages') || 'INSTALLED PACKAGES' }}</span>
            <span class="counter-pill">{{ totalPackages.toLocaleString() }}</span>
          </button>
          
          <button 
            @click="switchTab('updates')"
            class="tab-btn"
            :class="{ active: activeTab === 'updates' }"
          >
            <ArrowUpCircle :size="16" />
            <span>{{ $t('packages.pending_updates') || 'PENDING UPDATES' }}</span>
            <span v-if="updatesList.length" class="counter-pill warn">{{ updatesList.length }}</span>
          </button>
        </div>

        <!-- Search Bar (Active for Installed) -->
        <div class="search-box" v-if="activeTab === 'installed'">
          <Search :size="16" class="search-icon" />
          <input 
            v-model="searchQuery" 
            @input="handleSearchInput"
            type="text" 
            :placeholder="($t('packages.search_placeholder') || 'Filter packages by name...') + ' (' + selectedManager + ')'"
            class="cyber-input"
          />
        </div>

        <!-- Refresh & Filter for Updates Tab -->
        <div class="updates-header-actions" v-else>
          <div class="search-box small">
            <Search :size="14" class="search-icon" />
            <input 
              v-model="updatesSearch" 
              type="text" 
              :placeholder="$t('packages.filter_updates') || 'Filter updates...'"
              class="cyber-input"
            />
          </div>
          <button 
            @click="fetchUpdates" 
            class="cyber-btn"
            :disabled="loadingUpdates"
          >
            <RefreshCw :size="16" :class="{ 'spin': loadingUpdates }" />
            <span>{{ $t('packages.check_updates') || 'CHECK UPDATES' }}</span>
          </button>
        </div>
      </div>

      <!-- Tab Content: Installed Packages -->
      <div v-if="activeTab === 'installed'" class="tab-body">
        <div v-if="loadingPackages" class="loading-state">
          <RefreshCw :size="28" class="spin glow-cyan" />
          <span>{{ $t('common.loading') || 'SCANNING SYSTEM REPOSITORIES...' }}</span>
        </div>

        <div v-else-if="packagesList.length === 0" class="empty-state">
          <AlertCircle :size="32" class="glow-orange" />
          <p>{{ $t('packages.no_packages_found') || 'No packages found matching query.' }}</p>
        </div>

        <div v-else class="table-responsive">
          <table class="cyber-table">
            <thead>
              <tr>
                <th style="width: 70px;">#</th>
                <th>{{ $t('packages.package_name') || 'PACKAGE NAME' }}</th>
                <th>{{ $t('packages.version') || 'VERSION' }}</th>
                <th>{{ $t('packages.engine') || 'SOURCE' }}</th>
                <th v-if="selectedManager === 'flatpak'">{{ $t('packages.app_id') || 'APPLICATION ID' }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(pkg, idx) in packagesList" :key="pkg.name">
                <td class="mono text-dim">{{ (currentPage - 1) * pageSize + idx + 1 }}</td>
                <td class="pkg-name-cell">
                  <Package :size="15" class="pkg-icon" />
                  <span class="pkg-title">{{ pkg.name }}</span>
                </td>
                <td class="mono pkg-ver">{{ pkg.version }}</td>
                <td>
                  <span class="engine-tag">{{ pkg.source }}</span>
                </td>
                <td v-if="selectedManager === 'flatpak'" class="mono text-dim">
                  {{ pkg.description }}
                </td>
              </tr>
            </tbody>
          </table>

          <!-- Rich Interactive Pagination Footer -->
          <div class="pagination-footer" v-if="totalPackages > 0">
            <!-- Left Info & Per Page Picker -->
            <div class="pagination-meta">
              <div class="pagination-info">
                {{ $t('packages.showing') || 'Showing' }} 
                <span class="mono highlight">{{ (currentPage - 1) * pageSize + 1 }}</span> - 
                <span class="mono highlight">{{ Math.min(currentPage * pageSize, totalPackages) }}</span> 
                {{ $t('packages.of') || 'of' }} 
                <span class="mono highlight">{{ totalPackages.toLocaleString() }}</span>
              </div>

              <div class="page-size-picker">
                <span class="picker-label">{{ $t('packages.per_page') || 'Per page:' }}</span>
                <select v-model="pageSize" @change="handlePageSizeChange" class="cyber-select">
                  <option :value="25">25</option>
                  <option :value="50">50</option>
                  <option :value="100">100</option>
                  <option :value="200">200</option>
                </select>
              </div>
            </div>

            <!-- Right Navigation & Page Pills -->
            <div class="pagination-nav">
              <!-- First & Prev -->
              <button 
                @click="goToPage(1)" 
                :disabled="currentPage <= 1 || loadingPackages"
                class="nav-icon-btn"
                :title="$t('packages.first') || 'First page'"
              >
                <ChevronsLeft :size="16" />
              </button>
              <button 
                @click="goToPage(currentPage - 1)" 
                :disabled="currentPage <= 1 || loadingPackages"
                class="nav-icon-btn"
                :title="$t('packages.previous') || 'Previous page'"
              >
                <ChevronLeft :size="16" />
              </button>

              <!-- Page Number Pills -->
              <div class="page-pills">
                <button 
                  v-for="(p, idx) in visiblePages" 
                  :key="idx"
                  @click="goToPage(p)"
                  class="page-pill"
                  :class="{ 
                    'pill-active': p === currentPage,
                    'pill-dots': p === '...' 
                  }"
                  :disabled="p === '...' || loadingPackages"
                >
                  {{ p }}
                </button>
              </div>

              <!-- Next & Last -->
              <button 
                @click="goToPage(currentPage + 1)" 
                :disabled="currentPage >= totalPages || loadingPackages"
                class="nav-icon-btn"
                :title="$t('packages.next') || 'Next page'"
              >
                <ChevronRight :size="16" />
              </button>
              <button 
                @click="goToPage(totalPages)" 
                :disabled="currentPage >= totalPages || loadingPackages"
                class="nav-icon-btn"
                :title="$t('packages.last') || 'Last page'"
              >
                <ChevronsRight :size="16" />
              </button>

              <!-- Jump to Page Input -->
              <div class="jump-box" v-if="totalPages > 5">
                <span class="text-dim">{{ $t('packages.go_to') || 'Go to' }}:</span>
                <input 
                  type="number" 
                  v-model="jumpPage" 
                  @keyup.enter="handleJump"
                  min="1" 
                  :max="totalPages" 
                  class="jump-input mono"
                />
                <button @click="handleJump" class="jump-btn">GO</button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Tab Content: Pending Updates -->
      <div v-else class="tab-body">
        <div v-if="loadingUpdates" class="loading-state">
          <RefreshCw :size="28" class="spin glow-cyan" />
          <span>{{ $t('packages.checking_upstream') || 'QUERYING UPSTREAM REPOSITORIES...' }}</span>
        </div>

        <div v-else-if="updatesList.length === 0" class="empty-state clean">
          <CheckCircle2 :size="36" class="glow-cyan" />
          <h3>{{ $t('packages.system_up_to_date') || 'SYSTEM IS UP TO DATE' }}</h3>
          <p>{{ $t('packages.no_pending_updates', { manager: selectedManager }) || 'No pending updates found for ' + selectedManager + '.' }}</p>
        </div>

        <div v-else class="table-responsive">
          <table class="cyber-table">
            <thead>
              <tr>
                <th style="width: 70px;">#</th>
                <th>{{ $t('packages.package_name') || 'PACKAGE' }}</th>
                <th>{{ $t('packages.installed_version') || 'INSTALLED' }}</th>
                <th>{{ $t('packages.latest_version') || 'UPSTREAM NEW' }}</th>
                <th>{{ $t('packages.repository') || 'REPOSITORY' }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(up, idx) in paginatedUpdates" :key="up.name">
                <td class="mono text-dim">{{ (updatesPage - 1) * updatesPageSize + idx + 1 }}</td>
                <td class="pkg-name-cell">
                  <ArrowUpCircle :size="16" class="update-icon glow-orange" />
                  <span class="pkg-title">{{ up.name }}</span>
                </td>
                <td class="mono old-ver">{{ up.current_version }}</td>
                <td class="mono new-ver">
                  <span class="new-pill">{{ up.new_version }}</span>
                </td>
                <td class="text-dim">{{ up.repository || 'default' }}</td>
              </tr>
            </tbody>
          </table>

          <!-- Updates Pagination Footer -->
          <div class="pagination-footer" v-if="totalUpdatesPages > 1">
            <div class="pagination-info">
              {{ $t('packages.showing') || 'Showing' }} 
              <span class="mono highlight">{{ (updatesPage - 1) * updatesPageSize + 1 }}</span> - 
              <span class="mono highlight">{{ Math.min(updatesPage * updatesPageSize, filteredUpdates.length) }}</span> 
              {{ $t('packages.of') || 'of' }} 
              <span class="mono highlight">{{ filteredUpdates.length }}</span>
            </div>

            <div class="pagination-buttons">
              <button 
                @click="goToUpdatesPage(updatesPage - 1)" 
                :disabled="updatesPage <= 1"
                class="nav-icon-btn"
              >
                <ChevronLeft :size="16" />
              </button>
              <span class="page-current mono">{{ updatesPage }} / {{ totalUpdatesPages }}</span>
              <button 
                @click="goToUpdatesPage(updatesPage + 1)" 
                :disabled="updatesPage >= totalUpdatesPages"
                class="nav-icon-btn"
              >
                <ChevronRight :size="16" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.packages-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
  animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

/* OS Banner */
.os-banner {
  padding: 20px 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.os-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.os-avatar {
  width: 52px;
  height: 52px;
  border-radius: 4px;
  background: rgba(220, 38, 38, 0.1);
  border: 1px solid var(--border-crimson);
  display: flex;
  align-items: center;
  justify-content: center;
}

.os-label {
  font-size: 11px;
  letter-spacing: 2px;
  color: var(--text-muted);
}

.os-title {
  font-size: 22px;
  color: var(--text-primary);
  margin: 2px 0 6px 0;
}

.os-badges {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.cyber-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 2px;
  font-family: var(--font-data);
}

.cyber-badge.cyan {
  background: rgba(220, 38, 38, 0.12);
  color: var(--rdr-crimson);
  border: 1px solid rgba(220, 38, 38, 0.35);
}

.cyber-badge.muted {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-secondary);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.cyber-badge.primary-tag {
  background: rgba(217, 119, 6, 0.12);
  color: var(--rdr-amber);
  border: 1px solid rgba(217, 119, 6, 0.35);
}

/* Managers Grid */
.managers-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 16px;
}

.manager-card {
  padding: 16px 18px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 14px;
}

.manager-card:hover {
  border-color: rgba(220, 38, 38, 0.35);
}

.manager-card.active-card {
  border-color: var(--rdr-crimson);
  background: #181b24;
  box-shadow: 0 4px 16px rgba(220, 38, 38, 0.15);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.mgr-title-box {
  display: flex;
  align-items: center;
  gap: 10px;
}

.mgr-icon {
  color: var(--rdr-crimson);
}

.mgr-name {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 1px;
  color: #fff;
}

.mgr-badge {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 2px;
  letter-spacing: 0.5px;
}

.badge-primary {
  background: rgba(220, 38, 38, 0.12);
  color: var(--rdr-crimson);
  border: 1px solid var(--border-crimson);
}

.badge-universal {
  background: rgba(217, 119, 6, 0.12);
  color: var(--rdr-amber);
  border: 1px solid rgba(217, 119, 6, 0.35);
}

.card-metrics {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.metric-row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
}

.path-row {
  font-size: 11px;
  border-top: 1px dashed rgba(255, 255, 255, 0.08);
  padding-top: 6px;
  margin-top: 2px;
}

.metric-label {
  color: var(--text-secondary);
}

.metric-val.highlight {
  color: var(--rdr-crimson);
  font-weight: 700;
}

.card-glow-bar {
  height: 2px;
  width: 100%;
  background: transparent;
  transition: all 0.2s ease;
}

.card-glow-bar.bar-active {
  background: linear-gradient(90deg, transparent, var(--rdr-crimson), transparent);
}

/* Workspace */
.workspace-card {
  padding: 20px;
}

.workspace-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: 16px;
  margin-bottom: 16px;
}

.tab-controls {
  display: flex;
  gap: 8px;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border-subtle);
  color: var(--text-secondary);
  padding: 6px 12px;
  border-radius: 3px;
  font-family: var(--font-header);
  font-size: 13px;
  letter-spacing: 0.5px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.tab-btn:hover {
  color: #fff;
  border-color: rgba(220, 38, 38, 0.35);
}

.tab-btn.active {
  background: rgba(220, 38, 38, 0.12);
  color: #ffffff;
  border-color: var(--rdr-crimson);
}

.counter-pill {
  font-family: var(--font-data);
  font-size: 11px;
  background: rgba(220, 38, 38, 0.15);
  color: var(--rdr-crimson);
  padding: 1px 6px;
  border-radius: 10px;
}

.counter-pill.warn {
  background: rgba(217, 119, 6, 0.15);
  color: var(--rdr-amber);
}

.search-box {
  position: relative;
  min-width: 280px;
}

.search-box.small {
  min-width: 200px;
}

.search-icon {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
}

.updates-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.cyber-input {
  width: 100%;
  background: #11131a;
  border: 1px solid var(--border-subtle);
  border-radius: 3px;
  padding: 7px 12px 7px 32px;
  color: var(--text-primary);
  font-family: var(--font-data);
  font-size: 12px;
  outline: none;
  transition: border-color 0.15s ease;
}

.cyber-input:focus {
  border-color: var(--rdr-crimson);
}

.cyber-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #181b24;
  border: 1px solid var(--border-crimson);
  color: #ffffff;
  padding: 6px 12px;
  border-radius: 3px;
  font-family: var(--font-header);
  font-size: 12px;
  letter-spacing: 0.5px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.cyber-btn:hover:not(:disabled) {
  background: var(--rdr-crimson);
  color: #ffffff;
}

.cyber-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Table */
.table-responsive {
  overflow-x: auto;
}

.cyber-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 13px;
}

[dir="rtl"] .cyber-table {
  text-align: right;
}

.cyber-table th {
  padding: 9px 12px;
  color: var(--text-muted);
  font-weight: 600;
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--border-subtle);
  font-size: 11px;
  font-family: var(--font-data);
}

.cyber-table td {
  padding: 8px 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}

.cyber-table tr:hover td {
  background: rgba(220, 38, 38, 0.03);
}

.pkg-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pkg-icon {
  color: var(--rdr-crimson);
  opacity: 0.8;
}

.pkg-title {
  font-weight: 600;
  color: #fff;
}

.pkg-ver {
  color: var(--text-primary);
}

.engine-tag {
  font-size: 11px;
  font-family: var(--font-data);
  padding: 2px 6px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 2px;
  color: var(--text-secondary);
}

.new-pill {
  font-size: 12px;
  background: rgba(220, 38, 38, 0.12);
  color: var(--rdr-crimson);
  padding: 2px 8px;
  border-radius: 2px;
  border: 1px solid var(--border-crimson);
}

.old-ver {
  color: var(--text-muted);
  text-decoration: line-through;
}

.new-ver {
  color: var(--rdr-crimson);
  font-weight: 700;
}

/* Enhanced Pagination Footer */
.pagination-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
  padding: 14px 4px 4px 4px;
  font-size: 13px;
  border-top: 1px solid var(--border-subtle);
  margin-top: 12px;
}

.pagination-meta {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.pagination-info {
  color: var(--text-muted);
  font-size: 12px;
}

.highlight {
  color: var(--rdr-crimson);
  font-weight: 600;
}

.page-size-picker {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-secondary);
}

.cyber-select {
  background: #11131a;
  border: 1px solid var(--border-subtle);
  color: var(--text-primary);
  border-radius: 3px;
  padding: 3px 6px;
  font-family: var(--font-data);
  font-size: 12px;
  outline: none;
  cursor: pointer;
}

.cyber-select:focus {
  border-color: var(--rdr-crimson);
}

.pagination-nav {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.nav-icon-btn {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border-subtle);
  color: var(--text-primary);
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 3px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.nav-icon-btn:hover:not(:disabled) {
  border-color: var(--rdr-crimson);
  color: var(--rdr-crimson);
  background: rgba(220, 38, 38, 0.08);
}

.nav-icon-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.page-pills {
  display: flex;
  align-items: center;
  gap: 4px;
}

.page-pill {
  min-width: 30px;
  height: 30px;
  padding: 0 6px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border-subtle);
  color: var(--text-secondary);
  font-family: var(--font-data);
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 3px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.page-pill:hover:not(:disabled) {
  border-color: var(--rdr-crimson);
  color: #fff;
}

.page-pill.pill-active {
  background: var(--rdr-crimson);
  border-color: var(--rdr-crimson-hover);
  color: #ffffff;
  font-weight: 700;
}

.page-pill.pill-dots {
  border: none;
  background: transparent;
  cursor: default;
  color: var(--text-muted);
  padding: 0 4px;
  min-width: 20px;
}

.jump-box {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-left: 8px;
  font-size: 12px;
  color: var(--text-secondary);
}

.jump-input {
  width: 48px;
  background: #11131a;
  border: 1px solid var(--border-subtle);
  border-radius: 3px;
  padding: 3px 6px;
  color: #fff;
  font-size: 12px;
  text-align: center;
  outline: none;
}

.jump-input:focus {
  border-color: var(--rdr-crimson);
}

.jump-btn {
  background: #181b24;
  border: 1px solid var(--border-crimson);
  color: var(--rdr-crimson);
  font-size: 11px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 3px;
  cursor: pointer;
  letter-spacing: 0.5px;
}

.jump-btn:hover {
  background: var(--rdr-crimson);
  color: #ffffff;
}

.pagination-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-current {
  font-size: 12px;
  color: var(--rdr-crimson);
  padding: 0 6px;
}

/* Loading & Empty states */
.loading-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 16px;
  gap: 12px;
  color: var(--text-secondary);
}

.empty-state.clean h3 {
  color: #fff;
  margin-top: 4px;
}

.mono {
  font-family: var(--font-data);
}

.text-dim {
  color: var(--text-secondary);
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
