<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useSystemStore } from '@/stores/system'
import { useI18n } from '@/composables/useI18n'

const systemStore = useSystemStore()
const { t } = useI18n()

// Start polling on mount
onMounted(() => {
  systemStore.startPolling()
})

// Stop polling on unmount
onUnmounted(() => {
  systemStore.stopPolling()
})

// Get progress bar color based on usage percentage
function getProgressColor(percent: number): string {
  if (percent >= 90) return '#F56C6C'
  if (percent >= 70) return '#E6A23C'
  return '#67C23A'
}
</script>

<template>
  <div class="system-monitor">
    <!-- CPU -->
    <div class="monitor-item">
      <span class="monitor-label">{{ t('system.cpu') }}</span>
      <el-progress
        :percentage="systemStore.stats.cpu"
        :stroke-width="8"
        :color="getProgressColor(systemStore.stats.cpu)"
        :show-text="false"
        class="monitor-progress"
      />
      <span class="monitor-value">{{ systemStore.stats.cpu.toFixed(1) }}%</span>
    </div>
    
    <!-- Memory -->
    <div class="monitor-item">
      <span class="monitor-label">{{ t('system.memory') }}</span>
      <el-progress
        :percentage="systemStore.stats.memory"
        :stroke-width="8"
        :color="getProgressColor(systemStore.stats.memory)"
        :show-text="false"
        class="monitor-progress"
      />
      <span class="monitor-value">{{ systemStore.stats.memory.toFixed(1) }}%</span>
    </div>
    
    <!-- Disk -->
    <div class="monitor-item">
      <span class="monitor-label">{{ t('system.disk') }}</span>
      <el-progress
        :percentage="systemStore.stats.disk"
        :stroke-width="8"
        :color="getProgressColor(systemStore.stats.disk)"
        :show-text="false"
        class="monitor-progress"
      />
      <span class="monitor-value">{{ systemStore.stats.disk.toFixed(1) }}%</span>
    </div>
    
    <!-- Server Time -->
    <div class="monitor-item time-item">
      <span class="monitor-label">{{ t('system.serverTime') }}</span>
      <span class="monitor-time">{{ systemStore.stats.time || '--:--:--' }}</span>
    </div>
  </div>
</template>

<style scoped>
.system-monitor {
  display: flex;
  align-items: center;
  gap: 24px;
}

.monitor-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.monitor-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  white-space: nowrap;
}

.monitor-progress {
  width: 80px;
}

.monitor-value {
  font-size: 12px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  min-width: 45px;
  text-align: right;
}

.time-item {
  margin-left: 8px;
}

.monitor-time {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  font-family: 'Courier New', monospace;
}

/* Responsive: hide on smaller screens */
@media (max-width: 1200px) {
  .time-item {
    display: none;
  }
}

@media (max-width: 900px) {
  .system-monitor {
    gap: 16px;
  }
  
  .monitor-progress {
    width: 60px;
  }
}

@media (max-width: 768px) {
  .system-monitor {
    display: none;
  }
}
</style>
