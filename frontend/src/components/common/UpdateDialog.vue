<template>
  <div>
    <el-dialog v-model="visible" :title="t('update.title')" width="760px" top="8vh">
      <div v-loading="updateStore.loading" class="update-dialog">
        <template v-if="updateStore.info">
          <div class="summary-grid">
            <div class="summary-item"><span class="label">{{ t('update.currentVersion') }}</span><strong>{{ updateStore.info.current_version || 'dev' }}</strong></div>
            <div class="summary-item"><span class="label">{{ t('update.currentCommit') }}</span><strong>{{ updateStore.info.current_commit || '-' }}</strong></div>
            <div class="summary-item"><span class="label">{{ t('update.currentBuildTime') }}</span><strong>{{ formatTime(updateStore.info.current_build_time) }}</strong></div>
            <div class="summary-item"><span class="label">{{ t('update.latestVersion') }}</span><strong>{{ updateStore.info.latest_version || '-' }}</strong></div>
            <div class="summary-item"><span class="label">{{ t('update.deployMode') }}</span><strong>{{ deployModeLabel }}</strong></div>
            <div v-if="updateStore.info.deploy_mode === 'docker'" class="summary-item"><span class="label">{{ t('update.imageInfo') }}</span><strong>{{ updateStore.info.image }}:{{ updateStore.info.image_tag }}</strong></div>
          </div>

          <el-alert :type="updateStore.info.has_update ? 'warning' : 'info'" :closable="false" show-icon>
            <template #title>{{ updateStore.info.message }}</template>
            <div>{{ deployDescription }}</div>
          </el-alert>

          <el-alert v-if="updateStore.info.ignored" type="info" :closable="false" show-icon>
            <template #title>{{ t('update.ignoredNotice') }}</template>
          </el-alert>

          <div v-if="updateStore.taskStatus && (updateStore.taskStatus.status || updateStore.taskStatus.log_lines.length)" class="task-panel">
            <div class="task-header">
              <strong>{{ t('update.updateProgress') }}</strong>
              <div class="task-actions">
                <el-button :disabled="!updateStore.info?.ignored" text size="small" @click="handleClearIgnore">{{ t('update.clearIgnore') }}</el-button>
                <el-button :disabled="!updateStore.info?.last_update_status" text size="small" @click="handleClearState">{{ t('update.clearState') }}</el-button>
                <el-button text size="small" @click="minimize">{{ t('update.minimize') }}</el-button>
                <el-button text size="small" @click="updateStore.fetchTaskStatus()">{{ t('common.refresh') }}</el-button>
              </div>
            </div>
            <el-progress :percentage="displayProgress" :status="progressStatus" :indeterminate="!updateStore.taskStatus.progress_known && updateStore.taskStatus.status === 'running'" :duration="3" />
            <div class="meta">
              <span>{{ t('update.currentStage') }}: {{ updateStore.taskStatus.stage || '-' }}</span>
              <span>{{ t('update.lastUpdateStatus') }}: {{ updateStore.taskStatus.status || '-' }}</span>
              <span>{{ t('update.lastUpdateTime') }}: {{ formatTime(updateStore.taskStatus.last_update_time) }}</span>
            </div>
            <el-alert v-if="updateStore.taskStatus.last_error" type="error" :closable="false" show-icon>
              <template #title>{{ updateStore.taskStatus.last_error }}</template>
            </el-alert>
          </div>

          <div class="meta">
            <span>{{ t('update.lastCheckTime') }}: {{ formatTime(updateStore.info.last_check_time) }}</span>
            <span>{{ t('update.lastUpdateTime') }}: {{ formatTime(updateStore.info.last_update_time) }}</span>
            <span>{{ t('update.lastUpdateStatus') }}: {{ updateStore.info.last_update_status || '-' }}</span>
          </div>

          <el-alert v-if="updateStore.info.last_update_error" type="error" :closable="false" show-icon>
            <template #title>{{ updateStore.info.last_update_error }}</template>
          </el-alert>
        </template>
      </div>

      <template #footer>
        <div class="footer-actions">
          <el-button @click="updateStore.fetchUpdateInfo()">{{ t('update.checkNow') }}</el-button>
          <el-button type="primary" :disabled="!updateStore.info?.can_update || updateStore.taskStatus?.status === 'running'" :loading="updateStore.running" @click="handleRunUpdate">
            {{ t('update.updateNow') }}
          </el-button>
          <el-button @click="handleRemindLater">{{ t('update.remindLater') }}</el-button>
          <el-button :disabled="!updateStore.info?.latest_version" @click="handleIgnore">{{ t('update.ignoreVersion') }}</el-button>
          <el-button v-if="updateStore.info?.release_url" @click="openRelease">{{ t('update.viewRelease') }}</el-button>
        </div>
      </template>
    </el-dialog>

    <transition name="fade">
      <div v-if="updateStore.minimized && updateStore.taskStatus && updateStore.taskStatus.status === 'running'" class="mini-progress" @click="restore">
        <div class="mini-header">
          <strong>{{ t('update.updatingNow') }}</strong>
          <span>{{ updateStore.taskStatus.progress_known ? `${updateStore.taskStatus.progress}%` : '...' }}</span>
        </div>
        <el-progress :percentage="displayProgress" :show-text="false" :indeterminate="!updateStore.taskStatus.progress_known" :duration="3" />
        <div class="mini-stage">{{ updateStore.taskStatus.stage }}</div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from '@/composables/useI18n'
import { useUpdateStore } from '@/stores/update'

const { t } = useI18n()
const updateStore = useUpdateStore()

const props = defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ (e: 'update:modelValue', value: boolean): void }>()

const visible = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

const deployModeLabel = computed(() => {
  const mode = updateStore.info?.deploy_mode
  if (mode === 'docker') return 'Docker'
  if (mode === 'binary') return t('update.binary')
  return t('update.unknown')
})

const deployDescription = computed(() => {
  const info = updateStore.info
  if (!info) return ''
  if (info.deploy_mode === 'binary') return t('update.binaryDescription')
  if (info.deploy_mode === 'docker' && info.image_tag === 'latest') return t('update.dockerLatestDescription')
  if (info.deploy_mode === 'docker') return t('update.dockerFixedDescription')
  return t('update.unknownDescription')
})

const progressStatus = computed(() => {
  if (updateStore.taskStatus?.status === 'failed') return 'exception'
  if (updateStore.taskStatus?.status === 'success') return 'success'
  return undefined
})

const displayProgress = computed(() => {
  if (!updateStore.taskStatus) return 0
  if (updateStore.taskStatus.progress_known) return updateStore.taskStatus.progress
  return 100
})

function formatTime(value?: string): string {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

async function handleRunUpdate() {
  try {
    const message = await updateStore.runUpdate()
    ElMessage.success(message)
  } catch {
    // handled by interceptor
  }
}

async function handleIgnore() {
  await updateStore.ignoreVersion()
  ElMessage.success(t('update.ignoreSuccess'))
}

async function handleClearIgnore() {
  await updateStore.clearIgnoredVersion()
  ElMessage.success(t('update.clearIgnoreSuccess'))
}

async function handleClearState() {
  await updateStore.clearUpdateState()
  ElMessage.success(t('update.clearStateSuccess'))
}

function handleRemindLater() {
  updateStore.remindLater()
  visible.value = false
}

function openRelease() {
  if (updateStore.info?.release_url) {
    window.open(updateStore.info.release_url, '_blank')
  }
}

function minimize() {
  updateStore.setMinimized(true)
  visible.value = false
}

function restore() {
  updateStore.setMinimized(false)
  visible.value = true
}
</script>

<style scoped>
.update-dialog { display: flex; flex-direction: column; gap: 16px; }
.summary-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 12px; }
.summary-item { padding: 14px; border-radius: 10px; background: var(--el-fill-color-light); display: flex; flex-direction: column; gap: 6px; }
.label { font-size: 12px; color: var(--el-text-color-secondary); }
.meta { display: flex; flex-wrap: wrap; gap: 16px; color: var(--el-text-color-secondary); font-size: 13px; }
.task-panel { display: flex; flex-direction: column; gap: 12px; padding: 16px; border-radius: 14px; background: linear-gradient(180deg, rgba(64,158,255,0.08), rgba(64,158,255,0.03)); }
.task-header { display: flex; justify-content: space-between; align-items: center; gap: 12px; }
.task-actions { display: flex; gap: 6px; }
.footer-actions { width: 100%; display: flex; flex-wrap: wrap; justify-content: flex-start; gap: 8px; }
.notes-card { border-radius: 12px; }
.notes { margin: 0; white-space: pre-wrap; word-break: break-word; max-height: 240px; overflow: auto; line-height: 1.6; }
.mini-progress { position: fixed; right: 24px; bottom: 24px; width: 320px; z-index: 2200; padding: 14px 16px; border-radius: 16px; background: rgba(18, 22, 33, 0.92); color: #fff; box-shadow: 0 16px 40px rgba(0,0,0,0.24); cursor: pointer; backdrop-filter: blur(12px); }
.mini-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.mini-stage { margin-top: 8px; font-size: 12px; color: rgba(255,255,255,0.72); }
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
