import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

/**
 * Route meta interface for type safety
 */
declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
    title?: string
  }
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/Login.vue'),
      meta: {
        requiresAuth: false,
        title: '登录',
      },
    },
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      meta: {
        requiresAuth: true,
      },
      children: [
        {
          path: '',
          name: 'dashboard',
          component: () => import('@/views/Dashboard.vue'),
          meta: {
            title: '仪表板',
          },
        },
        {
          path: 'containers',
          name: 'containers',
          component: () => import('@/views/Containers.vue'),
          meta: {
            title: '容器管理',
          },
        },
        {
          path: 'images',
          name: 'images',
          component: () => import('@/views/Images.vue'),
          meta: {
            title: '镜像管理',
          },
        },
        {
          path: 'networks',
          name: 'networks',
          component: () => import('@/views/Networks.vue'),
          meta: {
            title: '网络管理',
          },
        },
        {
          path: 'compose',
          name: 'compose',
          component: () => import('@/views/Compose.vue'),
          meta: {
            title: 'Compose',
          },
        },
        {
          path: 'nodes',
          name: 'nodes',
          component: () => import('@/views/Nodes.vue'),
          meta: {
            title: '节点管理',
          },
        },
        {
          path: 'volumes',
          name: 'volumes',
          component: () => import('@/views/Volumes.vue'),
          meta: {
            title: '存储卷管理',
          },
        },
        {
          path: 'registry',
          name: 'registry',
          component: () => import('@/views/Registry.vue'),
          meta: {
            title: '仓库管理',
          },
        },
        {
          path: 'docker-config',
          name: 'dockerConfig',
          component: () => import('@/views/DockerConfig.vue'),
          meta: {
            title: 'Docker 配置',
          },
        },
        {
          path: 'agent',
          name: 'agent',
          component: () => import('@/views/AgentChat.vue'),
          meta: {
            title: '智能助手',
          },
        },
        {
          path: 'settings/agent',
          name: 'agentSettings',
          component: () => import('@/views/AgentSettings.vue'),
          meta: {
            title: '智能体配置',
          },
        },
      ],
    },
  ],
})

/**
 * Navigation guard to check authentication status
 * Redirects to login page if user is not authenticated
 */
router.beforeEach(async (to: RouteLocationNormalized, _from: RouteLocationNormalized) => {
  const authStore = useAuthStore()

  // Check if route or any parent requires authentication
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth !== false)

  // If route doesn't require auth, allow access
  if (!requiresAuth || to.meta.requiresAuth === false) {
    // If user is already authenticated and trying to access login, redirect to dashboard
    if (to.name === 'login' && authStore.isAuthenticated) {
      return { name: 'dashboard' }
    }
    return true
  }

  // Route requires authentication
  // First check if we have a token
  if (!authStore.isAuthenticated) {
    // No token, redirect to login
    return {
      name: 'login',
      query: { redirect: to.fullPath },
    }
  }

  // We have a token, verify it's still valid
  // Only check on initial load or if username is not set
  if (!authStore.username) {
    const isValid = await authStore.checkAuth()
    if (!isValid) {
      // Token is invalid, redirect to login
      return {
        name: 'login',
        query: { redirect: to.fullPath },
      }
    }
  }

  // User is authenticated, allow access
  return true
})

/**
 * Update document title after navigation
 */
router.afterEach((to: RouteLocationNormalized) => {
  const title = to.meta.title
  document.title = title ? `${title} - Rabbit Panel` : 'Rabbit Panel'
})

export default router
