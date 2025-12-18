// =============================================================================
// FAMLI - Store de Autenticação
// =============================================================================
// Gerencia estado de autenticação do usuário.
// 
// Funcionalidades:
// - Login/Logout/Registro
// - Verificação de sessão com retry
// - Mensagens de erro traduzidas
// - Tratamento de sessão expirada
// =============================================================================

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import i18n from '../i18n'

// Mapeamento de erros do backend para chaves i18n
const errorMap = {
  'E-mail ou senha incorretos.': 'auth.errors.invalidCredentials',
  'Dados inválidos.': 'auth.errors.invalidData',
  'E-mail inválido.': 'auth.errors.invalidEmail',
  'Senha precisa ter no mínimo 8 caracteres com letras e números.': 'auth.errors.weakPassword',
  'E-mail já cadastrado.': 'auth.errors.emailExists',
  'Sessão inválida.': 'auth.errors.sessionInvalid',
  'Sessão inválida': 'auth.errors.sessionInvalid',
  'Sessão não encontrada': 'auth.errors.sessionNotFound',
  'Sessão expirada': 'auth.errors.sessionExpired',
  'Muitas tentativas. Aguarde alguns minutos.': 'auth.errors.tooManyAttempts',
  'Erro ao preparar sua conta.': 'auth.errors.serverError'
}

// Códigos de erro do backend
const SESSION_ERROR_CODES = ['SESSION_NOT_FOUND', 'SESSION_INVALID', 'SESSION_EXPIRED']

// Função para traduzir erro do backend
function translateError(backendError, fallbackKey = 'auth.errors.generic') {
  const { t } = i18n.global
  
  // Tentar encontrar tradução pelo mapeamento
  const translationKey = errorMap[backendError]
  if (translationKey) {
    return t(translationKey)
  }
  
  // Se não encontrou, retornar a mensagem original ou fallback
  return t(fallbackKey)
}

// Verificar se é um erro de sessão
function isSessionError(response, data) {
  if (response.status === 401) return true
  if (data?.code && SESSION_ERROR_CODES.includes(data.code)) return true
  return false
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const loading = ref(false)
  const error = ref('')
  const sessionCheckInProgress = ref(false)
  const lastSessionCheck = ref(0)

  const isAuthenticated = computed(() => !!user.value)

  // Verificar sessão com debounce para evitar múltiplas chamadas
  async function checkSession(force = false) {
    // Evitar múltiplas verificações simultâneas
    if (sessionCheckInProgress.value) {
      return !!user.value
    }

    // Debounce: não verificar mais de uma vez a cada 5 segundos
    const now = Date.now()
    if (!force && (now - lastSessionCheck.value) < 5000) {
      return !!user.value
    }

    sessionCheckInProgress.value = true
    lastSessionCheck.value = now

    try {
      const res = await fetch('/api/auth/me', { 
        credentials: 'include',
        headers: {
          'Cache-Control': 'no-cache'
        }
      })
      
      if (res.ok) {
        const data = await res.json()
        user.value = data.user
        console.debug('[Auth] Sessão válida para:', data.user?.email)
        return true
      }

      // Sessão inválida ou expirada
      if (res.status === 401) {
        console.debug('[Auth] Sessão inválida ou expirada')
        user.value = null
        return false
      }

    } catch (e) {
      console.error('[Auth] Erro ao verificar sessão:', e)
      // Em caso de erro de rede, manter o estado atual
    } finally {
      sessionCheckInProgress.value = false
    }

    return false
  }

  // Handler global para erros de sessão em outras requisições
  async function handleSessionError() {
    console.debug('[Auth] Tratando erro de sessão...')
    user.value = null
    
    // Redirecionar para login apenas se não estiver já na página de auth
    if (!window.location.pathname.includes('/auth') && 
        !window.location.pathname.includes('/login') &&
        window.location.pathname !== '/') {
      window.location.href = '/'
    }
  }

  async function register(email, password, name) {
    loading.value = true
    error.value = ''

    try {
      const res = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ email, password, name })
      })

      const data = await res.json()

      if (!res.ok) {
        error.value = translateError(data.error, 'auth.errors.registerFailed')
        return false
      }

      user.value = data.user
      return true
    } catch (e) {
      error.value = translateError(null, 'auth.errors.connectionError')
      return false
    } finally {
      loading.value = false
    }
  }

  async function login(email, password) {
    loading.value = true
    error.value = ''

    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ email, password })
      })

      const data = await res.json()

      if (!res.ok) {
        error.value = translateError(data.error, 'auth.errors.invalidCredentials')
        return false
      }

      user.value = data.user
      return true
    } catch (e) {
      error.value = translateError(null, 'auth.errors.connectionError')
      return false
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      await fetch('/api/auth/logout', {
        method: 'POST',
        credentials: 'include'
      })
    } catch (e) {
      // Ignorar erro de logout
    }
    user.value = null
    lastSessionCheck.value = 0
  }

  // Login via Google OAuth
  async function loginWithGoogle(idToken) {
    loading.value = true
    error.value = ''

    try {
      const res = await fetch('/api/auth/oauth/google', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ token: idToken })
      })

      const data = await res.json()

      if (!res.ok) {
        error.value = translateError(data.error, 'auth.socialLoginError')
        return false
      }

      user.value = data.user
      return true
    } catch (e) {
      error.value = translateError(null, 'auth.errors.connectionError')
      return false
    } finally {
      loading.value = false
    }
  }

  // Login via Apple OAuth
  async function loginWithApple(idToken) {
    loading.value = true
    error.value = ''

    try {
      const res = await fetch('/api/auth/oauth/apple', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ token: idToken })
      })

      const data = await res.json()

      if (!res.ok) {
        error.value = translateError(data.error, 'auth.socialLoginError')
        return false
      }

      user.value = data.user
      return true
    } catch (e) {
      error.value = translateError(null, 'auth.errors.connectionError')
      return false
    } finally {
      loading.value = false
    }
  }

  function clearError() {
    error.value = ''
  }

  // Wrapper para fazer requisições autenticadas com tratamento de sessão
  async function authenticatedFetch(url, options = {}) {
    const res = await fetch(url, {
      ...options,
      credentials: 'include'
    })

    // Se for erro de sessão, tratar
    if (res.status === 401) {
      try {
        const data = await res.clone().json()
        if (isSessionError(res, data)) {
          await handleSessionError()
        }
      } catch (e) {
        await handleSessionError()
      }
    }

    return res
  }

  return {
    user,
    loading,
    error,
    isAuthenticated,
    checkSession,
    register,
    login,
    logout,
    loginWithGoogle,
    loginWithApple,
    clearError,
    handleSessionError,
    authenticatedFetch
  }
})
