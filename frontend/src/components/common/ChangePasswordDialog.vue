<template>
  <el-dialog
    v-model="visible"
    :title="isForced ? '首次登录 - 请修改密码' : '修改密码'"
    width="450px"
    :close-on-click-modal="!isForced"
    :close-on-press-escape="!isForced"
    :show-close="!isForced"
    @close="handleClose"
  >
    <el-alert
      v-if="isForced"
      type="warning"
      :closable="false"
      show-icon
      class="mb-4"
    >
      首次登录需要修改默认密码以确保账户安全
    </el-alert>

    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
      @submit.prevent="handleSubmit"
    >
      <el-form-item label="旧密码" prop="oldPassword" v-if="!isForced">
        <el-input
          v-model="form.oldPassword"
          type="password"
          placeholder="请输入旧密码"
          show-password
          autocomplete="current-password"
        />
      </el-form-item>

      <el-form-item label="新密码" prop="newPassword">
        <el-input
          v-model="form.newPassword"
          type="password"
          placeholder="请输入新密码"
          show-password
          autocomplete="new-password"
          @input="handlePasswordInput"
        />
      </el-form-item>

      <!-- Password strength indicator -->
      <el-form-item v-if="form.newPassword">
        <div class="password-strength">
          <span class="strength-label">密码强度：</span>
          <el-progress
            :percentage="strengthPercentage"
            :color="strengthColor"
            :stroke-width="8"
            :show-text="false"
            class="strength-progress"
          />
          <span class="strength-text" :style="{ color: strengthColor }">
            {{ strengthLabel }}
          </span>
        </div>
        <div v-if="passwordErrors.length > 0" class="password-errors">
          <div v-for="error in passwordErrors" :key="error" class="error-item">
            <el-icon><WarningFilled /></el-icon>
            <span>{{ error }}</span>
          </div>
        </div>
      </el-form-item>

      <el-form-item label="确认密码" prop="confirmPassword">
        <el-input
          v-model="form.confirmPassword"
          type="password"
          placeholder="请再次输入新密码"
          show-password
          autocomplete="new-password"
          @keyup.enter="handleSubmit"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button v-if="!isForced" @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">
        {{ loading ? '提交中...' : '确认修改' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { WarningFilled } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import {
  validatePassword,
  getStrengthColor,
  getStrengthPercentage,
  getStrengthLabel,
} from '@/utils/password'

interface Props {
  modelValue: boolean
  forced?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  forced: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'success': []
}>()

const authStore = useAuthStore()

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const isForced = computed(() => props.forced)

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

// Password validation state
const passwordErrors = ref<string[]>([])
const passwordStrength = ref<'weak' | 'medium' | 'strong'>('weak')

const strengthColor = computed(() => getStrengthColor(passwordStrength.value))
const strengthPercentage = computed(() => getStrengthPercentage(passwordStrength.value))
const strengthLabel = computed(() => getStrengthLabel(passwordStrength.value))

// Custom validator for new password
const validateNewPassword = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (!value) {
    callback(new Error('请输入新密码'))
    return
  }
  
  const result = validatePassword(value)
  if (!result.valid) {
    callback(new Error(result.errors[0]))
    return
  }
  
  // Also validate confirm password if it's filled
  if (form.confirmPassword) {
    formRef.value?.validateField('confirmPassword')
  }
  
  callback()
}

// Custom validator for confirm password
const validateConfirmPassword = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (!value) {
    callback(new Error('请再次输入新密码'))
    return
  }
  
  if (value !== form.newPassword) {
    callback(new Error('两次输入的密码不一致'))
    return
  }
  
  callback()
}

const rules: FormRules = {
  oldPassword: [
    { required: true, message: '请输入旧密码', trigger: 'blur' },
  ],
  newPassword: [
    { required: true, validator: validateNewPassword, trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, validator: validateConfirmPassword, trigger: 'blur' },
  ],
}

// Update password strength on input
function handlePasswordInput() {
  if (form.newPassword) {
    const result = validatePassword(form.newPassword)
    passwordErrors.value = result.errors
    passwordStrength.value = result.strength
  } else {
    passwordErrors.value = []
    passwordStrength.value = 'weak'
  }
}

async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch {
    return
  }

  loading.value = true

  try {
    await authStore.changePassword(
      isForced.value ? '' : form.oldPassword,
      form.newPassword
    )
    
    ElMessage.success('密码修改成功')
    emit('success')
    handleClose()
  } catch (error: unknown) {
    const err = error as { response?: { data?: { error?: string } } }
    ElMessage.error(err.response?.data?.error || '密码修改失败')
  } finally {
    loading.value = false
  }
}

function handleClose() {
  if (!isForced.value) {
    visible.value = false
    resetForm()
  }
}

function resetForm() {
  form.oldPassword = ''
  form.newPassword = ''
  form.confirmPassword = ''
  passwordErrors.value = []
  passwordStrength.value = 'weak'
  formRef.value?.resetFields()
}

// Reset form when dialog opens
watch(visible, (newVal) => {
  if (newVal) {
    resetForm()
  }
})
</script>

<style scoped>
.mb-4 {
  margin-bottom: 16px;
}

.password-strength {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.strength-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  white-space: nowrap;
}

.strength-progress {
  flex: 1;
  max-width: 120px;
}

.strength-text {
  font-size: 12px;
  font-weight: 500;
  min-width: 20px;
}

.password-errors {
  margin-top: 8px;
}

.error-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--el-color-warning);
  margin-bottom: 4px;
}

.error-item .el-icon {
  font-size: 14px;
}
</style>
