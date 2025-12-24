// =============================================================================
// FAMLI - useCookieConsent Composable
// =============================================================================
// Gerenciamento de consentimento de cookies conforme LGPD e GDPR
//
// Categorias de cookies:
// - necessary: Cookies essenciais (sempre habilitados)
// - analytics: Cookies de análise e métricas
// - preferences: Cookies de preferências do usuário
//
// Uso:
// import { useCookieConsent } from '@/composables/useCookieConsent'
// const { hasConsent, preferences, showBanner, acceptAll, rejectAll, savePreferences } = useCookieConsent()
// =============================================================================

import { ref, reactive, computed, watch } from 'vue'

const STORAGE_KEY = 'famli:cookie-consent'
const CONSENT_VERSION = '1.0' // Incrementar quando políticas mudarem

// Estado global compartilhado entre todas as instâncias
const globalState = {
  initialized: false,
  showBanner: ref(false),
  consentGiven: ref(false),
  preferences: reactive({
    necessary: true, // Sempre true - cookies essenciais
    analytics: false,
    preferences: false
  }),
  consentDate: ref(null),
  consentVersion: ref(null)
}

/**
 * Composable para gerenciamento de consentimento de cookies
 */
export function useCookieConsent() {
  // Inicializar apenas uma vez
  if (!globalState.initialized) {
    loadConsent()
    globalState.initialized = true
  }

  /**
   * Carrega o consentimento salvo do localStorage
   */
  function loadConsent() {
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored) {
        const data = JSON.parse(stored)
        
        // Verificar se a versão do consentimento ainda é válida
        if (data.version !== CONSENT_VERSION) {
          // Versão mudou, pedir novo consentimento
          globalState.showBanner.value = true
          return
        }
        
        globalState.consentGiven.value = true
        globalState.preferences.analytics = data.analytics || false
        globalState.preferences.preferences = data.preferences || false
        globalState.consentDate.value = data.date || null
        globalState.consentVersion.value = data.version
        globalState.showBanner.value = false
      } else {
        // Nenhum consentimento salvo, mostrar banner
        globalState.showBanner.value = true
      }
    } catch (error) {
      console.error('Error loading cookie consent:', error)
      globalState.showBanner.value = true
    }
  }

  /**
   * Salva o consentimento no localStorage
   */
  function saveConsent() {
    try {
      const data = {
        version: CONSENT_VERSION,
        date: new Date().toISOString(),
        necessary: true,
        analytics: globalState.preferences.analytics,
        preferences: globalState.preferences.preferences
      }
      localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
      globalState.consentGiven.value = true
      globalState.consentDate.value = data.date
      globalState.consentVersion.value = data.version
      globalState.showBanner.value = false
      
      // Aplicar as preferências
      applyPreferences()
    } catch (error) {
      console.error('Error saving cookie consent:', error)
    }
  }

  /**
   * Aceita todos os cookies
   */
  function acceptAll() {
    globalState.preferences.analytics = true
    globalState.preferences.preferences = true
    saveConsent()
  }

  /**
   * Rejeita todos os cookies opcionais (mantém apenas necessários)
   */
  function rejectAll() {
    globalState.preferences.analytics = false
    globalState.preferences.preferences = false
    saveConsent()
  }

  /**
   * Salva as preferências atuais
   */
  function savePreferences() {
    saveConsent()
  }

  /**
   * Abre o banner para o usuário revisar suas preferências
   */
  function openPreferences() {
    globalState.showBanner.value = true
  }

  /**
   * Aplica as preferências de cookies
   * Aqui você pode integrar com serviços de analytics, etc.
   */
  function applyPreferences() {
    // Desabilitar analytics se não consentido
    if (!globalState.preferences.analytics) {
      // Limpar cookies de analytics existentes (se houver)
      // Desabilitar rastreamento
      window._famliAnalyticsDisabled = true
    } else {
      window._famliAnalyticsDisabled = false
    }
  }

  /**
   * Verifica se há consentimento para uma categoria específica
   */
  function hasConsent(category) {
    if (category === 'necessary') return true
    return globalState.preferences[category] || false
  }

  /**
   * Retorna se analytics está permitido
   */
  const analyticsAllowed = computed(() => {
    return globalState.consentGiven.value && globalState.preferences.analytics
  })

  /**
   * Retorna se preferences cookies estão permitidos
   */
  const preferencesAllowed = computed(() => {
    return globalState.consentGiven.value && globalState.preferences.preferences
  })

  return {
    // Estado
    showBanner: globalState.showBanner,
    consentGiven: globalState.consentGiven,
    preferences: globalState.preferences,
    consentDate: globalState.consentDate,
    
    // Computed
    analyticsAllowed,
    preferencesAllowed,
    
    // Métodos
    hasConsent,
    acceptAll,
    rejectAll,
    savePreferences,
    openPreferences,
    loadConsent
  }
}

