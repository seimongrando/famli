// =============================================================================
// FAMLI - Store da Caixa (Pinia)
// =============================================================================
// Gerencia o estado dos itens e guardiões com suporte a:
// - Paginação infinita (cursor-based)
// - Cache local
// - Otimizações de performance
// - Tradução de erros de negócio
// =============================================================================

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import i18n from '../i18n'

// Mapeamento de erros do backend para chaves i18n
const errorMap = {
  // Erros em inglês
  'title is required': 'apiErrors.title_required',
  'title too long': 'apiErrors.title_too_long',
  'content too long': 'apiErrors.content_too_long',
  'name is required': 'apiErrors.name_required',
  'name too long': 'apiErrors.name_too_long',
  'invalid email': 'apiErrors.email_invalid',
  'email too long': 'apiErrors.email_too_long',
  'phone too long': 'apiErrors.phone_too_long',
  'pin is required': 'apiErrors.pin_required',
  'pin too short': 'apiErrors.pin_too_short',
  'pin too long': 'apiErrors.pin_too_long',
  'recipient too long': 'apiErrors.recipient_too_long',
  'invalid type': 'apiErrors.type_invalid',
  'item not found': 'apiErrors.item_not_found',
  'guardian not found': 'apiErrors.guardian_not_found',
  'duplicate email': 'apiErrors.duplicate_guardian_email',
  'unauthorized': 'apiErrors.unauthorized',
  'rate limit': 'apiErrors.rate_limit',
  // Erros em português (compatibilidade legada)
  'título é obrigatório': 'apiErrors.title_required',
  'título muito longo': 'apiErrors.title_too_long',
  'conteúdo muito longo': 'apiErrors.content_too_long',
  'nome é obrigatório': 'apiErrors.name_required',
  'nome muito longo': 'apiErrors.name_too_long',
  'e-mail inválido': 'apiErrors.email_invalid',
  'e-mail muito longo': 'apiErrors.email_too_long',
  'telefone muito longo': 'apiErrors.phone_too_long',
  'pin é obrigatório': 'apiErrors.pin_required',
  'pin muito curto': 'apiErrors.pin_too_short',
  'pin muito longo': 'apiErrors.pin_too_long',
  'destinatário muito longo': 'apiErrors.recipient_too_long',
  'tipo inválido': 'apiErrors.type_invalid',
  'item não encontrado': 'apiErrors.item_not_found',
  'guardião não encontrado': 'apiErrors.guardian_not_found',
  'e-mail já cadastrado': 'apiErrors.duplicate_guardian_email',
  'não autorizado': 'apiErrors.unauthorized',
  'muitas requisições': 'apiErrors.rate_limit'
}

// Função para traduzir erros do backend
function translateError(backendError) {
  const { t } = i18n.global
  
  if (!backendError) return t('apiErrors.generic')
  
  const lowerError = backendError.toLowerCase()
  
  for (const [key, translationKey] of Object.entries(errorMap)) {
    if (lowerError.includes(key.toLowerCase())) {
      return t(translationKey)
    }
  }
  
  // Verificar se é erro de conexão/rede
  if (lowerError.includes('network') || lowerError.includes('fetch') || lowerError.includes('connection')) {
    return t('apiErrors.network_error')
  }
  
  // Verificar se é erro de servidor
  if (lowerError.includes('server') || lowerError.includes('500')) {
    return t('apiErrors.server_error')
  }
  
  // Se não encontrou tradução, retornar o erro original ou genérico
  return t('apiErrors.generic')
}

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
      error.value = translateError('network error')
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
      error.value = translateError('network error')
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

  async function readErrorMessage(res, fallback) {
    try {
      const data = await res.json()
      if (data?.error) return data.error
    } catch (_) {
      // ignore json parse errors
    }
    try {
      const text = await res.text()
      if (text) return text
    } catch (_) {
      // ignore text errors
    }
    return fallback
  }

  async function createItem(payload) {
    try {
      const res = await fetch('/api/box/items', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(payload)
      })
      
      if (res.ok) {
        const item = await res.json()
        // Adicionar no início da lista
        items.value.unshift(item)
        itemsTotal.value++
        error.value = ''
        return item
      } else {
        const errorText = await readErrorMessage(res, 'server error')
        error.value = translateError(errorText)
      }
    } catch (e) {
      error.value = translateError('network error')
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
        error.value = ''
        return updated
      } else {
        const errorText = await readErrorMessage(res, 'server error')
        error.value = translateError(errorText)
      }
    } catch (e) {
      error.value = translateError('network error')
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
        error.value = ''
        return true
      } else {
        const errorText = await readErrorMessage(res, 'server error')
        error.value = translateError(errorText)
      }
    } catch (e) {
      error.value = translateError('network error')
    }
    return false
  }

  async function createGuardian(payload) {
    try {
      const res = await fetch('/api/guardians', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(payload)
      })
      
      if (res.ok) {
        const guardian = await res.json()
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
        const errorText = data?.error || (await res.text()) || 'server error'
        error.value = translateError(errorText)
      }
    } catch (e) {
      error.value = translateError('network error')
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
        error.value = ''
        return true
      } else {
        const errorText = await readErrorMessage(res, 'server error')
        error.value = translateError(errorText)
      }
    } catch (e) {
      error.value = translateError('network error')
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
