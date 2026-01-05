import { ref, watch, onMounted } from 'vue'
import type { Theme } from '@/types'

const THEME_KEY = 'rabbit_panel_theme'

// Global theme state (shared across all components)
const theme = ref<Theme>('light')
const isInitialized = ref(false)

/**
 * Theme composable for managing dark/light mode
 * Persists theme preference to localStorage
 */
export function useTheme() {
  /**
   * Initialize theme from localStorage or system preference
   */
  function initTheme(): void {
    if (isInitialized.value) return
    
    const stored = localStorage.getItem(THEME_KEY) as Theme | null
    
    if (stored && (stored === 'light' || stored === 'dark')) {
      theme.value = stored
    } else {
      // Auto-detect system preference
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      theme.value = prefersDark ? 'dark' : 'light'
    }
    
    applyTheme(theme.value)
    isInitialized.value = true
  }

  /**
   * Apply theme to document
   */
  function applyTheme(newTheme: Theme): void {
    if (newTheme === 'dark') {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  /**
   * Toggle between light and dark themes
   */
  function toggleTheme(): void {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
  }

  /**
   * Set specific theme
   */
  function setTheme(newTheme: Theme): void {
    theme.value = newTheme
  }

  // Watch for theme changes and persist
  watch(theme, (newTheme) => {
    applyTheme(newTheme)
    localStorage.setItem(THEME_KEY, newTheme)
  })

  // Initialize on mount
  onMounted(() => {
    initTheme()
  })

  return {
    theme,
    toggleTheme,
    setTheme,
    initTheme,
  }
}
