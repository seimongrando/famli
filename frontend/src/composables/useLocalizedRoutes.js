// =============================================================================
// FAMLI - Rotas Localizadas
// =============================================================================
// Composable para retornar URLs baseadas no idioma atual do usuário.
// 
// Uso:
//   const { getPath } = useLocalizedRoutes()
//   getPath('dashboard') // retorna '/minha-caixa' em pt-BR ou '/my-box' em en
// =============================================================================

import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

// Mapeamento de rotas por idioma
const routePaths = {
  'pt-BR': {
    landing: '/',
    auth: '/entrar',
    dashboard: '/minha-caixa',
    admin: '/administracao',
    profile: '/perfil'
  },
  'en': {
    landing: '/',
    auth: '/login',
    dashboard: '/my-box',
    admin: '/admin',
    profile: '/profile'
  }
}

// Idioma padrão
const defaultLocale = 'pt-BR'

export function useLocalizedRoutes() {
  const { locale } = useI18n()

  // Retorna o path localizado para uma rota
  const getPath = (routeName) => {
    const currentLocale = locale.value || defaultLocale
    const paths = routePaths[currentLocale] || routePaths[defaultLocale]
    return paths[routeName] || routePaths[defaultLocale][routeName] || '/'
  }

  // Paths computados para uso direto no template
  const paths = computed(() => {
    const currentLocale = locale.value || defaultLocale
    return routePaths[currentLocale] || routePaths[defaultLocale]
  })

  return {
    getPath,
    paths
  }
}

