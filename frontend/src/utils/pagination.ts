const ALLOWED_PAGE_SIZES = [10, 20, 50, 100] as const
const DEFAULT_PAGE_SIZE = 10

function isAllowedPageSize(size: number): boolean {
  return ALLOWED_PAGE_SIZES.includes(size as (typeof ALLOWED_PAGE_SIZES)[number])
}

export function getSavedPageSize(storageKey: string, fallback = DEFAULT_PAGE_SIZE): number {
  if (typeof window === 'undefined') {
    return fallback
  }

  try {
    const saved = window.localStorage.getItem(storageKey)
    if (!saved) return fallback

    const size = Number(saved)
    return isAllowedPageSize(size) ? size : fallback
  } catch {
    return fallback
  }
}

export function savePageSize(storageKey: string, size: number): void {
  if (typeof window === 'undefined') {
    return
  }

  if (!isAllowedPageSize(size)) {
    return
  }

  try {
    window.localStorage.setItem(storageKey, String(size))
  } catch {
    // Ignore write errors (e.g. private mode or quota limits)
  }
}
