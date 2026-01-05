<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('config.recreate')"
    width="700px"
    :close-on-click-modal="false"
    destroy-on-close
    append-to-body
    @open="handleOpen"
  >
    <el-alert
      type="warning"
      :title="t('config.recreateConfirm')"
      :closable="false"
      show-icon
      style="margin-bottom: 20px"
    />

    <el-form ref="formRef" :model="form" label-width="120px">
      <!-- Image -->
      <el-form-item :label="t('container.image')" prop="image" required>
        <el-input v-model="form.image" :placeholder="t('container.image')" />
      </el-form-item>

      <!-- Name -->
      <el-form-item :label="t('common.name')" prop="name" required>
        <el-input v-model="form.name" :placeholder="t('container.containerName')" />
      </el-form-item>

      <!-- Restart Policy -->
      <el-form-item :label="t('container.restartPolicy')">
        <el-select v-model="form.restart" style="width: 100%">
          <el-option value="no" :label="t('container.restartPolicies.no')" />
          <el-option value="always" :label="t('container.restartPolicies.always')" />
          <el-option value="on-failure" :label="t('container.restartPolicies.onFailure')" />
          <el-option value="unless-stopped" :label="t('container.restartPolicies.unlessStopped')" />
        </el-select>
      </el-form-item>

      <!-- Network -->
      <el-form-item :label="t('container.network')">
        <el-input v-model="form.network" placeholder="bridge" />
      </el-form-item>

      <!-- Ports -->
      <el-form-item :label="t('container.ports')">
        <div class="dynamic-list">
          <div v-for="(port, index) in form.ports" :key="index" class="list-item">
            <el-input
              v-model="port.host"
              :placeholder="t('container.hostPort')"
              style="width: 120px"
            />
            <span class="separator">:</span>
            <el-input
              v-model="port.container"
              :placeholder="t('container.containerPort')"
              style="width: 120px"
            />
            <el-button type="danger" :icon="Delete" circle size="small" @click="removePort(index)" />
          </div>
          <el-button type="primary" size="small" @click="addPort">
            <el-icon><Plus /></el-icon>
            {{ t('container.addPort') }}
          </el-button>
        </div>
      </el-form-item>

      <!-- Volumes -->
      <el-form-item :label="t('container.volumes')">
        <div class="dynamic-list">
          <div v-for="(volume, index) in form.volumes" :key="index" class="list-item">
            <el-input
              v-model="volume.host"
              :placeholder="t('container.hostPath')"
              style="width: 180px"
            />
            <span class="separator">:</span>
            <el-input
              v-model="volume.container"
              :placeholder="t('container.containerPath')"
              style="width: 180px"
            />
            <el-button type="danger" :icon="Delete" circle size="small" @click="removeVolume(index)" />
          </div>
          <el-button type="primary" size="small" @click="addVolume">
            <el-icon><Plus /></el-icon>
            {{ t('container.addVolume') }}
          </el-button>
        </div>
      </el-form-item>

      <!-- Environment Variables -->
      <el-form-item :label="t('container.envVars')">
        <div class="dynamic-list">
          <div v-for="(env, index) in form.envs" :key="index" class="list-item">
            <el-input
              v-model="env.key"
              :placeholder="t('container.key')"
              style="width: 150px"
            />
            <span class="separator">=</span>
            <el-input
              v-model="env.value"
              :placeholder="t('container.value')"
              style="width: 200px"
            />
            <el-button type="danger" :icon="Delete" circle size="small" @click="removeEnv(index)" />
          </div>
          <el-button type="primary" size="small" @click="addEnv">
            <el-icon><Plus /></el-icon>
            {{ t('container.addEnv') }}
          </el-button>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="warning" :loading="recreating" @click="handleRecreate">
        {{ t('config.recreate') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { Plus, Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useI18n } from '@/composables/useI18n'
import { containerApi } from '@/api/containers'
import type { ContainerConfig } from '@/types'

const props = defineProps<{
  visible: boolean
  containerId: string
  containerConfig: ContainerConfig | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  recreated: []
}>()

const { t } = useI18n()

// Dialog visibility
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value),
})

// Form state
const form = reactive({
  image: '',
  name: '',
  restart: 'no',
  network: 'bridge',
  ports: [] as Array<{ host: string; container: string }>,
  volumes: [] as Array<{ host: string; container: string }>,
  envs: [] as Array<{ key: string; value: string }>,
})

const recreating = ref(false)

// Initialize form from container config
function initForm() {
  if (!props.containerConfig) return
  
  const config = props.containerConfig
  form.image = config.image || ''
  form.name = config.name || ''
  form.restart = config.restart || 'no'
  form.network = config.networkMode || 'bridge'
  
  // Copy ports
  form.ports = (config.ports || []).map(p => ({
    host: p.host,
    container: p.container,
  }))
  
  // Copy volumes
  form.volumes = (config.volumes || []).map(v => ({
    host: v.host,
    container: v.container,
  }))
  
  // Copy environment variables
  form.envs = (config.env || []).map(e => ({
    key: e.key,
    value: e.value,
  }))
}

// Port management
function addPort() {
  form.ports.push({ host: '', container: '' })
}

function removePort(index: number) {
  form.ports.splice(index, 1)
}

// Volume management
function addVolume() {
  form.volumes.push({ host: '', container: '' })
}

function removeVolume(index: number) {
  form.volumes.splice(index, 1)
}

// Environment variable management
function addEnv() {
  form.envs.push({ key: '', value: '' })
}

function removeEnv(index: number) {
  form.envs.splice(index, 1)
}

// Handle recreate
async function handleRecreate() {
  if (!form.image.trim()) {
    ElMessage.warning(t('container.image') + ' ' + t('common.required'))
    return
  }
  
  if (!form.name.trim()) {
    ElMessage.warning(t('common.name') + ' ' + t('common.required'))
    return
  }
  
  recreating.value = true
  try {
    // Filter out empty entries
    const ports = form.ports.filter(p => p.host && p.container)
    const volumes = form.volumes.filter(v => v.host && v.container)
    const envs = form.envs.filter(e => e.key)
    
    await containerApi.recreate({
      container_id: props.containerId,
      image: form.image,
      name: form.name,
      restart: form.restart,
      network: form.network,
      ports,
      volumes,
      env: envs,
    })
    
    ElMessage.success(t('config.recreateSuccess'))
    emit('recreated')
    dialogVisible.value = false
  } catch {
    // Error handled by interceptor
  } finally {
    recreating.value = false
  }
}

// Handle dialog open
function handleOpen() {
  initForm()
}
</script>

<style scoped>
.dynamic-list {
  width: 100%;
}

.list-item {
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
