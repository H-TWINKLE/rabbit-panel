<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('container.create')"
    width="700px"
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
      <!-- Node Selection (only shown in Master mode) -->
      <el-form-item v-if="nodesStore.isMasterMode" :label="t('node.selectNode')" prop="nodeId">
        <el-select
          v-model="form.nodeId"
          :placeholder="t('node.autoSelect')"
          clearable
          style="width: 100%"
        >
          <el-option value="" :label="t('node.autoSelect')">
            <div class="node-option">
              <span>{{ t('node.autoSelect') }}</span>
              <el-tag size="small" type="info">{{ t('node.autoSelect') }}</el-tag>
            </div>
          </el-option>
          <el-option
            v-for="node in nodesStore.onlineNodes"
            :key="node.id"
            :value="node.id"
            :label="node.name"
          >
            <div class="node-option">
              <span>{{ node.name }}</span>
              <span class="node-info">
                <el-tag size="small" type="success">{{ t('node.online') }}</el-tag>
                <span class="node-stats">
                  CPU: {{ node.cpu.toFixed(1) }}% | 
                  {{ t('node.memory') }}: {{ node.memory.toFixed(1) }}%
                </span>
              </span>
            </div>
          </el-option>
        </el-select>
        <div class="node-hint">
          <el-icon><InfoFilled /></el-icon>
          <span>{{ nodesStore.onlineNodes.length }} {{ t('node.onlineNodes') }}</span>
        </div>
      </el-form-item>

      <!-- Image -->
      <el-form-item :label="t('container.image')" prop="image">
        <el-input v-model="form.image" placeholder="nginx:latest" />
      </el-form-item>

      <!-- Container Name -->
      <el-form-item :label="t('container.containerName')" prop="name">
        <el-input v-model="form.name" placeholder="my-container" />
      </el-form-item>

      <!-- Restart Policy -->
      <el-form-item :label="t('container.restartPolicy')" prop="restart">
        <el-select v-model="form.restart" style="width: 100%">
          <el-option value="no" :label="t('container.restartPolicies.no')" />
          <el-option value="always" :label="t('container.restartPolicies.always')" />
          <el-option value="on-failure" :label="t('container.restartPolicies.onFailure')" />
          <el-option value="unless-stopped" :label="t('container.restartPolicies.unlessStopped')" />
        </el-select>
      </el-form-item>

      <!-- Network -->
      <el-form-item :label="t('container.network')" prop="network">
        <el-input v-model="form.network" placeholder="bridge" />
      </el-form-item>

      <!-- Port Mappings -->
      <el-form-item :label="t('container.ports')">
        <div class="dynamic-list">
          <div v-for="(port, index) in form.ports" :key="index" class="dynamic-item">
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
            <el-button
              type="danger"
              :icon="Delete"
              circle
              size="small"
              @click="removePort(index)"
            />
          </div>
          <el-button type="primary" text @click="addPort">
            <el-icon><Plus /></el-icon>
            {{ t('container.addPort') }}
          </el-button>
        </div>
      </el-form-item>

      <!-- Volume Mappings -->
      <el-form-item :label="t('container.volumes')">
        <div class="dynamic-list">
          <div v-for="(volume, index) in form.volumes" :key="index" class="dynamic-item">
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
            <el-button
              type="danger"
              :icon="Delete"
              circle
              size="small"
              @click="removeVolume(index)"
            />
          </div>
          <el-button type="primary" text @click="addVolume">
            <el-icon><Plus /></el-icon>
            {{ t('container.addVolume') }}
          </el-button>
        </div>
      </el-form-item>

      <!-- Environment Variables -->
      <el-form-item :label="t('container.envVars')">
        <div class="dynamic-list">
          <div v-for="(env, index) in form.envs" :key="index" class="dynamic-item">
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
            <el-button
              type="danger"
              :icon="Delete"
              circle
              size="small"
              @click="removeEnv(index)"
            />
          </div>
          <el-button type="primary" text @click="addEnv">
            <el-icon><Plus /></el-icon>
            {{ t('container.addEnv') }}
          </el-button>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleClose">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">
        {{ t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { Plus, Delete, InfoFilled } from '@element-plus/icons-vue'
import { useI18n } from '@/composables/useI18n'
import { containerApi } from '@/api/containers'
import { useNodesStore } from '@/stores/nodes'

const { t } = useI18n()
const nodesStore = useNodesStore()

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  created: []
}>()

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
})

const formRef = ref<FormInstance>()
const loading = ref(false)

interface PortMapping {
  host: string
  container: string
}

interface VolumeMapping {
  host: string
  container: string
}

interface EnvVar {
  key: string
  value: string
}

interface FormData {
  image: string
  name: string
  restart: string
  network: string
  ports: PortMapping[]
  volumes: VolumeMapping[]
  envs: EnvVar[]
  nodeId: string
}

const initialForm = (): FormData => ({
  image: '',
  name: '',
  restart: 'no',
  network: 'bridge',
  ports: [],
  volumes: [],
  envs: [],
  nodeId: '',
})

const form = reactive<FormData>(initialForm())

const rules: FormRules = {
  image: [
    { required: true, message: () => t('container.image') + ' is required', trigger: 'blur' },
  ],
}

// Reset form when dialog opens
watch(() => props.visible, (val) => {
  if (val) {
    Object.assign(form, initialForm())
    formRef.value?.resetFields()
    // Fetch nodes when dialog opens (to check if in master mode)
    nodesStore.fetchNodes()
  }
})

// Port mapping methods
function addPort() {
  form.ports.push({ host: '', container: '' })
}

function removePort(index: number) {
  form.ports.splice(index, 1)
}

// Volume mapping methods
function addVolume() {
  form.volumes.push({ host: '', container: '' })
}

function removeVolume(index: number) {
  form.volumes.splice(index, 1)
}

// Environment variable methods
function addEnv() {
  form.envs.push({ key: '', value: '' })
}

function removeEnv(index: number) {
  form.envs.splice(index, 1)
}

function handleClose() {
  dialogVisible.value = false
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
    // Filter out empty entries
    const ports = form.ports.filter((p) => p.host && p.container)
    const volumes = form.volumes.filter((v) => v.host && v.container)
    const envs = form.envs.filter((e) => e.key)

    // Check if we should schedule to a node (Master mode)
    if (nodesStore.isMasterMode) {
      // Convert ports and envs to the format expected by schedule API
      const portsMap: Record<string, string> = {}
      ports.forEach((p) => {
        portsMap[p.host] = `${p.host}:${p.container}`
      })

      const envsMap: Record<string, string> = {}
      envs.forEach((e) => {
        envsMap[e.key] = e.value
      })

      const result = await nodesStore.scheduleContainer({
        image: form.image,
        name: form.name,
        ports: portsMap,
        env: envsMap,
        node_id: form.nodeId || undefined, // Empty string means auto-select
      })

      ElMessage.success(`${t('node.scheduleSuccess')} - ${result.node}`)
    } else {
      // Local container creation
      await containerApi.create({
        image: form.image,
        name: form.name,
        restart: form.restart,
        network: form.network,
        ports,
        volumes,
        env: envs,
      })

      ElMessage.success(t('container.createSuccess'))
    }

    emit('created')
    handleClose()
  } catch {
    // Error handled by request interceptor
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
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

.node-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.node-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.node-stats {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.node-hint {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
