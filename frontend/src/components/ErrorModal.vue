<!-- =============================================================================
  FAMLI - ErrorModal Component
  =============================================================================
  Modal amigável para exibir erros do backend ao usuário.
  
  Props:
  - show: boolean - Controla visibilidade do modal
  - title: string - Título do erro (opcional, usa padrão traduzido)
  - message: string - Mensagem de erro a ser exibida
  - type: 'error' | 'warning' | 'info' - Tipo do modal (padrão: error)
  
  Events:
  - close: Emitido quando o modal é fechado
  - retry: Emitido quando o usuário clica em tentar novamente
============================================================================== -->

<script setup>
import { computed, watch } from 'vue'
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
    required: true
  },
  type: {
    type: String,
    default: 'error',
    validator: (value) => ['error', 'warning', 'info'].includes(value)
  },
  showRetry: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close', 'retry'])

const displayTitle = computed(() => {
  if (props.title) return props.title
  return t(`errorModal.title.${props.type}`)
})

const icon = computed(() => {
  const icons = {
    error: '❌',
    warning: '⚠️',
    info: 'ℹ️'
  }
  return icons[props.type]
})

// Prevenir scroll do body quando modal está aberto
watch(() => props.show, (show) => {
  if (show) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})

function handleClose() {
  emit('close')
}

function handleRetry() {
  emit('retry')
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
      <div v-if="show" class="error-modal-overlay" @click="handleBackdropClick">
        <div :class="['error-modal', `error-modal--${type}`]" role="alertdialog" aria-modal="true">
          <div class="error-modal__icon">
            {{ icon }}
          </div>
          
          <h2 class="error-modal__title">{{ displayTitle }}</h2>
          
          <p class="error-modal__message">{{ message }}</p>
          
          <div class="error-modal__actions">
            <button 
              v-if="showRetry" 
              class="btn btn--secondary" 
              @click="handleRetry"
            >
              {{ t('errorModal.retry') }}
            </button>
            <button class="btn btn--primary" @click="handleClose">
              {{ t('errorModal.understood') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.error-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-lg);
  z-index: 1100;
}

.error-modal {
  background: var(--color-card, #ffffff);
  border-radius: var(--radius-xl, 16px);
  max-width: 400px;
  width: 100%;
  padding: var(--space-xl, 32px);
  text-align: center;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

.error-modal--error {
  border-top: 4px solid #dc2626;
}

.error-modal--warning {
  border-top: 4px solid #d97706;
}

.error-modal--info {
  border-top: 4px solid #2563eb;
}

.error-modal__icon {
  font-size: 3rem;
  margin-bottom: var(--space-md, 16px);
}

.error-modal__title {
  font-size: var(--font-size-lg, 1.25rem);
  font-weight: 600;
  margin: 0 0 var(--space-md, 16px);
  color: var(--color-text, #1f2937);
}

.error-modal__message {
  font-size: var(--font-size-base, 1rem);
  color: var(--color-text-soft, #6b7280);
  margin: 0 0 var(--space-xl, 32px);
  line-height: 1.6;
}

.error-modal__actions {
  display: flex;
  gap: var(--space-md, 16px);
  justify-content: center;
}

.btn {
  padding: var(--space-sm, 12px) var(--space-lg, 24px);
  border-radius: var(--radius-md, 8px);
  font-weight: 600;
  font-size: var(--font-size-base, 1rem);
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}

.btn--primary {
  background: var(--color-primary, #8b5cf6);
  color: white;
}

.btn--primary:hover {
  background: var(--color-primary-dark, #7c3aed);
}

.btn--secondary {
  background: var(--color-bg-warm, #f3f4f6);
  color: var(--color-text, #1f2937);
}

.btn--secondary:hover {
  background: var(--color-border, #e5e7eb);
}

/* Transições */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .error-modal,
.modal-leave-active .error-modal {
  transition: transform 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .error-modal,
.modal-leave-to .error-modal {
  transform: scale(0.95) translateY(10px);
}

@media (max-width: 480px) {
  .error-modal__actions {
    flex-direction: column;
  }
  
  .btn {
    width: 100%;
  }
}
</style>

