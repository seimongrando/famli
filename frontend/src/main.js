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

const routes = [
  { path: '/', name: 'landing', component: LandingPage },
  { path: '/entrar', name: 'auth', component: AuthPage },
  { path: '/minha-caixa', name: 'dashboard', component: DashboardPage, meta: { requiresAuth: true } },
  { path: '/admin', name: 'admin', component: AdminPage, meta: { requiresAuth: true } },
  { path: '/perfil', name: 'profile', component: ProfilePage, meta: { requiresAuth: true } },
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

// Navigation guard para rotas protegidas
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  
  // Se ainda não verificou a sessão, tentar verificar
  if (!authStore.user) {
    await authStore.checkSession()
  }
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'auth' })
  } else if (to.name === 'auth' && authStore.isAuthenticated) {
    // Se já está logado e tenta acessar /entrar, redirecionar para dashboard
    next({ name: 'dashboard' })
  } else {
    next()
  }
})

app.mount('#app')
