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
  Server,
  Trash2,
  PlusCircle,
  Zap,
  Sparkles,
  X,
  Copy,
  Check,
  ShieldAlert,
  Clock,
  DownloadCloud
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

// Operation execution & modals state
const executingAction = ref(false)
const actionTitle = ref('')
const showOutputModal = ref(false)
const outputResult = ref(null)
const copiedOutput = ref(false)

const showInstallModal = ref(false)
const installPackageName = ref('')

const showConfirmModal = ref(false)
const confirmData = ref({
  title: '',
  message: '',
  action: null,
  packageName: '',
  hasPurge: false,
  purge: false,
  btnText: '',
  isDanger: false
})

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

// ----------------------------------------------------
// Package Lifecycle Mutation Operations
// ----------------------------------------------------

const runOperation = async (title, apiCall) => {
  executingAction.value = true
  actionTitle.value = title
  outputResult.value = null
  showOutputModal.value = true
  copiedOutput.value = false

  try {
    const res = await apiCall()
    outputResult.value = res.data?.data || {
      success: true,
      action: title,
      manager: selectedManager.value,
      command: 'Executed successfully',
      output: res.data?.message || 'Operation completed.',
      execution_time_ms: 0
    }
    // Refresh background state
    await fetchOverview(true)
    if (activeTab.value === 'installed') {
      fetchPackages()
    } else {
      fetchUpdates()
    }
  } catch (err) {
    const errorData = err.response?.data?.data
    outputResult.value = errorData || {
      success: false,
      action: title,
      manager: selectedManager.value,
      command: 'Failed',
      output: err.response?.data?.message || err.message || 'Operation encountered an error',
      execution_time_ms: 0
    }
  } finally {
    executingAction.value = false
  }
}

// Clean Cache Action
const triggerCleanCache = () => {
  confirmData.value = {
    title: t('packages.clean_cache') || 'CLEAN CACHE',
    message: t('packages.clean_cache_desc') || 'Purge local cache and unused package archives to free up disk space.',
    action: () => {
      showConfirmModal.value = false
      runOperation(t('packages.clean_cache') || 'Clean Cache', () => {
        return api.post(`/packages/${selectedManager.value}/clean-cache`)
      })
    },
    packageName: '',
    hasPurge: false,
    purge: false,
    btnText: t('packages.clean_cache') || 'CLEAN CACHE',
    isDanger: false
  }
  showConfirmModal.value = true
}

// Refresh Repositories Action
const triggerRefreshRepos = () => {
  runOperation(t('packages.refresh_repos') || 'Refresh Repositories', () => {
    return api.post(`/packages/${selectedManager.value}/refresh`)
  })
}

// Upgrade System Action
const triggerUpgradeSystem = () => {
  confirmData.value = {
    title: t('packages.upgrade_system') || 'UPGRADE SYSTEM',
    message: t('packages.upgrade_system_desc') || 'Upgrade all installed packages with available upstream updates.',
    action: () => {
      showConfirmModal.value = false
      runOperation(t('packages.upgrade_system') || 'Upgrade System', () => {
        return api.post(`/packages/${selectedManager.value}/upgrade-system`)
      })
    },
    packageName: '',
    hasPurge: false,
    purge: false,
    btnText: t('packages.upgrade_system') || 'UPGRADE SYSTEM',
    isDanger: false
  }
  showConfirmModal.value = true
}

// Open Install Modal
const openInstallModal = () => {
  installPackageName.value = ''
  showInstallModal.value = true
}

// Submit Install Package
const submitInstallPackage = () => {
  const pkg = installPackageName.value.trim()
  if (!pkg) return
  showInstallModal.value = false
  runOperation(`${t('packages.install_pkg') || 'Install Package'}: ${pkg}`, () => {
    return api.post(`/packages/${selectedManager.value}/install`, { package: pkg })
  })
}

// Trigger Remove Package
const triggerRemovePackage = (pkgName) => {
  confirmData.value = {
    title: t('packages.remove_pkg_title') || 'Uninstall Package',
    message: (t('packages.remove_pkg_confirm', { name: pkgName }) || `Are you sure you want to remove package [${pkgName}]?`),
    packageName: pkgName,
    hasPurge: selectedManager.value === 'apt' || selectedManager.value === 'pacman' || selectedManager.value === 'apk' || selectedManager.value === 'snap',
    purge: false,
    action: () => {
      const purgeVal = confirmData.value.purge
      showConfirmModal.value = false
      runOperation(`${t('packages.remove_pkg') || 'Uninstall'}: ${pkgName}`, () => {
        return api.post(`/packages/${selectedManager.value}/remove`, { 
          package: pkgName, 
          purge: purgeVal 
        })
      })
    },
    btnText: t('packages.remove_pkg') || 'UNINSTALL',
    isDanger: true
  }
  showConfirmModal.value = true
}

// Trigger Upgrade Single Package
const triggerUpgradePackage = (pkgName) => {
  runOperation(`${t('packages.upgrade_pkg') || 'Upgrade'}: ${pkgName}`, () => {
    return api.post(`/packages/${selectedManager.value}/upgrade-package`, { package: pkgName })
  })
}

// Copy Console Output
const copyConsoleOutput = async () => {
  if (!outputResult.value) return
  const text = `Action: ${outputResult.value.action}\nCommand: ${outputResult.value.command}\n\nOutput:\n${outputResult.value.output}`
  try {
    await navigator.clipboard.writeText(text)
    copiedOutput.value = true
    setTimeout(() => {
      copiedOutput.value = false
    }, 2000)
  } catch (e) {
    // clipboard error
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

    <!-- Manager Maintenance & Action Toolbar -->
    <div class="tron-card manager-actions-bar" v-if="selectedManager">
      <div class="bar-left">
        <div class="engine-badge-pill">
          <span class="dot-indicator"></span>
          <span class="engine-name">{{ selectedManager.toUpperCase() }}</span>
          <span class="engine-mode">{{ $t('packages.manage_packages') || 'ENGINE OPERATIONS' }}</span>
        </div>
      </div>

      <div class="bar-right">
        <!-- Clean Cache Button -->
        <button 
          @click="triggerCleanCache" 
          class="action-pill-btn clean-btn"
          :title="$t('packages.clean_cache_desc') || 'Purge package cache'"
          :disabled="executingAction"
        >
          <Trash2 :size="15" />
          <span>{{ $t('packages.clean_cache') || 'CLEAN CACHE' }}</span>
        </button>

        <!-- Refresh Repositories Button -->
        <button 
          @click="triggerRefreshRepos" 
          class="action-pill-btn refresh-btn"
          :title="$t('packages.refresh_repos_desc') || 'Fetch latest metadata'"
          :disabled="executingAction"
        >
          <RefreshCw :size="15" :class="{ 'spin': executingAction && actionTitle.includes('Refresh') }" />
          <span>{{ $t('packages.refresh_repos') || 'REFRESH REPOS' }}</span>
        </button>

        <!-- Upgrade System Button -->
        <button 
          @click="triggerUpgradeSystem" 
          class="action-pill-btn upgrade-btn"
          :title="$t('packages.upgrade_system_desc') || 'Upgrade all packages'"
          :disabled="executingAction"
        >
          <Zap :size="15" />
          <span>{{ $t('packages.upgrade_system') || 'UPGRADE SYSTEM' }}</span>
        </button>

        <!-- Install Package Button -->
        <button 
          @click="openInstallModal" 
          class="action-pill-btn install-btn"
          :disabled="executingAction"
        >
          <PlusCircle :size="15" />
          <span>{{ $t('packages.install_pkg') || 'INSTALL PACKAGE' }}</span>
        </button>
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
            v-if="updatesList.length > 0"
            @click="triggerUpgradeSystem" 
            class="cyber-btn upgrade-all-btn"
            :disabled="executingAction"
          >
            <Zap :size="15" />
            <span>{{ $t('packages.upgrade_all') || 'UPGRADE ALL UPDATES' }}</span>
          </button>

          <button 
            @click="fetchUpdates" 
            class="cyber-btn"
            :disabled="loadingUpdates || executingAction"
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
                <th style="text-align: right; width: 140px;">{{ $t('packages.actions') || 'ACTIONS' }}</th>
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
                <td style="text-align: right;">
                  <div class="row-actions">
                    <button 
                      @click="triggerUpgradePackage(pkg.name)" 
                      class="row-action-btn upgrade-row-btn"
                      :title="$t('packages.upgrade_pkg') || 'Upgrade'"
                      :disabled="executingAction"
                    >
                      <ArrowUpCircle :size="13" />
                      <span>{{ $t('packages.upgrade_pkg') || 'UPGRADE' }}</span>
                    </button>
                    <button 
                      @click="triggerRemovePackage(pkg.name)" 
                      class="row-action-btn remove-row-btn"
                      :title="$t('packages.remove_pkg') || 'Uninstall'"
                      :disabled="executingAction"
                    >
                      <Trash2 :size="13" />
                      <span>{{ $t('packages.remove_pkg') || 'REMOVE' }}</span>
                    </button>
                  </div>
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
                <th style="text-align: right; width: 120px;">{{ $t('packages.actions') || 'ACTIONS' }}</th>
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
                <td style="text-align: right;">
                  <button 
                    @click="triggerUpgradePackage(up.name)" 
                    class="row-action-btn upgrade-row-btn"
                    :disabled="executingAction"
                  >
                    <ArrowUpCircle :size="13" />
                    <span>{{ $t('packages.upgrade_pkg') || 'UPGRADE' }}</span>
                  </button>
                </td>
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

    <!-- Modal 1: Install Package Modal -->
    <div class="modal-backdrop" v-if="showInstallModal" @click.self="showInstallModal = false">
      <div class="modal-card tron-card">
        <div class="modal-header">
          <div class="modal-title-box">
            <PlusCircle class="modal-title-icon glow-cyan" :size="20" />
            <h3 class="modal-title">{{ $t('packages.install_pkg_title') || 'Install New Package' }}</h3>
          </div>
          <button @click="showInstallModal = false" class="modal-close-btn">
            <X :size="18" />
          </button>
        </div>

        <div class="modal-body">
          <p class="modal-desc">
            Target Package Manager: <strong class="highlight mono">{{ selectedManager.toUpperCase() }}</strong>
          </p>
          <div class="modal-input-group">
            <label class="input-label">{{ $t('packages.package_name') || 'PACKAGE NAME' }}</label>
            <input 
              v-model="installPackageName" 
              @keyup.enter="submitInstallPackage"
              type="text" 
              :placeholder="$t('packages.install_pkg_placeholder') || 'Enter package name (e.g. htop, nginx, curl)...'"
              class="cyber-input modal-input mono"
              autofocus
            />
          </div>
        </div>

        <div class="modal-footer">
          <button @click="showInstallModal = false" class="btn-cancel">
            {{ $t('packages.cancel') || 'CANCEL' }}
          </button>
          <button 
            @click="submitInstallPackage" 
            class="btn-primary-action"
            :disabled="!installPackageName.trim()"
          >
            <PlusCircle :size="15" />
            <span>{{ $t('packages.install_submit') || 'INSTALL' }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Modal 2: Confirmation Modal (Uninstall, Clean, Upgrade) -->
    <div class="modal-backdrop" v-if="showConfirmModal" @click.self="showConfirmModal = false">
      <div class="modal-card tron-card" :class="{ 'danger-card': confirmData.isDanger }">
        <div class="modal-header">
          <div class="modal-title-box">
            <ShieldAlert v-if="confirmData.isDanger" class="modal-title-icon glow-orange" :size="20" />
            <Sparkles v-else class="modal-title-icon glow-cyan" :size="20" />
            <h3 class="modal-title">{{ confirmData.title }}</h3>
          </div>
          <button @click="showConfirmModal = false" class="modal-close-btn">
            <X :size="18" />
          </button>
        </div>

        <div class="modal-body">
          <p class="modal-confirm-msg">{{ confirmData.message }}</p>
          
          <div v-if="confirmData.hasPurge" class="purge-checkbox-box">
            <label class="custom-checkbox-label">
              <input type="checkbox" v-model="confirmData.purge" class="custom-checkbox" />
              <span>{{ $t('packages.purge_option') || 'Purge configuration files and orphaned dependencies' }}</span>
            </label>
          </div>
        </div>

        <div class="modal-footer">
          <button @click="showConfirmModal = false" class="btn-cancel">
            {{ $t('packages.cancel') || 'CANCEL' }}
          </button>
          <button 
            @click="confirmData.action" 
            class="btn-primary-action"
            :class="{ 'danger-btn': confirmData.isDanger }"
          >
            <span>{{ confirmData.btnText || ($t('packages.confirm') || 'CONFIRM') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Modal 3: Command Output & Execution Modal -->
    <div class="modal-backdrop" v-if="showOutputModal" @click.self="!executingAction ? (showOutputModal = false) : null">
      <div class="modal-card tron-card terminal-modal">
        <div class="modal-header">
          <div class="modal-title-box">
            <Terminal class="modal-title-icon glow-cyan" :size="20" />
            <h3 class="modal-title">{{ actionTitle }}</h3>
          </div>
          <button v-if="!executingAction" @click="showOutputModal = false" class="modal-close-btn">
            <X :size="18" />
          </button>
        </div>

        <div class="modal-body terminal-body">
          <!-- Execution Progress -->
          <div v-if="executingAction" class="terminal-running-state">
            <RefreshCw :size="32" class="spin glow-cyan" />
            <div class="running-info">
              <h4>{{ $t('packages.executing') || 'EXECUTING OPERATION...' }}</h4>
              <p class="mono text-dim">Invoking {{ selectedManager.toUpperCase() }} engine on host system...</p>
            </div>
          </div>

          <!-- Execution Complete Result -->
          <div v-else-if="outputResult" class="terminal-output-container">
            <!-- Header Meta -->
            <div class="output-meta-row">
              <div class="meta-status">
                <span 
                  class="status-pill"
                  :class="outputResult.success ? 'status-success' : 'status-failed'"
                >
                  <CheckCircle2 v-if="outputResult.success" :size="14" />
                  <AlertCircle v-else :size="14" />
                  {{ outputResult.success ? ($t('packages.operation_success') || 'SUCCESS') : ($t('packages.operation_failed') || 'FAILED') }}
                </span>
                <span class="meta-time mono" v-if="outputResult.execution_time_ms">
                  <Clock :size="13" />
                  {{ outputResult.execution_time_ms }} ms
                </span>
              </div>

              <button @click="copyConsoleOutput" class="copy-output-btn">
                <Check v-if="copiedOutput" :size="14" class="glow-cyan" />
                <Copy v-else :size="14" />
                <span>{{ copiedOutput ? ($t('packages.copied') || 'COPIED') : 'COPY' }}</span>
              </button>
            </div>

            <!-- Command line -->
            <div class="cmd-preview-box mono">
              <span class="cmd-prompt">$</span>
              <span class="cmd-text">{{ outputResult.command }}</span>
            </div>

            <!-- Terminal Output Log -->
            <div class="terminal-logs-window mono">
              <pre>{{ outputResult.output || 'No output returned.' }}</pre>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button 
            v-if="!executingAction" 
            @click="showOutputModal = false" 
            class="btn-primary-action"
          >
            {{ $t('packages.close') || 'CLOSE' }}
          </button>
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
}

/* Card System - Charcoal & Industrial Outlaw Tech */
.tron-card {
  background: var(--bg-surface, #141721);
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.08));
  border-radius: 4px;
  position: relative;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.35);
}

/* OS Banner */
.os-banner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  background: linear-gradient(135deg, rgba(20, 23, 33, 0.95) 0%, rgba(30, 20, 25, 0.85) 100%);
  border-left: 4px solid var(--rdr-crimson, #dc2626);
}

.os-info {
  display: flex;
  align-items: center;
  gap: 18px;
}

.os-avatar {
  width: 52px;
  height: 52px;
  background: rgba(220, 38, 38, 0.12);
  border: 1px solid rgba(220, 38, 38, 0.3);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.os-label {
  font-size: 11px;
  letter-spacing: 1.5px;
  color: var(--text-muted, #71717a);
  font-weight: 700;
  margin-bottom: 2px;
}

.os-title {
  font-size: 20px;
  font-weight: 700;
  color: #ffffff;
  margin: 0 0 8px 0;
  letter-spacing: 0.5px;
}

.os-badges {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.cyber-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 3px;
  font-family: var(--font-data, monospace);
  font-weight: 600;
  letter-spacing: 0.5px;
}

.cyber-badge.cyan {
  background: rgba(220, 38, 38, 0.15);
  color: var(--rdr-crimson, #dc2626);
  border: 1px solid rgba(220, 38, 38, 0.3);
}

.cyber-badge.muted {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-secondary, #a1a1aa);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.cyber-badge.primary-tag {
  background: rgba(217, 119, 6, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(217, 119, 6, 0.3);
}

.cyber-btn {
  background: #1c202d;
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.1));
  color: #ffffff;
  padding: 8px 16px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  letter-spacing: 0.8px;
  transition: all 0.2s ease;
}

.cyber-btn:hover:not(:disabled) {
  border-color: var(--rdr-crimson, #dc2626);
  background: rgba(220, 38, 38, 0.1);
  color: #ffffff;
}

.cyber-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Managers Grid */
.managers-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}

.manager-card {
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #141721;
}

.manager-card:hover {
  transform: translateY(-2px);
  border-color: rgba(220, 38, 38, 0.4);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.45);
}

.manager-card.active-card {
  border-color: var(--rdr-crimson, #dc2626);
  background: linear-gradient(180deg, #171a26 0%, #1c1822 100%);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.mgr-title-box {
  display: flex;
  align-items: center;
  gap: 8px;
}

.mgr-icon {
  color: var(--rdr-crimson, #dc2626);
}

.mgr-name {
  font-weight: 700;
  font-size: 15px;
  color: #fff;
  letter-spacing: 0.8px;
}

.mgr-badge {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 3px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.badge-primary {
  background: rgba(220, 38, 38, 0.2);
  color: var(--rdr-crimson, #dc2626);
  border: 1px solid rgba(220, 38, 38, 0.4);
}

.badge-universal {
  background: rgba(217, 119, 6, 0.2);
  color: #f59e0b;
  border: 1px solid rgba(217, 119, 6, 0.4);
}

.card-metrics {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.metric-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
}

.metric-label {
  color: var(--text-muted, #71717a);
}

.metric-val {
  color: #ffffff;
}

.metric-val.highlight {
  color: var(--rdr-crimson, #dc2626);
  font-weight: 700;
}

.path-row {
  border-top: 1px solid rgba(255, 255, 255, 0.05);
  padding-top: 6px;
  margin-top: 2px;
}

.card-glow-bar {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: transparent;
  transition: all 0.2s ease;
}

.card-glow-bar.bar-active {
  background: var(--rdr-crimson, #dc2626);
  box-shadow: 0 0 10px var(--rdr-crimson, #dc2626);
}

/* Manager Maintenance Toolbar */
.manager-actions-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: #171b26;
  border-left: 3px solid #f59e0b;
  flex-wrap: wrap;
  gap: 12px;
}

.engine-badge-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(0, 0, 0, 0.3);
  padding: 6px 12px;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.06);
}

.dot-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #10b981;
  box-shadow: 0 0 8px #10b981;
}

.engine-name {
  font-weight: 800;
  color: #ffffff;
  letter-spacing: 0.8px;
  font-size: 13px;
}

.engine-mode {
  color: var(--text-muted, #71717a);
  font-size: 11px;
  letter-spacing: 0.5px;
}

.bar-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.action-pill-btn {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.1));
  color: #e4e4e7;
  padding: 6px 12px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-pill-btn:hover:not(:disabled) {
  transform: translateY(-1px);
}

.clean-btn:hover:not(:disabled) {
  border-color: #f59e0b;
  color: #f59e0b;
  background: rgba(245, 158, 11, 0.08);
}

.refresh-btn:hover:not(:disabled) {
  border-color: #3b82f6;
  color: #60a5fa;
  background: rgba(59, 130, 246, 0.08);
}

.upgrade-btn {
  background: rgba(220, 38, 38, 0.12);
  border-color: rgba(220, 38, 38, 0.3);
  color: #ffffff;
}

.upgrade-btn:hover:not(:disabled) {
  background: var(--rdr-crimson, #dc2626);
  border-color: var(--rdr-crimson, #dc2626);
  color: #ffffff;
}

.install-btn {
  background: rgba(16, 185, 129, 0.12);
  border-color: rgba(16, 185, 129, 0.3);
  color: #34d399;
}

.install-btn:hover:not(:disabled) {
  background: #10b981;
  border-color: #10b981;
  color: #ffffff;
}

.action-pill-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Workspace Card */
.workspace-card {
  padding: 20px;
}

.workspace-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  padding-bottom: 16px;
}

.tab-controls {
  display: flex;
  gap: 8px;
}

.tab-btn {
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-muted, #71717a);
  padding: 8px 16px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  letter-spacing: 0.5px;
  transition: all 0.2s ease;
}

.tab-btn:hover {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.04);
}

.tab-btn.active {
  background: rgba(220, 38, 38, 0.15);
  border-color: rgba(220, 38, 38, 0.3);
  color: #ffffff;
}

.counter-pill {
  background: rgba(255, 255, 255, 0.1);
  color: #ffffff;
  font-size: 11px;
  font-family: var(--font-data, monospace);
  padding: 2px 6px;
  border-radius: 10px;
}

.counter-pill.warn {
  background: #f59e0b;
  color: #000;
  font-weight: 700;
}

.search-box {
  display: flex;
  align-items: center;
  background: #0d0f17;
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.1));
  border-radius: 4px;
  padding: 0 12px;
  width: 320px;
  transition: all 0.2s ease;
}

.search-box:focus-within {
  border-color: var(--rdr-crimson, #dc2626);
  box-shadow: 0 0 10px rgba(220, 38, 38, 0.2);
}

.search-box.small {
  width: 260px;
}

.search-icon {
  color: var(--text-muted, #71717a);
  margin-right: 8px;
}

.cyber-input {
  background: transparent;
  border: none;
  color: #ffffff;
  font-size: 13px;
  padding: 8px 0;
  width: 100%;
  outline: none;
}

.updates-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.upgrade-all-btn {
  background: var(--rdr-crimson, #dc2626);
  border-color: var(--rdr-crimson, #dc2626);
  color: #ffffff;
}

.upgrade-all-btn:hover:not(:disabled) {
  background: #b91c1c;
  border-color: #b91c1c;
}

/* Cyber Table */
.table-responsive {
  overflow-x: auto;
}

.cyber-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 13px;
}

.cyber-table th {
  padding: 12px 14px;
  color: var(--text-muted, #71717a);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 1px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.cyber-table td {
  padding: 12px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  vertical-align: middle;
}

.cyber-table tbody tr:hover {
  background: rgba(255, 255, 255, 0.02);
}

.pkg-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.pkg-icon {
  color: var(--text-muted, #71717a);
}

.pkg-title {
  font-weight: 600;
  color: #ffffff;
}

.pkg-ver {
  color: #a1a1aa;
}

.engine-tag {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 10px;
  color: var(--text-secondary, #a1a1aa);
  font-family: var(--font-data, monospace);
}

.old-ver {
  color: var(--text-muted, #71717a);
}

.new-ver .new-pill {
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
}

/* Row Action Buttons */
.row-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
}

.row-action-btn {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.1));
  padding: 4px 8px;
  border-radius: 3px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.5px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.upgrade-row-btn {
  color: #60a5fa;
  border-color: rgba(96, 165, 250, 0.3);
}

.upgrade-row-btn:hover:not(:disabled) {
  background: #3b82f6;
  color: #ffffff;
  border-color: #3b82f6;
}

.remove-row-btn {
  color: #f87171;
  border-color: rgba(248, 113, 113, 0.3);
}

.remove-row-btn:hover:not(:disabled) {
  background: #dc2626;
  color: #ffffff;
  border-color: #dc2626;
}

.row-action-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

/* Pagination Footer */
.pagination-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0 4px 0;
  flex-wrap: wrap;
  gap: 16px;
}

.pagination-meta {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.pagination-info {
  color: var(--text-muted, #71717a);
  font-size: 12px;
}

.highlight {
  color: var(--rdr-crimson, #dc2626);
  font-weight: 600;
}

.page-size-picker {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-secondary, #a1a1aa);
}

.cyber-select {
  background: #11131a;
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.1));
  color: var(--text-primary, #ffffff);
  border-radius: 3px;
  padding: 3px 6px;
  font-family: var(--font-data, monospace);
  font-size: 12px;
  outline: none;
  cursor: pointer;
}

.cyber-select:focus {
  border-color: var(--rdr-crimson, #dc2626);
}

.pagination-nav {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.nav-icon-btn {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.1));
  color: var(--text-primary, #ffffff);
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
  border-color: var(--rdr-crimson, #dc2626);
  color: var(--rdr-crimson, #dc2626);
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
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.1));
  color: var(--text-secondary, #a1a1aa);
  font-family: var(--font-data, monospace);
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 3px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.page-pill:hover:not(:disabled) {
  border-color: var(--rdr-crimson, #dc2626);
  color: #fff;
}

.page-pill.pill-active {
  background: var(--rdr-crimson, #dc2626);
  border-color: var(--rdr-crimson, #dc2626);
  color: #ffffff;
  font-weight: 700;
}

.page-pill.pill-dots {
  border: none;
  background: transparent;
  cursor: default;
  color: var(--text-muted, #71717a);
  padding: 0 4px;
  min-width: 20px;
}

.jump-box {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-left: 8px;
  font-size: 12px;
  color: var(--text-secondary, #a1a1aa);
}

.jump-input {
  width: 48px;
  background: #11131a;
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.1));
  border-radius: 3px;
  padding: 3px 6px;
  color: #fff;
  font-size: 12px;
  text-align: center;
  outline: none;
}

.jump-input:focus {
  border-color: var(--rdr-crimson, #dc2626);
}

.jump-btn {
  background: #181b24;
  border: 1px solid rgba(220, 38, 38, 0.4);
  color: var(--rdr-crimson, #dc2626);
  font-size: 11px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 3px;
  cursor: pointer;
  letter-spacing: 0.5px;
}

.jump-btn:hover {
  background: var(--rdr-crimson, #dc2626);
  color: #ffffff;
}

/* Modals */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 20px;
}

.modal-card {
  width: 100%;
  max-width: 520px;
  background: #141721;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-top: 3px solid var(--rdr-crimson, #dc2626);
  border-radius: 6px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.7);
  animation: modalIn 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-card.danger-card {
  border-top-color: #ef4444;
}

.modal-card.terminal-modal {
  max-width: 720px;
}

@keyframes modalIn {
  from { opacity: 0; transform: scale(0.96) translateY(-10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.modal-title-box {
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-title {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #ffffff;
  letter-spacing: 0.5px;
}

.modal-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted, #71717a);
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 3px;
  transition: all 0.15s ease;
}

.modal-close-btn:hover {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.06);
}

.modal-body {
  padding: 20px;
}

.modal-desc {
  color: var(--text-secondary, #a1a1aa);
  font-size: 13px;
  margin: 0 0 16px 0;
}

.modal-confirm-msg {
  color: #e4e4e7;
  font-size: 14px;
  line-height: 1.5;
  margin: 0 0 16px 0;
}

.modal-input-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.input-label {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted, #71717a);
  letter-spacing: 1px;
}

.modal-input {
  background: #0d0f17;
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.12));
  border-radius: 4px;
  padding: 10px 14px;
  font-size: 14px;
}

.modal-input:focus {
  border-color: var(--rdr-crimson, #dc2626);
}

.purge-checkbox-box {
  background: rgba(239, 68, 68, 0.08);
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: 4px;
  padding: 10px 14px;
  margin-top: 12px;
}

.custom-checkbox-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  color: #fca5a5;
  cursor: pointer;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 20px;
  background: rgba(0, 0, 0, 0.2);
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.btn-cancel {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-secondary, #a1a1aa);
  padding: 8px 16px;
  border-radius: 3px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  letter-spacing: 0.5px;
}

.btn-cancel:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #fff;
}

.btn-primary-action {
  background: var(--rdr-crimson, #dc2626);
  border: 1px solid var(--rdr-crimson, #dc2626);
  color: #ffffff;
  padding: 8px 18px;
  border-radius: 3px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.5px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.15s ease;
}

.btn-primary-action:hover:not(:disabled) {
  background: #b91c1c;
  border-color: #b91c1c;
}

.btn-primary-action:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.danger-btn {
  background: #ef4444;
  border-color: #ef4444;
}

.danger-btn:hover:not(:disabled) {
  background: #dc2626;
  border-color: #dc2626;
}

/* Terminal Modal Window */
.terminal-running-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 16px;
  gap: 16px;
}

.running-info {
  text-align: center;
}

.running-info h4 {
  margin: 0 0 4px 0;
  color: #ffffff;
  font-size: 15px;
}

.terminal-output-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.output-meta-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.meta-status {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 800;
  padding: 3px 10px;
  border-radius: 3px;
  letter-spacing: 0.5px;
}

.status-success {
  background: rgba(16, 185, 129, 0.15);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.status-failed {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.meta-time {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--text-muted, #71717a);
}

.copy-output-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-secondary, #a1a1aa);
  padding: 4px 10px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.copy-output-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.cmd-preview-box {
  background: #090a10;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 4px;
  padding: 8px 12px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.cmd-prompt {
  color: #10b981;
  font-weight: 700;
}

.cmd-text {
  color: #f3f4f6;
  word-break: break-all;
}

.terminal-logs-window {
  background: #090a10;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 4px;
  padding: 14px;
  max-height: 280px;
  overflow-y: auto;
}

.terminal-logs-window pre {
  margin: 0;
  color: #d1d5db;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

/* Loading & Empty states */
.loading-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 16px;
  gap: 12px;
  color: var(--text-secondary, #a1a1aa);
}

.empty-state.clean h3 {
  color: #fff;
  margin-top: 4px;
}

.mono {
  font-family: var(--font-data, monospace);
}

.text-dim {
  color: var(--text-secondary, #a1a1aa);
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
