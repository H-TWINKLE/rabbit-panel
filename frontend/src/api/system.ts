import request from '@/utils/request'
import type { SystemStats } from '@/types'

/**
 * System API service
 * Provides methods for fetching system statistics
 */
export const systemApi = {
  /**
   * Get current system statistics
   * @returns System stats including CPU, memory, disk usage and server time
   */
  async stats(): Promise<SystemStats> {
    const response = await request.get<SystemStats>('/system/stats')
    return response.data
  },
}
