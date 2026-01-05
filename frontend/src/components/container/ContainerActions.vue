<template>
  <div class="container-actions" :class="{ compact }">
    <!-- Compact mode: dropdown menu -->
    <template v-if="compact">
      <el-dropdown trigger="click" @command="handleCommand">
        <el-button type="primary" size="small">
          {{ t('common.actions') }}
          <el-icon class="el-icon--right"><ArrowDown /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item v-if="container.state !== 'running'" command="start">
              <el-icon><VideoPlay /></el-icon>
              {{ t('container.start') }}
            </el-dropdown-item>
            <el-dropdown-item v-if="container.state === 'running'" command="stop">
              <el-icon><VideoPause /></el-icon>
              {{ t('container.stop') }}
            </el-dropdown-item>
            <el-dropdown-item command="restart">
              <el-icon><RefreshRight /></el-icon>
              {{ t('container.restart') }}
            </el-dropdown-item>
            <el-dropdown-item divided command="logs">
              <el-icon><Document /></el-icon>
              {{ t('container.logs') }}
            </el-dropdown-item>
            <el-dropdown-item v-if="container.state === 'running'" command="terminal">
              <el-icon><Monitor /></el-icon>
              {{ t('container.terminal') }}
            </el-dropdown-item>
            <el-dropdown-item v-if="container.state === 'running'" command="files">
              <el-icon><FolderOpened /></el-icon>
              {{ t('container.files') }}
            </el-dropdown-item>
            <el-dropdown-item command="config">
              <el-icon><Setting /></el-icon>
              {{ t('container.config') }}
            </el-dropdown-item>
            <el-dropdown-item divided command="remove" class="danger-item">
              <el-icon><Delete /></el-icon>
              {{ t('common.delete') }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </template>

    <!-- Normal mode: icon buttons -->
    <template v-else>
      <el-button-group class="action-group">
        <!-- Start button (disabled when running or action in progress) -->
        <el-tooltip :content="t('container.start')" placement="top" :disabled="container.state === 'running'">
          <el-button
            type="success"
            size="small"
            :icon="VideoPlay"
            :loading="actionLoading === 'start'"
            :disabled="container.state === 'running' || containerStore.actionInProgress"
            @click="handleAction('start')"
          />
        </el-tooltip>

        <!-- Stop button (disabled when not running or action in progress) -->
        <el-tooltip :content="t('container.stop')" placement="top" :disabled="container.state !== 'running'">
          <el-button
            type="warning"
            size="small"
            :icon="VideoPause"
            :loading="actionLoading === 'stop'"
            :disabled="container.state !== 'running' || containerStore.actionInProgress"
            @click="handleAction('stop')"
          />
        </el-tooltip>

        <!-- Restart button (disabled when action in progress) -->
        <el-tooltip :content="t('container.restart')" placement="top">
          <el-button
            type="primary"
            size="small"
            :icon="RefreshRight"
            :loading="actionLoading === 'restart'"
            :disabled="containerStore.actionInProgress"
            @click="handleAction('restart')"
          />
        </el-tooltip>

        <!-- Logs button -->
        <el-tooltip :content="t('container.logs')" placement="top">
          <el-button
            type="info"
            size="small"
            :icon="Document"
            @click="handleViewLogs"
          />
        </el-tooltip>

        <!-- Terminal button (disabled when not running) -->
        <el-tooltip :content="t('container.terminal')" placement="top" :disabled="container.state !== 'running'">
          <el-button
            size="small"
            :icon="Monitor"
            :disabled="container.state !== 'running'"
            @click="handleOpenTerminal"
          />
        </el-tooltip>

        <!-- Files button (disabled when not running) -->
        <el-tooltip :content="t('container.files')" placement="top" :disabled="container.state !== 'running'">
          <el-button
            size="small"
            :icon="FolderOpened"
            :disabled="container.state !== 'running'"
            @click="handleOpenFiles"
          />
        </el-tooltip>

        <!-- Config button -->
        <el-tooltip :content="t('container.config')" placement="top">
          <el-button
            size="small"
            :icon="Setting"
            @click="handleOpenConfig"
          />
        </el-tooltip>

        <!-- Delete button with confirmation (disabled when action in progress) -->
        <el-popconfirm
          :title="t('container.confirmRemove')"
          :confirm-button-text="t('common.confirm')"
          :cancel-button-text="t('common.cancel')"
          :disabled="containerStore.actionInProgress"
          @confirm="handleAction('remove')"
        >
          <template #reference>
            <el-button
              type="danger"
              size="small"
              :icon="Delete"
              :loading="actionLoading === 'remove'"
              :disabled="containerStore.actionInProgress"
            />
          </template>
        </el-popconfirm>
      </el-button-group>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  VideoPlay,
  VideoPause,
  RefreshRight,
  Document,
  Delete,
  Monitor,
  FolderOpened,
  Setting,
  ArrowDown,
} from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'
import { useI18n } from '@/composables/useI18n'
import { useContainerStore } from '@/stores/containers'
import type { ContainerInfo } from '@/types'

const { t } = useI18n()
const containerStore = useContainerStore()

const props = withDefaults(defineProps<{
  container: ContainerInfo
  compact?: boolean
}>(), {
  compact: false
})

const emit = defineEmits<{
  action: [id: string, action: 'start' | 'stop' | 'restart' | 'remove']
  logs: [id: string, name: string]
  terminal: [id: string, name: string]
  files: [id: string, name: string]
  config: [id: string, name: string]
}>()

const actionLoading = ref<string | null>(null)

// 监听 store 的 actionInProgress 状态，当操作完成时重置 loading
watch(() => containerStore.actionInProgress, (inProgress) => {
  if (!inProgress) {
    actionLoading.value = null
  }
})

function handleAction(action: 'start' | 'stop' | 'restart' | 'remove') {
  actionLoading.value = action
  emit('action', props.container.id, action)
}

async function handleCommand(command: string) {
  if (command === 'remove') {
    try {
      await ElMessageBox.confirm(
        t('container.confirmRemove'),
        t('common.warning'),
        {
          confirmButtonText: t('common.confirm'),
          cancelButtonText: t('common.cancel'),
          type: 'warning',
        }
      )
      handleAction('remove')
    } catch {
      // User cancelled
    }
    return
  }
  
  switch (command) {
    case 'start':
    case 'stop':
    case 'restart':
      handleAction(command)
      break
    case 'logs':
      handleViewLogs()
      break
    case 'terminal':
      handleOpenTerminal()
      break
    case 'files':
      handleOpenFiles()
      break
    case 'config':
      handleOpenConfig()
      break
  }
}

function handleViewLogs() {
  emit('logs', props.container.id, props.container.name)
}

function handleOpenTerminal() {
  emit('terminal', props.container.id, props.container.name)
}

function handleOpenFiles() {
  emit('files', props.container.id, props.container.name)
}

function handleOpenConfig() {
  emit('config', props.container.id, props.container.name)
}
</script>

<style scoped>
.container-actions {
  display: flex;
  align-items: center;
  justify-content: center;
}

.container-actions.compact {
  justify-content: center;
}

.container-actions .action-group {
  display: flex;
  flex-wrap: nowrap;
}

/* Make icon-only buttons square and consistent */
.container-actions :deep(.el-button) {
  padding: 10px 12px;
  min-width: 42px;
  height: 42px;
  font-size: 18px;
}

.container-actions :deep(.el-button .el-icon) {
  margin: 0;
}

:deep(.danger-item) {
  color: var(--el-color-danger) !important;
}

:deep(.danger-item .el-icon) {
  color: var(--el-color-danger) !important;
}
</style>
