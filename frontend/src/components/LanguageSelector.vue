<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLocale, availableLocales } from '../i18n'

const { locale } = useI18n()
const isOpen = ref(false)

const currentLocale = computed(() => {
  return availableLocales.find(l => l.code === locale.value) || availableLocales[0]
})

function selectLocale(code) {
  setLocale(code)
  isOpen.value = false
}
</script>

<template>
  <div class="lang-selector">
    <button 
      class="lang-selector__button"
      @click="isOpen = !isOpen"
      :aria-expanded="isOpen"
    >
      <span class="lang-selector__flag">{{ currentLocale.flag }}</span>
      <span class="lang-selector__code">{{ currentLocale.code.split('-')[0].toUpperCase() }}</span>
      <svg class="lang-selector__arrow" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
      </svg>
    </button>
    
    <div v-if="isOpen" class="lang-selector__dropdown">
      <button
        v-for="loc in availableLocales"
        :key="loc.code"
        :class="['lang-selector__option', { 'lang-selector__option--active': loc.code === locale }]"
        @click="selectLocale(loc.code)"
      >
        <span class="lang-selector__flag">{{ loc.flag }}</span>
        <span class="lang-selector__name">{{ loc.name }}</span>
      </button>
    </div>
    
    <!-- Backdrop para fechar -->
    <div v-if="isOpen" class="lang-selector__backdrop" @click="isOpen = false"></div>
  </div>
</template>

<style scoped>
.lang-selector {
  position: relative;
}

.lang-selector__button {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  padding: var(--space-sm) var(--space-md);
  background: var(--color-bg-warm);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  color: var(--color-text);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.lang-selector__button:hover {
  background: var(--color-border);
}

.lang-selector__flag {
  font-size: 1.1rem;
}

.lang-selector__code {
  font-weight: 600;
}

.lang-selector__arrow {
  width: 16px;
  height: 16px;
  transition: transform var(--transition-fast);
}

.lang-selector__button[aria-expanded="true"] .lang-selector__arrow {
  transform: rotate(180deg);
}

.lang-selector__dropdown {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  overflow: hidden;
  z-index: 100;
  min-width: 140px;
}

.lang-selector__option {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  width: 100%;
  padding: var(--space-sm) var(--space-md);
  background: none;
  border: none;
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  color: var(--color-text);
  cursor: pointer;
  transition: background var(--transition-fast);
  text-align: left;
}

.lang-selector__option:hover {
  background: var(--color-bg-warm);
}

.lang-selector__option--active {
  background: var(--color-primary-soft);
  color: var(--color-primary);
}

.lang-selector__name {
  flex: 1;
}

.lang-selector__backdrop {
  position: fixed;
  inset: 0;
  z-index: 99;
}
</style>


