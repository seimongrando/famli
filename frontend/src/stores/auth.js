// =============================================================================
// FAMLI - Store de Autenticação
// =============================================================================
// Gerencia estado de autenticação do usuário.
// 
// Funcionalidades:
// - Login/Logout/Registro
// - Verificação de sessão
// - Mensagens de erro traduzidas
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
  'Sessão não encontrada': 'auth.errors.sessionNotFound',
  'Muitas tentativas. Aguarde alguns minutos.': 'auth.errors.tooManyAttempts',
  'Erro ao preparar sua conta.': 'auth.errors.serverError'
}

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

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const loading = ref(false)
  const error = ref('')

  const isAuthenticated = computed(() => !!user.value)

  async function checkSession() {
    try {
      const res = await fetch('/api/auth/me', { credentials: 'include' })
      if (res.ok) {
        const data = await res.json()
        user.value = data.user
        return true
      }
    } catch (e) {
      // Sessão não existe ou expirou
    }
    return false
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
  }

  function clearError() {
    error.value = ''
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
    clearError
  }
})
