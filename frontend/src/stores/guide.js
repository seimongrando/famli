import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useGuideStore = defineStore('guide', () => {
  const cards = ref([])
  const progress = ref({})
  const loading = ref(false)

  const completedCount = computed(() => {
    return Object.values(progress.value).filter(p => p === 'completed').length
  })

  const progressPercentage = computed(() => {
    if (cards.value.length === 0) return 0
    return Math.round((completedCount.value / cards.value.length) * 100)
  })

  async function fetchCards() {
    try {
      const res = await fetch('/api/guide/cards', { credentials: 'include' })
      if (res.ok) {
        const data = await res.json()
        cards.value = data.cards || []
      }
    } catch (e) {
      // Erro silencioso
    }
  }

  async function fetchProgress() {
    try {
      const res = await fetch('/api/guide/progress', { credentials: 'include' })
      if (res.ok) {
        const data = await res.json()
        progress.value = {}
        for (const p of data.progress || []) {
          progress.value[p.card_id] = p.status
        }
      }
    } catch (e) {
      // Erro silencioso
    }
  }

  async function fetchAll() {
    loading.value = true
    await Promise.all([fetchCards(), fetchProgress()])
    loading.value = false
  }

  async function markProgress(cardId, status) {
    console.log('[Guide Store] Marking progress:', cardId, status)
    
    try {
      const res = await fetch(`/api/guide/progress/${cardId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ status })
      })
      
      if (res.ok) {
        progress.value[cardId] = status
        console.log('[Guide Store] Progress marked successfully:', cardId, status)
        return true
      } else {
        console.error('[Guide Store] Failed to mark progress:', res.status)
        // Ainda atualizar localmente para feedback imediato
        progress.value[cardId] = status
        return false
      }
    } catch (e) {
      console.error('[Guide Store] Error marking progress:', e)
      // Atualizar localmente mesmo com erro de rede
      progress.value[cardId] = status
      return false
    }
  }

  function getCardStatus(cardId) {
    return progress.value[cardId] || 'pending'
  }

  return {
    cards,
    progress,
    loading,
    completedCount,
    progressPercentage,
    fetchCards,
    fetchProgress,
    fetchAll,
    markProgress,
    getCardStatus
  }
})

