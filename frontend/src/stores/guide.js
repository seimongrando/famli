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
    try {
      const res = await fetch(`/api/guide/progress/${cardId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ status })
      })
      if (res.ok) {
        progress.value[cardId] = status
      }
    } catch (e) {
      // Erro silencioso
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

