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
      <h2 class="glow-cyan">PERSONNEL DOSSIER</h2>
    </div>

    <div class="profile-grid">
      <!-- Profile Card -->
      <div class="tron-card settings-card">
        <div class="card-header">
           <User class="glow-cyan" :size="24" />
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
            <label>EMAIL TRANSMISSION VECTOR</label>
            <div class="input-with-icon">
              <Mail :size="16" class="field-icon" />
              <input v-model="email" type="email" placeholder="Grid Address" required />
            </div>
          </div>

          <div class="input-group">
            <label>SECURITY KEY (PASSWORD)</label>
            <span class="hint-text">Leave blank to maintain current clearance code</span>
            <div class="input-with-icon">
              <Lock :size="16" class="field-icon" />
              <input v-model="password" type="password" placeholder="********" />
            </div>
          </div>
          
          <div class="input-group">
            <label>BIOMETRIC SCAN (AVATAR)</label>
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
              <Save :size="18" />
              {{ isProfileSubmitting ? 'UPDATING...' : 'SAVE CONFIGURATION' }}
            </button>
          </div>
        </form>
      </div>

      <!-- Notifications Card -->
      <div class="tron-card settings-card">
        <div class="card-header">
           <BellRing class="glow-cyan" :size="24" />
           <h3>NOTIFICATION PROTOCOLS</h3>
        </div>
        
        <form @submit.prevent="submitNotificationSettings" class="settings-form">
          <div class="input-group">
            <label>PRIMARY ALERT DRIVER</label>
            <select v-model="notificationDriver" required class="full-width">
              <option value="Telegram">TELEGRAM NETWORK</option>
              <option value="Discord">DISCORD RELAY</option>
              <option value="Webhook">CUSTOM WEBHOOK</option>
              <option value="None">SILENT MODE (NONE)</option>
            </select>
          </div>

          <div v-if="notificationDriver === 'Telegram'" class="driver-settings">
            <h4 class="driver-title text-secondary">TELEGRAM CONFIGURATION</h4>
            <div class="input-group">
              <label>BOT TOKEN</label>
              <input v-model="telegramBotToken" type="text" placeholder="Bot API Token" />
            </div>
            <div class="input-group">
              <label>CHAT ID</label>
              <input v-model="telegramChatId" type="text" placeholder="Destination ID" />
            </div>
          </div>

          <div v-if="notificationDriver === 'Discord'" class="driver-settings">
            <h4 class="driver-title text-secondary">DISCORD CONFIGURATION</h4>
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
            <h4 class="driver-title text-secondary">WEBHOOK CONFIGURATION</h4>
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
              <Bell :size="18" />
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
  gap: 2rem;
}

.profile-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 2rem;
}

@media (max-width: 600px) {
  .profile-grid {
    grid-template-columns: 1fr;
  }
}

.settings-card {
  padding: 1.5rem;
  background: rgba(10, 15, 20, 0.4);
  backdrop-filter: blur(10px);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid rgba(0, 242, 255, 0.1);
}

.card-header h3 {
  font-size: 1.2rem;
  letter-spacing: 2px;
}

.settings-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.input-group label {
  display: block;
  font-size: 0.8rem;
  margin-bottom: 0.5rem;
  color: var(--text-secondary);
  letter-spacing: 1px;
}

.hint-text {
  font-size: 0.7rem;
  color: var(--neon-orange);
  margin-bottom: 0.5rem;
  display: inline-block;
}

.input-with-icon {
  position: relative;
  display: flex;
  align-items: center;
}

.field-icon {
  position: absolute;
  left: 1rem;
  color: var(--neon-cyan);
  opacity: 0.7;
}

.input-with-icon input {
  padding-left: 2.8rem;
}

.settings-form input,
.settings-form select {
  width: 100%;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(0, 242, 255, 0.2);
  padding: 0.8rem;
  color: var(--text-primary);
  font-family: var(--font-header);
  outline: none;
  border-radius: 4px;
}

.settings-form input:focus,
.settings-form select:focus {
  border-color: var(--neon-cyan);
  box-shadow: 0 0 10px var(--neon-cyan-glow);
}

.file-input {
  padding: 0.6rem!important;
  font-size: 0.9rem;
}

.current-avatar-preview {
  margin-bottom: 1rem;
  width: 64px;
  height: 64px;
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid var(--neon-cyan);
  box-shadow: 0 0 10px var(--neon-cyan-glow);
}

.profile-avatar {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.driver-settings {
  background: rgba(0, 242, 255, 0.02);
  padding: 1.5rem;
  border: 1px solid rgba(0, 242, 255, 0.1);
  border-radius: 4px;
  display: flex;
  flex-direction: column;
  gap: 1.2rem;
  margin-top: 1rem;
}

.driver-title {
  font-size: 0.8rem;
  letter-spacing: 2px;
  margin-bottom: 0.5rem;
}

.form-actions {
  margin-top: 1rem;
  display: flex;
  justify-content: flex-end;
}

.form-actions-spaced {
  margin-top: 2.5rem;
}
</style>
