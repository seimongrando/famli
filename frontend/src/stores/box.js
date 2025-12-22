// =============================================================================
// FAMLI - Store da Caixa (Pinia)
// =============================================================================
// Gerencia o estado dos itens e guardiões com suporte a:
// - Paginação infinita (cursor-based)
// - Cache local
// - Otimizações de performance
// =============================================================================

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useBoxStore = defineStore('box', () => {
  // Estado dos itens
  const items = ref([])
  const itemsCursor = ref(null)
  const itemsHasMore = ref(true)
  const itemsTotal = ref(0)

  // Estado dos guardiões
  const guardians = ref([])
  
  // Estado geral
  const loading = ref(false)
  const loadingMore = ref(false)
  const error = ref('')

  // Tamanho da página
  const PAGE_SIZE = 20

  // Contagem por tipo
  const counts = computed(() => {
    const infos = items.value.filter(i => i.type === 'info' || i.type === 'location' || i.type === 'access' || i.type === 'routine').length
    const memories = items.value.filter(i => i.type === 'memory' || i.type === 'note').length
    const people = guardians.value.length
    return { infos, memories, people, total: infos + memories + people }
  })

  // Itens unificados para o feed
  const unifiedEntries = computed(() => {
    const itemEntries = items.value.map(item => ({
      ...item,
      kind: item.type === 'memory' || item.type === 'note' ? 'memory' : 'info',
      icon: getIcon(item.type),
      typeLabel: getTypeLabel(item.type)
    }))

    const guardianEntries = guardians.value.map(g => ({
      ...g,
      kind: 'guardian',
      icon: '👤',
      typeLabel: 'Pessoa de confiança',
      title: g.name,
      content: g.relationship ? `${g.relationship} • ${g.email}` : g.email
    }))

    return [...itemEntries, ...guardianEntries].sort((a, b) => {
      const da = new Date(a.updated_at || a.created_at || 0).getTime()
      const db = new Date(b.updated_at || b.created_at || 0).getTime()
      return db - da
    })
  })

  function getIcon(type) {
    const icons = {
      info: '📋',
      memory: '💝',
      note: '📝',
      access: '🔑',
      routine: '🔄',
      location: '📍'
    }
    return icons[type] || '📄'
  }

  function getTypeLabel(type) {
    const labels = {
      info: 'Informação importante',
      memory: 'Memória',
      note: 'Nota pessoal',
      access: 'Instruções de acesso',
      routine: 'Rotina',
      location: 'Localização'
    }
    return labels[type] || 'Item'
  }

  // Buscar primeira página de itens
  async function fetchItems() {
    try {
      const res = await fetch(`/api/box/items?limit=${PAGE_SIZE}`, { 
        credentials: 'include' 
      })
      if (res.ok) {
        const data = await res.json()
        items.value = data.items || []
        itemsCursor.value = data.next_cursor || null
        itemsHasMore.value = data.has_more || false
        itemsTotal.value = data.total || 0
      }
    } catch (e) {
      console.error('[Box Store] Erro ao carregar itens:', e)
      error.value = 'Erro ao carregar itens'
    }
  }

  // Carregar mais itens (paginação infinita)
  async function loadMoreItems() {
    if (!itemsHasMore.value || loadingMore.value) return

    loadingMore.value = true
    try {
      const url = `/api/box/items?limit=${PAGE_SIZE}${itemsCursor.value ? `&cursor=${itemsCursor.value}` : ''}`
      const res = await fetch(url, { credentials: 'include' })
      
      if (res.ok) {
        const data = await res.json()
        // Adicionar itens sem duplicatas
        const newItems = (data.items || []).filter(
          newItem => !items.value.some(existing => existing.id === newItem.id)
        )
        items.value = [...items.value, ...newItems]
        itemsCursor.value = data.next_cursor || null
        itemsHasMore.value = data.has_more || false
      }
    } catch (e) {
      console.error('[Box Store] Erro ao carregar mais itens:', e)
    } finally {
      loadingMore.value = false
    }
  }

  async function fetchGuardians() {
    try {
      const res = await fetch('/api/guardians', { credentials: 'include' })
      if (res.ok) {
        const data = await res.json()
        guardians.value = data.guardians || []
      }
    } catch (e) {
      console.error('[Box Store] Erro ao carregar pessoas:', e)
      error.value = 'Erro ao carregar pessoas'
    }
  }

  async function fetchAll() {
    loading.value = true
    await Promise.all([fetchItems(), fetchGuardians()])
    loading.value = false
  }

  // Recarregar todos os dados (reset de paginação)
  async function refresh() {
    itemsCursor.value = null
    itemsHasMore.value = true
    await fetchAll()
  }

  async function createItem(payload) {
    console.log('[Box Store] Creating item:', payload.type, payload.title)
    
    try {
      const res = await fetch('/api/box/items', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(payload)
      })
      
      if (res.ok) {
        const item = await res.json()
        console.log('[Box Store] Item created successfully:', item.id)
        // Adicionar no início da lista
        items.value.unshift(item)
        itemsTotal.value++
        return item
      } else {
        const errorText = await res.text()
        console.error('[Box Store] Failed to create item:', res.status, errorText)
        error.value = 'Erro ao salvar item'
      }
    } catch (e) {
      console.error('[Box Store] Error creating item:', e)
      error.value = 'Erro ao salvar'
    }
    return null
  }

  async function updateItem(id, payload) {
    try {
      const res = await fetch(`/api/box/items/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(payload)
      })
      if (res.ok) {
        const updated = await res.json()
        const idx = items.value.findIndex(i => i.id === id)
        if (idx !== -1) items.value[idx] = updated
        return updated
      }
    } catch (e) {
      console.error('[Box Store] Error updating item:', e)
      error.value = 'Erro ao atualizar'
    }
    return null
  }

  async function deleteItem(id) {
    try {
      const res = await fetch(`/api/box/items/${id}`, {
        method: 'DELETE',
        credentials: 'include'
      })
      if (res.ok) {
        items.value = items.value.filter(i => i.id !== id)
        itemsTotal.value = Math.max(0, itemsTotal.value - 1)
        return true
      }
    } catch (e) {
      console.error('[Box Store] Error deleting item:', e)
      error.value = 'Erro ao excluir'
    }
    return false
  }

  async function createGuardian(payload) {
    console.log('[Box Store] Creating guardian:', payload.name)
    
    try {
      const res = await fetch('/api/guardians', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(payload)
      })
      
      if (res.ok) {
        const guardian = await res.json()
        console.log('[Box Store] Guardian created successfully:', guardian.id)
        error.value = ''
        guardians.value.unshift(guardian)
        return guardian
      } else {
        let data = null
        try {
          data = await res.json()
        } catch (_) {
          // ignore json parse errors
        }
        const errorText = data?.error || (await res.text())
        console.error('[Box Store] Failed to create guardian:', res.status, errorText)
        error.value = errorText || 'Erro ao adicionar pessoa'
      }
    } catch (e) {
      console.error('[Box Store] Error creating guardian:', e)
      error.value = 'Erro ao adicionar pessoa'
    }
    return null
  }

  async function deleteGuardian(id) {
    try {
      const res = await fetch(`/api/guardians/${id}`, {
        method: 'DELETE',
        credentials: 'include'
      })
      if (res.ok) {
        guardians.value = guardians.value.filter(g => g.id !== id)
        return true
      }
    } catch (e) {
      console.error('[Box Store] Error deleting guardian:', e)
      error.value = 'Erro ao remover pessoa'
    }
    return false
  }

  return {
    // Estado
    items,
    guardians,
    loading,
    loadingMore,
    error,
    itemsHasMore,
    itemsTotal,
    
    // Computeds
    counts,
    unifiedEntries,
    
    // Ações
    fetchItems,
    fetchGuardians,
    fetchAll,
    refresh,
    loadMoreItems,
    createItem,
    updateItem,
    deleteItem,
    createGuardian,
    deleteGuardian
  }
})
