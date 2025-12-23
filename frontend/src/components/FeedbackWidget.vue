<!--
=============================================================================
FAMLI - FeedbackWidget
=============================================================================
Componente flutuante para envio de feedback

Funcionalidades:
- Botão flutuante no canto inferior direito
- Modal simples para envio de feedback
- Tipos: sugestão, problema, elogio, dúvida
- Confirmação visual após envio

Uso:
<FeedbackWidget />

Importar em App.vue ou DashboardPage.vue
=============================================================================
-->

<template>
  <!-- Botão Flutuante -->
  <button 
    v-if="!isOpen" 
    class="feedback-fab"
    @click="openModal"
    :title="$t('feedback.title')"
  >
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
    </svg>
  </button>

  <!-- Modal de Feedback -->
  <Teleport to="body">
    <div v-if="isOpen" class="feedback-overlay" @click.self="closeModal">
      <div class="feedback-modal">
        <!-- Header -->
        <div class="feedback-header">
          <h3>{{ $t('feedback.title') }}</h3>
          <button class="feedback-close" @click="closeModal">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>

        <!-- Success State -->
        <div v-if="submitted" class="feedback-success">
          <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
            <polyline points="22 4 12 14.01 9 11.01"/>
          </svg>
          <p>{{ $t('feedback.successMessage') }}</p>
          <button class="btn btn-secondary" @click="closeModal">
            {{ $t('common.close') }}
          </button>
        </div>

        <!-- Form -->
        <form v-else @submit.prevent="submitFeedback" class="feedback-form">
          <!-- Tipo de feedback -->
          <div class="form-group">
            <label class="form-label">{{ $t('feedback.typeLabel') }}</label>
            <div class="feedback-types">
              <button 
                type="button"
                v-for="feedbackType in feedbackTypes" 
                :key="feedbackType.value"
                :class="['feedback-type-btn', { active: type === feedbackType.value }]"
                @click="type = feedbackType.value"
              >
                <span class="feedback-type-icon">{{ feedbackType.icon }}</span>
                <span class="feedback-type-label">{{ $t(feedbackType.label) }}</span>
              </button>
            </div>
          </div>

          <!-- Mensagem -->
          <div class="form-group">
            <label class="form-label" for="feedback-message">
              {{ $t('feedback.messageLabel') }}
            </label>
            <textarea
              id="feedback-message"
              v-model="message"
              :placeholder="$t('feedback.messagePlaceholder')"
              class="form-textarea"
              rows="4"
              maxlength="2000"
              required
            ></textarea>
            <span class="char-count">{{ message.length }}/2000</span>
          </div>

          <!-- Error -->
          <div v-if="error" class="feedback-error">
            {{ error }}
          </div>

          <!-- Submit -->
          <div class="feedback-actions">
            <button type="button" class="btn btn-secondary" @click="closeModal">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" class="btn btn-primary" :disabled="loading || !message.trim()">
              <span v-if="loading" class="loading-spinner"></span>
              {{ loading ? $t('common.sending') : $t('feedback.submit') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

const { t } = useI18n()
const route = useRoute()

// Estado do modal
const isOpen = ref(false)
const submitted = ref(false)
const loading = ref(false)
const error = ref('')

// Dados do formulário
const type = ref('suggestion')
const message = ref('')

// Tipos de feedback
const feedbackTypes = [
  { value: 'suggestion', label: 'feedback.types.suggestion', icon: '💡' },
  { value: 'problem', label: 'feedback.types.problem', icon: '🐛' },
  { value: 'praise', label: 'feedback.types.praise', icon: '❤️' },
  { value: 'question', label: 'feedback.types.question', icon: '❓' }
]

// Métodos
function openModal() {
  isOpen.value = true
  submitted.value = false
  error.value = ''
}

function closeModal() {
  isOpen.value = false
  // Reset form após um delay para animação
  setTimeout(() => {
    type.value = 'suggestion'
    message.value = ''
    submitted.value = false
    error.value = ''
  }, 300)
}

async function submitFeedback() {
  if (!message.value.trim()) return

  loading.value = true
  error.value = ''

  try {
    const response = await fetch('/api/feedback', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({
        type: type.value,
        message: message.value.trim(),
        page: route.path
      })
    })

    if (!response.ok) {
      const data = await response.json()
      throw new Error(data.error || 'Erro ao enviar feedback')
    }

    // Tracking do evento
    try {
      await fetch('/api/analytics/track', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          event_type: 'send_feedback',
          page: route.path,
          details: { type: type.value }
        })
      })
    } catch (e) { /* ignore tracking errors */ }

    submitted.value = true
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/* Botão Flutuante */
.feedback-fab {
  position: fixed;
  bottom: 24px;
  right: 24px;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--color-primary);
  color: white;
  border: none;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  z-index: 1000;
}

.feedback-fab:hover {
  transform: scale(1.05);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
}

.feedback-fab:active {
  transform: scale(0.95);
}

/* Overlay */
.feedback-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1001;
  padding: 16px;
}

/* Modal */
.feedback-modal {
  background: white;
  border-radius: 16px;
  width: 100%;
  max-width: 480px;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.2);
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Header */
.feedback-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid var(--color-border-light);
}

.feedback-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: var(--color-text);
}

.feedback-close {
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  color: var(--color-text-muted);
  transition: color 0.2s;
}

.feedback-close:hover {
  color: var(--color-text);
}

/* Form */
.feedback-form {
  padding: 24px;
}

.form-group {
  margin-bottom: 20px;
}

.form-label {
  display: block;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--color-text);
}

/* Tipos de Feedback */
.feedback-types {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}

.feedback-type-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 8px;
  background: var(--color-bg-light);
  border: 2px solid transparent;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.feedback-type-btn:hover {
  background: var(--color-bg);
}

.feedback-type-btn.active {
  border-color: var(--color-primary);
  background: var(--color-primary-light);
}

.feedback-type-icon {
  font-size: 1.5rem;
  margin-bottom: 4px;
}

.feedback-type-label {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  text-align: center;
}

.feedback-type-btn.active .feedback-type-label {
  color: var(--color-primary);
  font-weight: 600;
}

/* Textarea */
.form-textarea {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  font-size: 1rem;
  font-family: inherit;
  resize: vertical;
  min-height: 100px;
  transition: border-color 0.2s;
}

.form-textarea:focus {
  outline: none;
  border-color: var(--color-primary);
}

.char-count {
  display: block;
  text-align: right;
  font-size: 0.75rem;
  color: var(--color-text-muted);
  margin-top: 4px;
}

/* Error */
.feedback-error {
  padding: 12px 16px;
  background: #fef2f2;
  color: #dc2626;
  border-radius: 8px;
  margin-bottom: 16px;
  font-size: 0.875rem;
}

/* Actions */
.feedback-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.btn {
  padding: 12px 24px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.9375rem;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--color-primary-dark);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--color-bg-light);
  color: var(--color-text);
}

.btn-secondary:hover {
  background: var(--color-bg);
}

/* Success */
.feedback-success {
  padding: 48px 24px;
  text-align: center;
}

.feedback-success svg {
  color: #22c55e;
  margin-bottom: 16px;
}

.feedback-success p {
  font-size: 1rem;
  color: var(--color-text);
  margin-bottom: 24px;
}

/* Loading */
.loading-spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 1s linear infinite;
  margin-right: 8px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Responsivo */
@media (max-width: 480px) {
  .feedback-types {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .feedback-fab {
    bottom: 16px;
    right: 16px;
  }
}
</style>

