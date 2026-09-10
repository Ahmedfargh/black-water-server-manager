<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useToastStore } from '../stores/toast'
import { User, ShieldCheck, Mail, Save, Bell, BellRing, Lock, Camera } from 'lucide-vue-next'
import InteractiveImage from '../components/InteractiveImage.vue'

const authStore = useAuthStore()
const toast = useToastStore()

// Profile Data
const username = ref('')
const email = ref('')
const password = ref('')
const imageFile = ref(null)

// Notification Settings Data
const notificationDriver = ref('Telegram')
const telegramBotToken = ref('')
const telegramChatId = ref('')
const discordBotToken = ref('')
const discordChannelId = ref('')
const webhookUrl = ref('')
const webhookSecret = ref('')

const isProfileSubmitting = ref(false)
const isNotificationsSubmitting = ref(false)

onMounted(async () => {
  try {
    const userProfile = await authStore.fetchProfile()
    populateData(userProfile)
  } catch (err) {
    populateData(authStore.user)
  }
})

const populateData = (userData) => {
  if (!userData) return
  username.value = userData.username || ''
  email.value = userData.email || ''
  
  notificationDriver.value = userData.notification_driver || 'Telegram'
  telegramBotToken.value = userData.telegram_bot_token || ''
  telegramChatId.value = userData.telegram_chat_id || ''
  discordBotToken.value = userData.discord_bot_token || ''
  discordChannelId.value = userData.discord_channel_id || ''
  webhookUrl.value = userData.webhook_url || ''
  webhookSecret.value = userData.webhook_secret || ''
}

const handleFileChange = (e) => {
  if (e.target.files.length > 0) {
    imageFile.value = e.target.files[0]
  }
}

const submitProfileUpdate = async () => {
  isProfileSubmitting.value = true
  try {
    const formData = new FormData()
    formData.append('username', username.value)
    formData.append('email', email.value)
    if (password.value) {
      formData.append('password', password.value)
    }
    if (imageFile.value) {
      formData.append('avatar', imageFile.value)
    }
    
    await authStore.updateProfile(formData)
    password.value = ''
    toast.success('PROFILE DATA SYNCHRONIZED SUCCESSFULLY')
  } catch (err) {
    toast.error(`PROFILE UPDATE FAILED: ${err.response?.data?.error || err.message}`)
  } finally {
    isProfileSubmitting.value = false
  }
}

const submitNotificationSettings = async () => {
  isNotificationsSubmitting.value = true
  try {
    const payload = {
      notification_driver: notificationDriver.value,
      telegram_bot_token: telegramBotToken.value,
      telegram_chat_id: telegramChatId.value,
      discord_bot_token: discordBotToken.value,
      discord_channel_id: discordChannelId.value,
      webhook_url: webhookUrl.value,
      webhook_secret: webhookSecret.value
    }
    
    await authStore.updateNotifications(payload)
    toast.success('NOTIFICATION PROTOCOLS UPDATED')
  } catch (err) {
    toast.error(`NOTIFICATION UPDATE FAILED: ${err.response?.data?.error || err.message}`)
  } finally {
    isNotificationsSubmitting.value = false
  }
}
</script>

<template>
  <div class="profile-view">
    <div class="header-row">
      <div class="header-title-block">
        <h2 class="page-title">OPERATOR DOSSIER</h2>
        <span class="sub-label">ACCOUNT SECURITY & TELEMETRY DISPATCH</span>
      </div>
    </div>

    <div class="profile-grid">
      <!-- Profile Card -->
      <div class="tron-card settings-card">
        <div class="card-header">
           <User class="crimson-icon" :size="20" />
           <h3>ACCOUNT CONFIGURATION</h3>
        </div>
        
        <form @submit.prevent="submitProfileUpdate" class="settings-form">
          <div class="input-group">
            <label>USERNAME</label>
            <div class="input-with-icon">
              <User :size="16" class="field-icon" />
              <input v-model="username" type="text" placeholder="Access ID" required />
            </div>
          </div>
          
          <div class="input-group">
            <label>EMAIL ADDRESS</label>
            <div class="input-with-icon">
              <Mail :size="16" class="field-icon" />
              <input v-model="email" type="email" placeholder="Grid Address" required />
            </div>
          </div>

          <div class="input-group">
            <label>SECURITY KEY (PASSWORD)</label>
            <span class="hint-text">Leave blank to maintain current passphrase</span>
            <div class="input-with-icon">
              <Lock :size="16" class="field-icon" />
              <input v-model="password" type="password" placeholder="••••••••" />
            </div>
          </div>
          
          <div class="input-group">
            <label>OPERATOR BADGE (AVATAR)</label>
            <div class="current-avatar-preview" v-if="authStore.user?.image_path && authStore.user.image_path.includes('/uploads/')">
              <InteractiveImage :src="authStore.user.image_path" customClass="profile-avatar" />
            </div>
            <div class="input-with-icon">
              <Camera :size="16" class="field-icon" />
              <input type="file" @change="handleFileChange" accept="image/*" class="file-input" />
            </div>
          </div>

          <div class="form-actions">
            <button type="submit" class="tron-btn" :disabled="isProfileSubmitting">
              <Save :size="16" />
              {{ isProfileSubmitting ? 'UPDATING...' : 'SAVE CONFIGURATION' }}
            </button>
          </div>
        </form>
      </div>

      <!-- Notifications Card -->
      <div class="tron-card settings-card">
        <div class="card-header">
           <BellRing class="crimson-icon" :size="20" />
           <h3>ALERT & DISPATCH PROTOCOLS</h3>
        </div>
        
        <form @submit.prevent="submitNotificationSettings" class="settings-form">
          <div class="input-group">
            <label>PRIMARY ALERT DRIVER</label>
            <select v-model="notificationDriver" required class="full-width">
              <option value="Telegram">TELEGRAM BOT DISPATCH</option>
              <option value="Discord">DISCORD WEBHOOK RELAY</option>
              <option value="Webhook">CUSTOM HTTP WEBHOOK</option>
              <option value="None">SILENT MODE (MUTED)</option>
            </select>
          </div>

          <div v-if="notificationDriver === 'Telegram'" class="driver-settings">
            <h4 class="driver-title">TELEGRAM CONFIGURATION</h4>
            <div class="input-group">
              <label>BOT TOKEN</label>
              <input v-model="telegramBotToken" type="text" placeholder="Bot API Token" />
            </div>
            <div class="input-group">
              <label>CHAT ID</label>
              <input v-model="telegramChatId" type="text" placeholder="Destination Chat ID" />
            </div>
          </div>

          <div v-if="notificationDriver === 'Discord'" class="driver-settings">
            <h4 class="driver-title">DISCORD CONFIGURATION</h4>
            <div class="input-group">
              <label>BOT TOKEN</label>
              <input v-model="discordBotToken" type="text" placeholder="Bot API Token" />
            </div>
            <div class="input-group">
              <label>CHANNEL ID</label>
              <input v-model="discordChannelId" type="text" placeholder="Channel Target ID" />
            </div>
          </div>

          <div v-if="notificationDriver === 'Webhook'" class="driver-settings">
            <h4 class="driver-title">WEBHOOK CONFIGURATION</h4>
            <div class="input-group">
              <label>TARGET URL</label>
              <input v-model="webhookUrl" type="url" placeholder="https://api.yourdomain.com/webhook" />
            </div>
            <div class="input-group">
              <label>WEBHOOK SECRET</label>
              <input v-model="webhookSecret" type="password" placeholder="Signing Secret Key (Optional)" />
            </div>
          </div>

          <div class="form-actions form-actions-spaced">
            <button type="submit" class="tron-btn" :disabled="isNotificationsSubmitting">
              <Bell :size="16" />
              {{ isNotificationsSubmitting ? 'UPDATING...' : 'UPDATE PROTOCOLS' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.profile-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border-color);
}

.header-title-block {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.page-title {
  font-size: 1.25rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  color: var(--text-primary);
  margin: 0;
}

.sub-label {
  font-size: 0.7rem;
  font-weight: 700;
  color: var(--text-muted);
  font-family: var(--font-data);
  letter-spacing: 0.06em;
}

.profile-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(380px, 1fr));
  gap: 1.5rem;
}

@media (max-width: 600px) {
  .profile-grid {
    grid-template-columns: 1fr;
  }
}

.settings-card {
  padding: 1.5rem;
  border-inline-start: 3px solid var(--theme-crimson);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border-color);
}

.card-header h3 {
  font-size: 0.95rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: var(--text-primary);
  margin: 0;
}

.crimson-icon {
  color: var(--theme-crimson);
}

.settings-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.input-group label {
  display: block;
  font-size: 0.75rem;
  margin-bottom: 0.4rem;
  color: var(--text-secondary);
  font-weight: 600;
  letter-spacing: 0.04em;
}

.hint-text {
  font-size: 0.7rem;
  color: var(--theme-amber);
  margin-bottom: 0.4rem;
  display: inline-block;
  font-family: var(--font-data);
}

.input-with-icon {
  position: relative;
  display: flex;
  align-items: center;
}

.field-icon {
  position: absolute;
  inset-inline-start: 0.85rem;
  color: var(--text-muted);
  pointer-events: none;
}

.input-with-icon input {
  padding-inline-start: 2.5rem !important;
}

.settings-form input,
.settings-form select {
  width: 100%;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  padding: 0.6rem 0.85rem;
  color: var(--text-primary);
  font-family: var(--font-data);
  font-size: 0.85rem;
  outline: none;
  border-radius: var(--radius-sm);
  transition: border-color var(--transition-fast);
}

.settings-form input:focus,
.settings-form select:focus {
  border-color: var(--theme-crimson);
}

.file-input {
  padding: 0.45rem !important;
  font-size: 0.8rem;
}

.current-avatar-preview {
  margin-bottom: 0.75rem;
  width: 56px;
  height: 56px;
  border-radius: var(--radius-sm);
  overflow: hidden;
  border: 2px solid var(--theme-crimson);
}

.profile-avatar {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.driver-settings {
  background: rgba(0, 0, 0, 0.25);
  padding: 1.25rem;
  border: 1px solid var(--border-color);
  border-inline-start: 3px solid var(--theme-amber);
  border-radius: var(--radius-sm);
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-top: 0.5rem;
}

.driver-title {
  font-size: 0.75rem;
  letter-spacing: 0.05em;
  margin: 0;
  color: var(--theme-amber);
  font-weight: 700;
}

.form-actions {
  margin-top: 0.5rem;
  display: flex;
  justify-content: flex-end;
}

.form-actions-spaced {
  margin-top: 1.5rem;
}
</style>
