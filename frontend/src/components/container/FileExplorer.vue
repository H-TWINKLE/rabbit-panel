<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('container.files') + ' - ' + containerName"
    width="80%"
    :close-on-click-modal="false"
    destroy-on-close
    class="file-explorer-dialog"
    @open="handleOpen"
  >
    <div class="file-explorer">
      <!-- Toolbar -->
      <div class="toolbar">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item
            v-for="(segment, index) in pathSegments"
            :key="index"
            @click="navigateToSegment(index)"
          >
            <span class="breadcrumb-item" :class="{ clickable: index < pathSegments.length - 1 }">
              {{ segment || '/' }}
            </span>
          </el-breadcrumb-item>
        </el-breadcrumb>
        <div class="toolbar-actions">
          <el-button size="small" @click="handleRefresh">
            <el-icon><Refresh /></el-icon>
          </el-button>
          <el-button size="small" type="primary" @click="showCreateDirDialog = true">
            <el-icon><FolderAdd /></el-icon>
            {{ t('file.createDir') }}
          </el-button>
          <el-button size="small" type="success" @click="triggerUpload">
            <el-icon><Upload /></el-icon>
            {{ t('file.upload') }}
          </el-button>
          <input
            ref="fileInputRef"
            type="file"
            style="display: none"
            @change="handleFileSelect"
          />
        </div>
      </div>

      <!-- File table -->
      <el-table
        v-loading="loading"
        :data="files"
        stripe
        style="width: 100%"
        @row-dblclick="handleRowDblClick"
      >
        <el-table-column width="50">
          <template #default="{ row }">
            <el-icon :size="20" :color="row.is_dir ? '#e6a23c' : '#909399'">
              <Folder v-if="row.is_dir" />
              <Document v-else />
            </el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="name" :label="t('common.name')" min-width="200">
          <template #default="{ row }">
            <span class="file-name" @click="handleFileClick(row)">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="size" :label="t('common.size')" width="120">
          <template #default="{ row }">
            {{ row.is_dir ? '-' : formatSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="mode" :label="t('file.permissions')" width="120" />
        <el-table-column prop="mod_time" :label="t('file.modTime')" width="180">
          <template #default="{ row }">
            {{ formatTime(row.mod_time) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="!row.is_dir"
              type="primary"
              size="small"
              :icon="Download"
              circle
              @click="handleDownload(row)"
            />
            <el-button
              v-if="!row.is_dir && isTextFile(row.name)"
              type="info"
              size="small"
              :icon="Edit"
              circle
              @click="handleEdit(row)"
            />
            <el-popconfirm
              :title="t('file.confirmDelete')"
              :confirm-button-text="t('common.confirm')"
              :cancel-button-text="t('common.cancel')"
              @confirm="handleDelete(row)"
            >
              <template #reference>
                <el-button
                  type="danger"
                  size="small"
                  :icon="Delete"
                  circle
                  :disabled="isProtectedPath(row.path)"
                />
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- Create Directory Dialog -->
    <el-dialog
      v-model="showCreateDirDialog"
      :title="t('file.createDir')"
      width="400px"
      append-to-body
    >
      <el-form @submit.prevent="handleCreateDir">
        <el-form-item :label="t('file.dirName')">
          <el-input v-model="newDirName" :placeholder="t('file.dirNamePlaceholder')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDirDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreateDir">
          {{ t('common.create') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- File Editor Dialog -->
    <FileEditor
      v-model:visible="showEditorDialog"
      :container-id="containerId"
      :file-path="editingFilePath"
      :file-name="editingFileName"
      @saved="handleRefresh"
    />
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  Refresh,
  FolderAdd,
  Upload,
  Folder,
  Document,
  Download,
  Edit,
  Delete,
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useI18n } from '@/composables/useI18n'
import { containerApi } from '@/api/containers'
import type { FileInfo } from '@/types'
import FileEditor from './FileEditor.vue'

const MAX_UPLOAD_SIZE = 10 * 1024 * 1024 // 10MB

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

// State
const currentPath = ref('/')
const files = ref<FileInfo[]>([])
const loading = ref(false)
const creating = ref(false)
const showCreateDirDialog = ref(false)
const showEditorDialog = ref(false)
const newDirName = ref('')
const editingFilePath = ref('')
const editingFileName = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)

// Protected system paths
const PROTECTED_PATHS = ['/', '/bin', '/sbin', '/usr', '/lib', '/lib64', '/etc', '/var', '/root', '/home']

// Computed
const pathSegments = computed(() => {
  const segments = currentPath.value.split('/').filter(Boolean)
  return ['', ...segments]
})

// Methods
async function fetchFiles() {
  if (!props.containerId) return
  
  loading.value = true
  try {
    files.value = await containerApi.filesList(props.containerId, currentPath.value)
    // Sort: directories first, then by name
    files.value.sort((a, b) => {
      if (a.is_dir && !b.is_dir) return -1
      if (!a.is_dir && b.is_dir) return 1
      return a.name.localeCompare(b.name)
    })
  } catch {
    ElMessage.error(t('file.loadError'))
  } finally {
    loading.value = false
  }
}

function navigateToSegment(index: number) {
  if (index >= pathSegments.value.length - 1) return
  
  const newPath = '/' + pathSegments.value.slice(1, index + 1).join('/')
  currentPath.value = newPath || '/'
  fetchFiles()
}

function handleRowDblClick(row: FileInfo) {
  if (row.is_dir) {
    currentPath.value = row.path
    fetchFiles()
  } else if (isTextFile(row.name)) {
    handleEdit(row)
  }
}

function handleFileClick(row: FileInfo) {
  if (row.is_dir) {
    currentPath.value = row.path
    fetchFiles()
  }
}

function handleRefresh() {
  fetchFiles()
}

async function handleCreateDir() {
  if (!newDirName.value.trim()) {
    ElMessage.warning(t('file.dirNameRequired'))
    return
  }
  
  creating.value = true
  try {
    const path = currentPath.value === '/' 
      ? '/' + newDirName.value 
      : currentPath.value + '/' + newDirName.value
    await containerApi.fileMkdir(props.containerId, path)
    ElMessage.success(t('file.createSuccess'))
    showCreateDirDialog.value = false
    newDirName.value = ''
    fetchFiles()
  } catch {
    // Error handled by interceptor
  } finally {
    creating.value = false
  }
}

async function handleDelete(file: FileInfo) {
  if (isProtectedPath(file.path)) {
    ElMessage.error(t('file.protectedPath'))
    return
  }
  
  try {
    await containerApi.fileDelete(props.containerId, file.path)
    ElMessage.success(t('file.deleteSuccess'))
    fetchFiles()
  } catch {
    // Error handled by interceptor
  }
}

function handleDownload(file: FileInfo) {
  const url = containerApi.fileDownload(props.containerId, file.path)
  window.open(url, '_blank')
}

function handleEdit(file: FileInfo) {
  editingFilePath.value = file.path
  editingFileName.value = file.name
  showEditorDialog.value = true
}

function triggerUpload() {
  fileInputRef.value?.click()
}

async function handleFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  
  if (file.size > MAX_UPLOAD_SIZE) {
    ElMessage.error(t('file.fileTooLarge'))
    input.value = ''
    return
  }
  
  try {
    const reader = new FileReader()
    reader.onload = async () => {
      const result = reader.result as string
      const base64 = result.split(',')[1] || ''
      await containerApi.fileUpload(props.containerId, currentPath.value, file.name, base64)
      ElMessage.success(t('file.uploadSuccess'))
      fetchFiles()
    }
    reader.readAsDataURL(file)
  } catch {
    // Error handled by interceptor
  } finally {
    input.value = ''
  }
}

function isProtectedPath(path: string): boolean {
  return PROTECTED_PATHS.includes(path)
}

function isTextFile(filename: string): boolean {
  const textExtensions = [
    '.txt', '.md', '.json', '.xml', '.yml', '.yaml', '.toml', '.ini', '.cfg', '.conf',
    '.sh', '.bash', '.zsh', '.fish', '.ps1', '.bat', '.cmd',
    '.js', '.ts', '.jsx', '.tsx', '.vue', '.svelte',
    '.py', '.rb', '.php', '.go', '.rs', '.java', '.kt', '.scala', '.c', '.cpp', '.h', '.hpp',
    '.html', '.htm', '.css', '.scss', '.sass', '.less',
    '.sql', '.graphql', '.gql',
    '.env', '.gitignore', '.dockerignore', '.editorconfig',
    '.log', '.csv',
  ]
  const lowerName = filename.toLowerCase()
  return textExtensions.some(ext => lowerName.endsWith(ext)) || !filename.includes('.')
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function formatTime(time: string): string {
  if (!time) return '-'
  return new Date(time).toLocaleString()
}

function handleOpen() {
  currentPath.value = '/'
  fetchFiles()
}

// Watch for container ID changes
watch(() => props.containerId, () => {
  if (props.visible) {
    currentPath.value = '/'
    fetchFiles()
  }
})
</script>

<style scoped>
.file-explorer {
  min-height: 400px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding: 10px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
}

.toolbar-actions {
  display: flex;
  gap: 8px;
}

.breadcrumb-item {
  cursor: default;
}

.breadcrumb-item.clickable {
  cursor: pointer;
  color: var(--el-color-primary);
}

.breadcrumb-item.clickable:hover {
  text-decoration: underline;
}

.file-name {
  cursor: pointer;
  color: var(--el-color-primary);
}

.file-name:hover {
  text-decoration: underline;
}
</style>
