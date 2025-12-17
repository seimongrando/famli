<!-- =============================================================================
  FAMLI - Página de Administração
  =============================================================================
  Dashboard administrativo para monitoramento e gestão do sistema.
  
  Funcionalidades:
  - Visão geral com métricas principais
  - Gráficos de distribuição de itens
  - Lista de usuários (dados mascarados)
  - Atividade recente do sistema
  - Status de saúde do servidor
  
  Acesso:
  - Requer autenticação
  - Requer permissão de administrador (email na lista ADMIN_EMAILS)
============================================================================== -->

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { useLocalizedRoutes } from '../composables/useLocalizedRoutes'

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()
const { paths, getPath } = useLocalizedRoutes()

// Estados
const loading = ref(true)
const error = ref(null)
const refreshInterval = ref(null)
const activeTab = ref('overview')

// Dados do dashboard
const dashboard = ref({
  overview: {
    total_users: 0,
    total_items: 0,
    total_guardians: 0,
    avg_items_per_user: 0
  },
  items_by_type: {},
  items_by_category: {},
  recent_signups: 0,
  config: {
    admin_emails: [],
    environment: ''
  }
})

const health = ref({
  status: 'unknown',
  uptime: { human: '...' },
  memory: { alloc_mb: 0 },
  runtime: { goroutines: 0 }
})

const users = ref([])
const activity = ref([])

// Cores para gráficos
const typeColors = {
  info: '#3b82f6',
  memory: '#f59e0b',
  access: '#10b981',
  routine: '#8b5cf6',
  location: '#ec4899',
  contact: '#6366f1'
}

const categoryColors = {
  'família': '#f43f5e',
  'saúde': '#22c55e',
  'finanças': '#eab308',
  'documentos': '#3b82f6',
  'rotina': '#a855f7',
  'acesso': '#06b6d4'
}

// Computed para gráficos
const typeChartData = computed(() => {
  const items = dashboard.value.items_by_type
  const total = Object.values(items).reduce((a, b) => a + b, 0)
  
  return Object.entries(items).map(([type, count]) => ({
    type,
    count,
    percentage: total > 0 ? ((count / total) * 100).toFixed(1) : 0,
    color: typeColors[type] || '#9ca3af'
  }))
})

const categoryChartData = computed(() => {
  const items = dashboard.value.items_by_category
  const total = Object.values(items).reduce((a, b) => a + b, 0)
  
  return Object.entries(items).map(([category, count]) => ({
    category,
    count,
    percentage: total > 0 ? ((count / total) * 100).toFixed(1) : 0,
    color: categoryColors[category] || '#9ca3af'
  }))
})

// Funções de API
async function fetchDashboard() {
  try {
    const response = await fetch('/api/admin/dashboard', {
      credentials: 'include'
    })
    
    console.log('[Admin] Dashboard response status:', response.status)
    
    if (response.status === 401) {
      error.value = t('admin.sessionExpired')
      return
    }
    
    if (response.status === 403) {
      error.value = t('admin.accessDenied')
      return
    }
    
    if (!response.ok) {
      const errorData = await response.text()
      console.error('[Admin] Dashboard error:', errorData)
      throw new Error('Failed to fetch dashboard')
    }
    
    const data = await response.json()
    console.log('[Admin] Dashboard data:', data)
    dashboard.value = data
  } catch (err) {
    console.error('[Admin] Dashboard fetch error:', err)
    error.value = t('admin.loadError')
  }
}

async function fetchHealth() {
  try {
    const response = await fetch('/api/admin/health', {
      credentials: 'include'
    })
    
    if (!response.ok) throw new Error('Failed to fetch health')
    
    health.value = await response.json()
  } catch (err) {
    health.value.status = 'unknown'
  }
}

async function fetchUsers() {
  try {
    const response = await fetch('/api/admin/users', {
      credentials: 'include'
    })
    
    if (!response.ok) throw new Error('Failed to fetch users')
    
    const data = await response.json()
    users.value = data.users || []
  } catch (err) {
    // Silently fail
  }
}

async function fetchActivity() {
  try {
    const response = await fetch('/api/admin/activity', {
      credentials: 'include'
    })
    
    if (!response.ok) throw new Error('Failed to fetch activity')
    
    const data = await response.json()
    activity.value = data.activities || []
  } catch (err) {
    // Silently fail
  }
}

async function loadAll() {
  loading.value = true
  await Promise.all([
    fetchDashboard(),
    fetchHealth(),
    fetchUsers(),
    fetchActivity()
  ])
  loading.value = false
}

// Formatadores
function formatEventType(type) {
  const types = {
    'LOGIN_SUCCESS': '✓ Login',
    'LOGIN_FAILED': '✗ Login falhou',
    'REGISTER': '+ Novo usuário',
    'DATA_CREATE': '+ Item criado',
    'DATA_UPDATE': '↻ Item atualizado',
    'DATA_DELETE': '− Item removido',
    'RATE_LIMIT_EXCEEDED': '⚠ Rate limit',
    'UNAUTHORIZED_ACCESS': '🔒 Acesso negado'
  }
  return types[type] || type
}

function formatTimestamp(timestamp) {
  const date = new Date(timestamp)
  return date.toLocaleString('pt-BR', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// Lifecycle
onMounted(() => {
  if (!authStore.isAuthenticated) {
    router.push(paths.value.auth)
    return
  }
  
  loadAll()
  
  // Auto-refresh a cada 30 segundos
  refreshInterval.value = setInterval(loadAll, 30000)
})

onUnmounted(() => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
  }
})
</script>

<template>
  <div class="admin-page">
    <!-- Header -->
    <header class="admin-header">
      <div class="admin-header__left">
        <router-link :to="paths.dashboard" class="admin-header__back">
          ← {{ t('admin.backToDashboard') }}
        </router-link>
        <h1 class="admin-header__title">{{ t('admin.title') }}</h1>
      </div>
      <div class="admin-header__right">
        <span 
          class="health-badge" 
          :class="`health-badge--${health.status}`"
        >
          {{ health.status === 'healthy' ? '● ' : '○ ' }}
          {{ health.status }}
        </span>
        <button @click="loadAll" class="btn btn--secondary btn--small">
          ↻ {{ t('admin.refresh') }}
        </button>
      </div>
    </header>

    <!-- Loading -->
    <div v-if="loading" class="admin-loading">
      <div class="spinner"></div>
      <p>{{ t('admin.loading') }}</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="admin-error">
      <div class="admin-error__icon">🔒</div>
      <h2>{{ t('admin.error') }}</h2>
      <p>{{ error }}</p>
      <router-link :to="paths.dashboard" class="btn btn--primary">
        {{ t('admin.backToDashboard') }}
      </router-link>
    </div>

    <!-- Content -->
    <main v-else class="admin-main">
      <!-- Tabs -->
      <nav class="admin-tabs">
        <button 
          v-for="tab in ['overview', 'users', 'activity', 'system']" 
          :key="tab"
          :class="['admin-tab', { 'admin-tab--active': activeTab === tab }]"
          @click="activeTab = tab"
        >
          {{ t(`admin.tabs.${tab}`) }}
        </button>
      </nav>

      <!-- Overview Tab -->
      <section v-if="activeTab === 'overview'" class="admin-section">
        <!-- Metric Cards -->
        <div class="metrics-grid">
          <div class="metric-card metric-card--users">
            <div class="metric-card__icon">👥</div>
            <div class="metric-card__content">
              <span class="metric-card__value">{{ dashboard.overview.total_users }}</span>
              <span class="metric-card__label">{{ t('admin.metrics.totalUsers') }}</span>
            </div>
            <div class="metric-card__badge" v-if="dashboard.recent_signups > 0">
              +{{ dashboard.recent_signups }} {{ t('admin.metrics.thisWeek') }}
            </div>
          </div>

          <div class="metric-card metric-card--items">
            <div class="metric-card__icon">📦</div>
            <div class="metric-card__content">
              <span class="metric-card__value">{{ dashboard.overview.total_items }}</span>
              <span class="metric-card__label">{{ t('admin.metrics.totalItems') }}</span>
            </div>
          </div>

          <div class="metric-card metric-card--guardians">
            <div class="metric-card__icon">👼</div>
            <div class="metric-card__content">
              <span class="metric-card__value">{{ dashboard.overview.total_guardians }}</span>
              <span class="metric-card__label">{{ t('admin.metrics.totalGuardians') }}</span>
            </div>
          </div>

          <div class="metric-card metric-card--avg">
            <div class="metric-card__icon">📊</div>
            <div class="metric-card__content">
              <span class="metric-card__value">{{ dashboard.overview.avg_items_per_user?.toFixed(1) || '0' }}</span>
              <span class="metric-card__label">{{ t('admin.metrics.avgItemsPerUser') }}</span>
            </div>
          </div>
        </div>

        <!-- Charts Row -->
        <div class="charts-row">
          <!-- Items by Type -->
          <div class="chart-card">
            <h3 class="chart-card__title">{{ t('admin.charts.itemsByType') }}</h3>
            <div class="bar-chart">
              <div 
                v-for="item in typeChartData" 
                :key="item.type"
                class="bar-item"
              >
                <div class="bar-item__label">{{ item.type }}</div>
                <div class="bar-item__track">
                  <div 
                    class="bar-item__fill" 
                    :style="{ 
                      width: `${item.percentage}%`,
                      backgroundColor: item.color 
                    }"
                  ></div>
                </div>
                <div class="bar-item__value">{{ item.count }} ({{ item.percentage }}%)</div>
              </div>
              <p v-if="typeChartData.length === 0" class="chart-empty">
                {{ t('admin.charts.noData') }}
              </p>
            </div>
          </div>

          <!-- Items by Category -->
          <div class="chart-card">
            <h3 class="chart-card__title">{{ t('admin.charts.itemsByCategory') }}</h3>
            <div class="bar-chart">
              <div 
                v-for="item in categoryChartData" 
                :key="item.category"
                class="bar-item"
              >
                <div class="bar-item__label">{{ item.category }}</div>
                <div class="bar-item__track">
                  <div 
                    class="bar-item__fill" 
                    :style="{ 
                      width: `${item.percentage}%`,
                      backgroundColor: item.color 
                    }"
                  ></div>
                </div>
                <div class="bar-item__value">{{ item.count }} ({{ item.percentage }}%)</div>
              </div>
              <p v-if="categoryChartData.length === 0" class="chart-empty">
                {{ t('admin.charts.noData') }}
              </p>
            </div>
          </div>
        </div>
      </section>

      <!-- Users Tab -->
      <section v-if="activeTab === 'users'" class="admin-section">
        <div class="users-table-wrapper">
          <table class="users-table">
            <thead>
              <tr>
                <th>{{ t('admin.users.name') }}</th>
                <th>{{ t('admin.users.email') }}</th>
                <th>{{ t('admin.users.items') }}</th>
                <th>{{ t('admin.users.guardians') }}</th>
                <th>{{ t('admin.users.createdAt') }}</th>
                <th>{{ t('admin.users.admin') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.id">
                <td>{{ user.name || '-' }}</td>
                <td class="users-table__email">{{ user.email }}</td>
                <td>{{ user.items_count }}</td>
                <td>{{ user.guardians_count }}</td>
                <td>{{ formatTimestamp(user.created_at) }}</td>
                <td>
                  <span v-if="user.is_admin" class="badge badge--admin">Admin</span>
                </td>
              </tr>
              <tr v-if="users.length === 0">
                <td colspan="6" class="users-table__empty">
                  {{ t('admin.users.noUsers') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- Activity Tab -->
      <section v-if="activeTab === 'activity'" class="admin-section">
        <div class="activity-list">
          <div 
            v-for="event in activity" 
            :key="event.id"
            class="activity-item"
            :class="`activity-item--${event.severity.toLowerCase()}`"
          >
            <div class="activity-item__time">
              {{ formatTimestamp(event.timestamp) }}
            </div>
            <div class="activity-item__type">
              {{ formatEventType(event.type) }}
            </div>
            <div class="activity-item__result">
              {{ event.result }}
            </div>
            <div class="activity-item__ip">
              {{ event.client_ip }}
            </div>
          </div>
          <p v-if="activity.length === 0" class="activity-empty">
            {{ t('admin.activity.noActivity') }}
          </p>
        </div>
      </section>

      <!-- System Tab -->
      <section v-if="activeTab === 'system'" class="admin-section">
        <div class="system-grid">
          <!-- Status Card -->
          <div class="system-card">
            <h3 class="system-card__title">{{ t('admin.system.status') }}</h3>
            <div class="system-card__content">
              <div class="system-stat">
                <span class="system-stat__label">Status</span>
                <span 
                  class="system-stat__value"
                  :class="`system-stat__value--${health.status}`"
                >
                  {{ health.status }}
                </span>
              </div>
              <div class="system-stat">
                <span class="system-stat__label">Uptime</span>
                <span class="system-stat__value">{{ health.uptime?.human || '...' }}</span>
              </div>
            </div>
          </div>

          <!-- Memory Card -->
          <div class="system-card">
            <h3 class="system-card__title">{{ t('admin.system.memory') }}</h3>
            <div class="system-card__content">
              <div class="system-stat">
                <span class="system-stat__label">{{ t('admin.system.allocated') }}</span>
                <span class="system-stat__value">{{ health.memory?.alloc_mb?.toFixed(2) || '0' }} MB</span>
              </div>
              <div class="system-stat">
                <span class="system-stat__label">{{ t('admin.system.system') }}</span>
                <span class="system-stat__value">{{ health.memory?.sys_mb?.toFixed(2) || '0' }} MB</span>
              </div>
              <div class="system-stat">
                <span class="system-stat__label">GC Cycles</span>
                <span class="system-stat__value">{{ health.memory?.num_gc || 0 }}</span>
              </div>
            </div>
          </div>

          <!-- Runtime Card -->
          <div class="system-card">
            <h3 class="system-card__title">{{ t('admin.system.runtime') }}</h3>
            <div class="system-card__content">
              <div class="system-stat">
                <span class="system-stat__label">Goroutines</span>
                <span class="system-stat__value">{{ health.runtime?.goroutines || 0 }}</span>
              </div>
              <div class="system-stat">
                <span class="system-stat__label">CPUs</span>
                <span class="system-stat__value">{{ health.runtime?.cpus || 0 }}</span>
              </div>
              <div class="system-stat">
                <span class="system-stat__label">Go Version</span>
                <span class="system-stat__value">{{ health.runtime?.go_version || '-' }}</span>
              </div>
            </div>
          </div>

          <!-- Storage Card -->
          <div class="system-card">
            <h3 class="system-card__title">{{ t('admin.system.storage') }}</h3>
            <div class="system-card__content">
              <div class="system-stat">
                <span class="system-stat__label">{{ t('admin.system.type') }}</span>
                <span class="system-stat__value">{{ health.storage?.type || '-' }}</span>
              </div>
              <div class="system-stat">
                <span class="system-stat__label">Status</span>
                <span class="system-stat__value system-stat__value--healthy">
                  {{ health.storage?.status || '-' }}
                </span>
              </div>
            </div>
          </div>

          <!-- Config Card (Debug) -->
          <div class="system-card system-card--wide">
            <h3 class="system-card__title">{{ t('admin.system.config') }}</h3>
            <div class="system-card__content">
              <div class="system-stat">
                <span class="system-stat__label">{{ t('admin.system.environment') }}</span>
                <span class="system-stat__value">{{ dashboard.config?.environment || 'development' }}</span>
              </div>
              <div class="system-stat system-stat--vertical">
                <span class="system-stat__label">{{ t('admin.system.adminEmails') }}</span>
                <div class="system-stat__list">
                  <span 
                    v-for="email in (dashboard.config?.admin_emails || [])" 
                    :key="email"
                    class="system-stat__tag"
                  >
                    {{ email }}
                  </span>
                  <span v-if="!dashboard.config?.admin_emails?.length" class="system-stat__empty">
                    {{ t('admin.system.noAdminEmails') }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped>
/* =============================================================================
   ADMIN PAGE STYLES
============================================================================= */

.admin-page {
  min-height: 100vh;
  background: var(--color-bg);
}

/* Header */
.admin-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-md) var(--space-xl);
  background: var(--color-bg-card);
  border-bottom: 1px solid var(--color-border-light);
}

.admin-header__left {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.admin-header__back {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  text-decoration: none;
}

.admin-header__back:hover {
  color: var(--color-primary);
}

.admin-header__title {
  font-size: var(--font-size-xl);
  font-weight: 700;
  color: var(--color-text);
  margin: 0;
}

.admin-header__right {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.health-badge {
  padding: var(--space-xs) var(--space-sm);
  border-radius: var(--radius-full);
  font-size: var(--font-size-sm);
  font-weight: 600;
}

.health-badge--healthy {
  background: #dcfce7;
  color: #166534;
}

.health-badge--degraded {
  background: #fef3c7;
  color: #92400e;
}

.health-badge--unknown {
  background: var(--color-bg-warm);
  color: var(--color-text-soft);
}

/* Loading & Error */
.admin-loading,
.admin-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 50vh;
  gap: var(--space-md);
}

.admin-error__icon {
  font-size: 4rem;
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

/* Tabs */
.admin-tabs {
  display: flex;
  gap: var(--space-sm);
  padding: var(--space-md) var(--space-xl);
  background: var(--color-bg-card);
  border-bottom: 1px solid var(--color-border-light);
}

.admin-tab {
  padding: var(--space-sm) var(--space-md);
  border: none;
  background: transparent;
  color: var(--color-text-soft);
  font-size: var(--font-size-base);
  font-weight: 500;
  cursor: pointer;
  border-radius: var(--radius-md);
  transition: all 0.2s ease;
}

.admin-tab:hover {
  background: var(--color-bg-warm);
  color: var(--color-text);
}

.admin-tab--active {
  background: var(--color-primary);
  color: white;
}

/* Main Content */
.admin-main {
  max-width: 1400px;
  margin: 0 auto;
}

.admin-section {
  padding: var(--space-xl);
}

/* Metrics Grid */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: var(--space-lg);
  margin-bottom: var(--space-xl);
}

.metric-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  display: flex;
  align-items: center;
  gap: var(--space-md);
  box-shadow: var(--shadow-sm);
  position: relative;
  overflow: hidden;
}

.metric-card__icon {
  font-size: 2.5rem;
  flex-shrink: 0;
}

.metric-card__content {
  display: flex;
  flex-direction: column;
}

.metric-card__value {
  font-size: var(--font-size-2xl);
  font-weight: 700;
  color: var(--color-text);
  line-height: 1;
}

.metric-card__label {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  margin-top: var(--space-xs);
}

.metric-card__badge {
  position: absolute;
  top: var(--space-sm);
  right: var(--space-sm);
  background: #dcfce7;
  color: #166534;
  font-size: var(--font-size-xs);
  padding: 2px 8px;
  border-radius: var(--radius-full);
  font-weight: 600;
}

/* Charts */
.charts-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: var(--space-lg);
}

.chart-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  box-shadow: var(--shadow-sm);
}

.chart-card__title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: var(--space-lg);
}

.bar-chart {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.bar-item {
  display: grid;
  grid-template-columns: 100px 1fr 80px;
  align-items: center;
  gap: var(--space-md);
}

.bar-item__label {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  text-transform: capitalize;
}

.bar-item__track {
  height: 24px;
  background: var(--color-bg-warm);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.bar-item__fill {
  height: 100%;
  border-radius: var(--radius-sm);
  transition: width 0.5s ease;
}

.bar-item__value {
  font-size: var(--font-size-sm);
  color: var(--color-text);
  font-weight: 500;
  text-align: right;
}

.chart-empty {
  text-align: center;
  color: var(--color-text-soft);
  padding: var(--space-xl);
}

/* Users Table */
.users-table-wrapper {
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

.users-table {
  width: 100%;
  border-collapse: collapse;
}

.users-table th,
.users-table td {
  padding: var(--space-md);
  text-align: left;
  border-bottom: 1px solid var(--color-border-light);
}

.users-table th {
  background: var(--color-bg-warm);
  font-weight: 600;
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.users-table td {
  font-size: var(--font-size-base);
  color: var(--color-text);
}

.users-table__email {
  font-family: monospace;
  font-size: var(--font-size-sm);
}

.users-table__empty {
  text-align: center;
  color: var(--color-text-soft);
  padding: var(--space-xl) !important;
}

.badge {
  padding: 2px 8px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-xs);
  font-weight: 600;
}

.badge--admin {
  background: #fef3c7;
  color: #92400e;
}

/* Activity List */
.activity-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  padding: var(--space-md);
  box-shadow: var(--shadow-sm);
}

.activity-item {
  display: grid;
  grid-template-columns: 140px 200px 100px 1fr;
  gap: var(--space-md);
  padding: var(--space-md);
  border-radius: var(--radius-md);
  background: var(--color-bg);
  font-size: var(--font-size-sm);
}

.activity-item--warning {
  background: #fef3c7;
}

.activity-item--error {
  background: #fee2e2;
}

.activity-item--critical {
  background: #fecaca;
}

.activity-item__time {
  color: var(--color-text-soft);
  font-family: monospace;
}

.activity-item__type {
  color: var(--color-text);
  font-weight: 500;
}

.activity-item__result {
  color: var(--color-text-soft);
}

.activity-item__ip {
  color: var(--color-text-soft);
  font-family: monospace;
  text-align: right;
}

.activity-empty {
  text-align: center;
  color: var(--color-text-soft);
  padding: var(--space-xl);
}

/* System Grid */
.system-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: var(--space-lg);
}

.system-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  box-shadow: var(--shadow-sm);
}

.system-card__title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: var(--space-md);
  padding-bottom: var(--space-sm);
  border-bottom: 1px solid var(--color-border-light);
}

.system-card__content {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.system-stat {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.system-stat__label {
  color: var(--color-text-soft);
  font-size: var(--font-size-sm);
}

.system-stat__value {
  font-weight: 600;
  color: var(--color-text);
}

.system-stat__value--healthy {
  color: #16a34a;
}

.system-stat__value--degraded {
  color: #f59e0b;
}

.system-card--wide {
  grid-column: 1 / -1;
}

.system-stat--vertical {
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-sm);
}

.system-stat__list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-xs);
}

.system-stat__tag {
  background: var(--color-bg-warm);
  color: var(--color-text);
  padding: var(--space-xs) var(--space-sm);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-sm);
  font-family: monospace;
}

.system-stat__empty {
  color: var(--color-text-soft);
  font-style: italic;
  font-size: var(--font-size-sm);
}

/* Responsive */
@media (max-width: 768px) {
  .admin-header {
    flex-direction: column;
    gap: var(--space-md);
    align-items: flex-start;
  }

  .admin-tabs {
    overflow-x: auto;
    padding: var(--space-sm);
  }

  .admin-section {
    padding: var(--space-md);
  }

  .bar-item {
    grid-template-columns: 80px 1fr 60px;
  }

  .activity-item {
    grid-template-columns: 1fr;
    gap: var(--space-xs);
  }

  .charts-row {
    grid-template-columns: 1fr;
  }
}
</style>

