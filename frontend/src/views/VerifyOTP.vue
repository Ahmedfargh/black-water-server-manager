<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { useToastStore } from '../stores/toast'
import { ShieldCheck, Lock, Loader2, ArrowLeft } from 'lucide-vue-next'

const { t } = useI18n()
const authStore = useAuthStore()
const toast = useToastStore()
const router = useRouter()
const route = useRoute()

const isLoading = ref(false)
const isResending = ref(false)
const error = ref('')
const code = ref('')
const userId = ref(route.query.user_id)
const email = ref(route.query.email)

onMounted(() => {
  if (!userId.value) {
    router.push('/login')
  }
})

const handleVerify = async () => {
  if (code.value.length < 6) return
  
  isLoading.value = true
  error.value = ''
  try {
    await authStore.verifyOTP(userId.value, code.value)
    toast.success(t('common.uplink_established'))
    router.push('/')
  } catch (err) {
    const msg = err.response?.data?.error || t('auth.verification_failed')
    error.value = msg
    toast.error(msg)
  } finally {
    isLoading.value = false
  }
}

const handleResend = async () => {
  isResending.value = true
  try {
    await authStore.resendOTP(userId.value)
    toast.success(t('auth.code_resent'))
  } catch (err) {
    toast.error(t('common.action_failed'))
  } finally {
    isResending.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="bg-grid"></div>
    
    <div class="login-container">
      <div class="tron-card login-card">
        <div class="login-header">
          <div class="logo-glow">
            <ShieldCheck class="glow-cyan" :size="48" />
          </div>
          <h1>{{ $t('auth.verify_otp') }}</h1>
          <p class="subtitle">{{ $t('auth.code_sent') }}</p>
          <p class="email-display">{{ email }}</p>
        </div>

        <form @submit.prevent="handleVerify" class="login-form">
          <div class="input-group">
            <label>{{ $t('auth.verification_code') }}</label>
            <div class="input-wrapper">
              <Lock :size="18" />
              <input 
                v-model="code" 
                type="text" 
                maxlength="6"
                :placeholder="'_ _ _ _ _ _'" 
                required
                autocomplete="one-time-code"
                class="otp-input"
              />
            </div>
          </div>

          <div v-if="error" class="error-msg glow-orange">
            {{ error }}
          </div>

          <button :disabled="isLoading || code.length < 6" type="submit" class="login-btn">
            <Loader2 v-if="isLoading" class="spinner" :size="18" />
            <span v-else>{{ $t('auth.verify') }}</span>
          </button>
        </form>

        <div class="actions">
          <button @click="handleResend" :disabled="isResending" class="text-btn">
            {{ isResending ? $t('common.loading') : $t('auth.resend_code') }}
          </button>
          <router-link to="/login" class="text-btn back-btn">
            <ArrowLeft :size="14" />
            {{ $t('nav.login') }}
          </router-link>
        </div>

        <div class="login-footer">
          <span class="system-tag">{{ $t('app.system_v') }}</span>
          <span class="status-tag">{{ $t('app.status') }}: {{ $t('app.secure') }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  background-color: var(--bg-black);
}

.bg-grid {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image: 
    linear-gradient(var(--grid-line) 1px, transparent 1px),
    linear-gradient(90deg, var(--grid-line) 1px, transparent 1px);
  background-size: 80px 80px;
  opacity: 0.3;
}

.login-container {
  width: 100%;
  max-width: 420px;
  z-index: 10;
  padding: 1.5rem;
}

.login-card {
  padding: 3rem 2rem;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.login-header {
  text-align: center;
}

.logo-glow {
  margin-bottom: 1.5rem;
  display: inline-block;
  padding: 1rem;
  border: 1px solid rgba(0, 242, 255, 0.2);
  border-radius: 50%;
  box-shadow: 0 0 20px rgba(0, 242, 255, 0.1);
}

.login-header h1 {
  font-size: 1.5rem;
  margin-bottom: 0.5rem;
  text-shadow: var(--text-glow);
  letter-spacing: 2px;
}

.subtitle {
  font-size: 0.8rem;
  color: var(--text-secondary);
  line-height: 1.4;
}

.email-display {
  font-family: var(--font-data);
  color: var(--neon-cyan);
  font-size: 0.8rem;
  margin-top: 0.5rem;
  text-shadow: 0 0 5px var(--neon-cyan-glow);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.input-group label {
  display: block;
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
  letter-spacing: 2px;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  border: 1px solid rgba(0, 242, 255, 0.2);
  background: rgba(0, 242, 255, 0.02);
  transition: all 0.3s ease;
}

.input-wrapper:focus-within {
  border-color: var(--neon-cyan);
  box-shadow: 0 0 10px rgba(0, 242, 255, 0.1);
}

.input-wrapper svg {
  margin: 0 1rem;
  color: var(--text-secondary);
}

.input-wrapper input {
  flex: 1;
  background: transparent;
  border: none;
  padding: 0.8rem 0;
  color: var(--text-primary);
  font-family: var(--font-data);
  outline: none;
  letter-spacing: 4px;
}

.otp-input {
  text-align: center;
  font-size: 1.2rem !important;
  font-weight: bold;
}

.login-btn {
  margin-top: 1rem;
  background: transparent;
  border: 1px solid var(--neon-cyan);
  color: var(--neon-cyan);
  padding: 1rem;
  font-family: var(--font-header);
  font-weight: 700;
  letter-spacing: 3px;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.8rem;
}

.login-btn:hover:not(:disabled) {
  background: var(--neon-cyan);
  color: var(--bg-black);
  box-shadow: 0 0 20px var(--neon-cyan-glow);
}

.login-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.text-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 0.75rem;
  cursor: pointer;
  transition: color 0.3s ease;
  letter-spacing: 1px;
}

.text-btn:hover {
  color: var(--neon-cyan);
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  text-decoration: none;
}

.error-msg {
  font-size: 0.8rem;
  text-align: center;
  font-style: italic;
}

.login-footer {
  display: flex;
  justify-content: space-between;
  font-size: 0.7rem;
  color: var(--text-secondary);
  border-top: 1px solid rgba(0, 242, 255, 0.1);
  padding-top: 1.5rem;
}

.spinner {
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
