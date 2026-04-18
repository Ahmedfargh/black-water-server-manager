<script setup>
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { 
  Globe, 
  Activity, 
  Plus, 
  RefreshCw, 
  ExternalLink,
  ShieldCheck,
  ShieldX,
  Clock,
  LayoutGrid,
  Settings
} from 'lucide-vue-next'
import { useSiteStore } from '../stores/sites'
import { useToastStore } from '../stores/toast'

const { t } = useI18n()
const siteStore = useSiteStore()
const toast = useToastStore()
const showAddModal = ref(false)
const isEditing = ref(false)
const editingSiteId = ref(null)
const newSiteName = ref('')
const newSiteUrl = ref('')
const newHealthRoute = ref('')
const newDescription = ref('')
const newMethod = ref('GET')
const newExpectedStatus = ref(200)
const isSubmitting = ref(false)

onMounted(() => {
  siteStore.fetchSites()
})

const handleAddSite = async () => {
  if (!newSiteName.value || !newSiteUrl.value) return
  isSubmitting.value = true
  
  const payload = {
    name: newSiteName.value,
    url: newSiteUrl.value,
    health_route: newHealthRoute.value || newSiteUrl.value,
    description: newDescription.value,
    method: newMethod.value,
    expected_status: parseInt(newExpectedStatus.value)
  }

  try {
    if (isEditing.value) {
      await siteStore.updateSite(editingSiteId.value, payload)
      toast.success(t('sites.node_updated', { name: payload.name }))
    } else {
      await siteStore.addSite(payload)
      toast.success(t('sites.node_established', { name: payload.name }))
    }
    showAddModal.value = false
    resetForm()
  } catch (err) {
    toast.error(`${t('sites.uplink_failed')}: ${err.response?.data?.error || t('common.error_processing')}`)
  } finally {
    isSubmitting.value = false
  }
}

const openEditModal = (site) => {
  isEditing.value = true
  editingSiteId.value = site.id || site.ID
  newSiteName.value = site.name
  newSiteUrl.value = site.url
  newHealthRoute.value = site.health_route
  newDescription.value = site.description
  newMethod.value = site.method
  newExpectedStatus.value = site.expected_status
  showAddModal.value = true
}

const resetForm = () => {
  isEditing.value = false
  editingSiteId.value = null
  newSiteName.value = ''
  newSiteUrl.value = ''
  newHealthRoute.value = ''
  newDescription.value = ''
  newMethod.value = 'GET'
  newExpectedStatus.value = 200
}

const triggerCheckup = async () => {
  try {
    await siteStore.triggerCheckup()
    toast.success(t('sites.sync_complete'))
  } catch (err) {
    toast.error(t('sites.sync_failed'))
  }
}

const getStatusColor = (status) => {
  const s = status?.toUpperCase()
  if (s === 'UP') return 'var(--neon-cyan)'
  if (s === 'DOWN') return 'var(--neon-orange)'
  return 'var(--text-secondary)'
}
</script>

<template>
  <div class="sites-view">
    <div class="header-row">
      <h2 class="glow-cyan">{{ $t('sites.node_monitor') }}</h2>
      <div class="actions">
        <button @click="triggerCheckup" class="tron-btn secondary">
          <Activity :size="18" />
          {{ $t('sites.full_sync') }}
        </button>
        <button @click="resetForm(); showAddModal = true" class="tron-btn">
          <Plus :size="18" />
          {{ $t('sites.add_node') }}
        </button>
      </div>
    </div>

    <!-- Site Grid -->
    <div class="site-grid">
      <div 
        v-for="site in siteStore.sites" 
        :key="site.id" 
        class="tron-card site-card"
        :class="{ 'up': site.status?.toUpperCase() === 'UP', 'down': site.status?.toUpperCase() === 'DOWN' }"
      >
        <div class="card-header">
           <div class="icon-wrap">
             <Globe :size="24" :style="{ color: getStatusColor(site.status) }" />
           </div>
           <div class="node-info">
             <div class="title-row">
               <h3>{{ site.name || ($t('sites.external_node') || 'EXTERNAL NODE') }}</h3>
               <button @click="openEditModal(site)" class="edit-btn" :title="$t('sites.edit_node')">
                 <Settings :size="16" />
               </button>
             </div>
             <a :href="site.url" target="_blank" class="url-link font-data">
               {{ site.url }}
               <ExternalLink :size="12" />
             </a>
           </div>
        </div>

        <div class="card-body">
           <div class="status-indicator">
              <span class="status-label">{{ $t('sites.current_status') }}:</span>
              <span class="status-val" :style="{ color: getStatusColor(site.status) }">
                {{ site.status?.toUpperCase() || $t('common.unknown') }}
              </span>
           </div>
           
           <div class="metrics-row">
              <div class="metric">
                <Clock :size="14" />
                <span>{{ $t('sites.last_scan') }}: {{ site.last_checked ? new Date(site.last_checked).toLocaleTimeString() : '--:--' }}</span>
              </div>
           </div>
        </div>

        <div class="card-footer">
           <div class="activity-strip">
              <!-- Placeholder for recent history dots -->
              <div v-for="i in 10" :key="i" class="dot" :class="{ 'active': site.status?.toUpperCase() === 'UP' }"></div>
           </div>
        </div>
      </div>

      <div v-if="siteStore.sites.length === 0" class="empty-state">
         <Globe :size="48" class="pulse" />
         <p>{{ $t('sites.no_nodes') }}</p>
         <button @click="resetForm(); showAddModal = true" class="tron-btn">{{ $t('sites.initiate_uplink') }}</button>
      </div>
    </div>

    <!-- Add Site Modal -->
    <transition name="modal">
      <div v-if="showAddModal" class="modal-overlay">
        <div class="tron-card modal-container enhanced-modal">
          <div class="modal-header">
            <h3>{{ isEditing ? $t('sites.modify_config') : $t('sites.init_node') }}</h3>
          </div>
          <form @submit.prevent="handleAddSite" class="add-form">
            <div class="grid-inputs">
              <div class="input-group">
                <label>{{ $t('sites.designation') }}</label>
                <input v-model="newSiteName" type="text" :placeholder="$t('sites.name_placeholder') || 'e.g. PRIMARY GRID'" required />
              </div>
              <div class="input-group">
                <label>{{ $t('sites.coordinates') }}</label>
                <input v-model="newSiteUrl" type="url" placeholder="https://example.com" required />
              </div>
            </div>

            <div class="input-group">
              <label>{{ $t('sites.health_route') }}</label>
              <input v-model="newHealthRoute" type="text" :placeholder="$t('sites.health_placeholder') || 'e.g. https://example.com/health (defaults to URL)'" />
            </div>

            <div class="grid-inputs">
              <div class="input-group">
                <label>{{ $t('sites.method') }}</label>
                <select v-model="newMethod">
                  <option value="GET">GET</option>
                  <option value="HEAD">HEAD</option>
                  <option value="POST">POST</option>
                </select>
              </div>
              <div class="input-group">
                <label>{{ $t('sites.expected_status') }}</label>
                <input v-model="newExpectedStatus" type="number" placeholder="200" />
              </div>
            </div>

            <div class="input-group">
              <label>{{ $t('common.description') }}</label>
              <textarea v-model="newDescription" rows="2" :placeholder="$t('sites.desc_placeholder') || 'Describe this node...'"></textarea>
            </div>

            <div class="modal-footer">
              <button type="button" @click="showAddModal = false; resetForm()" class="tron-btn ghost">{{ $t('common.abort') }}</button>
              <button type="submit" :disabled="isSubmitting" class="tron-btn">
                <span v-if="isSubmitting">{{ isEditing ? $t('common.updating') : $t('common.connecting') }}</span>
                <span v-else>{{ isEditing ? $t('sites.save_config') : $t('sites.establish_uplink') }}</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.sites-view {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.actions {
  display: flex;
  gap: 1rem;
}

.site-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 2rem;
}

@media (max-width: 768px) {
  .site-grid { gap: 1.5rem; }
}

.site-card {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  border-top: 3px solid transparent;
}

@media (max-width: 480px) {
  .site-card { padding: 1rem; }
}

.site-card.up { border-top-color: var(--neon-cyan); }
.site-card.down { border-top-color: var(--neon-orange); }

.card-header {
  display: flex;
  align-items: center;
  gap: 1.2rem;
}

.icon-wrap {
  background: rgba(255, 255, 255, 0.03);
  padding: 0.8rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.node-info h3 {
  font-size: 1.1rem;
  margin-bottom: 0.2rem;
}

.title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.edit-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  opacity: 0.6;
}

.edit-btn:hover {
  color: var(--neon-cyan);
  opacity: 1;
  text-shadow: 0 0 8px var(--neon-cyan-glow);
}

.url-link {
  font-size: 0.75rem;
  color: var(--text-secondary);
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.url-link:hover { color: var(--neon-cyan); }

.card-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.status-indicator {
  display: flex;
  padding: 0.8rem;
  background: rgba(255, 255, 255, 0.02);
  justify-content: space-between;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.status-label { font-size: 0.8rem; color: var(--text-secondary); }
.status-val { font-weight: 700; letter-spacing: 2px; font-size: 0.9rem; }

.metrics-row {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.metric { display: flex; align-items: center; gap: 0.5rem; }

.activity-strip {
  display: flex;
  gap: 0.4rem;
  padding-top: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.dot {
  width: 8px;
  height: 8px;
  background: rgba(255, 255, 255, 0.1);
}

.dot.active {
  background: var(--neon-cyan);
  box-shadow: 0 0 5px var(--neon-cyan-glow);
}

.empty-state {
  grid-column: 1 / -1;
  padding: 5rem;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
  color: var(--text-secondary);
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-container {
  width: 100%;
  max-width: 450px;
  padding: 2rem;
}

.add-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.input-group label {
  display: block;
  font-size: 0.75rem;
  margin-bottom: 0.5rem;
  color: var(--text-secondary);
  letter-spacing: 2px;
}

.input-group input {
  width: 100%;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(0, 242, 255, 0.2);
  padding: 0.8rem;
  color: var(--text-primary);
  font-family: var(--font-header);
  outline: none;
}

.input-group input:focus { border-color: var(--neon-cyan); box-shadow: 0 0 10px var(--neon-cyan-glow); }

.modal-container.enhanced-modal {
  width: 95%;
  max-width: 550px;
  margin: 1rem;
}

.grid-inputs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

@media (max-width: 600px) {
  .grid-inputs {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
}

.input-group select,
.input-group textarea {
  width: 100%;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(0, 242, 255, 0.2);
  padding: 0.8rem;
  color: var(--text-primary);
  font-family: var(--font-header);
  outline: none;
}

.input-group select:focus,
.input-group textarea:focus { 
  border-color: var(--neon-cyan); 
  box-shadow: 0 0 10px var(--neon-cyan-glow); 
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1rem;
}
</style>
