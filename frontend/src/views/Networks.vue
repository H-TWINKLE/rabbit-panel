<template>
  <div class="networks-page">
    <!-- Header with title and actions -->
    <div class="page-header">
      <h2>{{ t('network.title') }}</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          {{ t('network.create') }}
        </el-button>
        <el-button :loading="networkStore.loading" @click="handleRefresh">
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

    <!-- Network table -->
    <el-table
      v-loading="networkStore.loading"
      :data="networkStore.paginatedNetworks"
      stripe
      style="width: 100%"
      @sort-change="handleSortChange"
    >
      <el-table-column
        prop="id"
        :label="t('network.id')"
        width="140"
        sortable="custom"
        class-name="hidden-sm-and-down"
      >
        <template #default="{ row }">
          <span class="network-id">{{ row.id.substring(0, 12) }}</span>
        </template>
      </el-table-column>
      <el-table-column
        prop="name"
        :label="t('common.name')"
        min-width="180"
        sortable="custom"
      >
        <template #default="{ row }">
          <span>{{ row.name }}</span>
          <el-tag 
            v-if="networkStore.isProtectedNetwork(row.name)" 
            size="small" 
            type="info"
            style="margin-left: 8px"
          >
            System
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column
        prop="driver"
        :label="t('network.driver')"
        width="120"
        sortable="custom"
      >
        <template #default="{ row }">
          <el-tag size="small" type="primary">{{ row.driver }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column
        prop="scope"
        :label="t('network.scope')"
        width="100"
        sortable="custom"
        class-name="hidden-xs-only"
      />
      <el-table-column
        prop="ipam"
        :label="t('network.ipam')"
        min-width="150"
        class-name="hidden-md-and-down"
      >
        <template #default="{ row }">
          <span>{{ row.ipam || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column
        prop="containers"
        :label="t('network.containers')"
        width="100"
        sortable="custom"
        align="center"
        class-name="hidden-xs-only"
      >
        <template #default="{ row }">
          <el-tag :type="row.container_count > 0 ? 'success' : 'info'" size="small">
            {{ row.container_count }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column
        prop="in_use"
        label="Usage"
        width="140"
        align="center"
      >
        <template #default="{ row }">
          <el-tooltip v-if="row.containers && row.containers.length > 0" :content="row.containers.join(', ')" placement="top">
            <el-tag type="success" size="small">
              In Use
            </el-tag>
          </el-tooltip>
          <el-tag v-else type="info" size="small">
            Unused
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column
        :label="t('common.actions')"
        width="160"
        fixed="right"
      >
        <template #default="{ row }">
          <el-button type="primary" size="small" link @click="handleViewDetails(row)">
            <el-icon><View /></el-icon>
            {{ t('network.details') }}
          </el-button>
          <el-popconfirm
            v-if="!networkStore.isProtectedNetwork(row.name)"
            :title="t('network.confirmRemove')"
            :confirm-button-text="t('common.confirm')"
            :cancel-button-text="t('common.cancel')"
            @confirm="handleRemoveNetwork(row.id)"
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
            :content="t('network.protectedNetwork')" 
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
        :total="networkStore.totalNetworks"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="handlePageChange"
        @size-change="handlePageSizeChange"
      />
    </div>

    <!-- Create Network Dialog -->
    <CreateNetworkDialog
      v-model:visible="showCreateDialog"
      @created="handleNetworkCreated"
    />

    <!-- Network Detail Dialog -->
    <NetworkDetailDialog
      v-model:visible="showDetailDialog"
      :network-id="selectedNetworkId"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Refresh, Search, Delete, Plus, View } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useNetworkStore, type SortField } from '@/stores/networks'
import { useI18n } from '@/composables/useI18n'
import type { NetworkInfo } from '@/types'
import CreateNetworkDialog from '@/components/network/CreateNetworkDialog.vue'
import NetworkDetailDialog from '@/components/network/NetworkDetailDialog.vue'

const { t } = useI18n()
const networkStore = useNetworkStore()

// Local state for v-model bindings
const searchInput = ref('')
const currentPageValue = ref(1)
const pageSizeValue = ref(10)

// Dialog visibility
const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const selectedNetworkId = ref('')

// Sync local state with store
watch(() => networkStore.searchQuery, (val) => { searchInput.value = val })
watch(() => networkStore.currentPage, (val) => { currentPageValue.value = val })
watch(() => networkStore.pageSize, (val) => { pageSizeValue.value = val })

// Handlers
function handleSearch(value: string) {
  networkStore.setSearch(value)
}

function handleSortChange({ prop, order }: { prop: string; order: string | null }) {
  if (prop && order) {
    const field = prop as SortField
    networkStore.sortField = field
    networkStore.sortOrder = order === 'ascending' ? 'asc' : 'desc'
  }
}

function handlePageChange(page: number) {
  networkStore.setPage(page)
}

function handlePageSizeChange(size: number) {
  networkStore.setPageSize(size)
}

async function handleRefresh() {
  await networkStore.fetchNetworks()
}

async function handleRemoveNetwork(id: string) {
  try {
    await networkStore.removeNetwork(id)
    ElMessage.success(t('network.removeSuccess'))
  } catch {
    // Error is handled by request interceptor
  }
}

function handleViewDetails(network: NetworkInfo) {
  selectedNetworkId.value = network.id
  showDetailDialog.value = true
}

function handleNetworkCreated() {
  networkStore.fetchNetworks()
}

// Fetch networks on mount
onMounted(() => {
  networkStore.fetchNetworks()
})
</script>

<style scoped>
.networks-page {
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

.network-id {
  font-family: monospace;
  font-size: 12px;
}

/* Responsive styles */
@media (max-width: 767px) {
  .networks-page {
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
