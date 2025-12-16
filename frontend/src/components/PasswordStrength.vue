<!-- =============================================================================
FAMLI - Componente de Validação de Senha
===============================================================================
Este componente exibe feedback visual em tempo real sobre a força da senha,
mostrando quais requisitos foram atendidos e quais ainda faltam.

Requisitos de senha:
- Mínimo 8 caracteres
- Pelo menos uma letra minúscula
- Pelo menos um número

Props:
- password: String - A senha digitada pelo usuário
- show: Boolean - Se deve mostrar o componente

Emits:
- valid: Boolean - Emite true quando a senha atende todos os requisitos
============================================================================= -->

<script setup>
import { computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'

// =============================================================================
// PROPS E EMITS
// =============================================================================

const props = defineProps({
  // Senha digitada pelo usuário
  password: {
    type: String,
    default: ''
  },
  // Se deve mostrar o componente
  show: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['valid'])

const { t } = useI18n()

// =============================================================================
// VALIDAÇÃO DE REQUISITOS
// =============================================================================

// Requisito: mínimo 8 caracteres
const hasMinLength = computed(() => props.password.length >= 8)

// Requisito: pelo menos uma letra minúscula
const hasLowercase = computed(() => /[a-z]/.test(props.password))

// Requisito: pelo menos um número
const hasNumber = computed(() => /[0-9]/.test(props.password))

// Bônus: letra maiúscula (não obrigatório, mas melhora a força)
const hasUppercase = computed(() => /[A-Z]/.test(props.password))

// Bônus: caractere especial (não obrigatório, mas melhora a força)
const hasSpecial = computed(() => /[!@#$%^&*(),.?":{}|<>]/.test(props.password))

// =============================================================================
// FORÇA DA SENHA
// =============================================================================

// Verifica se todos os requisitos obrigatórios foram atendidos
const isValid = computed(() => {
  return hasMinLength.value && hasLowercase.value && hasNumber.value
})

// Calcula a pontuação de força (0-5)
const strengthScore = computed(() => {
  let score = 0
  
  // Requisitos obrigatórios (1 ponto cada)
  if (hasMinLength.value) score++
  if (hasLowercase.value) score++
  if (hasNumber.value) score++
  
  // Bônus
  if (hasUppercase.value) score++
  if (hasSpecial.value) score++
  if (props.password.length >= 12) score++
  
  return Math.min(score, 5)
})

// Nível de força baseado na pontuação
const strengthLevel = computed(() => {
  if (props.password.length === 0) return 'empty'
  if (strengthScore.value <= 2) return 'weak'
  if (strengthScore.value <= 3) return 'fair'
  if (strengthScore.value <= 4) return 'good'
  return 'strong'
})

// Label de força
const strengthLabel = computed(() => {
  switch (strengthLevel.value) {
    case 'empty': return ''
    case 'weak': return t('password.weak')
    case 'fair': return t('password.fair')
    case 'good': return t('password.good')
    case 'strong': return t('password.strong')
    default: return ''
  }
})

// =============================================================================
// LISTA DE REQUISITOS
// =============================================================================

const requirements = computed(() => [
  {
    key: 'minLength',
    label: t('password.req.minLength'),
    met: hasMinLength.value,
    required: true
  },
  {
    key: 'lowercase',
    label: t('password.req.lowercase'),
    met: hasLowercase.value,
    required: true
  },
  {
    key: 'number',
    label: t('password.req.number'),
    met: hasNumber.value,
    required: true
  },
  {
    key: 'uppercase',
    label: t('password.req.uppercase'),
    met: hasUppercase.value,
    required: false
  },
  {
    key: 'special',
    label: t('password.req.special'),
    met: hasSpecial.value,
    required: false
  }
])

// =============================================================================
// EMITIR VALIDADE
// =============================================================================

// Observar mudanças na validação e emitir
watch(isValid, (newValue) => {
  emit('valid', newValue)
}, { immediate: true })
</script>

<template>
  <div v-if="show && password.length > 0" class="password-strength">
    <!-- Barra de força -->
    <div class="strength-bar">
      <div 
        class="strength-bar__fill"
        :class="`strength-bar__fill--${strengthLevel}`"
        :style="{ width: `${(strengthScore / 5) * 100}%` }"
      />
    </div>
    
    <!-- Label de força -->
    <div class="strength-label" :class="`strength-label--${strengthLevel}`">
      {{ strengthLabel }}
    </div>

    <!-- Lista de requisitos -->
    <ul class="requirements-list">
      <li 
        v-for="req in requirements" 
        :key="req.key"
        class="requirement"
        :class="{ 
          'requirement--met': req.met,
          'requirement--optional': !req.required
        }"
      >
        <span class="requirement__icon">
          {{ req.met ? '✓' : '○' }}
        </span>
        <span class="requirement__label">
          {{ req.label }}
          <span v-if="!req.required" class="requirement__optional">
            ({{ t('password.optional') }})
          </span>
        </span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
/* =============================================================================
   CONTAINER
   ============================================================================= */

.password-strength {
  margin-top: var(--space-sm);
  padding: var(--space-md);
  background: var(--color-bg-warm);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border-light);
}

/* =============================================================================
   BARRA DE FORÇA
   ============================================================================= */

.strength-bar {
  height: 6px;
  background: var(--color-border);
  border-radius: var(--radius-full);
  overflow: hidden;
  margin-bottom: var(--space-sm);
}

.strength-bar__fill {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width 0.3s ease, background-color 0.3s ease;
}

/* Cores por nível de força */
.strength-bar__fill--empty {
  background: var(--color-border);
}

.strength-bar__fill--weak {
  background: #ef4444; /* Vermelho */
}

.strength-bar__fill--fair {
  background: #f59e0b; /* Laranja */
}

.strength-bar__fill--good {
  background: #10b981; /* Verde */
}

.strength-bar__fill--strong {
  background: #059669; /* Verde escuro */
}

/* =============================================================================
   LABEL DE FORÇA
   ============================================================================= */

.strength-label {
  font-size: var(--font-size-sm);
  font-weight: 600;
  margin-bottom: var(--space-sm);
  text-align: right;
}

.strength-label--empty {
  color: var(--color-text-soft);
}

.strength-label--weak {
  color: #ef4444;
}

.strength-label--fair {
  color: #f59e0b;
}

.strength-label--good {
  color: #10b981;
}

.strength-label--strong {
  color: #059669;
}

/* =============================================================================
   LISTA DE REQUISITOS
   ============================================================================= */

.requirements-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.requirement {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  transition: color 0.2s ease;
}

.requirement--met {
  color: #10b981;
}

.requirement--met .requirement__icon {
  color: #10b981;
}

.requirement--optional {
  opacity: 0.8;
}

.requirement__icon {
  font-size: 0.875rem;
  width: 1.25rem;
  text-align: center;
  transition: color 0.2s ease;
}

.requirement__label {
  flex: 1;
}

.requirement__optional {
  font-size: var(--font-size-xs);
  color: var(--color-text-soft);
  opacity: 0.7;
}
</style>

