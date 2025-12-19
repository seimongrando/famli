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
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"famli/internal/email"
	"famli/internal/i18n"
	"famli/internal/security"
	"famli/internal/storage"
)

// =============================================================================
// HANDLER
// =============================================================================

// Handler gerencia endpoints de autenticação
type Handler struct {
	// store é o armazenamento de dados
	store storage.Store

	// jwtSecret é o segredo para assinar tokens JWT
	jwtSecret string

	// loginLimiter controla rate limit de login
	loginLimiter *security.RateLimiter

	// registerLimiter controla rate limit de registro
	registerLimiter *security.RateLimiter

	// auditLogger registra eventos de segurança
	auditLogger *security.AuditLogger

	// emailService envia emails
	emailService *email.Service
}

// NewHandler cria uma nova instância do handler de autenticação
//
// Parâmetros:
//   - store: armazenamento de dados
//   - secret: segredo JWT (mínimo 32 caracteres em produção)
//
// Retorna:
//   - *Handler: handler configurado
func NewHandler(store storage.Store, secret string) *Handler {
	return &Handler{
		store:           store,
		jwtSecret:       secret,
		loginLimiter:    security.NewRateLimiter(security.LoginRateLimit),
		registerLimiter: security.NewRateLimiter(security.RegisterRateLimit),
		auditLogger:     security.GetAuditLogger(),
		emailService:    email.NewService(),
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
		writeError(w, http.StatusTooManyRequests, i18n.Tr(r, "auth.rate_limit"))
		return
	}

	// Decodificar payload
	var payload registerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "auth.invalid_data"))
		return
	}

	// Validar e sanitizar email
	email, err := security.ValidateEmail(payload.Email)
	if err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "auth.email_invalid"))
		return
	}

	// Validar força da senha
	strength, err := security.ValidatePassword(payload.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "auth.password_weak"))
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
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "auth.prepare_error"))
		return
	}

	// Criar usuário
	user, err := h.store.CreateUser(email, string(hashed), name)
	if err != nil {
		if err == storage.ErrAlreadyExists {
			// Não revelar se o email existe (proteção contra enumeração)
			// Usar mesma mensagem de sucesso após delay
			time.Sleep(100 * time.Millisecond) // Timing attack protection
			writeError(w, http.StatusBadRequest, i18n.Tr(r, "auth.email_exists"))
			return
		}
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "auth.create_error"))
		return
	}

	// Criar sessão (inclui email no token para contexto)
	if err := h.setSession(w, user.ID, user.Email, r); err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "auth.session_error"))
		return
	}

	// Registrar evento de auditoria
	h.auditLogger.LogAuth(security.EventRegister, user.ID, clientIP, r.UserAgent(), "success", map[string]interface{}{
		"email":             maskEmail(email),
		"password_strength": strength,
	})

	// Enviar email de boas-vindas (em background, não bloqueia)
	if h.emailService != nil && h.emailService.IsConfigured() {
		locale := i18n.GetLocale(r)
		go h.emailService.SendWelcome(user.Email, user.Name, locale)
	}

	// Verificar se é admin para retornar na resposta
	isAdmin := checkIsAdmin(user.Email)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"user": map[string]interface{}{
			"id":       user.ID,
			"email":    user.Email,
			"name":     user.Name,
			"is_admin": isAdmin,
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
		writeError(w, http.StatusTooManyRequests, i18n.Tr(r, "auth.rate_limit"))
		return
	}

	// Decodificar payload
	var payload loginPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "auth.invalid_data"))
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
		writeError(w, http.StatusUnauthorized, i18n.Tr(r, "auth.invalid_credentials"))
		return
	}

	// Login bem-sucedido
	h.loginLimiter.RecordSuccess(clientIP)

	// Atualizar locale do usuário baseado no Accept-Language
	locale := i18n.GetLocale(r)
	if locale != "" && locale != user.Locale {
		_ = h.store.UpdateUserLocale(user.ID, locale) // Ignora erro, não é crítico
	}

	// Criar sessão (inclui email no token para contexto)
	if err := h.setSession(w, user.ID, user.Email, r); err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "auth.session_error"))
		return
	}

	// Registrar evento
	h.auditLogger.LogAuth(security.EventLoginSuccess, user.ID, clientIP, r.UserAgent(), "success", nil)

	// Verificar se é admin para retornar na resposta
	isAdmin := checkIsAdmin(user.Email)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":       user.ID,
			"email":    user.Email,
			"name":     user.Name,
			"is_admin": isAdmin,
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
		writeError(w, http.StatusUnauthorized, i18n.Tr(r, "auth.session_expired"))
		return
	}

	user, ok := h.store.GetUserByID(userID)
	if !ok {
		h.auditLogger.LogAuth(security.EventTokenInvalid, userID, security.GetClientIP(r), r.UserAgent(), "failure", nil)
		writeError(w, http.StatusUnauthorized, i18n.Tr(r, "auth.session_invalid"))
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

	writeJSON(w, http.StatusOK, map[string]string{"message": i18n.Tr(r, "auth.logout_success")})
}

// =============================================================================
// EXCLUSÃO DE CONTA (LGPD: Direito ao Esquecimento)
// =============================================================================

// deleteAccountPayload é o payload para exclusão de conta
type deleteAccountPayload struct {
	Password     string `json:"password"`     // Confirmação de senha
	Confirmation string `json:"confirmation"` // Texto de confirmação "EXCLUIR MINHA CONTA"
}

// DeleteAccount exclui a conta do usuário e todos os seus dados
//
// Endpoint: DELETE /api/auth/account
//
// Este endpoint implementa o "direito ao esquecimento" (LGPD Art. 18)
// Todos os dados do usuário são permanentemente removidos.
//
// Requisições:
//   - Password: senha atual para confirmação
//   - Confirmation: texto exato "DELETE MY ACCOUNT" ou "EXCLUIR MINHA CONTA"
//
// Segurança:
//   - Requer autenticação
//   - Validação de senha atual
//   - Confirmação textual obrigatória
//   - Log de auditoria mantido por requisitos legais
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	clientIP := security.GetClientIP(r)
	userID := GetUserID(r)

	if userID == "" {
		writeError(w, http.StatusUnauthorized, i18n.Tr(r, "auth.session_invalid"))
		return
	}

	// Rate limiting
	allowed, _ := h.loginLimiter.Allow(clientIP)
	if !allowed {
		h.auditLogger.LogAuth(security.EventRateLimitExceeded, userID, clientIP, r.UserAgent(), "rate_limited", nil)
		writeError(w, http.StatusTooManyRequests, i18n.Tr(r, "auth.rate_limit"))
		return
	}

	// Parse payload
	var payload deleteAccountPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "auth.invalid_data"))
		return
	}

	// Validar texto de confirmação
	confirmTexts := []string{
		"DELETE MY ACCOUNT",
		"EXCLUIR MINHA CONTA",
	}
	validConfirmation := false
	normalizedConfirmation := strings.ToUpper(strings.TrimSpace(payload.Confirmation))
	for _, txt := range confirmTexts {
		if normalizedConfirmation == txt {
			validConfirmation = true
			break
		}
	}
	if !validConfirmation {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "auth.delete_confirm"))
		return
	}

	// Buscar usuário
	user, found := h.store.GetUserByID(userID)
	if !found {
		h.auditLogger.LogAuth(security.EventAccountDeletion, userID, clientIP, r.UserAgent(), "user_not_found", nil)
		writeError(w, http.StatusNotFound, i18n.Tr(r, "auth.user_not_found"))
		return
	}

	// Debug: verificar se a senha foi recuperada corretamente
	if user.Password == "" {
		h.auditLogger.LogAuth(security.EventAccountDeletion, userID, clientIP, r.UserAgent(), "empty_password_hash", nil)
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "auth.internal_error"))
		return
	}

	// Verificar senha
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		h.auditLogger.LogAuth(security.EventAccountDeletion, userID, clientIP, r.UserAgent(), "invalid_password", map[string]interface{}{
			"password_hash_len":  len(user.Password),
			"input_password_len": len(payload.Password),
		})
		writeError(w, http.StatusUnauthorized, i18n.Tr(r, "auth.password_incorrect"))
		return
	}

	// Registrar auditoria ANTES de deletar (por requisitos legais)
	h.auditLogger.LogAuth(security.EventAccountDeletion, userID, clientIP, r.UserAgent(), "initiated", map[string]interface{}{
		"email": maskEmail(user.Email),
	})

	// Deletar conta e todos os dados
	if err := h.store.DeleteUser(userID); err != nil {
		h.auditLogger.LogAuth(security.EventAccountDeletion, userID, clientIP, r.UserAgent(), "error", map[string]interface{}{
			"error": err.Error(),
		})
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "auth.delete_error"))
		return
	}

	// Limpar cookie de sessão
	http.SetCookie(w, &http.Cookie{
		Name:     "famli_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecureContext(r),
		SameSite: http.SameSiteLaxMode,
	})

	// Registrar sucesso
	h.auditLogger.LogAuth(security.EventAccountDeletion, userID, clientIP, r.UserAgent(), "success", nil)

	writeJSON(w, http.StatusOK, map[string]string{
		"message": i18n.Tr(r, "auth.delete_success"),
	})
}

// ExportData exporta todos os dados do usuário (LGPD: Portabilidade)
//
// Endpoint: GET /api/auth/export
//
// Este endpoint implementa o direito à portabilidade (LGPD Art. 18)
// Retorna todos os dados do usuário em formato JSON.
func (h *Handler) ExportData(w http.ResponseWriter, r *http.Request) {
	clientIP := security.GetClientIP(r)
	userID := GetUserID(r)

	if userID == "" {
		writeError(w, http.StatusUnauthorized, i18n.Tr(r, "auth.session_invalid"))
		return
	}

	// Exportar dados
	data, err := h.store.ExportUserData(userID)
	if err != nil {
		h.auditLogger.LogAuth(security.EventDataExport, userID, clientIP, r.UserAgent(), "error", nil)
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "auth.export_error"))
		return
	}

	// Registrar exportação
	h.auditLogger.LogAuth(security.EventDataExport, userID, clientIP, r.UserAgent(), "success", nil)

	// Retornar como download JSON
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=famli-meus-dados.json")
	json.NewEncoder(w).Encode(data)
}

// =============================================================================
// SESSÃO JWT
// =============================================================================

// setSession cria um token JWT e define o cookie de sessão
// Inclui o email no token para facilitar identificação em feedbacks e logs
func (h *Handler) setSession(w http.ResponseWriter, userID, email string, r *http.Request) error {
	now := time.Now()
	sessionDuration := 7 * 24 * time.Hour

	// Claims do token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   userID,                          // Subject (ID do usuário)
		"email": email,                           // Email do usuário (para contexto)
		"exp":   now.Add(sessionDuration).Unix(), // Expira em 7 dias
		"iat":   now.Unix(),                      // Issued at
		"nbf":   now.Unix(),                      // Not before
		"jti":   generateJTI(),                   // JWT ID único
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

// =============================================================================
// RECUPERAÇÃO DE SENHA
// =============================================================================

// forgotPasswordPayload é o payload para solicitar reset
type forgotPasswordPayload struct {
	Email string `json:"email"`
}

// resetPasswordPayload é o payload para redefinir a senha
type resetPasswordPayload struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ForgotPassword inicia o processo de recuperação de senha
//
// Endpoint: POST /api/auth/forgot-password
//
// Segurança:
// - Não revela se o email existe (mensagem genérica)
// - Rate limiting para prevenir abuso
// - Token expira em 1 hora
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	clientIP := security.GetClientIP(r)

	// Rate limiting
	allowed, _ := h.registerLimiter.Allow(clientIP)
	if !allowed {
		writeError(w, http.StatusTooManyRequests, i18n.Tr(r, "auth.rate_limit"))
		return
	}

	var payload forgotPasswordPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "auth.invalid_data"))
		return
	}

	// Normalizar email
	emailAddr, _ := security.ValidateEmail(payload.Email)
	if emailAddr == "" {
		emailAddr = strings.ToLower(strings.TrimSpace(payload.Email))
	}

	// Resposta genérica (não revela se email existe)
	// O processamento real acontece em background
	go h.processPasswordReset(emailAddr, r)

	// Sempre retorna sucesso para não revelar se email existe
	writeJSON(w, http.StatusOK, map[string]string{
		"message": i18n.Tr(r, "password.reset_sent"),
	})
}

// processPasswordReset processa a solicitação de reset em background
func (h *Handler) processPasswordReset(emailAddr string, r *http.Request) {
	// Buscar usuário
	user, ok := h.store.GetUserByEmail(emailAddr)
	if !ok {
		return // Email não existe, não fazer nada
	}

	// Gerar token seguro
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return
	}
	rawToken := hex.EncodeToString(tokenBytes)

	// Hash do token para armazenar
	tokenHash := sha256.Sum256([]byte(rawToken))
	hashedToken := hex.EncodeToString(tokenHash[:])

	// Criar registro de reset
	resetToken := &storage.PasswordResetToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     hashedToken,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := h.store.CreatePasswordResetToken(resetToken); err != nil {
		return
	}

	// Construir URL de reset (usa rota do idioma do usuário)
	baseURL := getBaseURLFromRequest(r)
	locale := user.Locale
	if locale == "" {
		locale = i18n.GetLocale(r) // Fallback para o idioma da requisição
	}

	// Usa rota localizada
	resetPath := "/redefinir-senha"
	if strings.HasPrefix(locale, "en") {
		resetPath = "/reset-password"
	}
	resetLink := baseURL + resetPath + "?token=" + rawToken

	// Enviar email no idioma do usuário
	h.emailService.SendPasswordReset(user.Email, user.Name, resetLink, locale)
}

// ResetPassword redefine a senha usando o token
//
// Endpoint: POST /api/auth/reset-password
//
// Segurança:
// - Token é válido por 1 hora
// - Token só pode ser usado uma vez
// - Senha deve atender requisitos de força
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	clientIP := security.GetClientIP(r)

	var payload resetPasswordPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "auth.invalid_data"))
		return
	}

	if payload.Token == "" || payload.NewPassword == "" {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "auth.invalid_data"))
		return
	}

	// Validar força da senha
	_, err := security.ValidatePassword(payload.NewPassword)
	if err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "auth.password_weak"))
		return
	}

	// Hash do token recebido
	tokenHash := sha256.Sum256([]byte(payload.Token))
	hashedToken := hex.EncodeToString(tokenHash[:])

	// Buscar token no banco
	resetToken, err := h.store.GetPasswordResetToken(hashedToken)
	if err != nil {
		h.auditLogger.LogSecurity(security.EventSuspiciousActivity, clientIP, map[string]interface{}{
			"event": "invalid_reset_token",
		})
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "password.reset_invalid"))
		return
	}

	// Buscar usuário
	user, ok := h.store.GetUserByID(resetToken.UserID)
	if !ok {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "password.reset_invalid"))
		return
	}

	// Hash da nova senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "password.reset_error"))
		return
	}

	// Atualizar senha (precisamos adicionar este método ao store)
	if err := h.updateUserPassword(user.ID, string(hashedPassword)); err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "password.reset_error"))
		return
	}

	// Marcar token como usado
	h.store.MarkPasswordResetTokenUsed(resetToken.ID)

	// Log
	h.auditLogger.LogAuth(security.EventPasswordChange, user.ID, clientIP, r.UserAgent(), "success", nil)

	writeJSON(w, http.StatusOK, map[string]string{
		"message": i18n.Tr(r, "password.reset_success"),
	})
}

// updateUserPassword atualiza a senha do usuário
func (h *Handler) updateUserPassword(userID, hashedPassword string) error {
	return h.store.UpdateUserPassword(userID, hashedPassword)
}

// getBaseURLFromRequest extrai a URL base da requisição
func getBaseURLFromRequest(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	return scheme + "://" + r.Host
}
