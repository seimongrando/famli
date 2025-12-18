import { createI18n } from 'vue-i18n'
import ptBR from './locales/pt-BR.json'
import en from './locales/en.json'

// Detectar idioma do navegador
function getDefaultLocale() {
  const stored = localStorage.getItem('famli-locale')
  if (stored) return stored
  
  const browserLang = navigator.language || navigator.userLanguage
  if (browserLang.startsWith('pt')) return 'pt-BR'
  if (browserLang.startsWith('es')) return 'es' // futuro
  return 'en'
}

const i18n = createI18n({
  legacy: false, // Usar Composition API
  locale: getDefaultLocale(),
  fallbackLocale: 'en',
  messages: {
    'pt-BR': ptBR,
    'en': en
  }
})

// Helper para trocar idioma
export function setLocale(locale) {
  i18n.global.locale.value = locale
  localStorage.setItem('famli-locale', locale)
  document.documentElement.lang = locale
}

export function getLocale() {
  return i18n.global.locale.value
}

export const availableLocales = [
  { code: 'pt-BR', name: 'Português', flag: '🇧🇷' },
  { code: 'en', name: 'English', flag: '🇺🇸' }
]

export default i18n


