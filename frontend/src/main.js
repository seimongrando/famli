import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import i18n from './i18n'
import './styles/main.css'

// Pages
import LandingPage from './pages/LandingPage.vue'
import AuthPage from './pages/AuthPage.vue'
import DashboardPage from './pages/DashboardPage.vue'
import AdminPage from './pages/AdminPage.vue'
import ProfilePage from './pages/ProfilePage.vue'
import TermsPage from './pages/TermsPage.vue'
import PrivacyPolicyPage from './pages/PrivacyPolicyPage.vue'
import SharedPage from './pages/SharedPage.vue'
import ResetPasswordPage from './pages/ResetPasswordPage.vue'

// Rotas com aliases para suportar múltiplos idiomas
// URL principal em pt-BR, aliases em en
const routes = [
  { 
    path: '/', 
    name: 'landing', 
    component: LandingPage 
  },
  { 
    path: '/entrar', 
    alias: ['/login', '/sign-in'],
    name: 'auth', 
    component: AuthPage 
  },
  { 
    path: '/minha-caixa', 
    alias: ['/my-box', '/dashboard'],
    name: 'dashboard', 
    component: DashboardPage, 
    meta: { requiresAuth: true } 
  },
  { 
    path: '/administracao', 
    alias: ['/admin'],
    name: 'admin', 
    component: AdminPage, 
    meta: { requiresAuth: true } 
  },
  { 
    path: '/perfil', 
    alias: ['/profile', '/me'],
    name: 'profile', 
    component: ProfilePage, 
    meta: { requiresAuth: true } 
  },
  { 
    path: '/termos', 
    alias: ['/terms', '/terms-of-service'],
    name: 'terms', 
    component: TermsPage 
  },
  { 
    path: '/privacidade', 
    alias: ['/privacy', '/privacy-policy'],
    name: 'privacy', 
    component: PrivacyPolicyPage 
  },
  // Página pública de compartilhamento (para guardiões)
  { 
    path: '/compartilhado/:token', 
    alias: ['/shared/:token'],
    name: 'shared', 
    component: SharedPage 
  },
  // Recuperação de senha
  { 
    path: '/redefinir-senha', 
    alias: ['/reset-password', '/forgot-password'],
    name: 'reset-password', 
    component: ResetPasswordPage 
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

const pinia = createPinia()
const app = createApp(App)

app.use(pinia)
app.use(router)
app.use(i18n)

// Importar store após configurar o Pinia
import { useAuthStore } from './stores/auth'

// Mapeamento de rotas por idioma (duplicado do composable para uso no guard)
const routePaths = {
  'pt-BR': { auth: '/entrar', dashboard: '/minha-caixa' },
  'en': { auth: '/login', dashboard: '/my-box' }
}

// Função para obter path localizado no guard
function getLocalizedPath(routeName) {
  const locale = i18n.global.locale.value || 'pt-BR'
  const paths = routePaths[locale] || routePaths['pt-BR']
  return paths[routeName] || routePaths['pt-BR'][routeName]
}

// Navigation guard para rotas protegidas
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  
  // Se ainda não verificou a sessão, tentar verificar
  if (!authStore.user) {
    await authStore.checkSession()
  }
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next(getLocalizedPath('auth'))
  } else if (to.name === 'auth' && authStore.isAuthenticated) {
    // Se já está logado e tenta acessar /entrar, redirecionar para dashboard
    next(getLocalizedPath('dashboard'))
  } else {
    next()
  }
})

app.mount('#app')
