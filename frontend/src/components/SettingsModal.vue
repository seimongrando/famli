<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLocale, availableLocales, getLocale } from '../i18n'

const { t } = useI18n()
const emit = defineEmits(['close'])

const settings = ref({
  emergency_protocol_enabled: false,
  notifications_enabled: true,
  theme: 'light'
})

const currentLocale = ref(getLocale())
const saving = ref(false)

onMounted(async () => {
  try {
    const res = await fetch('/api/settings', { credentials: 'include' })
    if (res.ok) {
      const data = await res.json()
      settings.value = { ...settings.value, ...data }
    }
  } catch (e) {
    // Usar defaults
  }
})

function changeLocale(code) {
  currentLocale.value = code
  setLocale(code)
}

async function save() {
  saving.value = true
  try {
    await fetch('/api/settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(settings.value)
    })
  } catch (e) {
    // Erro silencioso
  } finally {
    saving.value = false
    emit('close')
  }
}
</script>

<template>
  <div class="modal-backdrop" @click.self="emit('close')">
    <div class="modal">
      <div class="modal__header">
        <h2 class="modal__title">⚙️ {{ t('settings.title') }}</h2>
      </div>
      
      <div class="settings-form">
        <!-- Language -->
        <div class="setting-item">
          <div class="setting-item__content">
            <h3 class="setting-item__title">{{ t('settings.language.title') }}</h3>
            <p class="setting-item__description">
              {{ t('settings.language.description') }}
            </p>
          </div>
          <div class="language-options">
            <button
              v-for="loc in availableLocales"
              :key="loc.code"
              :class="['lang-btn', { 'lang-btn--active': currentLocale === loc.code }]"
              @click="changeLocale(loc.code)"
            >
              <span class="lang-btn__flag">{{ loc.flag }}</span>
              <span class="lang-btn__name">{{ loc.name }}</span>
            </button>
          </div>
        </div>

        <!-- Emergency Protocol -->
        <div class="setting-item">
          <div class="setting-item__content">
            <h3 class="setting-item__title">{{ t('settings.emergency.title') }}</h3>
            <p class="setting-item__description">
              {{ t('settings.emergency.description') }}
            </p>
          </div>
          <label class="toggle">
            <input 
              type="checkbox" 
              v-model="settings.emergency_protocol_enabled"
              class="toggle__input"
            />
            <span class="toggle__slider"></span>
          </label>
        </div>

        <!-- Notifications -->
        <div class="setting-item">
          <div class="setting-item__content">
            <h3 class="setting-item__title">{{ t('settings.notifications.title') }}</h3>
            <p class="setting-item__description">
              {{ t('settings.notifications.description') }}
            </p>
          </div>
          <label class="toggle">
            <input 
              type="checkbox" 
              v-model="settings.notifications_enabled"
              class="toggle__input"
            />
            <span class="toggle__slider"></span>
          </label>
        </div>
      </div>

      <div class="modal__footer">
        <button class="btn btn--ghost" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button class="btn btn--primary" @click="save" :disabled="saving">
          {{ saving ? t('common.loading') : t('common.save') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.settings-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-lg);
  padding-bottom: var(--space-lg);
  border-bottom: 1px solid var(--color-border-light);
}

.setting-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.setting-item__title {
  font-size: var(--font-size-base);
  margin: 0 0 var(--space-xs);
}

.setting-item__description {
  font-size: var(--font-size-sm);
  color: var(--color-text-soft);
  margin: 0;
}

/* Language Options */
.language-options {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.lang-btn {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  background: var(--color-bg-warm);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.lang-btn:hover {
  border-color: var(--color-primary);
}

.lang-btn--active {
  background: var(--color-primary-soft);
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.lang-btn__flag {
  font-size: 1.25rem;
}

/* Toggle Switch */
.toggle {
  position: relative;
  display: inline-block;
  width: 52px;
  height: 28px;
  flex-shrink: 0;
}

.toggle__input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle__slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--color-border);
  transition: 0.3s;
  border-radius: 28px;
}

.toggle__slider:before {
  position: absolute;
  content: "";
  height: 22px;
  width: 22px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.toggle__input:checked + .toggle__slider {
  background-color: var(--color-primary);
}

.toggle__input:checked + .toggle__slider:before {
  transform: translateX(24px);
}
</style>
