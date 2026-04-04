<script setup>
import { useToastStore } from '../stores/toast'
import { 
  CheckCircle2, 
  AlertCircle, 
  Info, 
  AlertTriangle, 
  X 
} from 'lucide-vue-next'

const toastStore = useToastStore()

const getIcon = (type) => {
  switch (type) {
    case 'success': return CheckCircle2
    case 'error': return AlertCircle
    case 'warning': return AlertTriangle
    default: return Info
  }
}
</script>

<template>
  <div class="toast-container">
    <transition-group name="toast">
      <div 
        v-for="toast in toastStore.toasts" 
        :key="toast.id" 
        class="toast-item tron-card"
        :class="toast.type"
      >
        <div class="toast-content">
          <component :is="getIcon(toast.type)" :size="20" class="toast-icon" />
          <span class="toast-message">{{ toast.message }}</span>
        </div>
        <button @click="toastStore.remove(toast.id)" class="close-btn">
          <X :size="16" />
        </button>
        <div class="progress-bar"></div>
      </div>
    </transition-group>
  </div>
</template>

<style scoped>
.toast-container {
  position: fixed;
  top: 2rem;
  right: 2rem;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  pointer-events: none;
}

.toast-item {
  pointer-events: auto;
  min-width: 300px;
  max-width: 450px;
  padding: 1rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  background: rgba(10, 18, 25, 0.9);
  border-left: 4px solid var(--neon-cyan);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
}

.toast-item.success { border-left-color: var(--neon-cyan); }
.toast-item.error { border-left-color: var(--neon-orange); }
.toast-item.warning { border-left-color: #ffcc00; }
.toast-item.info { border-left-color: #0088ff; }

.toast-content {
  display: flex;
  align-items: center;
  gap: 0.8rem;
}

.toast-icon {
  flex-shrink: 0;
}

.success .toast-icon { color: var(--neon-cyan); }
.error .toast-icon { color: var(--neon-orange); }
.warning .toast-icon { color: #ffcc00; }
.info .toast-icon { color: #0088ff; }

.toast-message {
  font-family: var(--font-header);
  font-size: 0.9rem;
  letter-spacing: 1px;
  text-transform: uppercase;
  color: var(--text-primary);
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  color: #fff;
}

.progress-bar {
  position: absolute;
  bottom: 0;
  left: 0;
  height: 2px;
  width: 100%;
  background: var(--neon-cyan);
  opacity: 0.3;
  animation: progress 4s linear forwards;
}

.error .progress-bar { background: var(--neon-orange); }

@keyframes progress {
  from { width: 100%; }
  to { width: 0%; }
}

/* Animations */
.toast-enter-active,
.toast-leave-active {
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(50px) scale(0.9);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(100px);
}
</style>
