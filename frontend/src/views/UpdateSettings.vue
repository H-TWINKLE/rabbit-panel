<template>
  <div class="update-settings-page">
    <UpdateBanner @open-dialog="showDialog = true" />
    <el-card shadow="hover">
      <template #header>
        <div class="header">
          <span>{{ t('update.title') }}</span>
          <el-button :loading="updateStore.loading" @click="updateStore.fetchUpdateInfo()">
            {{ t('update.checkNow') }}
          </el-button>
        </div>
      </template>

      <div v-if="updateStore.info" class="info-grid">
        <div class="row"><span>{{ t('update.currentVersion') }}</span><strong>{{ updateStore.info.current_version || 'dev' }}</strong></div>
        <div class="row"><span>{{ t('update.currentCommit') }}</span><strong>{{ updateStore.info.current_commit || '-' }}</strong></div>
        <div class="row"><span>{{ t('update.currentBuildTime') }}</span><strong>{{ formatTime(updateStore.info.current_build_time) }}</strong></div>
        <div class="row"><span>{{ t('update.latestVersion') }}</span><strong>{{ updateStore.info.latest_version || '-' }}</strong></div>
        <div class="row"><span>{{ t('update.deployMode') }}</span><strong>{{ updateStore.info.deploy_mode }}</strong></div>
        <div class="row"><span>{{ t('update.imageInfo') }}</span><strong>{{ updateStore.info.image }}:{{ updateStore.info.image_tag }}</strong></div>
        <div class="row"><span>{{ t('update.lastCheckTime') }}</span><strong>{{ formatTime(updateStore.info.last_check_time) }}</strong></div>
        <div class="row"><span>{{ t('update.lastUpdateStatus') }}</span><strong>{{ updateStore.info.last_update_status || '-' }}</strong></div>
      </div>

      <div class="actions">
        <el-button type="primary" :disabled="!updateStore.info?.can_update" :loading="updateStore.running" @click="handleRunUpdate">
          {{ t('update.updateNow') }}
        </el-button>
        <el-button @click="updateStore.remindLater()">{{ t('update.remindLater') }}</el-button>
        <el-button @click="handleIgnore">{{ t('update.ignoreVersion') }}</el-button>
        <el-button v-if="updateStore.info?.release_url" @click="openRelease">{{ t('update.viewRelease') }}</el-button>
      </div>

      <el-card shadow="never" class="notes-card">
        <template #header>{{ t('update.releaseNotes') }}</template>
        <pre class="notes">{{ updateStore.info?.release_notes || '-' }}</pre>
      </el-card>
    </el-card>

    <UpdateDialog v-model="showDialog" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from '@/composables/useI18n'
import { useUpdateStore } from '@/stores/update'
import UpdateBanner from '@/components/common/UpdateBanner.vue'
import UpdateDialog from '@/components/common/UpdateDialog.vue'

const { t } = useI18n()
const updateStore = useUpdateStore()
const showDialog = ref(false)

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

function openRelease() {
  if (updateStore.info?.release_url) {
    window.open(updateStore.info.release_url, '_blank')
  }
}

onMounted(() => {
  if (!updateStore.info) {
    updateStore.fetchUpdateInfo()
  }
})
</script>

<style scoped>
.update-settings-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 12px;
  margin-bottom: 18px;
}

.row {
  border-radius: 12px;
  background: var(--el-fill-color-light);
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.row span {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 18px;
}

.notes-card {
  border-radius: 12px;
}

.notes {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 320px;
  overflow: auto;
  line-height: 1.6;
}
</style>
