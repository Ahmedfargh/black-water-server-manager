<script setup>
import { onMounted, onUnmounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSystemStore } from '../stores/system'
import { useAuthStore } from '../stores/auth'
import { Cpu, Zap, HardDrive, Share2, Thermometer, Flame, Layers } from 'lucide-vue-next'

const { t } = useI18n()
const systemStore = useSystemStore()
const authStore = useAuthStore()

let statsInterval
let wsTemp

const connectWebSockets = () => {
  // CPU & Hardware Temperature WebSocket
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws/cpu-temperature?token=${authStore.token}`
  
  wsTemp = new WebSocket(wsUrl)
  
  wsTemp.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (Array.isArray(data) && data.length > 0) {
        // Find CPU temp
        const cpuSensor = data.find(s => s.SensorKey.includes('package') || s.SensorKey.includes('core')) || data[0]
        if (cpuSensor) systemStore.updateCpuTemp(cpuSensor.Temperature)

        // Find and update NVMe / disk sensors if present in live broadcast
        const nvmeSensor = data.find(s => s.SensorKey.toLowerCase().includes('nvme'))
        if (nvmeSensor && systemStore.disk.devices) {
          systemStore.disk.devices.forEach(dev => {
            if (dev.name.toLowerCase().includes('nvme') && (!dev.temperature || Math.abs(dev.temperature - nvmeSensor.Temperature) > 0.1)) {
              dev.temperature = nvmeSensor.Temperature
            }
          })
        }
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
  }, 2500)
  
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

const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const formatRate = (bytesPerSec) => {
  if (!bytesPerSec || isNaN(bytesPerSec) || bytesPerSec <= 0) return '0 B/s'
  const k = 1024
  const units = ['B/s', 'KB/s', 'MB/s', 'GB/s', 'TB/s']
  const i = Math.floor(Math.log(bytesPerSec) / Math.log(k))
  const idx = Math.min(i, units.length - 1)
  const val = (bytesPerSec / Math.pow(k, idx)).toFixed(idx >= 2 ? 2 : 1)
  return `${parseFloat(val)} ${units[idx]}`
}

const getTempClass = (temp) => {
  if (temp === null || temp === undefined) return 'temp-na'
  if (temp >= 75) return 'temp-danger'
  if (temp >= 60) return 'temp-warning'
  return 'temp-cool'
}

const totalPartitionsCount = computed(() => {
  if (!systemStore.disk.devices) return 0
  return systemStore.disk.devices.reduce((acc, dev) => acc + (dev.partitions?.length || 0), 0)
})
</script>

<template>
  <div class="dashboard">
    <div class="grid-container">
      
      <!-- CPU HUD -->
      <div class="tron-card hud-item cpu-section">
        <div class="hud-header">
          <Cpu class="crimson-accent" />
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
              <span class="number">{{ systemStore.cpu.usage }}%</span>
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
              <span class="value temp-val">{{ systemStore.cpu.temp }}°C</span>
            </div>
          </div>
        </div>
        <div class="mini-chart">
           <svg preserveAspectRatio="none" viewBox="0 0 100 30" width="100%" height="40">
             <polyline
               fill="none"
               stroke="var(--rdr-crimson)"
               stroke-width="1.5"
               :points="systemStore.history.cpu.map((d, i) => `${(i / 19) * 100},${30 - (d.value / 100) * 30}`).join(' ')"
             />
           </svg>
        </div>
      </div>

      <!-- RAM HUD -->
      <div class="tron-card hud-item ram-section">
        <div class="hud-header">
          <Zap class="amber-accent" />
          <h3>{{ $t('dashboard.memory_bank') }}</h3>
        </div>
        <div class="hud-content">
          <div class="bar-container">
            <div class="bar-header">
              <span>{{ $t('dashboard.usage') }}</span>
              <span class="amber-accent">{{ systemStore.ram.usage }}%</span>
            </div>
            <div class="bar-outer">
              <div class="bar-inner ram-bar" :style="{ width: systemStore.ram.usage + '%' }"></div>
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

      <!-- STORAGE HUD SUMMARY -->
      <div class="tron-card hud-item storage-section">
        <div class="hud-header">
          <HardDrive class="crimson-accent" />
          <h3>{{ $t('dashboard.storage_array') }}</h3>
        </div>
        <div class="hud-content">
          <div class="storage-details">
             <div class="storage-percentage">
               {{ systemStore.disk.usage }}%
             </div>
             <div class="storage-text">
               <span>{{ $t('dashboard.disk_loaded') }}</span>
               <small>{{ formatBytes(systemStore.disk.used) }} / {{ formatBytes(systemStore.disk.total) }}</small>
             </div>
          </div>
          <div class="storage-bar-outer">
             <div class="storage-bar-inner" :style="{ width: `${Math.min(systemStore.disk.usage, 100)}%` }"></div>
          </div>
          <div class="storage-meta-chips">
            <span class="meta-chip">
              <HardDrive :size="12" />
              {{ systemStore.disk.devices?.length || 0 }} {{ $t('dashboard.connected_disks') }}
            </span>
            <span class="meta-chip">
              <Layers :size="12" />
              {{ totalPartitionsCount }} {{ $t('dashboard.partitions') }}
            </span>
          </div>
        </div>
      </div>

      <!-- NETWORK HUD -->
      <div class="tron-card hud-item network-section">
        <div class="hud-header">
          <Share2 class="sky-accent" />
          <h3>{{ $t('dashboard.signal_flow') }}</h3>
        </div>
        <div class="hud-content">
           <div class="net-stats">
              <div class="net-item">
                 <div class="net-item-meta">
                   <span class="label">{{ $t('dashboard.upload') }}</span>
                   <span class="net-total font-data" dir="ltr">{{ $t('dashboard.total') }}: {{ formatBytes(systemStore.network.totalSent || systemStore.network.bytesSent) }}</span>
                 </div>
                 <span class="value crimson-accent font-data" dir="ltr">{{ formatRate(systemStore.network.sentRate) }}</span>
              </div>
              <div class="net-item">
                 <div class="net-item-meta">
                   <span class="label">{{ $t('dashboard.download') }}</span>
                   <span class="net-total font-data" dir="ltr">{{ $t('dashboard.total') }}: {{ formatBytes(systemStore.network.totalRecv || systemStore.network.bytesRecv) }}</span>
                 </div>
                 <span class="value amber-accent font-data" dir="ltr">{{ formatRate(systemStore.network.recvRate) }}</span>
              </div>
           </div>
        </div>
      </div>

    </div>

    <!-- DETAILED PHYSICAL DISKS & PARTITIONS HUD -->
    <div class="disks-detailed-section tron-card">
      <div class="section-title-row">
        <div class="title-left">
          <HardDrive class="crimson-accent" :size="20" />
          <h2>{{ $t('dashboard.connected_disks') }} &amp; {{ $t('dashboard.partitions') }}</h2>
        </div>
        <div class="disk-count-badge font-data">
          {{ systemStore.disk.devices?.length || 0 }} DRIVES // {{ totalPartitionsCount }} PARTITIONS
        </div>
      </div>

      <!-- Physical Disks Grid -->
      <div class="physical-disks-grid" v-if="systemStore.disk.devices && systemStore.disk.devices.length > 0">
        <div 
          v-for="device in systemStore.disk.devices" 
          :key="device.name" 
          class="physical-disk-card"
        >
          <!-- Disk Card Top Bar -->
          <div class="disk-card-header">
            <div class="disk-main-info">
              <div class="disk-icon-box" :class="device.type.toLowerCase()">
                <HardDrive :size="20" />
              </div>
              <div class="disk-names">
                <div class="disk-title-line">
                  <span class="disk-device font-data">{{ device.device }}</span>
                  <span class="disk-type-badge" :class="device.type.toLowerCase()">{{ device.type }}</span>
                  <span v-if="device.is_removable" class="disk-removable-badge font-data">USB</span>
                  <span class="disk-state-dot" :title="device.state"></span>
                </div>
                <span class="disk-model">{{ device.model }}</span>
              </div>
            </div>

            <!-- Temperature & Capacity Badge -->
            <div class="disk-metrics-side">
              <div class="disk-temp-pill font-data" :class="getTempClass(device.temperature)" :title="$t('dashboard.temperature')">
                <Flame v-if="device.temperature !== null" :size="13" />
                <Thermometer v-else :size="13" />
                <span>{{ device.temperature !== null ? `${device.temperature.toFixed(1)}°C` : '--' }}</span>
              </div>
              <span class="disk-total-capacity font-data">{{ formatBytes(device.size) }}</span>
            </div>
          </div>

          <!-- Partitions Breakdown -->
          <div class="partitions-list">
            <div class="partitions-header font-data">
              <span>{{ $t('dashboard.partitions') }}</span>
              <span>MOUNT / STATUS</span>
            </div>

            <div 
              v-for="part in device.partitions" 
              :key="part.name" 
              class="partition-row"
              :class="{ 'is-mounted': part.is_mounted, 'is-unmounted': !part.is_mounted }"
            >
              <div class="partition-ident">
                <div class="part-device-badge">
                  <span class="part-name font-data">{{ part.name }}</span>
                  <span class="part-fstype font-data">{{ part.fstype }}</span>
                </div>
              </div>

              <div class="partition-mount-state">
                <div v-if="part.is_mounted" class="mounted-details">
                  <div class="mount-path-bar">
                    <span class="mount-path font-data" :title="part.mountpoint">{{ part.mountpoint }}</span>
                    <span class="mount-pct font-data">{{ part.used_percent.toFixed(1) }}%</span>
                  </div>
                  <div class="part-progress-track">
                    <div 
                      class="part-progress-fill" 
                      :style="{ width: `${Math.min(part.used_percent, 100)}%` }"
                      :class="{ 'high-usage': part.used_percent > 85, 'warn-usage': part.used_percent > 70 }"
                    ></div>
                  </div>
                  <span class="mount-usage-text font-data">{{ formatBytes(part.used_bytes) }} / {{ formatBytes(part.total_bytes) }}</span>
                </div>
                <div v-else class="unmounted-details">
                  <span class="unmounted-tag font-data">{{ $t('dashboard.unmounted') }}</span>
                  <span class="unmounted-size font-data">{{ formatBytes(part.size) }}</span>
                </div>
              </div>
            </div>

            <div v-if="!device.partitions || device.partitions.length === 0" class="no-partitions-msg font-data">
              Raw block device (no partition table)
            </div>
          </div>
        </div>
      </div>

      <div v-else class="no-disks-msg font-data">
        <HardDrive :size="28" class="crimson-accent" />
        <span>Scanning hardware bus for connected storage devices...</span>
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
  stroke-width: 7;
}

.gauge .bg {
  stroke: rgba(255, 255, 255, 0.06);
}

.gauge .progress {
  stroke: var(--rdr-crimson);
  stroke-linecap: round;
  transition: stroke-dasharray 0.5s ease;
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
  color: #ffffff;
}

.gauge-value .label {
  font-size: 0.7rem;
  color: var(--text-muted);
  text-transform: uppercase;
  font-family: var(--font-data);
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
  font-size: 0.72rem;
  color: var(--text-muted);
  font-family: var(--font-data);
}

.stat-item .value {
  font-size: 1.1rem;
  font-weight: 600;
  font-family: var(--font-data);
  color: #ffffff;
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
  font-family: var(--font-data);
}

.bar-outer {
  height: 6px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 3px;
  overflow: hidden;
  border: 1px solid var(--border-subtle);
}

.bar-inner {
  height: 100%;
  background: var(--rdr-crimson);
  transition: width 0.4s ease;
}

.info-grid {
  margin-top: 2rem;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

.info-box {
  background: rgba(255, 255, 255, 0.02);
  padding: 0.8rem;
  border-inline-start: 3px solid var(--rdr-crimson);
  border-radius: 2px;
}

/* Storage details */
.storage-details {
  display: flex;
  align-items: center;
  gap: 1.25rem;
}

.storage-percentage {
  font-size: 2.5rem;
  font-weight: 700;
  font-family: var(--font-data);
  color: #ffffff;
}

.storage-text {
  display: flex;
  flex-direction: column;
}

.storage-text span {
  font-size: 0.9rem;
  font-weight: 600;
  letter-spacing: 0.5px;
  color: var(--text-primary);
}

.storage-text small {
  color: var(--text-muted);
  font-family: var(--font-data);
  font-size: 0.78rem;
}

.storage-bar-outer {
  margin-top: 1rem;
  height: 6px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 3px;
  overflow: hidden;
  border: 1px solid var(--border-subtle);
}

.storage-bar-inner {
  height: 100%;
  background: var(--rdr-crimson);
  transition: width 0.4s ease;
}

.storage-meta-chips {
  margin-top: 1rem;
  display: flex;
  gap: 0.6rem;
  flex-wrap: wrap;
}

.meta-chip {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.25rem 0.6rem;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border-subtle);
  font-size: 0.72rem;
  color: var(--text-secondary);
  font-family: var(--font-data);
}

.ram-bar {
  background: var(--rdr-amber);
}

.crimson-accent { color: var(--rdr-crimson); }
.amber-accent { color: var(--rdr-amber); }
.sky-accent { color: #38bdf8; }
.temp-val { color: var(--rdr-amber); }

/* Mini Chart */
.mini-chart {
  margin-top: auto;
  opacity: 0.8;
}

.mini-chart polyline {
  stroke: var(--rdr-crimson);
}

/* Network */
.net-stats {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.net-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.8rem;
  background: rgba(217, 119, 6, 0.04);
  border-inline-start: 3px solid var(--rdr-amber);
  border-radius: 2px;
}

.net-item:first-child {
  background: rgba(220, 38, 38, 0.04);
  border-inline-start-color: var(--rdr-crimson);
}

.net-item-meta {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.net-item-meta .label {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--text-muted);
}

.net-total {
  font-size: 0.68rem;
  color: var(--text-muted);
  opacity: 0.75;
}

/* ==========================================================
   DETAILED PHYSICAL DISKS & PARTITIONS HUD
   ========================================================== */
.disks-detailed-section {
  margin-top: 2rem;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.section-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border-subtle);
}

.title-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.title-left h2 {
  font-size: 1.05rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  margin: 0;
  color: var(--text-primary);
}

.disk-count-badge {
  font-size: 0.72rem;
  color: var(--rdr-amber);
  background: rgba(217, 119, 6, 0.1);
  border: 1px solid rgba(217, 119, 6, 0.3);
  padding: 0.25rem 0.65rem;
  border-radius: var(--radius-sm);
  font-weight: 700;
}

/* Physical Disks Grid */
.physical-disks-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
  gap: 1.25rem;
}

@media (max-width: 768px) {
  .physical-disks-grid {
    grid-template-columns: 1fr;
  }
}

.physical-disk-card {
  background: rgba(12, 13, 17, 0.7);
  border: 1px solid var(--border-subtle);
  border-inline-start: 4px solid var(--rdr-crimson);
  border-radius: var(--radius-sm);
  padding: 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  transition: border-color var(--transition-fast);
}

.physical-disk-card:hover {
  border-color: rgba(255, 255, 255, 0.16);
}

/* Disk Card Header */
.disk-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 0.75rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.disk-main-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.disk-icon-box {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border-subtle);
  color: var(--text-secondary);
}

.disk-icon-box.nvme {
  color: var(--rdr-crimson);
  background: rgba(220, 38, 38, 0.1);
  border-color: rgba(220, 38, 38, 0.3);
}

.disk-icon-box.hdd {
  color: var(--rdr-amber);
  background: rgba(217, 119, 6, 0.1);
  border-color: rgba(217, 119, 6, 0.3);
}

.disk-names {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.disk-title-line {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.disk-device {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-primary);
}

.disk-type-badge {
  font-size: 0.65rem;
  padding: 0.1rem 0.4rem;
  border-radius: 2px;
  font-weight: 700;
  text-transform: uppercase;
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-secondary);
}

.disk-type-badge.nvme {
  background: rgba(220, 38, 38, 0.15);
  color: var(--rdr-crimson);
  border: 1px solid rgba(220, 38, 38, 0.3);
}

.disk-type-badge.hdd {
  background: rgba(217, 119, 6, 0.15);
  color: var(--rdr-amber);
  border: 1px solid rgba(217, 119, 6, 0.3);
}

.disk-removable-badge {
  font-size: 0.62rem;
  padding: 0.1rem 0.35rem;
  border-radius: 2px;
  background: rgba(56, 189, 248, 0.12);
  color: #38bdf8;
  border: 1px solid rgba(56, 189, 248, 0.3);
  font-weight: 700;
}

.disk-state-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--linux-green);
  box-shadow: 0 0 6px var(--linux-green);
}

.disk-model {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.disk-metrics-side {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.25rem;
}

.disk-temp-pill {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.18rem 0.5rem;
  border-radius: var(--radius-sm);
  font-size: 0.72rem;
  font-weight: 700;
  border: 1px solid transparent;
}

.temp-cool {
  background: rgba(16, 185, 129, 0.12);
  color: #10b981;
  border-color: rgba(16, 185, 129, 0.3);
}

.temp-warning {
  background: rgba(217, 119, 6, 0.15);
  color: var(--rdr-amber);
  border-color: rgba(217, 119, 6, 0.35);
}

.temp-danger {
  background: rgba(220, 38, 38, 0.2);
  color: #ef4444;
  border-color: rgba(220, 38, 38, 0.4);
}

.temp-na {
  background: rgba(255, 255, 255, 0.04);
  color: var(--text-muted);
  border-color: var(--border-subtle);
}

.disk-total-capacity {
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--text-secondary);
}

/* Partitions List */
.partitions-list {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.partitions-header {
  display: flex;
  justify-content: space-between;
  font-size: 0.65rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding-bottom: 0.25rem;
}

.partition-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0.65rem;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.04);
  transition: background var(--transition-fast);
}

.partition-row.is-mounted {
  border-inline-start: 3px solid #10b981;
}

.partition-row.is-unmounted {
  border-inline-start: 3px solid rgba(255, 255, 255, 0.2);
  opacity: 0.75;
}

.partition-ident {
  display: flex;
  align-items: center;
}

.part-device-badge {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
}

.part-name {
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--text-primary);
}

.part-fstype {
  font-size: 0.65rem;
  color: var(--text-muted);
  text-transform: uppercase;
}

.partition-mount-state {
  flex: 1;
  max-width: 200px;
}

.mounted-details {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.mount-path-bar {
  display: flex;
  justify-content: space-between;
  font-size: 0.72rem;
}

.mount-path {
  color: #10b981;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 140px;
}

.mount-pct {
  font-weight: 700;
  color: var(--text-secondary);
}

.part-progress-track {
  height: 4px;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 2px;
  overflow: hidden;
}

.part-progress-fill {
  height: 100%;
  background: #10b981;
  transition: width 0.3s ease;
}

.part-progress-fill.warn-usage {
  background: var(--rdr-amber);
}

.part-progress-fill.high-usage {
  background: var(--rdr-crimson);
}

.mount-usage-text {
  font-size: 0.65rem;
  color: var(--text-muted);
  text-align: end;
}

.unmounted-details {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.5rem;
}

.unmounted-tag {
  font-size: 0.65rem;
  padding: 0.1rem 0.35rem;
  border-radius: 2px;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-muted);
  border: 1px solid var(--border-subtle);
  text-transform: uppercase;
}

.unmounted-size {
  font-size: 0.72rem;
  color: var(--text-secondary);
}

.no-partitions-msg, .no-disks-msg {
  text-align: center;
  padding: 1.5rem;
  color: var(--text-muted);
  font-size: 0.78rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}
</style>
