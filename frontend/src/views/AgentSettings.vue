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
          <el-input v-model="form.api_key" type="password" show-password placeholder="sk-..." />
        </el-form-item>
        
        <el-form-item :label="t('agent.model')">
          <el-input v-model="form.model" placeholder="gpt-3.5-turbo" />
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="handleSave">{{ t('agent.save') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
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

const loadConfig = async () => {
  loading.value = true
  try {
    const res = await getAgentConfig()
    // Axios response data is usually in res.data, but request interceptor might return data directly?
    // Checking request.ts: interceptor returns response. So we need res.data
    form.value = (res.data as unknown as AgentConfig) 
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const handleSave = async () => {
  loading.value = true
  try {
    await saveAgentConfig(form.value)
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
</style>
