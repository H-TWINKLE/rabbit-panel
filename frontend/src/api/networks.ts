import request from '@/utils/request'
import type { NetworkInfo, NetworkDetail } from '@/types'

/**
 * Create network request parameters
 */
export interface CreateNetworkRequest {
  name: string
  driver: string
  subnet?: string
  gateway?: string
  internal?: boolean
}

/**
 * Network API service
 * Handles Docker network CRUD operations
 */
export const networkApi = {
  /**
   * Get list of all networks
   * @returns Array of network info
   */
  async list(): Promise<NetworkInfo[]> {
    const response = await request.get<NetworkInfo[]>('/networks')
    return response.data
  },

  /**
   * Create a new network (simple)
   * @param name Network name
   * @param driver Network driver (bridge, overlay, macvlan, host, none)
   */
  async create(name: string, driver: string): Promise<void> {
    await request.post('/networks/create', { name, driver })
  },

  /**
   * Create a new network with full options
   * @param data Network creation parameters
   */
  async createFull(data: CreateNetworkRequest): Promise<void> {
    await request.post('/networks/create', data)
  },

  /**
   * Remove a network
   * @param id Network ID
   */
  async remove(id: string): Promise<void> {
    await request.post('/networks/remove', { id })
  },

  /**
   * Get detailed network information
   * @param id Network ID
   * @returns Network details including connected containers
   */
  async inspect(id: string): Promise<NetworkDetail> {
    const response = await request.get<NetworkDetail>('/networks/inspect', {
      params: { id },
    })
    return response.data
  },
}

export default networkApi
