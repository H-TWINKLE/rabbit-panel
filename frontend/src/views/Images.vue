<template>
  <div class="images-page">
    <!-- Header with title and actions -->
    <div class="page-header">
      <h2>{{ t('image.title') }}</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showBuildDialog = true">
          <el-icon><Box /></el-icon>
          {{ t('image.build') }}
        </el-button>
        <el-button :loading="imageStore.loading" @click="handleRefresh">
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

    <!-- Image table -->
    <el-table
      v-loading="imageStore.loading"
      :data="imageStore.paginatedImages"
      stripe
      style="width: 100%"
      @sort-change="handleSortChange"
    >
      <el-table-column
        prop="id"
        :label="t('image.id')"
        width="140"
        sortable="custom"
        class-name="hidden-sm-and-down"
      >
        <template #default="{ row }">
          <span class="image-id">{{ row.id.substring(0, 12) }}</span>
        </template>
      </el-table-column>
      <el-table-column
        prop="name"
        :label="t('common.name')"
        min-width="200"
        sortable="custom"
      />
      <el-table-column
        prop="tag"
        :label="t('image.tag')"
        width="120"
        sortable="custom"
      >
        <template #default="{ row }">
          <el-tag size="small">{{ row.tag }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column
        prop="size"
        :label="t('common.size')"
        width="120"
        sortable="custom"
        class-name="hidden-xs-only"
      />
      <el-table-column
        prop="created"
        :label="t('image.created')"
        width="180"
        sortable="custom"
        class-name="hidden-md-and-down"
      />
      <el-table-column
        :label="t('common.actions')"
        width="120"
        fixed="right"
      >
        <template #default="{ row }">
          <el-popconfirm
            :title="t('image.confirmRemove')"
            :confirm-button-text="t('common.confirm')"
            :cancel-button-text="t('common.cancel')"
            @confirm="handleRemoveImage(row.id)"
          >
            <template #reference>
              <el-button type="danger" size="small" link>
                <el-icon><Delete /></el-icon>
                {{ t('common.delete') }}
              </el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="currentPageValue"
        v-model:page-size="pageSizeValue"
        :page-sizes="[10, 20, 50, 100]"
        :total="imageStore.totalImages"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="handlePageChange"
        @size-change="handlePageSizeChange"
      />
    </div>

    <!-- Build Image Dialog -->
    <BuildImageDialog
      v-model:visible="showBuildDialog"
      @built="handleImageBuilt"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Refresh, Search, Delete, Box } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useImageStore, type SortField } from '@/stores/images'
import { useI18n } from '@/composables/useI18n'
import BuildImageDialog from '@/components/image/BuildImageDialog.vue'

const { t } = useI18n()
const imageStore = useImageStore()

// Local state for v-model bindings
const searchInput = ref('')
const currentPageValue = ref(1)
const pageSizeValue = ref(10)

// Dialog visibility
const showBuildDialog = ref(false)

// Sync local state with store
watch(() => imageStore.searchQuery, (val) => { searchInput.value = val })
watch(() => imageStore.currentPage, (val) => { currentPageValue.value = val })
watch(() => imageStore.pageSize, (val) => { pageSizeValue.value = val })

// Handlers
function handleSearch(value: string) {
  imageStore.setSearch(value)
}

function handleSortChange({ prop, order }: { prop: string; order: string | null }) {
  if (prop && order) {
    const field = prop as SortField
    // Map Element Plus sort order to our format
    imageStore.sortField = field
    imageStore.sortOrder = order === 'ascending' ? 'asc' : 'desc'
  }
}

function handlePageChange(page: number) {
  imageStore.setPage(page)
}

function handlePageSizeChange(size: number) {
  imageStore.setPageSize(size)
}

async function handleRefresh() {
  await imageStore.fetchImages(true)
}

async function handleRemoveImage(id: string) {
  try {
    await imageStore.removeImage(id)
    ElMessage.success(t('image.removeSuccess'))
  } catch {
    // Error is handled by request interceptor
  }
}

function handleImageBuilt() {
  imageStore.fetchImages(true)
}

// Fetch images on mount
onMounted(() => {
  imageStore.fetchImages()
})
</script>

<style scoped>
.images-page {
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

.image-id {
  font-family: monospace;
  font-size: 12px;
}

/* Responsive styles */
@media (max-width: 767px) {
  .images-page {
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
