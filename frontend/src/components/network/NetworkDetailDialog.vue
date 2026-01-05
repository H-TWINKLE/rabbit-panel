<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('network.details')"
    width="700px"
    :close-on-click-modal="true"
    @open="loadNetworkDetails"
  >
    <div v-loading="loading" class="network-details">
      <template v-if="networkDetail">
        <!-- Basic Info -->
        <el-descriptions :title="t('config.basicInfo')" :column="2" border>
          <el-descriptions-item :label="t('network.id')">
            <span class="monospace">{{ networkDetail.id }}</span>
          </el-descriptions-item>
          <el-descriptions-item :label="t('common.name')">
            {{ networkDetail.name }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('network.driver')">
            <el-tag size="small" type="primary">{{ networkDetail.driver }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="t('network.scope')">
            {{ networkDetail.scope }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('network.subnet')">
            {{ networkDetail.subnet || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('network.gateway')">
            {{ networkDetail.gateway || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('network.internal')">
            <el-tag :type="networkDetail.internal ? 'warning' : 'success'" size="small">
              {{ networkDetail.internal ? t('common.yes') : t('common.no') }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>

        <!-- Connected Containers -->
        <div class="section-title">
          <h4>{{ t('network.connectedContainers') }}</h4>
        </div>
        
        <el-table
          v-if="networkDetail.connectedContainers && networkDetail.connectedContainers.length > 0"
          :data="networkDetail.connectedContainers"
          stripe
          size="small"
          style="width: 100%"
        >
          <el-table-column prop="name" :label="t('common.name')" min-width="150">
            <template #default="{ row }">
              <span class="container-name">{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="id" :label="t('network.id')" width="140">
            <template #default="{ row }">
              <span class="monospace">{{ row.id.substring(0, 12) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="ipAddress" :label="t('config.ipAddress')" width="140">
            <template #default="{ row }">
              <span class="monospace">{{ row.ipAddress || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="macAddress" :label="t('config.macAddress')" width="160">
            <template #default="{ row }">
              <span class="monospace">{{ row.macAddress || '-' }}</span>
            </template>
          </el-table-column>
        </el-table>

        <el-empty 
          v-else 
          :description="t('network.noContainers')" 
          :image-size="80"
        />
      </template>

      <el-empty v-else-if="error" :description="error" />
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible = false">{{ t('common.close') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useNetworkStore } from '@/stores/networks'
import { useI18n } from '@/composables/useI18n'
import type { NetworkDetail } from '@/types'

const props = defineProps<{
  visible: boolean
  networkId: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const { t } = useI18n()
const networkStore = useNetworkStore()

// Dialog visibility
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value),
})

// State
const loading = ref(false)
const error = ref<string | null>(null)
const networkDetail = ref<NetworkDetail | null>(null)

// Load network details when dialog opens
async function loadNetworkDetails() {
  if (!props.networkId) return

  try {
    loading.value = true
    error.value = null
    networkDetail.value = await networkStore.inspectNetwork(props.networkId)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load network details'
    networkDetail.value = null
  } finally {
    loading.value = false
  }
}

// Reset when network ID changes
watch(() => props.networkId, () => {
  networkDetail.value = null
  error.value = null
})
</script>

<style scoped>
.network-details {
  min-height: 200px;
}

.section-title {
  margin-top: 20px;
  margin-bottom: 10px;
}

.section-title h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.monospace {
  font-family: monospace;
  font-size: 12px;
}

.container-name {
  font-weight: 500;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
}
</style>
