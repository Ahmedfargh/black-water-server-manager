<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { useToastStore } from '../stores/toast'
import { Mail, Lock, Loader2, ArrowLeft } from 'lucide-vue-next'
import BrandLogo from '../components/BrandLogo.vue'

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
    await authStore.verifyEmail(userId.value, code.value)
    toast.success(t('auth.email_verified'))
    router.push('/login')
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
    await authStore.resendEmail(userId.value)
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
    <div class="frontier-backdrop"></div>
    
    <div class="login-container">
      <div class="tron-card login-card">
        <div class="login-brand">
          <BrandLogo size="md" :badgeText="$t('auth.verify_email')" />
        </div>

        <div class="login-header">
          <p class="subtitle">{{ $t('auth.code_sent') }}</p>
          <p class="email-display font-data">{{ email }}</p>
        </div>

        <form @submit.prevent="handleVerify" class="login-form">
          <div class="input-group">
            <label class="font-data">{{ $t('auth.verification_code') }}</label>
            <div class="input-wrapper">
              <Lock :size="17" class="input-icon" />
              <input 
                v-model="code" 
                type="text" 
                maxlength="6"
                placeholder="000000" 
                required
                autocomplete="one-time-code"
                class="otp-input font-data"
              />
            </div>
          </div>

          <div v-if="error" class="error-msg font-data">
            {{ error }}
          </div>

          <button :disabled="isLoading || code.length < 6" type="submit" class="frontier-btn font-data">
            <Loader2 v-if="isLoading" class="spinner" :size="18" />
            <span v-else>{{ $t('auth.verify') }}</span>
          </button>
        </form>

        <div class="actions">
          <button @click="handleResend" :disabled="isResending" class="text-btn font-data">
            {{ isResending ? $t('common.loading') : $t('auth.resend_code') }}
          </button>
          <router-link to="/login" class="text-btn back-btn font-data">
            <ArrowLeft :size="14" />
            {{ $t('nav.login') }}
          </router-link>
        </div>

        <div class="login-footer font-data">
          <span class="system-tag">{{ $t('app.system_v') }}</span>
          <span class="status-tag active">
            <span class="beacon-online"></span>
            {{ $t('app.secure') }}
          </span>
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
  background-color: #0c0d11;
  overflow: hidden;
}

.frontier-backdrop {
  position: absolute;
  inset: 0;
  background-image: 
    radial-gradient(circle at 50% 25%, rgba(220, 38, 38, 0.08) 0%, transparent 60%),
    radial-gradient(circle at 80% 80%, rgba(217, 119, 6, 0.04) 0%, transparent 50%);
  pointer-events: none;
}

.login-container {
  width: 100%;
  max-width: 440px;
  z-index: 10;
  padding: 1.5rem;
}

.login-card {
  padding: 2.8rem 2.2rem;
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
  border: 1px solid var(--border-crimson);
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.6);
  background: #13151d;
  border-radius: 4px;
}

.login-brand {
  display: flex;
  justify-content: center;
}

.login-header {
  text-align: center;
}

.subtitle {
  font-size: 0.82rem;
  color: var(--text-muted);
}

.email-display {
  color: var(--rdr-amber);
  font-size: 0.85rem;
  font-weight: 600;
  margin-top: 0.35rem;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.input-group label {
  display: block;
  font-size: 0.72rem;
  color: var(--text-muted);
  margin-bottom: 0.45rem;
  letter-spacing: 1px;
  text-transform: uppercase;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  border: 1px solid var(--border-subtle);
  background: #0f1118;
  border-radius: 3px;
  transition: all 0.15s ease;
}

.input-wrapper:focus-within {
  border-color: var(--rdr-crimson);
  background: #151722;
}

.input-icon {
  margin: 0 0.85rem;
  color: var(--text-muted);
  flex-shrink: 0;
}

.input-wrapper input {
  flex: 1;
  background: transparent;
  border: none;
  padding: 0.75rem 0.85rem 0.75rem 0;
  color: #ffffff;
  font-family: var(--font-data);
  font-size: 1.1rem;
  letter-spacing: 4px;
  text-align: center;
  outline: none;
}

.frontier-btn {
  margin-top: 0.5rem;
  background: var(--rdr-crimson);
  border: 1px solid var(--rdr-crimson-hover);
  color: #ffffff;
  padding: 0.75rem 1.2rem;
  font-family: var(--font-header);
  font-weight: 700;
  font-size: 0.88rem;
  letter-spacing: 1.5px;
  border-radius: 3px;
  cursor: pointer;
  transition: all 0.15s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  text-transform: uppercase;
}

.frontier-btn:hover:not(:disabled) {
  background: var(--rdr-crimson-hover);
  box-shadow: 0 4px 16px var(--rdr-crimson-glow);
}

.frontier-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 0.5rem;
}

.text-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 0.8rem;
  cursor: pointer;
  transition: color 0.15s ease;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.text-btn:hover:not(:disabled) {
  color: var(--rdr-crimson);
}

.error-msg {
  font-size: 0.78rem;
  text-align: center;
  color: #f87171;
  background: rgba(220, 38, 38, 0.1);
  padding: 0.5rem;
  border-radius: 3px;
  border: 1px solid rgba(220, 38, 38, 0.25);
}

.login-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.7rem;
  color: var(--text-muted);
  border-top: 1px solid var(--border-subtle);
  padding-top: 1.25rem;
}

.spinner {
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
