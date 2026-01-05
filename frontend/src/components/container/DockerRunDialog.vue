<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('container.dockerRun')"
    width="700px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <div class="docker-run-dialog">
      <!-- Command Input -->
      <el-form label-position="top">
        <el-form-item>
          <el-input
            v-model="command"
            type="textarea"
            :rows="4"
            :placeholder="t('container.commandPlaceholder')"
            :disabled="isExecuting"
          />
        </el-form-item>
        <el-form-item>
          <div class="help-text">{{ t('container.commandHelp') }}</div>
        </el-form-item>
      </el-form>

      <!-- Parsed Preview -->
      <el-collapse v-if="parsedCommand" v-model="activeCollapse">
        <el-collapse-item title="Preview" name="preview">
          <div class="parsed-preview">
            <div v-if="parsedCommand.image" class="preview-item">
              <span class="label">{{ t('container.image') }}:</span>
              <span class="value">{{ parsedCommand.image }}</span>
            </div>
            <div v-if="parsedCommand.name" class="preview-item">
              <span class="label">{{ t('container.containerName') }}:</span>
              <span class="value">{{ parsedCommand.name }}</span>
            </div>
            <div v-if="parsedCommand.ports.length" class="preview-item">
              <span class="label">{{ t('container.ports') }}:</span>
              <span class="value">{{ parsedCommand.ports.map(p => `${p.host}:${p.container}`).join(', ') }}</span>
            </div>
            <div v-if="parsedCommand.volumes.length" class="preview-item">
              <span class="label">{{ t('container.volumes') }}:</span>
              <span class="value">{{ parsedCommand.volumes.map(v => `${v.host}:${v.container}`).join(', ') }}</span>
            </div>
            <div v-if="parsedCommand.envs.length" class="preview-item">
              <span class="label">{{ t('container.envVars') }}:</span>
              <span class="value">{{ parsedCommand.envs.map(e => `${e.key}=${e.value}`).join(', ') }}</span>
            </div>
            <div v-if="parsedCommand.restart" class="preview-item">
              <span class="label">{{ t('container.restartPolicy') }}:</span>
              <span class="value">{{ parsedCommand.restart }}</span>
            </div>
            <div v-if="parsedCommand.network" class="preview-item">
              <span class="label">{{ t('container.network') }}:</span>
              <span class="value">{{ parsedCommand.network }}</span>
            </div>
          </div>
        </el-collapse-item>
      </el-collapse>

      <!-- Output -->
      <div v-if="output.length > 0" class="output-section">
        <div class="output-header">{{ t('container.output') }}</div>
        <div ref="outputRef" class="output-content">
          <div
            v-for="(line, index) in output"
            :key="index"
            :class="['output-line', line.type]"
          >
            {{ line.message }}
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <el-button @click="handleClose" :disabled="isExecuting">
        {{ t('common.cancel') }}
      </el-button>
      <el-button
        type="primary"
        :loading="isExecuting"
        :disabled="!command.trim()"
        @click="handleExecute"
      >
        {{ t('container.execute') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from '@/composables/useI18n'
import { containerApi } from '@/api/containers'

const { t } = useI18n()

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

const command = ref('')
const isExecuting = ref(false)
const output = ref<Array<{ type: string; message: string }>>([])
const outputRef = ref<HTMLElement>()
const activeCollapse = ref<string[]>([])

interface ParsedCommand {
  image: string
  name: string
  ports: Array<{ host: string; container: string }>
  volumes: Array<{ host: string; container: string }>
  envs: Array<{ key: string; value: string }>
  restart: string
  network: string
}

// Parse docker run command
const parsedCommand = computed((): ParsedCommand | null => {
  const cmd = command.value.trim()
  if (!cmd || !cmd.startsWith('docker run')) {
    return null
  }

  const result: ParsedCommand = {
    image: '',
    name: '',
    ports: [],
    volumes: [],
    envs: [],
    restart: '',
    network: '',
  }

  // Simple parser for docker run command
  const args = parseCommandArgs(cmd)
  
  // Collect all non-flag arguments (potential image name)
  const nonFlagArgs: string[] = []
  
  let i = 2 // Skip 'docker' and 'run'
  while (i < args.length) {
    const arg = args[i]
    if (!arg) {
      i++
      continue
    }
    
    if (arg === '-d' || arg === '--detach' || arg === '-it' || arg === '-i' || arg === '-t' || arg === '--rm') {
      i++
      continue
    }
    
    if (arg === '--name' && i + 1 < args.length) {
      result.name = args[i + 1] || ''
      i += 2
      continue
    }
    
    if ((arg === '-p' || arg === '--publish') && i + 1 < args.length) {
      const portMapping = args[i + 1] || ''
      const parts = portMapping.split(':')
      if (parts.length >= 2) {
        const hostPart = parts[0] || ''
        const containerPart = parts[parts.length - 1] || ''
        result.ports.push({
          host: hostPart,
          container: containerPart.split('/')[0] || '',
        })
      }
      i += 2
      continue
    }
    
    if ((arg === '-v' || arg === '--volume') && i + 1 < args.length) {
      const volumeMapping = args[i + 1] || ''
      const parts = volumeMapping.split(':')
      if (parts.length >= 2) {
        result.volumes.push({
          host: parts[0] || '',
          container: parts[1] || '',
        })
      }
      i += 2
      continue
    }
    
    if ((arg === '-e' || arg === '--env') && i + 1 < args.length) {
      const envVar = args[i + 1] || ''
      const eqIndex = envVar.indexOf('=')
      if (eqIndex > 0) {
        result.envs.push({
          key: envVar.substring(0, eqIndex),
          value: envVar.substring(eqIndex + 1),
        })
      }
      i += 2
      continue
    }
    
    if (arg === '--restart' && i + 1 < args.length) {
      result.restart = args[i + 1] || ''
      i += 2
      continue
    }
    
    if ((arg === '--network' || arg === '--net') && i + 1 < args.length) {
      result.network = args[i + 1] || ''
      i += 2
      continue
    }
    
    // Handle flags with = syntax (e.g., --name=xxx)
    if (arg.startsWith('--') && arg.includes('=')) {
      const eqIndex = arg.indexOf('=')
      const flagName = arg.substring(2, eqIndex)
      const flagValue = arg.substring(eqIndex + 1)
      
      if (flagName === 'name') {
        result.name = flagValue
      } else if (flagName === 'restart') {
        result.restart = flagValue
      } else if (flagName === 'network' || flagName === 'net') {
        result.network = flagValue
      }
      i++
      continue
    }
    
    // Skip other flags with values
    const nextArg = args[i + 1]
    if (arg.startsWith('-') && arg.length === 2 && i + 1 < args.length && nextArg && !nextArg.startsWith('-')) {
      // Short flag with value (e.g., -u root)
      i += 2
      continue
    }
    
    if (arg.startsWith('--') && i + 1 < args.length && nextArg && !nextArg.startsWith('-')) {
      // Long flag with value
      i += 2
      continue
    }
    
    // Skip standalone flags
    if (arg.startsWith('-')) {
      i++
      continue
    }
    
    // This is a non-flag argument (could be image name or command)
    nonFlagArgs.push(arg)
    i++
  }

  // The first non-flag argument is the image name
  // (subsequent ones would be the command to run in the container)
  if (nonFlagArgs.length > 0) {
    result.image = nonFlagArgs[0] || ''
  }

  return result.image ? result : null
})

// Parse command line arguments respecting quotes and line continuations
function parseCommandArgs(cmd: string): string[] {
  // First, handle line continuations (backslash followed by newline)
  let normalizedCmd = cmd
    .replace(/\\\r?\n/g, ' ')  // Replace backslash + newline with space
    .replace(/\s+/g, ' ')       // Normalize multiple spaces to single space
    .trim()

  const args: string[] = []
  let current = ''
  let inQuote = false
  let quoteChar = ''

  for (let i = 0; i < normalizedCmd.length; i++) {
    const char = normalizedCmd[i]

    if (inQuote) {
      if (char === quoteChar) {
        inQuote = false
      } else {
        current += char
      }
    } else if (char === '"' || char === "'") {
      inQuote = true
      quoteChar = char
    } else if (char === ' ' || char === '\t') {
      if (current) {
        args.push(current)
        current = ''
      }
    } else {
      current += char
    }
  }

  if (current) {
    args.push(current)
  }

  return args
}

// Reset when dialog opens
watch(() => props.visible, (val) => {
  if (val) {
    command.value = ''
    output.value = []
    isExecuting.value = false
    activeCollapse.value = []
  }
})

// Auto-scroll output
watch(output, async () => {
  await nextTick()
  if (outputRef.value) {
    outputRef.value.scrollTop = outputRef.value.scrollHeight
  }
}, { deep: true })

function handleClose() {
  if (!isExecuting.value) {
    dialogVisible.value = false
  }
}

async function handleExecute() {
  const cmd = command.value.trim()
  if (!cmd) return

  if (!cmd.startsWith('docker run')) {
    ElMessage.error('Only docker run commands are supported')
    return
  }

  isExecuting.value = true
  output.value = []

  try {
    await containerApi.createRawStream(
      cmd,
      (data) => {
        output.value.push({
          type: data.type || 'log',
          message: data.message || data.container_id || '',
        })
      },
      (error) => {
        output.value.push({ type: 'error', message: error })
      },
      () => {
        isExecuting.value = false
        // Check if last message was success
        const lastOutput = output.value[output.value.length - 1]
        if (lastOutput?.type === 'success') {
          ElMessage.success(t('container.createSuccess'))
          emit('created')
          // Close dialog after short delay
          setTimeout(() => {
            dialogVisible.value = false
          }, 1000)
        }
      }
    )
  } catch {
    isExecuting.value = false
  }
}
</script>

<style scoped>
.docker-run-dialog {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.help-text {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.parsed-preview {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.preview-item {
  display: flex;
  gap: 10px;
}

.preview-item .label {
  font-weight: 500;
  color: var(--el-text-color-secondary);
  min-width: 100px;
}

.preview-item .value {
  color: var(--el-text-color-primary);
  word-break: break-all;
}

.output-section {
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  overflow: hidden;
}

.output-header {
  background: var(--el-fill-color-light);
  padding: 8px 12px;
  font-weight: 500;
  border-bottom: 1px solid var(--el-border-color);
}

.output-content {
  max-height: 200px;
  overflow-y: auto;
  padding: 10px;
  background: var(--el-bg-color);
  font-family: monospace;
  font-size: 12px;
}

.output-line {
  padding: 2px 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.output-line.error {
  color: var(--el-color-danger);
}

.output-line.success {
  color: var(--el-color-success);
}

.output-line.log {
  color: var(--el-text-color-regular);
}
</style>
