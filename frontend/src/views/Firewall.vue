<script setup>
import { onMounted, computed, ref } from 'vue'
import { 
  ShieldCheck, 
  ShieldAlert, 
  List, 
  RefreshCw,
  Power,
  Ban,
  Unlock,
  Globe
} from 'lucide-vue-next'
import { useFirewallStore } from '../stores/firewall'
import { useToastStore } from '../stores/toast'

const firewallStore = useFirewallStore()
const toast = useToastStore()

const ipToBlock = ref('')
const ipToUnblock = ref('')
const isBlocking = ref(false)
const isUnblocking = ref(false)

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

const handleBlockIP = async () => {
  const ip = ipToBlock.value.trim()
  if (!ip) {
    toast.warning('Please enter an IP address or CIDR subnet to block.')
    return
  }

  isBlocking.value = true
  try {
    toast.info(`ENFORCING BLOCK RULE FOR ${ip}...`)
    const res = await firewallStore.blockIP(ip)
    toast.success(res.message || `IP ${ip} BLOCKED SUCCESSFULLY`)
    ipToBlock.value = ''
  } catch (err) {
    console.error('Block IP error:', err)
    const errMsg = err.response?.data?.error || err.response?.data?.message || 'Failed to block IP'
    toast.error(`BLOCK RULE FAILED: ${errMsg}`)
  } finally {
    isBlocking.value = false
  }
}

const handleUnblockIP = async () => {
  const ip = ipToUnblock.value.trim()
  if (!ip) {
    toast.warning('Please enter an IP address or CIDR subnet to unblock.')
    return
  }

  isUnblocking.value = true
  try {
    toast.info(`REMOVING BLOCK RULE FOR ${ip}...`)
    const res = await firewallStore.unblockIP(ip)
    toast.success(res.message || `IP ${ip} UNBLOCKED SUCCESSFULLY`)
    ipToUnblock.value = ''
  } catch (err) {
    console.error('Unblock IP error:', err)
    const errMsg = err.response?.data?.error || err.response?.data?.message || 'Failed to unblock IP'
    toast.error(`UNBLOCK RULE FAILED: ${errMsg}`)
  } finally {
    isUnblocking.value = false
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

    <!-- IP Blocking & Rule Management Grid -->
    <div class="firewall-grid">
      <!-- Block IP Card -->
      <div class="tron-card ip-action-card block-card">
        <div class="card-header">
          <Ban :size="20" class="glow-orange" />
          <h3>IP DEFENSE: BLOCK TRAFFIC</h3>
        </div>
        <div class="card-body">
          <p class="card-desc">
            Immediately deny all incoming packets from an IPv4 / IPv6 address or CIDR subnet block.
          </p>
          <form @submit.prevent="handleBlockIP" class="ip-form">
            <div class="input-wrap">
              <Globe :size="16" class="input-icon" />
              <input 
                v-model="ipToBlock" 
                type="text" 
                placeholder="e.g. 192.168.1.100 or 10.0.0.0/24" 
                class="tron-input font-data"
                :disabled="isBlocking"
              />
            </div>
            <button type="submit" class="action-btn btn-block" :disabled="isBlocking || !ipToBlock.trim()">
              <Ban :size="16" />
              <span>{{ isBlocking ? 'BLOCKING...' : 'ENFORCE BLOCK' }}</span>
            </button>
          </form>
        </div>
      </div>

      <!-- Unblock IP Card -->
      <div class="tron-card ip-action-card unblock-card">
        <div class="card-header">
          <Unlock :size="20" class="glow-cyan" />
          <h3>IP DEFENSE: UNBLOCK TRAFFIC</h3>
        </div>
        <div class="card-body">
          <p class="card-desc">
            Remove an active block rule to restore incoming packet flow for a specific IP or subnet.
          </p>
          <form @submit.prevent="handleUnblockIP" class="ip-form">
            <div class="input-wrap">
              <Globe :size="16" class="input-icon" />
              <input 
                v-model="ipToUnblock" 
                type="text" 
                placeholder="e.g. 192.168.1.100 or 10.0.0.0/24" 
                class="tron-input font-data"
                :disabled="isUnblocking"
              />
            </div>
            <button type="submit" class="action-btn btn-unblock" :disabled="isUnblocking || !ipToUnblock.trim()">
              <Unlock :size="16" />
              <span>{{ isUnblocking ? 'UNBLOCKING...' : 'LIFT BLOCK' }}</span>
            </button>
          </form>
        </div>
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
  font-size: 1.3rem;
  margin-bottom: 0.35rem;
}

.status-card.active { border-inline-start: 4px solid var(--linux-green); }
.status-card.inactive { border-inline-start: 4px solid var(--rdr-crimson); }

.status-card.active .status-text { color: var(--linux-green); }
.status-card.inactive .status-text { color: var(--rdr-crimson); }

.status-info p {
  color: var(--text-muted);
  font-size: 0.88rem;
}

.toggle-btn {
  margin-inline-start: auto;
  background: rgba(220, 38, 38, 0.08);
  border: 1px solid var(--border-crimson);
  color: #fca5a5;
  padding: 1rem 2rem;
  border-radius: 4px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.4rem;
  font-family: var(--font-header);
  font-weight: 700;
  font-size: 0.85rem;
  letter-spacing: 1px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.toggle-btn.on {
  background: rgba(16, 185, 129, 0.08);
  border-color: rgba(16, 185, 129, 0.35);
  color: var(--linux-green);
}

.toggle-btn:hover {
  filter: brightness(1.15);
}

/* IP Actions Grid */
.firewall-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 1.5rem;
}

.ip-action-card {
  display: flex;
  flex-direction: column;
}

.ip-action-card .card-body {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.card-desc {
  color: var(--text-muted);
  font-size: 0.88rem;
  line-height: 1.45;
}

.ip-form {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.input-wrap {
  position: relative;
  flex: 1;
  min-width: 200px;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 0.85rem;
  color: var(--text-muted);
  pointer-events: none;
}

.tron-input {
  width: 100%;
  background: rgba(13, 15, 20, 0.95);
  border: 1px solid var(--border-subtle);
  color: var(--text-primary);
  padding: 0.75rem 1rem 0.75rem 2.5rem;
  border-radius: 4px;
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.tron-input:focus {
  outline: none;
  border-color: var(--neon-cyan);
  box-shadow: 0 0 10px rgba(0, 240, 255, 0.2);
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  border-radius: 4px;
  font-family: var(--font-header);
  font-size: 0.82rem;
  font-weight: 700;
  letter-spacing: 0.5px;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-block {
  background: rgba(220, 38, 38, 0.15);
  border: 1px solid var(--border-crimson);
  color: #fca5a5;
}

.btn-block:not(:disabled):hover {
  background: rgba(220, 38, 38, 0.3);
  box-shadow: 0 0 12px rgba(220, 38, 38, 0.4);
}

.btn-unblock {
  background: rgba(0, 240, 255, 0.12);
  border: 1px solid rgba(0, 240, 255, 0.4);
  color: var(--neon-cyan);
}

.btn-unblock:not(:disabled):hover {
  background: rgba(0, 240, 255, 0.25);
  box-shadow: 0 0 12px rgba(0, 240, 255, 0.4);
}

/* Rules Display */
.rules-card {
  display: flex;
  flex-direction: column;
  min-height: 300px;
}

.card-header {
  padding: 1.25rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  border-bottom: 1px solid var(--border-subtle);
}

.rules-content {
  padding: 1.25rem;
  flex: 1;
  background: #0d0f14;
}

.raw-output {
  white-space: pre-wrap;
  word-wrap: break-word;
  color: var(--text-primary);
  font-size: 0.86rem;
  line-height: 1.5;
  background: #11131a;
  padding: 1.25rem;
  border-radius: 4px;
  border: 1px solid var(--border-subtle);
  font-family: var(--font-data);
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

