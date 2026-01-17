import { createI18n } from 'vue-i18n'
import zhCn from './zh-CN'
import enUs from './en-US'

const i18n = createI18n({
    legacy: false,
    locale: 'zh-CN',
    fallbackLocale: 'en-US',
    messages: {
        'zh-CN': zhCn,
        'en-US': enUs
    }
})

export default i18n
