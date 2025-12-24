// =============================================================================
// FAMLI - useAnalytics Composable
// =============================================================================
// Composable para rastreamento de eventos de analytics
// RESPEITA O CONSENTIMENTO DE COOKIES (LGPD/GDPR)
//
// Uso:
// import { useAnalytics } from '@/composables/useAnalytics'
// const { trackEvent, trackPageView } = useAnalytics()
//
// trackEvent('create_item', { type: 'memory' })
// trackPageView()
// =============================================================================

import { useRoute } from 'vue-router'
import { useCookieConsent } from './useCookieConsent'

/**
 * Composable para rastreamento de analytics
 */
export function useAnalytics() {
  const route = useRoute()
  const { analyticsAllowed } = useCookieConsent()

  /**
   * Verifica se o rastreamento está permitido
   */
  function isTrackingAllowed() {
    // Verificar flag global (definida pelo composable de cookies)
    if (window._famliAnalyticsDisabled) return false
    // Verificar consentimento
    return analyticsAllowed.value
  }

  /**
   * Rastreia um evento genérico
   * @param {string} eventType - Tipo do evento (create_item, edit_item, etc)
   * @param {Object} details - Detalhes adicionais do evento
   */
  async function trackEvent(eventType, details = {}) {
    // LGPD/GDPR: Só rastrear se o usuário consentiu
    if (!isTrackingAllowed()) {
      console.debug('Analytics tracking skipped: no consent')
      return
    }

    try {
      await fetch('/api/analytics/track', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          event_type: eventType,
          page: route?.path || window.location.pathname,
          details
        })
      })
    } catch (error) {
      // Silently ignore tracking errors - não deve afetar UX
      console.debug('Analytics tracking failed:', error)
    }
  }

  /**
   * Rastreia uma visualização de página
   */
  function trackPageView() {
    trackEvent('page_view')
  }

  /**
   * Rastreia criação de item
   * @param {string} type - Tipo do item (information, memory, trusted_person)
   */
  function trackCreateItem(type) {
    trackEvent('create_item', { item_type: type })
  }

  /**
   * Rastreia edição de item
   */
  function trackEditItem() {
    trackEvent('edit_item')
  }

  /**
   * Rastreia exclusão de item
   */
  function trackDeleteItem() {
    trackEvent('delete_item')
  }

  /**
   * Rastreia criação de guardião
   */
  function trackCreateGuardian() {
    trackEvent('create_guardian')
  }

  /**
   * Rastreia conclusão de guia
   * @param {string} cardId - ID do card do guia
   */
  function trackCompleteGuide(cardId) {
    trackEvent('complete_guide', { card_id: cardId })
  }

  /**
   * Rastreia login bem-sucedido
   */
  function trackLogin() {
    trackEvent('login')
  }

  /**
   * Rastreia registro bem-sucedido
   */
  function trackRegister() {
    trackEvent('register')
  }

  /**
   * Rastreia exportação de dados
   */
  function trackExportData() {
    trackEvent('export_data')
  }

  return {
    trackEvent,
    trackPageView,
    trackCreateItem,
    trackEditItem,
    trackDeleteItem,
    trackCreateGuardian,
    trackCompleteGuide,
    trackLogin,
    trackRegister,
    trackExportData
  }
}


