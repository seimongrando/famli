<!-- =============================================================================
  FAMLI - Cookie Consent Banner
  =============================================================================
  Banner de consentimento de cookies conforme LGPD e GDPR.
  
  Características:
  - Design moderno e não intrusivo
  - Opções granulares por categoria
  - Aceitar todos / Rejeitar opcionais
  - Detalhes expandíveis
  - Acessível (ARIA)
============================================================================== -->

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useCookieConsent } from '../composables/useCookieConsent'

const { t } = useI18n()
const { 
  showBanner, 
  preferences, 
  acceptAll, 
  rejectAll, 
  savePreferences 
} = useCookieConsent()

const showDetails = ref(false)

function handleAcceptAll() {
  acceptAll()
}

function handleRejectOptional() {
  rejectAll()
}

function handleSavePreferences() {
  savePreferences()
}

function toggleDetails() {
  showDetails.value = !showDetails.value
}
</script>

<template>
  <Teleport to="body">
    <Transition name="cookie-banner">
      <div 
        v-if="showBanner" 
        class="cookie-overlay"
        role="dialog"
        aria-modal="true"
        :aria-label="t('cookies.title')"
      >
        <div class="cookie-banner">
          <!-- Header -->
          <div class="cookie-header">
            <div class="cookie-icon">🍪</div>
            <div class="cookie-title-wrapper">
              <h2 class="cookie-title">{{ t('cookies.title') }}</h2>
              <p class="cookie-subtitle">{{ t('cookies.subtitle') }}</p>
            </div>
          </div>

          <!-- Descrição -->
          <div class="cookie-body">
            <p class="cookie-description">
              {{ t('cookies.description') }}
            </p>

            <!-- Toggle para detalhes -->
            <button 
              type="button" 
              class="cookie-details-toggle"
              @click="toggleDetails"
              :aria-expanded="showDetails"
            >
              <span>{{ showDetails ? t('cookies.hideDetails') : t('cookies.showDetails') }}</span>
              <svg 
                class="toggle-arrow" 
                :class="{ 'toggle-arrow--open': showDetails }"
                viewBox="0 0 24 24" 
                fill="none"
              >
                <path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </button>

            <!-- Detalhes expandíveis -->
            <Transition name="details">
              <div v-if="showDetails" class="cookie-categories">
                <!-- Cookies Necessários -->
                <div class="cookie-category">
                  <div class="category-header">
                    <div class="category-info">
                      <span class="category-icon">🔒</span>
                      <div>
                        <h3 class="category-title">{{ t('cookies.necessary.title') }}</h3>
                        <p class="category-desc">{{ t('cookies.necessary.description') }}</p>
                      </div>
                    </div>
                    <div class="category-toggle category-toggle--required">
                      <span class="required-badge">{{ t('cookies.required') }}</span>
                    </div>
                  </div>
                </div>

                <!-- Cookies de Analytics -->
                <div class="cookie-category">
                  <div class="category-header">
                    <div class="category-info">
                      <span class="category-icon">📊</span>
                      <div>
                        <h3 class="category-title">{{ t('cookies.analytics.title') }}</h3>
                        <p class="category-desc">{{ t('cookies.analytics.description') }}</p>
                      </div>
                    </div>
                    <label class="category-toggle">
                      <input 
                        type="checkbox" 
                        v-model="preferences.analytics" 
                        class="toggle-checkbox"
                      />
                      <span class="toggle-slider"></span>
                    </label>
                  </div>
                </div>

                <!-- Cookies de Preferências -->
                <div class="cookie-category">
                  <div class="category-header">
                    <div class="category-info">
                      <span class="category-icon">⚙️</span>
                      <div>
                        <h3 class="category-title">{{ t('cookies.preferences.title') }}</h3>
                        <p class="category-desc">{{ t('cookies.preferences.description') }}</p>
                      </div>
                    </div>
                    <label class="category-toggle">
                      <input 
                        type="checkbox" 
                        v-model="preferences.preferences" 
                        class="toggle-checkbox"
                      />
                      <span class="toggle-slider"></span>
                    </label>
                  </div>
                </div>
              </div>
            </Transition>
          </div>

          <!-- Links -->
          <div class="cookie-links">
            <router-link to="/privacidade" class="cookie-link">
              {{ t('cookies.privacyPolicy') }}
            </router-link>
            <span class="cookie-link-separator">•</span>
            <router-link to="/termos" class="cookie-link">
              {{ t('cookies.termsOfService') }}
            </router-link>
          </div>

          <!-- Ações -->
          <div class="cookie-actions">
            <button 
              type="button" 
              class="cookie-btn cookie-btn--secondary"
              @click="handleRejectOptional"
            >
              {{ t('cookies.rejectOptional') }}
            </button>
            
            <button 
              v-if="showDetails"
              type="button" 
              class="cookie-btn cookie-btn--outline"
              @click="handleSavePreferences"
            >
              {{ t('cookies.savePreferences') }}
            </button>
            
            <button 
              type="button" 
              class="cookie-btn cookie-btn--primary"
              @click="handleAcceptAll"
            >
              {{ t('cookies.acceptAll') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.cookie-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  padding: 1rem;
  z-index: 9999;
}

.cookie-banner {
  background: white;
  border-radius: 1.25rem 1.25rem 0.75rem 0.75rem;
  max-width: 560px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 
    0 -4px 24px rgba(0, 0, 0, 0.12),
    0 0 0 1px rgba(0, 0, 0, 0.04);
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.cookie-header {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 1.5rem 1.5rem 0;
}

.cookie-icon {
  font-size: 2rem;
  line-height: 1;
}

.cookie-title-wrapper {
  flex: 1;
}

.cookie-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--color-text);
  margin: 0;
  line-height: 1.3;
}

.cookie-subtitle {
  font-size: 0.85rem;
  color: var(--color-text-muted);
  margin: 0.25rem 0 0;
}

.cookie-body {
  padding: 1rem 1.5rem;
}

.cookie-description {
  font-size: 0.9rem;
  color: var(--color-text-soft);
  line-height: 1.6;
  margin: 0 0 1rem;
}

.cookie-details-toggle {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: none;
  border: none;
  color: var(--color-primary);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  padding: 0.5rem 0;
  transition: opacity 0.2s;
}

.cookie-details-toggle:hover {
  opacity: 0.8;
}

.toggle-arrow {
  width: 16px;
  height: 16px;
  transition: transform 0.2s ease;
}

.toggle-arrow--open {
  transform: rotate(180deg);
}

.cookie-categories {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border-light);
}

.cookie-category {
  background: #f8faf9;
  border-radius: 0.75rem;
  padding: 1rem;
  border: 1px solid var(--color-border-light);
}

.category-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.category-info {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  flex: 1;
}

.category-icon {
  font-size: 1.25rem;
  line-height: 1;
  margin-top: 0.125rem;
}

.category-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--color-text);
  margin: 0 0 0.25rem;
}

.category-desc {
  font-size: 0.8rem;
  color: var(--color-text-muted);
  margin: 0;
  line-height: 1.4;
}

.category-toggle {
  flex-shrink: 0;
}

.category-toggle--required {
  display: flex;
  align-items: center;
}

.required-badge {
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-soft);
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

/* Toggle Switch */
.toggle-checkbox {
  display: none;
}

.toggle-slider {
  display: block;
  width: 44px;
  height: 24px;
  background: #e0e0e0;
  border-radius: 12px;
  position: relative;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.toggle-slider::after {
  content: '';
  position: absolute;
  width: 20px;
  height: 20px;
  background: white;
  border-radius: 50%;
  top: 2px;
  left: 2px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
  transition: transform 0.2s ease;
}

.toggle-checkbox:checked + .toggle-slider {
  background: var(--color-primary);
}

.toggle-checkbox:checked + .toggle-slider::after {
  transform: translateX(20px);
}

.cookie-links {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0 1.5rem;
  margin-bottom: 1rem;
}

.cookie-link {
  font-size: 0.8rem;
  color: var(--color-text-muted);
  text-decoration: none;
  transition: color 0.2s;
}

.cookie-link:hover {
  color: var(--color-primary);
  text-decoration: underline;
}

.cookie-link-separator {
  color: var(--color-text-muted);
  opacity: 0.5;
}

.cookie-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  padding: 1rem 1.5rem 1.5rem;
  border-top: 1px solid var(--color-border-light);
}

.cookie-btn {
  flex: 1;
  min-width: 120px;
  padding: 0.875rem 1.25rem;
  font-size: 0.9rem;
  font-weight: 600;
  border-radius: 0.625rem;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}

.cookie-btn--primary {
  background: var(--color-primary);
  color: white;
}

.cookie-btn--primary:hover {
  background: var(--color-primary-dark);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(45, 90, 71, 0.25);
}

.cookie-btn--secondary {
  background: #f5f5f5;
  color: var(--color-text-soft);
}

.cookie-btn--secondary:hover {
  background: #ebebeb;
}

.cookie-btn--outline {
  background: transparent;
  border: 2px solid var(--color-primary);
  color: var(--color-primary);
}

.cookie-btn--outline:hover {
  background: var(--color-primary-soft);
}

/* Transições */
.cookie-banner-enter-active,
.cookie-banner-leave-active {
  transition: opacity 0.2s ease;
}

.cookie-banner-enter-active .cookie-banner,
.cookie-banner-leave-active .cookie-banner {
  transition: transform 0.3s ease;
}

.cookie-banner-enter-from,
.cookie-banner-leave-to {
  opacity: 0;
}

.cookie-banner-enter-from .cookie-banner,
.cookie-banner-leave-to .cookie-banner {
  transform: translateY(100%);
}

.details-enter-active,
.details-leave-active {
  transition: all 0.2s ease;
  overflow: hidden;
}

.details-enter-from,
.details-leave-to {
  opacity: 0;
  max-height: 0;
}

.details-enter-to,
.details-leave-from {
  opacity: 1;
  max-height: 500px;
}

/* Responsivo */
@media (max-width: 480px) {
  .cookie-overlay {
    padding: 0;
  }
  
  .cookie-banner {
    border-radius: 1rem 1rem 0 0;
    max-height: 85vh;
  }
  
  .cookie-header {
    padding: 1.25rem 1.25rem 0;
  }
  
  .cookie-body {
    padding: 0.875rem 1.25rem;
  }
  
  .cookie-actions {
    padding: 1rem 1.25rem 1.25rem;
  }
  
  .cookie-btn {
    min-width: 100%;
  }
  
  .category-header {
    flex-direction: column;
    gap: 0.75rem;
  }
  
  .category-toggle {
    align-self: flex-end;
  }
}
</style>

