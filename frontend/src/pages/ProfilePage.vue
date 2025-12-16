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

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()
const { paths } = useLocalizedRoutes()

// Estado local
const loading = ref(true)

// Computed
const user = computed(() => authStore.user)
const isAdmin = computed(() => user.value?.is_admin || false)

const formattedDate = computed(() => {
  if (!user.value?.created_at) return '-'
  const date = new Date(user.value.created_at)
  return date.toLocaleDateString('pt-BR', {
    day: '2-digit',
    month: 'long',
    year: 'numeric'
  })
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
  try {
    await authStore.logout()
  } catch (err) {
    console.error('[Profile] Logout error:', err)
  }
  // Sempre redirecionar, mesmo se o logout falhar no servidor
  router.push(paths.value.landing)
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

      <!-- Security Note -->
      <div class="profile-note">
        <span class="profile-note__icon">🔒</span>
        <p class="profile-note__text">
          {{ t('profile.securityNote') }}
        </p>
      </div>
    </main>
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
}
</style>

