// =============================================================================
// Famli - Vite Configuration
// =============================================================================
// Este arquivo configura o Vite para build do frontend Vue.js + PWA.
//
// Funcionalidades:
// - Vue 3 com Single File Components
// - PWA com Service Worker (auto update)
// - Proxy para API em desenvolvimento
// =============================================================================

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'
import { execSync } from 'node:child_process'

function buildVersion() {
  const buildTime = new Date().toISOString()
  let commit = ''
  try {
    commit = execSync('git rev-parse --short HEAD').toString().trim()
  } catch (_) {
    commit = ''
  }
  const build = commit ? `${commit}-${Date.now()}` : `build-${Date.now()}`
  return { build, commit, buildTime }
}

export default defineConfig({
  plugins: [
    vue(),
    {
      name: 'generate-version-file',
      apply: 'build',
      generateBundle() {
        const version = buildVersion()
        this.emitFile({
          type: 'asset',
          fileName: 'version.json',
          source: JSON.stringify(version)
        })
      }
    },
    VitePWA({
      // =======================================================================
      // CONFIGURAÇÃO DO SERVICE WORKER
      // =======================================================================
      
      // 'autoUpdate': Atualiza automaticamente sem perguntar ao usuário
      // Isso garante que os usuários sempre tenham a versão mais recente
      registerType: 'autoUpdate',
      
      // Assets estáticos para incluir no precache
      includeAssets: ['famli.png', 'favicon.ico', 'logo.svg'],
      
      // =======================================================================
      // WORKBOX - Service Worker Configuration
      // =======================================================================
      workbox: {
        // Forçar ativação imediata do novo SW (não esperar abas fecharem)
        skipWaiting: true,
        
        // Assumir controle de clientes imediatamente
        clientsClaim: true,
        
        // Arquivos para precache
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff,woff2}'],
        
        // NÃO cachear fontes do Google via Service Worker
        // O browser já faz cache nativo e evita conflitos com CSP
        // Removido o runtimeCaching para fonts.googleapis.com
        
        runtimeCaching: [
          // Cache de API - Network First (busca na rede, fallback para cache)
          {
            urlPattern: /\/api\/.*/i,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-cache',
              expiration: {
                maxEntries: 50,
                maxAgeSeconds: 60 * 5 // 5 minutos
              },
              cacheableResponse: {
                statuses: [0, 200]
              }
            }
          },
          // Cache de imagens externas
          {
            urlPattern: /^https:\/\/.*\.(png|jpg|jpeg|svg|gif|webp)$/i,
            handler: 'CacheFirst',
            options: {
              cacheName: 'images-cache',
              expiration: {
                maxEntries: 100,
                maxAgeSeconds: 60 * 60 * 24 * 30 // 30 dias
              },
              cacheableResponse: {
                statuses: [0, 200]
              }
            }
          }
        ],
        
        // Limpar caches antigos automaticamente
        cleanupOutdatedCaches: true,
        
        // Não precachear arquivos muito grandes
        maximumFileSizeToCacheInBytes: 3 * 1024 * 1024 // 3MB
      },
      
      // =======================================================================
      // MANIFEST (PWA)
      // =======================================================================
      manifest: {
        name: 'Famli',
        short_name: 'Famli',
        description: 'Sua caixa segura de memórias, documentos e orientações para quem você ama.',
        theme_color: '#355d4a',
        background_color: '#faf8f5',
        display: 'standalone',
        orientation: 'portrait',
        scope: '/',
        start_url: '/',
        categories: ['lifestyle', 'productivity'],
        icons: [
          {
            src: '/icons/icon-72x72.png',
            sizes: '72x72',
            type: 'image/png',
            purpose: 'maskable any'
          },
          {
            src: '/icons/icon-96x96.png',
            sizes: '96x96',
            type: 'image/png',
            purpose: 'maskable any'
          },
          {
            src: '/icons/icon-128x128.png',
            sizes: '128x128',
            type: 'image/png',
            purpose: 'maskable any'
          },
          {
            src: '/icons/icon-144x144.png',
            sizes: '144x144',
            type: 'image/png',
            purpose: 'maskable any'
          },
          {
            src: '/icons/icon-152x152.png',
            sizes: '152x152',
            type: 'image/png',
            purpose: 'maskable any'
          },
          {
            src: '/icons/icon-192x192.png',
            sizes: '192x192',
            type: 'image/png',
            purpose: 'maskable any'
          },
          {
            src: '/icons/icon-384x384.png',
            sizes: '384x384',
            type: 'image/png',
            purpose: 'maskable any'
          },
          {
            src: '/icons/icon-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'maskable any'
          }
        ]
      },
      
      // =======================================================================
      // DESENVOLVIMENTO
      // =======================================================================
      devOptions: {
        enabled: false // Desabilitar SW em desenvolvimento para evitar problemas de cache
      }
    })
  ],
  
  // ===========================================================================
  // SERVIDOR DE DESENVOLVIMENTO
  // ===========================================================================
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  
  // ===========================================================================
  // BUILD
  // ===========================================================================
  build: {
    // Gerar sourcemaps apenas em desenvolvimento
    sourcemap: false,
    
    // Tamanho máximo de chunk antes de warning
    chunkSizeWarningLimit: 500
  }
})
