<!-- =============================================================================
  FAMLI - Página de Perfil do Usuário
  =============================================================================
  Mostra os dados de acesso do usuário e permite gerenciar a conta.
  
  Funcionalidades:
  - Exibe dados do usuário (nome, email, data de criação)
  - Indica se é administrador
  - Link para área administrativa (se admin)
  - Opção de logout
============================================================================== -->

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { useLocalizedRoutes } from '../composables/useLocalizedRoutes'
import LanguageSelector from '../components/LanguageSelector.vue'
import ConfirmModal from '../components/ConfirmModal.vue'
import ShareLinksManager from '../components/ShareLinksManager.vue'

const { t, locale } = useI18n()
const router = useRouter()
const authStore = useAuthStore()
const { paths } = useLocalizedRoutes()

// Estado local
const loading = ref(true)
const exporting = ref(false)
const deleting = ref(false)

// Modal de exclusão
const showDeleteModal = ref(false)
const deletePassword = ref('')
const deleteConfirmation = ref('')
const deleteError = ref('')

// Computed
const user = computed(() => authStore.user)
const isAdmin = computed(() => user.value?.is_admin || false)

const formattedDate = computed(() => {
  if (!user.value?.created_at) return '-'
  const date = new Date(user.value.created_at)
  return date.toLocaleDateString(locale.value === 'pt-BR' ? 'pt-BR' : 'en-US', {
    day: '2-digit',
    month: 'long',
    year: 'numeric'
  })
})

// Texto de confirmação esperado para exclusão
const expectedConfirmation = computed(() => {
  return locale.value === 'pt-BR' ? 'EXCLUIR MINHA CONTA' : 'DELETE MY ACCOUNT'
})

const canDelete = computed(() => {
  return deletePassword.value.length > 0 && 
         deleteConfirmation.value.toUpperCase() === expectedConfirmation.value
})

// Lifecycle
onMounted(async () => {
  if (!authStore.isAuthenticated) {
    router.push({ name: 'auth' })
    return
  }
  
  // Atualizar dados do usuário
  await authStore.checkSession()
  loading.value = false
})

// Actions
async function handleLogout() {
  // Primeiro limpar o usuário do store
  await authStore.logout()
  
  // Usar location.href para garantir redirecionamento completo
  // e evitar problemas com o navigation guard
  window.location.href = '/'
}

// Exportar dados do usuário (LGPD)
async function exportUserData() {
  exporting.value = true
  try {
    const response = await fetch('/api/auth/export', {
      method: 'GET',
      credentials: 'include'
    })
    
    if (!response.ok) {
      throw new Error('Erro ao exportar dados')
    }
    
    const data = await response.json()
    
    // Criar e baixar arquivo JSON
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'famli-meus-dados.json'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch (error) {
    console.error('Erro ao exportar:', error)
    alert(t('profile.exportError'))
  } finally {
    exporting.value = false
  }
}

// Abrir modal de exclusão
function openDeleteModal() {
  deletePassword.value = ''
  deleteConfirmation.value = ''
  deleteError.value = ''
  showDeleteModal.value = true
}

// Cancelar exclusão
function cancelDelete() {
  showDeleteModal.value = false
  deletePassword.value = ''
  deleteConfirmation.value = ''
  deleteError.value = ''
}

// Confirmar exclusão de conta (LGPD)
async function confirmDeleteAccount() {
  if (!canDelete.value) return
  
  deleting.value = true
  deleteError.value = ''
  
  try {
    const response = await fetch('/api/auth/account', {
      method: 'DELETE',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        password: deletePassword.value,
        confirmation: deleteConfirmation.value
      })
    })
    
    if (!response.ok) {
      const data = await response.json()
      deleteError.value = data.error || t('profile.deleteError')
      return
    }
    
    // Conta excluída com sucesso
    showDeleteModal.value = false
    
    // Redirecionar para página inicial
    window.location.href = '/'
  } catch (error) {
    console.error('Erro ao excluir conta:', error)
    deleteError.value = t('profile.deleteError')
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div class="profile-page">
    <!-- Header -->
    <header class="profile-header">
      <div class="profile-header__left">
        <router-link :to="paths.dashboard" class="profile-header__back">
          ← {{ t('common.back') }}
        </router-link>
        <h1 class="profile-header__title">{{ t('profile.title') }}</h1>
      </div>
      <div class="profile-header__right">
        <LanguageSelector />
      </div>
    </header>

    <!-- Loading -->
    <div v-if="loading" class="profile-loading">
      <div class="spinner"></div>
      <p>{{ t('common.loading') }}</p>
    </div>

    <!-- Content -->
    <main v-else class="profile-main">
      <!-- User Info Card -->
      <div class="profile-card">
        <div class="profile-card__avatar">
          <span class="avatar-icon">👤</span>
        </div>
        
        <div class="profile-card__info">
          <h2 class="profile-card__name">{{ user?.name || t('profile.noName') }}</h2>
          <p class="profile-card__email">{{ user?.email }}</p>
          
          <div class="profile-badges">
            <span v-if="isAdmin" class="badge badge--admin">
              ⚙️ {{ t('profile.admin') }}
            </span>
            <span class="badge badge--member">
              📅 {{ t('profile.memberSince') }} {{ formattedDate }}
            </span>
          </div>
        </div>
      </div>

      <!-- Data Section -->
      <div class="profile-section">
        <h3 class="profile-section__title">{{ t('profile.accountData') }}</h3>
        
        <div class="data-grid">
          <div class="data-item">
            <span class="data-item__label">{{ t('profile.userId') }}</span>
            <span class="data-item__value data-item__value--mono">{{ user?.id }}</span>
          </div>
          
          <div class="data-item">
            <span class="data-item__label">{{ t('profile.email') }}</span>
            <span class="data-item__value">{{ user?.email }}</span>
          </div>
          
          <div class="data-item">
            <span class="data-item__label">{{ t('profile.name') }}</span>
            <span class="data-item__value">{{ user?.name || '-' }}</span>
          </div>
          
          <div class="data-item">
            <span class="data-item__label">{{ t('profile.createdAt') }}</span>
            <span class="data-item__value">{{ formattedDate }}</span>
          </div>
          
          <div class="data-item">
            <span class="data-item__label">{{ t('profile.role') }}</span>
            <span class="data-item__value">
              {{ isAdmin ? t('profile.adminRole') : t('profile.userRole') }}
            </span>
          </div>
        </div>
      </div>

      <!-- Share Links Section -->
      <div class="profile-section">
        <ShareLinksManager />
      </div>

      <!-- Actions Section -->
      <div class="profile-section">
        <h3 class="profile-section__title">{{ t('profile.actions') }}</h3>
        
        <div class="action-buttons">
          <router-link 
            v-if="isAdmin" 
            :to="paths.admin" 
            class="btn btn--primary"
          >
            ⚙️ {{ t('profile.goToAdmin') }}
          </router-link>
          
          <router-link :to="paths.dashboard" class="btn btn--secondary">
            📦 {{ t('profile.goToBox') }}
          </router-link>
          
          <button @click="handleLogout" class="btn btn--ghost btn--danger">
            🚪 {{ t('profile.logout') }}
          </button>
        </div>
      </div>

      <!-- LGPD Section - Direitos do Titular -->
      <div class="profile-section profile-section--lgpd">
        <h3 class="profile-section__title">
          🔐 {{ t('profile.lgpd.title') }}
        </h3>
        
        <p class="lgpd-description">{{ t('profile.lgpd.description') }}</p>
        
        <div class="lgpd-actions">
          <!-- Exportar Dados -->
          <div class="lgpd-action">
            <div class="lgpd-action__info">
              <span class="lgpd-action__icon">📥</span>
              <div>
                <h4 class="lgpd-action__title">{{ t('profile.lgpd.exportTitle') }}</h4>
                <p class="lgpd-action__description">{{ t('profile.lgpd.exportDescription') }}</p>
              </div>
            </div>
            <button 
              @click="exportUserData" 
              class="btn btn--secondary"
              :disabled="exporting"
            >
              {{ exporting ? t('common.loading') : t('profile.lgpd.exportButton') }}
            </button>
          </div>
          
          <!-- Excluir Conta -->
          <div class="lgpd-action lgpd-action--danger">
            <div class="lgpd-action__info">
              <span class="lgpd-action__icon">⚠️</span>
              <div>
                <h4 class="lgpd-action__title">{{ t('profile.lgpd.deleteTitle') }}</h4>
                <p class="lgpd-action__description">{{ t('profile.lgpd.deleteDescription') }}</p>
              </div>
            </div>
            <button 
              @click="openDeleteModal" 
              class="btn btn--danger"
            >
              {{ t('profile.lgpd.deleteButton') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Security Note -->
      <div class="profile-note">
        <span class="profile-note__icon">🔒</span>
        <p class="profile-note__text">
          {{ t('profile.securityNote') }}
        </p>
      </div>
    </main>

    <!-- Modal de Exclusão de Conta -->
    <div v-if="showDeleteModal" class="delete-modal-overlay" @click.self="cancelDelete">
      <div class="delete-modal">
        <div class="delete-modal__header">
          <h3>⚠️ {{ t('profile.lgpd.deleteModalTitle') }}</h3>
        </div>
        
        <div class="delete-modal__body">
          <div class="delete-warning">
            <p>{{ t('profile.lgpd.deleteWarning1') }}</p>
            <ul>
              <li>{{ t('profile.lgpd.deleteWarningItem1') }}</li>
              <li>{{ t('profile.lgpd.deleteWarningItem2') }}</li>
              <li>{{ t('profile.lgpd.deleteWarningItem3') }}</li>
              <li>{{ t('profile.lgpd.deleteWarningItem4') }}</li>
            </ul>
            <p class="delete-warning__highlight">{{ t('profile.lgpd.deleteWarning2') }}</p>
          </div>
          
          <div class="delete-form">
            <!-- Senha -->
            <div class="form-group">
              <label for="delete-password">{{ t('profile.lgpd.passwordLabel') }}</label>
              <input 
                id="delete-password"
                type="password" 
                v-model="deletePassword"
                :placeholder="t('profile.lgpd.passwordPlaceholder')"
                class="form-input"
              />
            </div>
            
            <!-- Confirmação -->
            <div class="form-group">
              <label for="delete-confirmation">
                {{ t('profile.lgpd.confirmationLabel') }}
                <code>{{ expectedConfirmation }}</code>
              </label>
              <input 
                id="delete-confirmation"
                type="text" 
                v-model="deleteConfirmation"
                :placeholder="expectedConfirmation"
                class="form-input"
              />
            </div>
            
            <!-- Erro -->
            <div v-if="deleteError" class="delete-error">
              {{ deleteError }}
            </div>
          </div>
        </div>
        
        <div class="delete-modal__footer">
          <button @click="cancelDelete" class="btn btn--ghost">
            {{ t('common.cancel') }}
          </button>
          <button 
            @click="confirmDeleteAccount" 
            class="btn btn--danger"
            :disabled="!canDelete || deleting"
          >
            {{ deleting ? t('common.loading') : t('profile.lgpd.deleteConfirmButton') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* =============================================================================
   PROFILE PAGE STYLES
============================================================================= */

.profile-page {
  min-height: 100vh;
  background: var(--color-bg);
}

/* Header */
.profile-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-md) var(--space-xl);
  background: var(--color-bg-card);
  border-bottom: 1px solid var(--color-border-light);
}

.profile-header__left {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.profile-header__back {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  text-decoration: none;
}

.profile-header__back:hover {
  color: var(--color-primary);
}

.profile-header__title {
  font-size: var(--font-size-xl);
  font-weight: 700;
  color: var(--color-text);
  margin: 0;
}

/* Loading */
.profile-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 50vh;
  gap: var(--space-md);
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid var(--color-border);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Main Content */
.profile-main {
  max-width: 600px;
  margin: 0 auto;
  padding: var(--space-xl);
}

/* Profile Card */
.profile-card {
  display: flex;
  align-items: center;
  gap: var(--space-lg);
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  padding: var(--space-xl);
  box-shadow: var(--shadow-sm);
  margin-bottom: var(--space-xl);
}

.profile-card__avatar {
  width: 80px;
  height: 80px;
  background: var(--color-bg-warm);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.avatar-icon {
  font-size: 2.5rem;
}

.profile-card__info {
  flex: 1;
}

.profile-card__name {
  font-size: var(--font-size-xl);
  font-weight: 700;
  color: var(--color-text);
  margin: 0 0 var(--space-xs);
}

.profile-card__email {
  font-size: var(--font-size-base);
  color: var(--color-text-soft);
  margin: 0 0 var(--space-md);
}

.profile-badges {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: var(--space-xs);
  padding: var(--space-xs) var(--space-sm);
  border-radius: var(--radius-full);
  font-size: var(--font-size-sm);
  font-weight: 500;
}

.badge--admin {
  background: #fef3c7;
  color: #92400e;
}

.badge--member {
  background: var(--color-bg-warm);
  color: var(--color-text-soft);
}

/* Section */
.profile-section {
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  box-shadow: var(--shadow-sm);
  margin-bottom: var(--space-lg);
}

.profile-section__title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--color-text);
  margin: 0 0 var(--space-lg);
  padding-bottom: var(--space-sm);
  border-bottom: 1px solid var(--color-border-light);
}

/* Data Grid */
.data-grid {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.data-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-sm) 0;
  border-bottom: 1px dashed var(--color-border-light);
}

.data-item:last-child {
  border-bottom: none;
}

.data-item__label {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
}

.data-item__value {
  font-size: var(--font-size-base);
  color: var(--color-text);
  font-weight: 500;
}

.data-item__value--mono {
  font-family: monospace;
  font-size: var(--font-size-sm);
  background: var(--color-bg-warm);
  padding: var(--space-xs) var(--space-sm);
  border-radius: var(--radius-sm);
}

/* Action Buttons */
.action-buttons {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.btn--danger {
  color: #dc2626;
}

.btn--danger:hover {
  background: #fee2e2;
}

/* Note */
.profile-note {
  display: flex;
  align-items: flex-start;
  gap: var(--space-md);
  background: var(--color-bg-warm);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
}

.profile-note__icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.profile-note__text {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  margin: 0;
  line-height: 1.5;
}

/* =============================================================================
   LGPD SECTION
============================================================================= */

.profile-section--lgpd {
  border: 2px solid var(--color-border-light);
}

.lgpd-description {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  margin-bottom: var(--space-lg);
  line-height: 1.6;
}

.lgpd-actions {
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
}

.lgpd-action {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-md);
  background: var(--color-bg);
  border-radius: var(--radius-md);
  gap: var(--space-md);
}

.lgpd-action--danger {
  background: #fef2f2;
  border: 1px solid #fecaca;
}

.lgpd-action__info {
  display: flex;
  align-items: flex-start;
  gap: var(--space-md);
  flex: 1;
}

.lgpd-action__icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.lgpd-action__title {
  font-size: var(--font-size-base);
  font-weight: 600;
  color: var(--color-text);
  margin: 0 0 var(--space-xs);
}

.lgpd-action__description {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  margin: 0;
}

.btn--danger {
  background: #dc2626;
  color: white;
  border: none;
}

.btn--danger:hover:not(:disabled) {
  background: #b91c1c;
}

.btn--danger:disabled {
  background: #fca5a5;
  cursor: not-allowed;
}

/* =============================================================================
   DELETE MODAL
============================================================================= */

.delete-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.75);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: var(--space-md);
  backdrop-filter: blur(4px);
}

.delete-modal {
  background: #ffffff;
  border-radius: var(--radius-lg);
  max-width: 500px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  border: 1px solid #e5e7eb;
}

.delete-modal__header {
  padding: var(--space-lg);
  border-bottom: 1px solid #e5e7eb;
  background: #fef2f2;
}

.delete-modal__header h3 {
  margin: 0;
  font-size: var(--font-size-lg);
  color: #dc2626;
}

.delete-modal__body {
  padding: var(--space-lg);
  background: #ffffff;
}

.delete-warning {
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: var(--radius-md);
  padding: var(--space-md);
  margin-bottom: var(--space-lg);
}

.delete-warning p {
  margin: 0 0 var(--space-sm);
  color: #991b1b;
  font-size: var(--font-size-sm);
}

.delete-warning ul {
  margin: var(--space-sm) 0;
  padding-left: var(--space-lg);
}

.delete-warning li {
  font-size: var(--font-size-sm);
  color: #7f1d1d;
  margin-bottom: var(--space-xs);
}

.delete-warning__highlight {
  font-weight: 700;
  color: #7f1d1d !important;
}

.delete-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.form-group label {
  font-size: var(--font-size-sm);
  color: var(--color-text);
  font-weight: 500;
}

.form-group label code {
  background: #fee2e2;
  color: #991b1b;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-size: var(--font-size-sm);
}

.form-input {
  padding: var(--space-sm) var(--space-md);
  border: 1px solid #d1d5db;
  border-radius: var(--radius-md);
  font-size: var(--font-size-base);
  font-family: inherit;
  background: #ffffff;
  color: #111827;
  width: 100%;
}

.form-input:focus {
  outline: none;
  border-color: #dc2626;
  box-shadow: 0 0 0 3px rgba(220, 38, 38, 0.15);
}

.delete-error {
  background: #fee2e2;
  color: #dc2626;
  padding: var(--space-sm);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  text-align: center;
}

.delete-modal__footer {
  padding: var(--space-lg);
  border-top: 1px solid #e5e7eb;
  display: flex;
  justify-content: flex-end;
  gap: var(--space-md);
  background: #f9fafb;
}

/* Responsive */
@media (max-width: 768px) {
  .profile-card {
    flex-direction: column;
    text-align: center;
  }
  
  .profile-badges {
    justify-content: center;
  }
  
  .data-item {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-xs);
  }
  
  .lgpd-action {
    flex-direction: column;
    text-align: center;
  }
  
  .lgpd-action__info {
    flex-direction: column;
    align-items: center;
  }
  
  .delete-modal__footer {
    flex-direction: column;
  }
  
  .delete-modal__footer .btn {
    width: 100%;
  }
}
</style>

