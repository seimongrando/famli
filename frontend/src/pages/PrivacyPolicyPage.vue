<!-- =============================================================================
  FAMLI - Página de Política de Privacidade
  =============================================================================
  Política de privacidade do serviço Famli.
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
      <h1>{{ t('legal.privacy.title') }}</h1>
    </header>

    <main class="legal-content">
      <p class="legal-updated">{{ t('legal.lastUpdated') }}: {{ t('legal.privacy.updatedDate') }}</p>

      <section>
        <h2>1. {{ t('legal.privacy.intro.title') }}</h2>
        <p>{{ t('legal.privacy.intro.text') }}</p>
      </section>

      <section>
        <h2>2. {{ t('legal.privacy.dataCollected.title') }}</h2>
        <p>{{ t('legal.privacy.dataCollected.text') }}</p>
        <ul>
          <li><strong>{{ t('legal.privacy.dataCollected.account') }}:</strong> {{ t('legal.privacy.dataCollected.accountDesc') }}</li>
          <li><strong>{{ t('legal.privacy.dataCollected.content') }}:</strong> {{ t('legal.privacy.dataCollected.contentDesc') }}</li>
          <li><strong>{{ t('legal.privacy.dataCollected.usage') }}:</strong> {{ t('legal.privacy.dataCollected.usageDesc') }}</li>
        </ul>
      </section>

      <section>
        <h2>3. {{ t('legal.privacy.dataUse.title') }}</h2>
        <p>{{ t('legal.privacy.dataUse.text') }}</p>
        <ul>
          <li>{{ t('legal.privacy.dataUse.item1') }}</li>
          <li>{{ t('legal.privacy.dataUse.item2') }}</li>
          <li>{{ t('legal.privacy.dataUse.item3') }}</li>
        </ul>
      </section>

      <section>
        <h2>4. {{ t('legal.privacy.security.title') }}</h2>
        <p>{{ t('legal.privacy.security.text') }}</p>
        <ul>
          <li>{{ t('legal.privacy.security.item1') }}</li>
          <li>{{ t('legal.privacy.security.item2') }}</li>
          <li>{{ t('legal.privacy.security.item3') }}</li>
        </ul>
      </section>

      <section>
        <h2>5. {{ t('legal.privacy.sharing.title') }}</h2>
        <p>{{ t('legal.privacy.sharing.text') }}</p>
      </section>

      <section>
        <h2>6. {{ t('legal.privacy.rights.title') }}</h2>
        <p>{{ t('legal.privacy.rights.text') }}</p>
        <ul>
          <li>{{ t('legal.privacy.rights.item1') }}</li>
          <li>{{ t('legal.privacy.rights.item2') }}</li>
          <li>{{ t('legal.privacy.rights.item3') }}</li>
          <li>{{ t('legal.privacy.rights.item4') }}</li>
        </ul>
      </section>

      <section>
        <h2>7. {{ t('legal.privacy.cookies.title') }}</h2>
        <p>{{ t('legal.privacy.cookies.text') }}</p>
      </section>

      <section>
        <h2>8. {{ t('legal.privacy.changes.title') }}</h2>
        <p>{{ t('legal.privacy.changes.text') }}</p>
      </section>

      <section>
        <h2>9. {{ t('legal.privacy.contact.title') }}</h2>
        <p>{{ t('legal.privacy.contact.text') }}</p>
        <p><strong>Email:</strong> privacidade@famli.net</p>
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


