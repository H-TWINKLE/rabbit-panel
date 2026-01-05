<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { Setting } from '@element-plus/icons-vue'

const SETTINGS_KEY = 'rabbit_panel_settings'

// 设置面板显示状态
const showPanel = ref(false)

// 设置项
const settings = ref({
  fontSize: 14,
  sidebarWidth: 200,
  primaryColor: '#409EFF',
})

// 预设主题色
const colorOptions = [
  { label: '默认蓝', value: '#409EFF' },
  { label: '翠绿', value: '#67C23A' },
  { label: '橙色', value: '#E6A23C' },
  { label: '红色', value: '#F56C6C' },
  { label: '紫色', value: '#9B59B6' },
  { label: '青色', value: '#00BCD4' },
]

// 应用设置
function applySettings() {
  document.documentElement.style.setProperty('--rp-font-size', `${settings.value.fontSize}px`)
  document.documentElement.style.setProperty('--rp-sidebar-width', `${settings.value.sidebarWidth}px`)
  document.documentElement.style.setProperty('--el-color-primary', settings.value.primaryColor)
  
  // 保存到 localStorage
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(settings.value))
}

// 重置设置
function resetSettings() {
  settings.value = {
    fontSize: 14,
    sidebarWidth: 200,
    primaryColor: '#409EFF',
  }
  applySettings()
}

// 确认并刷新页面
function confirmAndRefresh() {
  applySettings()
  window.location.reload()
}

// 监听设置变化
watch(settings, () => {
  applySettings()
}, { deep: true })

// 初始化加载设置
onMounted(() => {
  const saved = localStorage.getItem(SETTINGS_KEY)
  if (saved) {
    try {
      const parsed = JSON.parse(saved)
      settings.value = { ...settings.value, ...parsed }
    } catch (e) {
      console.error('Failed to parse settings', e)
    }
  }
  applySettings()
})
</script>

<template>
  <div class="settings-float">
    <!-- 悬浮按钮 -->
    <el-button
      class="float-btn"
      type="primary"
      circle
      size="large"
      :icon="Setting"
      @click="showPanel = !showPanel"
    />

    <!-- 设置面板 -->
    <transition name="slide">
      <div v-show="showPanel" class="settings-panel">
        <div class="panel-header">
          <span>页面设置</span>
          <el-button text size="small" @click="showPanel = false">×</el-button>
        </div>
        
        <div class="panel-content">
          <!-- 字体大小 -->
          <div class="setting-item">
            <label>字体大小</label>
            <div class="setting-control">
              <el-slider
                v-model="settings.fontSize"
                :min="12"
                :max="18"
                :step="1"
                :show-tooltip="false"
              />
              <span class="value">{{ settings.fontSize }}px</span>
            </div>
          </div>

          <!-- 侧边栏宽度 -->
          <div class="setting-item">
            <label>侧边栏宽度</label>
            <div class="setting-control">
              <el-slider
                v-model="settings.sidebarWidth"
                :min="180"
                :max="280"
                :step="10"
                :show-tooltip="false"
              />
              <span class="value">{{ settings.sidebarWidth }}px</span>
            </div>
          </div>

          <!-- 主题色 -->
          <div class="setting-item">
            <label>主题色</label>
            <div class="color-options">
              <div
                v-for="color in colorOptions"
                :key="color.value"
                class="color-item"
                :class="{ active: settings.primaryColor === color.value }"
                :style="{ backgroundColor: color.value }"
                :title="color.label"
                @click="settings.primaryColor = color.value"
              />
            </div>
          </div>

          <!-- 重置按钮 -->
          <div class="setting-item buttons">
            <el-button size="small" @click="resetSettings">恢复默认</el-button>
            <el-button size="small" type="primary" @click="confirmAndRefresh">确认刷新</el-button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.settings-float {
  position: fixed;
  right: 20px;
  bottom: 20px;
  z-index: 2000;
}

.float-btn {
  width: 50px;
  height: 50px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.settings-panel {
  position: absolute;
  right: 0;
  bottom: 60px;
  width: 280px;
  background: var(--el-bg-color);
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  border: 1px solid var(--el-border-color);
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color);
  font-weight: 600;
}

.panel-content {
  padding: 16px;
}

.setting-item {
  margin-bottom: 16px;
}

.setting-item:last-child {
  margin-bottom: 0;
}

.setting-item label {
  display: block;
  font-size: 13px;
  color: var(--el-text-color-regular);
  margin-bottom: 8px;
}

.setting-control {
  display: flex;
  align-items: center;
  gap: 12px;
}

.setting-control .el-slider {
  flex: 1;
}

.setting-control .value {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  min-width: 45px;
  text-align: right;
}

.color-options {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.color-item {
  width: 28px;
  height: 28px;
  border-radius: 4px;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  border: 2px solid transparent;
}

.color-item:hover {
  transform: scale(1.1);
}

.color-item.active {
  border-color: var(--el-text-color-primary);
  box-shadow: 0 0 0 2px var(--el-bg-color);
}

.setting-item.buttons {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

/* 动画 */
.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s ease;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>
