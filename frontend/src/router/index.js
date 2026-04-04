import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('../views/Layout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue'),
      },
      {
        path: 'docker',
        name: 'Docker',
        component: () => import('../views/Docker.vue'),
      },
      {
        path: 'processes',
        name: 'Processes',
        component: () => import('../views/Processes.vue'),
      },
      {
        path: 'firewall',
        name: 'Firewall',
        component: () => import('../views/Firewall.vue'),
      },
      {
        path: 'sites',
        name: 'Sites',
        component: () => import('../views/Sites.vue'),
      },
      {
        path: 'audit',
        name: 'AuditLogs',
        component: () => import('../views/AuditLogs.vue'),
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else {
    next()
  }
})

export default router
