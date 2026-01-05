import request from '@/utils/request'
import type { RegistryInfo, CreateRegistryRequest, RegistryTestResult } from '@/types'

/**
 * Registry API service
 * Handles Docker registry CRUD operations
 */
export const registryApi = {
  /**
   * Get list of all registries
   * @returns Array of registry info
   */
  async list(): Promise<RegistryInfo[]> {
    const response = await request.get<RegistryInfo[]>('/registries')
    return response.data
  },

  /**
   * Create a new registry
   * @param data Registry creation request
   * @returns Created registry info
   */
  async create(data: CreateRegistryRequest): Promise<RegistryInfo> {
    const response = await request.post<RegistryInfo>('/registries', data)
    return response.data
  },

  /**
   * Update an existing registry
   * @param id Registry ID
   * @param data Partial registry data to update
   * @returns Updated registry info
   */
  async update(id: string, data: Partial<CreateRegistryRequest>): Promise<RegistryInfo> {
    const response = await request.put<RegistryInfo>(`/registries/${id}`, data)
    return response.data
  },

  /**
   * Remove a registry by ID
   * @param id Registry ID
   */
  async remove(id: string): Promise<void> {
    await request.delete(`/registries/${id}`)
  },

  /**
   * Test registry connection
   * @param id Registry ID
   * @returns Test result with success status and message
   */
  async test(id: string): Promise<RegistryTestResult> {
    const response = await request.post<RegistryTestResult>(`/registries/${id}/test`)
    return response.data
  },
}

export default registryApi
