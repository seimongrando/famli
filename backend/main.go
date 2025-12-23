// =============================================================================
// FAMLI - Backend Server
// =============================================================================
// Este é o ponto de entrada principal do servidor Famli.
//
// Funcionalidades:
// - API REST para gerenciamento de dados (autenticação, itens, guardiões)
// - Integração com WhatsApp via Twilio
// - Servir frontend estático (SPA)
//
// Segurança implementada (OWASP Top 10):
// - Rate limiting (A04)
// - Headers de segurança (A05)
// - Validação de inputs (A03)
// - Criptografia de dados sensíveis (A02)
// - Auditoria de eventos (A09)
//
// Variáveis de ambiente:
// - PORT: porta do servidor (padrão: 8080)
// - STATIC_DIR: diretório do frontend buildado
// - JWT_SECRET: segredo para tokens JWT (mínimo 32 caracteres em produção)
// - ENCRYPTION_KEY: chave para criptografar dados sensíveis
// - ENV: ambiente (development, production)
// - TWILIO_*: configurações do WhatsApp
// =============================================================================

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"famli/internal/admin"
	"famli/internal/analytics"
	"famli/internal/auth"
	"famli/internal/box"
	"famli/internal/feedback"
	"famli/internal/guardian"
	"famli/internal/guide"
	"famli/internal/i18n"
	"famli/internal/oauth"
	"famli/internal/security"
	"famli/internal/settings"
	"famli/internal/share"
	"famli/internal/storage"
	"famli/internal/whatsapp"
)

func main() {
	// =========================================================================
	// CONFIGURAÇÃO
	// =========================================================================

	// Ambiente
	env := getenv("ENV", "development")
	isDev := env == "development"

	// Variáveis de ambiente com valores padrão para desenvolvimento
	port := getenv("PORT", "8080")
	staticDir := getenv("STATIC_DIR", filepath.Join("..", "frontend", "dist"))
	jwtSecret := getenv("JWT_SECRET", "famli-dev-secret-change-in-production")
	encryptionKey := getenv("ENCRYPTION_KEY", "famli-encryption-key-change-in-prod")

	// Validar segredo JWT em produção
	if !isDev && len(jwtSecret) < 32 {
		log.Fatal("❌ JWT_SECRET deve ter pelo menos 32 caracteres em produção")
	}

	// Configuração do WhatsApp/Twilio
	whatsappConfig := &whatsapp.Config{
		TwilioAccountSid:  getenv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:   getenv("TWILIO_AUTH_TOKEN", ""),
		TwilioPhoneNumber: getenv("TWILIO_PHONE_NUMBER", ""),
		WebhookBaseURL:    getenv("WEBHOOK_BASE_URL", "http://localhost:8080"),
		Enabled:           getenv("TWILIO_ACCOUNT_SID", "") != "",
	}

	// Configuração do OAuth (Google, Apple)
	oauthConfig := &oauth.Config{
		GoogleClientID:  getenv("GOOGLE_CLIENT_ID", ""),
		AppleClientID:   getenv("APPLE_CLIENT_ID", ""),
		AppleTeamID:     getenv("APPLE_TEAM_ID", ""),
		AppleKeyID:      getenv("APPLE_KEY_ID", ""),
		ApplePrivateKey: getenv("APPLE_PRIVATE_KEY", ""),
	}

	// =========================================================================
	// LOG DE INICIALIZAÇÃO
	// =========================================================================

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("🏠 Famli - Ambiente: %s", env)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if whatsappConfig.Enabled {
		log.Println("📱 WhatsApp: habilitado")
	} else {
		log.Println("📱 WhatsApp: desabilitado")
	}

	if oauthConfig.GoogleClientID != "" {
		log.Println("🔐 Google OAuth: habilitado")
	}
	if oauthConfig.AppleClientID != "" {
		log.Println("🍎 Apple Sign In: habilitado")
	}

	// =========================================================================
	// VERIFICAÇÃO DO FRONTEND
	// =========================================================================

	indexPath := filepath.Join(staticDir, "index.html")
	frontendBuilt := true
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		frontendBuilt = false
		log.Printf("⚠️  Frontend não encontrado em %s", staticDir)
	}

	// =========================================================================
	// INICIALIZAÇÃO DOS SERVIÇOS
	// =========================================================================

	// Verificar se há DATABASE_URL para usar PostgreSQL
	databaseURL := getenv("DATABASE_URL", "")
	var store storage.Store
	var storageType string

	if databaseURL != "" {
		// Usar PostgreSQL em produção
		pgStore, err := storage.NewPostgresStore(databaseURL)
		if err != nil {
			log.Fatalf("❌ Erro ao conectar ao PostgreSQL: %v", err)
		}
		store = pgStore
		storageType = "PostgreSQL"
		log.Println("💾 Storage: PostgreSQL")

	} else {
		// Usar memória em desenvolvimento
		store = storage.NewMemoryStore()
		storageType = "Memory"
		log.Println("💾 Storage: Memória (dados serão perdidos ao reiniciar)")
	}

	// Limpeza automática de logs antigos (economizar espaço)
	retentionDays := getenvInt("LOG_RETENTION_DAYS", 30)
	cleanupIntervalHours := getenvInt("LOG_CLEANUP_INTERVAL_HOURS", 24)
	if err := store.CleanupOldLogs(retentionDays); err != nil {
		log.Printf("⚠️  Erro na limpeza de logs: %v", err)
	} else {
		log.Printf("🧹 Logs antigos (>%d dias): limpos", retentionDays)
	}
	if err := store.CleanupExpiredPasswordResetTokens(); err != nil {
		log.Printf("⚠️  Erro na limpeza de tokens de senha: %v", err)
	}
	if cleanupIntervalHours > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(cleanupIntervalHours) * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				if err := store.CleanupOldLogs(retentionDays); err != nil {
					log.Printf("⚠️  Erro na limpeza periódica de logs: %v", err)
				}
				if err := store.CleanupExpiredPasswordResetTokens(); err != nil {
					log.Printf("⚠️  Erro na limpeza periódica de tokens de senha: %v", err)
				}
			}
		}()
	}

	// Encryptor para dados sensíveis
	encryptor, err := security.NewEncryptor(encryptionKey)
	if err != nil {
		log.Printf("⚠️  Criptografia não configurada: %v", err)
	} else {
		log.Println("🔐 Criptografia: habilitada")
		_ = encryptor // TODO: Usar encryptor no box handler para dados sensíveis
	}

	// Handlers organizados por domínio
	authHandler := auth.NewHandler(store, jwtSecret)
	boxHandler := box.NewHandler(store)
	guardianHandler := guardian.NewHandler(store)
	guideHandler := guide.NewHandler(store)
	settingsHandler := settings.NewHandler(store)
	adminHandler := admin.NewHandler(store, storageType)
	feedbackHandler := feedback.NewHandler(store)
	analyticsHandler := analytics.NewHandler(store)
	oauthHandler := oauth.NewHandler(store, jwtSecret, oauthConfig)
	shareHandler := share.NewHandler(store)

	// Serviço e handler do WhatsApp
	whatsappService := whatsapp.NewService(store, whatsappConfig)
	whatsappHandler := whatsapp.NewHandler(whatsappService, whatsappConfig)

	// Rate limiters
	apiLimiter := security.NewRateLimiter(security.APIRateLimit)
	webhookLimiter := security.NewRateLimiter(security.WebhookRateLimit)

	// =========================================================================
	// CONFIGURAÇÃO DO ROUTER
	// =========================================================================

	r := chi.NewRouter()

	// ─────────────────────────────────────────────────────────────────────────
	// MIDDLEWARES GLOBAIS
	// ─────────────────────────────────────────────────────────────────────────

	// Request ID para rastreamento
	r.Use(chimiddleware.RequestID)

	// IP real do cliente (quando atrás de proxy)
	r.Use(chimiddleware.RealIP)

	// Logger de requisições (com redação de tokens)
	r.Use(security.RedactingLogger())

	// Recuperar de panics
	r.Use(chimiddleware.Recoverer)

	// Headers de segurança (OWASP A05)
	var headersConfig security.SecurityHeadersConfig
	if isDev {
		headersConfig = security.DevelopmentSecurityHeadersConfig()
	} else {
		headersConfig = security.DefaultSecurityHeadersConfig()
	}
	r.Use(security.HeadersMiddleware(headersConfig))

	// CORS - Cross-Origin Resource Sharing
	allowedOrigins := []string{"http://localhost:5173", "http://localhost:8080"}
	if !isDev {
		allowedOrigins = append(allowedOrigins, "https://famli.net", "https://www.famli.net")
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Accept-Language"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// =========================================================================
	// ROTAS DA API
	// =========================================================================

	r.Route("/api", func(api chi.Router) {
		// Rate limiting para API (OWASP A04)
		api.Use(apiLimiter.Middleware(security.GetClientIP))

		// ─────────────────────────────────────────────────────────────────────
		// ROTAS PÚBLICAS (sem autenticação)
		// ─────────────────────────────────────────────────────────────────────

		// Health check público (para load balancers)
		api.Get("/health", adminHandler.PublicHealth)

		// Autenticação (rate limit adicional no handler)
		api.Post("/auth/register", authHandler.Register)
		api.Post("/auth/login", authHandler.Login)

		// Recuperação de senha
		api.Post("/auth/forgot-password", authHandler.ForgotPassword)
		api.Post("/auth/reset-password", authHandler.ResetPassword)

		// OAuth - Login Social (Google, Apple)
		api.Post("/auth/oauth/google", oauthHandler.Google)
		api.Post("/auth/oauth/apple", oauthHandler.Apple)
		api.Get("/auth/oauth/status", oauthHandler.Status)

		// Webhook do WhatsApp (chamado pelo Twilio)
		api.Group(func(wh chi.Router) {
			wh.Use(webhookLimiter.Middleware(security.GetClientIP))
			wh.Get("/whatsapp/webhook", whatsappHandler.WebhookVerify)
			wh.Post("/whatsapp/webhook", whatsappHandler.Webhook)
		})

		// Status da integração WhatsApp
		api.Get("/whatsapp/status", whatsappHandler.Status)

		// ─────────────────────────────────────────────────────────────────────
		// ROTAS PROTEGIDAS (requerem autenticação JWT)
		// ─────────────────────────────────────────────────────────────────────

		api.Group(func(pr chi.Router) {
			// Middleware de autenticação JWT
			pr.Use(auth.JWTMiddleware(jwtSecret))
			// CSRF - validar origem para requisições mutantes
			pr.Use(security.CSRFMiddleware(allowedOrigins, isDev))

			// Autenticação
			pr.Get("/auth/me", authHandler.Me)
			pr.Post("/auth/logout", authHandler.Logout)

			// LGPD - Direitos do Titular
			pr.Delete("/auth/account", authHandler.DeleteAccount) // Direito ao esquecimento
			pr.Get("/auth/export", authHandler.ExportData)        // Direito à portabilidade

			// Caixa Famli
			pr.Get("/box/items", boxHandler.List)
			pr.Post("/box/items", boxHandler.Create)
			pr.Put("/box/items/{itemID}", boxHandler.Update)
			pr.Delete("/box/items/{itemID}", boxHandler.Delete)

			// Guardiões
			pr.Get("/guardians", guardianHandler.List)
			pr.Post("/guardians", guardianHandler.Create)
			pr.Put("/guardians/{guardianID}", guardianHandler.Update)
			pr.Delete("/guardians/{guardianID}", guardianHandler.Delete)

			// Guia Famli
			pr.Get("/guide/cards", guideHandler.ListCards)
			pr.Get("/guide/progress", guideHandler.GetProgress)
			pr.Post("/guide/progress/{cardID}", guideHandler.MarkCardProgress)

			// Configurações
			pr.Get("/settings", settingsHandler.Get)
			pr.Put("/settings", settingsHandler.Update)

			// Assistente
			pr.Post("/assistant", boxHandler.Assistant)

			// WhatsApp (vincular/desvincular)
			pr.Post("/whatsapp/link", whatsappHandler.Link)
			pr.Delete("/whatsapp/link", whatsappHandler.Unlink)

			// Feedback - Usuários podem enviar feedback
			pr.Post("/feedback", feedbackHandler.Create)

			// Analytics - Rastreamento de eventos
			pr.Post("/analytics/track", analyticsHandler.Track)

			// Share - Gerenciar links de compartilhamento
			pr.Post("/share/links", shareHandler.CreateLink)
			pr.Get("/share/links", shareHandler.ListLinks)
			pr.Delete("/share/links/{id}", shareHandler.DeleteLink)
		})

		// ─────────────────────────────────────────────────────────────────────
		// ROTAS PÚBLICAS DE COMPARTILHAMENTO (não requerem autenticação)
		// ─────────────────────────────────────────────────────────────────────

		api.Route("/shared", func(sr chi.Router) {
			// Rate limit para prevenir brute force em PINs
			sr.Use(apiLimiter.Middleware(security.GetClientIP))

			// Acessar conteúdo compartilhado
			sr.Get("/{token}", shareHandler.AccessShared)
			// Verificar PIN e acessar
			sr.Post("/{token}/verify", shareHandler.VerifyPIN)
		})

		// ─────────────────────────────────────────────────────────────────────
		// ROTA DE ACESSO DO GUARDIÃO (nova arquitetura integrada)
		// ─────────────────────────────────────────────────────────────────────
		api.Route("/guardian-access", func(sr chi.Router) {
			sr.Use(apiLimiter.Middleware(security.GetClientIP))
			sr.Get("/{token}", shareHandler.AccessGuardianView)
			sr.Post("/{token}/verify", shareHandler.VerifyGuardianPIN)
		})

		// ─────────────────────────────────────────────────────────────────────
		// ROTAS ADMINISTRATIVAS (requerem autenticação JWT + permissão admin)
		// ─────────────────────────────────────────────────────────────────────

		api.Route("/admin", func(ar chi.Router) {
			// Autenticação JWT obrigatória
			ar.Use(auth.JWTMiddleware(jwtSecret))
			// CSRF - validar origem para requisições mutantes
			ar.Use(security.CSRFMiddleware(allowedOrigins, isDev))
			// Verificação de permissão admin
			ar.Use(adminHandler.AdminOnly)

			// Dashboard com métricas
			ar.Get("/dashboard", adminHandler.Dashboard)
			// Health check detalhado
			ar.Get("/health", adminHandler.Health)
			// Lista de usuários
			ar.Get("/users", adminHandler.Users)
			// Atividade recente
			ar.Get("/activity", adminHandler.Activity)

			// Feedbacks - Gerenciamento de feedbacks dos usuários
			ar.Get("/feedbacks", feedbackHandler.List)
			ar.Get("/feedbacks/stats", feedbackHandler.GetStats)
			ar.Patch("/feedbacks/{id}", feedbackHandler.Update)

			// Analytics - Métricas de uso da aplicação
			ar.Get("/analytics/summary", analyticsHandler.GetSummary)
			ar.Get("/analytics/events", analyticsHandler.GetRecentEvents)
			ar.Get("/analytics/daily", analyticsHandler.GetDailyStats)
		})
	})

	// =========================================================================
	// SERVIR FRONTEND (SPA)
	// =========================================================================

	if frontendBuilt {
		fileServer := http.FileServer(http.Dir(staticDir))
		indexPath := filepath.Join(staticDir, "index.html")

		// Ler o index.html uma vez
		indexHTML, err := os.ReadFile(indexPath)
		if err != nil {
			log.Printf("⚠️  Erro ao ler index.html: %v", err)
		}

		r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			urlPath := req.URL.Path
			filePath := filepath.Join(staticDir, urlPath)

			setStaticCacheHeaders := func(path string, isHTML bool) {
				switch {
				case isHTML:
					w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
				case path == "/sw.js" || path == "/service-worker.js" || strings.HasSuffix(path, "manifest.webmanifest") || path == "/version.json":
					w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
				case strings.HasPrefix(path, "/assets/"):
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				default:
					w.Header().Set("Cache-Control", "public, max-age=86400")
				}
			}

			// Verificar se é uma rota de página (não um arquivo estático)
			// Se termina em / ou não tem extensão, é uma rota de página SPA
			isPageRoute := urlPath == "/" ||
				(!strings.Contains(filepath.Base(urlPath), ".") &&
					!strings.HasPrefix(urlPath, "/assets/") &&
					!strings.HasPrefix(urlPath, "/icons/"))

			// Se é uma rota de página, servir index.html com meta tags localizadas
			if isPageRoute {
				// Detectar idioma preferido
				lang := i18n.GetPreferredLanguage(req)

				// Injetar meta tags no idioma correto
				localizedHTML := i18n.InjectMetaTags(string(indexHTML), lang)

				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				setStaticCacheHeaders(urlPath, true)
				io.WriteString(w, localizedHTML)
				return
			}

			// Para arquivos estáticos, verificar se existe
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				// Arquivo não existe, servir index.html (SPA fallback)
				lang := i18n.GetPreferredLanguage(req)
				localizedHTML := i18n.InjectMetaTags(string(indexHTML), lang)
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				setStaticCacheHeaders(urlPath, true)
				io.WriteString(w, localizedHTML)
				return
			}

			// Servir arquivo estático
			setStaticCacheHeaders(urlPath, false)
			fileServer.ServeHTTP(w, req)
		}))
	} else {
		r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, setupInstructionsHTML)
		}))
	}

	// =========================================================================
	// INICIAR SERVIDOR
	// =========================================================================

	log.Printf("🌐 Servidor: http://localhost:%s", port)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

// =============================================================================
// FUNÇÕES AUXILIARES
// =============================================================================

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}

// =============================================================================
// HTML DE INSTRUÇÕES
// =============================================================================

const setupInstructionsHTML = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Famli - Setup</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: 'Nunito', system-ui, -apple-system, sans-serif;
            background: linear-gradient(135deg, #faf8f5 0%, #f5f0e8 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .card {
            background: white;
            border-radius: 24px;
            padding: 48px;
            max-width: 600px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.1);
        }
        .logo { font-size: 48px; margin-bottom: 16px; }
        h1 { color: #2d5a47; font-size: 28px; margin-bottom: 8px; }
        .subtitle { color: #5c584f; margin-bottom: 32px; }
        .section { margin-bottom: 24px; }
        .section-title { 
            color: #2d5a47; 
            font-size: 18px; 
            font-weight: 700; 
            margin-bottom: 12px;
        }
        pre {
            background: #f5f0e8;
            padding: 16px;
            border-radius: 12px;
            overflow-x: auto;
            font-size: 14px;
            color: #2c2a26;
        }
        code { font-family: 'SF Mono', Monaco, monospace; }
        .api-status {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            background: #e8f4ee;
            color: #2d5a47;
            padding: 8px 16px;
            border-radius: 20px;
            font-size: 14px;
            font-weight: 600;
            margin-top: 24px;
        }
        .dot { 
            width: 8px; 
            height: 8px; 
            background: #3a8a5c; 
            border-radius: 50%;
            animation: pulse 2s infinite;
        }
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }
    </style>
</head>
<body>
    <div class="card">
        <div class="logo">🏠</div>
        <h1>Famli - Setup Necessário</h1>
        <p class="subtitle">O frontend precisa ser compilado antes de usar.</p>
        
        <div class="section">
            <div class="section-title">🚀 Setup Rápido</div>
            <pre><code>make setup
make run</code></pre>
        </div>

        <div class="api-status">
            <span class="dot"></span>
            API Backend funcionando em /api
        </div>
    </div>
</body>
</html>`
