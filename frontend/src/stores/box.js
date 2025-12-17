import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useBoxStore = defineStore('box', () => {
  const items = ref([])
  const guardians = ref([])
  const loading = ref(false)
  const error = ref('')

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

  async function fetchItems() {
    try {
      const res = await fetch('/api/box/items', { credentials: 'include' })
      if (res.ok) {
        const data = await res.json()
        items.value = data.items || []
      }
    } catch (e) {
      error.value = 'Erro ao carregar itens'
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
      error.value = 'Erro ao carregar pessoas'
    }
  }

  async function fetchAll() {
    loading.value = true
    await Promise.all([fetchItems(), fetchGuardians()])
    loading.value = false
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
        items.value.unshift(item)
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
        return true
      }
    } catch (e) {
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
        guardians.value.unshift(guardian)
        return guardian
      } else {
        const errorText = await res.text()
        console.error('[Box Store] Failed to create guardian:', res.status, errorText)
        error.value = 'Erro ao adicionar pessoa'
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
      error.value = 'Erro ao remover pessoa'
    }
    return false
  }

  return {
    items,
    guardians,
    loading,
    error,
    counts,
    unifiedEntries,
    fetchItems,
    fetchGuardians,
    fetchAll,
    createItem,
    updateItem,
    deleteItem,
    createGuardian,
    deleteGuardian
  }
})

