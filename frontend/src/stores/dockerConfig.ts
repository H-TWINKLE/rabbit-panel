import { defineStore } from 'pinia'
import { ref } from 'vue'
import { dockerConfigApi } from '@/api/dockerConfig'
import type { DockerConfig, DockerInfo } from '@/types'

/**
 * Docker Config store
 * Manages Docker daemon configuration and system info
 */
export const useDockerConfigStore = defineStore('dockerConfig', () => {
  // State
  const config = ref<DockerConfig | null>(null)
  const info = ref<DockerInfo | null>(null)
  const loading = ref(false)
  const saving = ref(false)
  const error = ref<string | null>(null)

  // Actions

  /**
   * Fetch Docker config and info from API
   */
  async function fetchConfig(): Promise<void> {
    try {
      loading.value = true
      error.value = null
      const [configResult, infoResult] = await Promise.allSettled([
        dockerConfigApi.getConfig(),
        dockerConfigApi.getInfo(),
      ])

      const errors: string[] = []

      if (configResult.status === 'fulfilled') {
        config.value = configResult.value
      } else {
        errors.push(configResult.reason instanceof Error ? configResult.reason.message : 'Failed to fetch Docker config')
      }

      if (infoResult.status === 'fulfilled') {
        info.value = infoResult.value
      } else {
        errors.push(infoResult.reason instanceof Error ? infoResult.reason.message : 'Failed to fetch Docker info')
      }

      if (errors.length > 0) {
        error.value = errors.join('; ')
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch Docker config'
    } finally {
      loading.value = false
    }
  }

  /**
   * Update Docker daemon configuration
   * @param newConfig Partial config to update
   */
  async function updateConfig(newConfig: Partial<DockerConfig>): Promise<void> {
    try {
      saving.value = true
      error.value = null
      await dockerConfigApi.updateConfig(newConfig)
      await fetchConfig()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to update Docker config'
      throw e
    } finally {
      saving.value = false
    }
  }

  /**
   * Restart Docker service
   * Warning: This will affect all running containers
   */
  async function restartDocker(): Promise<void> {
    try {
      loading.value = true
      error.value = null
      await dockerConfigApi.restart()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to restart Docker'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * Reset store state
   */
  function $reset(): void {
    config.value = null
    info.value = null
    loading.value = false
    saving.value = false
    error.value = null
  }

  return {
    // State
    config,
    info,
    loading,
    saving,
    error,
    // Actions
    fetchConfig,
    updateConfig,
    restartDocker,
    $reset,
  }
})
