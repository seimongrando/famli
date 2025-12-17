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

const { t } = useI18n()
const boxStore = useBoxStore()

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
const form = ref({
  title: '',
  content: '',
  category: '',
  name: '',
  email: '',
  phone: '',
  relationship: '',
  recipient: ''
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
      recipient: newItem.recipient || ''
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
  
  saving.value = true
  
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
        recipient: form.value.recipient
      })
      
      if (result) {
        emit('save', result)
        emit('close')
      }
    }
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
                required
              />
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
                rows="5"
              ></textarea>
            </div>
            
            <div class="modal__actions">
              <button type="button" class="btn btn--ghost" @click="handleClose">
                {{ t('common.cancel') }}
              </button>
              <button type="submit" class="btn btn--primary" :disabled="saving || !form.title">
                {{ saving ? t('common.loading') : t('common.save') }}
              </button>
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
                required
              />
            </div>
            
            <div class="form-group">
              <label class="form-label">{{ t('composer.memory.recipientLabel') }}</label>
              <input 
                v-model="form.recipient"
                type="text"
                class="form-input"
              />
            </div>
            
            <div class="form-group">
              <label class="form-label">{{ t('composer.memory.messageLabel') }}</label>
              <textarea 
                v-model="form.content"
                class="form-textarea"
                rows="6"
              ></textarea>
            </div>
            
            <div class="modal__actions">
              <button type="button" class="btn btn--ghost" @click="handleClose">
                {{ t('common.cancel') }}
              </button>
              <button type="submit" class="btn btn--primary" :disabled="saving || !form.title">
                {{ saving ? t('common.loading') : t('common.save') }}
              </button>
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
                disabled
              />
            </div>
            
            <div class="form-group">
              <label class="form-label">{{ t('composer.guardian.emailLabel') }}</label>
              <input 
                v-model="form.email"
                type="email"
                class="form-input"
                disabled
              />
            </div>
            
            <div class="form-group">
              <label class="form-label">{{ t('composer.guardian.phoneLabel') }}</label>
              <input 
                v-model="form.phone"
                type="tel"
                class="form-input"
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
</style>

