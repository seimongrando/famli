<!-- =============================================================================
  FAMLI - Página de Termos de Uso
  =============================================================================
  Termos de uso do serviço Famli.
============================================================================== -->

<script setup>
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useLocalizedRoutes } from '../composables/useLocalizedRoutes'

const { t } = useI18n()
const router = useRouter()
const { paths } = useLocalizedRoutes()

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push(paths.value.landing)
  }
}
</script>

<template>
  <div class="legal-page">
    <header class="legal-header">
      <button @click="goBack" class="btn btn--ghost">
        ← {{ t('common.back') }}
      </button>
      <h1>{{ t('legal.terms.title') }}</h1>
    </header>

    <main class="legal-content">
      <p class="legal-updated">{{ t('legal.lastUpdated') }}: {{ t('legal.terms.updatedDate') }}</p>

      <section>
        <h2>1. {{ t('legal.terms.acceptance.title') }}</h2>
        <p>{{ t('legal.terms.acceptance.text') }}</p>
      </section>

      <section>
        <h2>2. {{ t('legal.terms.service.title') }}</h2>
        <p>{{ t('legal.terms.service.text') }}</p>
      </section>

      <section>
        <h2>3. {{ t('legal.terms.account.title') }}</h2>
        <p>{{ t('legal.terms.account.text') }}</p>
      </section>

      <section>
        <h2>4. {{ t('legal.terms.content.title') }}</h2>
        <p>{{ t('legal.terms.content.text') }}</p>
      </section>

      <section>
        <h2>5. {{ t('legal.terms.prohibited.title') }}</h2>
        <p>{{ t('legal.terms.prohibited.text') }}</p>
        <ul>
          <li>{{ t('legal.terms.prohibited.item1') }}</li>
          <li>{{ t('legal.terms.prohibited.item2') }}</li>
          <li>{{ t('legal.terms.prohibited.item3') }}</li>
        </ul>
      </section>

      <section>
        <h2>6. {{ t('legal.terms.termination.title') }}</h2>
        <p>{{ t('legal.terms.termination.text') }}</p>
      </section>

      <section>
        <h2>7. {{ t('legal.terms.liability.title') }}</h2>
        <p>{{ t('legal.terms.liability.text') }}</p>
      </section>

      <section>
        <h2>8. {{ t('legal.terms.changes.title') }}</h2>
        <p>{{ t('legal.terms.changes.text') }}</p>
      </section>

      <section>
        <h2>9. {{ t('legal.terms.contact.title') }}</h2>
        <p>{{ t('legal.terms.contact.text') }}</p>
        <p><strong>Email:</strong> famli@famli.me</p>
      </section>
    </main>
  </div>
</template>

<style scoped>
.legal-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding: var(--space-xl);
}

.legal-header {
  max-width: 800px;
  margin: 0 auto var(--space-xl);
  display: flex;
  align-items: center;
  gap: var(--space-lg);
}

.legal-header h1 {
  font-size: var(--font-size-2xl);
  margin: 0;
}

.legal-content {
  max-width: 800px;
  margin: 0 auto;
  background: var(--color-card);
  padding: var(--space-2xl);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
}

.legal-updated {
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
  margin-bottom: var(--space-xl);
}

section {
  margin-bottom: var(--space-xl);
}

section h2 {
  font-size: var(--font-size-lg);
  margin-bottom: var(--space-md);
  color: var(--color-primary);
}

section p {
  line-height: 1.8;
  margin-bottom: var(--space-md);
}

section ul {
  margin-left: var(--space-lg);
  margin-bottom: var(--space-md);
}

section li {
  line-height: 1.8;
  margin-bottom: var(--space-sm);
}
</style>


