import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import { systemApi } from '@/api/system'
import type { SystemStats } from '@/types'

// Polling interval in milliseconds (5 seconds)
const POLLING_INTERVAL = 5000

/**
 * System store
 * Manages system statistics with automatic polling
 */
export const useSystemStore = defineStore('system', () => {
  // State
  const stats = reactive<SystemStats>({
    cpu: 0,
    memory: 0,
    disk: 0,
    time: '',
  })
  
  const loading = ref(false)
  const error = ref<string | null>(null)
  
  // Polling timer reference
  let pollingTimer: ReturnType<typeof setInterval> | null = null

  /**
   * Fetch system statistics from API
   */
  async function fetchStats(): Promise<void> {
    try {
      loading.value = true
      error.value = null
      
      const data = await systemApi.stats()
      
      stats.cpu = data.cpu ?? 0
      stats.memory = data.memory ?? 0
      stats.disk = data.disk ?? 0
      stats.time = data.time ?? ''
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch system stats'
      // Don't reset stats on error, keep last known values
    } finally {
      loading.value = false
    }
  }

  /**
   * Start automatic polling for system stats
   * Polls every 5 seconds as per requirements
   */
  function startPolling(): void {
    // Fetch immediately
    fetchStats()
    
    // Stop any existing polling
    stopPolling()
    
    // Start new polling interval
    pollingTimer = setInterval(() => {
      fetchStats()
    }, POLLING_INTERVAL)
  }

  /**
   * Stop automatic polling
   */
  function stopPolling(): void {
    if (pollingTimer) {
      clearInterval(pollingTimer)
      pollingTimer = null
    }
  }

  return {
    // State
    stats,
    loading,
    error,
    // Actions
    fetchStats,
    startPolling,
    stopPolling,
  }
})
