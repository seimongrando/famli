// =============================================================================
// FAMLI - Headers de Segurança HTTP
// =============================================================================
// Este módulo configura headers de segurança HTTP para proteger contra:
//
// OWASP A05:2021 – Security Misconfiguration
// - XSS (Cross-Site Scripting)
// - Clickjacking
// - MIME sniffing
// - Information disclosure
//
// Headers implementados:
// - Content-Security-Policy (CSP)
// - X-Content-Type-Options
// - X-Frame-Options
// - X-XSS-Protection
// - Strict-Transport-Security (HSTS)
// - Referrer-Policy
// - Permissions-Policy
// =============================================================================

package security

import (
	"net/http"
	"strings"
)

// =============================================================================
// CONFIGURAÇÃO
// =============================================================================

// SecurityHeadersConfig define configuração de headers de segurança
type SecurityHeadersConfig struct {
	// EnableHSTS habilita HTTP Strict Transport Security
	// IMPORTANTE: Só habilite em produção com HTTPS configurado
	EnableHSTS bool

	// HSTSMaxAge é o tempo em segundos para HSTS (padrão: 1 ano)
	HSTSMaxAge int

	// EnableCSP habilita Content-Security-Policy
	EnableCSP bool

	// CSPDirectives são as diretivas CSP customizadas
	// Se vazio, usa as diretivas padrão
	CSPDirectives string

	// FrameOptions define X-Frame-Options (DENY, SAMEORIGIN, ou vazio)
	FrameOptions string

	// ContentTypeOptions define X-Content-Type-Options (nosniff ou vazio)
	ContentTypeOptions string

	// ReferrerPolicy define a política de referrer
	ReferrerPolicy string

	// IsDevelopment indica modo de desenvolvimento (relaxa algumas políticas)
	IsDevelopment bool
}

// DefaultSecurityHeadersConfig retorna configuração padrão para produção
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		EnableHSTS:         true,
		HSTSMaxAge:         31536000, // 1 ano
		EnableCSP:          true,
		FrameOptions:       "DENY",
		ContentTypeOptions: "nosniff",
		ReferrerPolicy:     "strict-origin-when-cross-origin",
		IsDevelopment:      false,
	}
}

// DevelopmentSecurityHeadersConfig retorna configuração para desenvolvimento
func DevelopmentSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		EnableHSTS:         false, // Não usar HSTS em dev (sem HTTPS)
		EnableCSP:          true,
		FrameOptions:       "SAMEORIGIN",
		ContentTypeOptions: "nosniff",
		ReferrerPolicy:     "no-referrer-when-downgrade",
		IsDevelopment:      true,
	}
}

// =============================================================================
// MIDDLEWARE
// =============================================================================

// HeadersMiddleware retorna um middleware que adiciona headers de segurança
//
// Parâmetros:
//   - config: configuração de headers
//
// Retorna:
//   - func(http.Handler) http.Handler: middleware
func HeadersMiddleware(config SecurityHeadersConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ─────────────────────────────────────────────────────────────
			// X-Content-Type-Options
			// ─────────────────────────────────────────────────────────────
			// Previne MIME sniffing que pode levar a XSS
			if config.ContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", config.ContentTypeOptions)
			}

			// ─────────────────────────────────────────────────────────────
			// X-Frame-Options
			// ─────────────────────────────────────────────────────────────
			// Previne clickjacking
			if config.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", config.FrameOptions)
			}

			// ─────────────────────────────────────────────────────────────
			// X-XSS-Protection
			// ─────────────────────────────────────────────────────────────
			// Ativa filtro XSS do navegador (legado, mas não prejudica)
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// ─────────────────────────────────────────────────────────────
			// Referrer-Policy
			// ─────────────────────────────────────────────────────────────
			// Controla informações enviadas no header Referer
			if config.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
			}

			// ─────────────────────────────────────────────────────────────
			// Permissions-Policy
			// ─────────────────────────────────────────────────────────────
			// Controla quais APIs o navegador pode usar
			w.Header().Set("Permissions-Policy", buildPermissionsPolicy())

			// ─────────────────────────────────────────────────────────────
			// Strict-Transport-Security (HSTS)
			// ─────────────────────────────────────────────────────────────
			// Força HTTPS em acessos futuros
			if config.EnableHSTS && !config.IsDevelopment {
				w.Header().Set("Strict-Transport-Security",
					buildHSTSValue(config.HSTSMaxAge))
			}

			// ─────────────────────────────────────────────────────────────
			// Content-Security-Policy (CSP)
			// ─────────────────────────────────────────────────────────────
			// Controla de onde recursos podem ser carregados
			if config.EnableCSP {
				csp := config.CSPDirectives
				if csp == "" {
					csp = buildDefaultCSP(config.IsDevelopment)
				}
				w.Header().Set("Content-Security-Policy", csp)
			}

			// ─────────────────────────────────────────────────────────────
			// Cache-Control para respostas de API
			// ─────────────────────────────────────────────────────────────
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
			}

			// ─────────────────────────────────────────────────────────────
			// Remover headers que expõem informações
			// ─────────────────────────────────────────────────────────────
			// Nota: Alguns desses são adicionados pelo framework/servidor
			// e precisam ser removidos explicitamente

			next.ServeHTTP(w, r)
		})
	}
}

// =============================================================================
// BUILDERS
// =============================================================================

// buildDefaultCSP constrói a política CSP padrão
func buildDefaultCSP(isDevelopment bool) string {
	directives := []string{
		// Padrão: bloquear tudo que não for explicitamente permitido
		"default-src 'self'",

		// Scripts: próprio domínio
		// 'unsafe-eval' necessário para Vue.js runtime compilation
		// Em produção, idealmente usaríamos templates pré-compilados
		"script-src 'self' 'unsafe-inline' 'unsafe-eval'",

		// Estilos: próprio domínio + Google Fonts + inline (necessário para Vue)
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",

		// Fontes: próprio domínio + Google Fonts
		"font-src 'self' https://fonts.gstatic.com data:",

		// Imagens: próprio domínio + data URIs + https
		"img-src 'self' data: https: blob:",

		// Conexões: próprio domínio + Google Fonts (service worker) + WebSocket em dev
		"connect-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com" + conditionalCSP(isDevelopment, " ws://localhost:* wss://localhost:*"),

		// Frames: bloquear (prevenção de clickjacking adicional)
		"frame-ancestors 'none'",

		// Formulários: apenas próprio domínio
		"form-action 'self'",

		// Base URI: apenas próprio domínio
		"base-uri 'self'",

		// Object/Embed: bloquear
		"object-src 'none'",

		// Workers: próprio domínio (necessário para Service Worker/PWA)
		"worker-src 'self' blob:",

		// Manifest: próprio domínio (PWA manifest)
		"manifest-src 'self'",
	}

	// Só adicionar upgrade-insecure-requests em produção
	if !isDevelopment {
		directives = append(directives, "upgrade-insecure-requests")
	}

	return strings.Join(directives, "; ")
}

// buildHSTSValue constrói o valor do header HSTS
func buildHSTSValue(maxAge int) string {
	// Inclui subdomínios e permite preload list
	return "max-age=" + itoa(maxAge) + "; includeSubDomains; preload"
}

// buildPermissionsPolicy constrói a política de permissões
func buildPermissionsPolicy() string {
	// Desabilitar APIs potencialmente perigosas ou desnecessárias
	policies := []string{
		"accelerometer=()",          // Acelerômetro
		"camera=()",                 // Câmera
		"geolocation=()",            // Geolocalização
		"gyroscope=()",              // Giroscópio
		"magnetometer=()",           // Magnetômetro
		"microphone=()",             // Microfone
		"payment=()",                // Payment API
		"usb=()",                    // USB
		"interest-cohort=()",        // FLoC (tracking do Google)
		"autoplay=(self)",           // Autoplay apenas para self
		"fullscreen=(self)",         // Fullscreen apenas para self
		"picture-in-picture=(self)", // PiP apenas para self
	}

	return strings.Join(policies, ", ")
}

// =============================================================================
// FUNÇÕES AUXILIARES
// =============================================================================

// conditionalCSP adiciona string se condição for verdadeira
func conditionalCSP(condition bool, value string) string {
	if condition {
		return value
	}
	return ""
}

// itoa converte int para string (sem usar strconv para manter simples)
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	result := ""
	negative := n < 0
	if negative {
		n = -n
	}

	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}

	if negative {
		result = "-" + result
	}

	return result
}

// =============================================================================
// HEADERS ESPECÍFICOS POR TIPO DE RESPOSTA
// =============================================================================

// SetJSONHeaders configura headers para respostas JSON
func SetJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

// SetDownloadHeaders configura headers para download de arquivo
func SetDownloadHeaders(w http.ResponseWriter, filename string, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")
}

// SetNoCacheHeaders configura headers para não cachear
func SetNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}
