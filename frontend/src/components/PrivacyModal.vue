<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const emit = defineEmits(['close'])
</script>

<template>
  <div class="modal-backdrop" @click.self="emit('close')">
    <div class="modal">
      <div class="modal__header">
        <h2 class="modal__title">🔒 {{ t('privacy.title') }}</h2>
        <p class="modal__description">
          {{ t('privacy.subtitle') }}
        </p>
      </div>
      
      <div class="privacy-content">
        <section class="privacy-section">
          <h3 class="privacy-section__title">{{ t('privacy.whatWeDo.title') }}</h3>
          <ul class="privacy-list">
            <li>{{ t('privacy.whatWeDo.item1') }}</li>
            <li>{{ t('privacy.whatWeDo.item2') }}</li>
            <li>{{ t('privacy.whatWeDo.item3') }}</li>
          </ul>
        </section>

        <section class="privacy-section">
          <h3 class="privacy-section__title">{{ t('privacy.whatWeDont.title') }}</h3>
          <ul class="privacy-list privacy-list--not">
            <li>{{ t('privacy.whatWeDont.item1') }}</li>
            <li>{{ t('privacy.whatWeDont.item2') }}</li>
            <li>{{ t('privacy.whatWeDont.item3') }}</li>
          </ul>
        </section>

        <section class="privacy-section">
          <h3 class="privacy-section__title">{{ t('privacy.guardians.title') }}</h3>
          <p class="privacy-text">
            {{ t('privacy.guardians.text1', { bold: '' }) }}
            <strong>{{ t('privacy.guardians.text1Bold') }}</strong>
          </p>
          <p class="privacy-text">
            {{ t('privacy.guardians.text2') }}
          </p>
        </section>

        <section class="privacy-section">
          <h3 class="privacy-section__title">{{ t('privacy.rights.title') }}</h3>
          <ul class="privacy-list">
            <li>{{ t('privacy.rights.item1') }}</li>
            <li>{{ t('privacy.rights.item2') }}</li>
            <li>{{ t('privacy.rights.item3') }}</li>
            <li>{{ t('privacy.rights.item4') }}</li>
          </ul>
        </section>
      </div>

      <div class="modal__footer">
        <button class="btn btn--primary btn--full" @click="emit('close')">
          {{ t('privacy.understood') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.privacy-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
  max-height: 60vh;
  overflow-y: auto;
}

.privacy-section {
  padding-bottom: var(--space-lg);
  border-bottom: 1px solid var(--color-border-light);
}

.privacy-section:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.privacy-section__title {
  font-size: var(--font-size-base);
  margin: 0 0 var(--space-sm);
}

.privacy-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.privacy-list li {
  padding-left: var(--space-lg);
  position: relative;
  color: var(--color-text-soft);
}

.privacy-list li::before {
  content: "✓";
  position: absolute;
  left: 0;
  color: var(--color-success);
  font-weight: 700;
}

.privacy-list--not li::before {
  content: "✗";
  color: var(--color-danger);
}

.privacy-text {
  color: var(--color-text-soft);
  font-size: var(--font-size-sm);
  margin: 0 0 var(--space-sm);
}

.privacy-text:last-child {
  margin-bottom: 0;
}
</style>
