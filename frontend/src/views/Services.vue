<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Sliders,
  Play,
  Square,
  RotateCw,
  RefreshCw,
  Clock,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  FileText,
  Search,
  Plus,
  Trash2,
  Edit3,
  Terminal,
  Activity,
  Copy,
  Check,
  Pause,
  ArrowDownCircle,
  ShieldAlert,
  Server,
  Calendar,
  Layers,
  Sparkles
} from 'lucide-vue-next'
import api from '../api'
import { useToastStore } from '../stores/toast'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const toast = useToastStore()
const authStore = useAuthStore()

// State
const activeTab = ref('systemd') // 'systemd' | 'cron'
const loading = ref(false)
const systemdOverview = ref({
  available: true,
  total_units: 0,
  active_units: 0,
  failed_units: 0,
  units: []
})
const cronJobs = ref([])

// Filters & Search
const unitTypeFilter = ref('service')
const searchQuery = ref('')
const searchDebounce = ref(null)

// Action Modals & Drawers
const executing = ref(false)
const selectedUnit = ref(null)
const unitStatusModal = ref({
  show: false,
  unit: '',
  content: '',
  loading: false
})

// Journalctl Live Log Stream
const logStreamModal = ref({
  show: false,
  unit: '',
  logs: [],
  socket: null,
  autoScroll: true,
  paused: false
})
const logContainer = ref(null)

// Cron Task Modal (Create / Edit)
const cronModal = ref({
  show: false,
  isEdit: false,
  id: '',
  preset: '0 0 * * *',
  expression: '0 0 * * *',
  command: '',
  comment: '',
  enabled: true,
  saving: false
})

// Manual Cron Run Output Modal
const cronRunModal = ref({
  show: false,
  job: null,
  output: '',
  duration_ms: 0,
  success: false,
  running: false
})

// Cron Presets
const cronPresets = [
  { label: 'preset_every_5m', expr: '*/5 * * * *' },
  { label: 'preset_hourly', expr: '0 * * * *' },
  { label: 'preset_daily', expr: '0 0 * * *' },
  { label: 'preset_weekly', expr: '0 0 * * 0' },
  { label: 'preset_monthly', expr: '0 0 1 * *' },
  { label: 'preset_reboot', expr: '@reboot' },
  { label: 'preset_custom', expr: 'custom' }
]

// Fetch Systemd Units
const fetchSystemdUnits = async () => {
  loading.value = true
  try {
    const res = await api.get('/systemd/units', {
      params: {
        type: unitTypeFilter.value,
        search: searchQuery.value
      }
    })
    if (res.data?.data) {
      systemdOverview.value = res.data.data
    }
  } catch (err) {
    toast.error(err.response?.data?.message || t('services.action_error'))
  } finally {
    loading.value = false
  }
}

// Fetch Cron Jobs
const fetchCronJobs = async () => {
  loading.value = true
  try {
    const res = await api.get('/cron/jobs')
    if (res.data?.data) {
      cronJobs.value = res.data.data
    }
  } catch (err) {
    toast.error(err.response?.data?.message || t('services.action_error'))
  } finally {
    loading.value = false
  }
}

// Refresh active tab
const handleRefresh = () => {
  if (activeTab.value === 'systemd') {
    fetchSystemdUnits()
  } else {
    fetchCronJobs()
  }
}

// Search debounce
watch(searchQuery, () => {
  if (activeTab.value !== 'systemd') return
  clearTimeout(searchDebounce.value)
  searchDebounce.value = setTimeout(() => {
    fetchSystemdUnits()
  }, 350)
})

watch(unitTypeFilter, () => {
  if (activeTab.value === 'systemd') {
    fetchSystemdUnits()
  }
})

watch(activeTab, (tab) => {
  if (tab === 'systemd') {
    fetchSystemdUnits()
  } else {
    fetchCronJobs()
  }
})

// Execute Systemd Action
const executeSystemdAction = async (unitName, action) => {
  executing.value = true
  try {
    const res = await api.post(`/systemd/unit/${encodeURIComponent(unitName)}/action`, { action })
    toast.success(res.data?.message || t('services.action_success'))
    fetchSystemdUnits()
  } catch (err) {
    toast.error(err.response?.data?.message || err.response?.data?.error || t('services.action_error'))
  } finally {
    executing.value = false
  }
}

// Inspect Unit Status
const inspectUnitStatus = async (unitName) => {
  unitStatusModal.value = {
    show: true,
    unit: unitName,
    content: '',
    loading: true
  }
  try {
    const res = await api.get(`/systemd/unit/${encodeURIComponent(unitName)}/status`)
    unitStatusModal.value.content = res.data?.data?.status || 'No output available.'
  } catch (err) {
    unitStatusModal.value.content = err.response?.data?.error || err.message
  } finally {
    unitStatusModal.value.loading = false
  }
}

// Open Journalctl Live Stream
const openJournalStream = (unitName) => {
  closeJournalStream()
  logStreamModal.value = {
    show: true,
    unit: unitName,
    logs: [],
    socket: null,
    autoScroll: true,
    paused: false
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsHost = window.location.host
  const token = authStore.token
  const wsUrl = `${protocol}//${wsHost}/ws/systemd/${encodeURIComponent(unitName)}/logs?token=${encodeURIComponent(token)}`

  const ws = new WebSocket(wsUrl)
  logStreamModal.value.socket = ws

  ws.onopen = () => {
    logStreamModal.value.logs.push(`[SYSTEM] Attached to journalctl live stream for: ${unitName}`)
  }

  ws.onmessage = (event) => {
    if (logStreamModal.value.paused) return
    logStreamModal.value.logs.push(event.data)
    if (logStreamModal.value.logs.length > 1000) {
      logStreamModal.value.logs.shift()
    }
    if (logStreamModal.value.autoScroll) {
      nextTick(() => {
        if (logContainer.value) {
          logContainer.value.scrollTop = logContainer.value.scrollHeight
        }
      })
    }
  }

  ws.onerror = () => {
    logStreamModal.value.logs.push('[ERROR] WebSocket connection encountered an error.')
  }

  ws.onclose = () => {
    logStreamModal.value.logs.push('[SYSTEM] Journalctl log stream closed.')
  }
}

const closeJournalStream = () => {
  if (logStreamModal.value.socket) {
    logStreamModal.value.socket.close()
    logStreamModal.value.socket = null
  }
  logStreamModal.value.show = false
}

const toggleLogPause = () => {
  logStreamModal.value.paused = !logStreamModal.value.paused
}

const clearLogs = () => {
  logStreamModal.value.logs = []
}

// Cron Management
const openCreateCronModal = () => {
  cronModal.value = {
    show: true,
    isEdit: false,
    id: '',
    preset: '0 0 * * *',
    expression: '0 0 * * *',
    command: '',
    comment: '',
    enabled: true,
    saving: false
  }
}

const openEditCronModal = (job) => {
  cronModal.value = {
    show: true,
    isEdit: true,
    id: job.id,
    preset: 'custom',
    expression: job.expression,
    command: job.command,
    comment: job.comment,
    enabled: job.enabled,
    saving: false
  }
}

const handlePresetChange = () => {
  if (cronModal.value.preset !== 'custom') {
    cronModal.value.expression = cronModal.value.preset
  }
}

const saveCronJob = async () => {
  if (!cronModal.value.expression || !cronModal.value.command) {
    toast.error('Expression and command are required')
    return
  }
  cronModal.value.saving = true
  try {
    await api.post('/cron/jobs', {
      id: cronModal.value.isEdit ? cronModal.value.id : undefined,
      expression: cronModal.value.expression,
      command: cronModal.value.command,
      comment: cronModal.value.comment,
      enabled: cronModal.value.enabled
    })
    toast.success(t('services.action_success'))
    cronModal.value.show = false
    fetchCronJobs()
  } catch (err) {
    toast.error(err.response?.data?.message || t('services.action_error'))
  } finally {
    cronModal.value.saving = false
  }
}

const toggleCronJob = async (job) => {
  try {
    const targetState = !job.enabled
    await api.post(`/cron/jobs/${encodeURIComponent(job.id)}/toggle`, {
      enabled: targetState
    })
    job.enabled = targetState
    toast.success(t('services.action_success'))
  } catch (err) {
    toast.error(err.response?.data?.message || t('services.action_error'))
  }
}

const deleteCronJob = async (job) => {
  if (!confirm(t('services.delete_task_confirm'))) return
  try {
    await api.delete(`/cron/jobs/${encodeURIComponent(job.id)}`)
    toast.success(t('services.action_success'))
    fetchCronJobs()
  } catch (err) {
    toast.error(err.response?.data?.message || t('services.action_error'))
  }
}

const runCronJobNow = async (job) => {
  cronRunModal.value = {
    show: true,
    job: job,
    output: '',
    duration_ms: 0,
    success: false,
    running: true
  }
  try {
    const res = await api.post(`/cron/jobs/${encodeURIComponent(job.id)}/run`)
    cronRunModal.value.output = res.data?.data?.output || 'Command executed.'
    cronRunModal.value.duration_ms = res.data?.data?.duration_ms || 0
    cronRunModal.value.success = res.data?.data?.success || false
  } catch (err) {
    cronRunModal.value.output = err.response?.data?.error || err.message
    cronRunModal.value.success = false
  } finally {
    cronRunModal.value.running = false
  }
}

// Copy to clipboard helper
const copiedMap = ref({})
const copyToClipboard = (text, id) => {
  navigator.clipboard.writeText(text)
  copiedMap.value[id] = true
  setTimeout(() => {
    copiedMap.value[id] = false
  }, 2000)
}

onMounted(() => {
  fetchSystemdUnits()
})

onUnmounted(() => {
  closeJournalStream()
})
</script>

<template>
  <div class="services-view-container">
    <!-- Header Hero Banner -->
    <header class="engine-header">
      <div class="header-left">
        <div class="badge-tag">
          <span class="pulse-dot"></span>
          <span>{{ $t('services.subtitle') }}</span>
        </div>
        <h1 class="page-title">{{ $t('services.title') }}</h1>
      </div>

      <!-- Tab Switcher -->
      <div class="tab-controls">
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'systemd' }"
          @click="activeTab = 'systemd'"
        >
          <Sliders :size="16" />
          <span>{{ $t('services.tab_systemd') }}</span>
        </button>
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'cron' }"
          @click="activeTab = 'cron'"
        >
          <Calendar :size="16" />
          <span>{{ $t('services.tab_cron') }}</span>
        </button>
      </div>
    </header>

    <!-- Stat Metrics Row -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon-wrap neutral">
          <Layers :size="20" />
        </div>
        <div class="stat-meta">
          <span class="stat-label">{{ $t('services.total_units') }}</span>
          <span class="stat-value">{{ systemdOverview.total_units || 0 }}</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon-wrap optimal">
          <CheckCircle2 :size="20" />
        </div>
        <div class="stat-meta">
          <span class="stat-label">{{ $t('services.active_units') }}</span>
          <span class="stat-value text-optimal">{{ systemdOverview.active_units || 0 }}</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon-wrap danger">
          <XCircle :size="20" />
        </div>
        <div class="stat-meta">
          <span class="stat-label">{{ $t('services.failed_units') }}</span>
          <span class="stat-value text-danger">{{ systemdOverview.failed_units || 0 }}</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon-wrap warning">
          <Clock :size="20" />
        </div>
        <div class="stat-meta">
          <span class="stat-label">{{ $t('services.total_tasks') }}</span>
          <span class="stat-value text-warning">{{ cronJobs.length }}</span>
        </div>
      </div>
    </div>

    <!-- TAB 1: SYSTEMD UNITS -->
    <section v-if="activeTab === 'systemd'" class="content-panel">
      <!-- Toolbar -->
      <div class="panel-toolbar">
        <div class="search-wrap">
          <Search :size="16" class="search-icon" />
          <input
            v-model="searchQuery"
            type="text"
            class="tactical-input"
            :placeholder="$t('services.search_placeholder')"
          />
        </div>

        <div class="filter-actions">
          <select v-model="unitTypeFilter" class="tactical-select">
            <option value="service">{{ $t('services.services_type') }}</option>
            <option value="timer">{{ $t('services.timers_type') }}</option>
            <option value="socket">{{ $t('services.sockets_type') }}</option>
            <option value="all">{{ $t('services.all_types') }}</option>
          </select>

          <button class="action-btn-secondary" :disabled="loading" @click="handleRefresh">
            <RefreshCw :size="16" :class="{ 'spin-anim': loading }" />
            <span>{{ $t('services.refresh') }}</span>
          </button>
        </div>
      </div>

      <!-- Non-systemd banner -->
      <div v-if="!systemdOverview.available" class="notice-banner warning">
        <ShieldAlert :size="20" />
        <div>{{ $t('services.systemd_unavailable') }}</div>
      </div>

      <!-- Units Table -->
      <div class="table-container">
        <table class="tactical-table">
          <thead>
            <tr>
              <th>{{ $t('services.unit_name') }}</th>
              <th>{{ $t('services.type') }}</th>
              <th>{{ $t('services.load_state') }}</th>
              <th>{{ $t('services.active_state') }}</th>
              <th>{{ $t('services.sub_state') }}</th>
              <th>{{ $t('services.description') }}</th>
              <th class="text-right">{{ $t('services.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="7" class="text-center py-8">
                <div class="loading-state">
                  <RefreshCw :size="24" class="spin-anim text-crimson" />
                  <span>{{ $t('common.loading') }}</span>
                </div>
              </td>
            </tr>

            <tr v-else-if="!systemdOverview.units || systemdOverview.units.length === 0">
              <td colspan="7" class="text-center py-8 text-muted">
                {{ $t('services.no_units_found') }}
              </td>
            </tr>

            <tr v-for="u in systemdOverview.units" :key="u.unit" class="table-row">
              <td class="font-mono font-bold text-amber">
                {{ u.unit }}
              </td>
              <td>
                <span class="badge-type">{{ u.type }}</span>
              </td>
              <td>
                <span class="badge-load" :class="u.load">{{ u.load }}</span>
              </td>
              <td>
                <span
                  class="badge-state"
                  :class="{
                    'active-badge': u.active === 'active',
                    'failed-badge': u.active === 'failed',
                    'inactive-badge': u.active === 'inactive'
                  }"
                >
                  <span class="status-indicator"></span>
                  {{ u.active }}
                </span>
              </td>
              <td class="text-muted font-mono text-xs">
                {{ u.sub }}
              </td>
              <td class="text-secondary text-sm max-w-xs truncate" :title="u.description">
                {{ u.description || '—' }}
              </td>
              <td class="text-right">
                <div class="row-actions">
                  <!-- Start / Restart / Stop -->
                  <button
                    v-if="u.active !== 'active'"
                    class="btn-icon play"
                    :title="$t('services.start')"
                    :disabled="executing"
                    @click="executeSystemdAction(u.unit, 'start')"
                  >
                    <Play :size="14" />
                  </button>
                  <button
                    v-else
                    class="btn-icon restart"
                    :title="$t('services.restart')"
                    :disabled="executing"
                    @click="executeSystemdAction(u.unit, 'restart')"
                  >
                    <RotateCw :size="14" />
                  </button>

                  <button
                    v-if="u.active === 'active'"
                    class="btn-icon stop"
                    :title="$t('services.stop')"
                    :disabled="executing"
                    @click="executeSystemdAction(u.unit, 'stop')"
                  >
                    <Square :size="14" />
                  </button>

                  <!-- Status -->
                  <button
                    class="btn-icon inspect"
                    :title="$t('services.status')"
                    @click="inspectUnitStatus(u.unit)"
                  >
                    <FileText :size="14" />
                  </button>

                  <!-- Live Logs -->
                  <button
                    class="btn-icon terminal"
                    :title="$t('services.logs')"
                    @click="openJournalStream(u.unit)"
                  >
                    <Terminal :size="14" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- TAB 2: SCHEDULED CRONTAB TASKS -->
    <section v-if="activeTab === 'cron'" class="content-panel">
      <!-- Toolbar -->
      <div class="panel-toolbar">
        <div class="section-tagline">
          <Clock :size="18" class="text-amber" />
          <span>Automated cron daemon triggers</span>
        </div>

        <div class="filter-actions">
          <button class="action-btn-primary" @click="openCreateCronModal">
            <Plus :size="16" />
            <span>{{ $t('services.new_task') }}</span>
          </button>
          <button class="action-btn-secondary" :disabled="loading" @click="handleRefresh">
            <RefreshCw :size="16" :class="{ 'spin-anim': loading }" />
            <span>{{ $t('services.refresh') }}</span>
          </button>
        </div>
      </div>

      <!-- Tasks Table -->
      <div class="table-container">
        <table class="tactical-table">
          <thead>
            <tr>
              <th>{{ $t('services.cron_state') }}</th>
              <th>{{ $t('services.cron_expression') }}</th>
              <th>{{ $t('services.cron_schedule') }}</th>
              <th>{{ $t('services.cron_command') }}</th>
              <th>{{ $t('services.cron_comment') }}</th>
              <th class="text-right">{{ $t('services.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="6" class="text-center py-8">
                <div class="loading-state">
                  <RefreshCw :size="24" class="spin-anim text-crimson" />
                  <span>{{ $t('common.loading') }}</span>
                </div>
              </td>
            </tr>

            <tr v-else-if="cronJobs.length === 0">
              <td colspan="6" class="text-center py-8 text-muted">
                {{ $t('services.no_tasks_found') }}
              </td>
            </tr>

            <tr v-for="job in cronJobs" :key="job.id" class="table-row">
              <td>
                <label class="switch-toggle" :title="job.enabled ? $t('services.enabled') : $t('services.disabled')">
                  <input
                    type="checkbox"
                    :checked="job.enabled"
                    @change="toggleCronJob(job)"
                  />
                  <span class="slider"></span>
                </label>
              </td>
              <td>
                <span class="cron-badge-expr font-mono">{{ job.expression }}</span>
              </td>
              <td>
                <div class="schedule-human">
                  <Sparkles :size="14" class="text-amber" />
                  <span>{{ job.schedule_human }}</span>
                </div>
              </td>
              <td class="command-cell font-mono">
                <div class="cmd-wrap">
                  <span class="cmd-text" :title="job.command">{{ job.command }}</span>
                  <button
                    class="btn-copy"
                    :title="$t('packages.copied')"
                    @click="copyToClipboard(job.command, job.id)"
                  >
                    <Check v-if="copiedMap[job.id]" :size="12" class="text-optimal" />
                    <Copy v-else :size="12" />
                  </button>
                </div>
              </td>
              <td class="text-secondary text-sm">
                {{ job.comment || '—' }}
              </td>
              <td class="text-right">
                <div class="row-actions">
                  <!-- Run Now -->
                  <button
                    class="btn-icon play"
                    :title="$t('services.run_now')"
                    @click="runCronJobNow(job)"
                  >
                    <Play :size="14" />
                  </button>

                  <!-- Edit -->
                  <button
                    class="btn-icon inspect"
                    :title="$t('services.edit')"
                    @click="openEditCronModal(job)"
                  >
                    <Edit3 :size="14" />
                  </button>

                  <!-- Delete -->
                  <button
                    class="btn-icon stop"
                    :title="$t('services.delete')"
                    @click="deleteCronJob(job)"
                  >
                    <Trash2 :size="14" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- MODAL: LIVE JOURNALCTL STREAM DRAWER -->
    <div v-if="logStreamModal.show" class="tactical-modal-backdrop" @click.self="closeJournalStream">
      <div class="tactical-modal-card log-modal">
        <div class="modal-header">
          <div class="header-title-box">
            <Terminal :size="20" class="text-crimson" />
            <div>
              <h3>{{ $t('services.journal_logs_title') }}</h3>
              <p class="modal-sub font-mono text-amber">{{ logStreamModal.unit }}</p>
            </div>
          </div>
          <div class="modal-controls">
            <button
              class="control-pill"
              :class="{ active: !logStreamModal.paused }"
              @click="toggleLogPause"
            >
              <Pause v-if="!logStreamModal.paused" :size="14" />
              <Play v-else :size="14" />
              <span>{{ logStreamModal.paused ? $t('services.stream_paused') : $t('services.stream_active') }}</span>
            </button>
            <button
              class="control-pill"
              :class="{ active: logStreamModal.autoScroll }"
              @click="logStreamModal.autoScroll = !logStreamModal.autoScroll"
            >
              <ArrowDownCircle :size="14" />
              <span>{{ $t('services.auto_scroll') }}</span>
            </button>
            <button class="control-pill danger" @click="clearLogs">
              <span>{{ $t('services.clear_logs') }}</span>
            </button>
            <button class="btn-close" @click="closeJournalStream">✕</button>
          </div>
        </div>

        <div ref="logContainer" class="log-stream-terminal font-mono">
          <div v-for="(line, idx) in logStreamModal.logs" :key="idx" class="terminal-line">
            {{ line }}
          </div>
        </div>
      </div>
    </div>

    <!-- MODAL: UNIT STATUS INSPECTION -->
    <div v-if="unitStatusModal.show" class="tactical-modal-backdrop" @click.self="unitStatusModal.show = false">
      <div class="tactical-modal-card status-modal">
        <div class="modal-header">
          <div class="header-title-box">
            <FileText :size="20" class="text-amber" />
            <div>
              <h3>{{ $t('services.unit_status_title') }}</h3>
              <p class="modal-sub font-mono text-amber">{{ unitStatusModal.unit }}</p>
            </div>
          </div>
          <button class="btn-close" @click="unitStatusModal.show = false">✕</button>
        </div>

        <div class="status-content-terminal font-mono">
          <div v-if="unitStatusModal.loading" class="text-center py-6">
            <RefreshCw :size="20" class="spin-anim text-amber" />
          </div>
          <pre v-else>{{ unitStatusModal.content }}</pre>
        </div>
      </div>
    </div>

    <!-- MODAL: CREATE / EDIT CRON JOB -->
    <div v-if="cronModal.show" class="tactical-modal-backdrop" @click.self="cronModal.show = false">
      <div class="tactical-modal-card form-modal">
        <div class="modal-header">
          <div class="header-title-box">
            <Clock :size="20" class="text-crimson" />
            <h3>{{ cronModal.isEdit ? $t('services.modal_edit_task') : $t('services.modal_create_task') }}</h3>
          </div>
          <button class="btn-close" @click="cronModal.show = false">✕</button>
        </div>

        <div class="modal-body">
          <!-- Preset Selector -->
          <div class="form-group">
            <label class="form-label">{{ $t('services.preset_select') }}</label>
            <select v-model="cronModal.preset" class="tactical-select full-width" @change="handlePresetChange">
              <option v-for="p in cronPresets" :key="p.expr" :value="p.expr">
                {{ $t(`services.${p.label}`) }} ({{ p.expr }})
              </option>
            </select>
          </div>

          <!-- Cron Expression Input -->
          <div class="form-group">
            <label class="form-label">{{ $t('services.cron_expression') }} (e.g. */5 * * * * or 0 3 * * *)</label>
            <input
              v-model="cronModal.expression"
              type="text"
              class="tactical-input full-width font-mono"
              placeholder="0 0 * * *"
            />
          </div>

          <!-- Command -->
          <div class="form-group">
            <label class="form-label">{{ $t('services.cron_command') }}</label>
            <input
              v-model="cronModal.command"
              type="text"
              class="tactical-input full-width font-mono"
              placeholder="/usr/bin/python3 /opt/backup.py"
            />
          </div>

          <!-- Comment / Label -->
          <div class="form-group">
            <label class="form-label">{{ $t('services.cron_comment') }}</label>
            <input
              v-model="cronModal.comment"
              type="text"
              class="tactical-input full-width"
              placeholder="Nightly database backup"
            />
          </div>

          <!-- Enabled Toggle -->
          <div class="form-group-switch">
            <label class="form-label">{{ $t('services.cron_state') }}</label>
            <label class="switch-toggle">
              <input type="checkbox" v-model="cronModal.enabled" />
              <span class="slider"></span>
            </label>
          </div>
        </div>

        <div class="modal-footer">
          <button class="action-btn-secondary" @click="cronModal.show = false">
            {{ $t('packages.cancel') }}
          </button>
          <button class="action-btn-primary" :disabled="cronModal.saving" @click="saveCronJob">
            <RefreshCw v-if="cronModal.saving" :size="16" class="spin-anim" />
            <span>{{ $t('services.save_task') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL: MANUAL CRON RUN OUTPUT -->
    <div v-if="cronRunModal.show" class="tactical-modal-backdrop" @click.self="cronRunModal.show = false">
      <div class="tactical-modal-card status-modal">
        <div class="modal-header">
          <div class="header-title-box">
            <Terminal :size="20" class="text-amber" />
            <div>
              <h3>{{ $t('services.run_task_title') }}</h3>
              <p class="modal-sub font-mono text-amber">{{ cronRunModal.job?.command }}</p>
            </div>
          </div>
          <button class="btn-close" @click="cronRunModal.show = false">✕</button>
        </div>

        <div class="status-content-terminal font-mono">
          <div v-if="cronRunModal.running" class="text-center py-6">
            <RefreshCw :size="24" class="spin-anim text-crimson" />
            <p class="mt-2 text-muted">{{ $t('services.executing') }}</p>
          </div>
          <div v-else>
            <div class="meta-banner mb-3" :class="cronRunModal.success ? 'success' : 'danger'">
              <span>Duration: {{ cronRunModal.duration_ms }} ms</span>
              <span>Status: {{ cronRunModal.success ? 'SUCCESS' : 'FAILED' }}</span>
            </div>
            <pre>{{ cronRunModal.output }}</pre>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.services-view-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding-bottom: 3rem;
}

/* Header & Tab Controls */
.engine-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
  border-bottom: 1px solid rgba(220, 38, 38, 0.2);
  padding-bottom: 1.25rem;
}

.badge-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #f59e0b;
  margin-bottom: 0.25rem;
}

.pulse-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #dc2626;
  box-shadow: 0 0 10px #dc2626;
  animation: pulse-glow 2s infinite;
}

@keyframes pulse-glow {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(0.85); }
}

.page-title {
  font-family: 'Cinzel', serif;
  font-size: 1.85rem;
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: 0.05em;
  margin: 0;
}

.tab-controls {
  display: flex;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 4px;
  gap: 4px;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: #94a3b8;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab-btn:hover {
  color: #f8fafc;
  background: rgba(255, 255, 255, 0.05);
}

.tab-btn.active {
  background: linear-gradient(135deg, rgba(220, 38, 38, 0.3), rgba(185, 28, 28, 0.5));
  color: #ffffff;
  border: 1px solid rgba(220, 38, 38, 0.5);
  box-shadow: 0 0 15px rgba(220, 38, 38, 0.25);
}

/* Stats Row */
.stats-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  background: rgba(18, 24, 38, 0.7);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  padding: 1.1rem;
  backdrop-filter: blur(10px);
}

.stat-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-icon-wrap.neutral {
  background: rgba(148, 163, 184, 0.1);
  color: #94a3b8;
}

.stat-icon-wrap.optimal {
  background: rgba(34, 197, 94, 0.1);
  color: #22c55e;
}

.stat-icon-wrap.danger {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.stat-icon-wrap.warning {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.stat-meta {
  display: flex;
  flex-direction: column;
}

.stat-label {
  font-size: 0.7rem;
  font-weight: 600;
  color: #64748b;
  letter-spacing: 0.05em;
}

.stat-value {
  font-size: 1.35rem;
  font-weight: 700;
  color: #f8fafc;
}

.text-optimal { color: #22c55e; }
.text-danger { color: #ef4444; }
.text-warning { color: #f59e0b; }
.text-amber { color: #fbbf24; }
.text-crimson { color: #dc2626; }
.text-muted { color: #64748b; }

/* Panels & Tables */
.content-panel {
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  padding: 1.25rem;
  backdrop-filter: blur(12px);
}

.panel-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
  margin-bottom: 1.25rem;
}

.search-wrap {
  position: relative;
  min-width: 280px;
  flex: 1;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #64748b;
}

.tactical-input {
  width: 100%;
  background: rgba(10, 15, 28, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 6px;
  padding: 0.55rem 0.75rem 0.55rem 2.25rem;
  color: #f8fafc;
  font-size: 0.85rem;
  outline: none;
  transition: border-color 0.2s;
}

.tactical-input:focus {
  border-color: #dc2626;
  box-shadow: 0 0 10px rgba(220, 38, 38, 0.3);
}

.filter-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.tactical-select {
  background: rgba(10, 15, 28, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 6px;
  padding: 0.55rem 0.75rem;
  color: #f8fafc;
  font-size: 0.85rem;
  outline: none;
  cursor: pointer;
}

.action-btn-primary, .action-btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 0.95rem;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn-primary {
  background: linear-gradient(135deg, #dc2626, #991b1b);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #ffffff;
}

.action-btn-primary:hover {
  background: linear-gradient(135deg, #ef4444, #b91c1c);
  box-shadow: 0 0 12px rgba(220, 38, 38, 0.4);
}

.action-btn-secondary {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #cbd5e1;
}

.action-btn-secondary:hover {
  background: rgba(51, 65, 85, 0.8);
  color: #ffffff;
}

.section-tagline {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #94a3b8;
  font-size: 0.85rem;
}

.table-container {
  overflow-x: auto;
}

.tactical-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.tactical-table th {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: #94a3b8;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.tactical-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  font-size: 0.85rem;
}

.table-row:hover td {
  background: rgba(255, 255, 255, 0.02);
}

.badge-type {
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.3);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.7rem;
  font-family: monospace;
}

.badge-state {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
}

.status-indicator {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.active-badge {
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}
.active-badge .status-indicator { background: #22c55e; }

.failed-badge {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}
.failed-badge .status-indicator { background: #ef4444; }

.inactive-badge {
  background: rgba(148, 163, 184, 0.1);
  color: #94a3b8;
  border: 1px solid rgba(148, 163, 184, 0.2);
}
.inactive-badge .status-indicator { background: #94a3b8; }

.cron-badge-expr {
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.8rem;
}

.schedule-human {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  color: #e2e8f0;
  font-size: 0.8rem;
}

.cmd-wrap {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: rgba(0, 0, 0, 0.3);
  padding: 4px 8px;
  border-radius: 4px;
  max-width: 320px;
}

.cmd-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.8rem;
  color: #cbd5e1;
}

.btn-copy {
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
}
.btn-copy:hover { color: #f8fafc; }

.row-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.4rem;
}

.btn-icon {
  width: 28px;
  height: 28px;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.05);
  color: #94a3b8;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-icon.play:hover { background: rgba(34, 197, 94, 0.2); color: #22c55e; border-color: #22c55e; }
.btn-icon.restart:hover { background: rgba(245, 158, 11, 0.2); color: #fbbf24; border-color: #fbbf24; }
.btn-icon.stop:hover { background: rgba(239, 68, 68, 0.2); color: #ef4444; border-color: #ef4444; }
.btn-icon.inspect:hover { background: rgba(59, 130, 246, 0.2); color: #60a5fa; border-color: #60a5fa; }
.btn-icon.terminal:hover { background: rgba(220, 38, 38, 0.2); color: #dc2626; border-color: #dc2626; }

/* Switch Toggle */
.switch-toggle {
  position: relative;
  display: inline-block;
  width: 34px;
  height: 18px;
}
.switch-toggle input { opacity: 0; width: 0; height: 0; }
.slider {
  position: absolute;
  cursor: pointer;
  top: 0; left: 0; right: 0; bottom: 0;
  background-color: #334155;
  transition: 0.3s;
  border-radius: 18px;
}
.slider:before {
  position: absolute;
  content: "";
  height: 12px; width: 12px;
  left: 3px; bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}
input:checked + .slider { background-color: #22c55e; }
input:checked + .slider:before { transform: translateX(16px); }

/* Modals */
.tactical-modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.tactical-modal-card {
  background: #0f172a;
  border: 1px solid rgba(220, 38, 38, 0.3);
  box-shadow: 0 0 30px rgba(0, 0, 0, 0.8);
  border-radius: 12px;
  width: 100%;
  display: flex;
  flex-direction: column;
}

.log-modal {
  max-width: 900px;
  height: 80vh;
}

.status-modal {
  max-width: 800px;
  max-height: 80vh;
}

.form-modal {
  max-width: 520px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.header-title-box {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-title-box h3 {
  margin: 0;
  font-size: 1.1rem;
  color: #f8fafc;
}

.modal-sub {
  margin: 0;
  font-size: 0.8rem;
}

.modal-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.control-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #94a3b8;
  font-size: 0.75rem;
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
}

.control-pill.active {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
  border-color: rgba(34, 197, 94, 0.4);
}

.control-pill.danger {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.btn-close {
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 4px;
}
.btn-close:hover { color: #f8fafc; }

.log-stream-terminal, .status-content-terminal {
  flex: 1;
  background: #020617;
  padding: 1rem;
  overflow-y: auto;
  font-size: 0.8rem;
  line-height: 1.4;
  color: #cbd5e1;
}

.terminal-line {
  white-space: pre-wrap;
  word-break: break-all;
}

.modal-body {
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #94a3b8;
  text-transform: uppercase;
}

.form-group-switch {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 0.5rem;
}

.full-width {
  width: 100%;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.25rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.spin-anim {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.meta-banner {
  display: flex;
  justify-content: space-between;
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
}
.meta-banner.success { background: rgba(34, 197, 94, 0.2); color: #22c55e; }
.meta-banner.danger { background: rgba(239, 68, 68, 0.2); color: #ef4444; }

.notice-banner {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.85rem 1rem;
  border-radius: 8px;
  margin-bottom: 1rem;
  font-size: 0.85rem;
}
.notice-banner.warning {
  background: rgba(245, 158, 11, 0.15);
  border: 1px solid rgba(245, 158, 11, 0.3);
  color: #fbbf24;
}
</style>
