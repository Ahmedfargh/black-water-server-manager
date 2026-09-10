<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { useToastStore } from '../stores/toast'
import { Lock, Mail, Loader2, ArrowRight } from 'lucide-vue-next'
import BrandLogo from '../components/BrandLogo.vue'

const { t } = useI18n()
const authStore = useAuthStore()
const toast = useToastStore()
const router = useRouter()
let isLoading = ref(false)
let error = ref('')
let email = ref('')
let password = ref('')

const handleLogin = async () => {
  isLoading.value = true
  error.value = ''
  try {
    const data = await authStore.login(email.value, password.value)
    if (data.requires_verification) {
      toast.info(t('auth.code_sent'))
      router.push({ 
        name: 'VerifyEmail', 
        query: { user_id: data.user_id, email: data.email } 
      })
      return
    }
    if (data.requires_otp) {
      toast.info(t('auth.otp_required'))
      router.push({ 
        name: 'VerifyOTP', 
        query: { user_id: data.user_id, email: data.email } 
      })
      return
    }
    toast.success(t('common.uplink_established'))
    router.push('/')
  } catch (err) {
    const msg = err.response?.data?.error || t('common.access_denied')
    error.value = msg
    toast.error(msg)
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="frontier-backdrop"></div>
    
    <div class="login-container">
      <div class="tron-card login-card">
        <!-- New Brand Logo Presentation -->
        <div class="login-brand">
          <BrandLogo size="lg" :badgeText="$t('app.secure_access')" />
        </div>

        <form @submit.prevent="handleLogin" class="login-form">
          <div class="input-group">
            <label class="font-data">{{ $t('common.identifier') }}</label>
            <div class="input-wrapper">
              <Mail :size="17" class="input-icon" />
              <input 
                v-model="email" 
                type="email" 
                placeholder="admin@example.com" 
                required
                autocomplete="username"
              />
            </div>
          </div>

          <div class="input-group">
            <label class="font-data">{{ $t('common.access_key') }}</label>
            <div class="input-wrapper">
              <Lock :size="17" class="input-icon" />
              <input 
                v-model="password" 
                type="password" 
                placeholder="••••••••••••" 
                required
                autocomplete="current-password"
              />
            </div>
          </div>

          <div v-if="error" class="error-msg font-data">
            {{ error }}
          </div>

          <button :disabled="isLoading" type="submit" class="frontier-login-btn font-data">
            <Loader2 v-if="isLoading" class="spinner" :size="18" />
            <template v-else>
              <span>{{ $t('common.initiate_uplink') }}</span>
              <ArrowRight :size="16" />
            </template>
          </button>
        </form>

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
  gap: 2rem;
  border: 1px solid var(--border-crimson);
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.6);
  background: #13151d;
  border-radius: 4px;
}

.login-brand {
  display: flex;
  justify-content: center;
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
  box-shadow: 0 0 0 1px var(--rdr-crimson-dim);
}

.input-icon {
  margin: 0 0.85rem;
  color: var(--text-muted);
  flex-shrink: 0;
}

.input-wrapper:focus-within .input-icon {
  color: var(--rdr-crimson);
}

.input-wrapper input {
  flex: 1;
  background: transparent;
  border: none;
  padding: 0.75rem 0.85rem 0.75rem 0;
  color: #ffffff;
  font-family: var(--font-data);
  font-size: 0.88rem;
  outline: none;
}

[dir="rtl"] .input-wrapper input {
  padding: 0.75rem 0 0.75rem 0.85rem;
}

.frontier-login-btn {
  margin-top: 0.75rem;
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

.frontier-login-btn:hover:not(:disabled) {
  background: var(--rdr-crimson-hover);
  box-shadow: 0 4px 16px var(--rdr-crimson-glow);
}

.frontier-login-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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
