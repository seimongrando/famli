import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

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
        error.value = data.error || 'Erro ao criar conta'
        return false
      }

      user.value = data.user
      return true
    } catch (e) {
      error.value = 'Erro de conexão. Tente novamente.'
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
        error.value = data.error || 'E-mail ou senha incorretos'
        return false
      }

      user.value = data.user
      return true
    } catch (e) {
      error.value = 'Erro de conexão. Tente novamente.'
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
