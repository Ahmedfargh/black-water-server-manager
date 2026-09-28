<script setup>
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Key,
  Plus,
  Download,
  Upload,
  RefreshCw,
  Search,
  CheckCircle2,
  XCircle,
  ShieldCheck,
  ShieldAlert,
  Copy,
  Check,
  Trash2,
  AlertTriangle,
  Lock,
  Cpu,
  FileCode,
  Sparkles,
  X
} from 'lucide-vue-next'
import { useSSHStore } from '../stores/ssh'
import { useToastStore } from '../stores/toast'

const { t } = useI18n()
const sshStore = useSSHStore()
const toast = useToastStore()

// State
const searchQuery = ref('')
const filterType = ref('all') // 'all' | 'ed25519' | 'rsa' | 'authorized'
const copiedField = ref('')

// Modals
const showGenerateModal = ref(false)
const showImportModal = ref(false)
const showPrivateKeyModal = ref(false)
const showDeleteModal = ref(false)
const keyToDelete = ref(null)

// Generated Key Result
const generatedKeyResult = ref({
  key: null,
  private_key_pem: ''
})

// Generate Form
const genForm = ref({
  name: '',
  key_type: 'ed25519',
  comment: '',
  add_to_authorized_keys: true,
  rsa_bits: 4096
})

// Import Form
const importForm = ref({
  name: '',
  public_key: '',
  comment: '',
  add_to_authorized_keys: true
})

onMounted(async () => {
  await loadKeys()
})

const loadKeys = async () => {
  try {
    await sshStore.fetchKeys(1, 100)
  } catch (err) {
    toast.error(t('ssh.fetch_failed', 'Failed to load SSH keys'))
  }
}

// Filtered keys
const filteredKeys = computed(() => {
  let list = sshStore.keys || []
  
  if (filterType.value === 'ed25519') {
    list = list.filter(k => k.key_type?.toLowerCase().includes('ed25519'))
  } else if (filterType.value === 'rsa') {
    list = list.filter(k => k.key_type?.toLowerCase().includes('rsa'))
  } else if (filterType.value === 'authorized') {
    list = list.filter(k => k.added_to_authorized_keys)
  }

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim()
    list = list.filter(k => 
      k.name?.toLowerCase().includes(q) ||
      k.fingerprint?.toLowerCase().includes(q) ||
      k.comment?.toLowerCase().includes(q) ||
      k.key_type?.toLowerCase().includes(q)
    )
  }

  return list
})

const handleCopy = async (text, fieldId) => {
  try {
    await navigator.clipboard.writeText(text)
    copiedField.value = fieldId
    toast.success(t('ssh.copied_to_clipboard', 'Copied to clipboard!'))
    setTimeout(() => {
      if (copiedField.value === fieldId) {
        copiedField.value = ''
      }
    }, 2000)
  } catch (err) {
    toast.error(t('ssh.copy_failed', 'Failed to copy'))
  }
}

const handleToggleAuth = async (key) => {
  try {
    const res = await sshStore.toggleAuth(key.id)
    if (res.added_to_authorized_keys) {
      toast.success(t('ssh.key_authorized', 'Key added to host authorized_keys'))
    } else {
      toast.info(t('ssh.key_unauthorized', 'Key removed from host authorized_keys'))
    }
  } catch (err) {
    toast.error(t('ssh.toggle_failed', 'Failed to update key authorization'))
  }
}

const openGenerateModal = () => {
  genForm.value = {
    name: '',
    key_type: 'ed25519',
    comment: '',
    add_to_authorized_keys: true,
    rsa_bits: 4096
  }
  showGenerateModal.value = true
}

const submitGenerate = async () => {
  if (!genForm.value.name.trim()) {
    toast.error(t('ssh.name_required', 'Key name is required'))
    return
  }

  try {
    const res = await sshStore.generateKey(genForm.value)
    showGenerateModal.value = false
    generatedKeyResult.value = {
      key: res.key,
      private_key_pem: res.private_key_pem
    }
    showPrivateKeyModal.value = true
    toast.success(t('ssh.key_generated_success', 'SSH Key pair generated successfully!'))
  } catch (err) {
    toast.error(err.response?.data?.error || t('ssh.gen_failed', 'Key generation failed'))
  }
}

const openImportModal = () => {
  importForm.value = {
    name: '',
    public_key: '',
    comment: '',
    add_to_authorized_keys: true
  }
  showImportModal.value = true
}

const submitImport = async () => {
  if (!importForm.value.name.trim()) {
    toast.error(t('ssh.name_required', 'Key name is required'))
    return
  }
  if (!importForm.value.public_key.trim()) {
    toast.error(t('ssh.pubkey_required', 'Public key string is required'))
    return
  }

  try {
    await sshStore.importKey(importForm.value)
    showImportModal.value = false
    toast.success(t('ssh.key_imported_success', 'Public key imported successfully!'))
  } catch (err) {
    toast.error(err.response?.data?.error || t('ssh.import_failed', 'Key import failed'))
  }
}

const downloadPrivateKey = () => {
  if (!generatedKeyResult.value.private_key_pem) return
  const blob = new Blob([generatedKeyResult.value.private_key_pem], { type: 'application/x-pem-file' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  const filename = `${generatedKeyResult.value.key?.name || 'id'}_${generatedKeyResult.value.key?.key_type || 'key'}.pem`
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  toast.success(t('ssh.downloaded_private_key', 'Private key downloaded'))
}

const confirmDelete = (key) => {
  keyToDelete.value = key
  showDeleteModal.value = true
}

const submitDelete = async () => {
  if (!keyToDelete.value) return
  try {
    await sshStore.deleteKey(keyToDelete.value.id)
    showDeleteModal.value = false
    keyToDelete.value = null
    toast.success(t('ssh.key_deleted_success', 'SSH Key deleted successfully'))
  } catch (err) {
    toast.error(err.response?.data?.error || t('ssh.delete_failed', 'Failed to delete key'))
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return isNaN(d.getTime()) ? dateStr : d.toLocaleString()
}
</script>

<template>
  <div class="ssh-view-wrapper">
    <!-- Header Section -->
    <div class="header-section">
      <div class="header-title-block">
        <div class="title-with-icon">
          <div class="icon-box">
            <Key :size="24" class="header-icon" />
          </div>
          <div>
            <h1 class="page-title">{{ $t('ssh.title', 'SSH Key Management') }}</h1>
            <p class="page-subtitle">{{ $t('ssh.subtitle', 'Generate, import, and manage host authorized SSH keys') }}</p>
          </div>
        </div>
      </div>

      <div class="header-actions">
        <button class="btn btn-secondary" @click="loadKeys" :disabled="sshStore.loading">
          <RefreshCw :size="16" :class="{ 'spin-anim': sshStore.loading }" />
          <span>{{ $t('common.refresh', 'Refresh') }}</span>
        </button>
        <button class="btn btn-secondary" @click="openImportModal">
          <Upload :size="16" />
          <span>{{ $t('ssh.import_key', 'Import Public Key') }}</span>
        </button>
        <button class="btn btn-primary glow-button" @click="openGenerateModal">
          <Plus :size="16" />
          <span>{{ $t('ssh.generate_key', 'Generate Key Pair') }}</span>
        </button>
      </div>
    </div>

    <!-- Telemetry & Metrics Strip -->
    <div class="metrics-grid">
      <div class="metric-card">
        <div class="metric-header">
          <span class="metric-label">{{ $t('ssh.total_keys', 'Total Managed Keys') }}</span>
          <Key :size="18" class="metric-icon text-crimson" />
        </div>
        <div class="metric-value font-data">{{ sshStore.keys.length }}</div>
        <div class="metric-meta">{{ $t('ssh.active_database_records', 'In Blackwater Keyring') }}</div>
      </div>

      <div class="metric-card">
        <div class="metric-header">
          <span class="metric-label">{{ $t('ssh.authorized_in_host', 'Host Authorized') }}</span>
          <ShieldCheck :size="18" class="metric-icon text-green" />
        </div>
        <div class="metric-value font-data text-green">{{ sshStore.authorizedKeysCount }}</div>
        <div class="metric-meta">{{ $t('ssh.in_authorized_keys', 'Active in ~/.ssh/authorized_keys') }}</div>
      </div>

      <div class="metric-card">
        <div class="metric-header">
          <span class="metric-label">{{ $t('ssh.ed25519_keys', 'ED25519 Keys') }}</span>
          <Cpu :size="18" class="metric-icon text-amber" />
        </div>
        <div class="metric-value font-data text-amber">{{ sshStore.ed25519Count }}</div>
        <div class="metric-meta">{{ $t('ssh.high_security_fast', 'Modern Edwards Curve') }}</div>
      </div>

      <div class="metric-card">
        <div class="metric-header">
          <span class="metric-label">{{ $t('ssh.rsa_keys', 'RSA Keys') }}</span>
          <FileCode :size="18" class="metric-icon text-muted" />
        </div>
        <div class="metric-value font-data">{{ sshStore.rsaCount }}</div>
        <div class="metric-meta">{{ $t('ssh.legacy_compatible', 'Standard OpenSSH RSA') }}</div>
      </div>
    </div>

    <!-- Controls & Filter Toolbar -->
    <div class="toolbar-card">
      <div class="search-box">
        <Search :size="16" class="search-icon" />
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="$t('ssh.search_placeholder', 'Search by name, fingerprint, comment, algorithm...')"
          class="search-input"
        />
        <button v-if="searchQuery" @click="searchQuery = ''" class="clear-search-btn">
          <X :size="14" />
        </button>
      </div>

      <div class="filter-pills">
        <button
          class="pill-btn"
          :class="{ active: filterType === 'all' }"
          @click="filterType = 'all'"
        >
          {{ $t('common.all', 'All') }} ({{ sshStore.keys.length }})
        </button>
        <button
          class="pill-btn"
          :class="{ active: filterType === 'authorized' }"
          @click="filterType = 'authorized'"
        >
          {{ $t('ssh.authorized', 'Authorized') }} ({{ sshStore.authorizedKeysCount }})
        </button>
        <button
          class="pill-btn"
          :class="{ active: filterType === 'ed25519' }"
          @click="filterType = 'ed25519'"
        >
          ED25519 ({{ sshStore.ed25519Count }})
        </button>
        <button
          class="pill-btn"
          :class="{ active: filterType === 'rsa' }"
          @click="filterType = 'rsa'"
        >
          RSA ({{ sshStore.rsaCount }})
        </button>
      </div>
    </div>

    <!-- Keys Table Container -->
    <div class="table-container-card">
      <div v-if="sshStore.loading && sshStore.keys.length === 0" class="empty-state-box">
        <RefreshCw :size="32" class="spin-anim text-crimson mb-3" />
        <p>{{ $t('ssh.loading_keys', 'Loading SSH keyring from server...') }}</p>
      </div>

      <div v-else-if="filteredKeys.length === 0" class="empty-state-box">
        <Key :size="40" class="text-muted mb-3" />
        <h3 class="empty-title">{{ $t('ssh.no_keys_found', 'No SSH Keys Found') }}</h3>
        <p class="empty-subtitle">{{ $t('ssh.no_keys_desc', 'Generate a new Edwards Curve key pair or import your existing public key.') }}</p>
        <div class="mt-4 flex gap-3">
          <button class="btn btn-primary" @click="openGenerateModal">
            <Plus :size="16" />
            <span>{{ $t('ssh.generate_key', 'Generate Key Pair') }}</span>
          </button>
          <button class="btn btn-secondary" @click="openImportModal">
            <Upload :size="16" />
            <span>{{ $t('ssh.import_key', 'Import Public Key') }}</span>
          </button>
        </div>
      </div>

      <div v-else class="table-responsive">
        <table class="tactical-table">
          <thead>
            <tr>
              <th class="th-name">{{ $t('ssh.th_name', 'NAME / LABEL') }}</th>
              <th class="th-type">{{ $t('ssh.th_type', 'TYPE') }}</th>
              <th class="th-fingerprint">{{ $t('ssh.th_fingerprint', 'SHA256 FINGERPRINT') }}</th>
              <th class="th-pubkey">{{ $t('ssh.th_public_key', 'PUBLIC KEY') }}</th>
              <th class="th-auth">{{ $t('ssh.th_host_access', 'HOST ACCESS') }}</th>
              <th class="th-date">{{ $t('ssh.th_created', 'CREATED') }}</th>
              <th class="th-actions text-right">{{ $t('common.actions', 'ACTIONS') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="key in filteredKeys" :key="key.id" class="table-row">
              <td class="td-name">
                <div class="key-name-block">
                  <span class="key-name">{{ key.name }}</span>
                  <span v-if="key.comment" class="key-comment font-data text-muted">{{ key.comment }}</span>
                </div>
              </td>
              <td class="td-type">
                <span class="type-badge font-data" :class="key.key_type?.toLowerCase()">
                  {{ key.key_type?.toUpperCase() }}
                </span>
              </td>
              <td class="td-fingerprint">
                <div class="fingerprint-box">
                  <span class="fingerprint-text font-data">{{ key.fingerprint }}</span>
                  <button
                    class="btn-icon-subtle"
                    @click="handleCopy(key.fingerprint, 'fp-' + key.id)"
                    :title="$t('ssh.copy_fingerprint', 'Copy Fingerprint')"
                  >
                    <Check v-if="copiedField === 'fp-' + key.id" :size="14" class="text-green" />
                    <Copy v-else :size="14" />
                  </button>
                </div>
              </td>
              <td class="td-pubkey">
                <div class="pubkey-preview-box">
                  <span class="pubkey-snippet font-data">{{ key.public_key }}</span>
                  <button
                    class="btn-icon-subtle"
                    @click="handleCopy(key.public_key, 'pk-' + key.id)"
                    :title="$t('ssh.copy_public_key', 'Copy Public Key')"
                  >
                    <Check v-if="copiedField === 'pk-' + key.id" :size="14" class="text-green" />
                    <Copy v-else :size="14" />
                  </button>
                </div>
              </td>
              <td class="td-auth">
                <button
                  class="auth-toggle-pill"
                  :class="{ active: key.added_to_authorized_keys }"
                  @click="handleToggleAuth(key)"
                  :title="key.added_to_authorized_keys ? $t('ssh.click_to_unauthorize', 'Click to disable from authorized_keys') : $t('ssh.click_to_authorize', 'Click to authorize in host')"
                >
                  <span class="auth-dot"></span>
                  <span class="auth-text font-data">
                    {{ key.added_to_authorized_keys ? $t('ssh.authorized', 'AUTHORIZED') : $t('ssh.disabled', 'DISABLED') }}
                  </span>
                </button>
              </td>
              <td class="td-date font-data text-muted">
                {{ formatDate(key.created_at) }}
              </td>
              <td class="td-actions text-right">
                <button
                  class="btn-icon-danger"
                  @click="confirmDelete(key)"
                  :title="$t('ssh.delete_key', 'Delete Key')"
                >
                  <Trash2 :size="16" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- MODAL: Generate Key Pair -->
    <div v-if="showGenerateModal" class="modal-backdrop" @click.self="showGenerateModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <div class="flex items-center gap-2">
            <Sparkles :size="20" class="text-crimson" />
            <h2 class="modal-title">{{ $t('ssh.modal_generate_title', 'Generate SSH Key Pair') }}</h2>
          </div>
          <button class="modal-close-btn" @click="showGenerateModal = false">
            <X :size="18" />
          </button>
        </div>

        <form @submit.prevent="submitGenerate" class="modal-body">
          <div class="form-group">
            <label class="form-label">{{ $t('ssh.key_name_label', 'Key Identifier / Name') }} *</label>
            <input
              v-model="genForm.name"
              type="text"
              placeholder="e.g. devops-macbook-pro"
              class="form-input"
              required
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="form-group">
              <label class="form-label">{{ $t('ssh.algorithm_type', 'Algorithm') }}</label>
              <select v-model="genForm.key_type" class="form-select">
                <option value="ed25519">ED25519 (Recommended)</option>
                <option value="rsa">RSA</option>
              </select>
            </div>

            <div class="form-group" v-if="genForm.key_type === 'rsa'">
              <label class="form-label">{{ $t('ssh.rsa_bits_label', 'RSA Key Size (Bits)') }}</label>
              <select v-model.number="genForm.rsa_bits" class="form-select">
                <option :value="2048">2048-bit</option>
                <option :value="4096">4096-bit (Recommended)</option>
              </select>
            </div>

            <div class="form-group" :class="{ 'col-span-2': genForm.key_type !== 'rsa' }">
              <label class="form-label">{{ $t('ssh.comment_label', 'Comment / Email') }}</label>
              <input
                v-model="genForm.comment"
                type="text"
                placeholder="e.g. admin@blackwater.internal"
                class="form-input"
              />
            </div>
          </div>

          <div class="checkbox-group">
            <label class="checkbox-label">
              <input
                type="checkbox"
                v-model="genForm.add_to_authorized_keys"
                class="form-checkbox"
              />
              <div>
                <span class="font-bold text-white">{{ $t('ssh.auto_authorize_label', 'Authorize for direct host login') }}</span>
                <p class="text-xs text-muted mt-0.5">{{ $t('ssh.auto_authorize_desc', 'Appends public key directly to ~/.ssh/authorized_keys') }}</p>
              </div>
            </label>
          </div>

          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showGenerateModal = false">
              {{ $t('common.cancel', 'Cancel') }}
            </button>
            <button type="submit" class="btn btn-primary glow-button" :disabled="sshStore.generating">
              <RefreshCw v-if="sshStore.generating" :size="16" class="spin-anim" />
              <Sparkles v-else :size="16" />
              <span>{{ sshStore.generating ? $t('ssh.generating', 'Generating...') : $t('ssh.generate_now', 'Generate Key Pair') }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL: Private Key Display & Download (ONE TIME) -->
    <div v-if="showPrivateKeyModal" class="modal-backdrop" @click.self="showPrivateKeyModal = false">
      <div class="modal-card modal-lg">
        <div class="modal-header border-crimson">
          <div class="flex items-center gap-2">
            <ShieldAlert :size="22" class="text-crimson" />
            <h2 class="modal-title">{{ $t('ssh.private_key_ready', 'Private Key Generated') }}</h2>
          </div>
          <button class="modal-close-btn" @click="showPrivateKeyModal = false">
            <X :size="18" />
          </button>
        </div>

        <div class="modal-body">
          <div class="alert-box-warning mb-4">
            <AlertTriangle :size="20" class="text-amber flex-shrink-0" />
            <div>
              <p class="font-bold text-amber">{{ $t('ssh.one_time_warning_title', 'IMPORTANT: Save Your Private Key Now') }}</p>
              <p class="text-xs text-secondary mt-1">
                {{ $t('ssh.one_time_warning_desc', 'Blackwater does NOT store your private key in the database for security reasons. If you lose this key, you will have to generate a new one.') }}
              </p>
            </div>
          </div>

          <div class="key-summary-strip mb-4">
            <div>
              <span class="text-xs text-muted block">{{ $t('ssh.th_name', 'Name') }}</span>
              <span class="font-bold text-white">{{ generatedKeyResult.key?.name }}</span>
            </div>
            <div>
              <span class="text-xs text-muted block">{{ $t('ssh.th_type', 'Type') }}</span>
              <span class="type-badge font-data">{{ generatedKeyResult.key?.key_type?.toUpperCase() }}</span>
            </div>
            <div>
              <span class="text-xs text-muted block">{{ $t('ssh.th_fingerprint', 'Fingerprint') }}</span>
              <span class="font-data text-xs text-secondary">{{ generatedKeyResult.key?.fingerprint }}</span>
            </div>
          </div>

          <div class="code-block-container mb-4">
            <div class="code-block-header">
              <span class="font-data text-xs text-muted">PRIVATE KEY (PEM FORMAT)</span>
              <div class="flex gap-2">
                <button class="btn-xs btn-secondary" @click="handleCopy(generatedKeyResult.private_key_pem, 'priv-key')">
                  <Check v-if="copiedField === 'priv-key'" :size="12" class="text-green" />
                  <Copy v-else :size="12" />
                  <span>{{ copiedField === 'priv-key' ? $t('common.copied', 'Copied') : $t('common.copy', 'Copy') }}</span>
                </button>
                <button class="btn-xs btn-primary" @click="downloadPrivateKey">
                  <Download :size="12" />
                  <span>{{ $t('ssh.download_pem', 'Download .PEM') }}</span>
                </button>
              </div>
            </div>
            <pre class="code-pre font-data">{{ generatedKeyResult.private_key_pem }}</pre>
          </div>

          <div class="code-block-container">
            <div class="code-block-header">
              <span class="font-data text-xs text-muted">PUBLIC KEY (OPENSSH FORMAT)</span>
              <button class="btn-xs btn-secondary" @click="handleCopy(generatedKeyResult.key?.public_key, 'gen-pub-key')">
                <Check v-if="copiedField === 'gen-pub-key'" :size="12" class="text-green" />
                <Copy v-else :size="12" />
                <span>{{ copiedField === 'gen-pub-key' ? $t('common.copied', 'Copied') : $t('common.copy', 'Copy') }}</span>
              </button>
            </div>
            <pre class="code-pre font-data text-xs">{{ generatedKeyResult.key?.public_key }}</pre>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn btn-primary" @click="showPrivateKeyModal = false">
            {{ $t('ssh.done_saved', 'I Have Saved My Private Key') }}
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL: Import Public Key -->
    <div v-if="showImportModal" class="modal-backdrop" @click.self="showImportModal = false">
      <div class="modal-card modal-md">
        <div class="modal-header">
          <div class="flex items-center gap-2">
            <Upload :size="20" class="text-crimson" />
            <h2 class="modal-title">{{ $t('ssh.modal_import_title', 'Import Existing Public Key') }}</h2>
          </div>
          <button class="modal-close-btn" @click="showImportModal = false">
            <X :size="18" />
          </button>
        </div>

        <form @submit.prevent="submitImport" class="modal-body">
          <div class="form-group">
            <label class="form-label">{{ $t('ssh.key_name_label', 'Key Identifier / Name') }} *</label>
            <input
              v-model="importForm.name"
              type="text"
              placeholder="e.g. workstation-rsa"
              class="form-input"
              required
            />
          </div>

          <div class="form-group">
            <label class="form-label">{{ $t('ssh.public_key_string', 'OpenSSH Public Key String') }} *</label>
            <textarea
              v-model="importForm.public_key"
              rows="4"
              placeholder="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA... or ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ..."
              class="form-textarea font-data text-xs"
              required
            ></textarea>
          </div>

          <div class="form-group">
            <label class="form-label">{{ $t('ssh.comment_label', 'Comment / Description (Optional)') }}</label>
            <input
              v-model="importForm.comment"
              type="text"
              placeholder="e.g. Imported from CI/CD pipeline"
              class="form-input"
            />
          </div>

          <div class="checkbox-group">
            <label class="checkbox-label">
              <input
                type="checkbox"
                v-model="importForm.add_to_authorized_keys"
                class="form-checkbox"
              />
              <div>
                <span class="font-bold text-white">{{ $t('ssh.auto_authorize_label', 'Authorize for direct host login') }}</span>
                <p class="text-xs text-muted mt-0.5">{{ $t('ssh.auto_authorize_desc', 'Appends public key directly to ~/.ssh/authorized_keys') }}</p>
              </div>
            </label>
          </div>

          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showImportModal = false">
              {{ $t('common.cancel', 'Cancel') }}
            </button>
            <button type="submit" class="btn btn-primary" :disabled="sshStore.importing">
              <RefreshCw v-if="sshStore.importing" :size="16" class="spin-anim" />
              <Upload v-else :size="16" />
              <span>{{ sshStore.importing ? $t('ssh.importing', 'Importing...') : $t('ssh.import_btn', 'Import Public Key') }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL: Delete Confirmation -->
    <div v-if="showDeleteModal" class="modal-backdrop" @click.self="showDeleteModal = false">
      <div class="modal-card modal-sm">
        <div class="modal-header border-crimson">
          <div class="flex items-center gap-2">
            <Trash2 :size="20" class="text-crimson" />
            <h2 class="modal-title">{{ $t('ssh.delete_modal_title', 'Delete SSH Key') }}</h2>
          </div>
          <button class="modal-close-btn" @click="showDeleteModal = false">
            <X :size="18" />
          </button>
        </div>

        <div class="modal-body">
          <p class="text-secondary text-sm">
            {{ $t('ssh.delete_modal_confirm', 'Are you sure you want to delete this SSH key?') }}
          </p>
          <div class="mt-3 p-3 bg-black/40 border border-subtle rounded">
            <div class="font-bold text-white">{{ keyToDelete?.name }}</div>
            <div class="font-data text-xs text-muted mt-1">{{ keyToDelete?.fingerprint }}</div>
          </div>
          <p class="text-xs text-crimson mt-3">
            {{ $t('ssh.delete_modal_subwarning', 'This key will also be removed from the host\'s ~/.ssh/authorized_keys file immediately.') }}
          </p>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDeleteModal = false">
            {{ $t('common.cancel', 'Cancel') }}
          </button>
          <button class="btn btn-danger" @click="submitDelete">
            <Trash2 :size="16" />
            <span>{{ $t('common.delete', 'Delete Key') }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ssh-view-wrapper {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  max-width: 1600px;
  margin: 0 auto;
  width: 100%;
}

/* Header */
.header-section {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}

.header-title-block {
  display: flex;
  align-items: center;
}

.title-with-icon {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.icon-box {
  width: 44px;
  height: 44px;
  border-radius: 6px;
  background: var(--rdr-crimson-dim);
  border: 1px solid var(--border-crimson);
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-icon {
  color: var(--rdr-crimson);
}

.page-title {
  font-family: var(--font-header);
  font-size: 1.5rem;
  font-weight: 700;
  color: #fff;
  letter-spacing: 0.5px;
}

.page-subtitle {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

/* Metrics Strip */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.metric-card {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  padding: 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  transition: transform 0.2s, border-color 0.2s;
}

.metric-card:hover {
  border-color: rgba(255, 255, 255, 0.15);
  transform: translateY(-2px);
}

.metric-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.metric-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-muted);
  letter-spacing: 0.5px;
  text-transform: uppercase;
}

.metric-value {
  font-size: 1.8rem;
  font-weight: 700;
  color: #fff;
  line-height: 1.1;
}

.metric-meta {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

/* Toolbar */
.toolbar-card {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  padding: 0.75rem 1rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}

.search-box {
  position: relative;
  flex: 1;
  min-width: 260px;
}

.search-icon {
  position: absolute;
  left: 0.85rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
}

[dir="rtl"] .search-icon {
  left: auto;
  right: 0.85rem;
}

.search-input {
  width: 100%;
  background: #0f1117;
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
  padding: 0.55rem 2rem 0.55rem 2.4rem;
  color: #fff;
  font-size: 0.85rem;
  outline: none;
  transition: border-color 0.2s;
}

[dir="rtl"] .search-input {
  padding: 0.55rem 2.4rem 0.55rem 2rem;
}

.search-input:focus {
  border-color: var(--rdr-crimson);
}

.clear-search-btn {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
}

[dir="rtl"] .clear-search-btn {
  right: auto;
  left: 0.75rem;
}

.filter-pills {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.pill-btn {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border-subtle);
  color: var(--text-secondary);
  padding: 0.4rem 0.8rem;
  border-radius: 4px;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
}

.pill-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}

.pill-btn.active {
  background: var(--rdr-crimson-dim);
  border-color: var(--border-crimson);
  color: #fff;
}

/* Table */
.table-container-card {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  overflow: hidden;
}

.table-responsive {
  width: 100%;
  overflow-x: auto;
}

.tactical-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

[dir="rtl"] .tactical-table {
  text-align: right;
}

.tactical-table th {
  background: #11131a;
  padding: 0.85rem 1rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.5px;
  color: var(--text-muted);
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}

.tactical-table td {
  padding: 0.9rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  font-size: 0.85rem;
  vertical-align: middle;
}

.table-row:hover {
  background: rgba(255, 255, 255, 0.02);
}

.key-name-block {
  display: flex;
  flex-direction: column;
}

.key-name {
  font-weight: 600;
  color: #fff;
}

.key-comment {
  font-size: 0.72rem;
}

.type-badge {
  display: inline-block;
  font-size: 0.72rem;
  font-weight: 700;
  padding: 0.2rem 0.5rem;
  border-radius: 3px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid var(--border-subtle);
  color: var(--text-secondary);
}

.type-badge.ed25519 {
  background: var(--rdr-amber-dim);
  border-color: var(--border-crimson);
  color: var(--rdr-amber-hover);
}

.type-badge.rsa {
  background: rgba(59, 130, 246, 0.12);
  border-color: rgba(59, 130, 246, 0.3);
  color: #93c5fd;
}

.fingerprint-box, .pubkey-preview-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.fingerprint-text {
  font-size: 0.75rem;
  color: var(--text-secondary);
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pubkey-snippet {
  font-size: 0.72rem;
  color: var(--text-muted);
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Auth Pill Toggle */
.auth-toggle-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.35rem 0.7rem;
  border-radius: 9999px;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #fca5a5;
  cursor: pointer;
  font-size: 0.72rem;
  font-weight: 700;
  transition: all 0.2s ease;
}

.auth-toggle-pill .auth-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #ef4444;
}

.auth-toggle-pill.active {
  background: rgba(16, 185, 129, 0.15);
  border-color: rgba(16, 185, 129, 0.35);
  color: #6ee7b7;
}

.auth-toggle-pill.active .auth-dot {
  background: var(--linux-green);
  box-shadow: 0 0 8px var(--linux-green);
}

.auth-toggle-pill:hover {
  transform: scale(1.03);
}

/* Action buttons */
.btn-icon-subtle {
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-muted);
  padding: 0.25rem;
  border-radius: 3px;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-icon-subtle:hover {
  color: #fff;
  border-color: var(--border-subtle);
  background: rgba(255, 255, 255, 0.05);
}

.btn-icon-danger {
  background: rgba(220, 38, 38, 0.08);
  border: 1px solid rgba(220, 38, 38, 0.25);
  color: #fca5a5;
  padding: 0.35rem 0.5rem;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-icon-danger:hover {
  background: var(--rdr-crimson);
  color: #fff;
}

/* Empty State */
.empty-state-box {
  padding: 4rem 2rem;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.empty-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: #fff;
}

.empty-subtitle {
  font-size: 0.85rem;
  color: var(--text-secondary);
  max-width: 450px;
  margin-top: 0.4rem;
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.55rem 1rem;
  border-radius: 4px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.18s;
  border: 1px solid transparent;
  text-decoration: none;
}

.btn-primary {
  background: var(--rdr-crimson);
  color: #fff;
  border-color: var(--rdr-crimson-hover);
}

.btn-primary:hover:not(:disabled) {
  background: var(--rdr-crimson-hover);
  box-shadow: 0 0 12px var(--rdr-crimson-glow);
}

.btn-secondary {
  background: #191c25;
  color: var(--text-primary);
  border-color: var(--border-subtle);
}

.btn-secondary:hover:not(:disabled) {
  background: #222632;
  border-color: rgba(255, 255, 255, 0.2);
}

.btn-danger {
  background: #b91c1c;
  color: #fff;
}

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
}

.btn-xs {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.25rem 0.55rem;
  border-radius: 3px;
  font-size: 0.72rem;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid var(--border-subtle);
}

/* Modals */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(5, 7, 12, 0.85);
  backdrop-filter: blur(8px);
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
}

.modal-card {
  background: #141720;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  width: 100%;
  max-width: 550px;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.7);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-md {
  max-width: 650px;
}

.modal-lg {
  max-width: 780px;
}

.modal-sm {
  max-width: 440px;
}

.modal-header {
  padding: 1.2rem 1.4rem;
  border-bottom: 1px solid var(--border-subtle);
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #171b26;
}

.modal-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: #fff;
}

.modal-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
}

.modal-close-btn:hover {
  color: #fff;
}

.modal-body {
  padding: 1.4rem;
  overflow-y: auto;
  max-height: calc(85vh - 120px);
}

.modal-footer {
  padding: 1rem 1.4rem;
  border-top: 1px solid var(--border-subtle);
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
  background: #11131b;
}

/* Form inputs */
.form-group {
  margin-bottom: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-label {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-secondary);
}

.form-input, .form-select, .form-textarea {
  background: #0d0f15;
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
  padding: 0.6rem 0.85rem;
  color: #fff;
  font-size: 0.88rem;
  outline: none;
  transition: border-color 0.15s;
  width: 100%;
}

.form-input:focus, .form-select:focus, .form-textarea:focus {
  border-color: var(--rdr-crimson);
}

.checkbox-group {
  margin-top: 1rem;
  margin-bottom: 0.5rem;
  padding: 0.85rem;
  background: #0f1118;
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
}

.checkbox-label {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  cursor: pointer;
}

.form-checkbox {
  margin-top: 0.2rem;
  accent-color: var(--rdr-crimson);
  width: 16px;
  height: 16px;
}

/* Alerts & Code Blocks */
.alert-box-warning {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
  padding: 0.9rem;
  border-radius: 4px;
  background: rgba(217, 119, 6, 0.12);
  border: 1px solid rgba(217, 119, 6, 0.35);
}

.key-summary-strip {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  padding: 0.85rem;
  background: #0c0e14;
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
}

.code-block-container {
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
  overflow: hidden;
  background: #090a0f;
}

.code-block-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.45rem 0.85rem;
  background: #11131b;
  border-bottom: 1px solid var(--border-subtle);
}

.code-pre {
  padding: 0.85rem;
  font-size: 0.78rem;
  color: #facc15;
  overflow-x: auto;
  max-height: 180px;
  white-space: pre-wrap;
  word-break: break-all;
}

/* Utilities */
.text-crimson { color: var(--rdr-crimson); }
.text-green { color: var(--linux-green); }
.text-amber { color: var(--rdr-amber); }
.text-muted { color: var(--text-muted); }

.spin-anim {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
