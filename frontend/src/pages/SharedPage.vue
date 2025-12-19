<template>
  <div class="shared-page">
    <!-- Loading -->
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>{{ $t('common.loading') }}</p>
    </div>

    <!-- PIN Required -->
    <div v-else-if="requiresPin" class="pin-container">
      <div class="pin-card">
        <div class="logo-icon">
          <img src="/logo.svg" alt="Famli" width="60" />
        </div>
        <h2>{{ $t('shared.protected_access') }}</h2>
        <p>{{ $t('shared.protected_message') }}</p>
        
        <form @submit.prevent="verifyPin" class="pin-form">
          <input 
            v-model="pin" 
            type="password" 
            :placeholder="$t('shared.enter_pin')"
            maxlength="10"
            class="pin-input"
            :disabled="verifying"
          />
          <button type="submit" class="btn-verify" :disabled="verifying || !pin">
            {{ verifying ? $t('common.verifying') : $t('shared.access') }}
          </button>
        </form>

        <p v-if="pinError" class="error-message">{{ pinError }}</p>
      </div>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="error-container">
      <div class="error-card">
        <div class="error-icon">😔</div>
        <h2>{{ $t('shared.invalid_link') }}</h2>
        <p>{{ error }}</p>
        <a href="/" class="btn-home">{{ $t('shared.go_to_famli') }}</a>
      </div>
    </div>

    <!-- Content -->
    <div v-else-if="sharedView" class="content-container">
      <!-- Header -->
      <header class="shared-header" :class="linkTypeClass">
        <div class="header-content">
          <div class="header-icon">
            {{ linkTypeIcon }}
          </div>
          <div class="header-text">
            <h1>{{ headerTitle }}</h1>
            <p v-if="sharedView.message">{{ sharedView.message }}</p>
          </div>
        </div>
      </header>

      <!-- User Info -->
      <section class="user-info" v-if="sharedView.user_name">
        <h2>{{ $t('shared.info_of', { name: sharedView.user_name }) }}</h2>
        <p class="access-time">{{ $t('shared.accessed_at', { date: formatDate(sharedView.accessed_at) }) }}</p>
      </section>

      <!-- Items by Category -->
      <section class="items-section">
        <div v-for="category in groupedItems" :key="category.name" class="category-group">
          <h3 class="category-title">
            <span class="category-icon">{{ getCategoryIcon(category.name) }}</span>
            {{ formatCategory(category.name) }}
          </h3>
          
          <div class="items-grid">
            <div v-for="item in category.items" :key="item.id" class="item-card" :class="{ important: item.is_important }">
              <div class="item-header">
                <span class="item-type">{{ getTypeIcon(item.type) }}</span>
                <h4>{{ item.title }}</h4>
                <span v-if="item.is_important" class="important-badge">⭐</span>
              </div>
              <div class="item-content" v-if="item.content">
                <p>{{ item.content }}</p>
              </div>
              <div class="item-footer" v-if="item.recipient">
                <small>Para: {{ item.recipient }}</small>
              </div>
            </div>
          </div>
        </div>

        <div v-if="!sharedView.items?.length" class="empty-state">
          <p>{{ $t('shared.no_items') }}</p>
        </div>
      </section>

      <!-- Guardians (Memorial only) -->
      <section v-if="sharedView.guardians?.length" class="guardians-section">
        <h3>{{ $t('shared.trusted_people') }}</h3>
        <div class="guardians-grid">
          <div v-for="guardian in sharedView.guardians" :key="guardian.id" class="guardian-card">
            <div class="guardian-icon">👤</div>
            <div class="guardian-info">
              <h4>{{ guardian.name }}</h4>
              <p v-if="guardian.relationship">{{ guardian.relationship }}</p>
              <p v-if="guardian.email">{{ guardian.email }}</p>
              <p v-if="guardian.phone">{{ guardian.phone }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- Footer -->
      <footer class="shared-footer">
        <p>
          <a href="/">Famli</a> - {{ $t('shared.tagline') }}
        </p>
      </footer>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const { t, locale } = useI18n()

const loading = ref(true)
const requiresPin = ref(false)
const linkType = ref('normal')
const error = ref(null)
const sharedView = ref(null)
const pin = ref('')
const pinError = ref(null)
const verifying = ref(false)

const token = computed(() => route.params.token)

const linkTypeClass = computed(() => {
  return {
    'type-normal': linkType.value === 'normal',
    'type-emergency': linkType.value === 'emergency',
    'type-memorial': linkType.value === 'memorial'
  }
})

const linkTypeIcon = computed(() => {
  switch (linkType.value) {
    case 'emergency': return '🚨'
    case 'memorial': return '🕊️'
    default: return '📦'
  }
})

const headerTitle = computed(() => {
  switch (linkType.value) {
    case 'emergency': return t('shared.emergency_access')
    case 'memorial': return t('shared.memorial')
    default: return t('shared.famli_box')
  }
})

const groupedItems = computed(() => {
  if (!sharedView.value?.items) return []
  
  const groups = {}
  for (const item of sharedView.value.items) {
    const cat = item.category || 'outros'
    if (!groups[cat]) {
      groups[cat] = { name: cat, items: [] }
    }
    groups[cat].items.push(item)
  }
  
  // Ordenar por prioridade
  const order = ['saúde', 'finanças', 'família', 'documentos', 'memórias', 'outros']
  return Object.values(groups).sort((a, b) => {
    return order.indexOf(a.name) - order.indexOf(b.name)
  })
})

onMounted(async () => {
  await fetchSharedContent()
})

async function fetchSharedContent() {
  try {
    loading.value = true
    error.value = null
    
    const response = await fetch(`/api/shared/${token.value}`)
    const data = await response.json()
    
    if (!response.ok) {
      error.value = data.error || 'Link inválido ou expirado'
      return
    }
    
    if (data.requires_pin) {
      requiresPin.value = true
      linkType.value = data.link_type
      return
    }
    
    sharedView.value = data
    linkType.value = data.link_type
  } catch (err) {
    error.value = 'Erro ao carregar conteúdo'
  } finally {
    loading.value = false
  }
}

async function verifyPin() {
  if (!pin.value) return
  
  try {
    verifying.value = true
    pinError.value = null
    
    const response = await fetch(`/api/shared/${token.value}/verify`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ pin: pin.value })
    })
    
    const data = await response.json()
    
    if (!response.ok) {
      pinError.value = data.error || 'PIN incorreto'
      return
    }
    
    sharedView.value = data
    linkType.value = data.link_type
    requiresPin.value = false
  } catch (err) {
    pinError.value = 'Erro ao verificar PIN'
  } finally {
    verifying.value = false
  }
}

function formatDate(date) {
  const loc = locale.value === 'en' ? 'en-US' : 'pt-BR'
  return new Date(date).toLocaleString(loc)
}

function formatCategory(cat) {
  // Usar traduções para categorias
  const key = `categories.${cat}`
  const translated = t(key)
  // Se a tradução retornar a chave, usar o valor original capitalizado
  if (translated === key) {
    return cat.charAt(0).toUpperCase() + cat.slice(1)
  }
  return translated
}

function getCategoryIcon(cat) {
  const icons = {
    'saúde': '🏥',
    'finanças': '💰',
    'família': '👨‍👩‍👧‍👦',
    'documentos': '📄',
    'memórias': '💭',
    'outros': '📦'
  }
  return icons[cat] || '📦'
}

function getTypeIcon(type) {
  const icons = {
    'info': 'ℹ️',
    'memory': '💭',
    'note': '📝',
    'access': '🔑',
    'routine': '🔄',
    'location': '📍'
  }
  return icons[type] || '📄'
}
</script>

<style scoped>
.shared-page {
  min-height: 100vh;
  background: var(--color-bg, #faf8f5);
}

/* Loading */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  gap: 1rem;
}

.loading-spinner {
  width: 50px;
  height: 50px;
  border: 4px solid var(--color-border, #e5ddd0);
  border-top-color: var(--color-primary, #2d5a47);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* PIN Container */
.pin-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 1rem;
}

.pin-card {
  background: white;
  border-radius: 1rem;
  padding: 2rem;
  text-align: center;
  box-shadow: 0 10px 40px rgba(0,0,0,0.1);
  max-width: 400px;
  width: 100%;
}

.pin-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.pin-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-top: 1.5rem;
}

.pin-input {
  padding: 1rem;
  font-size: 1.5rem;
  text-align: center;
  border: 2px solid #e2e8f0;
  border-radius: 0.5rem;
  letter-spacing: 0.5rem;
}

.pin-input:focus {
  outline: none;
  border-color: var(--color-primary, #2d5a47);
}

.btn-verify {
  padding: 1rem;
  background: var(--color-accent, #e07b39);
  color: white;
  border: none;
  border-radius: var(--radius-md, 12px);
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-verify:hover:not(:disabled) {
  background: var(--color-accent-light, #f4a876);
}

.btn-verify:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.error-message {
  color: #ef4444;
  margin-top: 1rem;
}

/* Error Container */
.error-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 1rem;
}

.error-card {
  background: white;
  border-radius: 1rem;
  padding: 2rem;
  text-align: center;
  box-shadow: 0 10px 40px rgba(0,0,0,0.1);
  max-width: 400px;
}

.error-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.btn-home {
  display: inline-block;
  margin-top: 1.5rem;
  padding: 0.75rem 1.5rem;
  background: var(--color-accent, #e07b39);
  color: white;
  text-decoration: none;
  border-radius: var(--radius-md, 12px);
  font-weight: 600;
}

/* Content */
.content-container {
  max-width: 900px;
  margin: 0 auto;
  padding-bottom: 2rem;
}

/* Header */
.shared-header {
  padding: 3rem 2rem;
  color: white;
  text-align: center;
}

.shared-header.type-normal {
  background: var(--color-primary, #2d5a47);
}

.shared-header.type-emergency {
  background: linear-gradient(135deg, #c04a4a 0%, #e07b39 100%);
}

.shared-header.type-memorial {
  background: linear-gradient(135deg, #2c2a26 0%, #5c584f 100%);
}

.header-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.header-icon {
  font-size: 4rem;
}

.header-text h1 {
  margin: 0;
  font-size: 2rem;
}

.header-text p {
  margin: 0.5rem 0 0;
  opacity: 0.9;
}

/* User Info */
.user-info {
  background: white;
  padding: 1.5rem 2rem;
  text-align: center;
  border-bottom: 1px solid #e2e8f0;
}

.user-info h2 {
  margin: 0;
  color: #1e293b;
}

.access-time {
  margin: 0.5rem 0 0;
  color: #64748b;
  font-size: 0.875rem;
}

/* Items Section */
.items-section {
  padding: 2rem;
}

.category-group {
  margin-bottom: 2rem;
}

.category-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
  color: #1e293b;
  font-size: 1.25rem;
}

.items-grid {
  display: grid;
  gap: 1rem;
}

.item-card {
  background: white;
  border-radius: 0.75rem;
  padding: 1.25rem;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  border-left: 4px solid #e2e8f0;
  transition: transform 0.2s, box-shadow 0.2s;
}

.item-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0,0,0,0.1);
}

.item-card.important {
  border-left-color: #eab308;
  background: linear-gradient(135deg, #fffbeb 0%, white 100%);
}

.item-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.item-header h4 {
  margin: 0;
  flex: 1;
  color: #1e293b;
}

.important-badge {
  font-size: 1rem;
}

.item-content {
  margin-top: 0.75rem;
  color: #475569;
  line-height: 1.6;
}

.item-content p {
  margin: 0;
  white-space: pre-wrap;
}

.item-footer {
  margin-top: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid #e2e8f0;
  color: #64748b;
}

/* Guardians */
.guardians-section {
  padding: 0 2rem 2rem;
}

.guardians-section h3 {
  margin-bottom: 1rem;
  color: #1e293b;
}

.guardians-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 1rem;
}

.guardian-card {
  background: white;
  border-radius: 0.75rem;
  padding: 1.25rem;
  display: flex;
  gap: 1rem;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

.guardian-icon {
  font-size: 2rem;
}

.guardian-info h4 {
  margin: 0 0 0.25rem;
  color: #1e293b;
}

.guardian-info p {
  margin: 0;
  color: #64748b;
  font-size: 0.875rem;
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: 3rem;
  color: #64748b;
}

/* Footer */
.shared-footer {
  text-align: center;
  padding: 2rem;
  color: #64748b;
}

.shared-footer a {
  color: var(--color-primary, #2d5a47);
  text-decoration: none;
  font-weight: 600;
}

/* Responsive */
@media (max-width: 640px) {
  .shared-header {
    padding: 2rem 1rem;
  }
  
  .header-icon {
    font-size: 3rem;
  }
  
  .header-text h1 {
    font-size: 1.5rem;
  }
  
  .items-section,
  .guardians-section {
    padding: 1rem;
  }
}
</style>

