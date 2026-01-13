<template>
  <div class="containers-page">
    <!-- Header with title and actions -->
    <div class="page-header">
      <h2>{{ t('container.title') }}</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          {{ t('container.create') }}
        </el-button>
        <el-button @click="showDockerRunDialog = true">
          <el-icon><Monitor /></el-icon>
          {{ t('container.dockerRun') }}
        </el-button>
        <el-button :loading="containerStore.loading" @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <!-- Search and filter bar -->
    <div class="filter-bar">
      <el-input
        v-model="searchInput"
        :placeholder="t('common.search')"
        clearable
        style="width: 300px"
        @input="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>

      <el-select
        v-model="statusFilterValue"
        :placeholder="t('common.status')"
        style="width: 150px"
        @change="handleStatusFilter"
      >
        <el-option :label="t('common.all')" value="all" />
        <el-option :label="t('container.running')" value="running" />
        <el-option :label="t('container.stopped')" value="exited" />
        <el-option :label="t('container.paused')" value="paused" />
        <el-option :label="t('container.created')" value="created" />
      </el-select>
    </div>

    <!-- Container table -->
    <el-table
      v-loading="containerStore.loading"
      :data="containerStore.paginatedContainers"
      stripe
      style="width: 100%"
      @sort-change="handleSortChange"
    >
      <el-table-column
        prop="id"
        label="ID"
        min-width="120"
        sortable="custom"
        show-overflow-tooltip
        class-name="hidden-sm-and-down"
      />
      <el-table-column
        prop="name"
        :label="t('common.name')"
        min-width="120"
        sortable="custom"
        show-overflow-tooltip
      />
      <el-table-column
        prop="image"
        :label="t('container.image')"
        min-width="250"
        sortable="custom"
        show-overflow-tooltip
        class-name="hidden-xs-only"
      />
      <el-table-column
        prop="state"
        :label="t('common.status')"
        min-width="80"
        sortable="custom"
      >
        <template #default="{ row }">
          <el-tag :type="getStateTagType(row.state)" size="small">
            {{ getStateLabel(row.state) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column
        prop="ports"
        :label="t('container.ports')"
        min-width="200"
        show-overflow-tooltip
        class-name="hidden-sm-and-down"
      />
      <el-table-column
        prop="cpu"
        label="CPU"
        min-width="80"
        class-name="hidden-md-and-down"
      >
        <template #default="{ row }">
          <template v-if="row.state === 'running'">
            <span v-if="loadingStats[row.id]">...</span>
            <span v-else-if="containerStats[row.id]">{{ containerStats[row.id]!.cpu_percent.toFixed(1) }}%</span>
            <span v-else>-</span>
          </template>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column
        prop="memory"
        :label="t('system.memory')"
        min-width="120"
        class-name="hidden-md-and-down"
      >
        <template #default="{ row }">
          <template v-if="row.state === 'running'">
            <span v-if="loadingStats[row.id]">...</span>
            <template v-else-if="containerStats[row.id]">
              <span>{{ formatMemory(containerStats[row.id]!.memory_usage) }}</span>
              <span v-if="containerStats[row.id]!.has_memory_limit" class="memory-limit"> / {{ formatMemory(containerStats[row.id]!.memory_limit) }}</span>
              <span v-else class="memory-limit"> / 无限制</span>
            </template>
            <span v-else>-</span>
          </template>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column
        prop="created"
        :label="t('common.time')"
        min-width="170"
        sortable="custom"
        show-overflow-tooltip
        class-name="hidden-md-and-down"
      />
      <el-table-column
        :label="t('common.actions')"
        :width="isMobile ? 120 : 380"
        fixed="right"
        align="center"
      >
        <template #default="{ row }">
          <ContainerActions
            :container="row"
            :compact="isMobile"
            @action="handleContainerAction"
            @logs="handleViewLogs"
            @terminal="handleOpenTerminal"
            @files="handleOpenFiles"
            @config="handleOpenConfig"
          />
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="currentPageValue"
        v-model:page-size="pageSizeValue"
        :page-sizes="[10, 20, 50, 100]"
        :total="containerStore.totalContainers"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="handlePageChange"
        @size-change="handlePageSizeChange"
      />
    </div>

    <!-- Create Container Dialog -->
    <CreateContainerDialog
      v-model:visible="showCreateDialog"
      @created="handleContainerCreated"
    />

    <!-- Docker Run Dialog -->
    <DockerRunDialog
      v-model:visible="showDockerRunDialog"
      @created="handleContainerCreated"
    />

    <!-- Container Logs Dialog -->
    <ContainerLogsDialog
      v-model:visible="showLogsDialog"
      :container-id="selectedContainerId"
      :container-name="selectedContainerName"
    />

    <!-- Container Terminal Dialog -->
    <ContainerTerminal
      v-model:visible="showTerminalDialog"
      :container-id="selectedContainerId"
      :container-name="selectedContainerName"
    />

    <!-- Container File Explorer Dialog -->
    <FileExplorer
      v-model:visible="showFilesDialog"
      :container-id="selectedContainerId"
      :container-name="selectedContainerName"
    />

    <!-- Container Config Dialog -->
    <ContainerConfig
      v-model:visible="showConfigDialog"
      :container-id="selectedContainerId"
      :container-name="selectedContainerName"
      @updated="handleContainerCreated"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed, onUnmounted, reactive } from 'vue'
import { useRoute } from 'vue-router'
import { Plus, Refresh, Search, Monitor } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useContainerStore, type ContainerState, type SortField } from '@/stores/containers'
import { useI18n } from '@/composables/useI18n'
import { containerApi } from '@/api/containers'
import type { ContainerStats } from '@/types'
import ContainerActions from '@/components/container/ContainerActions.vue'
import CreateContainerDialog from '@/components/container/CreateContainerDialog.vue'
import DockerRunDialog from '@/components/container/DockerRunDialog.vue'
import ContainerLogsDialog from '@/components/container/ContainerLogsDialog.vue'
import ContainerTerminal from '@/components/container/ContainerTerminal.vue'
import FileExplorer from '@/components/container/FileExplorer.vue'
import ContainerConfig from '@/components/container/ContainerConfig.vue'

const { t } = useI18n()
const route = useRoute()
const containerStore = useContainerStore()

// 容器资源统计缓存
const containerStats = reactive<Record<string, ContainerStats | null>>({})
const loadingStats = reactive<Record<string, boolean>>({})

// 格式化内存大小
function formatMemory(bytes: number): string {
  if (bytes === 0) return '-'
  const mb = bytes / 1024 / 1024
  if (mb >= 1024) {
    return `${(mb / 1024).toFixed(1)} GB`
  }
  return `${mb.toFixed(0)} MB`
}

// 获取容器 stats
async function fetchContainerStats(containerId: string) {
  if (loadingStats[containerId]) return
  loadingStats[containerId] = true
  try {
    const stats = await containerApi.stats(containerId)
    containerStats[containerId] = stats
  } catch {
    containerStats[containerId] = null
  } finally {
    loadingStats[containerId] = false
  }
}

// 批量获取运行中容器的 stats
async function fetchAllRunningStats() {
  const runningContainers = containerStore.containers.filter(c => c.state === 'running')
  // 并行获取，但限制并发数
  const batchSize = 5
  for (let i = 0; i < runningContainers.length; i += batchSize) {
    const batch = runningContainers.slice(i, i + batchSize)
    await Promise.all(batch.map(c => fetchContainerStats(c.id)))
  }
}

// 监听容器列表变化，自动获取 stats
watch(() => containerStore.containers, () => {
  fetchAllRunningStats()
}, { immediate: true })

// Responsive state
const windowWidth = ref(window.innerWidth)
const isMobile = computed(() => windowWidth.value < 768)

function handleResize() {
  windowWidth.value = window.innerWidth
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})

// Local state for v-model bindings
const searchInput = ref('')
const statusFilterValue = ref<ContainerState>('all')
const currentPageValue = ref(1)
const pageSizeValue = ref(10)

// Dialog visibility
const showCreateDialog = ref(false)
const showDockerRunDialog = ref(false)
const showLogsDialog = ref(false)
const showTerminalDialog = ref(false)
const showFilesDialog = ref(false)
const showConfigDialog = ref(false)

// Selected container for logs
const selectedContainerId = ref('')
const selectedContainerName = ref('')

// Sync local state with store
watch(() => containerStore.searchQuery, (val) => { searchInput.value = val })
watch(() => containerStore.statusFilter, (val) => { statusFilterValue.value = val })
watch(() => containerStore.currentPage, (val) => { currentPageValue.value = val })
watch(() => containerStore.pageSize, (val) => { pageSizeValue.value = val })

// Handlers
function handleSearch(value: string) {
  containerStore.setSearch(value)
}

function handleStatusFilter(value: ContainerState) {
  containerStore.setStatusFilter(value)
}

function handleSortChange({ prop, order }: { prop: string; order: string | null }) {
  if (prop && order) {
    const field = prop as SortField
    // Map Element Plus sort order to our format
    containerStore.sortField = field
    containerStore.sortOrder = order === 'ascending' ? 'asc' : 'desc'
  }
}

function handlePageChange(page: number) {
  containerStore.setPage(page)
}

function handlePageSizeChange(size: number) {
  containerStore.setPageSize(size)
}

async function handleRefresh() {
  await containerStore.fetchContainers()
}

async function handleContainerAction(id: string, action: 'start' | 'stop' | 'restart' | 'remove') {
  try {
    await containerStore.containerAction(id, action)
    ElMessage.success(t('common.success'))
  } catch {
    // Error is handled by request interceptor
  }
}

function handleViewLogs(id: string, name: string) {
  selectedContainerId.value = id
  selectedContainerName.value = name
  showLogsDialog.value = true
}

function handleOpenTerminal(id: string, name: string) {
  selectedContainerId.value = id
  selectedContainerName.value = name
  showTerminalDialog.value = true
}

function handleOpenFiles(id: string, name: string) {
  selectedContainerId.value = id
  selectedContainerName.value = name
  showFilesDialog.value = true
}

function handleOpenConfig(id: string, name: string) {
  selectedContainerId.value = id
  selectedContainerName.value = name
  showConfigDialog.value = true
}

function handleContainerCreated() {
  containerStore.fetchContainers()
}

// Get tag type based on container state
function getStateTagType(state: string): 'success' | 'danger' | 'warning' | 'info' {
  switch (state) {
    case 'running':
      return 'success'
    case 'exited':
      return 'danger'
    case 'paused':
      return 'warning'
    default:
      return 'info'
  }
}

// Get localized state label
function getStateLabel(state: string): string {
  switch (state) {
    case 'running':
      return t('container.running')
    case 'exited':
      return t('container.stopped')
    case 'paused':
      return t('container.paused')
    case 'created':
      return t('container.created')
    default:
      return state
  }
}

// Fetch containers on mount and handle query params
onMounted(async () => {
  await containerStore.fetchContainers()
  
  // 处理从 Compose 页面跳转过来的操作
  const action = route.query.action as string
  const containerName = route.query.container as string
  
  if (action && containerName) {
    // 等待容器列表加载完成后查找容器
    const container = containerStore.containers.find(c => c.name === containerName)
    if (container) {
      selectedContainerId.value = container.id
      selectedContainerName.value = container.name
      switch (action) {
        case 'terminal':
          showTerminalDialog.value = true
          break
        case 'files':
          showFilesDialog.value = true
          break
        case 'logs':
          showLogsDialog.value = true
          break
      }
    } else {
      ElMessage.warning(`容器 ${containerName} 未找到`)
    }
  }
})
</script>

<style scoped>
.containers-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.filter-bar {
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.memory-limit {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

/* Responsive styles */
@media (max-width: 767px) {
  .containers-page {
    padding: 12px;
  }
  
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .header-actions {
    width: 100%;
    flex-wrap: wrap;
  }
  
  .header-actions .el-button {
    flex: 1;
    min-width: 80px;
  }
  
  .filter-bar {
    flex-direction: column;
  }
  
  .filter-bar .el-input,
  .filter-bar .el-select {
    width: 100% !important;
  }
  
  .pagination-wrapper {
    justify-content: center;
  }
}

/* Hide columns on smaller screens */
:deep(.hidden-xs-only) {
  @media (max-width: 767px) {
    display: none !important;
  }
}

:deep(.hidden-sm-and-down) {
  @media (max-width: 991px) {
    display: none !important;
  }
}

:deep(.hidden-md-and-down) {
  @media (max-width: 1199px) {
    display: none !important;
  }
}
</style>
