<!-- =============================================================================
  FAMLI - CharCounter Component
  =============================================================================
  Componente para exibir contador de caracteres com feedback visual.
  
  Props:
  - current: number - Quantidade atual de caracteres
  - max: number - Limite máximo de caracteres
  - showRemaining: boolean - Se deve mostrar caracteres restantes (padrão: true)
  
  Estilos:
  - Verde: < 70% do limite
  - Amarelo: 70-90% do limite
  - Vermelho: > 90% do limite
============================================================================== -->

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  current: {
    type: Number,
    default: 0
  },
  max: {
    type: Number,
    required: true
  },
  showRemaining: {
    type: Boolean,
    default: true
  }
})

const remaining = computed(() => props.max - props.current)
const percentage = computed(() => (props.current / props.max) * 100)

const statusClass = computed(() => {
  if (percentage.value >= 90) return 'char-counter--danger'
  if (percentage.value >= 70) return 'char-counter--warning'
  return 'char-counter--normal'
})

const displayText = computed(() => {
  if (props.showRemaining) {
    return t('charCounter.remaining', { remaining: remaining.value, max: props.max })
  }
  return `${props.current}/${props.max}`
})
</script>

<template>
  <span :class="['char-counter', statusClass]">
    {{ displayText }}
  </span>
</template>

<style scoped>
.char-counter {
  font-size: var(--font-size-xs, 0.75rem);
  font-variant-numeric: tabular-nums;
  transition: color 0.2s ease;
}

.char-counter--normal {
  color: var(--color-text-muted, #6b7280);
}

.char-counter--warning {
  color: var(--color-warning, #d97706);
}

.char-counter--danger {
  color: var(--color-danger, #dc2626);
  font-weight: 500;
}
</style>

