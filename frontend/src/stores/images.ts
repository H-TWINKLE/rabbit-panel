import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { imageApi } from '@/api/images'
import type { ImageInfo } from '@/types'
import { getSavedPageSize, savePageSize } from '@/utils/pagination'

export type SortField = 'id' | 'name' | 'tag' | 'size' | 'created'
export type SortOrder = 'asc' | 'desc'
const PAGE_SIZE_STORAGE_KEY = 'rabbit-page-size-images'

/**
 * Image store
 * Manages image list with search, sort, and pagination
 */
export const useImageStore = defineStore('images', () => {
  // State
  const images = ref<ImageInfo[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Search state
  const searchQuery = ref('')

  // Sort state
  const sortField = ref<SortField>('name')
  const sortOrder = ref<SortOrder>('asc')

  // Pagination state
  const currentPage = ref(1)
  const pageSize = ref(getSavedPageSize(PAGE_SIZE_STORAGE_KEY))

  // Getters

  /**
   * Filter images by search query (name or tag)
   */
  const filteredImages = computed(() => {
    let result = [...images.value]

    // Filter by search query (name or tag)
    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase()
      result = result.filter(
        (img) =>
          img.name.toLowerCase().includes(query) ||
          img.tag.toLowerCase().includes(query)
      )
    }

    // Sort
    result.sort((a, b) => {
      let aVal: string | number
      let bVal: string | number

      switch (sortField.value) {
        case 'id':
          aVal = a.id
          bVal = b.id
          break
        case 'name':
          aVal = a.name.toLowerCase()
          bVal = b.name.toLowerCase()
          break
        case 'tag':
          aVal = a.tag.toLowerCase()
          bVal = b.tag.toLowerCase()
          break
        case 'size':
          // Parse size string to bytes for comparison
          aVal = parseSizeToBytes(a.size)
          bVal = parseSizeToBytes(b.size)
          break
        case 'created':
          aVal = a.created
          bVal = b.created
          break
        default:
          aVal = a.name.toLowerCase()
          bVal = b.name.toLowerCase()
      }

      if (aVal < bVal) return sortOrder.value === 'asc' ? -1 : 1
      if (aVal > bVal) return sortOrder.value === 'asc' ? 1 : -1
      return 0
    })

    return result
  })

  /**
   * Get paginated images from filtered list
   */
  const paginatedImages = computed(() => {
    const start = (currentPage.value - 1) * pageSize.value
    const end = start + pageSize.value
    return filteredImages.value.slice(start, end)
  })

  /**
   * Total number of pages
   */
  const totalPages = computed(() => {
    return Math.ceil(filteredImages.value.length / pageSize.value) || 1
  })

  /**
   * Total number of filtered images
   */
  const totalImages = computed(() => filteredImages.value.length)

  // Actions

  /**
   * Fetch images from API
   * @param refresh Force refresh from Docker daemon
   */
  async function fetchImages(refresh?: boolean): Promise<void> {
    try {
      loading.value = true
      error.value = null
      images.value = await imageApi.list(refresh)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch images'
    } finally {
      loading.value = false
    }
  }

  /**
   * Remove an image
   * @param id Image ID
   */
  async function removeImage(id: string): Promise<void> {
    await imageApi.remove(id)
    // Refresh list after removal
    await fetchImages()
  }

  /**
   * Set search query
   * @param query Search string
   */
  function setSearch(query: string): void {
    searchQuery.value = query
    // Reset to first page when search changes
    currentPage.value = 1
  }

  /**
   * Set sort field and toggle order if same field
   * @param field Field to sort by
   */
  function setSort(field: SortField): void {
    if (sortField.value === field) {
      // Toggle order if same field
      sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
    } else {
      sortField.value = field
      sortOrder.value = 'asc'
    }
  }

  /**
   * Set current page
   * @param page Page number
   */
  function setPage(page: number): void {
    if (page >= 1 && page <= totalPages.value) {
      currentPage.value = page
    }
  }

  /**
   * Set page size
   * @param size Number of items per page
   */
  function setPageSize(size: number): void {
    pageSize.value = size
    savePageSize(PAGE_SIZE_STORAGE_KEY, size)
    // Reset to first page when page size changes
    currentPage.value = 1
  }

  /**
   * Reset all filters and pagination
   */
  function resetFilters(): void {
    searchQuery.value = ''
    sortField.value = 'name'
    sortOrder.value = 'asc'
    currentPage.value = 1
  }

  return {
    // State
    images,
    loading,
    error,
    searchQuery,
    sortField,
    sortOrder,
    currentPage,
    pageSize,
    // Getters
    filteredImages,
    paginatedImages,
    totalPages,
    totalImages,
    // Actions
    fetchImages,
    removeImage,
    setSearch,
    setSort,
    setPage,
    setPageSize,
    resetFilters,
  }
})

/**
 * Parse size string to bytes for comparison
 * Supports formats like "1.5 GB", "500 MB", "100 KB", "1024 B"
 */
function parseSizeToBytes(sizeStr: string): number {
  const match = sizeStr.match(/^([\d.]+)\s*(B|KB|MB|GB|TB)?$/i)
  if (!match) return 0

  const value = parseFloat(match[1] || '0')
  const unit = (match[2] ?? 'B').toUpperCase()

  const multipliers: Record<string, number> = {
    'B': 1,
    'KB': 1024,
    'MB': 1024 * 1024,
    'GB': 1024 * 1024 * 1024,
    'TB': 1024 * 1024 * 1024 * 1024,
  }

  return value * (multipliers[unit] || 1)
}
