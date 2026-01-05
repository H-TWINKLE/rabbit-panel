<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('container.terminal') + ' - ' + containerName"
    width="80%"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    destroy-on-close
    class="terminal-dialog"
    @open="handleOpen"
    @close="handleClose"
  >
    <div class="terminal-container">
      <div ref="terminalRef" class="terminal-wrapper"></div>
      <div v-if="connectionStatus !== 'connected'" class="terminal-overlay">
        <div class="overlay-content">
          <el-icon v-if="connectionStatus === 'connecting'" class="is-loading" :size="32">
            <Loading />
          </el-icon>
          <el-icon v-else-if="connectionStatus === 'disconnected'" :size="32">
            <CircleClose />
          </el-icon>
          <el-icon v-else-if="connectionStatus === 'error'" :size="32" color="#f56c6c">
            <WarningFilled />
          </el-icon>
          <p class="status-text">{{ statusText }}</p>
          <el-button
            v-if="connectionStatus === 'disconnected' || connectionStatus === 'error'"
            type="primary"
            @click="reconnect"
          >
            {{ t('container.reconnect') }}
          </el-button>
        </div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted, nextTick } from 'vue'
import { Loading, CircleClose, WarningFilled } from '@element-plus/icons-vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'
import { useI18n } from '@/composables/useI18n'
import { containerApi } from '@/api/containers'

type ConnectionStatus = 'idle' | 'connecting' | 'connected' | 'disconnected' | 'error'

const props = defineProps<{
  visible: boolean
  containerId: string
  containerName: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const { t } = useI18n()

// Dialog visibility
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value),
})

// Terminal refs
const terminalRef = ref<HTMLElement | null>(null)
let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let ws: WebSocket | null = null
let resizeObserver: ResizeObserver | null = null

// Connection status
const connectionStatus = ref<ConnectionStatus>('idle')
const errorMessage = ref('')

// Status text
const statusText = computed(() => {
  switch (connectionStatus.value) {
    case 'connecting':
      return t('container.terminalConnecting')
    case 'disconnected':
      return t('container.terminalDisconnected')
    case 'error':
      return errorMessage.value || t('container.terminalError')
    default:
      return ''
  }
})

// Initialize terminal
function initTerminal() {
  if (!terminalRef.value) return

  terminal = new Terminal({
    cursorBlink: true,
    cursorStyle: 'block',
    fontSize: 14,
    fontFamily: 'Menlo, Monaco, "Courier New", monospace',
    theme: {
      background: '#1e1e1e',
      foreground: '#d4d4d4',
      cursor: '#d4d4d4',
      cursorAccent: '#1e1e1e',
      selectionBackground: '#264f78',
      black: '#000000',
      red: '#cd3131',
      green: '#0dbc79',
      yellow: '#e5e510',
      blue: '#2472c8',
      magenta: '#bc3fbc',
      cyan: '#11a8cd',
      white: '#e5e5e5',
      brightBlack: '#666666',
      brightRed: '#f14c4c',
      brightGreen: '#23d18b',
      brightYellow: '#f5f543',
      brightBlue: '#3b8eea',
      brightMagenta: '#d670d6',
      brightCyan: '#29b8db',
      brightWhite: '#e5e5e5',
    },
    allowProposedApi: true,
  })

  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(new WebLinksAddon())

  terminal.open(terminalRef.value)
  fitAddon.fit()

  // Handle terminal input
  terminal.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      // 直接发送原始输入数据，不包装成 JSON
      ws.send(data)
    }
  })

  // Handle terminal resize
  terminal.onResize(({ cols, rows }) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'resize', cols, rows }))
    }
  })

  // Setup resize observer
  resizeObserver = new ResizeObserver(() => {
    if (fitAddon) {
      fitAddon.fit()
    }
  })
  resizeObserver.observe(terminalRef.value)
}

// Connect to WebSocket
function connect() {
  if (!props.containerId) return

  connectionStatus.value = 'connecting'
  errorMessage.value = ''

  const wsUrl = containerApi.getTerminalWsUrl(props.containerId)
  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    connectionStatus.value = 'connected'
    terminal?.focus()
    
    // 延迟发送 resize，等待后端完成 exec attach
    setTimeout(() => {
      if (terminal && ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'resize',
          cols: terminal.cols,
          rows: terminal.rows,
        }))
      }
    }, 100)
  }

  ws.onmessage = (event) => {
    // 后端发送的是二进制数据，直接写入终端
    if (event.data instanceof Blob) {
      event.data.text().then(text => {
        terminal?.write(text)
      })
    } else if (typeof event.data === 'string') {
      // 尝试解析 JSON 错误消息
      try {
        const message = JSON.parse(event.data)
        if (message.type === 'error') {
          errorMessage.value = message.data || t('container.terminalError')
          connectionStatus.value = 'error'
          return
        }
      } catch {
        // 不是 JSON，当作普通文本输出
      }
      terminal?.write(event.data)
    }
  }

  ws.onerror = (error) => {
    console.error('[Terminal] WebSocket error:', error)
    errorMessage.value = t('container.terminalError')
    connectionStatus.value = 'error'
  }

  ws.onclose = (event) => {
    console.log('[Terminal] WebSocket closed:', event.code, event.reason)
    if (connectionStatus.value === 'connected') {
      connectionStatus.value = 'disconnected'
    }
    if (!event.wasClean && connectionStatus.value !== 'error') {
      errorMessage.value = t('container.terminalDisconnected')
      connectionStatus.value = 'disconnected'
    }
  }
}

// Disconnect WebSocket
function disconnect() {
  if (ws) {
    ws.close()
    ws = null
  }
}

// Reconnect
function reconnect() {
  disconnect()
  terminal?.clear()
  connect()
}

// Handle dialog open
async function handleOpen() {
  await nextTick()
  initTerminal()
  connect()
}

// Handle dialog close
function handleClose() {
  disconnect()
  
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }
  
  if (terminal) {
    terminal.dispose()
    terminal = null
  }
  
  fitAddon = null
  connectionStatus.value = 'idle'
}

// Watch for container ID changes
watch(() => props.containerId, (newId, oldId) => {
  if (newId !== oldId && props.visible) {
    disconnect()
    terminal?.clear()
    connect()
  }
})

// Cleanup on unmount
onUnmounted(() => {
  handleClose()
})
</script>

<style scoped>
.terminal-dialog :deep(.el-dialog__body) {
  padding: 0;
}

.terminal-container {
  position: relative;
  height: 500px;
  background: #1e1e1e;
}

.terminal-wrapper {
  width: 100%;
  height: 100%;
  padding: 10px;
  box-sizing: border-box;
}

.terminal-wrapper :deep(.xterm) {
  height: 100%;
}

.terminal-wrapper :deep(.xterm-viewport) {
  overflow-y: auto !important;
}

.terminal-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(30, 30, 30, 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}

.overlay-content {
  text-align: center;
  color: #d4d4d4;
}

.status-text {
  margin: 15px 0;
  font-size: 14px;
}

.is-loading {
  animation: rotating 2s linear infinite;
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
