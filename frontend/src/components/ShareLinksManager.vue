<template>
  <div class="share-links-manager">
    <div class="section-header">
      <h3>🔗 {{ $t('share.title') }}</h3>
      <button @click="showCreateModal = true" class="btn-create">
        + {{ $t('share.create') }}
      </button>
    </div>

    <p class="section-description">
      {{ $t('share.description') }}
    </p>

    <!-- Lista de Links -->
    <div v-if="loading" class="loading">
      {{ $t('common.loading') }}
    </div>

    <div v-else-if="links.length === 0" class="empty-state">
      <div class="empty-icon">🔗</div>
      <p>{{ $t('share.empty') }}</p>
    </div>

    <div v-else class="links-list">
      <div v-for="link in links" :key="link.id" class="link-card" :class="{ inactive: !link.is_active }">
        <div class="link-header">
          <span class="link-type-icon">{{ getLinkTypeIcon(link.type) }}</span>
          <div class="link-info">
            <h4>{{ link.name }}</h4>
            <span class="link-type-badge" :class="link.type">
              {{ getLinkTypeName(link.type) }}
            </span>
          </div>
          <button @click="deleteLink(link.id)" class="btn-delete" :title="$t('common.delete')">
            🗑️
          </button>
        </div>

        <div class="link-details">
          <div class="link-url">
            <input type="text" :value="link.url" readonly @focus="$event.target.select()" />
            <button @click="copyLink(link.url)" class="btn-copy" :title="$t('share.copy')">
              📋
            </button>
          </div>

          <div class="link-stats">
            <span v-if="link.usage_count > 0">
              👁️ {{ link.usage_count }} {{ $t('share.accesses') }}
            </span>
            <span v-if="link.expires_at">
              ⏰ {{ $t('share.expires') }}: {{ formatDate(link.expires_at) }}
            </span>
            <span v-if="link.max_uses > 0">
              🔢 {{ $t('share.max_uses') }}: {{ link.max_uses }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal de Criação -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ $t('share.create_link') }}</h3>
          <button @click="showCreateModal = false" class="btn-close">✕</button>
        </div>

        <form @submit.prevent="createLink" class="create-form">
          <div class="form-group">
            <label>{{ $t('share.link_name') }}</label>
            <input 
              v-model="newLink.name" 
              type="text" 
              :placeholder="$t('share.link_name_placeholder')"
              required
            />
          </div>

          <div class="form-group">
            <label>{{ $t('share.link_type') }}</label>
            <select v-model="newLink.type">
              <option value="normal">📦 {{ $t('share.type_normal') }}</option>
              <option value="emergency">🚨 {{ $t('share.type_emergency') }}</option>
              <option value="memorial">🕊️ {{ $t('share.type_memorial') }}</option>
            </select>
          </div>

          <div class="form-group">
            <label>{{ $t('share.categories') }}</label>
            <div class="categories-grid">
              <label v-for="cat in categories" :key="cat.value" class="category-checkbox">
                <input type="checkbox" :value="cat.value" v-model="newLink.categories" />
                <span>{{ cat.icon }} {{ cat.label }}</span>
              </label>
            </div>
            <small>{{ $t('share.categories_hint') }}</small>
          </div>

          <div class="form-group">
            <label>👤 {{ $t('share.guardians') }}</label>
            <div v-if="guardians.length > 0" class="guardians-grid">
              <label v-for="guardian in guardians" :key="guardian.id" class="guardian-checkbox">
                <input type="checkbox" :value="guardian.id" v-model="newLink.guardianIds" />
                <span class="guardian-option">
                  <strong>{{ guardian.name }}</strong>
                  <small v-if="guardian.relationship">{{ guardian.relationship }}</small>
                </span>
              </label>
            </div>
            <div v-else class="no-guardians">
              <span>{{ $t('share.no_guardians') }}</span>
            </div>
            <small>{{ $t('share.guardians_hint') }}</small>
          </div>

          <div class="form-group">
            <label>{{ $t('share.pin_optional') }}</label>
            <input 
              v-model="newLink.pin" 
              type="password" 
              :placeholder="$t('share.pin_placeholder')"
              maxlength="10"
            />
            <small>{{ $t('share.pin_hint') }}</small>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>{{ $t('share.expires_in') }}</label>
              <select v-model="newLink.expiresIn">
                <option :value="0">{{ $t('share.never') }}</option>
                <option :value="7">7 {{ $t('share.days') }}</option>
                <option :value="30">30 {{ $t('share.days') }}</option>
                <option :value="90">90 {{ $t('share.days') }}</option>
                <option :value="365">1 {{ $t('share.year') }}</option>
              </select>
            </div>

            <div class="form-group">
              <label>{{ $t('share.max_uses') }}</label>
              <select v-model="newLink.maxUses">
                <option :value="0">{{ $t('share.unlimited') }}</option>
                <option :value="1">1</option>
                <option :value="5">5</option>
                <option :value="10">10</option>
                <option :value="50">50</option>
              </select>
            </div>
          </div>

          <div class="modal-actions">
            <button type="button" @click="showCreateModal = false" class="btn-cancel">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" class="btn-submit" :disabled="creating">
              {{ creating ? $t('common.creating') : $t('share.create') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Toast de Sucesso -->
    <div v-if="toast" class="toast" :class="toast.type">
      {{ toast.message }}
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useBoxStore } from '@/stores/box'

const { t } = useI18n()
const boxStore = useBoxStore()

const loading = ref(true)
const links = ref([])
const showCreateModal = ref(false)
const creating = ref(false)
const toast = ref(null)

const newLink = ref({
  name: '',
  type: 'normal',
  categories: [],
  guardianIds: [],
  pin: '',
  expiresIn: 0,
  maxUses: 0
})

// Guardiões do store
const guardians = computed(() => boxStore.guardians || [])

// Categorias com i18n
const categories = computed(() => [
  { value: 'health', label: t('categories.health'), icon: '🏥' },
  { value: 'finances', label: t('categories.finances'), icon: '💰' },
  { value: 'family', label: t('categories.family'), icon: '👨‍👩‍👧‍👦' },
  { value: 'documents', label: t('categories.documents'), icon: '📄' },
  { value: 'memories', label: t('categories.memories'), icon: '💭' },
])

onMounted(async () => {
  // Carregar guardiões se não estiverem carregados
  if (boxStore.guardians.length === 0) {
    await boxStore.fetchAll()
  }
  await fetchLinks()
})

async function fetchLinks() {
  try {
    loading.value = true
    const response = await fetch('/api/share/links', {
      credentials: 'include'
    })
    const data = await response.json()
    links.value = data.links || []
  } catch (err) {
    console.error('Error fetching links:', err)
  } finally {
    loading.value = false
  }
}

async function createLink() {
  try {
    creating.value = true
    
    const response = await fetch('/api/share/links', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({
        name: newLink.value.name,
        type: newLink.value.type,
        categories: newLink.value.categories,
        guardian_ids: newLink.value.guardianIds,
        pin: newLink.value.pin,
        expires_in: newLink.value.expiresIn,
        max_uses: newLink.value.maxUses
      })
    })

    if (!response.ok) {
      throw new Error('Failed to create link')
    }

    const created = await response.json()
    links.value.unshift(created)
    
    showCreateModal.value = false
    resetNewLink()
    showToast('success', t('share.created'))
    
    // Copiar automaticamente
    copyLink(created.url)
  } catch (err) {
    showToast('error', t('share.create_error'))
  } finally {
    creating.value = false
  }
}

async function deleteLink(linkId) {
  if (!confirm(t('share.delete_confirm'))) return
  
  try {
    const response = await fetch(`/api/share/links/${linkId}`, {
      method: 'DELETE',
      credentials: 'include'
    })

    if (response.ok) {
      links.value = links.value.filter(l => l.id !== linkId)
      showToast('success', t('share.deleted'))
    }
  } catch (err) {
    showToast('error', t('share.delete_error'))
  }
}

function copyLink(url) {
  navigator.clipboard.writeText(url)
  showToast('success', t('share.copied'))
}

function resetNewLink() {
  newLink.value = {
    name: '',
    type: 'normal',
    categories: [],
    guardianIds: [],
    pin: '',
    expiresIn: 0,
    maxUses: 0
  }
}

function getLinkTypeIcon(type) {
  const icons = { normal: '📦', emergency: '🚨', memorial: '🕊️' }
  return icons[type] || '📦'
}

function getLinkTypeName(type) {
  const names = { 
    normal: t('share.type_normal'), 
    emergency: t('share.type_emergency'), 
    memorial: t('share.type_memorial') 
  }
  return names[type] || type
}

function formatDate(date) {
  return new Date(date).toLocaleDateString('pt-BR')
}

function showToast(type, message) {
  toast.value = { type, message }
  setTimeout(() => { toast.value = null }, 3000)
}
</script>

<style scoped>
.share-links-manager {
  background: white;
  border-radius: 1rem;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.section-header h3 {
  margin: 0;
  color: #1e293b;
}

.section-description {
  color: #64748b;
  margin-bottom: 1.5rem;
}

.btn-create {
  padding: 0.5rem 1rem;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background 0.2s;
}

.btn-create:hover {
  background: var(--color-primary-light);
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: 2rem;
  color: #64748b;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 0.5rem;
}

/* Links List */
.links-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.link-card {
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  padding: 1rem;
  transition: border-color 0.2s;
}

.link-card:hover {
  border-color: var(--color-primary);
}

.link-card.inactive {
  opacity: 0.5;
}

.link-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.link-type-icon {
  font-size: 1.5rem;
}

.link-info {
  flex: 1;
}

.link-info h4 {
  margin: 0;
  color: #1e293b;
}

.link-type-badge {
  display: inline-block;
  font-size: 0.75rem;
  padding: 0.125rem 0.5rem;
  border-radius: 1rem;
  margin-top: 0.25rem;
}

.link-type-badge.normal {
  background: var(--color-primary-soft);
  color: var(--color-primary);
}

.link-type-badge.emergency {
  background: #fee2e2;
  color: #dc2626;
}

.link-type-badge.memorial {
  background: #f1f5f9;
  color: #475569;
}

.btn-delete {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1.25rem;
  opacity: 0.5;
  transition: opacity 0.2s;
}

.btn-delete:hover {
  opacity: 1;
}

.link-url {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.link-url input {
  flex: 1;
  padding: 0.5rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  background: #f8fafc;
}

.btn-copy {
  padding: 0.5rem;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  cursor: pointer;
}

.btn-copy:hover {
  background: #e2e8f0;
}

.link-stats {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
  font-size: 0.75rem;
  color: #64748b;
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: white;
  border-radius: 1rem;
  max-width: 500px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #e2e8f0;
}

.modal-header h3 {
  margin: 0;
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.25rem;
  cursor: pointer;
  color: #64748b;
}

.create-form {
  padding: 1.5rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #1e293b;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  font-size: 1rem;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-soft);
}

.form-group small {
  display: block;
  margin-top: 0.25rem;
  color: #64748b;
  font-size: 0.75rem;
}

.categories-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.5rem;
}

.category-checkbox {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
}

.category-checkbox:has(input:checked) {
  background: var(--color-primary-soft);
  border-color: var(--color-primary);
}

.category-checkbox input {
  width: auto;
}

/* Guardians Grid */
.guardians-grid {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.guardian-checkbox {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s;
}

.guardian-checkbox:hover {
  border-color: var(--color-primary);
  background: rgba(45, 90, 71, 0.02);
}

.guardian-checkbox:has(input:checked) {
  background: var(--color-primary-soft);
  border-color: var(--color-primary);
}

.guardian-checkbox input {
  width: auto;
  accent-color: var(--color-primary);
}

.guardian-option {
  display: flex;
  flex-direction: column;
}

.guardian-option strong {
  color: #1e293b;
  font-size: 0.9rem;
}

.guardian-option small {
  color: #64748b;
  font-size: 0.8rem;
}

.no-guardians {
  padding: 1rem;
  text-align: center;
  color: #64748b;
  font-size: 0.9rem;
  background: #f8fafc;
  border-radius: 0.5rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  margin-top: 1.5rem;
}

.btn-cancel {
  flex: 1;
  padding: 0.75rem;
  background: #f1f5f9;
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
}

.btn-submit {
  flex: 1;
  padding: 0.75rem;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-submit:hover:not(:disabled) {
  background: var(--color-primary-light);
}

.btn-submit:disabled {
  opacity: 0.5;
}

/* Toast */
.toast {
  position: fixed;
  bottom: 2rem;
  left: 50%;
  transform: translateX(-50%);
  padding: 0.75rem 1.5rem;
  border-radius: 0.5rem;
  color: white;
  z-index: 1001;
  animation: slideUp 0.3s ease;
}

.toast.success {
  background: var(--color-primary);
}

.toast.error {
  background: #ef4444;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translate(-50%, 1rem);
  }
  to {
    opacity: 1;
    transform: translate(-50%, 0);
  }
}

/* Loading */
.loading {
  text-align: center;
  padding: 2rem;
  color: #64748b;
}
</style>

