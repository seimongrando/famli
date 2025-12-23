<!-- =============================================================================
  FAMLI - Feed da Caixa Famli
  =============================================================================
  Exibe todos os itens guardados pelo usuário (informações, memórias, pessoas).
  
  Funcionalidades:
  - Filtros por tipo
  - Edição de itens
  - Exclusão com confirmação modal
  - Formatação de datas e categorias
============================================================================== -->

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useBoxStore } from '../stores/box'
import ConfirmModal from './ConfirmModal.vue'
import EditItemModal from './EditItemModal.vue'

const { t, locale } = useI18n()
const boxStore = useBoxStore()

// Filtro ativo
const filter = ref('all')

// Estado dos modais
const showDeleteModal = ref(false)
const showEditModal = ref(false)
const selectedItem = ref(null)

// Referência para o container do feed
const feedContainer = ref(null)

// Scroll infinito - detectar quando usuário chega no final
function handleScroll() {
  if (!feedContainer.value) return
  
  const container = feedContainer.value
  const scrollPosition = container.scrollTop + container.clientHeight
  const scrollHeight = container.scrollHeight
  
  // Carregar mais quando estiver a 200px do final
  if (scrollHeight - scrollPosition < 200) {
    boxStore.loadMoreItems()
  }
}

// Usar Intersection Observer para detecção mais eficiente
const loadMoreTrigger = ref(null)
let observer = null

onMounted(() => {
  // Criar observer para o trigger de "carregar mais"
  observer = new IntersectionObserver(
    (entries) => {
      if (entries[0].isIntersecting && boxStore.itemsHasMore && !boxStore.loadingMore) {
        boxStore.loadMoreItems()
      }
    },
    { threshold: 0.1 }
  )
  
  // Observar o elemento trigger quando ele existir
  if (loadMoreTrigger.value) {
    observer.observe(loadMoreTrigger.value)
  }
})

onUnmounted(() => {
  if (observer) {
    observer.disconnect()
  }
})

const filters = [
  { id: 'all', labelKey: 'box.filters.all' },
  { id: 'info', labelKey: 'box.filters.info' },
  { id: 'guardian', labelKey: 'box.filters.guardian' },
  { id: 'memory', labelKey: 'box.filters.memory' }
]

const filteredEntries = computed(() => {
  if (filter.value === 'all') return boxStore.unifiedEntries
  return boxStore.unifiedEntries.filter(e => e.kind === filter.value)
})

// Mensagem de confirmação baseada no tipo
const deleteMessage = computed(() => {
  if (!selectedItem.value) return ''
  return selectedItem.value.kind === 'guardian'
    ? t('confirmations.deleteGuardianMessage')
    : t('confirmations.deleteItemMessage')
})

function formatDate(dateStr) {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return ''
  
  // Usar locale do usuário (pt-BR ou en)
  const userLocale = locale.value === 'pt-BR' ? 'pt-BR' : 'en-US'
  return date.toLocaleDateString(userLocale, {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric'
  })
}

function getTypeLabel(type) {
  const key = `itemTypes.${type}`
  return t(key)
}

function getCategoryLabel(category) {
  if (!category) return ''
  const key = `composer.categories.${category}`
  const translated = t(key)
  // Se a tradução retornar a própria chave, mostrar o valor original
  return translated === key ? category : translated
}

function getRelationshipLabel(relationship) {
  if (!relationship) return ''
  const key = `composer.relationships.${relationship}`
  const translated = t(key)
  return translated === key ? relationship : translated
}

// Detectar se um texto parece estar criptografado (base64 longo sem espaços, ou caracteres estranhos)
function looksEncrypted(text) {
  if (!text || typeof text !== 'string') return false
  
  // Se o texto é muito longo sem espaços e parece base64
  const hasNoSpaces = !text.includes(' ') && text.length > 50
  const looksLikeBase64 = /^[A-Za-z0-9+/=]{50,}$/.test(text)
  const hasNonPrintable = /[\x00-\x08\x0E-\x1F]/.test(text)
  
  return (hasNoSpaces && looksLikeBase64) || hasNonPrintable
}

// Obter texto seguro para exibição
function getSafeDisplayText(text, fallback = '') {
  if (!text) return fallback
  if (looksEncrypted(text)) {
    return t('errors.encryptedContent')
  }
  return text
}

// Abrir modal de edição
function openEditModal(entry) {
  selectedItem.value = entry
  showEditModal.value = true
}

// Abrir modal de exclusão
function openDeleteModal(entry) {
  selectedItem.value = entry
  showDeleteModal.value = true
}

// Confirmar exclusão
async function confirmDelete() {
  if (!selectedItem.value) return
  
  if (selectedItem.value.kind === 'guardian') {
    await boxStore.deleteGuardian(selectedItem.value.id)
  } else {
    await boxStore.deleteItem(selectedItem.value.id)
  }
  
  showDeleteModal.value = false
  selectedItem.value = null
}

// Cancelar exclusão
function cancelDelete() {
  showDeleteModal.value = false
  selectedItem.value = null
}

// Fechar modal de edição
function closeEditModal() {
  showEditModal.value = false
  selectedItem.value = null
}

// Ao salvar edição
function onEditSaved() {
  showEditModal.value = false
  selectedItem.value = null
}

// Copiar link de acesso do guardião
async function copyGuardianLink(guardian) {
  if (!guardian.access_token) return
  if (!guardian.has_pin) {
    alert(t('guardian.pinRequired'))
    return
  }
  
  const baseUrl = window.location.origin
  const link = `${baseUrl}/g/${guardian.access_token}`
  
  try {
    await navigator.clipboard.writeText(link)
    // Feedback visual simples
    alert(t('guardian.linkCopied'))
  } catch (err) {
    console.error('Erro ao copiar link:', err)
  }
}
</script>

<template>
  <div class="feed">
    <div class="feed__header">
      <div class="feed__title-row">
        <h2 class="feed__title">📦 {{ t('box.title') }}</h2>
        <span class="feed__count">{{ boxStore.unifiedEntries.length }} {{ t('common.items') }}</span>
      </div>
      <p class="feed__subtitle">{{ t('box.subtitle') }}</p>
    </div>

    <!-- Filters -->
    <div class="feed__filters">
      <button 
        v-for="f in filters"
        :key="f.id"
        :class="['chip', { 'chip--active': filter === f.id }]"
        @click="filter = f.id"
      >
        {{ t(f.labelKey) }}
      </button>
    </div>

    <!-- Empty State -->
    <div v-if="filteredEntries.length === 0" class="feed__empty">
      <span class="feed__empty-icon">📭</span>
      <p class="feed__empty-text">
        {{ filter === 'all' ? t('box.empty') : t('box.emptyFiltered') }}
      </p>
    </div>

    <!-- Entries -->
    <ul v-else class="feed__list">
      <li v-for="entry in filteredEntries" :key="entry.kind + '-' + entry.id" class="feed-item">
        <span class="feed-item__icon">{{ entry.icon }}</span>
        
        <div class="feed-item__content">
          <div class="feed-item__header">
            <h3 class="feed-item__title">{{ getSafeDisplayText(entry.title, '...') }}</h3>
            <span class="feed-item__type">{{ getTypeLabel(entry.type || entry.kind) }}</span>
          </div>
          
          <p v-if="entry.content && !looksEncrypted(entry.content)" class="feed-item__description">
            {{ entry.content.length > 120 ? entry.content.slice(0, 120) + '...' : entry.content }}
          </p>
          <p v-else-if="entry.content && looksEncrypted(entry.content)" class="feed-item__description feed-item__description--warning">
            ⚠️ {{ t('errors.encryptedContent') }}
          </p>
          
          <div class="feed-item__meta">
            <span v-if="entry.category" class="feed-item__category">
              {{ getCategoryLabel(entry.category) }}
            </span>
            <span v-if="entry.recipient" class="feed-item__recipient">
              {{ entry.recipient }}
            </span>
            <span v-if="entry.relationship" class="feed-item__relationship">
              {{ getRelationshipLabel(entry.relationship) }}
            </span>
            <span class="feed-item__date">{{ formatDate(entry.updated_at || entry.created_at) }}</span>
          </div>
        </div>
        
        <div class="feed-item__actions">
          <!-- Botão de link apenas para guardiões -->
          <button 
            v-if="entry.kind === 'guardian' && entry.access_token"
            class="btn btn--ghost btn--small btn--icon btn--primary-text" 
            @click="copyGuardianLink(entry)"
            :title="t('guardian.copyLink')"
          >
            🔗
          </button>
          <button 
            class="btn btn--ghost btn--small btn--icon" 
            @click="openEditModal(entry)"
            :title="t('common.edit')"
          >
            ✏️
          </button>
          <button 
            class="btn btn--ghost btn--small btn--icon btn--danger-text" 
            @click="openDeleteModal(entry)"
            :title="t('common.delete')"
          >
            🗑️
          </button>
        </div>
      </li>
    </ul>
    
    <!-- Trigger para paginação infinita -->
    <div 
      v-if="boxStore.itemsHasMore" 
      ref="loadMoreTrigger" 
      class="load-more-trigger"
    >
      <div v-if="boxStore.loadingMore" class="loading-more">
        <span class="spinner-small"></span>
        {{ t('common.loading') }}
      </div>
      <button 
        v-else 
        @click="boxStore.loadMoreItems()" 
        class="btn btn--ghost btn--load-more"
      >
        {{ t('box.loadMore') }}
      </button>
    </div>
    
    <!-- Indicador de fim da lista -->
    <div v-if="!boxStore.itemsHasMore && filteredEntries.length > 0" class="end-of-list">
      {{ t('box.endOfList') }}
    </div>
    
    <!-- Modal de Confirmação de Exclusão -->
    <ConfirmModal
      :show="showDeleteModal"
      :title="t('confirmations.deleteTitle')"
      :message="deleteMessage"
      :confirm-text="t('common.delete')"
      type="danger"
      @confirm="confirmDelete"
      @cancel="cancelDelete"
    />
    
    <!-- Modal de Edição -->
    <EditItemModal
      :show="showEditModal"
      :item="selectedItem"
      @save="onEditSaved"
      @close="closeEditModal"
    />
  </div>
</template>

<style scoped>
.feed {
  background: var(--color-card);
  border-radius: var(--radius-xl);
  padding: var(--space-xl);
  border: 1px solid var(--color-border-light);
}

.feed__header {
  margin-bottom: var(--space-lg);
}

.feed__title-row {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  margin-bottom: var(--space-xs);
}

.feed__title {
  font-size: var(--font-size-lg);
  margin: 0;
}

.feed__count {
  padding: 2px 10px;
  background: var(--color-primary-soft);
  border-radius: var(--radius-full);
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-primary);
}

.feed__subtitle {
  color: var(--color-text-soft);
  margin: 0;
  font-size: var(--font-size-sm);
}

.feed__filters {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
  margin-bottom: var(--space-lg);
}

.feed__empty {
  text-align: center;
  padding: var(--space-2xl) var(--space-lg);
}

.feed__empty-icon {
  font-size: 3rem;
  display: block;
  margin-bottom: var(--space-md);
}

.feed__empty-text {
  color: var(--color-text-muted);
  margin: 0;
}

.feed__list {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.feed-item {
  display: grid;
  grid-template-columns: 48px 1fr auto;
  gap: var(--space-md);
  padding: var(--space-md);
  background: var(--color-bg);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-light);
  transition: all var(--transition-fast);
}

.feed-item:hover {
  background: var(--color-bg-warm);
  border-color: var(--color-border);
}

.feed-item__icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  background: var(--color-card);
  border-radius: var(--radius-md);
}

.feed-item__content {
  min-width: 0;
  overflow: hidden;
}

.feed-item__header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  margin-bottom: var(--space-xs);
  flex-wrap: nowrap;
  overflow: hidden;
}

.feed-item__title {
  font-size: var(--font-size-base);
  font-weight: 600;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.feed-item__type {
  padding: 2px 8px;
  background: var(--color-primary-soft);
  border-radius: var(--radius-full);
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--color-primary);
  flex-shrink: 0;
  white-space: nowrap;
}

.feed-item__description {
  color: var(--color-text-soft);
  font-size: var(--font-size-sm);
  margin: 0 0 var(--space-sm);
  line-height: 1.5;
  /* Limitar a 2 linhas */
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

.feed-item__description--warning {
  color: var(--color-warning, #d97706);
  font-style: italic;
}

.feed-item__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
  font-size: 0.75rem;
  color: var(--color-text-muted);
  overflow: hidden;
}

.feed-item__category,
.feed-item__recipient,
.feed-item__relationship {
  padding: 2px 6px;
  background: var(--color-bg-warm);
  border-radius: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 150px;
}

.feed-item__date {
  flex-shrink: 0;
  white-space: nowrap;
}

.feed-item__actions {
  display: flex;
  align-items: flex-start;
  gap: var(--space-xs);
}

.btn--icon {
  padding: var(--space-xs);
  min-width: 32px;
}

.btn--danger-text:hover {
  color: #dc2626;
}

@media (max-width: 600px) {
  .feed-item {
    grid-template-columns: 40px 1fr;
  }
  
  .feed-item__actions {
    grid-column: 1 / -1;
    justify-content: flex-end;
  }
}

/* Paginação infinita */
.load-more-trigger {
  display: flex;
  justify-content: center;
  padding: var(--space-lg);
}

.loading-more {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  color: var(--color-text-soft);
  font-size: var(--font-size-sm);
}

.spinner-small {
  width: 16px;
  height: 16px;
  border: 2px solid var(--color-border);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.btn--load-more {
  font-size: var(--font-size-sm);
  color: var(--color-primary);
}

.btn--load-more:hover {
  text-decoration: underline;
}

.end-of-list {
  text-align: center;
  padding: var(--space-lg);
  color: var(--color-text-soft);
  font-size: var(--font-size-sm);
  font-style: italic;
}
</style>
