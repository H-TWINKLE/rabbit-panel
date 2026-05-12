import request from '@/utils/request'
import type { DockerConfig, DockerInfo } from '@/types'

/**
 * Docker Config API service
 * Handles Docker daemon configuration operations
 */
export const dockerConfigApi = {
  /**
   * Get Docker system information
   * @returns Docker info including version, OS, containers count, etc.
   */
  async getInfo(): Promise<DockerInfo> {
    const response = await request.get<DockerInfo>('/docker/info')
    return response.data
  },

  /**
   * Get current Docker daemon configuration
   * @returns Current Docker config from daemon.json
   */
  async getConfig(): Promise<DockerConfig> {
    const response = await request.get<DockerConfig>('/docker/config')
    return response.data
  },

  /**
   * Update Docker daemon configuration
   * @param config Partial config to update
   */
  async updateConfig(config: Partial<DockerConfig>): Promise<void> {
    await request.post('/docker/config/update', config)
  },

  /**
   * Restart Docker service
   * Warning: This will affect all running containers
   */
  async restart(): Promise<void> {
    await request.post('/docker/restart')
  },
}

export default dockerConfigApi
