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

// Limites de caracteres (MVP)
const LIMITS = {
  title: 100,
  content: 2000,
  name: 100,
  email: 254,
  phone: 20,
  pin: 20,
  recipient: 100
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
const infoForm = ref({ title: '', content: '', category: '', isShared: false, guardianIds: [] })
const guardianForm = ref({ name: '', email: '', phone: '', relationship: '', accessPin: '' })
const memoryForm = ref({ title: '', content: '', recipient: '', isShared: false, guardianIds: [] })

// Guardiões do store
const guardians = computed(() => boxStore.guardians || [])

// Computed para seleção de guardiões (Info form)
const infoAllGuardiansSelected = computed(() => {
  return guardians.value.length > 0 && infoForm.value.guardianIds.length === guardians.value.length
})
const infoSomeGuardiansSelected = computed(() => {
  return infoForm.value.guardianIds.length > 0 && infoForm.value.guardianIds.length < guardians.value.length
})

// Computed para seleção de guardiões (Memory form)
const memoryAllGuardiansSelected = computed(() => {
  return guardians.value.length > 0 && memoryForm.value.guardianIds.length === guardians.value.length
})
const memorySomeGuardiansSelected = computed(() => {
  return memoryForm.value.guardianIds.length > 0 && memoryForm.value.guardianIds.length < guardians.value.length
})

// Funções para selecionar/deselecionar todos
function toggleInfoSelectAll() {
  if (infoAllGuardiansSelected.value) {
    infoForm.value.guardianIds = []
  } else {
    infoForm.value.guardianIds = guardians.value.map(g => g.id)
  }
}

function toggleMemorySelectAll() {
  if (memoryAllGuardiansSelected.value) {
    memoryForm.value.guardianIds = []
  } else {
    memoryForm.value.guardianIds = guardians.value.map(g => g.id)
  }
}

// Funções para toggle individual de guardião
function toggleGuardianInfo(guardianId) {
  const index = infoForm.value.guardianIds.indexOf(guardianId)
  if (index === -1) {
    infoForm.value.guardianIds.push(guardianId)
  } else {
    infoForm.value.guardianIds.splice(index, 1)
  }
}

function toggleGuardianMemory(guardianId) {
  const index = memoryForm.value.guardianIds.indexOf(guardianId)
  if (index === -1) {
    memoryForm.value.guardianIds.push(guardianId)
  } else {
    memoryForm.value.guardianIds.splice(index, 1)
  }
}

// Quando o toggle é ativado, selecionar todos por padrão
watch(() => infoForm.value.isShared, (newVal) => {
  if (newVal && infoForm.value.guardianIds.length === 0 && guardians.value.length > 0) {
    infoForm.value.guardianIds = guardians.value.map(g => g.id)
  }
})

watch(() => memoryForm.value.isShared, (newVal) => {
  if (newVal && memoryForm.value.guardianIds.length === 0 && guardians.value.length > 0) {
    memoryForm.value.guardianIds = guardians.value.map(g => g.id)
  }
})

// Quando todos os guardiões forem desmarcados, desativar isShared automaticamente
watch(() => infoForm.value.guardianIds, (newIds) => {
  if (infoForm.value.isShared && newIds.length === 0) {
    infoForm.value.isShared = false
  }
}, { deep: true })

watch(() => memoryForm.value.guardianIds, (newIds) => {
  if (memoryForm.value.isShared && newIds.length === 0) {
    memoryForm.value.isShared = false
  }
}, { deep: true })

// Função para gerar cor de avatar baseada no nome
function getAvatarColor(name) {
  const colors = [
    '#2D5A47', // verde escuro (primária)
    '#7C5E4A', // marrom quente
    '#4A6572', // azul acinzentado
    '#8B6B4F', // caramelo
    '#5A7A6B', // verde médio
    '#6B5B73', // roxo suave
    '#5C6B5E', // verde oliva
    '#7A6055'  // terracota
  ]
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
}

// Função para obter iniciais do nome
function getInitials(name) {
  if (!name) return '?'
  const parts = name.trim().split(/\s+/)
  if (parts.length === 1) {
    return parts[0].charAt(0).toUpperCase()
  }
  return (parts[0].charAt(0) + parts[parts.length - 1].charAt(0)).toUpperCase()
}

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
      is_shared: infoForm.value.isShared,
      guardian_ids: infoForm.value.isShared ? infoForm.value.guardianIds : []
    })
    
    if (result) {
      infoForm.value = { title: '', content: '', category: '', isShared: false, guardianIds: [] }
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
      is_shared: memoryForm.value.isShared,
      guardian_ids: memoryForm.value.isShared ? memoryForm.value.guardianIds : []
    })
    
    if (result) {
      memoryForm.value = { title: '', content: '', recipient: '', isShared: false, guardianIds: [] }
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
        
        <!-- Seleção de guardiões (aparece quando toggle ativo) -->
        <div v-if="infoForm.isShared && guardians.length > 0" class="guardians-selection-modern">
          <div class="guardians-header">
            <span class="guardians-count">
              {{ t('share.selected_count', { count: infoForm.guardianIds.length, total: guardians.length }) }}
            </span>
            <button 
              type="button" 
              class="select-all-btn"
              @click="toggleInfoSelectAll"
            >
              {{ infoAllGuardiansSelected ? t('share.deselect_all') : t('share.select_all') }}
            </button>
          </div>
          
          <!-- Grid de guardiões com avatares -->
          <div class="guardians-cards">
            <div 
              v-for="guardian in guardians" 
              :key="guardian.id" 
              class="guardian-card"
              :class="{ 'guardian-card--selected': infoForm.guardianIds.includes(guardian.id) }"
              @click="toggleGuardianInfo(guardian.id)"
            >
              <div 
                class="guardian-avatar"
                :style="{ backgroundColor: getAvatarColor(guardian.name) }"
              >
                {{ getInitials(guardian.name) }}
              </div>
              <div class="guardian-info">
                <span class="guardian-name">{{ guardian.name }}</span>
                <span v-if="guardian.relationship" class="guardian-relationship">
                  {{ t(`composer.relationships.${guardian.relationship}`) }}
                </span>
              </div>
              <div class="guardian-check">
                <svg v-if="infoForm.guardianIds.includes(guardian.id)" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="10" fill="currentColor"/>
                  <path d="M8 12l2.5 2.5L16 9" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                <svg v-else viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="9.5" stroke="currentColor" stroke-opacity="0.3"/>
                </svg>
              </div>
            </div>
          </div>
          
          <p v-if="infoForm.guardianIds.length === 0" class="guardians-hint">
            {{ t('share.select_hint') }}
          </p>
        </div>
        
        <!-- Info quando não há guardiões -->
        <div v-if="infoForm.isShared && guardians.length === 0" class="no-guardians-warning">
          <span class="warning-icon">⚠️</span>
          <span>{{ t('share.no_guardians') }}</span>
        </div>
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
        
        <!-- Seleção de guardiões (aparece quando toggle ativo) -->
        <div v-if="memoryForm.isShared && guardians.length > 0" class="guardians-selection-modern">
          <div class="guardians-header">
            <span class="guardians-count">
              {{ t('share.selected_count', { count: memoryForm.guardianIds.length, total: guardians.length }) }}
            </span>
            <button 
              type="button" 
              class="select-all-btn"
              @click="toggleMemorySelectAll"
            >
              {{ memoryAllGuardiansSelected ? t('share.deselect_all') : t('share.select_all') }}
            </button>
          </div>
          
          <!-- Grid de guardiões com avatares -->
          <div class="guardians-cards">
            <div 
              v-for="guardian in guardians" 
              :key="guardian.id" 
              class="guardian-card"
              :class="{ 'guardian-card--selected': memoryForm.guardianIds.includes(guardian.id) }"
              @click="toggleGuardianMemory(guardian.id)"
            >
              <div 
                class="guardian-avatar"
                :style="{ backgroundColor: getAvatarColor(guardian.name) }"
              >
                {{ getInitials(guardian.name) }}
              </div>
              <div class="guardian-info">
                <span class="guardian-name">{{ guardian.name }}</span>
                <span v-if="guardian.relationship" class="guardian-relationship">
                  {{ t(`composer.relationships.${guardian.relationship}`) }}
                </span>
              </div>
              <div class="guardian-check">
                <svg v-if="memoryForm.guardianIds.includes(guardian.id)" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="10" fill="currentColor"/>
                  <path d="M8 12l2.5 2.5L16 9" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                <svg v-else viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="9.5" stroke="currentColor" stroke-opacity="0.3"/>
                </svg>
              </div>
            </div>
          </div>
          
          <p v-if="memoryForm.guardianIds.length === 0" class="guardians-hint">
            {{ t('share.select_hint') }}
          </p>
        </div>
        
        <!-- Info quando não há guardiões -->
        <div v-if="memoryForm.isShared && guardians.length === 0" class="no-guardians-warning">
          <span class="warning-icon">⚠️</span>
          <span>{{ t('share.no_guardians') }}</span>
        </div>
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

/* Guardians Selection Panel */
/* Design moderno para seleção de guardiões */
.guardians-selection-modern {
  margin-top: var(--space-md);
  padding: var(--space-md);
  background: linear-gradient(135deg, #fafbfc 0%, #f5f7f6 100%);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-light);
}

.guardians-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-md);
}

.guardians-count {
  font-size: 0.85rem;
  color: var(--color-text-muted);
  font-weight: 500;
}

.select-all-btn {
  background: none;
  border: none;
  color: var(--color-primary);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: var(--radius-sm);
  transition: all 0.2s ease;
}

.select-all-btn:hover {
  background: var(--color-primary-soft);
}

.guardians-cards {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.guardian-card {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  padding: 0.875rem 1rem;
  background: white;
  border-radius: var(--radius-md);
  border: 2px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

.guardian-card:hover {
  border-color: var(--color-border);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.guardian-card--selected {
  border-color: var(--color-primary);
  background: linear-gradient(135deg, rgba(45, 90, 71, 0.03) 0%, rgba(45, 90, 71, 0.06) 100%);
  box-shadow: 0 2px 8px rgba(45, 90, 71, 0.12);
}

.guardian-avatar {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 0.9rem;
  letter-spacing: 0.02em;
  flex-shrink: 0;
}

.guardian-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.guardian-name {
  font-weight: 600;
  color: var(--color-text);
  font-size: 0.95rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.guardian-relationship {
  font-size: 0.8rem;
  color: var(--color-text-muted);
}

.guardian-check {
  width: 24px;
  height: 24px;
  color: var(--color-primary);
  flex-shrink: 0;
}

.guardian-check svg {
  width: 100%;
  height: 100%;
}

.guardians-hint {
  margin: 0;
  margin-top: var(--space-sm);
  padding: 0.75rem;
  background: #fef3cd;
  border-radius: var(--radius-sm);
  font-size: 0.8rem;
  color: #856404;
  text-align: center;
}

.no-guardians-warning {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: var(--space-sm);
  padding: 0.75rem 1rem;
  background: #fef3cd;
  border-radius: var(--radius-sm);
  color: #856404;
  font-size: 0.85rem;
}

.warning-icon {
  font-size: 1rem;
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
