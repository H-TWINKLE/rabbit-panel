<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  Menu,
  Monitor,
  Box,
  Picture,
  Connection,
  Files,
  Cpu,
  User,
  SwitchButton,
  Moon,
  Sunny,
  Fold,
  Expand,
  FolderOpened,
  OfficeBuilding,
  Setting,
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'
import { useI18n } from '@/composables/useI18n'
import SystemMonitor from '@/components/common/SystemMonitor.vue'
import ChangePasswordDialog from '@/components/common/ChangePasswordDialog.vue'
import SettingsFloat from '@/components/common/SettingsFloat.vue'

const router = useRouter()
const authStore = useAuthStore()
const { theme, toggleTheme } = useTheme()
const { language, t, setLanguage } = useI18n()

// Logo based on theme
import mikuLight from '@/assets/miku-light.svg'
import mikuDark from '@/assets/miku-dark.svg'
const logoSrc = computed(() => theme.value === 'dark' ? mikuDark : mikuLight)

// Responsive state
const windowWidth = ref(window.innerWidth)
const isMobile = computed(() => windowWidth.value < 768)

function handleResize() {
  windowWidth.value = window.innerWidth
  // Close drawer when resizing to desktop
  if (!isMobile.value) {
    drawerVisible.value = false
  }
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})

// Sidebar collapse state (desktop)
const isCollapse = ref(false)

// Mobile drawer state
const drawerVisible = ref(false)

// Password dialog
const showPasswordDialog = ref(false)

// Current theme icon
const themeIcon = computed(() => theme.value === 'dark' ? Sunny : Moon)
const themeTooltip = computed(() => theme.value === 'dark' ? t('header.lightMode') : t('header.darkMode'))

// Language options
const languageOptions = [
  { label: '简体中文', value: 'zh-CN' as const },
  { label: 'English', value: 'en-US' as const },
]

// Navigation menu items
const menuItems = computed(() => [
  { index: '/', icon: Monitor, title: t('sideNav.dashboard') },
  { index: '/containers', icon: Box, title: t('sideNav.containers') },
  { index: '/images', icon: Picture, title: t('sideNav.images') },
  { index: '/networks', icon: Connection, title: t('sideNav.networks') },
  { index: '/volumes', icon: FolderOpened, title: t('sideNav.volumes') },
  { index: '/compose', icon: Files, title: t('sideNav.compose') },
  { index: '/nodes', icon: Cpu, title: t('sideNav.nodes') },
  { index: '/registry', icon: OfficeBuilding, title: t('sideNav.registry') },
  { index: '/docker-config', icon: Setting, title: t('sideNav.dockerConfig') },
])

// Handle menu select
function handleMenuSelect(index: string) {
  router.push(index)
  // Close drawer on mobile after selection
  if (isMobile.value) {
    drawerVisible.value = false
  }
}

// Handle logout
async function handleLogout() {
  try {
    await ElMessageBox.confirm(
      t('auth.logout') + '?',
      t('common.confirm'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning',
      }
    )
    
    await authStore.logout()
    ElMessage.success(t('auth.logoutSuccess'))
    router.push('/login')
  } catch {
    // User cancelled
  }
}

// Handle language change
function handleLanguageChange(lang: 'zh-CN' | 'en-US') {
  setLanguage(lang)
}

// Handle password change
function handleChangePassword() {
  showPasswordDialog.value = true
  // Close drawer on mobile
  if (isMobile.value) {
    drawerVisible.value = false
  }
}

function handlePasswordChangeSuccess() {
  showPasswordDialog.value = false
  ElMessage.success(t('auth.passwordChanged'))
}

// Open GitHub repository
function openGitHub() {
  window.open('https://github.com/reisen7/rabbit-panel', '_blank')
}

// Toggle mobile drawer
function toggleDrawer() {
  drawerVisible.value = !drawerVisible.value
}
</script>

<template>
  <el-container class="main-layout">
    <!-- Desktop Sidebar -->
    <el-aside v-if="!isMobile" :width="isCollapse ? '64px' : '200px'" class="main-sidebar">
      <div class="sidebar-header">
        <img :src="logoSrc" alt="Logo" class="logo" />
        <span v-show="!isCollapse" class="logo-text">Rabbit Panel</span>
      </div>
      
      <el-menu
        :default-active="$route.path"
        :collapse="isCollapse"
        :collapse-transition="false"
        class="sidebar-menu"
        @select="handleMenuSelect"
      >
        <el-menu-item
          v-for="item in menuItems"
          :key="item.index"
          :index="item.index"
        >
          <el-icon><component :is="item.icon" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </el-menu>
      
      <!-- Collapse button -->
      <div class="collapse-btn" @click="isCollapse = !isCollapse">
        <el-icon :size="20">
          <component :is="isCollapse ? Expand : Fold" />
        </el-icon>
      </div>
    </el-aside>

    <!-- Mobile Drawer -->
    <el-drawer
      v-model="drawerVisible"
      direction="ltr"
      :size="280"
      :with-header="false"
      class="mobile-drawer"
    >
      <div class="drawer-header">
        <img :src="logoSrc" alt="Logo" class="logo" />
        <span class="logo-text">Rabbit Panel</span>
      </div>
      
      <el-menu
        :default-active="$route.path"
        class="drawer-menu"
        @select="handleMenuSelect"
      >
        <el-menu-item
          v-for="item in menuItems"
          :key="item.index"
          :index="item.index"
        >
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.title }}</span>
        </el-menu-item>
      </el-menu>
      
      <!-- Drawer footer with user actions -->
      <div class="drawer-footer">
        <div class="drawer-user-info">
          <el-avatar :size="40" :icon="User" />
          <span class="drawer-username">{{ authStore.username }}</span>
        </div>
        
        <div class="drawer-actions">
          <el-button text @click="handleChangePassword">
            <el-icon><SwitchButton /></el-icon>
            {{ t('header.changePassword') }}
          </el-button>
          <el-button text type="danger" @click="handleLogout">
            <el-icon><SwitchButton /></el-icon>
            {{ t('nav.logout') }}
          </el-button>
        </div>
      </div>
    </el-drawer>

    <el-container>
      <!-- Header -->
      <el-header class="main-header">
        <div class="header-left">
          <!-- Mobile menu button -->
          <el-button
            v-if="isMobile"
            :icon="Menu"
            class="mobile-menu-btn"
            @click="toggleDrawer"
          />
          <SystemMonitor />
        </div>
        
        <div class="header-right">
          <!-- GitHub link -->
          <el-tooltip content="GitHub" placement="bottom">
            <el-button circle @click="openGitHub">
              <svg class="github-icon" viewBox="0 0 16 16" width="18" height="18" fill="currentColor">
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
              </svg>
            </el-button>
          </el-tooltip>
          
          <!-- Theme toggle -->
          <el-tooltip :content="themeTooltip" placement="bottom">
            <el-button
              :icon="themeIcon"
              circle
              @click="toggleTheme"
            />
          </el-tooltip>
          
          <!-- Language selector -->
          <el-dropdown @command="handleLanguageChange">
            <el-button circle>
              <span class="lang-icon">{{ language === 'zh-CN' ? '中' : 'En' }}</span>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  v-for="opt in languageOptions"
                  :key="opt.value"
                  :command="opt.value"
                  :class="{ 'is-active': language === opt.value }"
                >
                  {{ opt.label }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          
          <!-- User dropdown (desktop only) -->
          <el-dropdown v-if="!isMobile">
            <div class="user-info">
              <el-avatar :size="32" :icon="User" />
              <span class="username">{{ authStore.username }}</span>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="handleChangePassword">
                  <el-icon><SwitchButton /></el-icon>
                  {{ t('header.changePassword') }}
                </el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">
                  <el-icon><SwitchButton /></el-icon>
                  {{ t('nav.logout') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- Main content -->
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
  
  <!-- Password change dialog -->
  <ChangePasswordDialog
    v-model="showPasswordDialog"
    @success="handlePasswordChangeSuccess"
  />
  
  <!-- Settings float button -->
  <SettingsFloat />
</template>

<style scoped>
.main-layout {
  height: 100vh;
  overflow: hidden;
}

.main-sidebar {
  background-color: var(--rp-sidebar-bg);
  border-right: 1px solid var(--rp-border-color);
  display: flex;
  flex-direction: column;
  transition: width 0.3s;
}

.sidebar-header {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 16px;
  border-bottom: 1px solid var(--rp-border-color);
}

.logo {
  width: 32px;
  height: 32px;
}

.logo-text {
  margin-left: 12px;
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  white-space: nowrap;
}

.sidebar-menu {
  flex: 1;
  border-right: none;
  overflow-y: auto;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 100%;
}

.collapse-btn {
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-top: 1px solid var(--rp-border-color);
  color: var(--el-text-color-secondary);
  transition: color 0.3s, background-color 0.3s;
}

.collapse-btn:hover {
  color: var(--el-color-primary);
  background-color: var(--rp-hover-bg);
}

.main-header {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  background-color: var(--rp-header-bg);
  border-bottom: 1px solid var(--rp-border-color);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-right > * {
  margin: 0 !important;
}

.lang-icon {
  font-size: 12px;
  font-weight: 600;
}

.github-icon {
  display: block;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.user-info:hover {
  background-color: var(--rp-hover-bg);
}

.username {
  font-size: 14px;
  color: var(--el-text-color-primary);
}

.main-content {
  background-color: var(--rp-bg-color);
  overflow-y: auto;
  padding: 20px;
}

/* Active language item */
:deep(.el-dropdown-menu__item.is-active) {
  color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-9);
}

/* Mobile menu button */
.mobile-menu-btn {
  margin-right: 8px;
}

/* Mobile Drawer Styles */
.mobile-drawer :deep(.el-drawer__body) {
  padding: 0;
  display: flex;
  flex-direction: column;
}

.drawer-header {
  height: 60px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  border-bottom: 1px solid var(--rp-border-color);
  background-color: var(--rp-header-bg);
}

.drawer-menu {
  flex: 1;
  border-right: none;
  overflow-y: auto;
}

.drawer-menu .el-menu-item {
  height: 50px;
  line-height: 50px;
  font-size: 15px;
}

.drawer-menu .el-menu-item .el-icon {
  font-size: 18px;
  margin-right: 12px;
}

.drawer-footer {
  padding: 16px 20px;
  border-top: 1px solid var(--rp-border-color);
  background-color: var(--rp-header-bg);
}

.drawer-user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.drawer-username {
  font-size: 16px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.drawer-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.drawer-actions .el-button {
  justify-content: flex-start;
  width: 100%;
}

/* Mobile responsive styles */
@media (max-width: 767px) {
  .main-header {
    padding: 0 12px;
  }
  
  .main-content {
    padding: 12px;
  }
  
  .header-right {
    gap: 8px;
  }
}
</style>
