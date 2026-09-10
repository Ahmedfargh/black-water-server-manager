<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSystemStore } from '../stores/system'
import { 
  BarChart3, 
  Cpu, 
  Zap, 
  HardDrive, 
  Calendar,
  RefreshCw,
  Clock,
  ArrowBigRightDash
} from 'lucide-vue-next'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const { t } = useI18n()
const systemStore = useSystemStore()
const isLoading = ref(false)
const timeRange = ref('1h') // 1h, 6h, 24h, 7d, custom
const customStart = ref('')
const customEnd = ref('')

const isCustomRangeValid = computed(() => {
  return customStart.value && customEnd.value
})

const ranges = [
  { label: '1H', value: '1h', duration: 3600 * 1000 },
  { label: '6H', value: '6h', duration: 6 * 3600 * 1000 },
  { label: '24H', value: '24h', duration: 24 * 3600 * 1000 },
  { label: '7D', value: '7d', duration: 7 * 24 * 3600 * 1000 },
  { label: t('reports.custom') || 'CUSTOM', value: 'custom', duration: 0 },
]

const fetchData = async () => {
  isLoading.value = true
  try {
    let start, end

    if (timeRange.value === 'custom') {
      if (!customStart.value || !customEnd.value) {
        return
      }
      start = new Date(customStart.value).toISOString()
      end = new Date(customEnd.value).toISOString()
    } else {
      const range = ranges.find(r => r.value === timeRange.value)
      end = new Date().toISOString()
      start = new Date(Date.now() - range.duration).toISOString()
    }

    await Promise.all([
      systemStore.fetchHistoryReports(start, end),
      systemStore.fetchAverageReports(start, end)
    ])
  } catch (error) {
    console.error('Failed to sync with grid:', error)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchData()
})

watch(timeRange, (newRange) => {
  if (newRange === 'custom') {
    // Default to last 24h if empty
    if (!customStart.value || !customEnd.value) {
      const now = new Date()
      const yesterday = new Date(now.getTime() - 24 * 3600 * 1000)
      
      // Format as YYYY-MM-DDTHH:mm for datetime-local input
      const formatLocal = (d) => {
        const pad = (n) => n.toString().padStart(2, '0')
        return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
      }
      
      customStart.value = formatLocal(yesterday)
      customEnd.value = formatLocal(now)
    }
  }
  fetchData()
})

const chartData = computed(() => {
  const reports = systemStore.reports || []
  const labels = reports.map(r => {
    const date = new Date(r.CreatedAt)
    return timeRange.value === '7d' 
      ? date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      : date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  })

  return {
    labels,
    datasets: [
      {
        label: `${t('reports.cpu_usage_p') || 'CPU Usage (%)'}`,
        data: reports.map(r => r.cpu_usage),
        borderColor: '#dc2626',
        backgroundColor: 'rgba(220, 38, 38, 0.12)',
        fill: true,
        tension: 0.4,
        pointRadius: 0,
        pointHoverRadius: 5,
        borderWidth: 2
      },
      {
        label: `${t('reports.ram_usage_p') || 'RAM Usage (%)'}`,
        data: reports.map(r => r.memory_usage),
        borderColor: '#d97706',
        backgroundColor: 'rgba(217, 119, 6, 0.12)',
        fill: true,
        tension: 0.4,
        pointRadius: 0,
        pointHoverRadius: 5,
        borderWidth: 2
      },
      {
        label: `${t('reports.disk_usage_p') || 'Disk Usage (%)'}`,
        data: reports.map(r => r.disk_usage),
        borderColor: '#38bdf8',
        backgroundColor: 'rgba(56, 189, 248, 0.12)',
        fill: true,
        tension: 0.4,
        pointRadius: 0,
        pointHoverRadius: 5,
        borderWidth: 2
      }
    ]
  }
})

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'top',
      labels: {
        color: '#94a3b8',
        usePointStyle: true,
        padding: 20,
        font: {
          family: 'JetBrains Mono, monospace',
          size: 11
        }
      }
    },
    tooltip: {
      mode: 'index',
      intersect: false,
      backgroundColor: 'rgba(12, 13, 17, 0.95)',
      titleColor: '#f1f2f6',
      bodyColor: '#cbd5e1',
      borderColor: 'rgba(255, 255, 255, 0.1)',
      borderWidth: 1,
      padding: 12,
      displayColors: true,
      callbacks: {
        label: function(context) {
          return `${context.dataset.label}: ${context.parsed.y.toFixed(2)}%`
        }
      }
    }
  },
  scales: {
    y: {
      beginAtZero: true,
      max: 100,
      grid: {
        color: 'rgba(255, 255, 255, 0.05)',
        borderDash: [4, 4]
      },
      ticks: {
        color: '#64748b',
        font: { family: 'JetBrains Mono, monospace', size: 10 },
        callback: (value) => value + '%'
      }
    },
    x: {
      grid: {
        display: false
      },
      ticks: {
        color: '#64748b',
        font: { family: 'JetBrains Mono, monospace', size: 10 },
        maxRotation: 45,
        minRotation: 0,
        autoSkip: true,
        maxTicksLimit: 10
      }
    }
  },
  interaction: {
    intersect: false,
    mode: 'index',
  }
}
</script>

<template>
  <div class="reports-page">
    <header class="section-header">
      <div class="header-left">
        <div class="icon-orb">
          <BarChart3 :size="22" />
        </div>
        <div>
          <h1>{{ $t('reports.analysis_grid') }}</h1>
          <p class="subtitle">{{ $t('reports.archive') }}</p>
        </div>
      </div>

      <div class="range-selector">
        <div class="selector-label">
          <Clock :size="14" />
          <span>{{ $t('reports.temporal_scope') || 'TEMPORAL SCOPE' }}:</span>
        </div>
        <div class="range-buttons">
          <button 
            v-for="r in ranges" 
            :key="r.value"
            @click="timeRange = r.value"
            :class="{ active: timeRange === r.value }"
            class="range-btn"
          >
            {{ r.label }}
          </button>
        </div>

        <div v-if="timeRange === 'custom'" class="custom-inputs">
          <div class="input-group">
            <span class="input-label">{{ $t('reports.from') }}</span>
            <input type="datetime-local" v-model="customStart" class="tron-input" />
          </div>
          <div class="input-group">
            <span class="input-label">{{ $t('reports.to') }}</span>
            <input type="datetime-local" v-model="customEnd" class="tron-input" />
          </div>
          <button 
            class="apply-btn" 
            @click="fetchData" 
            :disabled="!customStart || !customEnd || isLoading"
          >
            {{ $t('reports.go') || 'GO' }}
          </button>
        </div>

        <button class="refresh-btn" @click="fetchData" :disabled="isLoading">
          <RefreshCw :size="16" :class="{ spinning: isLoading }" />
        </button>
      </div>
    </header>

    <div class="averages-grid">
      <div class="avg-card cpu">
        <div class="avg-header">
          <Cpu :size="18" class="cpu-icon" />
          <span>{{ $t('reports.avg_cpu') }}</span>
        </div>
        <div class="avg-value cpu-val">
          {{ systemStore.averages.cpu.toFixed(2) }}%
        </div>
        <div class="avg-footer">{{ $t('reports.throughput') }}</div>
      </div>

      <div class="avg-card memory">
        <div class="avg-header">
          <Zap :size="18" class="ram-icon" />
          <span>{{ $t('reports.avg_mem') }}</span>
        </div>
        <div class="avg-value ram-val">
          {{ systemStore.averages.memory.toFixed(2) }}%
        </div>
        <div class="avg-footer">{{ $t('reports.allocation') }}</div>
      </div>

      <div class="avg-card disk">
        <div class="avg-header">
          <HardDrive :size="18" class="disk-icon" />
          <span>{{ $t('reports.avg_storage') }}</span>
        </div>
        <div class="avg-value disk-val">
          {{ systemStore.averages.disk.toFixed(2) }}%
        </div>
        <div class="avg-footer">{{ $t('reports.density') }}</div>
      </div>
    </div>

    <div class="chart-container-wrapper tron-card">
      <div class="card-header">
        <h3><ArrowBigRightDash :size="16" class="crimson-accent" /> {{ $t('reports.trendlines') }}</h3>
        <div class="status-pill">
          <span class="dot"></span> {{ $t('reports.realtime_archive') || 'REAL-TIME ARCHIVE' }}
        </div>
      </div>
      <div class="chart-area">
        <Line 
          v-if="!isLoading && systemStore.reports.length > 0"
          :data="chartData" 
          :options="chartOptions" 
        />
        <div v-else-if="isLoading" class="chart-loading">
          <RefreshCw class="spinning" :size="36" />
          <span>{{ $t('reports.syncing') || 'SYNCING WITH GRID...' }}</span>
        </div>
        <div v-else class="no-data">
           <Calendar :size="36" />
           <span>{{ $t('reports.no_data') }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.reports-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border-color);
}

@media (max-width: 900px) {
  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.icon-orb {
  width: 44px;
  height: 44px;
  background: rgba(220, 38, 38, 0.1);
  border: 1px solid rgba(220, 38, 38, 0.3);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--theme-crimson);
}

h1 {
  font-size: 1.35rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  margin: 0;
  color: var(--text-primary);
}

.subtitle {
  font-size: 0.72rem;
  color: var(--text-muted);
  letter-spacing: 0.05em;
  margin: 0.2rem 0 0 0;
  font-family: var(--font-data);
}

.range-selector {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: var(--bg-card);
  padding: 0.4rem 0.75rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
}

.selector-label {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.7rem;
  color: var(--text-secondary);
  font-weight: 700;
  letter-spacing: 0.04em;
}

.range-buttons {
  display: flex;
  gap: 0.25rem;
  background: rgba(0, 0, 0, 0.3);
  padding: 0.2rem;
  border-radius: var(--radius-sm);
}

.range-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 0.3rem 0.65rem;
  font-size: 0.72rem;
  font-weight: 700;
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
  font-family: var(--font-data);
}

.range-btn:hover {
  color: var(--text-primary);
}

.range-btn.active {
  background: var(--theme-crimson);
  color: #ffffff;
}

.refresh-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0.4rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color var(--transition-fast);
}

.refresh-btn:hover:not(:disabled) {
  color: var(--theme-crimson);
}

.refresh-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Custom Inputs */
.custom-inputs {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding-inline-start: 0.75rem;
  border-inline-start: 1px solid var(--border-color);
}

.input-group {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.input-label {
  font-size: 0.65rem;
  font-weight: 700;
  color: var(--text-secondary);
  letter-spacing: 0.04em;
}

.apply-btn {
  background: var(--theme-crimson);
  border: none;
  color: #ffffff;
  padding: 0.35rem 0.75rem;
  border-radius: var(--radius-sm);
  font-weight: 700;
  font-size: 0.72rem;
  cursor: pointer;
  transition: opacity var(--transition-fast);
}

.apply-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.apply-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

@media (max-width: 1200px) {
  .range-selector {
    flex-wrap: wrap;
  }
  .custom-inputs {
    border-inline-start: none;
    padding-inline-start: 0;
    margin-top: 0.5rem;
    width: 100%;
  }
}

/* Averages Grid */
.averages-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 1.25rem;
}

.avg-card {
  position: relative;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  padding: 1.25rem 1.5rem;
  border-radius: var(--radius-sm);
  overflow: hidden;
  transition: border-color var(--transition-fast);
}

.avg-card.cpu { border-inline-start: 4px solid var(--theme-crimson); }
.avg-card.memory { border-inline-start: 4px solid var(--theme-amber); }
.avg-card.disk { border-inline-start: 4px solid var(--theme-sky); }

.avg-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--text-secondary);
  letter-spacing: 0.04em;
  margin-bottom: 0.75rem;
}

.cpu-icon { color: var(--theme-crimson); }
.ram-icon { color: var(--theme-amber); }
.disk-icon { color: var(--theme-sky); }

.avg-value {
  font-size: 2rem;
  font-weight: 800;
  font-family: var(--font-data);
  margin-bottom: 0.5rem;
  letter-spacing: -0.02em;
}

.cpu-val { color: var(--theme-crimson); }
.ram-val { color: var(--theme-amber); }
.disk-val { color: var(--theme-sky); }

.avg-footer {
  font-size: 0.65rem;
  color: var(--text-muted);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  font-weight: 600;
}

/* Chart Area */
.chart-container-wrapper {
  padding: 1.5rem;
  border-radius: var(--radius-sm);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
}

.card-header h3 {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0;
  font-size: 0.95rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: var(--text-primary);
}

.crimson-accent {
  color: var(--theme-crimson);
}

.status-pill {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.65rem;
  font-weight: 700;
  color: var(--text-secondary);
  background: rgba(0, 0, 0, 0.3);
  padding: 0.25rem 0.65rem;
  border-radius: 20px;
  border: 1px solid var(--border-color);
  font-family: var(--font-data);
}

.status-pill .dot {
  width: 6px;
  height: 6px;
  background: var(--theme-emerald);
  border-radius: 50%;
}

.chart-area {
  height: 420px;
  position: relative;
}

.chart-loading, .no-data {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  color: var(--text-muted);
  font-weight: 600;
  font-family: var(--font-data);
  font-size: 0.85rem;
}

.spinning {
  animation: spin 1.5s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
