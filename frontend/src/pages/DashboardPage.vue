<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { useBoxStore } from '../stores/box'
import { useGuideStore } from '../stores/guide'

// Components
import LanguageSelector from '../components/LanguageSelector.vue'
import GuideCard from '../components/GuideCard.vue'
import BoxComposer from '../components/BoxComposer.vue'
import BoxFeed from '../components/BoxFeed.vue'
import AssistantChat from '../components/AssistantChat.vue'
import SettingsModal from '../components/SettingsModal.vue'
import PrivacyModal from '../components/PrivacyModal.vue'

const { t } = useI18n()
const authStore = useAuthStore()
const boxStore = useBoxStore()
const guideStore = useGuideStore()

// UI State
const activeTab = ref('caixa') // 'caixa' | 'guia'
const showSettings = ref(false)
const showPrivacy = ref(false)

// Computed
const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 12) return t('dashboard.greeting.morning')
  if (hour < 18) return t('dashboard.greeting.afternoon')
  return t('dashboard.greeting.evening')
})

const userName = computed(() => {
  return authStore.user?.name || ''
})

// Lifecycle
onMounted(async () => {
  await Promise.all([
    boxStore.fetchAll(),
    guideStore.fetchAll()
  ])
})
</script>

<template>
  <div class="dashboard">
    <!-- Header -->
    <header class="dashboard-header">
      <div class="container">
        <div class="dashboard-header__content">
          <div class="header__brand">
            <img src="/famli.png" alt="Famli" class="header__logo" />
            <span class="header__name">{{ t('brand.name') }}</span>
          </div>
          
          <div class="dashboard-header__center">
            <span class="dashboard-tagline">{{ t('brand.tagline') }}</span>
          </div>
          
          <div class="dashboard-header__actions">
            <LanguageSelector />
            <router-link 
              v-if="authStore.user?.is_admin" 
              :to="{ name: 'admin' }" 
              class="btn btn--ghost btn--small admin-link" 
              title="Admin"
            >
              ⚙️
            </router-link>
            <router-link :to="{ name: 'profile' }" class="btn btn--ghost btn--small" title="Perfil">
              👤
            </router-link>
            <button class="btn btn--ghost btn--small" @click="showSettings = true">
              {{ t('nav.settings') }}
            </button>
            <button class="btn btn--ghost btn--small" @click="authStore.logout">
              {{ t('nav.logout') }}
            </button>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="dashboard-main">
      <div class="container">
        <!-- Welcome Section -->
        <section class="welcome-section">
          <div class="welcome-content">
            <h1 class="welcome-title">{{ greeting }}{{ userName ? `, ${userName}` : '' }}! 👋</h1>
            <p class="welcome-subtitle">
              {{ t('dashboard.welcome', { count: boxStore.counts.total }) }}
            </p>
          </div>
          
          <!-- Quick Stats -->
          <div class="quick-stats">
            <div class="stat-item">
              <span class="stat-icon">📋</span>
              <span class="stat-value">{{ boxStore.counts.infos }}</span>
              <span class="stat-label">{{ t('dashboard.stats.info') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-icon">👥</span>
              <span class="stat-value">{{ boxStore.counts.people }}</span>
              <span class="stat-label">{{ t('dashboard.stats.people') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-icon">💝</span>
              <span class="stat-value">{{ boxStore.counts.memories }}</span>
              <span class="stat-label">{{ t('dashboard.stats.memories') }}</span>
            </div>
          </div>
        </section>

        <!-- Tabs -->
        <div class="dashboard-tabs">
          <button 
            :class="['dashboard-tab', { 'dashboard-tab--active': activeTab === 'caixa' }]"
            @click="activeTab = 'caixa'"
          >
            <span class="dashboard-tab__icon">📦</span>
            {{ t('dashboard.tabs.box') }}
          </button>
          <button 
            :class="['dashboard-tab', { 'dashboard-tab--active': activeTab === 'guia' }]"
            @click="activeTab = 'guia'"
          >
            <span class="dashboard-tab__icon">🗺️</span>
            {{ t('dashboard.tabs.guide') }}
            <span v-if="guideStore.progressPercentage < 100" class="dashboard-tab__badge">
              {{ guideStore.progressPercentage }}%
            </span>
          </button>
        </div>

        <!-- Tab Content -->
        <div class="dashboard-content">
          <!-- Caixa Famli Tab -->
          <div v-if="activeTab === 'caixa'" class="caixa-layout">
            <div class="caixa-main">
              <!-- Composer -->
              <BoxComposer />
              
              <!-- Feed -->
              <BoxFeed />
            </div>
            
            <aside class="caixa-sidebar">
              <!-- Assistente -->
              <AssistantChat />
              
              <!-- Info Cards -->
              <div class="sidebar-card">
                <h3 class="sidebar-card__title">💡 {{ t('sidebar.tip.title') }}</h3>
                <p class="sidebar-card__text">
                  {{ t('sidebar.tip.text') }}
                </p>
              </div>
              
              <div class="sidebar-card">
                <h3 class="sidebar-card__title">🔒 {{ t('sidebar.privacy.title') }}</h3>
                <p class="sidebar-card__text">
                  {{ t('sidebar.privacy.text') }}
                </p>
                <button class="btn btn--link btn--small" @click="showPrivacy = true">
                  {{ t('sidebar.privacy.link') }}
                </button>
              </div>
            </aside>
          </div>

          <!-- Guia Famli Tab -->
          <div v-else class="guia-layout">
            <div class="guia-header">
              <h2 class="guia-title">{{ t('guide.title') }}</h2>
              <p class="guia-subtitle">
                {{ t('guide.subtitle') }}
              </p>
              
              <!-- Progress Bar -->
              <div class="progress-bar">
                <div 
                  class="progress-bar__fill" 
                  :style="{ width: guideStore.progressPercentage + '%' }"
                ></div>
              </div>
              <p class="progress-text">
                {{ t('guide.progress', { completed: guideStore.completedCount, total: guideStore.cards.length }) }}
              </p>
            </div>

            <div class="guia-cards">
              <GuideCard 
                v-for="card in guideStore.cards" 
                :key="card.id"
                :card="card"
                :status="guideStore.getCardStatus(card.id)"
                @start="guideStore.markProgress(card.id, 'started'); activeTab = 'caixa'"
                @complete="guideStore.markProgress(card.id, 'completed')"
                @skip="guideStore.markProgress(card.id, 'skipped')"
              />
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Modals -->
    <SettingsModal v-if="showSettings" @close="showSettings = false" />
    <PrivacyModal v-if="showPrivacy" @close="showPrivacy = false" />
  </div>
</template>

<style scoped>
.dashboard {
  min-height: 100vh;
  background: var(--color-bg);
}

/* Header */
.dashboard-header {
  background: var(--color-card);
  border-bottom: 1px solid var(--color-border-light);
  position: sticky;
  top: 0;
  z-index: 40;
}

.dashboard-header__content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-md) 0;
}

.dashboard-header__center {
  flex: 1;
  text-align: center;
}

.dashboard-tagline {
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
}

.dashboard-header__actions {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

/* Main */
.dashboard-main {
  padding: var(--space-xl) 0 var(--space-3xl);
}

/* Welcome */
.welcome-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-xl);
  padding: var(--space-xl);
  background: var(--color-card);
  border-radius: var(--radius-xl);
  border: 1px solid var(--color-border-light);
  box-shadow: var(--shadow-sm);
}

.welcome-title {
  font-size: var(--font-size-2xl);
  margin-bottom: var(--space-sm);
}

.welcome-subtitle {
  color: var(--color-text-soft);
  margin: 0;
}

.quick-stats {
  display: flex;
  gap: var(--space-xl);
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-xs);
}

.stat-icon {
  font-size: 1.5rem;
}

.stat-value {
  font-size: var(--font-size-xl);
  font-weight: 700;
  color: var(--color-primary);
}

.stat-label {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

/* Tabs */
.dashboard-tabs {
  display: flex;
  gap: var(--space-sm);
  margin-bottom: var(--space-lg);
}

.dashboard-tab {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-md) var(--space-lg);
  background: var(--color-card);
  border: 2px solid var(--color-border);
  border-radius: var(--radius-full);
  font-family: var(--font-family);
  font-size: var(--font-size-base);
  font-weight: 600;
  color: var(--color-text-soft);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.dashboard-tab:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.dashboard-tab--active {
  background: var(--color-primary);
  border-color: var(--color-primary);
  color: white;
}

.dashboard-tab--active:hover {
  background: var(--color-primary-light);
  border-color: var(--color-primary-light);
  color: white;
}

.dashboard-tab__icon {
  font-size: 1.25rem;
}

.dashboard-tab__badge {
  padding: 2px 8px;
  background: var(--color-accent);
  border-radius: var(--radius-full);
  font-size: 0.75rem;
  color: white;
}

/* Caixa Layout */
.caixa-layout {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: var(--space-xl);
}

.caixa-main {
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
}

.caixa-sidebar {
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
}

.sidebar-card {
  background: var(--color-card);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  border: 1px solid var(--color-border-light);
}

.sidebar-card__title {
  font-size: var(--font-size-base);
  margin-bottom: var(--space-sm);
}

.sidebar-card__text {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  margin-bottom: var(--space-sm);
}

/* Guia Layout */
.guia-layout {
  max-width: 800px;
}

.guia-header {
  margin-bottom: var(--space-xl);
}

.guia-title {
  font-size: var(--font-size-2xl);
  margin-bottom: var(--space-sm);
}

.guia-subtitle {
  color: var(--color-text-soft);
  margin-bottom: var(--space-lg);
}

.progress-bar {
  height: 8px;
  background: var(--color-border);
  border-radius: var(--radius-full);
  overflow: hidden;
  margin-bottom: var(--space-sm);
}

.progress-bar__fill {
  height: 100%;
  background: linear-gradient(90deg, var(--color-primary), var(--color-accent));
  border-radius: var(--radius-full);
  transition: width 0.5s ease;
}

.progress-text {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
  margin: 0;
}

.guia-cards {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

/* Admin Link */
.admin-link {
  font-size: 1.25rem;
  padding: var(--space-xs) var(--space-sm);
  opacity: 0.7;
  transition: opacity 0.2s ease;
}

.admin-link:hover {
  opacity: 1;
}

/* Responsive */
@media (max-width: 1024px) {
  .caixa-layout {
    grid-template-columns: 1fr;
  }
  
  .caixa-sidebar {
    order: -1;
  }
}

@media (max-width: 768px) {
  .welcome-section {
    flex-direction: column;
    text-align: center;
    gap: var(--space-lg);
  }
  
  .dashboard-tabs {
    flex-direction: column;
  }
  
  .dashboard-tab {
    justify-content: center;
  }
  
  .dashboard-header__center {
    display: none;
  }
}
</style>
