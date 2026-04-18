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
        borderColor: '#00f2ff',
        backgroundColor: 'rgba(0, 242, 255, 0.1)',
        fill: true,
        tension: 0.4,
        pointRadius: 0,
        pointHoverRadius: 5,
        borderWidth: 2
      },
      {
        label: `${t('reports.ram_usage_p') || 'RAM Usage (%)'}`,
        data: reports.map(r => r.memory_usage),
        borderColor: '#ff8c00',
        backgroundColor: 'rgba(255, 140, 0, 0.1)',
        fill: true,
        tension: 0.4,
        pointRadius: 0,
        pointHoverRadius: 5,
        borderWidth: 2
      },
      {
        label: `${t('reports.disk_usage_p') || 'Disk Usage (%)'}`,
        data: reports.map(r => r.disk_usage),
        borderColor: '#7000ff',
        backgroundColor: 'rgba(112, 0, 255, 0.1)',
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
        color: '#a0aec0',
        usePointStyle: true,
        padding: 20,
        font: {
          family: 'Inter, sans-serif'
        }
      }
    },
    tooltip: {
      mode: 'index',
      intersect: false,
      backgroundColor: 'rgba(5, 7, 10, 0.9)',
      titleColor: '#00f2ff',
      bodyColor: '#fff',
      borderColor: 'rgba(0, 242, 255, 0.2)',
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
        borderDash: [5, 5]
      },
      ticks: {
        color: '#a0aec0',
        callback: (value) => value + '%'
      }
    },
    x: {
      grid: {
        display: false
      },
      ticks: {
        color: '#a0aec0',
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
          <BarChart3 class="glow-cyan" />
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
          <RefreshCw :size="18" :class="{ spinning: isLoading }" />
        </button>
      </div>
    </header>

    <div class="averages-grid">
      <div class="avg-card cpu">
        <div class="card-glow"></div>
        <div class="avg-header">
          <Cpu :size="20" class="glow-cyan" />
          <span>{{ $t('reports.avg_cpu') }}</span>
        </div>
        <div class="avg-value glow-cyan">
          {{ systemStore.averages.cpu.toFixed(2) }}%
        </div>
        <div class="avg-footer">{{ $t('reports.throughput') }}</div>
      </div>

      <div class="avg-card memory">
        <div class="card-glow"></div>
        <div class="avg-header">
          <Zap :size="20" class="glow-orange" />
          <span>{{ $t('reports.avg_mem') }}</span>
        </div>
        <div class="avg-value glow-orange">
          {{ systemStore.averages.memory.toFixed(2) }}%
        </div>
        <div class="avg-footer">{{ $t('reports.allocation') }}</div>
      </div>

      <div class="avg-card disk">
        <div class="card-glow"></div>
        <div class="avg-header">
          <HardDrive :size="20" class="glow-purple" />
          <span>{{ $t('reports.avg_storage') }}</span>
        </div>
        <div class="avg-value glow-purple">
          {{ systemStore.averages.disk.toFixed(2) }}%
        </div>
        <div class="avg-footer">{{ $t('reports.density') }}</div>
      </div>
    </div>

    <div class="chart-container-wrapper tron-card">
      <div class="card-header">
        <h3> <ArrowBigRightDash :size="16" class="glow-cyan" /> {{ $t('reports.trendlines') }}</h3>
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
          <RefreshCw class="spinning" :size="48" />
          <span>{{ $t('reports.syncing') || 'SYNCING WITH GRID...' }}</span>
        </div>
        <div v-else class="no-data">
           <Calendar :size="48" />
           <span>{{ $t('reports.no_data') }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.reports-page {
  animation: slideUp 0.6s cubic-bezier(0.23, 1, 0.32, 1);
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(30px); filter: blur(10px); }
  to { opacity: 1; transform: translateY(0); filter: blur(0); }
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 2.5rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid rgba(0, 242, 255, 0.1);
}

@media (max-width: 900px) {
  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1.5rem;
  }
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.icon-orb {
  width: 54px;
  height: 54px;
  background: rgba(0, 242, 255, 0.05);
  border: 1px solid rgba(0, 242, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 0 20px rgba(0, 242, 255, 0.1);
}

h1 {
  font-size: 1.8rem;
  font-weight: 800;
  letter-spacing: 4px;
  margin: 0;
  text-shadow: 0 0 15px rgba(0, 242, 255, 0.3);
}

.subtitle {
  font-size: 0.75rem;
  color: var(--text-secondary);
  letter-spacing: 2px;
  margin: 0.3rem 0 0 0;
}

.range-selector {
  display: flex;
  align-items: center;
  gap: 1rem;
  background: rgba(255, 255, 255, 0.02);
  padding: 0.5rem 1rem;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.selector-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.7rem;
  color: var(--text-secondary);
  font-weight: 600;
  letter-spacing: 1px;
}

.range-buttons {
  display: flex;
  gap: 0.3rem;
  background: rgba(0, 0, 0, 0.2);
  padding: 0.2rem;
  border-radius: 6px;
}

.range-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 0.4rem 0.8rem;
  font-size: 0.75rem;
  font-weight: 700;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.range-btn:hover {
  color: var(--text-primary);
}

.range-btn.active {
  background: rgba(0, 242, 255, 0.1);
  color: var(--neon-cyan);
  box-shadow: 0 0 10px rgba(0, 242, 255, 0.2);
}

.refresh-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.2s;
}

.refresh-btn:hover:not(:disabled) {
  color: var(--neon-cyan);
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Custom Inputs */
.custom-inputs {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding-left: 1rem;
  border-left: 1px solid rgba(0, 242, 255, 0.1);
  animation: fadeInRight 0.3s ease-out;
}

@keyframes fadeInRight {
  from { opacity: 0; transform: translateX(10px); }
  to { opacity: 1; transform: translateX(0); }
}

.input-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.input-label {
  font-size: 0.6rem;
  font-weight: 700;
  color: var(--text-secondary);
  letter-spacing: 1px;
}

.tron-input {
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(0, 242, 255, 0.2);
  color: var(--text-primary);
  font-family: var(--font-data);
  font-size: 0.8rem;
  padding: 0.4rem 0.6rem;
  border-radius: 4px;
  outline: none;
  transition: all 0.2s;
}

.tron-input:focus {
  border-color: var(--neon-cyan);
  box-shadow: 0 0 10px rgba(0, 242, 255, 0.1);
}

/* For browser support of dark mode in inputs */
::-webkit-calendar-picker-indicator {
  filter: invert(1);
  cursor: pointer;
}

.apply-btn {
  background: rgba(0, 242, 255, 0.1);
  border: 1px solid rgba(0, 242, 255, 0.3);
  color: var(--neon-cyan);
  padding: 0.4rem 1rem;
  border-radius: 4px;
  font-weight: 800;
  font-size: 0.7rem;
  cursor: pointer;
  transition: all 0.2s;
}

.apply-btn:hover:not(:disabled) {
  background: var(--neon-cyan);
  color: #000;
  box-shadow: 0 0 15px var(--neon-cyan-glow);
}

.apply-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

@media (max-width: 1200px) {
  .range-selector {
    flex-wrap: wrap;
    justify-content: center;
  }
  .custom-inputs {
    border-left: none;
    padding-left: 0;
    margin-top: 0.5rem;
    width: 100%;
    justify-content: center;
  }
}

/* Averages Grid */
.averages-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2.5rem;
}

.avg-card {
  position: relative;
  background: var(--bg-card);
  border: 1px solid rgba(255, 255, 255, 0.05);
  padding: 1.8rem;
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.3s ease;
}

.avg-card:hover {
  transform: translateY(-5px);
  border-color: rgba(255, 255, 255, 0.1);
}

.card-glow {
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle at center, rgba(255, 255, 255, 0.03) 0%, transparent 70%);
  pointer-events: none;
}

.avg-header {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-secondary);
  letter-spacing: 1px;
  margin-bottom: 1rem;
}

.avg-value {
  font-size: 2.5rem;
  font-weight: 800;
  font-family: var(--font-data);
  margin-bottom: 1rem;
}

.avg-footer {
  font-size: 0.65rem;
  color: rgba(255, 255, 255, 0.3);
  letter-spacing: 1px;
  text-transform: uppercase;
}

.avg-card.cpu { border-left: 3px solid var(--neon-cyan); }
.avg-card.memory { border-left: 3px solid var(--neon-orange); }
.avg-card.disk { border-left: 3px solid var(--neon-purple); }

/* Chart Area */
.chart-container-wrapper {
  padding: 2rem;
  border-radius: 12px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.card-header h3 {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  margin: 0;
  font-size: 1rem;
  letter-spacing: 2px;
}

.status-pill {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.65rem;
  font-weight: 700;
  color: var(--text-secondary);
  background: rgba(0, 0, 0, 0.2);
  padding: 0.3rem 0.8rem;
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.status-pill .dot {
  width: 6px;
  height: 6px;
  background: var(--neon-cyan);
  border-radius: 50%;
  box-shadow: 0 0 5px var(--neon-cyan);
}

.chart-area {
  height: 450px;
  position: relative;
}

.chart-loading, .no-data {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1.5rem;
  color: var(--text-secondary);
  font-weight: 600;
  letter-spacing: 1px;
}

.spinning {
  animation: spin 2s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.glow-purple {
  color: #7000ff;
  filter: drop-shadow(0 0 5px rgba(112, 0, 255, 0.5));
}
</style>
