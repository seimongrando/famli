<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  card: {
    type: Object,
    required: true
  },
  status: {
    type: String,
    default: 'pending'
  }
})

const emit = defineEmits(['start', 'complete', 'skip'])

const isCompleted = computed(() => props.status === 'completed')
const isSkipped = computed(() => props.status === 'skipped')
const isStarted = computed(() => props.status === 'started')

// Mapear IDs dos cards para as chaves de tradução
const cardTranslationKeys = {
  'welcome': 'guide.cards.welcome',
  'people': 'guide.cards.people',
  'locations': 'guide.cards.locations',
  'routines': 'guide.cards.routines',
  'access': 'guide.cards.access',
  'memories': 'guide.cards.memories'
}

const cardTitle = computed(() => {
  const key = cardTranslationKeys[props.card.id]
  return key ? t(`${key}.title`) : props.card.title
})

const cardDescription = computed(() => {
  const key = cardTranslationKeys[props.card.id]
  return key ? t(`${key}.description`) : props.card.description
})
</script>

<template>
  <div :class="['guide-card', { 
    'guide-card--completed': isCompleted,
    'guide-card--skipped': isSkipped 
  }]">
    <div class="guide-card__icon">
      <span v-if="isCompleted">✅</span>
      <span v-else-if="isSkipped">⏭️</span>
      <span v-else>{{ card.icon }}</span>
    </div>
    
    <div class="guide-card__content">
      <div class="guide-card__header">
        <h3 class="guide-card__title">{{ cardTitle }}</h3>
        <span v-if="isCompleted" class="guide-card__badge guide-card__badge--success">
          {{ t('guide.status.completed') }}
        </span>
        <span v-else-if="isSkipped" class="guide-card__badge guide-card__badge--muted">
          {{ t('guide.status.skipped') }}
        </span>
        <span v-else-if="isStarted" class="guide-card__badge guide-card__badge--active">
          {{ t('guide.status.started') }}
        </span>
      </div>
      
      <p class="guide-card__description">{{ cardDescription }}</p>
      
      <div v-if="!isCompleted && !isSkipped" class="guide-card__actions">
        <button 
          class="btn btn--primary btn--small"
          @click="emit('start')"
        >
          {{ isStarted ? t('guide.actions.continue') : t('guide.actions.start') }}
        </button>
        <button 
          v-if="!isStarted"
          class="btn btn--ghost btn--small"
          @click="emit('skip')"
        >
          {{ t('guide.actions.skip') }}
        </button>
        <button 
          v-if="isStarted"
          class="btn btn--ghost btn--small"
          @click="emit('complete')"
        >
          {{ t('guide.actions.markDone') }}
        </button>
      </div>
      
      <div v-else class="guide-card__done">
        <button 
          class="btn btn--link btn--small"
          @click="emit('start')"
        >
          {{ t('guide.actions.review') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.guide-card {
  display: flex;
  gap: var(--space-lg);
  padding: var(--space-lg);
  background: var(--color-card);
  border-radius: var(--radius-lg);
  border: 2px solid var(--color-border-light);
  transition: all var(--transition-normal);
}

.guide-card:hover {
  border-color: var(--color-border);
  box-shadow: var(--shadow-sm);
}

.guide-card--completed {
  background: var(--color-primary-soft);
  border-color: var(--color-primary);
}

.guide-card--skipped {
  opacity: 0.7;
}

.guide-card__icon {
  font-size: 2rem;
  line-height: 1;
}

.guide-card__content {
  flex: 1;
}

.guide-card__header {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  margin-bottom: var(--space-sm);
}

.guide-card__title {
  font-size: var(--font-size-lg);
  margin: 0;
}

.guide-card__badge {
  padding: 2px 10px;
  border-radius: var(--radius-full);
  font-size: 0.75rem;
  font-weight: 600;
}

.guide-card__badge--success {
  background: var(--color-success);
  color: white;
}

.guide-card__badge--active {
  background: var(--color-accent);
  color: white;
}

.guide-card__badge--muted {
  background: var(--color-border);
  color: var(--color-text-muted);
}

.guide-card__description {
  color: var(--color-text-soft);
  margin-bottom: var(--space-md);
}

.guide-card__actions {
  display: flex;
  gap: var(--space-sm);
}

.guide-card__done {
  margin-top: var(--space-sm);
}
</style>
