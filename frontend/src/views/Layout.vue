<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { 
  LayoutDashboard, 
  Box, 
  Terminal, 
  ShieldCheck, 
  Globe, 
  History, 
  LogOut,
  User,
  Users,
  Settings,
  Cpu,
  Menu,
  X,
  BarChart3,
  Languages
} from 'lucide-vue-next'
import { useAuthStore } from '../stores/auth'
import { useToastStore } from '../stores/toast'
import { useSettingsStore } from '../stores/settings'
import InteractiveImage from '../components/InteractiveImage.vue'

const { t, locale } = useI18n()
const authStore = useAuthStore()
const settingsStore = useSettingsStore()
const toast = useToastStore()
const router = useRouter()
const route = useRoute()
const isSidebarOpen = ref(window.innerWidth > 1024)

const updateSidebarState = () => {
  if (window.innerWidth <= 1024) {
    isSidebarOpen.value = false
  } else {
    isSidebarOpen.value = true
  }
}

onMounted(() => {
  window.addEventListener('resize', updateSidebarState)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateSidebarState)
})

const toggleSidebar = () => {
  isSidebarOpen.value = !isSidebarOpen.value
}

const handleNavClick = () => {
  if (window.innerWidth <= 1024) {
    isSidebarOpen.value = false
  }
}

const localizedRouteName = computed(() => {
  const name = route.name?.toLowerCase()
  if (!name) return t('common.system')
  
  // Map route names to translation keys
  const routeMap = {
    'dashboard': 'nav.dashboard',
    'docker': 'nav.docker',
    'terminal': 'nav.terminal',
    'processes': 'nav.processes',
    'firewall': 'nav.firewall',
    'sites': 'nav.sites',
    'users': 'nav.users',
    'auditlogs': 'nav.audit_logs',
    'profile': 'nav.profile',
    'reports': 'nav.reports',
    'login': 'nav.login'
  }
  
  const key = routeMap[name] || `nav.${name}`
  const translated = t(key)
  return translated !== key ? translated.toUpperCase() : name.toUpperCase()
})

const menuItems = computed(() => [
  { name: t('nav.dashboard'), path: '/', icon: LayoutDashboard },
  { name: t('nav.docker'), path: '/docker', icon: Box },
  { name: t('nav.terminal'), path: '/terminal', icon: Terminal },
  { name: t('nav.processes'), path: '/processes', icon: Cpu },
  { name: t('nav.firewall'), path: '/firewall', icon: ShieldCheck },
  { name: t('nav.sites'), path: '/sites', icon: Globe },
  { name: t('nav.users'), path: '/users', icon: Users },
  { name: t('nav.audit_logs'), path: '/audit', icon: History },
  { name: t('nav.reports'), path: '/reports', icon: BarChart3 },
])

const toggleLanguage = () => {
  const newLang = locale.value === 'en' ? 'ar' : 'en'
  locale.value = newLang
  settingsStore.setLanguage(newLang)
  toast.info(newLang === 'ar' ? 'تم تغيير اللغة إلى العربية' : 'Language switched to English')
}

const handleLogout = () => {
  toast.info(t('common.connection_terminated'))
  authStore.logout()
}
</script>

<template>
  <div class="layout-wrapper">
    <!-- Sidebar Overlay for Mobile -->
    <div 
      v-if="isSidebarOpen" 
      class="sidebar-overlay" 
      @click="isSidebarOpen = false"
    ></div>

    <!-- Tron Sidebar -->
    <aside :class="['sidebar', { 'closed': !isSidebarOpen }]">
      <div class="logo-container">
        <Cpu class="glow-cyan" :size="32" />
        <span class="logo-text">{{ $t('app.logo_text') }}</span>
        <button class="close-sidebar-btn" @click="isSidebarOpen = false">
          <X :size="24" />
        </button>
      </div>

      <nav class="nav-menu">
        <router-link 
          v-for="item in menuItems" 
          :key="item.path" 
          :to="item.path"
          class="nav-item"
          :class="{ 'active': route.path === item.path }"
          @click="handleNavClick"
        >
          <component :is="item.icon" :size="20" class="nav-icon" />
          <span class="nav-text">{{ item.name }}</span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <div class="user-profile enhanced-link">
          <div class="avatar-wrap" v-if="authStore.user?.image_path && authStore.user.image_path.includes('/uploads/')">
            <InteractiveImage :src="authStore.user.image_path" customClass="user-avatar" :alt="$t('common.avatar')" />
          </div>
          <User v-else :size="18" />
          <router-link to="/profile" class="username-link">
            <span class="username">{{ authStore.user?.username || $t('common.admin') }}</span>
          </router-link>
        </div>
        <button @click="handleLogout" class="logout-btn">
          <LogOut :size="18" />
          <span>{{ $t('nav.logout') }}</span>
        </button>
      </div>
    </aside>

    <!-- Main Content Area -->
    <main class="main-content">
      <header class="top-bar">
        <div class="top-bar-left">
          <button class="menu-toggle-btn" @click="toggleSidebar">
             <Menu :size="24" />
          </button>
          <div class="breadcrumb">
            <span class="text-secondary">{{ $t('common.system') }}</span>
            <span class="separator">/</span>
            <span class="glow-cyan">{{ localizedRouteName }}</span>
          </div>
        </div>
        <div class="top-bar-right">
          <button @click="toggleLanguage" class="lang-switcher-btn" :title="locale === 'en' ? 'العربية' : 'English'">
            <Languages :size="20" />
            <span class="lang-text">{{ locale === 'en' ? 'AR' : 'EN' }}</span>
          </button>
          <div class="status-indicator">
            <span class="pulse-dot"></span>
            <span class="status-text glow-cyan">{{ $t('common.grid_online') }}</span>
          </div>
        </div>
      </header>

      <div class="content-container">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>
    </main>
  </div>
</template>

<style scoped>
.layout-wrapper {
  display: flex;
  height: 100vh;
  background-color: var(--bg-black);
}

/* Sidebar Styling */
.sidebar {
  width: 260px;
  background: var(--bg-card);
  border-right: 1px solid var(--neon-cyan-glow);
  display: flex;
  flex-direction: column;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  backdrop-filter: blur(20px);
  z-index: 1000;
  height: 100vh;
}

[dir="rtl"] .sidebar {
  border-right: none;
  border-left: 1px solid var(--neon-cyan-glow);
}

@media (max-width: 1024px) {
  .sidebar {
    position: fixed;
    left: 0;
    top: 0;
    transform: translateX(0);
  }
  
  [dir="rtl"] .sidebar {
    left: auto;
    right: 0;
  }
  
  .sidebar.closed {
    transform: translateX(-100%);
    pointer-events: none;
  }
  
  [dir="rtl"] .sidebar.closed {
    transform: translateX(100%);
  }
  
  .layout-wrapper {
    overflow: hidden;
  }
}

.sidebar-overlay {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(4px);
  z-index: 999;
}

@media (max-width: 1024px) {
  .sidebar-overlay {
     display: block;
  }
}

.logo-container {
  padding: 2.5rem 2rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgba(0, 242, 255, 0.1);
}

.close-sidebar-btn {
  display: none;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
}

@media (max-width: 1024px) {
  .close-sidebar-btn { display: block; }
}

.logo-text {
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: 3px;
  color: var(--text-primary);
  text-shadow: var(--text-glow);
}

.nav-menu {
  flex: 1;
  padding: 1.5rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.8rem 1.2rem;
  text-decoration: none;
  color: var(--text-secondary);
  border-radius: 4px;
  transition: all 0.2s ease;
  border: 1px solid transparent;
  text-transform: uppercase;
  font-size: 0.9rem;
  letter-spacing: 1px;
}

.nav-item:hover {
  background: rgba(0, 242, 255, 0.05);
  color: var(--neon-cyan);
  border-color: rgba(0, 242, 255, 0.1);
}

.nav-item.active {
  background: rgba(0, 242, 255, 0.1);
  color: var(--neon-cyan);
  border-color: var(--neon-cyan-glow);
  box-shadow: inset 0 0 10px rgba(0, 242, 255, 0.1);
}

.sidebar-footer {
  padding: 1.5rem;
  border-top: 1px solid rgba(0, 242, 255, 0.1);
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  color: var(--text-primary);
  text-decoration: none;
  padding: 0.5rem;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.user-profile:hover {
  background: rgba(0, 242, 255, 0.05);
  color: var(--neon-cyan);
}

.username-link {
  color: inherit;
  text-decoration: none;
}

.username-link:hover {
  text-shadow: 0 0 10px var(--neon-cyan-glow);
}

.avatar-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  overflow: hidden;
  border: 1px solid rgba(0, 242, 255, 0.3);
}

.user-avatar {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.logout-btn {
  background: transparent;
  border: 1px solid var(--neon-orange);
  color: var(--neon-orange);
  padding: 0.6rem;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  font-family: inherit;
  font-weight: 600;
  transition: all 0.3s ease;
}

.logout-btn:hover {
  background: var(--neon-orange);
  color: #fff;
  box-shadow: 0 0 15px var(--neon-orange-glow);
}

/* Main Content area */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  height: 100vh;
}

.top-bar {
  height: 60px;
  padding: 0 2rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(0, 242, 255, 0.1);
  background: rgba(5, 7, 10, 0.5);
}

@media (max-width: 600px) {
  .top-bar { padding: 0 1rem; }
}

.top-bar-left {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.top-bar-right {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.lang-switcher-btn {
  background: transparent;
  border: 1px solid rgba(0, 242, 255, 0.2);
  color: var(--text-secondary);
  padding: 0.4rem 0.8rem;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-family: var(--font-data);
  font-size: 0.8rem;
  transition: all 0.2s ease;
}

.lang-switcher-btn:hover {
  border-color: var(--neon-cyan);
  color: var(--neon-cyan);
  background: rgba(0, 242, 255, 0.05);
}

.lang-text {
  font-weight: 700;
}

.menu-toggle-btn {
  display: none;
  background: transparent;
  border: none;
  color: var(--neon-cyan);
  cursor: pointer;
  padding: 0.5rem;
}

@media (max-width: 1024px) {
  .menu-toggle-btn { display: flex; }
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  font-size: 0.9rem;
  letter-spacing: 2px;
}

@media (max-width: 768px) {
  .breadcrumb .text-secondary,
  .breadcrumb .separator {
    display: none;
  }
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 0.8rem;
}

@media (max-width: 480px) {
  .status-text { display: none; }
}

.pulse-dot {
  width: 8px;
  height: 8px;
  background-color: var(--neon-cyan);
  border-radius: 50%;
  box-shadow: 0 0 10px var(--neon-cyan);
  animation: pulse-cyan 2s infinite;
}

.content-container {
  flex: 1;
  padding: 2rem;
  overflow-y: auto;
}

@media (max-width: 768px) {
  .content-container { padding: 1.5rem 1rem; }
}

/* Transition */
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
