<script setup>
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useBoxStore } from '../stores/box'
import CharCounter from './CharCounter.vue'
import ErrorModal from './ErrorModal.vue'

const { t } = useI18n()
const boxStore = useBoxStore()
const formError = ref('')
const showErrorModal = ref(false)
const errorModalMessage = ref('')

// Limites de caracteres
const LIMITS = {
  title: 255,
  content: 10000,
  name: 255,
  email: 254,
  phone: 20,
  pin: 20,
  recipient: 255
}

// Função para traduzir erros do backend
function translateError(error) {
  // Mapeamento de erros conhecidos do backend
  const errorMap = {
    'title is required': 'apiErrors.title_required',
    'title too long': 'apiErrors.title_too_long',
    'content too long': 'apiErrors.content_too_long',
    'name is required': 'apiErrors.name_required',
    'name too long': 'apiErrors.name_too_long',
    'invalid email': 'apiErrors.email_invalid',
    'email too long': 'apiErrors.email_too_long',
    'phone too long': 'apiErrors.phone_too_long',
    'pin is required': 'apiErrors.pin_required',
    'pin too short': 'apiErrors.pin_too_short',
    'pin too long': 'apiErrors.pin_too_long',
    'recipient too long': 'apiErrors.recipient_too_long',
    'invalid type': 'apiErrors.type_invalid',
    'item not found': 'apiErrors.item_not_found',
    'guardian not found': 'apiErrors.guardian_not_found',
    'duplicate email': 'apiErrors.duplicate_guardian_email',
    'unauthorized': 'apiErrors.unauthorized',
    'rate limit exceeded': 'apiErrors.rate_limit',
    'Título é obrigatório': 'apiErrors.title_required',
    'Título muito longo': 'apiErrors.title_too_long',
    'Conteúdo muito longo': 'apiErrors.content_too_long',
    'Nome é obrigatório': 'apiErrors.name_required',
    'Nome muito longo': 'apiErrors.name_too_long',
    'E-mail inválido': 'apiErrors.email_invalid',
    'E-mail muito longo': 'apiErrors.email_too_long',
    'Telefone muito longo': 'apiErrors.phone_too_long',
    'PIN é obrigatório': 'apiErrors.pin_required',
    'PIN muito curto': 'apiErrors.pin_too_short',
    'PIN muito longo': 'apiErrors.pin_too_long',
    'Destinatário muito longo': 'apiErrors.recipient_too_long',
    'Tipo inválido': 'apiErrors.type_invalid',
    'Item não encontrado': 'apiErrors.item_not_found',
    'Guardião não encontrado': 'apiErrors.guardian_not_found',
    'E-mail já cadastrado': 'apiErrors.duplicate_guardian_email',
    'Não autorizado': 'apiErrors.unauthorized',
    'Muitas requisições': 'apiErrors.rate_limit'
  }

  if (!error) return t('apiErrors.generic')
  
  // Procurar match exato ou parcial
  const lowerError = error.toLowerCase()
  for (const [key, translationKey] of Object.entries(errorMap)) {
    if (lowerError.includes(key.toLowerCase())) {
      return t(translationKey)
    }
  }
  
  // Se não encontrou tradução, retornar erro genérico
  return t('apiErrors.generic')
}

// Mostrar erro em modal
function showError(error) {
  errorModalMessage.value = translateError(error)
  showErrorModal.value = true
}

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
    formError.value = t('apiErrors.title_required')
    return
  }
  
  saving.value = true
  formError.value = ''
  
  try {
    const result = await boxStore.createItem({
      type: 'info',
      title: infoForm.value.title,
      content: infoForm.value.content,
      category: infoForm.value.category,
      is_shared: infoForm.value.isShared
    })
    
    if (result) {
      infoForm.value = { title: '', content: '', category: '', isShared: false }
      emit('saved', 'info')
    } else {
      showError(boxStore.error)
    }
  } catch (error) {
    showError(boxStore.error || error.message)
  } finally {
    saving.value = false
  }
}

async function saveGuardian() {
  if (!guardianForm.value.name) {
    formError.value = t('apiErrors.name_required')
    return
  }
  if (!guardianForm.value.accessPin) {
    formError.value = t('apiErrors.pin_required')
    return
  }
  if (guardianForm.value.accessPin.length < 4) {
    formError.value = t('apiErrors.pin_too_short')
    return
  }
  
  saving.value = true
  formError.value = ''
  
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
      guardianForm.value = { name: '', email: '', phone: '', relationship: '', accessPin: '' }
      emit('saved', 'guardian')
    } else {
      showError(boxStore.error)
    }
  } catch (error) {
    showError(boxStore.error || error.message)
  } finally {
    saving.value = false
  }
}

async function saveMemory() {
  if (!memoryForm.value.title) {
    formError.value = t('apiErrors.title_required')
    return
  }
  
  saving.value = true
  formError.value = ''
  
  try {
    const result = await boxStore.createItem({
      type: 'memory',
      title: memoryForm.value.title,
      content: memoryForm.value.content,
      recipient: memoryForm.value.recipient,
      is_shared: memoryForm.value.isShared
    })
    
    if (result) {
      memoryForm.value = { title: '', content: '', recipient: '', isShared: false }
      emit('saved', 'memory')
    } else {
      showError(boxStore.error)
    }
  } catch (error) {
    showError(boxStore.error || error.message)
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
          :class="{ 'form-input--error': infoForm.title.length > LIMITS.title }"
          :placeholder="t('composer.info.titlePlaceholder')"
          :maxlength="LIMITS.title"
          required
        />
        <div class="form-hint-row">
          <CharCounter :current="infoForm.title.length" :max="LIMITS.title" />
        </div>
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
          :class="{ 'form-input--error': infoForm.content.length > LIMITS.content }"
          :placeholder="t('composer.info.detailsPlaceholder')"
          rows="4"
          :maxlength="LIMITS.content"
        ></textarea>
        <div class="form-hint-row">
          <CharCounter :current="infoForm.content.length" :max="LIMITS.content" />
        </div>
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
          :class="{ 'form-input--error': guardianForm.name.length > LIMITS.name }"
          :placeholder="t('composer.guardian.namePlaceholder')"
          :maxlength="LIMITS.name"
          required
        />
        <div class="form-hint-row">
          <CharCounter :current="guardianForm.name.length" :max="LIMITS.name" />
        </div>
      </div>
      
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">{{ t('composer.guardian.emailLabel') }}</label>
          <input 
            v-model="guardianForm.email"
            type="email"
            class="form-input"
            :class="{ 'form-input--error': guardianForm.email.length > LIMITS.email }"
            placeholder="email@exemplo.com"
            :maxlength="LIMITS.email"
          />
          <div class="form-hint-row">
            <CharCounter :current="guardianForm.email.length" :max="LIMITS.email" />
          </div>
        </div>
        
        <div class="form-group">
          <label class="form-label">{{ t('composer.guardian.phoneLabel') }}</label>
          <input 
            v-model="guardianForm.phone"
            type="tel"
            class="form-input"
            :class="{ 'form-input--error': guardianForm.phone.length > LIMITS.phone }"
            :placeholder="t('composer.guardian.phonePlaceholder')"
            :maxlength="LIMITS.phone"
          />
          <div class="form-hint-row">
            <CharCounter :current="guardianForm.phone.length" :max="LIMITS.phone" />
          </div>
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
          :class="{ 'form-input--error': formError && (!guardianForm.accessPin || guardianForm.accessPin.length < 4) }"
          :placeholder="t('composer.guardian.pinPlaceholder')"
          minlength="4"
          :maxlength="LIMITS.pin"
          required
        />
        <div class="form-hint-row">
          <small class="form-hint">{{ t('composer.guardian.pinHint') }}</small>
          <CharCounter :current="guardianForm.accessPin.length" :max="LIMITS.pin" />
        </div>
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
          :class="{ 'form-input--error': memoryForm.title.length > LIMITS.title }"
          :placeholder="t('composer.memory.titlePlaceholder')"
          :maxlength="LIMITS.title"
          required
        />
        <div class="form-hint-row">
          <CharCounter :current="memoryForm.title.length" :max="LIMITS.title" />
        </div>
      </div>
      
      <div class="form-group">
        <label class="form-label">{{ t('composer.memory.recipientLabel') }}</label>
        <input 
          v-model="memoryForm.recipient"
          type="text"
          class="form-input"
          :class="{ 'form-input--error': memoryForm.recipient.length > LIMITS.recipient }"
          :placeholder="t('composer.memory.recipientPlaceholder')"
          :maxlength="LIMITS.recipient"
        />
        <div class="form-hint-row">
          <CharCounter :current="memoryForm.recipient.length" :max="LIMITS.recipient" />
        </div>
      </div>
      
      <div class="form-group">
        <label class="form-label">{{ t('composer.memory.messageLabel') }}</label>
        <textarea 
          v-model="memoryForm.content"
          class="form-textarea"
          :class="{ 'form-input--error': memoryForm.content.length > LIMITS.content }"
          :placeholder="t('composer.memory.messagePlaceholder')"
          rows="6"
          :maxlength="LIMITS.content"
        ></textarea>
        <div class="form-hint-row">
          <CharCounter :current="memoryForm.content.length" :max="LIMITS.content" />
        </div>
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
    
    <!-- Error Modal -->
    <ErrorModal 
      :show="showErrorModal"
      :message="errorModalMessage"
      @close="showErrorModal = false"
    />
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

/* Form hint row - alinha hint e contador */
.form-hint-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: var(--space-xs);
  gap: var(--space-sm);
}

.form-hint-row .form-hint {
  margin: 0;
  flex: 1;
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
