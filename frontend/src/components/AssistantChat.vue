<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()

const input = ref('')
const reply = ref('')
const loading = ref(false)

async function sendMessage() {
  if (!input.value.trim() || loading.value) return
  
  loading.value = true
  reply.value = ''
  
  try {
    const res = await fetch('/api/assistant', {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Accept-Language': locale.value
      },
      credentials: 'include',
      body: JSON.stringify({ input: input.value })
    })
    
    if (res.ok) {
      const data = await res.json()
      reply.value = data.reply
    } else {
      reply.value = t('errors.generic')
    }
  } catch (e) {
    reply.value = t('errors.connection')
  } finally {
    loading.value = false
    input.value = ''
  }
}
</script>

<template>
  <div class="assistant">
    <div class="assistant__header">
      <span class="assistant__icon">🤖</span>
      <h3 class="assistant__title">{{ t('assistant.title') }}</h3>
    </div>
    
    <p class="assistant__intro">
      {{ t('assistant.intro') }}
    </p>
    
    <form @submit.prevent="sendMessage" class="assistant__form">
      <textarea
        v-model="input"
        class="form-textarea assistant__input"
        :placeholder="t('assistant.placeholder')"
        rows="3"
        :disabled="loading"
      ></textarea>
      
      <button 
        type="submit" 
        class="btn btn--primary btn--full"
        :disabled="loading || !input.trim()"
      >
        {{ loading ? t('assistant.thinking') : t('assistant.send') }}
      </button>
    </form>
    
    <div v-if="reply" class="assistant__reply">
      <span class="assistant__reply-icon">💬</span>
      <p class="assistant__reply-text">{{ reply }}</p>
    </div>
  </div>
</template>

<style scoped>
.assistant {
  background: var(--color-accent-soft);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  border: 1px solid rgba(224, 123, 57, 0.2);
}

.assistant__header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  margin-bottom: var(--space-sm);
}

.assistant__icon {
  font-size: 1.25rem;
}

.assistant__title {
  font-size: var(--font-size-base);
  margin: 0;
}

.assistant__intro {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  margin-bottom: var(--space-md);
}

.assistant__form {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.assistant__input {
  min-height: 80px;
  resize: vertical;
}

.assistant__reply {
  display: flex;
  gap: var(--space-sm);
  margin-top: var(--space-md);
  padding: var(--space-md);
  background: var(--color-card);
  border-radius: var(--radius-md);
}

.assistant__reply-icon {
  font-size: 1.25rem;
  flex-shrink: 0;
}

.assistant__reply-text {
  font-size: var(--font-size-sm);
  color: var(--color-text);
  margin: 0;
  line-height: 1.6;
}
</style>
