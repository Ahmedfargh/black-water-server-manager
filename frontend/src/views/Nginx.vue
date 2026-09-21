<script setup>
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Server,
  Activity,
  FileText,
  CheckCircle,
  AlertTriangle,
  XCircle,
  RefreshCw,
  Plus,
  Edit,
  Trash2,
  Power,
  Search,
  Sliders,
  ShieldCheck,
  ShieldAlert,
  ArrowUpRight,
  TrendingUp,
  Globe,
  Database,
  Terminal,
  Play
} from 'lucide-vue-next'
import api from '../api'
import { useToastStore } from '../stores/toast'

const { t } = useI18n()
const toast = useToastStore()

// State
const activeTab = ref('sites') // 'sites' | 'analytics' | 'logs' | 'editor'
const isLoading = ref(false)
const isReloading = ref(false)
const isTesting = ref(false)

// Overview
const overview = ref({
  is_installed: false,
  is_running: false,
  version: '',
  sites_count: 0,
  enabled_count: 0,
  disabled_count: 0,
  config_valid: true,
  config_output: ''
})

// Sites
const sites = ref([])
const showSiteModal = ref(false)
const editingFilename = ref('')
const siteContent = ref('')
const isSavingSite = ref(false)

// Logs
const logFiles = ref({ access_logs: [], error_logs: [] })
const selectedAccessFile = ref('')
const selectedErrorFile = ref('')
const accessLogs = ref([])
const errorLogs = ref([])
const logType = ref('access') // 'access' | 'error'
const accessSearch = ref('')
const errorSearch = ref('')
const statusFilter = ref(0)
const levelFilter = ref('')
const logLimit = ref(100)
const selectedLogEntry = ref(null)

// Analytics
const analytics = ref(null)
const isLoadingAnalytics = ref(false)

// Lifecycle
onMounted(async () => {
  await fetchOverview()
  await fetchSites()
  await fetchLogFiles()
  await fetchAnalytics()
})

// Fetch Overview
const fetchOverview = async () => {
  try {
    const res = await api.get('/nginx/overview')
    if (res.data?.success) {
      overview.value = res.data.data
    }
  } catch (err) {
    console.debug('Nginx overview fetch error:', err)
  }
}

// Fetch Sites
const fetchSites = async () => {
  isLoading.value = true
  try {
    const res = await api.get('/nginx/sites?include_content=true')
    if (res.data?.success) {
      sites.value = res.data.data || []
    }
  } catch (err) {
    toast.error('Failed to load Nginx sites: ' + (err.response?.data?.error || err.message))
  } finally {
    isLoading.value = false
  }
}

// Fetch Log Files List
const fetchLogFiles = async () => {
  try {
    const res = await api.get('/nginx/logs/files')
    if (res.data?.success) {
      logFiles.value = res.data.data
      if (logFiles.value.access_logs?.length > 0 && !selectedAccessFile.value) {
        selectedAccessFile.value = logFiles.value.access_logs[0]
      }
      if (logFiles.value.error_logs?.length > 0 && !selectedErrorFile.value) {
        selectedErrorFile.value = logFiles.value.error_logs[0]
      }
    }
  } catch (err) {
    console.debug('Log files fetch error:', err)
  }
}

// Fetch Access Logs
const fetchAccessLogs = async () => {
  isLoading.value = true
  try {
    const params = new URLSearchParams({
      file: selectedAccessFile.value || '',
      limit: logLimit.value,
      status: statusFilter.value,
      search: accessSearch.value
    })
    const res = await api.get(`/nginx/logs/access?${params.toString()}`)
    if (res.data?.success) {
      accessLogs.value = res.data.data || []
    }
  } catch (err) {
    toast.error('Failed to fetch access logs: ' + (err.response?.data?.error || err.message))
  } finally {
    isLoading.value = false
  }
}

// Fetch Error Logs
const fetchErrorLogs = async () => {
  isLoading.value = true
  try {
    const params = new URLSearchParams({
      file: selectedErrorFile.value || '',
      limit: logLimit.value,
      level: levelFilter.value,
      search: errorSearch.value
    })
    const res = await api.get(`/nginx/logs/error?${params.toString()}`)
    if (res.data?.success) {
      errorLogs.value = res.data.data || []
    }
  } catch (err) {
    toast.error('Failed to fetch error logs: ' + (err.response?.data?.error || err.message))
  } finally {
    isLoading.value = false
  }
}

// Fetch Analytics
const fetchAnalytics = async () => {
  isLoadingAnalytics.value = true
  try {
    const res = await api.get(`/nginx/logs/analytics?file=${encodeURIComponent(selectedAccessFile.value || '')}`)
    if (res.data?.success) {
      analytics.value = res.data.data
    }
  } catch (err) {
    console.debug('Analytics fetch error:', err)
  } finally {
    isLoadingAnalytics.value = false
  }
}

// Switch Tab
const setTab = (tab) => {
  activeTab.value = tab
  if (tab === 'logs') {
    if (logType.value === 'access') fetchAccessLogs()
    else fetchErrorLogs()
  } else if (tab === 'analytics') {
    fetchAnalytics()
  } else if (tab === 'sites') {
    fetchSites()
  }
}

// Actions
const testConfig = async () => {
  isTesting.value = true
  try {
    const res = await api.post('/nginx/test')
    if (res.data?.valid) {
      toast.success('Nginx Configuration Syntax OK')
    } else {
      toast.error('Nginx Syntax Error: ' + res.data?.output)
    }
    await fetchOverview()
  } catch (err) {
    toast.error('Config test failed: ' + (err.response?.data?.error || err.message))
  } finally {
    isTesting.value = false
  }
}

const reloadNginx = async () => {
  isReloading.value = true
  try {
    const res = await api.post('/nginx/reload')
    if (res.data?.success) {
      toast.success('Nginx reloaded successfully')
      await fetchOverview()
    }
  } catch (err) {
    toast.error('Reload failed: ' + (err.response?.data?.error || err.message))
  } finally {
    isReloading.value = false
  }
}

const toggleSite = async (site) => {
  try {
    const res = await api.post(`/nginx/sites/${encodeURIComponent(site.filename)}/toggle`, {
      enable: !site.is_enabled
    })
    if (res.data?.success) {
      toast.success(`Site ${site.filename} ${!site.is_enabled ? 'enabled' : 'disabled'}`)
      await fetchSites()
      await fetchOverview()
    }
  } catch (err) {
    toast.error('Toggle failed: ' + (err.response?.data?.error || err.message))
  }
}

const openNewSiteModal = () => {
  editingFilename.value = 'new-site.conf'
  siteContent.value = `server {
    listen 80;
    server_name example.com;

    root /var/www/html;
    index index.html;

    location / {
        try_files $uri $uri/ =404;
    }
}`
  showSiteModal.value = true
}

const openEditSiteModal = (site) => {
  editingFilename.value = site.filename
  siteContent.value = site.raw_content || ''
  showSiteModal.value = true
}

const saveSiteConfig = async () => {
  if (!editingFilename.value || !siteContent.value) return
  isSavingSite.value = true
  try {
    const res = await api.post('/nginx/sites', {
      filename: editingFilename.value,
      content: siteContent.value
    })
    if (res.data?.success) {
      toast.success('Site configuration saved & validated!')
      showSiteModal.value = false
      await fetchSites()
      await fetchOverview()
    }
  } catch (err) {
    toast.error('Save failed: ' + (err.response?.data?.error || err.message))
  } finally {
    isSavingSite.value = false
  }
}

const deleteSite = async (site) => {
  if (!confirm(`Are you sure you want to delete ${site.filename}?`)) return
  try {
    const res = await api.delete(`/nginx/sites/${encodeURIComponent(site.filename)}`)
    if (res.data?.success) {
      toast.success('Site deleted successfully')
      await fetchSites()
      await fetchOverview()
    }
  } catch (err) {
    toast.error('Delete failed: ' + (err.response?.data?.error || err.message))
  }
}

const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const getStatusBadgeClass = (status) => {
  if (status >= 200 && status < 300) return 'status-badge-2xx'
  if (status >= 300 && status < 400) return 'status-badge-3xx'
  if (status >= 400 && status < 500) return 'status-badge-4xx'
  if (status >= 500) return 'status-badge-5xx'
  return 'status-badge-default'
}

const getLogLevelClass = (level) => {
  const l = level?.toLowerCase()
  if (l === 'error' || l === 'crit' || l === 'alert' || l === 'emerg') return 'level-error'
  if (l === 'warn') return 'level-warn'
  if (l === 'notice' || l === 'info') return 'level-info'
  return 'level-default'
}
</script>

<template>
  <div class="nginx-view">
    <!-- Header Controls -->
    <div class="header-card">
      <div class="header-main">
        <div class="header-title-wrap">
          <div class="header-icon-box">
            <Server class="icon-server" />
          </div>
          <div>
            <h1 class="page-title">{{ t('nginx.title', 'Nginx Manager & Logs') }}</h1>
            <p class="page-subtitle">
              {{ overview.version || 'Nginx Web Server' }} •
              <span :class="overview.is_running ? 'text-green' : 'text-red'">
                {{ overview.is_running ? '● Active' : '○ Inactive' }}
              </span>
              <span v-if="overview.config_valid" class="config-tag valid">
                <CheckCircle class="w-3.5 h-3.5 inline" /> Syntax Valid
              </span>
              <span v-else class="config-tag invalid">
                <AlertTriangle class="w-3.5 h-3.5 inline" /> Syntax Error
              </span>
            </p>
          </div>
        </div>

        <div class="header-actions">
          <button @click="testConfig" :disabled="isTesting" class="action-btn secondary">
            <Play class="w-4 h-4" :class="{ 'animate-spin': isTesting }" />
            <span>Test Syntax</span>
          </button>
          <button @click="reloadNginx" :disabled="isReloading" class="action-btn primary">
            <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': isReloading }" />
            <span>Reload Nginx</span>
          </button>
        </div>
      </div>

      <!-- Quick Metrics Bar -->
      <div class="overview-metrics">
        <div class="metric-chip">
          <Globe class="chip-icon" />
          <div class="chip-data">
            <span class="chip-val">{{ overview.sites_count }}</span>
            <span class="chip-lbl">Virtual Hosts</span>
          </div>
        </div>
        <div class="metric-chip">
          <CheckCircle class="chip-icon text-green" />
          <div class="chip-data">
            <span class="chip-val">{{ overview.enabled_count }}</span>
            <span class="chip-lbl">Active Sites</span>
          </div>
        </div>
        <div class="metric-chip">
          <Power class="chip-icon text-muted" />
          <div class="chip-data">
            <span class="chip-val">{{ overview.disabled_count }}</span>
            <span class="chip-lbl">Disabled</span>
          </div>
        </div>
        <div class="metric-chip" v-if="analytics">
          <TrendingUp class="chip-icon text-amber" />
          <div class="chip-data">
            <span class="chip-val">{{ analytics.total_requests }}</span>
            <span class="chip-lbl">Parsed Requests</span>
          </div>
        </div>
      </div>

      <!-- Navigation Tabs -->
      <div class="nav-tabs">
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'sites' }"
          @click="setTab('sites')"
        >
          <Globe class="w-4 h-4" />
          <span>Virtual Hosts ({{ sites.length }})</span>
        </button>
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'analytics' }"
          @click="setTab('analytics')"
        >
          <TrendingUp class="w-4 h-4" />
          <span>Log Analytics</span>
        </button>
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'logs' }"
          @click="setTab('logs')"
        >
          <FileText class="w-4 h-4" />
          <span>Live Log Explorer</span>
        </button>
      </div>
    </div>

    <!-- TAB 1: SITES / VIRTUAL HOSTS -->
    <div v-if="activeTab === 'sites'" class="tab-content">
      <div class="section-toolbar">
        <div>
          <h2 class="section-heading">Nginx Virtual Hosts & Server Blocks</h2>
          <p class="section-desc">Manage site configs in <code>/etc/nginx/sites-available</code> & <code>/etc/nginx/conf.d</code></p>
        </div>
        <button @click="openNewSiteModal" class="action-btn primary">
          <Plus class="w-4 h-4" />
          <span>New Virtual Host</span>
        </button>
      </div>

      <div v-if="isLoading" class="loading-state">
        <RefreshCw class="animate-spin w-8 h-8" />
        <span>Parsing Nginx Configurations...</span>
      </div>

      <div v-else-if="sites.length === 0" class="empty-state">
        <Server class="w-12 h-12 text-muted" />
        <p>No Nginx virtual hosts discovered.</p>
        <button @click="openNewSiteModal" class="action-btn primary mt-4">Create First Site</button>
      </div>

      <div v-else class="sites-grid">
        <div
          v-for="site in sites"
          :key="site.filename"
          class="site-card"
          :class="{ 'site-enabled': site.is_enabled }"
        >
          <div class="site-card-header">
            <div class="site-title-box">
              <span class="status-indicator" :class="site.is_enabled ? 'active' : 'inactive'"></span>
              <h3 class="site-filename">{{ site.filename }}</h3>
            </div>
            <div class="site-actions">
              <button
                @click="toggleSite(site)"
                class="icon-btn"
                :title="site.is_enabled ? 'Disable Site' : 'Enable Site'"
                :class="site.is_enabled ? 'text-green' : 'text-muted'"
              >
                <Power class="w-4 h-4" />
              </button>
              <button
                @click="openEditSiteModal(site)"
                class="icon-btn text-cyan"
                title="Edit Configuration"
              >
                <Edit class="w-4 h-4" />
              </button>
              <button
                @click="deleteSite(site)"
                class="icon-btn text-red"
                title="Delete Configuration"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>

          <div class="site-card-body">
            <!-- Server Blocks Info -->
            <div v-for="(srv, srvIdx) in site.servers" :key="srvIdx" class="server-block-summary">
              <div class="server-block-domains">
                <Globe class="w-3.5 h-3.5 text-muted" />
                <span class="domain-text">
                  {{ srv.server_names?.length ? srv.server_names.join(', ') : 'Default Server' }}
                </span>
                <span v-if="srv.is_ssl" class="ssl-badge">
                  <ShieldCheck class="w-3 h-3" /> SSL
                </span>
              </div>

              <div class="server-block-details">
                <div class="detail-row">
                  <span class="detail-lbl">Listen:</span>
                  <span class="detail-val font-mono">{{ srv.listen_ports?.join(', ') || '80' }}</span>
                </div>
                <div v-if="srv.root" class="detail-row">
                  <span class="detail-lbl">Root:</span>
                  <span class="detail-val font-mono truncate">{{ srv.root }}</span>
                </div>
                <div v-if="srv.locations?.length" class="detail-row">
                  <span class="detail-lbl">Locations:</span>
                  <span class="detail-val font-mono">
                    {{ srv.locations.map(l => l.path).join(', ') }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Upstreams -->
            <div v-if="site.upstreams?.length" class="upstreams-badge">
              <Database class="w-3 h-3" /> Upstreams: {{ site.upstreams.join(', ') }}
            </div>
          </div>

          <div class="site-card-footer">
            <span class="path-text">{{ site.path }}</span>
            <span class="state-pill" :class="site.is_enabled ? 'enabled' : 'disabled'">
              {{ site.is_enabled ? 'ENABLED' : 'DISABLED' }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- TAB 2: LOG ANALYTICS -->
    <div v-if="activeTab === 'analytics'" class="tab-content">
      <div class="section-toolbar">
        <div>
          <h2 class="section-heading">Access Log Traffic & Performance Analytics</h2>
          <p class="section-desc">Computed metrics parsed from Nginx access logs</p>
        </div>
        <div class="flex gap-2">
          <select v-model="selectedAccessFile" @change="fetchAnalytics" class="form-select">
            <option v-for="f in logFiles.access_logs" :key="f" :value="f">{{ f }}</option>
          </select>
          <button @click="fetchAnalytics" class="action-btn secondary">
            <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': isLoadingAnalytics }" />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      <div v-if="isLoadingAnalytics" class="loading-state">
        <RefreshCw class="animate-spin w-8 h-8" />
        <span>Parsing & Aggregating Log Metrics...</span>
      </div>

      <div v-else-if="!analytics || analytics.total_requests === 0" class="empty-state">
        <FileText class="w-12 h-12 text-muted" />
        <p>No log data found in {{ selectedAccessFile || 'access.log' }}.</p>
      </div>

      <div v-else class="analytics-dashboard">
        <!-- KPI Cards Grid -->
        <div class="kpi-grid">
          <div class="kpi-card">
            <span class="kpi-lbl">Total Requests</span>
            <span class="kpi-val">{{ analytics.total_requests.toLocaleString() }}</span>
            <span class="kpi-sub">Processed hits</span>
          </div>
          <div class="kpi-card">
            <span class="kpi-lbl">Bandwidth Transferred</span>
            <span class="kpi-val">{{ formatBytes(analytics.total_bytes) }}</span>
            <span class="kpi-sub">Total body bytes</span>
          </div>
          <div class="kpi-card">
            <span class="kpi-lbl">Error Rate</span>
            <span class="kpi-val" :class="analytics.error_rate > 5 ? 'text-red' : 'text-green'">
              {{ analytics.error_rate.toFixed(2) }}%
            </span>
            <span class="kpi-sub">4xx & 5xx responses</span>
          </div>
          <div class="kpi-card">
            <span class="kpi-lbl">HTTP Status Breakdown</span>
            <div class="status-mini-bars">
              <span class="mini-bar green" title="2xx Success">2xx: {{ analytics.status_2xx }}</span>
              <span class="mini-bar blue" title="3xx Redirect">3xx: {{ analytics.status_3xx }}</span>
              <span class="mini-bar yellow" title="4xx Client Error">4xx: {{ analytics.status_4xx }}</span>
              <span class="mini-bar red" title="5xx Server Error">5xx: {{ analytics.status_5xx }}</span>
            </div>
          </div>
        </div>

        <!-- Tables: Top IPs & Top Paths -->
        <div class="analytics-tables-grid">
          <!-- Top IPs -->
          <div class="analytics-card">
            <h3 class="card-heading">Top Client IP Addresses</h3>
            <div class="table-container">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>Client IP</th>
                    <th class="text-right">Hits</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="ip in analytics.top_ips" :key="ip.item">
                    <td class="font-mono text-cyan">{{ ip.item }}</td>
                    <td class="text-right font-bold">{{ ip.count }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Top Paths -->
          <div class="analytics-card">
            <h3 class="card-heading">Top Requested Endpoints</h3>
            <div class="table-container">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>Path</th>
                    <th class="text-right">Requests</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="p in analytics.top_paths" :key="p.item">
                    <td class="font-mono truncate max-w-xs">{{ p.item }}</td>
                    <td class="text-right font-bold">{{ p.count }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- Top User Agents -->
        <div class="analytics-card mt-6" v-if="analytics.top_user_agents?.length">
          <h3 class="card-heading">Top User Agents</h3>
          <div class="table-container">
            <table class="data-table">
              <thead>
                <tr>
                  <th>User Agent</th>
                  <th class="text-right">Count</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="ua in analytics.top_user_agents" :key="ua.item">
                  <td class="font-mono text-xs text-muted truncate max-w-lg">{{ ua.item }}</td>
                  <td class="text-right font-bold">{{ ua.count }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- TAB 3: LIVE LOG EXPLORER -->
    <div v-if="activeTab === 'logs'" class="tab-content">
      <div class="logs-header-bar">
        <div class="log-type-selector">
          <button
            class="type-btn"
            :class="{ active: logType === 'access' }"
            @click="logType = 'access'; fetchAccessLogs()"
          >
            Access Logs ({{ accessLogs.length }})
          </button>
          <button
            class="type-btn"
            :class="{ active: logType === 'error' }"
            @click="logType = 'error'; fetchErrorLogs()"
          >
            Error Logs ({{ errorLogs.length }})
          </button>
        </div>

        <div class="logs-filter-bar">
          <!-- Access Log Filters -->
          <template v-if="logType === 'access'">
            <select v-model="selectedAccessFile" @change="fetchAccessLogs" class="form-select">
              <option v-for="f in logFiles.access_logs" :key="f" :value="f">{{ f }}</option>
            </select>
            <select v-model="statusFilter" @change="fetchAccessLogs" class="form-select">
              <option :value="0">All Status Codes</option>
              <option :value="200">200 OK</option>
              <option :value="301">301 Moved</option>
              <option :value="400">400 Bad Request</option>
              <option :value="401">401 Unauthorized</option>
              <option :value="403">403 Forbidden</option>
              <option :value="404">404 Not Found</option>
              <option :value="500">500 Server Error</option>
              <option :value="502">502 Bad Gateway</option>
            </select>
            <div class="search-input-wrap">
              <Search class="search-icon" />
              <input
                v-model="accessSearch"
                @keyup.enter="fetchAccessLogs"
                placeholder="Search IP, path, UA..."
                class="form-input search"
              />
            </div>
          </template>

          <!-- Error Log Filters -->
          <template v-else>
            <select v-model="selectedErrorFile" @change="fetchErrorLogs" class="form-select">
              <option v-for="f in logFiles.error_logs" :key="f" :value="f">{{ f }}</option>
            </select>
            <select v-model="levelFilter" @change="fetchErrorLogs" class="form-select">
              <option value="">All Levels</option>
              <option value="error">Error</option>
              <option value="warn">Warning</option>
              <option value="crit">Critical</option>
              <option value="notice">Notice</option>
            </select>
            <div class="search-input-wrap">
              <Search class="search-icon" />
              <input
                v-model="errorSearch"
                @keyup.enter="fetchErrorLogs"
                placeholder="Search message, IP..."
                class="form-input search"
              />
            </div>
          </template>

          <button @click="logType === 'access' ? fetchAccessLogs() : fetchErrorLogs()" class="action-btn secondary">
            <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': isLoading }" />
          </button>
        </div>
      </div>

      <!-- Access Logs Table -->
      <div v-if="logType === 'access'" class="log-table-wrap">
        <table class="log-table">
          <thead>
            <tr>
              <th>Status</th>
              <th>Method</th>
              <th>Path</th>
              <th>Client IP</th>
              <th>Bytes</th>
              <th>Timestamp</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="accessLogs.length === 0">
              <td colspan="7" class="text-center py-8 text-muted">No access log entries matching criteria.</td>
            </tr>
            <tr
              v-for="(entry, idx) in accessLogs"
              :key="idx"
              class="log-row"
              @click="selectedLogEntry = entry"
            >
              <td>
                <span class="status-badge" :class="getStatusBadgeClass(entry.status_code)">
                  {{ entry.status_code }}
                </span>
              </td>
              <td class="font-bold text-cyan">{{ entry.method }}</td>
              <td class="font-mono text-white truncate max-w-md">{{ entry.path }}</td>
              <td class="font-mono text-muted">{{ entry.remote_ip }}</td>
              <td class="font-mono text-xs">{{ formatBytes(entry.body_bytes) }}</td>
              <td class="text-xs text-muted whitespace-nowrap">{{ entry.raw_timestamp }}</td>
              <td class="text-right">
                <button class="inspect-btn">Inspect</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Error Logs Table -->
      <div v-else class="log-table-wrap">
        <table class="log-table">
          <thead>
            <tr>
              <th>Level</th>
              <th>Message</th>
              <th>Client IP</th>
              <th>Server</th>
              <th>Timestamp</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="errorLogs.length === 0">
              <td colspan="6" class="text-center py-8 text-muted">No error log entries matching criteria.</td>
            </tr>
            <tr
              v-for="(entry, idx) in errorLogs"
              :key="idx"
              class="log-row"
              @click="selectedLogEntry = entry"
            >
              <td>
                <span class="level-badge" :class="getLogLevelClass(entry.log_level)">
                  {{ entry.log_level?.toUpperCase() }}
                </span>
              </td>
              <td class="font-mono text-xs text-white truncate max-w-xl">{{ entry.message }}</td>
              <td class="font-mono text-muted">{{ entry.client_ip || '-' }}</td>
              <td class="font-mono text-xs text-muted">{{ entry.server || '-' }}</td>
              <td class="text-xs text-muted whitespace-nowrap">{{ entry.raw_timestamp }}</td>
              <td class="text-right">
                <button class="inspect-btn">Inspect</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- SITE CONFIG MODAL -->
    <div v-if="showSiteModal" class="modal-backdrop" @click.self="showSiteModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <div>
            <h3 class="modal-title">Nginx Virtual Host Editor</h3>
            <p class="modal-subtitle">File: {{ editingFilename }}</p>
          </div>
          <button @click="showSiteModal = false" class="icon-btn text-muted"><XCircle class="w-5 h-5" /></button>
        </div>

        <div class="modal-body">
          <div class="form-group mb-3">
            <label class="form-label">Filename (in /etc/nginx/sites-available)</label>
            <input v-model="editingFilename" class="form-input" placeholder="mysite.conf" />
          </div>

          <div class="form-group">
            <label class="form-label">Nginx Configuration Block</label>
            <textarea
              v-model="siteContent"
              class="form-textarea code-editor"
              rows="18"
              spellcheck="false"
            ></textarea>
          </div>
        </div>

        <div class="modal-footer">
          <button @click="showSiteModal = false" class="action-btn secondary">Cancel</button>
          <button @click="saveSiteConfig" :disabled="isSavingSite" class="action-btn primary">
            <CheckCircle class="w-4 h-4" />
            <span>{{ isSavingSite ? 'Testing & Saving...' : 'Save & Validate' }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- LOG INSPECT MODAL -->
    <div v-if="selectedLogEntry" class="modal-backdrop" @click.self="selectedLogEntry = null">
      <div class="modal-card log-inspect-card">
        <div class="modal-header">
          <h3 class="modal-title">Log Entry Inspector</h3>
          <button @click="selectedLogEntry = null" class="icon-btn text-muted"><XCircle class="w-5 h-5" /></button>
        </div>
        <div class="modal-body font-mono text-xs">
          <div class="inspect-grid">
            <div v-for="(val, key) in selectedLogEntry" :key="key" class="inspect-item">
              <span class="inspect-key">{{ key }}:</span>
              <span class="inspect-val">{{ val }}</span>
            </div>
          </div>
          <div class="raw-log-box mt-4">
            <span class="text-muted block mb-1">Raw Log Line:</span>
            <pre class="raw-pre">{{ selectedLogEntry.raw }}</pre>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.nginx-view {
  padding: 1.5rem;
  max-width: 1400px;
  margin: 0 auto;
}

/* Header */
.header-card {
  background: var(--bg-card);
  border: 1px solid var(--border-card);
  border-radius: 6px;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}
.header-main {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
  margin-bottom: 1.25rem;
}
.header-title-wrap {
  display: flex;
  align-items: center;
  gap: 1rem;
}
.header-icon-box {
  width: 48px;
  height: 48px;
  background: var(--rdr-crimson-dim);
  border: 1px solid var(--border-crimson);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.icon-server {
  width: 24px;
  height: 24px;
  color: var(--rdr-crimson);
}
.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.25rem;
}
.page-subtitle {
  font-size: 0.875rem;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.config-tag {
  font-size: 0.75rem;
  padding: 0.15rem 0.5rem;
  border-radius: 3px;
  font-weight: 600;
}
.config-tag.valid {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
}
.config-tag.invalid {
  background: rgba(220, 38, 38, 0.15);
  color: #ef4444;
  border: 1px solid rgba(220, 38, 38, 0.3);
}
.header-actions {
  display: flex;
  gap: 0.75rem;
}

/* Metrics Bar */
.overview-metrics {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 1rem;
  padding: 1rem 0;
  border-top: 1px solid var(--border-subtle);
  border-bottom: 1px solid var(--border-subtle);
  margin-bottom: 1.25rem;
}
.metric-chip {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
.chip-icon {
  width: 28px;
  height: 28px;
  color: var(--rdr-crimson);
}
.chip-data {
  display: flex;
  flex-direction: column;
}
.chip-val {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
  font-family: var(--font-data);
}
.chip-lbl {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: uppercase;
}

/* Navigation Tabs */
.nav-tabs {
  display: flex;
  gap: 0.5rem;
}
.tab-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 1.2rem;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 4px;
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}
.tab-btn:hover {
  background: var(--bg-card-hover);
  color: var(--text-primary);
}
.tab-btn.active {
  background: var(--rdr-crimson-dim);
  border-color: var(--border-crimson);
  color: var(--rdr-crimson-hover);
}

/* Section Toolbar */
.section-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
}
.section-heading {
  font-size: 1.15rem;
  color: var(--text-primary);
  margin-bottom: 0.2rem;
}
.section-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
}

/* Sites Grid */
.sites-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 1.25rem;
}
.site-card {
  background: var(--bg-card);
  border: 1px solid var(--border-card);
  border-radius: 6px;
  overflow: hidden;
  transition: transform 0.2s ease, border-color 0.2s ease;
}
.site-card:hover {
  border-color: rgba(255, 255, 255, 0.2);
}
.site-card.site-enabled {
  border-left: 3px solid #10b981;
}
.site-card-header {
  padding: 1rem 1.25rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--border-subtle);
}
.site-title-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.status-indicator.active { background: #10b981; box-shadow: 0 0 6px rgba(16, 185, 129, 0.6); }
.status-indicator.inactive { background: #64748b; }
.site-filename {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-primary);
}
.site-actions {
  display: flex;
  gap: 0.4rem;
}
.site-card-body {
  padding: 1rem 1.25rem;
}
.server-block-summary {
  margin-bottom: 0.75rem;
}
.server-block-domains {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.4rem;
}
.domain-text {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 0.9rem;
}
.ssl-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  font-size: 0.7rem;
  padding: 0.1rem 0.4rem;
  border-radius: 3px;
  font-weight: 700;
}
.server-block-details {
  font-size: 0.78rem;
  color: var(--text-muted);
}
.detail-row {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.2rem;
}
.detail-lbl {
  color: #94a3b8;
  min-width: 60px;
}
.upstreams-badge {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.75rem;
  color: #d97706;
  background: rgba(217, 119, 6, 0.1);
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
}
.site-card-footer {
  padding: 0.75rem 1.25rem;
  background: rgba(0, 0, 0, 0.2);
  border-top: 1px solid var(--border-subtle);
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.72rem;
}
.path-text {
  color: var(--text-muted);
  font-family: var(--font-data);
}
.state-pill {
  font-weight: 700;
  padding: 0.1rem 0.4rem;
  border-radius: 3px;
}
.state-pill.enabled { background: rgba(16, 185, 129, 0.2); color: #10b981; }
.state-pill.disabled { background: rgba(100, 116, 139, 0.2); color: #94a3b8; }

/* Analytics */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}
.kpi-card {
  background: var(--bg-card);
  border: 1px solid var(--border-card);
  border-radius: 6px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
}
.kpi-lbl {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: uppercase;
  font-weight: 600;
  margin-bottom: 0.25rem;
}
.kpi-val {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--text-primary);
  font-family: var(--font-data);
}
.kpi-sub {
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-top: 0.25rem;
}
.status-mini-bars {
  display: flex;
  gap: 0.4rem;
  margin-top: 0.5rem;
  flex-wrap: wrap;
}
.mini-bar {
  font-size: 0.7rem;
  padding: 0.1rem 0.4rem;
  border-radius: 3px;
  font-family: var(--font-data);
  font-weight: 600;
}
.mini-bar.green { background: rgba(16, 185, 129, 0.2); color: #10b981; }
.mini-bar.blue { background: rgba(59, 130, 246, 0.2); color: #60a5fa; }
.mini-bar.yellow { background: rgba(245, 158, 11, 0.2); color: #fbbf24; }
.mini-bar.red { background: rgba(239, 68, 68, 0.2); color: #f87171; }

.analytics-tables-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 1.25rem;
}
.analytics-card {
  background: var(--bg-card);
  border: 1px solid var(--border-card);
  border-radius: 6px;
  padding: 1.25rem;
}
.card-heading {
  font-size: 1rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 1rem;
}

/* Logs Explorer */
.logs-header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
  margin-bottom: 1rem;
}
.log-type-selector {
  display: flex;
  background: var(--bg-card);
  border: 1px solid var(--border-card);
  border-radius: 4px;
  padding: 0.2rem;
}
.type-btn {
  padding: 0.4rem 0.8rem;
  font-size: 0.8rem;
  font-weight: 600;
  background: transparent;
  border: none;
  color: var(--text-muted);
  border-radius: 3px;
  cursor: pointer;
}
.type-btn.active {
  background: var(--rdr-crimson);
  color: #fff;
}
.logs-filter-bar {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  align-items: center;
}
.form-select, .form-input {
  background: var(--bg-card);
  border: 1px solid var(--border-card);
  color: var(--text-primary);
  padding: 0.45rem 0.75rem;
  border-radius: 4px;
  font-size: 0.85rem;
}
.search-input-wrap {
  position: relative;
}
.search-icon {
  position: absolute;
  left: 0.6rem;
  top: 50%;
  transform: translateY(-50%);
  width: 14px;
  height: 14px;
  color: var(--text-muted);
}
.form-input.search {
  padding-left: 2rem;
  min-width: 200px;
}

/* Log Tables */
.log-table-wrap {
  background: var(--bg-card);
  border: 1px solid var(--border-card);
  border-radius: 6px;
  overflow-x: auto;
}
.log-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.82rem;
}
.log-table th {
  text-align: left;
  padding: 0.75rem 1rem;
  background: rgba(0, 0, 0, 0.3);
  color: var(--text-muted);
  font-size: 0.75rem;
  text-transform: uppercase;
  border-bottom: 1px solid var(--border-subtle);
}
.log-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-subtle);
}
.log-row {
  cursor: pointer;
  transition: background 0.15s ease;
}
.log-row:hover {
  background: var(--bg-card-hover);
}
.status-badge, .level-badge {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.15rem 0.45rem;
  border-radius: 3px;
  font-family: var(--font-data);
}
.status-badge-2xx { background: rgba(16, 185, 129, 0.15); color: #10b981; }
.status-badge-3xx { background: rgba(59, 130, 246, 0.15); color: #60a5fa; }
.status-badge-4xx { background: rgba(245, 158, 11, 0.15); color: #fbbf24; }
.status-badge-5xx { background: rgba(239, 68, 68, 0.15); color: #f87171; }
.level-error { background: rgba(239, 68, 68, 0.15); color: #f87171; }
.level-warn { background: rgba(245, 158, 11, 0.15); color: #fbbf24; }
.level-info { background: rgba(59, 130, 246, 0.15); color: #60a5fa; }
.inspect-btn {
  font-size: 0.72rem;
  background: var(--rdr-crimson-dim);
  color: var(--rdr-crimson-hover);
  border: 1px solid var(--border-crimson);
  border-radius: 3px;
  padding: 0.2rem 0.5rem;
  cursor: pointer;
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
  padding: 1rem;
}
.modal-card {
  background: #14161d;
  border: 1px solid var(--border-card);
  border-radius: 6px;
  width: 100%;
  max-width: 800px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}
.modal-header {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border-subtle);
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.modal-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-primary);
}
.modal-subtitle {
  font-size: 0.8rem;
  color: var(--text-muted);
}
.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
}
.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border-subtle);
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}
.code-editor {
  width: 100%;
  background: #090a0f;
  border: 1px solid var(--border-card);
  border-radius: 4px;
  color: #38bdf8;
  font-family: var(--font-data);
  font-size: 0.85rem;
  padding: 1rem;
  line-height: 1.5;
  resize: vertical;
}

/* Common UI */
.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  font-size: 0.85rem;
  font-weight: 600;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
}
.action-btn.primary {
  background: var(--rdr-crimson);
  color: #fff;
  border: 1px solid var(--rdr-crimson-hover);
}
.action-btn.primary:hover { background: var(--rdr-crimson-hover); }
.action-btn.secondary {
  background: var(--bg-card);
  color: var(--text-primary);
  border: 1px solid var(--border-card);
}
.action-btn.secondary:hover { background: var(--bg-card-hover); }
.icon-btn {
  background: transparent;
  border: none;
  padding: 0.35rem;
  border-radius: 4px;
  cursor: pointer;
}
.icon-btn:hover { background: var(--bg-card-hover); }
.text-green { color: #10b981; }
.text-red { color: #ef4444; }
.text-cyan { color: #38bdf8; }
.text-amber { color: #f59e0b; }
.text-muted { color: #64748b; }
.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}
.data-table th, .data-table td {
  padding: 0.6rem 0.75rem;
  border-bottom: 1px solid var(--border-subtle);
}
.data-table th { color: var(--text-muted); font-size: 0.75rem; text-transform: uppercase; }
.loading-state, .empty-state {
  padding: 3rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  color: var(--text-muted);
}
.inspect-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 0.75rem;
}
.inspect-item {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}
.inspect-key { color: var(--text-muted); font-weight: 600; text-transform: uppercase; font-size: 0.7rem; }
.inspect-val { color: var(--text-primary); word-break: break-all; }
.raw-pre {
  background: #090a0f;
  padding: 0.75rem;
  border-radius: 4px;
  color: #38bdf8;
  overflow-x: auto;
  white-space: pre-wrap;
}
</style>
