<template>
  <el-dialog
    v-model="visible"
    :title="t('settings.pageSettings')"
    width="400px"
    :close-on-click-modal="true"
  >
    <div class="settings-content">
      <!-- Font Size -->
      <div class="setting-item">
        <label>{{ t('settings.fontSize') }}</label>
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

      <!-- Sidebar Width -->
      <div class="setting-item">
        <label>{{ t('settings.sidebarWidth') }}</label>
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

      <!-- Theme Color -->
      <div class="setting-item">
        <label>{{ t('settings.themeColor') }}</label>
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
    </div>

    <template #footer>
      <el-button @click="resetSettings">{{ t('settings.reset') }}</el-button>
      <el-button type="primary" @click="confirmAndRefresh">{{ t('settings.confirmRefresh') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const SETTINGS_KEY = 'rabbit_panel_settings'

const settings = ref({
  fontSize: 14,
  sidebarWidth: 200,
  primaryColor: '#409EFF',
})

const colorOptions = [
  { label: '默认蓝', value: '#409EFF' },
  { label: '翠绿', value: '#67C23A' },
  { label: '橙色', value: '#E6A23C' },
  { label: '红色', value: '#F56C6C' },
  { label: '紫色', value: '#9B59B6' },
  { label: '青色', value: '#00BCD4' },
]

function applySettings() {
  document.documentElement.style.setProperty('--rp-font-size', `${settings.value.fontSize}px`)
  document.documentElement.style.setProperty('--rp-sidebar-width', `${settings.value.sidebarWidth}px`)
  document.documentElement.style.setProperty('--el-color-primary', settings.value.primaryColor)
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(settings.value))
}

function resetSettings() {
  settings.value = {
    fontSize: 14,
    sidebarWidth: 200,
    primaryColor: '#409EFF',
  }
  applySettings()
}

function confirmAndRefresh() {
  applySettings()
  window.location.reload()
}

watch(settings, () => {
  applySettings()
}, { deep: true })

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

<style scoped>
.settings-content {
  padding: 8px 0;
}

.setting-item {
  margin-bottom: 20px;
}

.setting-item:last-child {
  margin-bottom: 0;
}

.setting-item label {
  display: block;
  font-size: 14px;
  color: var(--el-text-color-regular);
  margin-bottom: 10px;
}

.setting-control {
  display: flex;
  align-items: center;
  gap: 16px;
}

.setting-control .el-slider {
  flex: 1;
}

.setting-control .value {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  min-width: 50px;
  text-align: right;
}

.color-options {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.color-item {
  width: 32px;
  height: 32px;
  border-radius: 6px;
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
</style>
