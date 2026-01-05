<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('container.config') + ' - ' + containerName"
    width="700px"
    :close-on-click-modal="false"
    destroy-on-close
    class="container-config-dialog"
    @open="handleOpen"
  >
    <div v-loading="loading" class="container-config">
      <el-tabs v-model="activeTab">
        <!-- Basic Info Tab -->
        <el-tab-pane :label="t('config.basicInfo')" name="basic">
          <el-descriptions :column="2" border>
            <el-descriptions-item :label="t('config.containerId')">
              {{ config?.id }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('config.fullId')">
              <el-tooltip :content="config?.fullId" placement="top">
                <span class="truncate">{{ config?.fullId }}</span>
              </el-tooltip>
            </el-descriptions-item>
            <el-descriptions-item :label="t('common.name')">
              <div class="editable-field">
                <span v-if="!editingName">{{ config?.name }}</span>
                <el-input v-else v-model="newName" size="small" style="width: 200px" />
                <el-button
                  v-if="!editingName"
                  type="primary"
                  size="small"
                  :icon="Edit"
                  circle
                  @click="startEditName"
                />
                <template v-else>
                  <el-button type="success" size="small" :icon="Check" circle @click="handleRename" />
                  <el-button type="info" size="small" :icon="Close" circle @click="cancelEditName" />
                </template>
              </div>
            </el-descriptions-item>
            <el-descriptions-item :label="t('container.image')">
              {{ config?.image }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('common.status')">
              <el-tag :type="config?.running ? 'success' : 'danger'">
                {{ config?.state }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('config.platform')">
              {{ config?.platform }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('config.created')">
              {{ formatTime(config?.created) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('config.started')">
              {{ formatTime(config?.started) }}
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>

        <!-- Network Tab -->
        <el-tab-pane :label="t('config.network')" name="network">
          <el-descriptions :column="2" border>
            <el-descriptions-item :label="t('config.networkMode')">
              {{ config?.networkMode }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('config.hostname')">
              {{ config?.hostname }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('config.ipAddress')">
              {{ config?.ipAddress || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('config.gateway')">
              {{ config?.gateway || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('config.macAddress')">
              {{ config?.macAddress || '-' }}
            </el-descriptions-item>
          </el-descriptions>
          
          <h4>{{ t('container.ports') }}</h4>
          <el-table :data="config?.ports || []" stripe size="small">
            <el-table-column prop="hostIP" :label="t('config.hostIP')" />
            <el-table-column prop="host" :label="t('container.hostPort')" />
            <el-table-column prop="container" :label="t('container.containerPort')" />
          </el-table>
        </el-tab-pane>

        <!-- Storage Tab -->
        <el-tab-pane :label="t('config.storage')" name="storage">
          <h4>{{ t('container.volumes') }}</h4>
          <el-table :data="config?.volumes || []" stripe size="small">
            <el-table-column prop="host" :label="t('container.hostPath')" />
            <el-table-column prop="container" :label="t('container.containerPath')" />
            <el-table-column prop="mode" :label="t('config.mode')" width="100" />
          </el-table>
          
          <el-descriptions :column="2" border style="margin-top: 20px">
            <el-descriptions-item :label="t('config.workingDir')">
              {{ config?.workingDir || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('config.readOnly')">
              {{ config?.readOnly ? t('common.yes') : t('common.no') }}
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>

        <!-- Environment Tab -->
        <el-tab-pane :label="t('container.envVars')" name="env">
          <el-table :data="config?.env || []" stripe size="small">
            <el-table-column prop="key" :label="t('container.key')" />
            <el-table-column prop="value" :label="t('container.value')" />
          </el-table>
        </el-tab-pane>

        <!-- Resources Tab -->
        <el-tab-pane :label="t('config.resources')" name="resources">
          <el-form label-width="150px">
            <el-form-item :label="t('config.memory')">
              <div class="resource-field">
                <el-input-number
                  v-model="resourceForm.memory"
                  :min="0"
                  :step="64"
                  controls-position="right"
                />
                <span class="unit">MB (0 = {{ t('config.unlimited') }})</span>
                <el-button type="primary" size="small" @click="updateMemory">
                  {{ t('config.apply') }}
                </el-button>
              </div>
            </el-form-item>
            <el-form-item :label="t('config.cpus')">
              <div class="resource-field">
                <el-input-number
                  v-model="resourceForm.cpus"
                  :min="0"
                  :max="64"
                  :step="0.1"
                  :precision="1"
                  controls-position="right"
                />
                <span class="unit">{{ t('config.cores') }} (0 = {{ t('config.unlimited') }})</span>
                <el-button type="primary" size="small" @click="updateCpus">
                  {{ t('config.apply') }}
                </el-button>
              </div>
            </el-form-item>
            <el-form-item :label="t('container.restartPolicy')">
              <div class="resource-field">
                <el-select v-model="resourceForm.restart" style="width: 200px">
                  <el-option value="no" :label="t('container.restartPolicies.no')" />
                  <el-option value="always" :label="t('container.restartPolicies.always')" />
                  <el-option value="on-failure" :label="t('container.restartPolicies.onFailure')" />
                  <el-option value="unless-stopped" :label="t('container.restartPolicies.unlessStopped')" />
                </el-select>
                <el-button type="primary" size="small" @click="updateRestart">
                  {{ t('config.apply') }}
                </el-button>
              </div>
            </el-form-item>
          </el-form>
          
          <el-descriptions :column="2" border style="margin-top: 20px">
            <el-descriptions-item :label="t('config.cpuShares')">
              {{ config?.cpuShares }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('config.pidsLimit')">
              {{ config?.pidsLimit === 0 ? t('config.unlimited') : config?.pidsLimit }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('config.oomKillDisable')">
              {{ config?.oomKillDisable ? t('common.yes') : t('common.no') }}
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>

        <!-- Security Tab -->
        <el-tab-pane :label="t('config.security')" name="security">
          <el-descriptions :column="2" border>
            <el-descriptions-item :label="t('config.privileged')">
              <el-tag :type="config?.privileged ? 'danger' : 'success'">
                {{ config?.privileged ? t('common.yes') : t('common.no') }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('config.user')">
              {{ config?.user || 'root' }}
            </el-descriptions-item>
          </el-descriptions>
          
          <h4>{{ t('config.capAdd') }}</h4>
          <div class="tag-list">
            <el-tag v-for="cap in config?.capAdd" :key="cap" type="success" size="small">
              {{ cap }}
            </el-tag>
            <span v-if="!config?.capAdd?.length">-</span>
          </div>
          
          <h4>{{ t('config.capDrop') }}</h4>
          <div class="tag-list">
            <el-tag v-for="cap in config?.capDrop" :key="cap" type="danger" size="small">
              {{ cap }}
            </el-tag>
            <span v-if="!config?.capDrop?.length">-</span>
          </div>
        </el-tab-pane>

        <!-- Command Tab -->
        <el-tab-pane :label="t('config.command')" name="command">
          <el-descriptions :column="1" border>
            <el-descriptions-item :label="t('config.entrypoint')">
              <code>{{ config?.entrypoint?.join(' ') || '-' }}</code>
            </el-descriptions-item>
            <el-descriptions-item :label="t('config.cmd')">
              <code>{{ config?.cmd?.join(' ') || '-' }}</code>
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
      </el-tabs>
    </div>
    
    <template #footer>
      <el-button type="warning" @click="showRecreateDialog = true">
        {{ t('config.recreate') }}
      </el-button>
      <el-button @click="dialogVisible = false">{{ t('common.close') }}</el-button>
    </template>

    <!-- Recreate Container Dialog -->
    <RecreateContainerDialog
      v-model:visible="showRecreateDialog"
      :container-id="containerId"
      :container-config="config"
      @recreated="handleRecreated"
    />
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { Edit, Check, Close } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useI18n } from '@/composables/useI18n'
import { containerApi } from '@/api/containers'
import type { ContainerConfig } from '@/types'
import RecreateContainerDialog from './RecreateContainerDialog.vue'

const props = defineProps<{
  visible: boolean
  containerId: string
  containerName: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  updated: []
}>()

const { t } = useI18n()

// Dialog visibility
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value),
})

// State
const loading = ref(false)
const activeTab = ref('basic')
const config = ref<ContainerConfig | null>(null)
const editingName = ref(false)
const newName = ref('')
const showRecreateDialog = ref(false)

// Resource form
const resourceForm = reactive({
  memory: 0,
  cpus: 0,
  restart: 'no',
})

// Load container config
async function loadConfig() {
  if (!props.containerId) return
  
  loading.value = true
  try {
    config.value = await containerApi.inspect(props.containerId)
    // Initialize resource form
    resourceForm.memory = Math.round((config.value.memory || 0) / 1024 / 1024)
    resourceForm.cpus = config.value.cpus || 0
    resourceForm.restart = config.value.restart || 'no'
  } catch {
    ElMessage.error(t('config.loadError'))
    dialogVisible.value = false
  } finally {
    loading.value = false
  }
}

// Name editing
function startEditName() {
  newName.value = config.value?.name || ''
  editingName.value = true
}

function cancelEditName() {
  editingName.value = false
  newName.value = ''
}

async function handleRename() {
  if (!newName.value.trim()) {
    ElMessage.warning(t('config.nameRequired'))
    return
  }
  
  try {
    await containerApi.rename(props.containerId, newName.value)
    ElMessage.success(t('config.renameSuccess'))
    if (config.value) {
      config.value.name = newName.value
    }
    editingName.value = false
    emit('updated')
  } catch {
    // Error handled by interceptor
  }
}

// Resource updates
async function updateMemory() {
  try {
    await containerApi.update(props.containerId, { memory: resourceForm.memory * 1024 * 1024 })
    ElMessage.success(t('config.updateSuccess'))
    emit('updated')
  } catch {
    // Error handled by interceptor
  }
}

async function updateCpus() {
  try {
    await containerApi.update(props.containerId, { cpus: resourceForm.cpus })
    ElMessage.success(t('config.updateSuccess'))
    emit('updated')
  } catch {
    // Error handled by interceptor
  }
}

async function updateRestart() {
  try {
    await containerApi.update(props.containerId, { restart: resourceForm.restart })
    ElMessage.success(t('config.updateSuccess'))
    emit('updated')
  } catch {
    // Error handled by interceptor
  }
}

// Handle recreated
function handleRecreated() {
  emit('updated')
  dialogVisible.value = false
}

// Format time
function formatTime(time?: string): string {
  if (!time) return '-'
  return new Date(time).toLocaleString()
}

// Handle dialog open
function handleOpen() {
  activeTab.value = 'basic'
  editingName.value = false
  loadConfig()
}
</script>

<style scoped>
.container-config {
  min-height: 400px;
}

.editable-field {
  display: flex;
  align-items: center;
  gap: 8px;
}

.truncate {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
}

.resource-field {
  display: flex;
  align-items: center;
  gap: 10px;
}

.unit {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 15px;
}

h4 {
  margin: 20px 0 10px;
  font-size: 14px;
  font-weight: 600;
}

code {
  background: var(--el-fill-color-light);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
}
</style>
