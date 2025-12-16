<!-- =============================================================================
FAMLI - Página de Autenticação
===============================================================================
Esta página gerencia login e registro de usuários.

Funcionalidades:
- Alternar entre login e registro
- Validação de senha em tempo real
- Feedback visual de força da senha
- Tradução completa (i18n)

Segurança:
- Validação de email
- Requisitos de senha claros
- Feedback de erros sem expor detalhes sensíveis
============================================================================= -->

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import LanguageSelector from '../components/LanguageSelector.vue'
import PasswordStrength from '../components/PasswordStrength.vue'

// =============================================================================
// SETUP
// =============================================================================

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// =============================================================================
// ESTADO
// =============================================================================

// Modo atual: 'login' ou 'register'
const mode = ref('login')

// Dados do formulário
const form = ref({
  name: '',
  email: '',
  password: ''
})

// Se a senha é válida (atende todos os requisitos)
const isPasswordValid = ref(false)

// =============================================================================
// COMPUTED
// =============================================================================

// Se está no modo login
const isLogin = computed(() => mode.value === 'login')

// Se o formulário pode ser enviado
const canSubmit = computed(() => {
  if (isLogin.value) {
    // No login, apenas email e senha são necessários
    return form.value.email && form.value.password
  } else {
    // No registro, precisa validar a força da senha
    return form.value.email && form.value.password && isPasswordValid.value
  }
})

// =============================================================================
// LIFECYCLE
// =============================================================================

onMounted(() => {
  // Definir modo baseado na query string
  mode.value = route.query.mode === 'register' ? 'register' : 'login'
  
  // Se já autenticado, redirecionar para dashboard
  if (authStore.isAuthenticated) {
    router.push({ name: 'dashboard' })
  }
})

// =============================================================================
// MÉTODOS
// =============================================================================

// Alternar entre login e registro
function toggleMode() {
  mode.value = isLogin.value ? 'register' : 'login'
  authStore.clearError()
}

// Callback quando a validação de senha muda
function onPasswordValidChange(valid) {
  isPasswordValid.value = valid
}

// Submeter formulário
async function handleSubmit() {
  // Verificar se pode submeter
  if (!canSubmit.value) return
  
  let success = false
  
  if (isLogin.value) {
    success = await authStore.login(form.value.email, form.value.password)
  } else {
    success = await authStore.register(form.value.email, form.value.password, form.value.name)
  }
  
  if (success) {
    router.push({ name: 'dashboard' })
  }
}
</script>

<template>
  <div class="auth-page">
    <!-- Header simples -->
    <header class="auth-header">
      <router-link :to="{ name: 'landing' }" class="header__brand">
        <img src="/famli.png" alt="Famli" class="header__logo" />
        <span class="header__name">{{ t('brand.name') }}</span>
      </router-link>
      <LanguageSelector />
    </header>

    <main class="auth-main">
      <div class="auth-card">
        <!-- Tabs -->
        <div class="auth-tabs">
          <button 
            :class="['auth-tab', { 'auth-tab--active': isLogin }]"
            @click="mode = 'login'; authStore.clearError()"
          >
            {{ t('nav.login') }}
          </button>
          <button 
            :class="['auth-tab', { 'auth-tab--active': !isLogin }]"
            @click="mode = 'register'; authStore.clearError()"
          >
            {{ t('nav.register') }}
          </button>
        </div>

        <!-- Título -->
        <div class="auth-header-content">
          <h1 class="auth-title">
            {{ isLogin ? t('auth.loginTitle') : t('auth.registerTitle') }}
          </h1>
          <p class="auth-subtitle">
            {{ isLogin ? t('auth.loginSubtitle') : t('auth.registerSubtitle') }}
          </p>
        </div>

        <!-- Form -->
        <form @submit.prevent="handleSubmit" class="auth-form">
          <!-- Nome (apenas registro) -->
          <div v-if="!isLogin" class="form-group">
            <label class="form-label" for="name">{{ t('auth.name') }}</label>
            <input
              id="name"
              v-model="form.name"
              type="text"
              class="form-input"
              :placeholder="t('auth.namePlaceholder')"
              autocomplete="name"
            />
          </div>

          <!-- Email -->
          <div class="form-group">
            <label class="form-label" for="email">{{ t('auth.email') }}</label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              class="form-input"
              :placeholder="t('auth.emailPlaceholder')"
              required
              autocomplete="email"
            />
            <p v-if="!isLogin" class="form-hint">
              {{ t('auth.emailHint') }}
            </p>
          </div>

          <!-- Senha -->
          <div class="form-group">
            <label class="form-label" for="password">{{ t('auth.password') }}</label>
            <input
              id="password"
              v-model="form.password"
              type="password"
              class="form-input"
              :placeholder="t('auth.passwordPlaceholder')"
              required
              :minlength="isLogin ? 1 : 8"
              :autocomplete="isLogin ? 'current-password' : 'new-password'"
            />
            
            <!-- Componente de validação de senha (apenas registro) -->
            <PasswordStrength 
              v-if="!isLogin"
              :password="form.password"
              :show="true"
              @valid="onPasswordValidChange"
            />
          </div>

          <!-- Erro -->
          <p v-if="authStore.error" class="form-error">
            {{ authStore.error }}
          </p>

          <!-- Aviso de senha inválida -->
          <p v-if="!isLogin && !isPasswordValid && form.password.length > 0" class="form-warning">
            {{ t('auth.passwordInvalid') }}
          </p>

          <!-- Botão -->
          <button 
            type="submit" 
            class="btn btn--primary btn--large btn--full"
            :class="{ 'btn--disabled-hint': !isLogin && !canSubmit && form.password.length > 0 }"
            :disabled="authStore.loading || (!isLogin && !canSubmit)"
          >
            {{ authStore.loading 
              ? (isLogin ? t('auth.loginLoading') : t('auth.registerLoading')) 
              : (isLogin ? t('auth.loginButton') : t('auth.registerButton')) 
            }}
          </button>
        </form>

        <!-- Alternar modo -->
        <p class="auth-switch">
          {{ isLogin ? t('auth.noAccount') : t('auth.hasAccount') }}
          <button class="btn btn--link" @click="toggleMode">
            {{ isLogin ? t('auth.createNow') : t('nav.login') }}
          </button>
        </p>

        <!-- Info de segurança -->
        <div class="auth-security">
          <span class="auth-security__icon">🔒</span>
          <p class="auth-security__text">
            {{ t('auth.securityNote') }}
          </p>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
/* =============================================================================
   PÁGINA
   ============================================================================= */

.auth-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, var(--color-bg) 0%, var(--color-bg-warm) 100%);
}

/* =============================================================================
   HEADER
   ============================================================================= */

.auth-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-lg);
}

/* =============================================================================
   MAIN
   ============================================================================= */

.auth-main {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-lg);
}

/* =============================================================================
   CARD
   ============================================================================= */

.auth-card {
  background: var(--color-card);
  border-radius: var(--radius-xl);
  padding: var(--space-2xl);
  width: 100%;
  max-width: 440px;
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--color-border-light);
}

/* =============================================================================
   TABS
   ============================================================================= */

.auth-tabs {
  display: flex;
  gap: var(--space-sm);
  margin-bottom: var(--space-xl);
}

.auth-tab {
  flex: 1;
  padding: var(--space-md);
  background: var(--color-bg-warm);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  font-family: var(--font-family);
  font-size: var(--font-size-base);
  font-weight: 600;
  color: var(--color-text-soft);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.auth-tab:hover {
  background: var(--color-border);
}

.auth-tab--active {
  background: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}

.auth-tab--active:hover {
  background: var(--color-primary-light);
}

/* =============================================================================
   HEADER CONTENT
   ============================================================================= */

.auth-header-content {
  text-align: center;
  margin-bottom: var(--space-xl);
}

.auth-title {
  font-size: var(--font-size-xl);
  margin-bottom: var(--space-sm);
}

.auth-subtitle {
  color: var(--color-text-soft);
  margin: 0;
}

/* =============================================================================
   FORM
   ============================================================================= */

.auth-form {
  margin-bottom: var(--space-lg);
}

/* =============================================================================
   SWITCH
   ============================================================================= */

.auth-switch {
  text-align: center;
  color: var(--color-text-soft);
  margin-bottom: var(--space-lg);
}

/* =============================================================================
   SECURITY NOTE
   ============================================================================= */

.auth-security {
  display: flex;
  align-items: flex-start;
  gap: var(--space-sm);
  padding: var(--space-md);
  background: var(--color-primary-soft);
  border-radius: var(--radius-md);
}

.auth-security__icon {
  font-size: 1.25rem;
}

.auth-security__text {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  margin: 0;
}
</style>
