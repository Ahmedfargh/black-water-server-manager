<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useToastStore } from '../stores/toast'
import { Cpu, Lock, Mail, Loader2 } from 'lucide-vue-next'

const authStore = useAuthStore()
const toast = useToastStore()
const router = useRouter()
...
const handleLogin = async () => {
  isLoading.value = true
  error.value = ''
  try {
    await authStore.login(email.value, password.value)
    toast.success('UPLINK ESTABLISHED: Welcome back, Commander.')
    router.push('/')
  } catch (err) {
    const msg = err.response?.data?.error || 'ACCESS DENIED: Authentication failed.'
    error.value = msg
    toast.error(msg)
  } finally {
    isLoading.value = false
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
            <Cpu class="glow-cyan" :size="48" />
          </div>
          <h1>BLACKWATER</h1>
          <p class="subtitle">SECURE GRID ACCESS</p>
        </div>

        <form @submit.prevent="handleLogin" class="login-form">
          <div class="input-group">
            <label>IDENTIFIER</label>
            <div class="input-wrapper">
              <Mail :size="18" />
              <input 
                v-model="email" 
                type="email" 
                placeholder="USER@GRID.SYS" 
                required
              />
            </div>
          </div>

          <div class="input-group">
            <label>ACCESS KEY</label>
            <div class="input-wrapper">
              <Lock :size="18" />
              <input 
                v-model="password" 
                type="password" 
                placeholder="********" 
                required
              />
            </div>
          </div>

          <div v-if="error" class="error-msg glow-orange">
            {{ error }}
          </div>

          <button :disabled="isLoading" type="submit" class="login-btn">
            <Loader2 v-if="isLoading" class="spinner" :size="18" />
            <span v-else>INITIATE UPLINK</span>
          </button>
        </form>

        <div class="login-footer">
          <span class="system-tag">SYSTEM v1.24.03</span>
          <span class="status-tag">STATUS: SECURE</span>
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
  gap: 2.5rem;
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
  font-size: 2rem;
  margin-bottom: 0.5rem;
  text-shadow: var(--text-glow);
}

.subtitle {
  font-size: 0.9rem;
  color: var(--text-secondary);
  letter-spacing: 4px;
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
