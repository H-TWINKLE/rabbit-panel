<template>
  <div class="agent-settings">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('agent.settingsTitle') }}</span>
        </div>
      </template>
      
      <el-form :model="form" label-width="120px" v-loading="loading">
        <el-form-item :label="t('agent.enable')">
          <el-switch v-model="form.enabled" />
        </el-form-item>
        
        <el-form-item :label="t('agent.apiUrl')">
          <el-input v-model="form.api_url" placeholder="https://api.openai.com/v1" />
          <div class="form-tip">{{ t('agent.apiTip') }}</div>
        </el-form-item>
        
        <el-form-item :label="t('agent.apiKey')">
          <div class="key-row">
            <span class="key-mask">{{ displayApiKey }}</span>
            <el-button @click="showKeyDialog = true">{{ t('agent.changeKey') }}</el-button>
          </div>
        </el-form-item>
        
        <el-form-item :label="t('agent.model')">
          <el-input v-model="form.model" placeholder="gpt-3.5-turbo" />
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="handleSave">{{ t('agent.save') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-dialog v-model="showKeyDialog" :title="t('agent.changeKey')" width="460px">
      <el-form label-width="100px">
        <el-form-item :label="t('agent.apiKey')">
          <el-input v-model="newApiKey" type="password" show-password placeholder="sk-..." />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleCancelKeyDialog">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleConfirmKeyChange">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { getAgentConfig, saveAgentConfig, type AgentConfig } from '@/api/agent'
import { ElMessage } from 'element-plus'

const { t } = useI18n()

const loading = ref(false)
const form = ref<AgentConfig>({
  api_url: 'https://api.openai.com/v1',
  api_key: '',
  model: 'gpt-3.5-turbo',
  enabled: false
})
const showKeyDialog = ref(false)
const newApiKey = ref('')

const displayApiKey = computed(() => form.value.api_key || t('agent.notSet'))

const loadConfig = async () => {
  loading.value = true
  try {
    const res = await getAgentConfig()
    form.value = res.data as AgentConfig
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const handleSave = async () => {
  loading.value = true
  try {
    await saveAgentConfig({
      ...form.value,
      api_key: '',
    })
    await loadConfig()
    ElMessage.success(t('agent.saveSuccess'))
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const handleCancelKeyDialog = () => {
  showKeyDialog.value = false
  newApiKey.value = ''
}

const handleConfirmKeyChange = async () => {
  const key = newApiKey.value.trim()
  if (!key) {
    return
  }
  loading.value = true
  try {
    await saveAgentConfig({
      ...form.value,
      api_key: key,
    })
    await loadConfig()
    showKeyDialog.value = false
    newApiKey.value = ''
    ElMessage.success(t('agent.saveSuccess'))
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<style scoped>
.agent-settings {
  max-width: 800px;
  margin: 0 auto;
}
.form-tip {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}
.key-row {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.key-mask {
  color: var(--el-text-color-regular);
  font-family: monospace;
}
</style>
