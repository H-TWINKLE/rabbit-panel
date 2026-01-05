import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { nodesApi, type ScheduleRequest, type ScheduleResponse } from '@/api/nodes'
import type { NodeInfo } from '@/types'

// Polling interval in milliseconds (5 seconds)
const POLLING_INTERVAL = 5000

/**
 * Nodes store
 * Manages multi-node Docker infrastructure state
 * Only functional when backend is running in Master mode
 */
export const useNodesStore = defineStore('nodes', () => {
  // State
  const nodes = ref<NodeInfo[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const isMasterMode = ref(false)
  
  // Search and filter state
  const searchQuery = ref('')
  const statusFilter = ref<'all' | 'online' | 'offline'>('all')
  
  // Polling timer reference
  let pollingTimer: ReturnType<typeof setInterval> | null = null

  // Getters
  const filteredNodes = computed(() => {
    let result = nodes.value

    // Apply status filter
    if (statusFilter.value !== 'all') {
      result = result.filter(node => node.status === statusFilter.value)
    }

    // Apply search filter
    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase()
      result = result.filter(node =>
        node.name.toLowerCase().includes(query) ||
        node.address.toLowerCase().includes(query) ||
        node.id.toLowerCase().includes(query)
      )
    }

    return result
  })

  const onlineNodes = computed(() => 
    nodes.value.filter(node => node.status === 'online')
  )

  const offlineNodes = computed(() => 
    nodes.value.filter(node => node.status === 'offline')
  )

  const totalNodes = computed(() => nodes.value.length)

  const onlineCount = computed(() => onlineNodes.value.length)

  const offlineCount = computed(() => offlineNodes.value.length)

  /**
   * Fetch nodes list from API
   */
  async function fetchNodes(): Promise<void> {
    try {
      loading.value = true
      error.value = null
      
      const data = await nodesApi.list()
      nodes.value = data
      isMasterMode.value = true
    } catch (e: any) {
      // Check if error is because we're not in master mode
      if (e.response?.status === 400 && 
          e.response?.data?.includes?.('Master')) {
        isMasterMode.value = false
        error.value = null
        nodes.value = []
      } else {
        error.value = e instanceof Error ? e.message : 'Failed to fetch nodes'
      }
    } finally {
      loading.value = false
    }
  }

  /**
   * Schedule a container to a specific node or auto-select best node
   * @param data Schedule request configuration
   * @returns Schedule response with node and container info
   */
  async function scheduleContainer(data: ScheduleRequest): Promise<ScheduleResponse> {
    return await nodesApi.schedule(data)
  }

  /**
   * Get the best available node for scheduling
   * @returns Best node or null if none available
   */
  async function getBestNode(): Promise<NodeInfo | null> {
    return await nodesApi.getBestNode()
  }

  /**
   * Start automatic polling for nodes status
   * Polls every 5 seconds as per requirements
   */
  function startPolling(): void {
    // Fetch immediately
    fetchNodes()
    
    // Stop any existing polling
    stopPolling()
    
    // Start new polling interval
    pollingTimer = setInterval(() => {
      fetchNodes()
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

  /**
   * Set search query for filtering nodes
   */
  function setSearch(query: string): void {
    searchQuery.value = query
  }

  /**
   * Set status filter
   */
  function setStatusFilter(status: 'all' | 'online' | 'offline'): void {
    statusFilter.value = status
  }

  /**
   * Get a specific node by ID
   */
  function getNodeById(nodeId: string): NodeInfo | undefined {
    return nodes.value.find(node => node.id === nodeId)
  }

  return {
    // State
    nodes,
    loading,
    error,
    isMasterMode,
    searchQuery,
    statusFilter,
    // Getters
    filteredNodes,
    onlineNodes,
    offlineNodes,
    totalNodes,
    onlineCount,
    offlineCount,
    // Actions
    fetchNodes,
    scheduleContainer,
    getBestNode,
    startPolling,
    stopPolling,
    setSearch,
    setStatusFilter,
    getNodeById,
  }
})
