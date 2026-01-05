<template>
  <div class="volumes-page">
    <!-- Header with title and actions -->
    <div class="page-header">
      <h2>{{ t('volumes.title') }}</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          {{ t('volumes.create') }}
        </el-button>
        <el-popconfirm
          :title="t('volumes.pruneConfirm')"
          :confirm-button-text="t('common.confirm')"
          :cancel-button-text="t('common.cancel')"
          @confirm="handlePrune"
        >
          <template #reference>
            <el-button type="warning" :disabled="volumeStore.unusedCount === 0">
              <el-icon><Delete /></el-icon>
              {{ t('volumes.prune') }} ({{ volumeStore.unusedCount }})
            </el-button>
          </template>
        </el-popconfirm>
        <el-button :loading="volumeStore.loading" @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <!-- Search bar -->
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
    </div>

    <!-- Volume table -->
    <el-table
      v-loading="volumeStore.loading"
      :data="volumeStore.paginatedVolumes"
      stripe
      style="width: 100%"
      @sort-change="handleSortChange"
    >
      <el-table-column
        prop="name"
        :label="t('volumes.name')"
        min-width="200"
        sortable="custom"
      >
        <template #default="{ row }">
          <span class="volume-name">{{ row.name }}</span>
        </template>
      </el-table-column>
      <el-table-column
        prop="driver"
        :label="t('volumes.driver')"
        width="120"
        sortable="custom"
      >
        <template #default="{ row }">
          <el-tag size="small" type="primary">{{ row.driver }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column
        prop="mountpoint"
        :label="t('volumes.mountpoint')"
        min-width="250"
        class-name="hidden-md-and-down"
      >
        <template #default="{ row }">
          <span class="mountpoint">{{ row.mountpoint }}</span>
        </template>
      </el-table-column>
      <el-table-column
        prop="containers"
        :label="t('volumes.usedBy')"
        width="150"
        class-name="hidden-sm-and-down"
      >
        <template #default="{ row }">
          <template v-if="row.containers && row.containers.length > 0">
            <el-tooltip
              :content="row.containers.join(', ')"
              placement="top"
            >
              <el-tag size="small" type="success">
                {{ row.containers.length }} {{ t('volumes.containers') }}
              </el-tag>
            </el-tooltip>
          </template>
          <el-tag v-else size="small" type="info">
            {{ t('volumes.unused') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column
        prop="created"
        :label="t('common.createdAt')"
        width="180"
        sortable="custom"
        class-name="hidden-xs-only"
      >
        <template #default="{ row }">
          <span>{{ formatDate(row.created) }}</span>
        </template>
      </el-table-column>
      <el-table-column
        :label="t('common.actions')"
        width="120"
        fixed="right"
      >
        <template #default="{ row }">
          <el-popconfirm
            v-if="!row.containers || row.containers.length === 0"
            :title="t('volumes.confirmRemove')"
            :confirm-button-text="t('common.confirm')"
            :cancel-button-text="t('common.cancel')"
            @confirm="handleRemove(row.name)"
          >
            <template #reference>
              <el-button type="danger" size="small" link>
                <el-icon><Delete /></el-icon>
                {{ t('common.delete') }}
              </el-button>
            </template>
          </el-popconfirm>
          <el-tooltip
            v-else
            :content="t('volumes.inUse')"
            placement="top"
          >
            <el-button type="info" size="small" link disabled>
              <el-icon><Delete /></el-icon>
              {{ t('common.delete') }}
            </el-button>
          </el-tooltip>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="currentPageValue"
        v-model:page-size="pageSizeValue"
        :page-sizes="[10, 20, 50, 100]"
        :total="volumeStore.totalVolumes"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="handlePageChange"
        @size-change="handlePageSizeChange"
      />
    </div>

    <!-- Create Volume Dialog -->
    <CreateVolumeDialog
      v-model:visible="showCreateDialog"
      @created="handleVolumeCreated"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Refresh, Search, Delete, Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useVolumeStore, type SortField } from '@/stores/volumes'
import { useI18n } from '@/composables/useI18n'
import CreateVolumeDialog from '@/components/volumes/CreateVolumeDialog.vue'

const { t } = useI18n()
const volumeStore = useVolumeStore()

// Local state for v-model bindings
const searchInput = ref('')
const currentPageValue = ref(1)
const pageSizeValue = ref(10)

// Dialog visibility
const showCreateDialog = ref(false)

// Sync local state with store
watch(() => volumeStore.searchQuery, (val) => { searchInput.value = val })
watch(() => volumeStore.currentPage, (val) => { currentPageValue.value = val })
watch(() => volumeStore.pageSize, (val) => { pageSizeValue.value = val })

// Format date helper
function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString()
}

// Format bytes helper
function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// Handlers
function handleSearch(value: string) {
  volumeStore.setSearch(value)
}

function handleSortChange({ prop, order }: { prop: string; order: string | null }) {
  if (prop && order) {
    const field = prop as SortField
    volumeStore.sortField = field
    volumeStore.sortOrder = order === 'ascending' ? 'asc' : 'desc'
  }
}

function handlePageChange(page: number) {
  volumeStore.setPage(page)
}

function handlePageSizeChange(size: number) {
  volumeStore.setPageSize(size)
}

async function handleRefresh() {
  await volumeStore.fetchVolumes()
}

async function handleRemove(name: string) {
  try {
    await volumeStore.removeVolume(name)
    ElMessage.success(t('volumes.removeSuccess'))
  } catch {
    // Error is handled by request interceptor
  }
}

async function handlePrune() {
  try {
    const result = await volumeStore.pruneVolumes()
    const count = result.volumesDeleted?.length || 0
    const size = formatBytes(result.spaceReclaimed || 0)
    ElMessage.success(t('volumes.pruneSuccess').replace('{count}', String(count)).replace('{size}', size))
  } catch {
    // Error is handled by request interceptor
  }
}

function handleVolumeCreated() {
  volumeStore.fetchVolumes()
}

// Fetch volumes on mount
onMounted(() => {
  volumeStore.fetchVolumes()
})
</script>

<style scoped>
.volumes-page {
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

.volume-name {
  font-family: monospace;
  font-size: 13px;
}

.mountpoint {
  font-family: monospace;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

/* Responsive styles */
@media (max-width: 767px) {
  .volumes-page {
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

  .filter-bar .el-input {
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
