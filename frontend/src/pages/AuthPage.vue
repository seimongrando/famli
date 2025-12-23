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
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { useLocalizedRoutes } from '../composables/useLocalizedRoutes'
import LanguageSelector from '../components/LanguageSelector.vue'
import PasswordStrength from '../components/PasswordStrength.vue'

// =============================================================================
// SETUP
// =============================================================================

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { paths } = useLocalizedRoutes()

// =============================================================================
// SOCIAL LOGIN
// =============================================================================

const oauthStatus = ref({ google: { enabled: false }, apple: { enabled: false } })
const socialLoginLoading = ref(false)

// Verificar status do OAuth ao montar
async function checkOAuthStatus() {
  try {
    const res = await fetch('/api/auth/oauth/status')
    if (res.ok) {
      oauthStatus.value = await res.json()
    }
  } catch (e) {
    // OAuth não disponível
  }
}

// Inicializar Google Sign-In
function initGoogleSignIn() {
  if (!oauthStatus.value.google.enabled) return
  
  // Carregar script do Google se não existir
  if (!window.google && !document.getElementById('google-gsi-script')) {
    const script = document.createElement('script')
    script.id = 'google-gsi-script'
    script.src = 'https://accounts.google.com/gsi/client'
    script.async = true
    script.defer = true
    script.onload = renderGoogleButton
    document.head.appendChild(script)
  } else if (window.google) {
    renderGoogleButton()
  }
}

function renderGoogleButton() {
  if (!window.google || !oauthStatus.value.google.client_id) return
  
  window.google.accounts.id.initialize({
    client_id: oauthStatus.value.google.client_id,
    callback: handleGoogleCallback,
    auto_select: false,
    cancel_on_tap_outside: true
  })
  
  // Renderizar botão customizado
  const container = document.getElementById('google-signin-btn')
  if (container) {
    window.google.accounts.id.renderButton(container, {
      theme: 'outline',
      size: 'large',
      width: '100%',
      text: 'continue_with',
      shape: 'rectangular',
      logo_alignment: 'center'
    })
  }
}

async function handleGoogleCallback(response) {
  if (response.credential) {
    socialLoginLoading.value = true
    const success = await authStore.loginWithGoogle(response.credential)
    socialLoginLoading.value = false
    if (success) {
      router.push(paths.value.dashboard)
    }
  }
}

// Apple Sign In (usando API nativa do browser)
async function handleAppleLogin() {
  if (!window.AppleID) {
    // Carregar script da Apple se necessário
    const script = document.createElement('script')
    script.src = 'https://appleid.cdn-apple.com/appleauth/static/jsapi/appleid/1/en_US/appleid.auth.js'
    script.onload = () => performAppleLogin()
    document.head.appendChild(script)
  } else {
    performAppleLogin()
  }
}

async function performAppleLogin() {
  try {
    socialLoginLoading.value = true
    
    window.AppleID.auth.init({
      clientId: oauthStatus.value.apple.client_id,
      scope: 'name email',
      redirectURI: window.location.origin + '/auth',
      usePopup: true
    })
    
    const response = await window.AppleID.auth.signIn()
    
    if (response.authorization?.id_token) {
      const success = await authStore.loginWithApple(response.authorization.id_token)
      if (success) {
        router.push(paths.value.dashboard)
      }
    }
  } catch (e) {
    if (e.error !== 'popup_closed_by_user') {
      authStore.error = t('auth.socialLoginError')
    }
  } finally {
    socialLoginLoading.value = false
  }
}

// =============================================================================
// ESTADO
// =============================================================================

// Modo atual: 'login' ou 'register'
const mode = ref('login')

// Dados do formulário
const form = ref({
  name: '',
  email: '',
  password: '',
  termsAccepted: false
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
    // No registro, precisa validar a força da senha e aceite dos termos
    return form.value.email && form.value.password && isPasswordValid.value && form.value.termsAccepted
  }
})

// =============================================================================
// LIFECYCLE
// =============================================================================

onMounted(async () => {
  // Definir modo baseado na query string
  mode.value = route.query.mode === 'register' ? 'register' : 'login'
  
  // Se já autenticado, redirecionar para dashboard
  if (authStore.isAuthenticated) {
    router.push(paths.value.dashboard)
    return
  }

  // Verificar status do OAuth
  await checkOAuthStatus()
  
  // Inicializar Google Sign-In se disponível
  if (oauthStatus.value.google.enabled) {
    initGoogleSignIn()
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
    router.push(paths.value.dashboard)
  }
}
</script>

<template>
  <div class="auth-page">
    <!-- Header simples -->
    <header class="auth-header">
      <router-link :to="paths.landing" class="header__brand">
        <img src="/logo.svg" alt="Famli" class="header__logo" />
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
              maxlength="100"
              autocomplete="name"
            />
            <p class="form-hint">{{ t('common.maxChars', { count: 100 }) }}</p>
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
              maxlength="254"
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
              maxlength="128"
              :autocomplete="isLogin ? 'current-password' : 'new-password'"
            />
            
            <!-- Componente de validação de senha (apenas registro) -->
            <PasswordStrength 
              v-if="!isLogin"
              :password="form.password"
              :show="true"
              @valid="onPasswordValidChange"
            />

            <!-- Link para recuperação de senha (apenas login) -->
            <router-link 
              v-if="isLogin" 
              to="/redefinir-senha" 
              class="forgot-password-link"
            >
              {{ t('auth.forgotPassword') }}
            </router-link>
          </div>

          <!-- Aceite dos termos (apenas registro) -->
          <div v-if="!isLogin" class="form-group form-group--checkbox">
            <label class="checkbox-label">
              <input
                v-model="form.termsAccepted"
                type="checkbox"
                class="checkbox-input"
              />
              <span class="checkbox-text">
                {{ t('legal.termsAndPrivacy') }}
                <router-link :to="paths.terms" target="_blank" class="link">
                  {{ t('legal.terms.link') }}
                </router-link>
                {{ t('common.and') }}
                <router-link :to="paths.privacy" target="_blank" class="link">
                  {{ t('legal.privacy.link') }}
                </router-link>
              </span>
            </label>
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

        <!-- Login Social -->
        <div v-if="oauthStatus.google.enabled || oauthStatus.apple.enabled" class="social-login">
          <div class="social-divider">
            <span>{{ t('auth.orContinueWith') }}</span>
          </div>

          <div class="social-buttons">
            <!-- Google Sign-In -->
            <div v-if="oauthStatus.google.enabled" id="google-signin-btn" class="social-btn-container"></div>

            <!-- Apple Sign-In -->
            <button 
              v-if="oauthStatus.apple.enabled"
              @click="handleAppleLogin"
              class="social-btn social-btn--apple"
              :disabled="socialLoginLoading"
            >
              <svg class="social-btn__icon" viewBox="0 0 24 24" fill="currentColor">
                <path d="M17.05 20.28c-.98.95-2.05.8-3.08.35-1.09-.46-2.09-.48-3.24 0-1.44.62-2.2.44-3.06-.35C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.53 4.09zM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25z"/>
              </svg>
              <span>{{ t('auth.continueWithApple') }}</span>
            </button>
          </div>
        </div>

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

/* =============================================================================
   CHECKBOX TERMOS
   ============================================================================= */

.form-group--checkbox {
  margin-top: var(--space-lg);
}

.checkbox-label {
  display: flex;
  align-items: flex-start;
  gap: var(--space-sm);
  cursor: pointer;
}

.checkbox-input {
  margin-top: 0.25rem;
  width: 18px;
  height: 18px;
  accent-color: var(--color-primary);
  cursor: pointer;
}

.checkbox-text {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  line-height: 1.5;
}

.checkbox-text .link {
  color: var(--color-primary);
  text-decoration: underline;
}

.checkbox-text .link:hover {
  color: var(--color-primary-light);
}

/* =============================================================================
   SOCIAL LOGIN
   ============================================================================= */

.social-login {
  margin-top: var(--space-lg);
}

.social-divider {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  margin-bottom: var(--space-lg);
}

.social-divider::before,
.social-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--color-border);
}

.social-divider span {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  white-space: nowrap;
}

.social-buttons {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.social-btn-container {
  width: 100%;
  display: flex;
  justify-content: center;
}

.social-btn-container > div {
  width: 100% !important;
}

.social-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-sm);
  width: 100%;
  padding: var(--space-md);
  border-radius: var(--radius-md);
  font-family: var(--font-family);
  font-size: var(--font-size-base);
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition-fast);
  border: 1px solid var(--color-border);
}

.social-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.social-btn__icon {
  width: 20px;
  height: 20px;
}

.social-btn--apple {
  background: #000;
  color: #fff;
  border-color: #000;
}

.social-btn--apple:hover:not(:disabled) {
  background: #333;
}

.social-btn--google {
  background: #fff;
  color: #333;
}

.social-btn--google:hover:not(:disabled) {
  background: var(--color-bg-warm);
}

/* Forgot Password Link */
.forgot-password-link {
  display: block;
  text-align: right;
  margin-top: 0.5rem;
  font-size: 0.875rem;
  color: var(--color-primary);
  text-decoration: none;
}

.forgot-password-link:hover {
  text-decoration: underline;
}
</style>
