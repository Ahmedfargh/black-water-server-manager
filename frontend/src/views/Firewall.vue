<script setup>
import { onMounted, computed } from 'vue'
import { 
  ShieldCheck, 
  ShieldAlert, 
  List, 
  RefreshCw,
  Power
} from 'lucide-vue-next'
import { useFirewallStore } from '../stores/firewall'
import { useToastStore } from '../stores/toast'

const firewallStore = useFirewallStore()
const toast = useToastStore()

const isActive = computed(() => {
  if (!firewallStore.status) return false
  const status = firewallStore.status.toLowerCase()
  // Must check 'not running' BEFORE 'running' because 'not running'.includes('running') is true
  if (status.includes('not running') || status.includes('inactive')) return false
  return status.includes('running') || status.includes('active')
})

const actionText = computed(() => isActive.value ? 'DEACTIVATE' : 'ACTIVATE')

onMounted(() => {
  firewallStore.fetchStatus()
  firewallStore.fetchRules()
})

const handleToggle = async () => {
  const action = actionText.value
  const targetState = !isActive.value
  
    try {
      toast.info(`INITIATING FIREWALL ${action} SEQUENCE...`)
      await firewallStore.toggleFirewall(targetState)
      toast.success(`FIREWALL ${action}D SUCCESSFULLY`)
    } catch (err) {
      console.error('Firewall toggle error:', err)
      toast.error(`PROTOCOL FAILED: Unable to ${action} firewall.`)
    }
  
}
</script>

<template>
  <div class="firewall-view">
    <div class="header-row">
      <h2 class="glow-cyan">FIREWALL DEFENSE GRID</h2>
      <button @click="firewallStore.fetchRules(); toast.info('RESCANNING RULES...')" class="tron-btn">
        <RefreshCw :size="18" />
        RESCAN RULES
      </button>
    </div>

    <!-- Status Card -->
    <div class="tron-card status-card" :class="{ 'active': isActive, 'inactive': !isActive }">
      <div class="status-content">
        <div class="status-icon-wrap">
          <ShieldCheck v-if="isActive" :size="64" class="glow-cyan" />
          <ShieldAlert v-else :size="64" class="glow-orange" />
        </div>
        <div class="status-info">
          <h3>SYSTEM STATUS: <span class="status-text">{{ (firewallStore.status || 'UNKNOWN').toUpperCase() }}</span></h3>
          <p v-if="isActive">Defense grid is operational. All incoming traffic is being filtered.</p>
          <p v-else>Defense grid is OFFLINE. System is vulnerable to external signals.</p>
        </div>
        <button @click.stop="handleToggle" class="toggle-btn" :class="{ 'on': isActive }">
          <Power :size="24" />
          <span>{{ actionText }}</span>
        </button>
      </div>
    </div>

    <!-- Rules List -->
    <div class="tron-card rules-card">
      <div class="card-header">
        <List :size="20" class="glow-cyan" />
        <h3>ACTIVE SECURITY CONFIGURATION</h3>
      </div>
      <div class="rules-content">
        <pre v-if="firewallStore.rules" class="raw-output font-data">{{ firewallStore.rules }}</pre>
        <div v-else-if="firewallStore.loading" class="loading-state">
          <RefreshCw :size="32" class="spinner" />
          <p>SCANNING GRID SECURITY...</p>
        </div>
        <div v-else class="empty-msg">NO SECURITY CONFIGURATION IDENTIFIED.</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.firewall-view {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* Status Card */
.status-card {
  padding: 2.5rem;
  border-left-width: 6px;
}

.status-card.active { border-left-color: var(--neon-cyan); }
.status-card.inactive { border-left-color: var(--neon-orange); }

.status-content {
  display: flex;
  align-items: center;
  gap: 3rem;
}

@media (max-width: 768px) {
  .status-content {
    flex-direction: column;
    text-align: center;
    gap: 1.5rem;
  }
  .toggle-btn { margin-left: 0 !important; }
}

.status-info h3 {
  font-size: 1.5rem;
  margin-bottom: 0.5rem;
}

.status-card.active .status-text { color: var(--neon-cyan); text-shadow: var(--text-glow); }
.status-card.inactive .status-text { color: var(--neon-orange); text-shadow: 0 0 10px var(--neon-orange-glow); }

.status-info p {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.toggle-btn {
  margin-left: auto;
  background: transparent;
  border: 1px solid var(--neon-orange);
  color: var(--neon-orange);
  padding: 1.5rem 2.5rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  font-family: var(--font-header);
  font-weight: 700;
  letter-spacing: 2px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.toggle-btn.on {
  border-color: var(--neon-cyan);
  color: var(--neon-cyan);
}

.toggle-btn:hover {
  box-shadow: 0 0 20px currentColor;
  background: rgba(255, 255, 255, 0.02);
}

/* Rules Display */
.rules-card {
  display: flex;
  flex-direction: column;
  min-height: 300px;
}

.card-header {
  padding: 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  border-bottom: 1px solid rgba(0, 242, 255, 0.1);
}

.rules-content {
  padding: 1.5rem;
  flex: 1;
  background: rgba(0, 0, 0, 0.3);
}

.raw-output {
  white-space: pre-wrap;
  word-wrap: break-word;
  color: var(--text-primary);
  font-size: 0.9rem;
  line-height: 1.5;
  background: rgba(255, 255, 255, 0.02);
  padding: 1.5rem;
  border-radius: 4px;
  border: 1px solid rgba(0, 242, 255, 0.1);
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: 4rem;
  color: var(--neon-cyan);
}

.spinner {
  animation: rotate 2s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.empty-msg {
  text-align: center;
  padding: 4rem;
  color: var(--text-secondary);
  font-style: italic;
  letter-spacing: 2px;
}
</style>
