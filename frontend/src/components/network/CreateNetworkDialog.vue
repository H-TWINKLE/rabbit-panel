<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('network.create')"
    width="550px"
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
      <el-form-item :label="t('network.networkName')" prop="name">
        <el-input
          v-model="form.name"
          :placeholder="t('network.nameRequired')"
          clearable
        />
      </el-form-item>

      <el-form-item :label="t('network.networkDriver')" prop="driver">
        <el-select v-model="form.driver" style="width: 100%">
          <el-option
            v-for="driver in driverOptions"
            :key="driver.value"
            :label="driver.label"
            :value="driver.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('network.subnet')" prop="subnet">
        <el-input
          v-model="form.subnet"
          placeholder="例如: 172.20.0.0/16"
          clearable
        />
        <div class="form-tip">可选，留空则由 Docker 自动分配</div>
      </el-form-item>

      <el-form-item :label="t('network.gateway')" prop="gateway">
        <el-input
          v-model="form.gateway"
          placeholder="例如: 172.20.0.1"
          clearable
          :disabled="!form.subnet"
        />
        <div class="form-tip">可选，需要先设置子网</div>
      </el-form-item>

      <el-form-item :label="t('network.internal')">
        <el-switch v-model="form.internal" />
        <span class="switch-label">{{ form.internal ? '是' : '否' }}</span>
        <div class="form-tip">内部网络不允许外部连接</div>
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
import { networkApi } from '@/api/networks'
import { useI18n } from '@/composables/useI18n'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'created'): void
}>()

const { t } = useI18n()

// Dialog visibility
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value),
})

// Form state
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = ref({
  name: '',
  driver: 'bridge',
  subnet: '',
  gateway: '',
  internal: false,
})

// Driver options
const driverOptions = computed(() => [
  { value: 'bridge', label: t('network.drivers.bridge') },
  { value: 'overlay', label: t('network.drivers.overlay') },
  { value: 'macvlan', label: t('network.drivers.macvlan') },
  { value: 'host', label: t('network.drivers.host') },
  { value: 'none', label: t('network.drivers.none') },
])

// CIDR validation regex
const cidrRegex = /^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/
const ipRegex = /^(\d{1,3}\.){3}\d{1,3}$/

// Form validation rules
const rules = computed<FormRules>(() => ({
  name: [
    { required: true, message: t('network.nameRequired'), trigger: 'blur' },
    { 
      pattern: /^[a-zA-Z0-9][a-zA-Z0-9_.-]*$/, 
      message: '网络名称必须以字母或数字开头，只能包含字母、数字、下划线、点或连字符',
      trigger: 'blur' 
    },
  ],
  driver: [
    { required: true, message: t('network.driverRequired'), trigger: 'change' },
  ],
  subnet: [
    { 
      validator: (_rule, value, callback) => {
        if (value && !cidrRegex.test(value)) {
          callback(new Error('子网格式不正确，例如: 172.20.0.0/16'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    },
  ],
  gateway: [
    { 
      validator: (_rule, value, callback) => {
        if (value && !ipRegex.test(value)) {
          callback(new Error('网关格式不正确，例如: 172.20.0.1'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    },
  ],
}))

// Reset form when dialog opens
watch(() => props.visible, (newVal) => {
  if (newVal) {
    form.value = {
      name: '',
      driver: 'bridge',
      subnet: '',
      gateway: '',
      internal: false,
    }
    formRef.value?.clearValidate()
  }
})

// Clear gateway when subnet is cleared
watch(() => form.value.subnet, (newVal) => {
  if (!newVal) {
    form.value.gateway = ''
  }
})

// Handlers
function handleClose() {
  dialogVisible.value = false
}

async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    loading.value = true

    await networkApi.createFull({
      name: form.value.name,
      driver: form.value.driver,
      subnet: form.value.subnet || undefined,
      gateway: form.value.gateway || undefined,
      internal: form.value.internal,
    })
    ElMessage.success(t('network.createSuccess'))
    emit('created')
    handleClose()
  } catch (error) {
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

.form-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.4;
  margin-top: 4px;
}

.switch-label {
  margin-left: 8px;
  font-size: 14px;
  color: var(--el-text-color-regular);
}
</style>
