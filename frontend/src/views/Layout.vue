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
  Cpu,
  Menu,
  X,
  BarChart3,
  Languages,
  FolderClosed,
  Package,
  Activity,
  Server
} from 'lucide-vue-next'
import { useAuthStore } from '../stores/auth'
import { useToastStore } from '../stores/toast'
import { useSettingsStore } from '../stores/settings'
import { useSystemStore } from '../stores/system'
import api from '../api'
import InteractiveImage from '../components/InteractiveImage.vue'
import BrandLogo from '../components/BrandLogo.vue'

const { t, locale } = useI18n()
const authStore = useAuthStore()
const settingsStore = useSettingsStore()
const systemStore = useSystemStore()
const toast = useToastStore()
const router = useRouter()
const route = useRoute()

const isSidebarOpen = ref(window.innerWidth > 1024)
const hostOs = ref('')
let statsInterval = null

const updateSidebarState = () => {
  if (window.innerWidth <= 1024) {
    isSidebarOpen.value = false
  } else {
    isSidebarOpen.value = true
  }
}

const fetchHostInfo = async () => {
  try {
    const res = await api.get('/packages/overview')
    if (res.data?.data?.os) {
      const os = res.data.data.os
      hostOs.value = `${os.name || 'Linux'} ${os.arch ? '(' + os.arch + ')' : ''}`.trim()
    }
  } catch (err) {
    // Fallback quietly if endpoint is warming up
    console.debug('Host overview warmup:', err)
  }
}

onMounted(async () => {
  window.addEventListener('resize', updateSidebarState)
  
  // Initial stats & host OS detection
  fetchHostInfo()
  if (authStore.token) {
    systemStore.fetchAllStats().catch(() => {})
    statsInterval = setInterval(() => {
      systemStore.fetchAllStats().catch(() => {})
    }, 10000)
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', updateSidebarState)
  if (statsInterval) {
    clearInterval(statsInterval)
    statsInterval = null
  }
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
  
  const routeMap = {
    'dashboard': 'nav.dashboard',
    'docker': 'nav.docker',
    'terminal': 'nav.terminal',
    'processes': 'nav.processes',
    'firewall': 'nav.firewall',
    'sites': 'nav.sites',
    'packages': 'nav.packages',
    'users': 'nav.users',
    'auditlogs': 'nav.audit_logs',
    'profile': 'nav.profile',
    'reports': 'nav.reports',
    'files': 'nav.files',
    'login': 'nav.login'
  }
  
  const key = routeMap[name] || `nav.${name}`
  const translated = t(key)
  return translated !== key ? translated : name.toUpperCase()
})

const menuItems = computed(() => [
  { name: t('nav.dashboard'), path: '/', icon: LayoutDashboard },
  { name: t('nav.docker'), path: '/docker', icon: Box },
  { name: t('nav.terminal'), path: '/terminal', icon: Terminal },
  { name: t('nav.processes'), path: '/processes', icon: Cpu },
  { name: t('nav.firewall'), path: '/firewall', icon: ShieldCheck },
  { name: t('nav.sites'), path: '/sites', icon: Globe },
  { name: t('nav.packages'), path: '/packages', icon: Package },
  { name: t('nav.users'), path: '/users', icon: Users },
  { name: t('nav.audit_logs'), path: '/audit', icon: History },
  { name: t('nav.reports'), path: '/reports', icon: BarChart3 },
  { name: t('nav.files'), path: '/files', icon: FolderClosed },
])

const toggleLanguage = (lang) => {
  const target = lang || (locale.value === 'en' ? 'ar' : 'en')
  if (locale.value === target) return
  locale.value = target
  settingsStore.setLanguage(target)
  toast.info(target === 'ar' ? 'تم تغيير لغة النظام إلى العربية' : 'System language set to English')
}

const handleLogout = () => {
  toast.info(t('common.connection_terminated'))
  authStore.logout()
}

const cpuColorClass = computed(() => {
  const u = systemStore.cpu.usage || 0
  if (u >= 85) return 'stat-danger'
  if (u >= 60) return 'stat-warning'
  return 'stat-optimal'
})

const ramColorClass = computed(() => {
  const u = systemStore.ram.usage || 0
  if (u >= 85) return 'stat-danger'
  if (u >= 60) return 'stat-warning'
  return 'stat-optimal'
})
</script>

<template>
  <div class="layout-wrapper">
    <!-- Mobile Sidebar Backdrop -->
    <div 
      v-if="isSidebarOpen" 
      class="sidebar-overlay" 
      @click="isSidebarOpen = false"
    ></div>

    <!-- Stealth Obsidian Sidebar -->
    <aside :class="['sidebar', { 'closed': !isSidebarOpen }]">
      <!-- Brand & Tactical Logo -->
      <div class="brand-container">
        <BrandLogo size="md" />
        <button class="close-sidebar-btn" @click="isSidebarOpen = false" :title="$t('common.close')">
          <X :size="20" />
        </button>
      </div>

      <!-- Navigation Links -->
      <nav class="nav-menu">
        <router-link 
          v-for="item in menuItems" 
          :key="item.path" 
          :to="item.path"
          class="nav-item"
          :class="{ 'active': route.path === item.path }"
          @click="handleNavClick"
        >
          <component :is="item.icon" :size="18" class="nav-icon" />
          <span class="nav-text">{{ item.name }}</span>
        </router-link>
      </nav>

      <!-- Sidebar User & Session Footer -->
      <div class="sidebar-footer">
        <div class="user-card">
          <div class="user-avatar-box">
            <div class="avatar-wrap" v-if="authStore.user?.image_path && authStore.user.image_path.includes('/uploads/')">
              <InteractiveImage :src="authStore.user.image_path" customClass="user-avatar" :alt="$t('common.avatar')" />
            </div>
            <div class="avatar-fallback" v-else>
              <User :size="16" />
            </div>
            <span class="user-online-dot"></span>
          </div>
          <div class="user-info">
            <router-link to="/profile" class="user-name-link">
              <span class="user-display-name">{{ authStore.user?.username || $t('common.admin') }}</span>
            </router-link>
            <span class="user-role-tag">{{ authStore.user?.role?.toUpperCase() || 'OPERATOR' }}</span>
          </div>
        </div>

        <button @click="handleLogout" class="tactical-logout-btn">
          <LogOut :size="16" />
          <span>{{ $t('nav.logout') }}</span>
        </button>
      </div>
    </aside>

    <!-- Main Content Area -->
    <main class="main-content">
      <!-- Stealth Top Bar -->
      <header class="top-bar">
        <!-- Left: Mobile Toggle & Breadcrumbs -->
        <div class="top-bar-left">
          <button class="menu-toggle-btn" @click="toggleSidebar" aria-label="Toggle Navigation">
            <Menu :size="20" />
          </button>
          <div class="breadcrumb-trail">
            <Server :size="15" class="breadcrumb-icon" />
            <span class="breadcrumb-root">{{ $t('common.system') }}</span>
            <span class="breadcrumb-sep">/</span>
            <span class="breadcrumb-current">{{ localizedRouteName }}</span>
          </div>
        </div>

        <!-- Right: Telemetry Capsule & Controls -->
        <div class="top-bar-right">
          <!-- System Telemetry Capsule -->
          <div class="telemetry-capsule">
            <div class="capsule-section">
              <span class="beacon-online"></span>
              <span class="capsule-label glow-cyan">{{ $t('common.grid_online') }}</span>
            </div>

            <div class="capsule-divider" v-if="hostOs"></div>
            <div class="capsule-section distro-badge" v-if="hostOs">
              <span class="font-data distro-text">{{ hostOs }}</span>
            </div>

            <div class="capsule-divider"></div>
            <div class="capsule-section stat-pill font-data">
              <span class="pill-title">CPU</span>
              <span class="pill-value" :class="cpuColorClass">{{ systemStore.cpu.usage }}%</span>
            </div>

            <div class="capsule-divider"></div>
            <div class="capsule-section stat-pill font-data">
              <span class="pill-title">RAM</span>
              <span class="pill-value" :class="ramColorClass">{{ systemStore.ram.usage }}%</span>
            </div>
          </div>

          <!-- Language Selector Switch -->
          <div class="lang-switch-capsule">
            <button 
              class="lang-pill-btn" 
              :class="{ 'active': locale === 'en' }" 
              @click="toggleLanguage('en')"
            >
              EN
            </button>
            <button 
              class="lang-pill-btn" 
              :class="{ 'active': locale === 'ar' }" 
              @click="toggleLanguage('ar')"
            >
              عربي
            </button>
          </div>
        </div>
      </header>

      <!-- Routed Page Container -->
      <div class="content-container">
        <router-view v-slot="{ Component }">
          <transition name="view-fade" mode="out-in">
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
  width: 100vw;
  background-color: var(--bg-black);
  overflow: hidden;
  color: var(--text-primary);
}

/* ==================== SIDEBAR ==================== */
.sidebar {
  width: 270px;
  background: var(--bg-card);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border-right: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  transition: transform 0.28s cubic-bezier(0.16, 1, 0.3, 1), width 0.28s ease;
  z-index: 1000;
  height: 100vh;
  flex-shrink: 0;
  box-shadow: 4px 0 24px rgba(0, 0, 0, 0.5);
}

[dir="rtl"] .sidebar {
  border-right: none;
  border-left: 1px solid var(--border-subtle);
  box-shadow: -4px 0 24px rgba(0, 0, 0, 0.5);
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
}

.sidebar-overlay {
  position: fixed;
  inset: 0;
  background: rgba(4, 7, 12, 0.75);
  backdrop-filter: blur(6px);
  z-index: 999;
}

/* Brand Header */
.brand-container {
  padding: 1.25rem 1.2rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-subtle);
  background: rgba(220, 38, 38, 0.02);
}

.brand-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.brand-symbol {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  background: rgba(220, 38, 38, 0.12);
  border: 1px solid var(--border-crimson);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--rdr-crimson);
}

.brand-icon {
  color: var(--rdr-crimson);
}

.brand-meta {
  display: flex;
  flex-direction: column;
}

.brand-title {
  font-family: var(--font-header);
  font-size: 1.1rem;
  font-weight: 700;
  letter-spacing: 2px;
  color: #fff;
  line-height: 1.2;
}

[dir="rtl"] .brand-title {
  font-family: var(--font-arabic);
  letter-spacing: 0px !important;
}

.brand-badge {
  font-family: var(--font-data);
  font-size: 0.65rem;
  letter-spacing: 0.5px;
  color: var(--rdr-amber);
  font-weight: 600;
}

.close-sidebar-btn {
  display: none;
  background: transparent;
  border: 1px solid var(--border-subtle);
  border-radius: 3px;
  color: var(--text-secondary);
  padding: 0.35rem;
  cursor: pointer;
  transition: all 0.15s;
}

.close-sidebar-btn:hover {
  color: var(--rdr-crimson);
  border-color: var(--rdr-crimson);
}

@media (max-width: 1024px) {
  .close-sidebar-btn {
    display: flex;
    align-items: center;
    justify-content: center;
  }
}

/* Nav Menu */
.nav-menu {
  flex: 1;
  padding: 1rem 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 0.85rem;
  text-decoration: none;
  color: var(--text-secondary);
  border-radius: 3px;
  font-size: 0.86rem;
  letter-spacing: 0.3px;
  font-weight: 500;
  position: relative;
  transition: all 0.15s ease;
  border: 1px solid transparent;
  border-inline-start: 3px solid transparent;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.04);
  color: #ffffff;
  border-color: rgba(255, 255, 255, 0.06);
}

.nav-item.active {
  background: rgba(220, 38, 38, 0.12);
  color: #ffffff;
  border-color: rgba(220, 38, 38, 0.25);
  border-inline-start: 3px solid var(--rdr-crimson);
  font-weight: 600;
}

.nav-icon {
  flex-shrink: 0;
}

.nav-item.active .nav-icon {
  color: var(--rdr-crimson);
}

.nav-text {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Sidebar Footer */
.sidebar-footer {
  padding: 1.1rem;
  border-top: 1px solid var(--border-subtle);
  background: rgba(5, 8, 14, 0.6);
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.user-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.6rem 0.8rem;
  background: #161822;
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
}

.user-avatar-box {
  position: relative;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
}

.avatar-wrap {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  overflow: hidden;
  border: 1px solid var(--border-crimson);
}

.user-avatar {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-fallback {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: rgba(220, 38, 38, 0.12);
  border: 1px solid var(--border-crimson);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--rdr-crimson);
}

.user-online-dot {
  position: absolute;
  bottom: -1px;
  right: -1px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--linux-green);
  border: 2px solid var(--bg-black);
}

.user-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.user-name-link {
  text-decoration: none;
  color: var(--text-primary);
  font-weight: 600;
  font-size: 0.85rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: color 0.15s;
}

.user-name-link:hover {
  color: var(--rdr-crimson);
}

.user-role-tag {
  font-family: var(--font-data);
  font-size: 0.62rem;
  color: var(--rdr-amber);
  font-weight: 600;
  letter-spacing: 0.5px;
}

.tactical-logout-btn {
  background: rgba(220, 38, 38, 0.08);
  border: 1px solid var(--border-crimson);
  color: #fca5a5;
  padding: 0.5rem;
  border-radius: 3px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  font-family: var(--font-header);
  font-weight: 600;
  font-size: 0.8rem;
  letter-spacing: 0.5px;
  transition: all 0.15s ease;
}

[dir="rtl"] .tactical-logout-btn {
  font-family: var(--font-arabic);
  letter-spacing: 0px !important;
}

.tactical-logout-btn:hover {
  background: var(--rdr-crimson);
  color: #ffffff;
  border-color: var(--rdr-crimson-hover);
}

/* ==================== MAIN CONTENT AREA ==================== */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  position: relative;
}

/* Top Bar */
.top-bar {
  height: 56px;
  padding: 0 1.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #111319;
  border-bottom: 1px solid var(--border-subtle);
  z-index: 10;
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .top-bar {
    padding: 0 0.85rem;
    height: 52px;
  }
}

.top-bar-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.menu-toggle-btn {
  display: none;
  background: rgba(220, 38, 38, 0.08);
  border: 1px solid var(--border-crimson);
  color: var(--rdr-crimson);
  border-radius: 3px;
  padding: 0.4rem;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.menu-toggle-btn:hover {
  background: var(--rdr-crimson);
  color: #ffffff;
}

@media (max-width: 1024px) {
  .menu-toggle-btn {
    display: flex;
  }
}

/* Breadcrumbs */
.breadcrumb-trail {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.88rem;
  letter-spacing: 0.5px;
}

.breadcrumb-icon {
  color: var(--rdr-crimson);
}

.breadcrumb-root {
  color: var(--text-muted);
  font-weight: 500;
  font-family: var(--font-data);
}

.breadcrumb-sep {
  color: var(--border-subtle);
  font-family: var(--font-data);
}

.breadcrumb-current {
  color: #ffffff;
  font-weight: 700;
}

[dir="rtl"] .breadcrumb-trail {
  letter-spacing: 0px !important;
}

/* Top Bar Right */
.top-bar-right {
  display: flex;
  align-items: center;
  gap: 0.85rem;
}

/* Telemetry Capsule */
.telemetry-capsule {
  display: flex;
  align-items: center;
  background: #181b24;
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
  padding: 0.3rem 0.75rem;
  gap: 0.65rem;
}

.capsule-section {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.76rem;
}

.capsule-label {
  font-size: 0.74rem;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: 0.5px;
}

.capsule-divider {
  width: 1px;
  height: 12px;
  background: var(--border-subtle);
}

.distro-badge {
  color: var(--text-secondary);
}

.distro-text {
  font-size: 0.72rem;
  letter-spacing: 0.5px;
}

.stat-pill {
  font-size: 0.74rem;
  display: flex;
  gap: 0.3rem;
}

.pill-title {
  color: var(--text-muted);
}

.pill-value {
  font-weight: 600;
}

.stat-optimal {
  color: var(--linux-green);
}

.stat-warning {
  color: var(--rdr-amber);
}

.stat-danger {
  color: var(--rdr-crimson);
}

@media (max-width: 900px) {
  .distro-badge,
  .capsule-divider:first-of-type {
    display: none;
  }
}

@media (max-width: 680px) {
  .telemetry-capsule {
    display: none;
  }
}

/* Language Segmented Toggle */
.lang-switch-capsule {
  display: flex;
  align-items: center;
  background: #181b24;
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
  padding: 2px;
}

.lang-pill-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-family: var(--font-data);
  font-size: 0.72rem;
  font-weight: 600;
  padding: 0.25rem 0.6rem;
  border-radius: 3px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.lang-pill-btn:hover {
  color: var(--text-primary);
}

.lang-pill-btn.active {
  background: var(--rdr-crimson);
  color: #ffffff;
}

/* Routed View Container */
.content-container {
  flex: 1;
  padding: 1.8rem;
  overflow-y: auto;
  overflow-x: hidden;
}

@media (max-width: 768px) {
  .content-container {
    padding: 1rem 0.85rem;
  }
}

/* Page Transition */
.view-fade-enter-active,
.view-fade-leave-active {
  transition: opacity 0.2s cubic-bezier(0.16, 1, 0.3, 1), transform 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.view-fade-enter-from {
  opacity: 0;
  transform: translateY(6px);
}

.view-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
