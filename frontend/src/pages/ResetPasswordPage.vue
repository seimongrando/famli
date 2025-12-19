<template>
  <div class="reset-password-page">
    <div class="reset-container">
      <!-- Logo -->
      <div class="logo">
        <router-link to="/">
          <span class="logo-icon">🏠</span>
          <span class="logo-text">famli</span>
        </router-link>
      </div>

      <!-- Forgot Password Form -->
      <div v-if="mode === 'forgot'" class="reset-card">
        <h1>{{ $t('password.forgot_title') }}</h1>
        <p class="subtitle">{{ $t('password.forgot_subtitle') }}</p>

        <form @submit.prevent="handleForgotPassword" class="reset-form">
          <div class="form-group">
            <label for="email">{{ $t('auth.email') }}</label>
            <input
              id="email"
              v-model="email"
              type="email"
              :placeholder="$t('auth.emailPlaceholder')"
              required
              :disabled="loading"
            />
          </div>

          <button type="submit" class="btn-submit" :disabled="loading || !email">
            {{ loading ? $t('common.sending') : $t('password.send_link') }}
          </button>

          <p v-if="message" class="success-message">
            ✅ {{ message }}
          </p>

          <p v-if="error" class="error-message">
            {{ error }}
          </p>
        </form>

        <div class="back-link">
          <router-link :to="{ name: 'auth' }">
            ← {{ $t('password.back_to_login') }}
          </router-link>
        </div>
      </div>

      <!-- Reset Password Form (with token) -->
      <div v-else-if="mode === 'reset'" class="reset-card">
        <h1>{{ $t('password.reset_title') }}</h1>
        <p class="subtitle">{{ $t('password.reset_subtitle') }}</p>

        <form @submit.prevent="handleResetPassword" class="reset-form">
          <div class="form-group">
            <label for="newPassword">{{ $t('password.new_password') }}</label>
            <input
              id="newPassword"
              v-model="newPassword"
              type="password"
              :placeholder="$t('password.new_password_placeholder')"
              required
              minlength="8"
              :disabled="loading"
            />
          </div>

          <div class="form-group">
            <label for="confirmPassword">{{ $t('password.confirm_password') }}</label>
            <input
              id="confirmPassword"
              v-model="confirmPassword"
              type="password"
              :placeholder="$t('password.confirm_password_placeholder')"
              required
              minlength="8"
              :disabled="loading"
            />
          </div>

          <p v-if="newPassword && confirmPassword && newPassword !== confirmPassword" class="validation-error">
            {{ $t('password.passwords_not_match') }}
          </p>

          <button 
            type="submit" 
            class="btn-submit" 
            :disabled="loading || !canSubmitReset"
          >
            {{ loading ? $t('common.saving') : $t('password.reset_button') }}
          </button>

          <p v-if="successReset" class="success-message">
            ✅ {{ $t('password.reset_success') }}
            <router-link :to="{ name: 'auth' }">{{ $t('password.login_now') }}</router-link>
          </p>

          <p v-if="error" class="error-message">
            {{ error }}
          </p>
        </form>
      </div>

      <!-- Invalid/Expired Token -->
      <div v-else-if="mode === 'invalid'" class="reset-card">
        <div class="invalid-icon">😕</div>
        <h1>{{ $t('password.link_expired_title') }}</h1>
        <p class="subtitle">{{ $t('password.link_expired_subtitle') }}</p>
        
        <button @click="mode = 'forgot'" class="btn-submit">
          {{ $t('password.request_new_link') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const mode = ref('forgot') // 'forgot', 'reset', 'invalid'
const token = ref('')
const email = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const message = ref('')
const error = ref('')
const successReset = ref(false)

const canSubmitReset = computed(() => {
  return newPassword.value.length >= 8 && 
         newPassword.value === confirmPassword.value
})

onMounted(() => {
  // Verificar se há token na URL
  const urlToken = route.query.token
  if (urlToken) {
    token.value = urlToken
    mode.value = 'reset'
  }
})

async function handleForgotPassword() {
  if (!email.value) return
  
  try {
    loading.value = true
    error.value = ''
    message.value = ''
    
    const response = await fetch('/api/auth/forgot-password', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: email.value })
    })
    
    const data = await response.json()
    
    if (response.ok) {
      message.value = data.message
      email.value = ''
    } else {
      error.value = data.error
    }
  } catch (err) {
    error.value = 'Erro ao enviar. Tente novamente.'
  } finally {
    loading.value = false
  }
}

async function handleResetPassword() {
  if (!canSubmitReset.value) return
  
  try {
    loading.value = true
    error.value = ''
    
    const response = await fetch('/api/auth/reset-password', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        token: token.value,
        new_password: newPassword.value
      })
    })
    
    const data = await response.json()
    
    if (response.ok) {
      successReset.value = true
      // Redirecionar após 3 segundos
      setTimeout(() => {
        router.push({ name: 'auth' })
      }, 3000)
    } else {
      if (response.status === 400) {
        mode.value = 'invalid'
      } else {
        error.value = data.error
      }
    }
  } catch (err) {
    error.value = 'Erro ao redefinir senha. Tente novamente.'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.reset-password-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg, #faf8f5);
  padding: 1rem;
}

.reset-container {
  width: 100%;
  max-width: 420px;
}

.logo {
  text-align: center;
  margin-bottom: 2rem;
}

.logo a {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  text-decoration: none;
}

.logo-icon {
  font-size: 2rem;
}

.logo-text {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-primary, #2d5a47);
}

.reset-card {
  background: white;
  border-radius: 1rem;
  padding: 2rem;
  box-shadow: 0 10px 40px rgba(0,0,0,0.1);
}

.reset-card h1 {
  margin: 0 0 0.5rem;
  font-size: 1.5rem;
  color: #1e293b;
  text-align: center;
}

.subtitle {
  color: #64748b;
  text-align: center;
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
}

.reset-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  font-weight: 500;
  color: #1e293b;
  font-size: 0.875rem;
}

.form-group input {
  padding: 0.875rem 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  font-size: 1rem;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: var(--color-primary, #2d5a47);
  box-shadow: 0 0 0 3px rgba(45, 90, 71, 0.1);
}

.form-group input:disabled {
  background: #f8fafc;
  cursor: not-allowed;
}

.btn-submit {
  padding: 1rem;
  background: var(--color-accent, #e07b39);
  color: white;
  border: none;
  border-radius: var(--radius-md, 12px);
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
  margin-top: 0.5rem;
}

.btn-submit:hover:not(:disabled) {
  background: var(--color-accent-light, #f4a876);
}

.btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.success-message {
  background: #ecfdf5;
  color: #065f46;
  padding: 1rem;
  border-radius: 0.5rem;
  text-align: center;
  font-size: 0.9rem;
}

.success-message a {
  color: var(--color-primary, #2d5a47);
  font-weight: 600;
}

.error-message {
  background: #fef2f2;
  color: #991b1b;
  padding: 1rem;
  border-radius: 0.5rem;
  text-align: center;
  font-size: 0.9rem;
}

.validation-error {
  color: #ef4444;
  font-size: 0.875rem;
  margin: 0;
}

.back-link {
  text-align: center;
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid #e2e8f0;
}

.back-link a {
  color: var(--color-primary, #2d5a47);
  text-decoration: none;
  font-size: 0.9rem;
}

.back-link a:hover {
  text-decoration: underline;
}

.invalid-icon {
  font-size: 4rem;
  text-align: center;
  margin-bottom: 1rem;
}
</style>

