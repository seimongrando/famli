<script setup>
// O App.vue agora é simples, já que a verificação de sessão
// é feita pelo navigation guard no main.js
import CookieConsent from './components/CookieConsent.vue'
</script>

<template>
  <router-view />
  <CookieConsent />
</template>
