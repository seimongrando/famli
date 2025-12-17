<script setup>
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useBoxStore } from '../stores/box'

const { t } = useI18n()
const boxStore = useBoxStore()

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
const infoForm = ref({ title: '', content: '', category: '' })
const guardianForm = ref({ name: '', email: '', phone: '', relationship: '' })
const memoryForm = ref({ title: '', content: '', recipient: '' })

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
  console.log('[Composer] Saving info:', infoForm.value.title)
  
  try {
    const result = await boxStore.createItem({
      type: 'info',
      title: infoForm.value.title,
      content: infoForm.value.content,
      category: infoForm.value.category
    })
    
    if (result) {
      console.log('[Composer] Info saved successfully:', result.id)
      infoForm.value = { title: '', content: '', category: '' }
      emit('saved', 'info')
    } else {
      console.error('[Composer] Info save failed - no result')
    }
  } catch (error) {
    console.error('[Composer] Info save error:', error)
  } finally {
    saving.value = false
  }
}

async function saveGuardian() {
  if (!guardianForm.value.name) {
    console.log('[Composer] Guardian form validation failed: name required')
    return
  }
  
  saving.value = true
  console.log('[Composer] Saving guardian:', guardianForm.value.name)
  
  try {
    const result = await boxStore.createGuardian({
      name: guardianForm.value.name,
      email: guardianForm.value.email,
      phone: guardianForm.value.phone,
      relationship: guardianForm.value.relationship
    })
    
    if (result) {
      console.log('[Composer] Guardian saved successfully:', result.id)
      guardianForm.value = { name: '', email: '', phone: '', relationship: '' }
      emit('saved', 'guardian')
    } else {
      console.error('[Composer] Guardian save failed - no result')
    }
  } catch (error) {
    console.error('[Composer] Guardian save error:', error)
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
  console.log('[Composer] Saving memory:', memoryForm.value.title)
  
  try {
    const result = await boxStore.createItem({
      type: 'memory',
      title: memoryForm.value.title,
      content: memoryForm.value.content,
      recipient: memoryForm.value.recipient
    })
    
    if (result) {
      console.log('[Composer] Memory saved successfully:', result.id)
      memoryForm.value = { title: '', content: '', recipient: '' }
      emit('saved', 'memory')
    } else {
      console.error('[Composer] Memory save failed - no result')
    }
  } catch (error) {
    console.error('[Composer] Memory save error:', error)
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
        ></textarea>
      </div>
      
      <button type="submit" class="btn btn--primary" :disabled="saving || !infoForm.title">
        {{ saving ? t('composer.info.saving') : t('composer.info.saveButton') }}
      </button>
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
          required
        />
      </div>
      
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">{{ t('composer.guardian.emailLabel') }}</label>
          <input 
            v-model="guardianForm.email"
            type="email"
            class="form-input"
            placeholder="email@exemplo.com"
          />
        </div>
        
        <div class="form-group">
          <label class="form-label">{{ t('composer.guardian.phoneLabel') }}</label>
          <input 
            v-model="guardianForm.phone"
            type="tel"
            class="form-input"
            :placeholder="t('composer.guardian.phonePlaceholder')"
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
      
      <button type="submit" class="btn btn--primary" :disabled="saving || !guardianForm.name">
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
          required
        />
      </div>
      
      <div class="form-group">
        <label class="form-label">{{ t('composer.memory.recipientLabel') }}</label>
        <input 
          v-model="memoryForm.recipient"
          type="text"
          class="form-input"
          :placeholder="t('composer.memory.recipientPlaceholder')"
        />
      </div>
      
      <div class="form-group">
        <label class="form-label">{{ t('composer.memory.messageLabel') }}</label>
        <textarea 
          v-model="memoryForm.content"
          class="form-textarea"
          :placeholder="t('composer.memory.messagePlaceholder')"
          rows="6"
        ></textarea>
      </div>
      
      <button type="submit" class="btn btn--primary" :disabled="saving || !memoryForm.title">
        {{ saving ? t('composer.memory.saving') : t('composer.memory.saveButton') }}
      </button>
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

@media (max-width: 600px) {
  .form-row {
    grid-template-columns: 1fr;
  }
}
</style>
