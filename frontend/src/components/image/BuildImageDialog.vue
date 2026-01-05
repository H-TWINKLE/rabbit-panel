<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('image.build')"
    width="700px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      :disabled="building"
    >
      <el-form-item :label="t('image.imageName')" prop="name">
        <el-input
          v-model="form.name"
          :placeholder="t('image.nameRequired')"
        />
      </el-form-item>

      <el-form-item :label="t('image.imageTag')" prop="tag">
        <el-input
          v-model="form.tag"
          placeholder="latest"
        />
      </el-form-item>

      <el-form-item :label="t('image.dockerfile')" prop="dockerfile">
        <el-input
          v-model="form.dockerfile"
          type="textarea"
          :rows="12"
          :placeholder="dockerfilePlaceholder"
          class="dockerfile-editor"
        />
      </el-form-item>
    </el-form>

    <!-- Build Logs -->
    <div v-if="buildLogs.length > 0" class="build-logs">
      <div class="logs-header">
        <span>{{ t('image.buildLogs') }}</span>
        <el-tag v-if="building" type="warning" size="small">
          {{ t('image.building') }}
        </el-tag>
        <el-tag v-else-if="buildSuccess" type="success" size="small">
          {{ t('image.buildSuccess') }}
        </el-tag>
        <el-tag v-else-if="buildError" type="danger" size="small">
          {{ t('image.buildFailed') }}
        </el-tag>
      </div>
      <div ref="logsContainer" class="logs-content">
        <div
          v-for="(log, index) in buildLogs"
          :key="index"
          :class="['log-line', log.type]"
        >
          {{ log.message }}
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose" :disabled="building">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          type="primary"
          :loading="building"
          @click="handleBuild"
        >
          {{ building ? t('image.building') : t('image.build') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { useI18n } from '@/composables/useI18n'
import { imageApi } from '@/api/images'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'built'): void
}>()

const { t } = useI18n()

// Dialog visibility
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value),
})

// Form
const formRef = ref<FormInstance>()
const form = ref({
  name: '',
  tag: 'latest',
  dockerfile: '',
})

// Validation rules
const rules = computed<FormRules>(() => ({
  name: [
    { required: true, message: t('image.nameRequired'), trigger: 'blur' },
  ],
  tag: [
    { required: true, message: t('image.tagRequired'), trigger: 'blur' },
  ],
  dockerfile: [
    { required: true, message: t('image.dockerfileRequired'), trigger: 'blur' },
  ],
}))

// Build state
const building = ref(false)
const buildSuccess = ref(false)
const buildError = ref(false)
const buildLogs = ref<Array<{ type: string; message: string }>>([])
const logsContainer = ref<HTMLElement>()

// Dockerfile placeholder
const dockerfilePlaceholder = `FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3000
CMD ["npm", "start"]`

// Auto-scroll logs
watch(buildLogs, () => {
  nextTick(() => {
    if (logsContainer.value) {
      logsContainer.value.scrollTop = logsContainer.value.scrollHeight
    }
  })
}, { deep: true })

// Handlers
function handleClose() {
  if (building.value) return
  resetForm()
  dialogVisible.value = false
}

function resetForm() {
  form.value = {
    name: '',
    tag: 'latest',
    dockerfile: '',
  }
  buildLogs.value = []
  buildSuccess.value = false
  buildError.value = false
  formRef.value?.resetFields()
}

async function handleBuild() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch {
    return
  }

  building.value = true
  buildSuccess.value = false
  buildError.value = false
  buildLogs.value = []

  try {
    await imageApi.build(
      form.value.name,
      form.value.tag,
      form.value.dockerfile,
      // onMessage
      (data) => {
        if (data.message) {
          buildLogs.value.push({
            type: data.type || 'info',
            message: data.message,
          })
        }
        if (data.error) {
          buildLogs.value.push({
            type: 'error',
            message: data.error,
          })
          buildError.value = true
        }
      },
      // onError
      (error) => {
        buildLogs.value.push({
          type: 'error',
          message: error,
        })
        buildError.value = true
        building.value = false
        ElMessage.error(t('image.buildFailed'))
      },
      // onComplete
      () => {
        building.value = false
        if (!buildError.value) {
          buildSuccess.value = true
          ElMessage.success(t('image.buildSuccess'))
          emit('built')
        }
      }
    )
  } catch (error) {
    building.value = false
    buildError.value = true
    ElMessage.error(t('image.buildFailed'))
  }
}
</script>

<style scoped>
.dockerfile-editor :deep(textarea) {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
}

.build-logs {
  margin-top: 20px;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
}

.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 15px;
  background-color: var(--el-fill-color-light);
  border-bottom: 1px solid var(--el-border-color);
  font-weight: 500;
}

.logs-content {
  max-height: 300px;
  overflow-y: auto;
  padding: 10px 15px;
  background-color: var(--el-bg-color);
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.6;
}

.log-line {
  white-space: pre-wrap;
  word-break: break-all;
}

.log-line.error {
  color: var(--el-color-danger);
}

.log-line.warning {
  color: var(--el-color-warning);
}

.log-line.success {
  color: var(--el-color-success);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
