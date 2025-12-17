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
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useBoxStore } from '../stores/box'
import ConfirmModal from './ConfirmModal.vue'
import EditItemModal from './EditItemModal.vue'

const { t } = useI18n()
const boxStore = useBoxStore()

// Filtro ativo
const filter = ref('all')

// Estado dos modais
const showDeleteModal = ref(false)
const showEditModal = ref(false)
const selectedItem = ref(null)

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
  return date.toLocaleDateString(undefined, {
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
            <h3 class="feed-item__title">{{ entry.title }}</h3>
            <span class="feed-item__type">{{ getTypeLabel(entry.type || entry.kind) }}</span>
          </div>
          
          <p v-if="entry.content" class="feed-item__description">
            {{ entry.content.length > 120 ? entry.content.slice(0, 120) + '...' : entry.content }}
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
}

.feed-item__header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  margin-bottom: var(--space-xs);
  flex-wrap: wrap;
}

.feed-item__title {
  font-size: var(--font-size-base);
  font-weight: 600;
  margin: 0;
}

.feed-item__type {
  padding: 2px 8px;
  background: var(--color-primary-soft);
  border-radius: var(--radius-full);
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--color-primary);
}

.feed-item__description {
  color: var(--color-text-soft);
  font-size: var(--font-size-sm);
  margin: 0 0 var(--space-sm);
  line-height: 1.5;
}

.feed-item__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.feed-item__category,
.feed-item__recipient,
.feed-item__relationship {
  padding: 2px 6px;
  background: var(--color-bg-warm);
  border-radius: 4px;
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
</style>
