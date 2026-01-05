import { ref, computed, watch, onMounted } from 'vue'
import type { Language } from '@/types'
import zhCN from '@/locales/zh-CN'
import enUS from '@/locales/en-US'

const LANGUAGE_KEY = 'rabbit_panel_language'

// Available locales
const locales: Record<Language, typeof zhCN> = {
  'zh-CN': zhCN,
  'en-US': enUS,
}

// Global language state (shared across all components)
const language = ref<Language>('zh-CN')
const isInitialized = ref(false)

/**
 * i18n composable for managing internationalization
 * Persists language preference to localStorage
 */
export function useI18n() {
  // Current locale messages
  const messages = computed(() => locales[language.value])

  /**
   * Initialize language from localStorage or browser preference
   */
  function initLanguage(): void {
    if (isInitialized.value) return
    
    const stored = localStorage.getItem(LANGUAGE_KEY) as Language | null
    
    if (stored && (stored === 'zh-CN' || stored === 'en-US')) {
      language.value = stored
    } else {
      // Auto-detect browser language
      const browserLang = navigator.language
      if (browserLang.startsWith('zh')) {
        language.value = 'zh-CN'
      } else {
        language.value = 'en-US'
      }
    }
    
    isInitialized.value = true
  }

  /**
   * Set language
   */
  function setLanguage(lang: Language): void {
    language.value = lang
  }

  /**
   * Toggle between Chinese and English
   */
  function toggleLanguage(): void {
    language.value = language.value === 'zh-CN' ? 'en-US' : 'zh-CN'
  }

  /**
   * Get translation by key path (e.g., 'common.confirm')
   */
  function t(key: string): string {
    const keys = key.split('.')
    let result: unknown = messages.value
    
    for (const k of keys) {
      if (result && typeof result === 'object' && k in result) {
        result = (result as Record<string, unknown>)[k]
      } else {
        return key // Return key if translation not found
      }
    }
    
    return typeof result === 'string' ? result : key
  }

  // Watch for language changes and persist
  watch(language, (newLang) => {
    localStorage.setItem(LANGUAGE_KEY, newLang)
    // Update document lang attribute
    document.documentElement.lang = newLang
  })

  // Initialize on mount
  onMounted(() => {
    initLanguage()
  })

  return {
    language,
    messages,
    t,
    setLanguage,
    toggleLanguage,
    initLanguage,
  }
}
