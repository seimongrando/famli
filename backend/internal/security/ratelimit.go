// =============================================================================
// FAMLI - Rate Limiter
// =============================================================================
// Este módulo implementa rate limiting para proteger contra:
//
// OWASP A04:2021 – Insecure Design
// - Ataques de força bruta
// - Denial of Service (DoS)
// - Credential stuffing
//
// Estratégias implementadas:
// - Rate limit por IP (geral)
// - Rate limit por usuário (login)
// - Sliding window algorithm
// - Bloqueio progressivo após falhas
// =============================================================================

package security

import (
	"net/http"
	"sync"
	"time"
)

// =============================================================================
// CONFIGURAÇÃO
// =============================================================================

// RateLimitConfig define configuração de rate limiting
type RateLimitConfig struct {
	// Requests é o número máximo de requisições permitidas
	Requests int

	// Window é a janela de tempo para contagem
	Window time.Duration

	// BlockDuration é o tempo de bloqueio após exceder o limite
	BlockDuration time.Duration
}

// Configurações padrão para diferentes endpoints
var (
	// DefaultRateLimit para endpoints gerais
	DefaultRateLimit = RateLimitConfig{
		Requests:      100,
		Window:        time.Minute,
		BlockDuration: time.Minute * 5,
	}

	// LoginRateLimit para tentativas de login (mais restritivo)
	LoginRateLimit = RateLimitConfig{
		Requests:      5,
		Window:        time.Minute,
		BlockDuration: time.Minute * 15,
	}

	// RegisterRateLimit para criação de contas
	RegisterRateLimit = RateLimitConfig{
		Requests:      3,
		Window:        time.Hour,
		BlockDuration: time.Hour,
	}

	// APIRateLimit para chamadas de API
	APIRateLimit = RateLimitConfig{
		Requests:      60,
		Window:        time.Minute,
		BlockDuration: time.Minute * 5,
	}

	// WebhookRateLimit para webhooks externos (mais permissivo)
	WebhookRateLimit = RateLimitConfig{
		Requests:      200,
		Window:        time.Minute,
		BlockDuration: time.Minute,
	}
)

// =============================================================================
// RATE LIMITER
// =============================================================================

// RateLimiter implementa rate limiting com sliding window
type RateLimiter struct {
	// config é a configuração do limiter
	config RateLimitConfig

	// clients armazena estado por identificador (IP, userID, etc.)
	clients map[string]*clientState

	// mu protege acesso concorrente
	mu sync.RWMutex

	// cleanupInterval define intervalo de limpeza de entradas antigas
	cleanupInterval time.Duration
}

// clientState armazena o estado de rate limit para um cliente
type clientState struct {
	// requests é o número de requisições na janela atual
	requests int

	// windowStart é o início da janela atual
	windowStart time.Time

	// blockedUntil indica até quando o cliente está bloqueado
	blockedUntil time.Time

	// failedAttempts conta tentativas falhas consecutivas
	failedAttempts int

	// lastRequest é o timestamp da última requisição
	lastRequest time.Time
}

// NewRateLimiter cria um novo rate limiter
//
// Parâmetros:
//   - config: configuração de limites
//
// Retorna:
//   - *RateLimiter: limiter configurado
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		config:          config,
		clients:         make(map[string]*clientState),
		cleanupInterval: time.Minute * 5,
	}

	// Iniciar goroutine de limpeza
	go rl.cleanup()

	return rl
}

// Allow verifica se uma requisição deve ser permitida
//
// Parâmetros:
//   - identifier: identificador do cliente (IP, userID, etc.)
//
// Retorna:
//   - bool: true se permitido, false se bloqueado
//   - time.Duration: tempo restante de bloqueio (se bloqueado)
func (rl *RateLimiter) Allow(identifier string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Obter ou criar estado do cliente
	state, exists := rl.clients[identifier]
	if !exists {
		state = &clientState{
			windowStart: now,
			lastRequest: now,
		}
		rl.clients[identifier] = state
	}

	// Verificar se está bloqueado
	if now.Before(state.blockedUntil) {
		return false, state.blockedUntil.Sub(now)
	}

	// Verificar se a janela expirou
	if now.Sub(state.windowStart) > rl.config.Window {
		// Resetar janela
		state.requests = 0
		state.windowStart = now
	}

	// Verificar limite
	if state.requests >= rl.config.Requests {
		// Bloquear cliente
		state.blockedUntil = now.Add(rl.config.BlockDuration)
		return false, rl.config.BlockDuration
	}

	// Permitir requisição
	state.requests++
	state.lastRequest = now
	return true, 0
}

// RecordFailure registra uma tentativa falha (ex: login incorreto)
// Aumenta progressivamente o tempo de bloqueio
//
// Parâmetros:
//   - identifier: identificador do cliente
func (rl *RateLimiter) RecordFailure(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	state, exists := rl.clients[identifier]
	if !exists {
		state = &clientState{
			windowStart: time.Now(),
		}
		rl.clients[identifier] = state
	}

	state.failedAttempts++

	// Bloqueio progressivo baseado em falhas
	// 3 falhas: 1 min, 5 falhas: 5 min, 10 falhas: 30 min, 15+: 1 hora
	var blockDuration time.Duration
	switch {
	case state.failedAttempts >= 15:
		blockDuration = time.Hour
	case state.failedAttempts >= 10:
		blockDuration = time.Minute * 30
	case state.failedAttempts >= 5:
		blockDuration = time.Minute * 5
	case state.failedAttempts >= 3:
		blockDuration = time.Minute
	}

	if blockDuration > 0 {
		state.blockedUntil = time.Now().Add(blockDuration)
	}
}

// RecordSuccess registra uma tentativa bem-sucedida
// Reseta o contador de falhas
//
// Parâmetros:
//   - identifier: identificador do cliente
func (rl *RateLimiter) RecordSuccess(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if state, exists := rl.clients[identifier]; exists {
		state.failedAttempts = 0
	}
}

// GetStatus retorna o status atual de rate limit para um cliente
func (rl *RateLimiter) GetStatus(identifier string) (remaining int, resetIn time.Duration, blocked bool) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	state, exists := rl.clients[identifier]
	if !exists {
		return rl.config.Requests, rl.config.Window, false
	}

	now := time.Now()

	// Verificar bloqueio
	if now.Before(state.blockedUntil) {
		return 0, state.blockedUntil.Sub(now), true
	}

	// Verificar janela
	elapsed := now.Sub(state.windowStart)
	if elapsed > rl.config.Window {
		return rl.config.Requests, rl.config.Window, false
	}

	remaining = rl.config.Requests - state.requests
	if remaining < 0 {
		remaining = 0
	}

	return remaining, rl.config.Window - elapsed, false
}

// cleanup remove entradas antigas periodicamente
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.config.Window * 2)

		for id, state := range rl.clients {
			// Remover se última requisição foi há muito tempo e não está bloqueado
			if state.lastRequest.Before(cutoff) && now.After(state.blockedUntil) {
				delete(rl.clients, id)
			}
		}
		rl.mu.Unlock()
	}
}

// =============================================================================
// MIDDLEWARE HTTP
// =============================================================================

// Middleware retorna um middleware HTTP para rate limiting
//
// Parâmetros:
//   - getIdentifier: função para extrair identificador da requisição (geralmente IP)
//
// Retorna:
//   - func(http.Handler) http.Handler: middleware
func (rl *RateLimiter) Middleware(getIdentifier func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identifier := getIdentifier(r)

			allowed, retryAfter := rl.Allow(identifier)

			// Adicionar headers de rate limit
			remaining, resetIn, _ := rl.GetStatus(identifier)
			w.Header().Set("X-RateLimit-Limit", string(rune(rl.config.Requests)))
			w.Header().Set("X-RateLimit-Remaining", string(rune(remaining)))
			w.Header().Set("X-RateLimit-Reset", string(rune(int(resetIn.Seconds()))))

			if !allowed {
				w.Header().Set("Retry-After", string(rune(int(retryAfter.Seconds()))))
				http.Error(w, `{"error":"Muitas requisições. Tente novamente em alguns minutos."}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// =============================================================================
// FUNÇÕES AUXILIARES
// =============================================================================

// GetClientIP extrai o IP real do cliente considerando proxies
func GetClientIP(r *http.Request) string {
	// Verificar X-Forwarded-For (quando atrás de proxy/load balancer)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Pegar o primeiro IP (cliente original)
		ips := splitAndTrim(xff, ",")
		if len(ips) > 0 {
			return ips[0]
		}
	}

	// Verificar X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback para RemoteAddr
	// Remover porta se presente
	ip := r.RemoteAddr
	if colonIdx := lastIndex(ip, ':'); colonIdx != -1 {
		ip = ip[:colonIdx]
	}

	return ip
}

// splitAndTrim divide string e remove espaços
func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	if s == "" {
		return nil
	}
	result := make([]string, 0)
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func lastIndex(s string, char byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == char {
			return i
		}
	}
	return -1
}
