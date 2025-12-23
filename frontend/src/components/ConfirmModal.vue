<!-- =============================================================================
  FAMLI - Modal de Confirmação
  =============================================================================
  Modal moderno para confirmações de ações destrutivas ou importantes.
  
  Props:
  - show: boolean - Controla visibilidade do modal
  - title: string - Título do modal
  - message: string - Mensagem de confirmação
  - confirmText: string - Texto do botão de confirmação
  - cancelText: string - Texto do botão de cancelar
  - type: 'danger' | 'warning' | 'info' - Estilo do modal
  
  Events:
  - confirm: Emitido quando usuário confirma
  - cancel: Emitido quando usuário cancela
============================================================================== -->

<script setup>
import { watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: ''
  },
  message: {
    type: String,
    default: ''
  },
  details: {
    type: Array,
    default: () => []
  },
  confirmText: {
    type: String,
    default: ''
  },
  cancelText: {
    type: String,
    default: ''
  },
  showCancel: {
    type: Boolean,
    default: true
  },
  type: {
    type: String,
    default: 'danger',
    validator: (value) => ['danger', 'warning', 'info'].includes(value)
  }
})

const emit = defineEmits(['confirm', 'cancel'])

// Prevenir scroll do body quando modal está aberto
watch(() => props.show, (show) => {
  if (show) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})

function handleConfirm() {
  emit('confirm')
}

function handleCancel() {
  emit('cancel')
}

function handleBackdropClick(e) {
  if (e.target === e.currentTarget) {
    handleCancel()
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="modal-overlay" @click="handleBackdropClick">
        <div class="modal" role="dialog" aria-modal="true">
          <div class="modal__icon" :class="`modal__icon--${type}`">
            <span v-if="type === 'danger'">🗑️</span>
            <span v-else-if="type === 'warning'">⚠️</span>
            <span v-else>ℹ️</span>
          </div>
          
          <h2 class="modal__title">{{ title || t('modal.confirm') }}</h2>
          
          <p class="modal__message">{{ message }}</p>
          <ul v-if="details && details.length" class="modal__details">
            <li v-for="detail in details" :key="detail">{{ detail }}</li>
          </ul>
          
          <div class="modal__actions">
            <button 
              v-if="showCancel"
              class="btn btn--ghost" 
              @click="handleCancel"
            >
              {{ cancelText || t('common.cancel') }}
            </button>
            <button 
              :class="['btn', type === 'danger' ? 'btn--danger' : 'btn--primary']"
              @click="handleConfirm"
            >
              {{ confirmText || t('common.confirm') }}
            </button>
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
  padding: var(--space-xl);
  max-width: 400px;
  width: 100%;
  text-align: center;
  box-shadow: var(--shadow-lg);
}

.modal__icon {
  width: 64px;
  height: 64px;
  margin: 0 auto var(--space-lg);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
}

.modal__icon--danger {
  background: #fee2e2;
}

.modal__icon--warning {
  background: #fef3c7;
}

.modal__icon--info {
  background: var(--color-primary-soft);
}

.modal__title {
  font-size: var(--font-size-lg);
  margin: 0 0 var(--space-sm);
}

.modal__message {
  color: var(--color-text-soft);
  margin: 0 0 var(--space-xl);
  line-height: 1.6;
}

.modal__details {
  margin: -12px 0 var(--space-xl);
  padding-left: 1.2rem;
  text-align: left;
  color: var(--color-text-soft);
  line-height: 1.6;
}

.modal__details li {
  margin-bottom: var(--space-xs);
}

.modal__actions {
  display: flex;
  gap: var(--space-md);
  justify-content: center;
}

.modal__actions .btn {
  flex: 1;
  max-width: 150px;
}

.btn--danger {
  background: #dc2626;
  color: white;
}

.btn--danger:hover {
  background: #b91c1c;
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
  transform: scale(0.95);
}
</style>

