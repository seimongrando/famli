<template>
  <div class="shared-page">
    <!-- Loading -->
    <div v-if="loading" class="loading-container">
      <div class="loading-card">
        <img src="/logo.svg" alt="Famli" class="loading-logo" />
        <div class="loading-spinner"></div>
        <p class="loading-text">{{ $t('common.loading') }}</p>
      </div>
    </div>

    <!-- PIN Required -->
    <div v-else-if="requiresPin" class="pin-container">
      <div class="pin-card">
        <img src="/logo.svg" alt="Famli" class="pin-logo" />
        <h1 class="pin-title">Famli</h1>
        <p class="pin-tagline">{{ $t('brand.tagline') }}</p>
        
        <div class="pin-divider"></div>
        
        <h2>🔒 {{ $t('shared.protected_access') }}</h2>
        <p class="pin-description">{{ $t('shared.protected_message') }}</p>
        
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
        <img src="/logo.svg" alt="Famli" class="error-logo" />
        <h1 class="error-brand">Famli</h1>
        <div class="error-divider"></div>
        <div class="error-icon">😔</div>
        <h2>{{ $t('shared.invalid_link') }}</h2>
        <p class="error-text">{{ error }}</p>
        <a href="/" class="btn-home">{{ $t('shared.go_to_famli') }}</a>
      </div>
    </div>

    <!-- Content -->
    <div v-else-if="sharedView" class="content-wrapper">
      <!-- Header com Logo -->
      <header class="famli-header">
        <div class="header-container">
          <div class="header-brand">
            <img src="/logo.svg" alt="Famli" class="header-logo" />
            <div class="header-title">
              <h1>Famli</h1>
              <p class="header-tagline">{{ $t('brand.tagline') }}</p>
            </div>
          </div>
          <div class="header-badge" :class="linkType">
            {{ linkTypeIcon }} {{ headerTitle }}
          </div>
        </div>
      </header>

      <!-- User Info Card -->
      <section class="user-section" v-if="sharedView.user_name || sharedView.guardian_name">
        <div class="user-card">
          <div class="user-avatar">👤</div>
          <div class="user-info">
            <h2 v-if="sharedView.user_name">{{ $t('shared.info_of', { name: sharedView.user_name }) }}</h2>
            <p v-if="sharedView.guardian_name" class="guardian-label">
              {{ $t('shared.viewing_as', { name: sharedView.guardian_name }) }}
            </p>
            <p class="access-time">{{ $t('shared.accessed_at', { date: formatDate(sharedView.accessed_at) }) }}</p>
          </div>
        </div>
      </section>

      <!-- Main Content -->
      <main class="content-main">
        <!-- Items by Category -->
        <section class="items-section" v-if="groupedItems.length > 0">
          <div v-for="category in groupedItems" :key="category.name" class="category-group">
            <h3 class="category-title">
              <span class="category-icon">{{ getCategoryIcon(category.name) }}</span>
              {{ formatCategory(category.name) }}
            </h3>
            
            <div class="items-grid">
              <article v-for="item in category.items" :key="item.id" class="item-card" :class="{ important: item.is_important }">
                <div class="item-header">
                  <span class="item-type-icon">{{ getTypeIcon(item.type) }}</span>
                  <h4 class="item-title">{{ item.title || '...' }}</h4>
                  <span v-if="item.is_important" class="important-badge" title="Importante">⭐</span>
                </div>
                <div class="item-content" v-if="item.content">
                  <p>{{ item.content }}</p>
                </div>
                <div class="item-footer" v-if="item.recipient">
                  <span class="recipient-label">💌 {{ $t('shared.for') }}: {{ item.recipient }}</span>
                </div>
              </article>
            </div>
          </div>
        </section>

        <!-- Empty State -->
        <div v-else class="empty-state">
          <div class="empty-icon">📭</div>
          <p>{{ $t('shared.no_items') }}</p>
        </div>

        <!-- Guardians (Memorial only) -->
        <section v-if="sharedView.guardians?.length" class="guardians-section">
          <h3 class="section-title">👥 {{ $t('shared.trusted_people') }}</h3>
          <div class="guardians-grid">
            <div v-for="guardian in sharedView.guardians" :key="guardian.id" class="guardian-card">
              <div class="guardian-avatar">👤</div>
              <div class="guardian-info">
                <h4>{{ guardian.name }}</h4>
                <p v-if="guardian.relationship" class="guardian-relationship">{{ guardian.relationship }}</p>
                <p v-if="guardian.email" class="guardian-contact">📧 {{ guardian.email }}</p>
                <p v-if="guardian.phone" class="guardian-contact">📱 {{ guardian.phone }}</p>
              </div>
            </div>
          </div>
        </section>
      </main>

      <!-- Footer -->
      <footer class="famli-footer">
        <div class="footer-content">
          <img src="/logo.svg" alt="Famli" class="footer-logo" />
          <p class="footer-text">
            <a href="/" class="footer-link">Famli</a> — {{ $t('shared.tagline') }}
          </p>
          <p class="footer-privacy">🔒 {{ $t('shared.privacy_note') }}</p>
        </div>
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
const isGuardianAccess = computed(() => route.path.startsWith('/g/') || route.path.startsWith('/guardian/'))

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
    const cat = item.category || 'other'
    if (!groups[cat]) {
      groups[cat] = { name: cat, items: [] }
    }
    groups[cat].items.push(item)
  }
  
  // Ordenar por prioridade (usar chaves em inglês como padrão)
  const order = ['health', 'saúde', 'finances', 'finanças', 'family', 'família', 'documents', 'documentos', 'memories', 'memórias', 'other', 'outros']
  return Object.values(groups).sort((a, b) => {
    const aIndex = order.indexOf(a.name) >= 0 ? order.indexOf(a.name) : 999
    const bIndex = order.indexOf(b.name) >= 0 ? order.indexOf(b.name) : 999
    return aIndex - bIndex
  })
})

onMounted(async () => {
  await fetchSharedContent()
})

async function fetchSharedContent() {
  try {
    loading.value = true
    error.value = null
    
    // Determinar endpoint baseado no tipo de acesso
    const endpoint = isGuardianAccess.value 
      ? `/api/guardian-access/${token.value}`
      : `/api/shared/${token.value}`
    
    const response = await fetch(endpoint)
    const data = await response.json()
    
    if (!response.ok) {
      error.value = data.error || t('shared.invalid_link')
      return
    }
    
    // Verificar se precisa de PIN (tanto para guardião quanto para link)
    if (data.requires_pin) {
      requiresPin.value = true
      linkType.value = data.link_type || 'normal'
      // Guardar info do owner/guardian para mostrar na tela de PIN
      if (data.owner) {
        sharedView.value = {
          user_name: data.owner.name,
          guardian_name: data.guardian?.name
        }
      }
      return
    }
    
    // Processar resposta do guardião
    if (isGuardianAccess.value) {
      sharedView.value = {
        items: data.items,
        guardians: [],
        user_name: data.owner?.name,
        link_type: data.access_type || 'normal',
        accessed_at: data.accessed_at,
        guardian_name: data.guardian?.name,
        guardian_relationship: data.guardian?.relationship
      }
      linkType.value = data.access_type || 'normal'
      return
    }
    
    sharedView.value = data
    linkType.value = data.link_type
  } catch (err) {
    error.value = t('shared.error_loading')
  } finally {
    loading.value = false
  }
}

async function verifyPin() {
  if (!pin.value) return
  
  try {
    verifying.value = true
    pinError.value = null
    
    // Usar endpoint correto baseado no tipo de acesso
    const endpoint = isGuardianAccess.value 
      ? `/api/guardian-access/${token.value}/verify`
      : `/api/shared/${token.value}/verify`
    
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ pin: pin.value })
    })
    
    const data = await response.json()
    
    if (!response.ok) {
      pinError.value = data.error || t('shared.invalid_pin')
      return
    }
    
    // Processar resposta do guardião
    if (isGuardianAccess.value) {
      sharedView.value = {
        items: data.items,
        guardians: [],
        user_name: data.owner?.name,
        link_type: data.access_type || 'normal',
        accessed_at: data.accessed_at,
        guardian_name: data.guardian?.name,
        guardian_relationship: data.guardian?.relationship
      }
      linkType.value = data.access_type || 'normal'
    } else {
      sharedView.value = data
      linkType.value = data.link_type
    }
    requiresPin.value = false
  } catch (err) {
    pinError.value = t('shared.error_verifying_pin')
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
    // Chaves em inglês (padrão)
    'health': '🏥',
    'finances': '💰',
    'family': '👨‍👩‍👧‍👦',
    'documents': '📄',
    'memories': '💭',
    'other': '📦',
    // Chaves em português (legado)
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
/* =============================================================================
   FAMLI SHARED PAGE - Visual Identity
   ============================================================================= */

.shared-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #faf8f5 0%, #f5f0e8 100%);
  font-family: 'Nunito', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

/* =============================================================================
   LOADING STATE
   ============================================================================= */
.loading-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #2d5a47 0%, #1e3d30 100%);
}

.loading-card {
  text-align: center;
  padding: 3rem;
}

.loading-logo {
  width: 80px;
  height: 80px;
  margin-bottom: 1.5rem;
  animation: pulse 2s ease-in-out infinite;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  margin: 0 auto 1rem;
  border: 3px solid rgba(255, 255, 255, 0.2);
  border-top-color: #e07b39;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.loading-text {
  color: rgba(255, 255, 255, 0.9);
  font-size: 1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.8; transform: scale(0.95); }
}

/* =============================================================================
   PIN VERIFICATION
   ============================================================================= */
.pin-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 1.5rem;
  background: linear-gradient(135deg, #2d5a47 0%, #1e3d30 100%);
}

.pin-card {
  background: white;
  border-radius: 1.5rem;
  padding: 2.5rem 2rem;
  text-align: center;
  box-shadow: 0 25px 80px rgba(0, 0, 0, 0.3);
  max-width: 420px;
  width: 100%;
}

.pin-logo {
  width: 70px;
  height: 70px;
  margin-bottom: 0.75rem;
}

.pin-title {
  font-size: 2rem;
  font-weight: 800;
  color: #2d5a47;
  margin: 0 0 0.25rem;
}

.pin-tagline {
  color: #6b665c;
  font-size: 0.9rem;
  margin: 0;
}

.pin-divider {
  height: 1px;
  background: linear-gradient(90deg, transparent, #e5ddd0, transparent);
  margin: 1.5rem 0;
}

.pin-description {
  color: #6b665c;
  font-size: 0.95rem;
  margin: 0.5rem 0 0;
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
  border: 2px solid #e5ddd0;
  border-radius: 12px;
  letter-spacing: 0.5rem;
  font-family: inherit;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.pin-input:focus {
  outline: none;
  border-color: #2d5a47;
  box-shadow: 0 0 0 3px rgba(45, 90, 71, 0.1);
}

.btn-verify {
  padding: 1rem;
  background: #e07b39;
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 1rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.btn-verify:hover:not(:disabled) {
  background: #c66a2e;
  transform: translateY(-1px);
}

.btn-verify:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.error-message {
  color: #dc2626;
  margin-top: 1rem;
  font-weight: 600;
  font-size: 0.9rem;
}

/* =============================================================================
   ERROR STATE
   ============================================================================= */
.error-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 1.5rem;
  background: linear-gradient(135deg, #2d5a47 0%, #1e3d30 100%);
}

.error-card {
  background: white;
  border-radius: 1.5rem;
  padding: 2.5rem 2rem;
  text-align: center;
  box-shadow: 0 25px 80px rgba(0, 0, 0, 0.3);
  max-width: 420px;
  width: 100%;
}

.error-logo {
  width: 60px;
  height: 60px;
  margin-bottom: 0.5rem;
}

.error-brand {
  font-size: 1.75rem;
  font-weight: 800;
  color: #2d5a47;
  margin: 0;
}

.error-divider {
  height: 1px;
  background: linear-gradient(90deg, transparent, #e5ddd0, transparent);
  margin: 1.5rem 0;
}

.error-icon {
  font-size: 4rem;
  margin-bottom: 0.5rem;
}

.error-text {
  color: #6b665c;
  margin: 0.5rem 0 0;
}

.btn-home {
  display: inline-block;
  margin-top: 1.5rem;
  padding: 0.875rem 2rem;
  background: #e07b39;
  color: white;
  text-decoration: none;
  border-radius: 12px;
  font-weight: 700;
  transition: all 0.2s;
}

.btn-home:hover {
  background: #c66a2e;
  transform: translateY(-1px);
}

/* =============================================================================
   MAIN CONTENT WRAPPER
   ============================================================================= */
.content-wrapper {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* =============================================================================
   FAMLI HEADER
   ============================================================================= */
.famli-header {
  background: linear-gradient(135deg, #2d5a47 0%, #1e3d30 100%);
  padding: 1.5rem 2rem;
  color: white;
}

.header-container {
  max-width: 1000px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.header-brand {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.header-logo {
  width: 50px;
  height: 50px;
}

.header-title h1 {
  margin: 0;
  font-size: 1.75rem;
  font-weight: 800;
}

.header-tagline {
  margin: 0;
  font-size: 0.85rem;
  opacity: 0.9;
}

.header-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border-radius: 50px;
  font-weight: 600;
  font-size: 0.9rem;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(10px);
}

.header-badge.normal {
  background: rgba(224, 123, 57, 0.3);
}

.header-badge.emergency {
  background: rgba(220, 38, 38, 0.3);
}

.header-badge.memorial {
  background: rgba(0, 0, 0, 0.3);
}

/* =============================================================================
   USER SECTION
   ============================================================================= */
.user-section {
  padding: 0 2rem;
  margin-top: -1.5rem;
}

.user-card {
  max-width: 1000px;
  margin: 0 auto;
  background: white;
  border-radius: 1rem;
  padding: 1.5rem 2rem;
  display: flex;
  align-items: center;
  gap: 1.25rem;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
}

.user-avatar {
  font-size: 2.5rem;
  background: linear-gradient(135deg, #e5ddd0 0%, #d5cec3 100%);
  border-radius: 50%;
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-info h2 {
  margin: 0 0 0.25rem;
  font-size: 1.25rem;
  color: #2c2a26;
}

.guardian-label {
  color: #2d5a47;
  font-weight: 600;
  margin: 0 0 0.25rem;
  font-size: 0.9rem;
}

.access-time {
  margin: 0;
  color: #8a857a;
  font-size: 0.85rem;
}

/* =============================================================================
   MAIN CONTENT
   ============================================================================= */
.content-main {
  flex: 1;
  max-width: 1000px;
  margin: 0 auto;
  padding: 2rem;
  width: 100%;
  box-sizing: border-box;
  /* Evitar overflow horizontal */
  overflow-x: hidden;
}

/* =============================================================================
   ITEMS SECTION
   ============================================================================= */
.items-section {
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
}

.category-group {
  /* grouped categories */
}

.category-title {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 1.25rem;
  color: #2c2a26;
  margin: 0 0 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 2px solid #e5ddd0;
}

.category-icon {
  font-size: 1.5rem;
}

.items-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
}

.item-card {
  background: white;
  border-radius: 1rem;
  padding: 1.25rem 1.5rem;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.05);
  border: 1px solid #f0ebe3;
  transition: all 0.2s;
  /* Evitar que conteúdo longo quebre o layout */
  overflow: hidden;
  min-width: 0;
}

.item-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
}

.item-card.important {
  border-left: 4px solid #e07b39;
  background: linear-gradient(90deg, #fef7f1 0%, white 30%);
}

.item-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.item-type-icon {
  font-size: 1.25rem;
}

.item-title {
  flex: 1;
  margin: 0;
  font-size: 1rem;
  font-weight: 700;
  color: #2c2a26;
  /* Evitar quebra de layout com títulos longos */
  word-break: break-word;
  overflow-wrap: break-word;
  min-width: 0;
}

.important-badge {
  font-size: 1rem;
}

.item-content {
  color: #5c584f;
  font-size: 0.95rem;
  line-height: 1.6;
  /* Evitar quebra de layout com conteúdos longos */
  max-width: 100%;
  overflow: hidden;
}

.item-content p {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: break-word;
}

.item-footer {
  margin-top: 1rem;
  padding-top: 0.75rem;
  border-top: 1px solid #f0ebe3;
}

.recipient-label {
  color: #8a857a;
  font-size: 0.85rem;
}

/* =============================================================================
   EMPTY STATE
   ============================================================================= */
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.empty-state p {
  color: #8a857a;
  font-size: 1.1rem;
}

/* =============================================================================
   GUARDIANS SECTION
   ============================================================================= */
.guardians-section {
  margin-top: 3rem;
  padding-top: 2rem;
  border-top: 2px solid #e5ddd0;
}

.section-title {
  font-size: 1.25rem;
  color: #2c2a26;
  margin: 0 0 1.5rem;
}

.guardians-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

.guardian-card {
  background: white;
  border-radius: 1rem;
  padding: 1.5rem;
  display: flex;
  gap: 1rem;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.05);
  border: 1px solid #f0ebe3;
}

.guardian-avatar {
  font-size: 2rem;
  background: linear-gradient(135deg, #e5ddd0 0%, #d5cec3 100%);
  border-radius: 50%;
  width: 50px;
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.guardian-info h4 {
  margin: 0 0 0.25rem;
  color: #2c2a26;
  font-size: 1rem;
}

.guardian-relationship {
  margin: 0 0 0.5rem;
  color: #2d5a47;
  font-weight: 600;
  font-size: 0.85rem;
}

.guardian-contact {
  margin: 0.25rem 0 0;
  color: #8a857a;
  font-size: 0.85rem;
}

/* =============================================================================
   FOOTER
   ============================================================================= */
.famli-footer {
  background: #2c2a26;
  padding: 2rem;
  margin-top: auto;
}

.footer-content {
  max-width: 1000px;
  margin: 0 auto;
  text-align: center;
}

.footer-logo {
  width: 40px;
  height: 40px;
  margin-bottom: 0.75rem;
  opacity: 0.9;
}

.footer-text {
  color: rgba(255, 255, 255, 0.9);
  margin: 0 0 0.5rem;
}

.footer-link {
  color: #e07b39;
  text-decoration: none;
  font-weight: 700;
}

.footer-link:hover {
  text-decoration: underline;
}

.footer-privacy {
  color: rgba(255, 255, 255, 0.6);
  font-size: 0.8rem;
  margin: 0;
}

/* =============================================================================
   RESPONSIVE
   ============================================================================= */
@media (max-width: 768px) {
  .famli-header {
    padding: 1.25rem 1rem;
  }
  
  .header-container {
    flex-direction: column;
    text-align: center;
  }
  
  .header-brand {
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .user-section {
    padding: 0 1rem;
    margin-top: -1rem;
  }
  
  .user-card {
    flex-direction: column;
    text-align: center;
    padding: 1.25rem;
  }
  
  .content-main {
    padding: 1.5rem 1rem;
  }
  
  .items-grid {
    grid-template-columns: 1fr;
  }
  
  .guardians-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 480px) {
  .pin-card,
  .error-card {
    padding: 1.75rem 1.25rem;
    border-radius: 1rem;
  }
  
  .pin-title,
  .error-brand {
    font-size: 1.5rem;
  }
  
  .header-title h1 {
    font-size: 1.25rem;
  }
}
</style>

