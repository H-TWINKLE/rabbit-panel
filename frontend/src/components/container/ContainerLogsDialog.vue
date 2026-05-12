<template>
  <el-dialog
    v-model="dialogVisible"
    :title="`${t('container.logs')} - ${containerName}`"
    width="900px"
    :close-on-click-modal="false"
    @open="handleOpen"
    @close="handleClose"
  >
    <div class="logs-dialog">
      <!-- Search bar -->
      <div class="logs-toolbar">
        <el-input
          v-model="searchQuery"
          :placeholder="t('container.searchLogs')"
          clearable
          style="width: 240px"
          @input="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <div class="toolbar-right">
          <el-switch
            v-model="showAllLogs"
            active-text="全部日志"
            inactive-text="实时日志"
            @change="handleModeChange"
          />
          <el-checkbox v-model="autoScroll" :disabled="showAllLogs">Auto-scroll</el-checkbox>
          <el-button size="small" @click="clearLogs">
            <el-icon><Delete /></el-icon>
            Clear
          </el-button>
        </div>
      </div>

      <!-- Logs content -->
      <div ref="logsRef" class="logs-content">
        <div v-if="filteredLogs.length === 0" class="no-logs">
          {{ t('container.noLogs') }}
        </div>
        <div
          v-for="(log, index) in filteredLogs"
          :key="index"
          class="log-line"
          v-html="highlightSearch(log)"
        />
      </div>

      <!-- Status bar -->
      <div class="logs-status">
        <span v-if="showAllLogs" class="status connected">
          <el-icon><CircleCheck /></el-icon>
          History Loaded
        </span>
        <span v-else-if="isConnected" class="status connected">
          <el-icon><CircleCheck /></el-icon>
          Connected
        </span>
        <span v-else class="status disconnected">
          <el-icon><CircleClose /></el-icon>
          Disconnected
        </span>
        <span class="log-count">{{ logs.length }} lines</span>
      </div>
    </div>

    <template #footer>
      <el-button @click="handleClose">{{ t('common.close') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted } from 'vue'
import { Search, Delete, CircleCheck, CircleClose } from '@element-plus/icons-vue'
import { useI18n } from '@/composables/useI18n'
import { containerApi } from '@/api/containers'

const { t } = useI18n()

const props = defineProps<{
  visible: boolean
  containerId: string
  containerName: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
})

const logs = ref<string[]>([])
const searchQuery = ref('')
const autoScroll = ref(true)
const isConnected = ref(false)
const logsRef = ref<HTMLElement>()
const showAllLogs = ref(false)

let eventSource: EventSource | null = null

// Filter logs by search query
const filteredLogs = computed(() => {
  if (!searchQuery.value) {
    return logs.value
  }
  const query = searchQuery.value.toLowerCase()
  return logs.value.filter((log) => log.toLowerCase().includes(query))
})

// Highlight search matches
function highlightSearch(text: string): string {
  if (!searchQuery.value) {
    return escapeHtml(text)
  }
  
  const query = searchQuery.value
  const escapedText = escapeHtml(text)
  const escapedQuery = escapeHtml(query)
  
  // Case-insensitive replace with highlight
  const regex = new RegExp(`(${escapeRegex(escapedQuery)})`, 'gi')
  return escapedText.replace(regex, '<mark class="highlight">$1</mark>')
}

function escapeHtml(text: string): string {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

function escapeRegex(text: string): string {
  return text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

// Auto-scroll to bottom
async function scrollToBottom() {
  if (autoScroll.value && logsRef.value) {
    await nextTick()
    logsRef.value.scrollTop = logsRef.value.scrollHeight
  }
}

// Watch for new logs
watch(logs, scrollToBottom, { deep: true })

function handleOpen() {
  if (!props.containerId) return
  showAllLogs.value = false
  connectLogs()
}

function connectLogs() {
  // 关闭现有连接
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }

  logs.value = []
  isConnected.value = false

  if (showAllLogs.value) {
    loadAllLogs()
    return
  }

  // Create SSE connection
  eventSource = containerApi.logs(props.containerId, 100, true)

  eventSource.onopen = () => {
    isConnected.value = true
  }

  eventSource.onmessage = (event) => {
    if (event.data) {
      // Handle escaped newlines in log data
      const logLine = event.data.replace(/\\n/g, '\n').replace(/\\r/g, '\r')
      logs.value.push(logLine)
      
      // Limit log buffer to prevent memory issues (only for realtime mode)
      if (!showAllLogs.value && logs.value.length > 10000) {
        logs.value = logs.value.slice(-5000)
      }
    }
  }

  eventSource.onerror = () => {
    isConnected.value = false
    eventSource?.close()
    eventSource = null
  }
}

async function loadAllLogs() {
  try {
    logs.value = await containerApi.logsOnce(props.containerId, 'all')
    isConnected.value = true
  } catch {
    isConnected.value = false
  }
}

function handleModeChange() {
  connectLogs()
}

function handleClose() {
  // Close SSE connection
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
  isConnected.value = false
  dialogVisible.value = false
}

function handleSearch() {
  // Search is reactive through computed property
}

function clearLogs() {
  logs.value = []
}

// Cleanup on unmount
onUnmounted(() => {
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
})
</script>

<style scoped>
.logs-dialog {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.logs-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 15px;
}

.logs-content {
  height: 400px;
  overflow-y: auto;
  background: #1e1e1e;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  padding: 10px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1;
  letter-spacing: 0;
  color: #d4d4d4;
}

.no-logs {
  color: var(--el-text-color-secondary);
  text-align: center;
  padding: 50px;
}

.log-line {
  white-space: pre;
  word-break: break-all;
  padding: 0;
  line-height: 1;
}

.log-line:hover {
  background: var(--el-fill-color-light);
}

:deep(.highlight) {
  background: var(--el-color-warning-light-5);
  color: var(--el-color-warning-dark-2);
  padding: 0 2px;
  border-radius: 2px;
}

.logs-status {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.status {
  display: flex;
  align-items: center;
  gap: 5px;
}

.status.connected {
  color: var(--el-color-success);
}

.status.disconnected {
  color: var(--el-color-danger);
}

.log-count {
  color: var(--el-text-color-secondary);
}
</style>
