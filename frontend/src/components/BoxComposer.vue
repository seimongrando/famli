<script setup>
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useBoxStore } from '../stores/box'

const { t } = useI18n()
const boxStore = useBoxStore()
const formError = ref('')

// Props para permitir selecionar tipo externamente
const props = defineProps({
  initialType: {
    type: String,
    default: 'info'
  }
})

const emit = defineEmits(['saved'])

const activeType = ref(props.initialType)
const saving = ref(false)

// Observar mudanças na prop para atualizar o tipo ativo
watch(() => props.initialType, (newType) => {
  if (newType) {
    activeType.value = newType
  }
})

// Forms para cada tipo
const infoForm = ref({ title: '', content: '', category: '', isShared: false })
const guardianForm = ref({ name: '', email: '', phone: '', relationship: '', accessPin: '' })
const memoryForm = ref({ title: '', content: '', recipient: '', isShared: false })

const types = [
  { id: 'info', labelKey: 'composer.types.info', icon: '📋' },
  { id: 'guardian', labelKey: 'composer.types.guardian', icon: '👤' },
  { id: 'memory', labelKey: 'composer.types.memory', icon: '💝' }
]

const categories = ['saude', 'financas', 'documentos', 'casa', 'familia', 'outro']
const relationships = ['filho', 'neto', 'conjuge', 'irmao', 'amigo', 'outro']

async function saveInfo() {
  if (!infoForm.value.title) {
    console.log('[Composer] Info form validation failed: title required')
    return
  }
  
  saving.value = true
  formError.value = ''
  console.log('[Composer] Saving info:', infoForm.value.title)
  
  try {
    const result = await boxStore.createItem({
      type: 'info',
      title: infoForm.value.title,
      content: infoForm.value.content,
      category: infoForm.value.category,
      is_shared: infoForm.value.isShared
    })
    
    if (result) {
      console.log('[Composer] Info saved successfully:', result.id)
      infoForm.value = { title: '', content: '', category: '', isShared: false }
      emit('saved', 'info')
    } else {
      console.error('[Composer] Info save failed - no result')
      formError.value = boxStore.error || t('errors.generic')
    }
  } catch (error) {
    console.error('[Composer] Info save error:', error)
    formError.value = boxStore.error || t('errors.generic')
  } finally {
    saving.value = false
  }
}

async function saveGuardian() {
  if (!guardianForm.value.name) {
    console.log('[Composer] Guardian form validation failed: name required')
    formError.value = t('errors.requiredField')
    return
  }
  if (!guardianForm.value.accessPin) {
    console.log('[Composer] Guardian form validation failed: pin required')
    formError.value = t('guardian.pinRequired')
    return
  }
  
  saving.value = true
  formError.value = ''
  console.log('[Composer] Saving guardian:', guardianForm.value.name)
  
  try {
    const payload = {
      name: guardianForm.value.name,
      email: guardianForm.value.email,
      phone: guardianForm.value.phone,
      relationship: guardianForm.value.relationship
    }
    
    payload.access_pin = guardianForm.value.accessPin
    
    const result = await boxStore.createGuardian(payload)
    
    if (result) {
      console.log('[Composer] Guardian saved successfully:', result.id)
      guardianForm.value = { name: '', email: '', phone: '', relationship: '', accessPin: '' }
      emit('saved', 'guardian')
    } else {
      console.error('[Composer] Guardian save failed - no result')
      formError.value = boxStore.error || t('errors.generic')
    }
  } catch (error) {
    console.error('[Composer] Guardian save error:', error)
    formError.value = boxStore.error || t('errors.generic')
  } finally {
    saving.value = false
  }
}

async function saveMemory() {
  if (!memoryForm.value.title) {
    console.log('[Composer] Memory form validation failed: title required')
    return
  }
  
  saving.value = true
  formError.value = ''
  console.log('[Composer] Saving memory:', memoryForm.value.title)
  
  try {
    const result = await boxStore.createItem({
      type: 'memory',
      title: memoryForm.value.title,
      content: memoryForm.value.content,
      recipient: memoryForm.value.recipient,
      is_shared: memoryForm.value.isShared
    })
    
    if (result) {
      console.log('[Composer] Memory saved successfully:', result.id)
      memoryForm.value = { title: '', content: '', recipient: '', isShared: false }
      emit('saved', 'memory')
    } else {
      console.error('[Composer] Memory save failed - no result')
      formError.value = boxStore.error || t('errors.generic')
    }
  } catch (error) {
    console.error('[Composer] Memory save error:', error)
    formError.value = boxStore.error || t('errors.generic')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="composer">
    <div class="composer__header">
      <h2 class="composer__title">{{ t('composer.title') }}</h2>
      <p class="composer__subtitle">
        {{ t('composer.subtitle') }}
      </p>
    </div>

    <!-- Type Selector -->
    <div class="composer__types">
      <button 
        v-for="type in types" 
        :key="type.id"
        :class="['chip', { 'chip--active': activeType === type.id }]"
        @click="activeType = type.id"
      >
        <span>{{ type.icon }}</span>
        {{ t(type.labelKey) }}
      </button>
    </div>

    <!-- Info Form -->
    <form v-if="activeType === 'info'" @submit.prevent="saveInfo" class="composer__form">
      <p class="composer__hint">
        💡 {{ t('composer.info.hint') }}
      </p>
      
      <div class="form-group">
        <label class="form-label">{{ t('composer.info.titleLabel') }}</label>
        <input 
          v-model="infoForm.title"
          type="text"
          class="form-input"
          :placeholder="t('composer.info.titlePlaceholder')"
          maxlength="255"
          required
        />
        <small class="form-hint">{{ t('common.maxChars', { count: 255 }) }}</small>
      </div>
      
      <div class="form-group">
        <label class="form-label">{{ t('composer.info.categoryLabel') }}</label>
        <div class="category-chips">
          <button 
            v-for="cat in categories"
            :key="cat"
            type="button"
            :class="['chip', 'chip--small', { 'chip--active': infoForm.category === cat }]"
            @click="infoForm.category = infoForm.category === cat ? '' : cat"
          >
            {{ t(`composer.categories.${cat}`) }}
          </button>
        </div>
      </div>
      
      <div class="form-group">
        <label class="form-label">{{ t('composer.info.detailsLabel') }}</label>
        <textarea 
          v-model="infoForm.content"
          class="form-textarea"
          :placeholder="t('composer.info.detailsPlaceholder')"
          rows="4"
          maxlength="10000"
        ></textarea>
      </div>

      <div class="form-group share-toggle">
        <label class="toggle-label">
          <input type="checkbox" v-model="infoForm.isShared" class="toggle-input" />
          <span class="toggle-switch"></span>
          <span class="toggle-text">
            <span class="toggle-icon">👥</span>
            {{ t('composer.shareWithGuardians') }}
          </span>
        </label>
        <small class="toggle-hint">{{ t('composer.shareHint') }}</small>
      </div>
      
      <button type="submit" class="btn btn--primary" :disabled="saving || !infoForm.title">
        {{ saving ? t('composer.info.saving') : t('composer.info.saveButton') }}
      </button>
      
      <div v-if="formError" class="form-error-box">
        <span class="form-error-icon">⚠️</span>
        <span>{{ formError }}</span>
      </div>
    </form>

    <!-- Guardian Form -->
    <form v-else-if="activeType === 'guardian'" @submit.prevent="saveGuardian" class="composer__form">
      <p class="composer__hint">
        👥 {{ t('composer.guardian.hint', { bold: '' }) }}
        <strong>{{ t('composer.guardian.hintBold') }}</strong>
      </p>
      
      <div class="form-group">
        <label class="form-label">{{ t('composer.guardian.nameLabel') }}</label>
        <input 
          v-model="guardianForm.name"
          type="text"
          class="form-input"
          :placeholder="t('composer.guardian.namePlaceholder')"
          maxlength="255"
          required
        />
        <small class="form-hint">{{ t('common.maxChars', { count: 255 }) }}</small>
      </div>
      
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">{{ t('composer.guardian.emailLabel') }}</label>
          <input 
            v-model="guardianForm.email"
            type="email"
            class="form-input"
            placeholder="email@exemplo.com"
            maxlength="254"
          />
        </div>
        
        <div class="form-group">
          <label class="form-label">{{ t('composer.guardian.phoneLabel') }}</label>
          <input 
            v-model="guardianForm.phone"
            type="tel"
            class="form-input"
            :placeholder="t('composer.guardian.phonePlaceholder')"
            maxlength="20"
          />
        </div>
      </div>
      
      <div class="form-group">
        <label class="form-label">{{ t('composer.guardian.relationshipLabel') }}</label>
        <div class="category-chips">
          <button 
            v-for="rel in relationships"
            :key="rel"
            type="button"
            :class="['chip', 'chip--small', { 'chip--active': guardianForm.relationship === rel }]"
            @click="guardianForm.relationship = guardianForm.relationship === rel ? '' : rel"
          >
            {{ t(`composer.relationships.${rel}`) }}
          </button>
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">
          🔒 {{ t('composer.guardian.pinLabel') }} <span class="required-indicator">*</span>
        </label>
        <input 
          v-model="guardianForm.accessPin"
          type="password"
          class="form-input"
          :class="{ 'form-input--error': formError && !guardianForm.accessPin }"
          :placeholder="t('composer.guardian.pinPlaceholder')"
          minlength="4"
          maxlength="20"
          required
        />
        <small class="form-hint">{{ t('composer.guardian.pinHint') }}</small>
      </div>
      
      <div v-if="formError" class="form-error-box">
        <span class="form-error-icon">⚠️</span>
        <span>{{ formError }}</span>
      </div>

      <button type="submit" class="btn btn--primary" :disabled="saving || !guardianForm.name || !guardianForm.accessPin">
        {{ saving ? t('composer.guardian.saving') : t('composer.guardian.saveButton') }}
      </button>
    </form>

    <!-- Memory Form -->
    <form v-else @submit.prevent="saveMemory" class="composer__form">
      <p class="composer__hint">
        💝 {{ t('composer.memory.hint') }}
      </p>
      
      <div class="form-group">
        <label class="form-label">{{ t('composer.memory.titleLabel') }}</label>
        <input 
          v-model="memoryForm.title"
          type="text"
          class="form-input"
          :placeholder="t('composer.memory.titlePlaceholder')"
          maxlength="255"
          required
        />
        <small class="form-hint">{{ t('common.maxChars', { count: 255 }) }}</small>
      </div>
      
      <div class="form-group">
        <label class="form-label">{{ t('composer.memory.recipientLabel') }}</label>
        <input 
          v-model="memoryForm.recipient"
          type="text"
          class="form-input"
          :placeholder="t('composer.memory.recipientPlaceholder')"
          maxlength="255"
        />
        <small class="form-hint">{{ t('common.maxChars', { count: 255 }) }}</small>
      </div>
      
      <div class="form-group">
        <label class="form-label">{{ t('composer.memory.messageLabel') }}</label>
        <textarea 
          v-model="memoryForm.content"
          class="form-textarea"
          :placeholder="t('composer.memory.messagePlaceholder')"
          rows="6"
          maxlength="10000"
        ></textarea>
      </div>

      <div class="form-group share-toggle">
        <label class="toggle-label">
          <input type="checkbox" v-model="memoryForm.isShared" class="toggle-input" />
          <span class="toggle-switch"></span>
          <span class="toggle-text">
            <span class="toggle-icon">👥</span>
            {{ t('composer.shareWithGuardians') }}
          </span>
        </label>
        <small class="toggle-hint">{{ t('composer.shareHint') }}</small>
      </div>
      
      <button type="submit" class="btn btn--primary" :disabled="saving || !memoryForm.title">
        {{ saving ? t('composer.memory.saving') : t('composer.memory.saveButton') }}
      </button>
      
      <div v-if="formError" class="form-error-box">
        <span class="form-error-icon">⚠️</span>
        <span>{{ formError }}</span>
      </div>
    </form>
  </div>
</template>

<style scoped>
.composer {
  background: var(--color-card);
  border-radius: var(--radius-xl);
  padding: var(--space-xl);
  border: 1px solid var(--color-border-light);
  box-shadow: var(--shadow-sm);
}

.composer__header {
  margin-bottom: var(--space-lg);
}

.composer__title {
  font-size: var(--font-size-xl);
  margin-bottom: var(--space-xs);
}

.composer__subtitle {
  color: var(--color-text-soft);
  margin: 0;
}

.composer__types {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
  margin-bottom: var(--space-lg);
}

.composer__form {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.composer__hint {
  padding: var(--space-md);
  background: var(--color-bg-warm);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  margin: 0;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-md);
}

.category-chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
}

.chip--small {
  padding: var(--space-xs) var(--space-sm);
  font-size: 0.8125rem;
}

/* Share Toggle */
.share-toggle {
  padding: var(--space-md);
  background: var(--color-primary-soft);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-primary);
  border-opacity: 0.2;
}

.toggle-label {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  cursor: pointer;
  user-select: none;
}

.toggle-input {
  display: none;
}

.toggle-switch {
  position: relative;
  width: 44px;
  height: 24px;
  background: var(--color-border);
  border-radius: 12px;
  transition: background 0.2s;
  flex-shrink: 0;
}

.toggle-switch::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 20px;
  height: 20px;
  background: white;
  border-radius: 50%;
  transition: transform 0.2s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
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
  font-weight: 500;
  color: var(--color-text);
}

.toggle-icon {
  font-size: 1.2em;
}

.toggle-hint {
  display: block;
  margin-top: var(--space-xs);
  margin-left: 56px;
  color: var(--color-text-muted);
  font-size: var(--font-size-xs);
}

/* Required field indicator */
.required-indicator {
  color: var(--color-danger);
  font-weight: 600;
}

/* Error state for inputs */
.form-input--error {
  border-color: var(--color-danger);
  background-color: #fef2f2;
}

.form-input--error:focus {
  border-color: var(--color-danger);
  box-shadow: 0 0 0 3px rgba(192, 74, 74, 0.15);
}

/* Error box - more visible error display */
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
}

.form-error-icon {
  font-size: 1.2em;
}

@media (max-width: 600px) {
  .form-row {
    grid-template-columns: 1fr;
  }
}
</style>
