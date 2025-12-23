<!-- =============================================================================
  FAMLI - Modal de Edição de Item
  =============================================================================
  Modal para editar itens existentes (informações, memórias, pessoas).
  
  Props:
  - show: boolean - Controla visibilidade do modal
  - item: Object - Item a ser editado
  
  Events:
  - save: Emitido com o item atualizado
  - close: Emitido quando o modal é fechado
============================================================================== -->

<script setup>
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useBoxStore } from '../stores/box'
import CharCounter from './CharCounter.vue'
import ErrorModal from './ErrorModal.vue'

const { t } = useI18n()
const boxStore = useBoxStore()

// Limites de caracteres
const LIMITS = {
  title: 255,
  content: 10000,
  name: 255,
  email: 254,
  phone: 20,
  recipient: 255
}

// Função para traduzir erros do backend
function translateError(error) {
  const errorMap = {
    'title is required': 'apiErrors.title_required',
    'title too long': 'apiErrors.title_too_long',
    'content too long': 'apiErrors.content_too_long',
    'name is required': 'apiErrors.name_required',
    'name too long': 'apiErrors.name_too_long',
    'invalid email': 'apiErrors.email_invalid',
    'email too long': 'apiErrors.email_too_long',
    'phone too long': 'apiErrors.phone_too_long',
    'recipient too long': 'apiErrors.recipient_too_long',
    'invalid type': 'apiErrors.type_invalid',
    'item not found': 'apiErrors.item_not_found',
    'unauthorized': 'apiErrors.unauthorized',
    'Título é obrigatório': 'apiErrors.title_required',
    'Título muito longo': 'apiErrors.title_too_long',
    'Conteúdo muito longo': 'apiErrors.content_too_long',
    'Nome é obrigatório': 'apiErrors.name_required',
    'Nome muito longo': 'apiErrors.name_too_long',
    'E-mail inválido': 'apiErrors.email_invalid',
    'E-mail muito longo': 'apiErrors.email_too_long',
    'Telefone muito longo': 'apiErrors.phone_too_long',
    'Destinatário muito longo': 'apiErrors.recipient_too_long',
    'Tipo inválido': 'apiErrors.type_invalid',
    'Item não encontrado': 'apiErrors.item_not_found',
    'Não autorizado': 'apiErrors.unauthorized'
  }

  if (!error) return t('apiErrors.generic')
  
  const lowerError = error.toLowerCase()
  for (const [key, translationKey] of Object.entries(errorMap)) {
    if (lowerError.includes(key.toLowerCase())) {
      return t(translationKey)
    }
  }
  
  return t('apiErrors.generic')
}

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  item: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['save', 'close'])

const saving = ref(false)
const errorMessage = ref('')
const showErrorModal = ref(false)
const errorModalMessage = ref('')

// Mostrar erro em modal
function showError(error) {
  errorModalMessage.value = translateError(error)
  showErrorModal.value = true
}

const form = ref({
  title: '',
  content: '',
  category: '',
  name: '',
  email: '',
  phone: '',
  relationship: '',
  recipient: '',
  isShared: false
})

const categories = ['saude', 'financas', 'documentos', 'casa', 'familia', 'outro']
const relationships = ['filho', 'neto', 'conjuge', 'irmao', 'amigo', 'outro']

// Tipo do item atual
const itemKind = computed(() => {
  if (!props.item) return null
  if (props.item.kind === 'guardian') return 'guardian'
  if (props.item.type === 'memory') return 'memory'
  return 'info'
})

// Observar mudanças no item para preencher o form
watch(() => props.item, (newItem) => {
  if (newItem) {
    form.value = {
      title: newItem.title || '',
      content: newItem.content || '',
      category: newItem.category || '',
      name: newItem.name || '',
      email: newItem.email || '',
      phone: newItem.phone || '',
      relationship: newItem.relationship || '',
      recipient: newItem.recipient || '',
      isShared: newItem.is_shared || false
    }
  }
}, { immediate: true })

// Prevenir scroll do body quando modal está aberto
watch(() => props.show, (show) => {
  if (show) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})

async function handleSave() {
  if (!props.item) return
  
  // Validação frontend
  if (!form.value.title && itemKind.value !== 'guardian') {
    errorMessage.value = t('apiErrors.title_required')
    return
  }
  
  saving.value = true
  errorMessage.value = ''
  
  try {
    if (itemKind.value === 'guardian') {
      // Atualizar guardian não está implementado no store ainda
      // Por enquanto, apenas fechar
      emit('close')
    } else {
      const result = await boxStore.updateItem(props.item.id, {
        title: form.value.title,
        content: form.value.content,
        category: form.value.category,
        recipient: form.value.recipient,
        is_shared: form.value.isShared
      })
      
      if (result) {
        emit('save', result)
        emit('close')
      } else {
        showError(boxStore.error)
      }
    }
  } catch (error) {
    showError(boxStore.error || error.message)
  } finally {
    saving.value = false
  }
}

function handleClose() {
  emit('close')
}

function handleBackdropClick(e) {
  if (e.target === e.currentTarget) {
    handleClose()
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show && item" class="modal-overlay" @click="handleBackdropClick">
        <div class="modal modal--large" role="dialog" aria-modal="true">
          <div class="modal__header">
            <h2 class="modal__title">{{ t('edit.title') }}</h2>
            <button class="modal__close" @click="handleClose">✕</button>
          </div>
          
          <!-- Info Form -->
          <form v-if="itemKind === 'info'" @submit.prevent="handleSave" class="modal__form">
            <div class="form-group">
              <label class="form-label">{{ t('composer.info.titleLabel') }}</label>
              <input 
                v-model="form.title"
                type="text"
                class="form-input"
                :class="{ 'form-input--error': form.title.length > LIMITS.title }"
                :maxlength="LIMITS.title"
                required
              />
              <div class="form-hint-row">
                <CharCounter :current="form.title.length" :max="LIMITS.title" />
              </div>
            </div>
            
            <div class="form-group">
              <label class="form-label">{{ t('composer.info.categoryLabel') }}</label>
              <div class="category-chips">
                <button 
                  v-for="cat in categories"
                  :key="cat"
                  type="button"
                  :class="['chip', 'chip--small', { 'chip--active': form.category === cat }]"
                  @click="form.category = form.category === cat ? '' : cat"
                >
                  {{ t(`composer.categories.${cat}`) }}
                </button>
              </div>
            </div>
            
            <div class="form-group">
              <label class="form-label">{{ t('composer.info.detailsLabel') }}</label>
              <textarea 
                v-model="form.content"
                class="form-textarea"
                :class="{ 'form-input--error': form.content.length > LIMITS.content }"
                rows="5"
                :maxlength="LIMITS.content"
              ></textarea>
              <div class="form-hint-row">
                <CharCounter :current="form.content.length" :max="LIMITS.content" />
              </div>
            </div>

            <div class="form-group share-toggle">
              <label class="toggle-label">
                <input type="checkbox" v-model="form.isShared" class="toggle-input" />
                <span class="toggle-switch"></span>
                <span class="toggle-text">
                  <span class="toggle-icon">👥</span>
                  {{ t('composer.shareLabel') }}
                </span>
              </label>
            </div>
            
            <div class="modal__actions">
              <button type="button" class="btn btn--ghost" @click="handleClose">
                {{ t('common.cancel') }}
              </button>
              <button type="submit" class="btn btn--primary" :disabled="saving || !form.title">
                {{ saving ? t('common.loading') : t('common.save') }}
              </button>
            </div>
            <div v-if="errorMessage" class="form-error-box">
              <span class="form-error-icon">⚠️</span>
              <span>{{ errorMessage }}</span>
            </div>
          </form>
          
          <!-- Memory Form -->
          <form v-else-if="itemKind === 'memory'" @submit.prevent="handleSave" class="modal__form">
            <div class="form-group">
              <label class="form-label">{{ t('composer.memory.titleLabel') }}</label>
              <input 
                v-model="form.title"
                type="text"
                class="form-input"
                :class="{ 'form-input--error': form.title.length > LIMITS.title }"
                :maxlength="LIMITS.title"
                required
              />
              <div class="form-hint-row">
                <CharCounter :current="form.title.length" :max="LIMITS.title" />
              </div>
            </div>
            
            <div class="form-group">
              <label class="form-label">{{ t('composer.memory.recipientLabel') }}</label>
              <input 
                v-model="form.recipient"
                type="text"
                class="form-input"
                :class="{ 'form-input--error': form.recipient.length > LIMITS.recipient }"
                :maxlength="LIMITS.recipient"
              />
              <div class="form-hint-row">
                <CharCounter :current="form.recipient.length" :max="LIMITS.recipient" />
              </div>
            </div>
            
            <div class="form-group">
              <label class="form-label">{{ t('composer.memory.messageLabel') }}</label>
              <textarea 
                v-model="form.content"
                class="form-textarea"
                :class="{ 'form-input--error': form.content.length > LIMITS.content }"
                rows="6"
                :maxlength="LIMITS.content"
              ></textarea>
              <div class="form-hint-row">
                <CharCounter :current="form.content.length" :max="LIMITS.content" />
              </div>
            </div>

            <div class="form-group share-toggle">
              <label class="toggle-label">
                <input type="checkbox" v-model="form.isShared" class="toggle-input" />
                <span class="toggle-switch"></span>
                <span class="toggle-text">
                  <span class="toggle-icon">👥</span>
                  {{ t('composer.shareLabel') }}
                </span>
              </label>
            </div>
            
            <div class="modal__actions">
              <button type="button" class="btn btn--ghost" @click="handleClose">
                {{ t('common.cancel') }}
              </button>
              <button type="submit" class="btn btn--primary" :disabled="saving || !form.title">
                {{ saving ? t('common.loading') : t('common.save') }}
              </button>
            </div>
            <div v-if="errorMessage" class="form-error-box">
              <span class="form-error-icon">⚠️</span>
              <span>{{ errorMessage }}</span>
            </div>
          </form>
          
          <!-- Guardian Form (view only for now) -->
          <div v-else-if="itemKind === 'guardian'" class="modal__form">
            <div class="form-group">
              <label class="form-label">{{ t('composer.guardian.nameLabel') }}</label>
              <input 
                v-model="form.name"
                type="text"
                class="form-input"
                :maxlength="LIMITS.name"
                disabled
              />
            </div>
            
            <div class="form-group">
              <label class="form-label">{{ t('composer.guardian.emailLabel') }}</label>
              <input 
                v-model="form.email"
                type="email"
                class="form-input"
                :maxlength="LIMITS.email"
                disabled
              />
            </div>
            
            <div class="form-group">
              <label class="form-label">{{ t('composer.guardian.phoneLabel') }}</label>
              <input 
                v-model="form.phone"
                type="tel"
                class="form-input"
                :maxlength="LIMITS.phone"
                disabled
              />
            </div>
            
            <p class="modal__note">
              {{ t('edit.guardianNote') }}
            </p>
            
            <div class="modal__actions">
              <button type="button" class="btn btn--primary" @click="handleClose">
                {{ t('common.close') }}
              </button>
            </div>
          </div>
        </div>
        
        <!-- Error Modal -->
        <ErrorModal 
          :show="showErrorModal"
          :message="errorModalMessage"
          @close="showErrorModal = false"
        />
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-lg);
  z-index: 1000;
}

.modal {
  background: var(--color-card);
  border-radius: var(--radius-xl);
  max-width: 500px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: var(--shadow-lg);
}

.modal--large {
  max-width: 550px;
}

.modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-lg);
  border-bottom: 1px solid var(--color-border-light);
}

.modal__title {
  font-size: var(--font-size-lg);
  margin: 0;
}

.modal__close {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  font-size: 1.25rem;
  cursor: pointer;
  border-radius: var(--radius-md);
  color: var(--color-text-muted);
  transition: all var(--transition-fast);
}

.modal__close:hover {
  background: var(--color-bg-warm);
  color: var(--color-text);
}

.modal__form {
  padding: var(--space-lg);
}

.modal__note {
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
  font-style: italic;
  text-align: center;
  margin: var(--space-lg) 0;
}

.modal__actions {
  display: flex;
  gap: var(--space-md);
  justify-content: flex-end;
  margin-top: var(--space-xl);
  padding-top: var(--space-lg);
  border-top: 1px solid var(--color-border-light);
}

.category-chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
}

/* Transições */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .modal,
.modal-leave-active .modal {
  transition: transform 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal,
.modal-leave-to .modal {
  transform: scale(0.95) translateY(20px);
}

/* Share Toggle */
.share-toggle {
  margin-top: var(--space-md);
}

.toggle-label {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  cursor: pointer;
}

.toggle-input {
  display: none;
}

.toggle-switch {
  width: 44px;
  height: 24px;
  background: var(--color-border);
  border-radius: 12px;
  position: relative;
  transition: background-color var(--transition-fast);
}

.toggle-switch::after {
  content: '';
  position: absolute;
  width: 20px;
  height: 20px;
  background: white;
  border-radius: 50%;
  top: 2px;
  left: 2px;
  transition: transform var(--transition-fast);
}

.toggle-input:checked + .toggle-switch {
  background: var(--color-primary);
}

.toggle-input:checked + .toggle-switch::after {
  transform: translateX(20px);
}

.toggle-text {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

.toggle-icon {
  font-size: 1rem;
}

.form-error-box {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-md);
  background: #fef2f2;
  border: 1px solid var(--color-danger);
  border-radius: var(--radius-md);
  color: var(--color-danger);
  font-size: var(--font-size-sm);
  font-weight: 500;
  margin-top: var(--space-md);
}

.form-error-icon {
  font-size: 1.2em;
}

/* Form hint row */
.form-hint-row {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--space-xs);
}

/* Error input state */
.form-input--error {
  border-color: var(--color-danger);
  background-color: #fef2f2;
}
</style>
