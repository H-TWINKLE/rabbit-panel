<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('volumes.create')"
    width="600px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      label-position="right"
    >
      <el-form-item :label="t('volumes.name')" prop="name">
        <el-input
          v-model="form.name"
          :placeholder="t('volumes.namePlaceholder')"
          clearable
        />
      </el-form-item>

      <el-form-item :label="t('volumes.driver')" prop="driver">
        <el-select v-model="form.driver" style="width: 100%" :placeholder="t('volumes.driverPlaceholder')">
          <el-option
            v-for="driver in driverOptions"
            :key="driver.value"
            :label="driver.label"
            :value="driver.value"
          />
        </el-select>
      </el-form-item>

      <!-- Driver Options -->
      <el-form-item :label="t('volumes.driverOpts')">
        <div class="dynamic-list">
          <div v-for="(opt, index) in form.driverOpts" :key="index" class="dynamic-item">
            <el-input
              v-model="opt.key"
              :placeholder="t('volumes.optKey')"
              style="width: 150px"
            />
            <span class="separator">=</span>
            <el-input
              v-model="opt.value"
              :placeholder="t('volumes.optValue')"
              style="width: 200px"
            />
            <el-button
              type="danger"
              :icon="Delete"
              circle
              size="small"
              @click="removeDriverOpt(index)"
            />
          </div>
          <el-button type="primary" text size="small" @click="addDriverOpt">
            <el-icon><Plus /></el-icon>
            {{ t('volumes.addDriverOpt') }}
          </el-button>
        </div>
      </el-form-item>

      <!-- Labels -->
      <el-form-item :label="t('volumes.labels')">
        <div class="dynamic-list">
          <div v-for="(label, index) in form.labels" :key="index" class="dynamic-item">
            <el-input
              v-model="label.key"
              :placeholder="t('volumes.labelKey')"
              style="width: 150px"
            />
            <span class="separator">=</span>
            <el-input
              v-model="label.value"
              :placeholder="t('volumes.labelValue')"
              style="width: 200px"
            />
            <el-button
              type="danger"
              :icon="Delete"
              circle
              size="small"
              @click="removeLabel(index)"
            />
          </div>
          <el-button type="primary" text size="small" @click="addLabel">
            <el-icon><Plus /></el-icon>
            {{ t('volumes.addLabel') }}
          </el-button>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="loading" @click="handleSubmit">
          {{ t('common.create') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import { useVolumeStore } from '@/stores/volumes'
import { useI18n } from '@/composables/useI18n'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'created'): void
}>()

const { t } = useI18n()
const volumeStore = useVolumeStore()

// Dialog visibility
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value),
})

// Form state
const formRef = ref<FormInstance>()
const loading = ref(false)

interface KeyValue {
  key: string
  value: string
}

const form = ref({
  name: '',
  driver: 'local',
  driverOpts: [] as KeyValue[],
  labels: [] as KeyValue[],
})

// Driver options
const driverOptions = [
  { value: 'local', label: 'local' },
]

// Form validation rules
const rules = computed<FormRules>(() => ({
  name: [
    { required: true, message: t('volumes.nameRequired'), trigger: 'blur' },
    {
      pattern: /^[a-zA-Z0-9][a-zA-Z0-9_.-]*$/,
      message: 'Volume name must start with alphanumeric and contain only alphanumeric, underscore, period, or hyphen',
      trigger: 'blur',
    },
  ],
}))

// Driver options management
function addDriverOpt() {
  form.value.driverOpts.push({ key: '', value: '' })
}

function removeDriverOpt(index: number) {
  form.value.driverOpts.splice(index, 1)
}

// Labels management
function addLabel() {
  form.value.labels.push({ key: '', value: '' })
}

function removeLabel(index: number) {
  form.value.labels.splice(index, 1)
}

// Reset form when dialog opens
watch(
  () => props.visible,
  (newVal) => {
    if (newVal) {
      form.value = {
        name: '',
        driver: 'local',
        driverOpts: [],
        labels: [],
      }
      formRef.value?.clearValidate()
    }
  }
)

// Handlers
function handleClose() {
  dialogVisible.value = false
}

async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    loading.value = true

    // Convert arrays to objects
    const driverOpts: Record<string, string> = {}
    form.value.driverOpts.forEach((opt) => {
      if (opt.key) {
        driverOpts[opt.key] = opt.value
      }
    })

    const labels: Record<string, string> = {}
    form.value.labels.forEach((label) => {
      if (label.key) {
        labels[label.key] = label.value
      }
    })

    await volumeStore.createVolume({
      name: form.value.name,
      driver: form.value.driver,
      driverOpts: Object.keys(driverOpts).length > 0 ? driverOpts : undefined,
      labels: Object.keys(labels).length > 0 ? labels : undefined,
    })
    ElMessage.success(t('volumes.createSuccess'))
    emit('created')
    handleClose()
  } catch {
    // Validation error or API error (handled by interceptor)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.dynamic-list {
  width: 100%;
}

.dynamic-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.separator {
  color: var(--el-text-color-secondary);
  font-weight: bold;
}
</style>
