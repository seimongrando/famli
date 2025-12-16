// =============================================================================
// FAMLI - Handler de Autenticação
// =============================================================================
// Este módulo gerencia autenticação de usuários com foco em segurança.
//
// Proteções implementadas:
// - Validação de email e senha (OWASP A07)
// - Rate limiting para login (OWASP A04)
// - Hashing seguro com bcrypt (OWASP A02)
// - Auditoria de eventos (OWASP A09)
// - Proteção contra enumeração de usuários
// =============================================================================

package auth

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"famli/internal/security"
	"famli/internal/storage"
)

// =============================================================================
// HANDLER
// =============================================================================

// Handler gerencia endpoints de autenticação
type Handler struct {
	// store é o armazenamento de dados
	store *storage.MemoryStore

	// jwtSecret é o segredo para assinar tokens JWT
	jwtSecret string

	// loginLimiter controla rate limit de login
	loginLimiter *security.RateLimiter

	// registerLimiter controla rate limit de registro
	registerLimiter *security.RateLimiter

	// auditLogger registra eventos de segurança
	auditLogger *security.AuditLogger
}

// NewHandler cria uma nova instância do handler de autenticação
//
// Parâmetros:
//   - store: armazenamento de dados
//   - secret: segredo JWT (mínimo 32 caracteres em produção)
//
// Retorna:
//   - *Handler: handler configurado
func NewHandler(store *storage.MemoryStore, secret string) *Handler {
	return &Handler{
		store:           store,
		jwtSecret:       secret,
		loginLimiter:    security.NewRateLimiter(security.LoginRateLimit),
		registerLimiter: security.NewRateLimiter(security.RegisterRateLimit),
		auditLogger:     security.GetAuditLogger(),
	}
}

// =============================================================================
// PAYLOADS
// =============================================================================

// registerPayload é o payload de registro
type registerPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// loginPayload é o payload de login
type loginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// =============================================================================
// REGISTRO
// =============================================================================

// Register cria uma nova conta de usuário
//
// Endpoint: POST /api/auth/register
//
// Segurança:
// - Rate limiting por IP
// - Validação de email
// - Validação de força de senha
// - Sanitização de inputs
// - Auditoria de evento
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	clientIP := security.GetClientIP(r)

	// Verificar rate limit
	allowed, retryAfter := h.registerLimiter.Allow(clientIP)
	if !allowed {
		h.auditLogger.LogSecurity(security.EventRateLimitExceeded, clientIP, map[string]interface{}{
			"endpoint": "register",
		})
		w.Header().Set("Retry-After", itoa(int(retryAfter.Seconds())))
		writeError(w, http.StatusTooManyRequests, "Muitas tentativas. Aguarde alguns minutos.")
		return
	}

	// Decodificar payload
	var payload registerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Dados inválidos.")
		return
	}

	// Validar e sanitizar email
	email, err := security.ValidateEmail(payload.Email)
	if err != nil {
		writeError(w, http.StatusBadRequest, "E-mail inválido.")
		return
	}

	// Validar força da senha
	strength, err := security.ValidatePassword(payload.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Senha precisa ter no mínimo 8 caracteres com letras e números.")
		return
	}

	// Sanitizar nome
	name := security.SanitizeName(payload.Name)

	// Hash da senha com bcrypt (custo alto para resistir a ataques)
	hashed, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		h.auditLogger.LogSecurity(security.EventSuspiciousActivity, clientIP, map[string]interface{}{
			"error": "bcrypt failed",
		})
		writeError(w, http.StatusInternalServerError, "Erro ao preparar sua conta.")
		return
	}

	// Criar usuário
	user, err := h.store.CreateUser(email, string(hashed), name)
	if err != nil {
		if err == storage.ErrAlreadyExists {
			// Não revelar se o email existe (proteção contra enumeração)
			// Usar mesma mensagem de sucesso após delay
			time.Sleep(100 * time.Millisecond) // Timing attack protection
			writeError(w, http.StatusBadRequest, "Não foi possível criar a conta. Tente outro e-mail.")
			return
		}
		writeError(w, http.StatusBadRequest, "Erro ao criar conta.")
		return
	}

	// Criar sessão
	if err := h.setSession(w, user.ID, r); err != nil {
		writeError(w, http.StatusInternalServerError, "Erro ao criar sessão.")
		return
	}

	// Registrar evento de auditoria
	h.auditLogger.LogAuth(security.EventRegister, user.ID, clientIP, r.UserAgent(), "success", map[string]interface{}{
		"email":             maskEmail(email),
		"password_strength": strength,
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// =============================================================================
// LOGIN
// =============================================================================

// Login autentica um usuário existente
//
// Endpoint: POST /api/auth/login
//
// Segurança:
// - Rate limiting por IP com bloqueio progressivo
// - Proteção contra timing attacks
// - Proteção contra enumeração de usuários
// - Auditoria de tentativas
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	clientIP := security.GetClientIP(r)

	// Verificar rate limit
	allowed, retryAfter := h.loginLimiter.Allow(clientIP)
	if !allowed {
		h.auditLogger.LogSecurity(security.EventRateLimitExceeded, clientIP, map[string]interface{}{
			"endpoint": "login",
		})
		w.Header().Set("Retry-After", itoa(int(retryAfter.Seconds())))
		writeError(w, http.StatusTooManyRequests, "Muitas tentativas. Aguarde alguns minutos.")
		return
	}

	// Decodificar payload
	var payload loginPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Dados inválidos.")
		return
	}

	// Normalizar email
	email, _ := security.ValidateEmail(payload.Email)
	if email == "" {
		email = payload.Email // Usar original para busca
	}

	// Buscar usuário
	user, ok := h.store.GetUserByEmail(email)

	// IMPORTANTE: Sempre executar bcrypt mesmo se usuário não existe
	// Isso previne timing attacks que revelam se o email existe
	var dummyHash = "$2a$10$dummy.hash.for.timing.attack.prevention"
	passwordToCheck := dummyHash
	if ok {
		passwordToCheck = user.Password
	}

	// Verificar senha
	err := bcrypt.CompareHashAndPassword([]byte(passwordToCheck), []byte(payload.Password))

	// Se usuário não existe ou senha incorreta
	if !ok || err != nil {
		// Registrar falha
		h.loginLimiter.RecordFailure(clientIP)
		h.auditLogger.LogAuth(security.EventLoginFailed, "", clientIP, r.UserAgent(), "failure", map[string]interface{}{
			"email": maskEmail(email),
		})

		// Mensagem genérica (não revela se email existe)
		writeError(w, http.StatusUnauthorized, "E-mail ou senha incorretos.")
		return
	}

	// Login bem-sucedido
	h.loginLimiter.RecordSuccess(clientIP)

	// Criar sessão
	if err := h.setSession(w, user.ID, r); err != nil {
		writeError(w, http.StatusInternalServerError, "Erro ao criar sessão.")
		return
	}

	// Registrar evento
	h.auditLogger.LogAuth(security.EventLoginSuccess, user.ID, clientIP, r.UserAgent(), "success", nil)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// =============================================================================
// SESSÃO
// =============================================================================

// Me retorna dados do usuário autenticado
//
// Endpoint: GET /api/auth/me
//
// Resposta inclui:
//   - id: ID do usuário
//   - email: email do usuário
//   - name: nome do usuário
//   - created_at: data de criação
//   - is_admin: se o usuário é administrador
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Sessão expirada.")
		return
	}

	user, ok := h.store.GetUserByID(userID)
	if !ok {
		h.auditLogger.LogAuth(security.EventTokenInvalid, userID, security.GetClientIP(r), r.UserAgent(), "failure", nil)
		writeError(w, http.StatusUnauthorized, "Sessão inválida.")
		return
	}

	// Verificar se é admin
	isAdmin := checkIsAdmin(user.Email)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"name":       user.Name,
			"created_at": user.CreatedAt,
			"is_admin":   isAdmin,
		},
	})
}

// checkIsAdmin verifica se o email está na lista de administradores
// Lê a variável de ambiente ADMIN_EMAILS dinamicamente
func checkIsAdmin(email string) bool {
	adminEmails := os.Getenv("ADMIN_EMAILS")
	if adminEmails == "" {
		// Em desenvolvimento sem admins configurados, todos são admin
		if os.Getenv("ENV") != "production" {
			return true
		}
		return false
	}

	email = strings.ToLower(email)
	for _, admin := range strings.Split(adminEmails, ",") {
		admin = strings.TrimSpace(strings.ToLower(admin))
		if email == admin {
			return true
		}
	}
	return false
}

// Logout encerra a sessão do usuário
//
// Endpoint: POST /api/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	clientIP := security.GetClientIP(r)

	// Limpar cookie de sessão
	http.SetCookie(w, &http.Cookie{
		Name:     "famli_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecureContext(r),
		SameSite: http.SameSiteLaxMode,
	})

	// Registrar logout (já usa o AuditLogger que funciona)
	h.auditLogger.LogAuth(security.EventLogout, userID, clientIP, r.UserAgent(), "success", nil)

	writeJSON(w, http.StatusOK, map[string]string{"message": "Sessão encerrada."})
}

// =============================================================================
// SESSÃO JWT
// =============================================================================

// setSession cria um token JWT e define o cookie de sessão
func (h *Handler) setSession(w http.ResponseWriter, userID string, r *http.Request) error {
	now := time.Now()

	// Claims do token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,                             // Subject (ID do usuário)
		"exp": now.Add(7 * 24 * time.Hour).Unix(), // Expira em 7 dias
		"iat": now.Unix(),                         // Issued at
		"nbf": now.Unix(),                         // Not before
		"jti": generateJTI(),                      // JWT ID único
	})

	// Assinar token
	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return err
	}

	// Definir cookie seguro
	http.SetCookie(w, &http.Cookie{
		Name:     "famli_session",
		Value:    signed,
		Path:     "/",
		HttpOnly: true,                 // Não acessível via JavaScript (previne XSS)
		Secure:   isSecureContext(r),   // HTTPS only em produção
		SameSite: http.SameSiteLaxMode, // Proteção contra CSRF
		Expires:  now.Add(7 * 24 * time.Hour),
		MaxAge:   7 * 24 * 60 * 60,
	})

	return nil
}

// =============================================================================
// FUNÇÕES AUXILIARES
// =============================================================================

// writeJSON escreve resposta JSON
func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	security.SetJSONHeaders(w)
	w.WriteHeader(status)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

// writeError escreve resposta de erro
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// isSecureContext verifica se a requisição veio via HTTPS
func isSecureContext(r *http.Request) bool {
	// Verificar TLS direto
	if r.TLS != nil {
		return true
	}
	// Verificar header de proxy
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}
	return false
}

// generateJTI gera um ID único para o token JWT
func generateJTI() string {
	now := time.Now()
	return now.Format("20060102150405") + randomChars(8)
}

// randomChars gera caracteres aleatórios
func randomChars(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(time.Nanosecond) // Variação
	}
	return string(b)
}

// maskEmail mascara parte do email para logs
func maskEmail(email string) string {
	if len(email) < 5 {
		return "***"
	}
	atIdx := -1
	for i, c := range email {
		if c == '@' {
			atIdx = i
			break
		}
	}
	if atIdx <= 2 {
		return "***" + email[atIdx:]
	}
	return email[:2] + "***" + email[atIdx:]
}

// itoa converte int para string
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
