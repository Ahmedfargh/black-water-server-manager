<script setup>
import { onMounted, onUnmounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSystemStore } from '../stores/system'
import { useAuthStore } from '../stores/auth'
import { Cpu, Zap, HardDrive, Share2, Thermometer } from 'lucide-vue-next'

const { t } = useI18n()
const systemStore = useSystemStore()
const authStore = useAuthStore()

let statsInterval
let wsTemp

const connectWebSockets = () => {
  // CPU Temperature WebSocket
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws/cpu-temperature?token=${authStore.token}`
  
  wsTemp = new WebSocket(wsUrl)
  
  wsTemp.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      // data is an array of sensors: [{SensorKey, Temperature}, ...]
      if (Array.isArray(data) && data.length > 0) {
        // Try to find a package/core sensor or just use the first one
        const sensor = data.find(s => s.SensorKey.includes('package') || s.SensorKey.includes('core')) || data[0]
        systemStore.updateCpuTemp(sensor.Temperature)
      } else if (data.value !== undefined) {
        systemStore.updateCpuTemp(data.value)
      }
    } catch (e) {
      const temp = parseFloat(event.data)
      if (!isNaN(temp)) systemStore.updateCpuTemp(temp)
    }
  }
}

onMounted(() => {
  systemStore.fetchAllStats()
  statsInterval = setInterval(() => {
    systemStore.fetchAllStats()
  }, 2000)
  
  connectWebSockets()
})

onUnmounted(() => {
  clearInterval(statsInterval)
  if (wsTemp) wsTemp.close()
})

const getStrokeDash = (percentage) => {
  const radius = 45
  const circumference = 2 * Math.PI * radius
  return `${(percentage / 100) * circumference} ${circumference}`
}
</script>

<template>
  <div class="dashboard">
    <div class="grid-container">
      
      <!-- CPU HUD -->
      <div class="tron-card hud-item cpu-section">
        <div class="hud-header">
          <Cpu class="glow-cyan" />
          <h3>{{ $t('dashboard.cpu_unit') }}</h3>
        </div>
        <div class="hud-content">
          <div class="gauge-container">
            <svg viewBox="0 0 100 100" class="gauge">
              <circle cx="50" cy="50" r="45" class="bg" />
              <circle cx="50" cy="50" r="45" class="progress" 
                :style="{ strokeDasharray: getStrokeDash(systemStore.cpu.usage) }" />
            </svg>
            <div class="gauge-value">
              <span class="number glow-cyan">{{ systemStore.cpu.usage }}%</span>
              <span class="label">{{ $t('dashboard.load') }}</span>
            </div>
          </div>
          <div class="stats-list">
            <div class="stat-item">
              <span class="label">{{ $t('dashboard.cores') }}</span>
              <span class="value">{{ systemStore.cpu.cores }}</span>
            </div>
            <div class="stat-item">
              <span class="label">{{ $t('dashboard.temp') }}</span>
              <span class="value glow-orange">{{ systemStore.cpu.temp }}°C</span>
            </div>
          </div>
        </div>
        <div class="mini-chart">
           <svg preserveAspectRatio="none" viewBox="0 0 100 30" width="100%" height="40">
             <polyline
               fill="none"
               stroke="var(--neon-cyan)"
               stroke-width="1.5"
               :points="systemStore.history.cpu.map((d, i) => `${(i / 19) * 100},${30 - (d.value / 100) * 30}`).join(' ')"
             />
           </svg>
        </div>
      </div>

      <!-- RAM HUD -->
      <div class="tron-card hud-item ram-section">
        <div class="hud-header">
          <Zap class="glow-cyan" />
          <h3>{{ $t('dashboard.memory_bank') }}</h3>
        </div>
        <div class="hud-content">
          <div class="bar-container">
            <div class="bar-header">
              <span>{{ $t('dashboard.usage') }}</span>
              <span class="glow-cyan">{{ systemStore.ram.usage }}%</span>
            </div>
            <div class="bar-outer">
              <div class="bar-inner" :style="{ width: systemStore.ram.usage + '%' }"></div>
            </div>
          </div>
          <div class="info-grid">
            <div class="info-box">
              <span class="label">{{ $t('dashboard.total') }}</span>
              <span class="value">{{ (systemStore.ram.total / 1024 / 1024 / 1024).toFixed(1) }} GB</span>
            </div>
            <div class="info-box">
              <span class="label">{{ $t('dashboard.used') }}</span>
              <span class="value">{{ (systemStore.ram.used / 1024 / 1024 / 1024).toFixed(1) }} GB</span>
            </div>
          </div>
        </div>
      </div>

      <!-- STORAGE HUD -->
      <div class="tron-card hud-item storage-section">
        <div class="hud-header">
          <HardDrive class="glow-cyan" />
          <h3>{{ $t('dashboard.storage_array') }}</h3>
        </div>
        <div class="hud-content">
          <div class="storage-details">
             <div class="storage-percentage glow-cyan">
               {{ systemStore.disk.usage }}%
             </div>
             <div class="storage-text">
               <span>{{ $t('dashboard.disk_loaded') }}</span>
               <small>{{ (systemStore.disk.used / 1024 / 1024 / 1024).toFixed(1) }}GB / {{ (systemStore.disk.total / 1024 / 1024 / 1024).toFixed(1) }}GB</small>
             </div>
          </div>
        </div>
      </div>

      <!-- NETWORK HUD -->
      <div class="tron-card hud-item network-section">
        <div class="hud-header">
          <Share2 class="glow-cyan" />
          <h3>{{ $t('dashboard.signal_flow') }}</h3>
        </div>
        <div class="hud-content">
           <div class="net-stats">
              <div class="net-item">
                 <span class="label">{{ $t('dashboard.upload') }}</span>
                 <span class="value glow-cyan">{{ (systemStore.network.bytesSent / 1024 / 1024).toFixed(2) }} MB/s</span>
              </div>
              <div class="net-item">
                 <span class="label">{{ $t('dashboard.download') }}</span>
                 <span class="value glow-orange">{{ (systemStore.network.bytesRecv / 1024 / 1024).toFixed(2) }} MB/s</span>
              </div>
           </div>
        </div>
      </div>

    </div>
  </div>
</template>

<style scoped>
.dashboard {
  animation: fadeIn 0.5s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; filter: blur(10px); }
  to { opacity: 1; filter: blur(0); }
}

.grid-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 2rem;
}

@media (max-width: 768px) {
  .grid-container {
    gap: 1.5rem;
  }
}

.hud-item {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.hud-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  border-bottom: 1px solid rgba(0, 242, 255, 0.1);
  padding-bottom: 1rem;
}

.hud-header h3 {
  font-size: 1.1rem;
  margin: 0;
}

.hud-content {
  flex: 1;
}

/* Gauge Styles */
.gauge-container {
  position: relative;
  width: 120px;
  height: 120px;
  margin: 0 auto;
}

.gauge {
  transform: rotate(-90deg);
}

.gauge circle {
  fill: none;
  stroke-width: 8;
}

.gauge .bg {
  stroke: rgba(0, 242, 255, 0.05);
}

.gauge .progress {
  stroke: var(--neon-cyan);
  stroke-linecap: round;
  transition: stroke-dasharray 0.5s ease;
  filter: drop-shadow(0 0 5px var(--neon-cyan-glow));
}

.gauge-value {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
}

.gauge-value .number {
  display: block;
  font-size: 1.4rem;
  font-weight: 700;
  font-family: var(--font-data);
}

.gauge-value .label {
  font-size: 0.7rem;
  color: var(--text-secondary);
  text-transform: uppercase;
}

/* Stats List */
.stats-list {
  margin-top: 1.5rem;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.stat-item .label {
  font-size: 0.7rem;
  color: var(--text-secondary);
}

.stat-item .value {
  font-size: 1.1rem;
  font-weight: 600;
  font-family: var(--font-data);
}

/* Bar styles */
.bar-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.bar-header {
  display: flex;
  justify-content: space-between;
  font-size: 0.8rem;
}

.bar-outer {
  height: 8px;
  background: rgba(0, 242, 255, 0.05);
  border-radius: 4px;
  overflow: hidden;
  border: 1px solid rgba(0, 242, 255, 0.1);
}

.bar-inner {
  height: 100%;
  background: var(--neon-cyan);
  box-shadow: 0 0 10px var(--neon-cyan-glow);
  transition: width 0.5s ease;
}

.info-grid {
  margin-top: 2rem;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

.info-box {
  background: rgba(0, 242, 255, 0.03);
  padding: 0.8rem;
  border-left: 2px solid var(--neon-cyan);
}

/* Storage details */
.storage-details {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.storage-percentage {
  font-size: 3rem;
  font-weight: 700;
  font-family: var(--font-data);
}

.storage-text {
  display: flex;
  flex-direction: column;
}

.storage-text span {
  font-size: 1rem;
  letter-spacing: 1px;
}

.storage-text small {
  color: var(--text-secondary);
  font-family: var(--font-data);
}

/* Mini Chart */
.mini-chart {
  margin-top: auto;
  opacity: 0.6;
}

.mini-chart polyline {
  filter: drop-shadow(0 0 3px var(--neon-cyan-glow));
}

/* Network */
.net-stats {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.net-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.8rem;
  background: rgba(255, 140, 0, 0.03);
  border-left: 2px solid var(--neon-orange);
}

.net-item:first-child {
  background: rgba(0, 242, 255, 0.03);
  border-color: var(--neon-cyan);
}
</style>
