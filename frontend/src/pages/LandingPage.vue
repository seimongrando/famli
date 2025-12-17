<script setup>
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { useLocalizedRoutes } from '../composables/useLocalizedRoutes'
import { onMounted } from 'vue'
import LanguageSelector from '../components/LanguageSelector.vue'

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()
const { paths, getPath } = useLocalizedRoutes()

onMounted(() => {
  if (authStore.isAuthenticated) {
    router.push(paths.value.dashboard)
  }
})
</script>

<template>
  <div class="landing">
    <!-- Header -->
    <header class="landing-header">
      <div class="container">
        <div class="landing-header__content">
          <div class="header__brand">
            <img src="/logo.svg" alt="Famli" class="header__logo" />
            <span class="header__name">{{ t('brand.name') }}</span>
          </div>
          <nav class="landing-header__nav">
            <LanguageSelector />
            <router-link :to="{ path: paths.auth, query: { mode: 'login' } }" class="btn btn--ghost">
              {{ t('nav.login') }}
            </router-link>
            <router-link :to="{ path: paths.auth, query: { mode: 'register' } }" class="btn btn--primary">
              {{ t('nav.register') }}
            </router-link>
          </nav>
        </div>
      </div>
    </header>

    <!-- Hero -->
    <section class="hero">
      <div class="container">
        <div class="hero__content">
          <div class="hero__text">
            <h1 class="hero__title">
              {{ t('landing.hero.title') }}<br />
              <span class="text-accent">{{ t('landing.hero.titleHighlight') }}</span> {{ t('landing.hero.titleEnd') }}
            </h1>
            <p class="hero__subtitle">
              {{ t('landing.hero.subtitle') }}
            </p>
            <div class="hero__actions">
              <router-link :to="{ path: paths.auth, query: { mode: 'register' } }" class="btn btn--accent btn--large">
                {{ t('landing.hero.cta') }}
              </router-link>
              <router-link :to="{ path: paths.auth, query: { mode: 'login' } }" class="btn btn--link">
                {{ t('landing.hero.ctaSecondary') }}
              </router-link>
            </div>
            <p class="hero__trust-badge">{{ t('landing.hero.trustBadge') }}</p>
          </div>
          <div class="hero__card">
            <div class="hero__card-header">
              <span class="hero__card-icon">🏠</span>
              <span class="hero__card-title">{{ t('landing.card.title') }}</span>
            </div>
            <ul class="hero__card-list">
              <li>✓ {{ t('landing.card.item1') }}</li>
              <li>✓ {{ t('landing.card.item2') }}</li>
              <li>✓ {{ t('landing.card.item3') }}</li>
              <li>✓ {{ t('landing.card.item4') }}</li>
            </ul>
          </div>
        </div>
      </div>
    </section>

    <!-- Para quem é -->
    <section class="section section--warm">
      <div class="container">
        <h2 class="section__title text-center">{{ t('landing.audience.title') }}</h2>
        <div class="audience-grid">
          <div class="audience-card">
            <span class="audience-card__icon">🌟</span>
            <h3 class="audience-card__title">{{ t('landing.audience.senior.title') }}</h3>
            <p class="audience-card__description">{{ t('landing.audience.senior.description') }}</p>
          </div>
          <div class="audience-card">
            <span class="audience-card__icon">👨‍👩‍👧‍👦</span>
            <h3 class="audience-card__title">{{ t('landing.audience.family.title') }}</h3>
            <p class="audience-card__description">{{ t('landing.audience.family.description') }}</p>
          </div>
          <div class="audience-card">
            <span class="audience-card__icon">💚</span>
            <h3 class="audience-card__title">{{ t('landing.audience.caregiver.title') }}</h3>
            <p class="audience-card__description">{{ t('landing.audience.caregiver.description') }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Funcionalidades -->
    <section class="section">
      <div class="container">
        <h2 class="section__title text-center">{{ t('landing.features.title') }}</h2>
        <p class="section__subtitle text-center">
          {{ t('landing.features.subtitle') }}
        </p>
        <div class="features-grid">
          <div class="feature-card">
            <span class="feature-card__icon">📋</span>
            <h3 class="feature-card__title">{{ t('landing.features.info.title') }}</h3>
            <p class="feature-card__description">{{ t('landing.features.info.description') }}</p>
          </div>
          <div class="feature-card">
            <span class="feature-card__icon">👥</span>
            <h3 class="feature-card__title">{{ t('landing.features.people.title') }}</h3>
            <p class="feature-card__description">{{ t('landing.features.people.description') }}</p>
          </div>
          <div class="feature-card">
            <span class="feature-card__icon">🔑</span>
            <h3 class="feature-card__title">{{ t('landing.features.access.title') }}</h3>
            <p class="feature-card__description">{{ t('landing.features.access.description') }}</p>
          </div>
          <div class="feature-card">
            <span class="feature-card__icon">💝</span>
            <h3 class="feature-card__title">{{ t('landing.features.memories.title') }}</h3>
            <p class="feature-card__description">{{ t('landing.features.memories.description') }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Diferenciador -->
    <section class="section section--warm">
      <div class="container container--narrow">
        <div class="differentiator">
          <h2 class="differentiator__title">{{ t('landing.differentiator.title') }}</h2>
          <p class="differentiator__subtitle">{{ t('landing.differentiator.subtitle') }}</p>
          <ul class="differentiator__list">
            <li>
              <span class="differentiator__icon">📝</span>
              {{ t('landing.differentiator.point1') }}
            </li>
            <li>
              <span class="differentiator__icon">💬</span>
              {{ t('landing.differentiator.point2') }}
            </li>
            <li>
              <span class="differentiator__icon">🎯</span>
              {{ t('landing.differentiator.point3') }}
            </li>
          </ul>
        </div>
      </div>
    </section>

    <!-- Trust Signals -->
    <section class="section">
      <div class="container">
        <h2 class="section__title text-center">{{ t('landing.trust.title') }}</h2>
        <p class="section__subtitle text-center">{{ t('landing.trust.subtitle') }}</p>
        <div class="trust-grid">
          <div class="trust-card">
            <span class="trust-card__icon">🔐</span>
            <h3 class="trust-card__title">{{ t('landing.trust.encryption.title') }}</h3>
            <p class="trust-card__description">{{ t('landing.trust.encryption.description') }}</p>
          </div>
          <div class="trust-card">
            <span class="trust-card__icon">🎛️</span>
            <h3 class="trust-card__title">{{ t('landing.trust.control.title') }}</h3>
            <p class="trust-card__description">{{ t('landing.trust.control.description') }}</p>
          </div>
          <div class="trust-card">
            <span class="trust-card__icon">🛡️</span>
            <h3 class="trust-card__title">{{ t('landing.trust.privacy.title') }}</h3>
            <p class="trust-card__description">{{ t('landing.trust.privacy.description') }}</p>
          </div>
          <div class="trust-card">
            <span class="trust-card__icon">🗑️</span>
            <h3 class="trust-card__title">{{ t('landing.trust.delete.title') }}</h3>
            <p class="trust-card__description">{{ t('landing.trust.delete.description') }}</p>
          </div>
          <div class="trust-card">
            <span class="trust-card__icon">📜</span>
            <h3 class="trust-card__title">{{ t('landing.trust.lgpd.title') }}</h3>
            <p class="trust-card__description">{{ t('landing.trust.lgpd.description') }}</p>
          </div>
          <div class="trust-card">
            <span class="trust-card__icon">👁️</span>
            <h3 class="trust-card__title">{{ t('landing.trust.opensource.title') }}</h3>
            <p class="trust-card__description">{{ t('landing.trust.opensource.description') }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Segurança (resumo) -->
    <section class="section section--highlight">
      <div class="container container--narrow">
        <div class="security-card">
          <h2 class="security-card__title">
            🔒 {{ t('landing.security.title') }}
          </h2>
          <p class="security-card__subtitle">{{ t('landing.security.subtitle') }}</p>
          <ul class="security-card__list">
            <li>✓ {{ t('landing.security.item1') }}</li>
            <li>✓ {{ t('landing.security.item2') }}</li>
            <li>✓ {{ t('landing.security.item3') }}</li>
            <li>✓ {{ t('landing.security.item4') }}</li>
          </ul>
        </div>
      </div>
    </section>

    <!-- CTA Final -->
    <section class="section section--cta">
      <div class="container container--narrow text-center">
        <h2 class="section__title">{{ t('landing.cta.title') }}</h2>
        <p class="section__subtitle">
          {{ t('landing.cta.subtitle') }}
        </p>
        <router-link :to="{ path: paths.auth, query: { mode: 'register' } }" class="btn btn--primary btn--large">
          {{ t('landing.cta.button') }}
        </router-link>
      </div>
    </section>

    <!-- Footer -->
    <footer class="footer">
      <div class="container">
        <div class="footer__content">
          <div class="footer__brand">
            <div class="header__brand">
              <img src="/logo.svg" alt="Famli" class="header__logo" style="height: 32px" />
              <span class="header__name" style="font-size: 1rem">{{ t('brand.name') }}</span>
            </div>
            <p class="footer__tagline">
              {{ t('brand.slogan') }}
            </p>
          </div>
          <div class="footer__links">
            <router-link :to="paths.terms" class="footer__link">
              {{ t('legal.terms.link') }}
            </router-link>
            <router-link :to="paths.privacy" class="footer__link">
              {{ t('legal.privacy.link') }}
            </router-link>
          </div>
        </div>
        <p class="footer__copyright">
          © {{ new Date().getFullYear() }} Famli. {{ t('privacy.rights.item4') }}
        </p>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.landing {
  min-height: 100vh;
}

/* Header */
.landing-header {
  position: sticky;
  top: 0;
  background: rgba(250, 248, 245, 0.95);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--color-border-light);
  z-index: 50;
}

.landing-header__content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-md) 0;
}

.landing-header__nav {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

/* Hero */
.hero {
  padding: var(--space-3xl) 0;
  background: linear-gradient(180deg, var(--color-bg) 0%, var(--color-bg-warm) 100%);
}

.hero__content {
  display: grid;
  grid-template-columns: 1.4fr 1fr;
  gap: var(--space-2xl);
  align-items: center;
}

.hero__title {
  font-size: clamp(2rem, 4vw, 2.75rem);
  line-height: 1.2;
  margin-bottom: var(--space-lg);
}

.hero__subtitle {
  font-size: var(--font-size-lg);
  color: var(--color-text-soft);
  margin-bottom: var(--space-xl);
  max-width: 540px;
}

.hero__actions {
  display: flex;
  align-items: center;
  gap: var(--space-lg);
  flex-wrap: wrap;
}

.hero__card {
  background: var(--color-card);
  border-radius: var(--radius-xl);
  padding: var(--space-xl);
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--color-border-light);
}

.hero__card-header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  margin-bottom: var(--space-lg);
}

.hero__card-icon {
  font-size: 1.5rem;
}

.hero__card-title {
  font-weight: 700;
  color: var(--color-text);
}

.hero__card-list {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.hero__card-list li {
  color: var(--color-text-soft);
  font-size: var(--font-size-base);
}

/* Sections */
.section {
  padding: var(--space-3xl) 0;
}

.section--warm {
  background: var(--color-bg-warm);
}

.section--highlight {
  background: var(--color-primary-soft);
}

.section--cta {
  background: linear-gradient(180deg, var(--color-bg) 0%, var(--color-bg-warm) 100%);
  padding: var(--space-3xl) 0 var(--space-2xl);
}

.section__title {
  margin-bottom: var(--space-md);
}

.section__subtitle {
  font-size: var(--font-size-lg);
  color: var(--color-text-soft);
  margin-bottom: var(--space-xl);
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

/* Audience cards */
.audience-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-lg);
  margin-top: var(--space-xl);
}

.audience-card {
  background: var(--color-card);
  border-radius: var(--radius-lg);
  padding: var(--space-xl);
  text-align: center;
  border: 1px solid var(--color-border-light);
  box-shadow: var(--shadow-sm);
}

.audience-card__icon {
  font-size: 2.5rem;
  display: block;
  margin-bottom: var(--space-md);
}

.audience-card__title {
  font-size: var(--font-size-lg);
  margin-bottom: var(--space-sm);
}

.audience-card__description {
  color: var(--color-text-soft);
  margin: 0;
}

/* Feature cards */
.features-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--space-lg);
}

.feature-card {
  background: var(--color-card);
  border-radius: var(--radius-lg);
  padding: var(--space-xl);
  border: 1px solid var(--color-border-light);
  box-shadow: var(--shadow-sm);
  transition: transform var(--transition-normal), box-shadow var(--transition-normal);
}

.feature-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-md);
}

.feature-card__icon {
  font-size: 2rem;
  display: block;
  margin-bottom: var(--space-md);
}

.feature-card__title {
  font-size: var(--font-size-lg);
  margin-bottom: var(--space-sm);
}

.feature-card__description {
  color: var(--color-text-soft);
  margin: 0;
}

/* Security card */
.security-card {
  background: var(--color-card);
  border-radius: var(--radius-xl);
  padding: var(--space-2xl);
  text-align: center;
  box-shadow: var(--shadow-md);
}

.security-card__title {
  font-size: var(--font-size-xl);
  margin-bottom: var(--space-lg);
}

.security-card__list {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  margin-bottom: var(--space-lg);
}

.security-card__list li {
  font-size: var(--font-size-lg);
  color: var(--color-text-soft);
}

.security-card__note {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
  margin: 0;
}

.security-card__subtitle {
  color: var(--color-text-soft);
  margin-bottom: var(--space-lg);
}

/* Hero Trust Badge */
.hero__trust-badge {
  margin-top: var(--space-lg);
  padding: var(--space-sm) var(--space-md);
  background: var(--color-success-light, rgba(34, 197, 94, 0.1));
  color: var(--color-success, #16a34a);
  border-radius: var(--radius-full);
  font-size: var(--font-size-sm);
  font-weight: 500;
  display: inline-block;
}

/* Differentiator */
.differentiator {
  background: var(--color-card);
  border-radius: var(--radius-xl);
  padding: var(--space-2xl);
  text-align: center;
  box-shadow: var(--shadow-md);
  border: 2px solid var(--color-primary-soft);
}

.differentiator__title {
  font-size: var(--font-size-2xl);
  color: var(--color-primary);
  margin-bottom: var(--space-md);
}

.differentiator__subtitle {
  font-size: var(--font-size-lg);
  color: var(--color-text-soft);
  margin-bottom: var(--space-xl);
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

.differentiator__list {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  text-align: left;
  max-width: 400px;
  margin: 0 auto;
}

.differentiator__list li {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  font-size: var(--font-size-lg);
  color: var(--color-text);
}

.differentiator__icon {
  font-size: 1.5rem;
}

/* Trust Grid */
.trust-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-lg);
}

.trust-card {
  background: var(--color-card);
  border-radius: var(--radius-lg);
  padding: var(--space-xl);
  text-align: center;
  border: 1px solid var(--color-border-light);
  box-shadow: var(--shadow-sm);
  transition: transform var(--transition-normal), box-shadow var(--transition-normal);
}

.trust-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-md);
}

.trust-card__icon {
  font-size: 2.5rem;
  display: block;
  margin-bottom: var(--space-md);
}

.trust-card__title {
  font-size: var(--font-size-base);
  font-weight: 700;
  margin-bottom: var(--space-sm);
  color: var(--color-text);
}

.trust-card__description {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  margin: 0;
  line-height: 1.6;
}

/* Footer */
.footer {
  background: var(--color-bg-warm);
  padding: var(--space-xl) 0;
  border-top: 1px solid var(--color-border);
}

.footer__content {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: var(--space-lg);
}

.footer__brand {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.footer__tagline {
  color: var(--color-text-muted);
  margin: 0;
  font-size: var(--font-size-sm);
}

.footer__links {
  display: flex;
  gap: var(--space-lg);
}

.footer__link {
  color: var(--color-text-soft);
  font-size: var(--font-size-sm);
  text-decoration: none;
  transition: color var(--transition-fast);
}

.footer__link:hover {
  color: var(--color-primary);
  text-decoration: underline;
}

.footer__copyright {
  text-align: center;
  color: var(--color-text-muted);
  font-size: var(--font-size-xs);
  margin: 0;
  padding-top: var(--space-md);
  border-top: 1px solid var(--color-border-light);
}

/* Responsive */
@media (max-width: 900px) {
  .hero__content {
    grid-template-columns: 1fr;
  }
  
  .hero__card {
    max-width: 400px;
  }
  
  .audience-grid {
    grid-template-columns: 1fr;
  }
  
  .features-grid {
    grid-template-columns: 1fr;
  }

  .trust-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 650px) {
  .trust-grid {
    grid-template-columns: 1fr;
  }

  .differentiator__list {
    text-align: center;
  }

  .differentiator__list li {
    justify-content: center;
  }
}

@media (max-width: 600px) {
  .landing-header__nav .btn--ghost {
    display: none;
  }
  
  .hero__actions {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .footer__content {
    flex-direction: column;
    gap: var(--space-lg);
    text-align: center;
    align-items: center;
  }

  .footer__brand {
    align-items: center;
  }

  .footer__links {
    flex-direction: column;
    gap: var(--space-sm);
  }
}
</style>
