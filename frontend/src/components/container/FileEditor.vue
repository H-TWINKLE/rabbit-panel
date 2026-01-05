<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('file.edit') + ' - ' + fileName"
    width="70%"
    :close-on-click-modal="false"
    destroy-on-close
    class="file-editor-dialog"
    @open="handleOpen"
  >
    <div class="file-editor">
      <div v-if="loading" class="loading-container">
        <el-icon class="is-loading" :size="32"><Loading /></el-icon>
        <p>{{ t('common.loading') }}</p>
      </div>
      <div v-else class="editor-container">
        <el-input
          v-model="content"
          type="textarea"
          :rows="20"
          :placeholder="t('file.fileContent')"
          class="code-editor"
        />
      </div>
    </div>
    <template #footer>
      <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave">
        {{ saving ? t('file.saving') : t('common.save') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Loading } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useI18n } from '@/composables/useI18n'
import { containerApi } from '@/api/containers'

const props = defineProps<{
  visible: boolean
  containerId: string
  filePath: string
  fileName: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  saved: []
}>()

const { t } = useI18n()

// Dialog visibility
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value),
})

// State
const content = ref('')
const loading = ref(false)
const saving = ref(false)

// Load file content
async function loadContent() {
  if (!props.containerId || !props.filePath) return
  
  loading.value = true
  try {
    const result = await containerApi.fileRead(props.containerId, props.filePath)
    content.value = result.content
  } catch {
    ElMessage.error(t('file.loadError'))
    dialogVisible.value = false
  } finally {
    loading.value = false
  }
}

// Save file content
async function handleSave() {
  if (!props.containerId || !props.filePath) return
  
  saving.value = true
  try {
    await containerApi.fileWrite(props.containerId, props.filePath, content.value)
    ElMessage.success(t('file.saveSuccess'))
    emit('saved')
    dialogVisible.value = false
  } catch {
    // Error handled by interceptor
  } finally {
    saving.value = false
  }
}

// Handle dialog open
function handleOpen() {
  content.value = ''
  loadContent()
}

// Watch for file path changes
watch(() => props.filePath, () => {
  if (props.visible && props.filePath) {
    loadContent()
  }
})
</script>

<style scoped>
.file-editor {
  min-height: 400px;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 400px;
  color: var(--el-text-color-secondary);
}

.loading-container .is-loading {
  animation: rotating 2s linear infinite;
}

.editor-container {
  height: 100%;
}

.code-editor :deep(.el-textarea__inner) {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.5;
  resize: vertical;
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
