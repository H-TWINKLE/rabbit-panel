<template>
  <el-alert
    v-if="updateStore.shouldShowBanner && updateStore.info"
    class="update-banner"
    type="warning"
    show-icon
    :closable="true"
    @close="updateStore.remindLater"
  >
    <template #title>{{ updateStore.info.message }}</template>
    <div class="content">
      <span>{{ description }}</span>
      <div class="actions">
        <el-button size="small" @click="updateStore.fetchUpdateInfo()">{{ t('update.checkNow') }}</el-button>
        <el-button size="small" type="primary" :disabled="!updateStore.info.has_update || !updateStore.info.can_update" :loading="updateStore.running" @click="emit('open-dialog')">
          {{ t('update.updateNow') }}
        </el-button>
        <el-button size="small" @click="emit('open-dialog')">{{ t('update.details') }}</el-button>
        <el-button size="small" @click="handleIgnore">{{ t('update.ignoreVersion') }}</el-button>
      </div>
    </div>
  </el-alert>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from '@/composables/useI18n'
import { useUpdateStore } from '@/stores/update'

const emit = defineEmits<{ (e: 'open-dialog'): void }>()
const { t } = useI18n()
const updateStore = useUpdateStore()

const description = computed(() => {
  const info = updateStore.info
  if (!info) return ''
  if (info.deploy_mode === 'binary') return t('update.binaryDescription')
  if (info.deploy_mode === 'docker' && info.image_tag === 'latest') return t('update.dockerLatestDescription')
  if (info.deploy_mode === 'docker') return t('update.dockerFixedDescription')
  return t('update.unknownDescription')
})

async function handleIgnore() {
  await updateStore.ignoreVersion()
  ElMessage.success(t('update.ignoreSuccess'))
}
</script>

<style scoped>
.update-banner {
  margin-bottom: 16px;
  border-radius: 14px;
}

.content {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
