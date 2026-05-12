import request from '@/utils/request'
import type { VolumeInfo, CreateVolumeRequest, VolumePruneResult } from '@/types'

/**
 * Volume API service
 * Handles Docker volume CRUD operations
 */
export const volumeApi = {
  /**
   * Get list of all volumes
   * @returns Array of volume info
   */
  async list(): Promise<VolumeInfo[]> {
    const response = await request.get<VolumeInfo[]>('/volumes')
    return response.data
  },

  /**
   * Create a new volume
   * @param data Volume creation request
   * @returns Created volume info
   */
  async create(data: CreateVolumeRequest): Promise<VolumeInfo> {
    const response = await request.post<VolumeInfo>('/volumes/create', data)
    return response.data
  },

  /**
   * Remove a volume by name
   * @param name Volume name
   */
  async remove(name: string): Promise<void> {
    await request.post('/volumes/remove', { name })
  },

  /**
   * Prune unused volumes
   * @returns Prune result with deleted volumes and reclaimed space
   */
  async prune(): Promise<VolumePruneResult> {
    const response = await request.post<VolumePruneResult>('/volumes/prune')
    return response.data
  },

  /**
   * Get volume details
   * @param name Volume name
   * @returns Volume info
   */
  async inspect(name: string): Promise<VolumeInfo> {
    const volumes = await volumeApi.list()
    const found = volumes.find((volume) => volume.name === name)
    if (!found) {
      throw new Error(`Volume ${name} not found`)
    }
    return found
  },
}

export default volumeApi
