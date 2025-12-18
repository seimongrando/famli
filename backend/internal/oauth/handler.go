// =============================================================================
// FAMLI - Handler OAuth (Google, Apple)
// =============================================================================
// Este módulo gerencia autenticação via provedores OAuth2.
//
// Provedores suportados:
// - Google (Google Identity Services)
// - Apple (Sign in with Apple)
//
// Fluxo:
// 1. Frontend obtém token do provedor (Google/Apple SDK)
// 2. Frontend envia token para /api/auth/oauth/{provider}
// 3. Backend valida token com o provedor
// 4. Backend cria/atualiza usuário e retorna sessão JWT
//
// Variáveis de ambiente:
// - GOOGLE_CLIENT_ID: Client ID do Google OAuth
// - APPLE_CLIENT_ID: Bundle ID do app Apple (ex: com.famli.app)
// - APPLE_TEAM_ID: Team ID da Apple Developer
// - APPLE_KEY_ID: Key ID da chave privada Apple
// - APPLE_PRIVATE_KEY: Chave privada para Sign in with Apple
// =============================================================================

package oauth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"famli/internal/i18n"
	"famli/internal/security"
	"famli/internal/storage"
)

// =============================================================================
// HANDLER
// =============================================================================

// Handler gerencia endpoints de autenticação OAuth
type Handler struct {
	store           storage.Store
	jwtSecret       string
	googleClientID  string
	appleClientID   string
	appleTeamID     string
	appleKeyID      string
	applePrivateKey string
	auditLogger     *security.AuditLogger
}

// Config contém as configurações para OAuth
type Config struct {
	GoogleClientID  string
	AppleClientID   string
	AppleTeamID     string
	AppleKeyID      string
	ApplePrivateKey string
}

// NewHandler cria uma nova instância do handler OAuth
func NewHandler(store storage.Store, jwtSecret string, config *Config) *Handler {
	return &Handler{
		store:           store,
		jwtSecret:       jwtSecret,
		googleClientID:  config.GoogleClientID,
		appleClientID:   config.AppleClientID,
		appleTeamID:     config.AppleTeamID,
		appleKeyID:      config.AppleKeyID,
		applePrivateKey: config.ApplePrivateKey,
		auditLogger:     security.GetAuditLogger(),
	}
}

// =============================================================================
// PAYLOADS
// =============================================================================

// oauthPayload é o payload enviado pelo frontend
type oauthPayload struct {
	Token string `json:"token"` // ID Token do provedor
	Nonce string `json:"nonce"` // Nonce para validação (Apple)
}

// googleUserInfo representa os dados do usuário do Google
type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// =============================================================================
// GOOGLE
// =============================================================================

// Google processa login via Google Identity Services
//
// Endpoint: POST /api/auth/oauth/google
//
// O frontend deve usar Google Identity Services (gsi) para obter o ID token
// e enviar para este endpoint.
func (h *Handler) Google(w http.ResponseWriter, r *http.Request) {
	clientIP := security.GetClientIP(r)

	// Verificar se Google está configurado
	if h.googleClientID == "" {
		writeError(w, http.StatusServiceUnavailable, i18n.Tr(r, "oauth.google_not_configured"))
		return
	}

	// Decodificar payload
	var payload oauthPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "auth.invalid_data"))
		return
	}

	if payload.Token == "" {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "oauth.token_required"))
		return
	}

	// Validar token com Google
	userInfo, err := h.validateGoogleToken(payload.Token)
	if err != nil {
		h.auditLogger.LogSecurity(security.EventLoginFailed, clientIP, map[string]interface{}{
			"provider": "google",
			"error":    err.Error(),
		})
		writeError(w, http.StatusUnauthorized, i18n.Tr(r, "oauth.invalid_token"))
		return
	}

	// Verificar email
	if !userInfo.EmailVerified {
		writeError(w, http.StatusUnauthorized, i18n.Tr(r, "oauth.email_not_verified"))
		return
	}

	// Criar ou atualizar usuário
	user, err := h.store.CreateOrUpdateSocialUser(
		storage.AuthProviderGoogle,
		userInfo.Sub,
		userInfo.Email,
		userInfo.Name,
		userInfo.Picture,
	)
	if err != nil {
		h.auditLogger.LogSecurity(security.EventSuspiciousActivity, clientIP, map[string]interface{}{
			"provider": "google",
			"error":    err.Error(),
		})
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "auth.create_error"))
		return
	}

	// Criar sessão JWT
	if err := h.setSession(w, user.ID, user.Email, r); err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "auth.session_error"))
		return
	}

	// Log de sucesso
	h.auditLogger.LogAuth(security.EventLoginSuccess, user.ID, clientIP, r.UserAgent(), "success", map[string]interface{}{
		"provider": "google",
	})

	// Verificar se é admin
	isAdmin := checkIsAdmin(user.Email)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"name":       user.Name,
			"avatar_url": user.AvatarURL,
			"provider":   user.Provider,
			"is_admin":   isAdmin,
		},
	})
}

// validateGoogleToken valida um ID token do Google
func (h *Handler) validateGoogleToken(idToken string) (*googleUserInfo, error) {
	// Usar a API tokeninfo do Google para validar
	resp, err := http.Get(fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken))
	if err != nil {
		return nil, fmt.Errorf("erro ao validar token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token inválido: %s", string(body))
	}

	var tokenInfo struct {
		Aud           string `json:"aud"`
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	// Verificar audience (client ID)
	if tokenInfo.Aud != h.googleClientID {
		return nil, errors.New("audience inválido")
	}

	return &googleUserInfo{
		Sub:           tokenInfo.Sub,
		Email:         tokenInfo.Email,
		EmailVerified: tokenInfo.EmailVerified == "true",
		Name:          tokenInfo.Name,
		GivenName:     tokenInfo.GivenName,
		FamilyName:    tokenInfo.FamilyName,
		Picture:       tokenInfo.Picture,
	}, nil
}

// =============================================================================
// APPLE
// =============================================================================

// Apple processa login via Sign in with Apple
//
// Endpoint: POST /api/auth/oauth/apple
//
// O frontend deve usar Sign in with Apple JS para obter o ID token
// e enviar para este endpoint.
func (h *Handler) Apple(w http.ResponseWriter, r *http.Request) {
	clientIP := security.GetClientIP(r)

	// Verificar se Apple está configurado
	if h.appleClientID == "" {
		writeError(w, http.StatusServiceUnavailable, i18n.Tr(r, "oauth.apple_not_configured"))
		return
	}

	// Decodificar payload
	var payload oauthPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "auth.invalid_data"))
		return
	}

	if payload.Token == "" {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "oauth.token_required"))
		return
	}

	// Validar token com Apple
	claims, err := h.validateAppleToken(payload.Token)
	if err != nil {
		h.auditLogger.LogSecurity(security.EventLoginFailed, clientIP, map[string]interface{}{
			"provider": "apple",
			"error":    err.Error(),
		})
		writeError(w, http.StatusUnauthorized, i18n.Tr(r, "oauth.invalid_token"))
		return
	}

	// Extrair email (pode estar no token ou vir separado)
	email, _ := claims["email"].(string)
	sub, _ := claims["sub"].(string)

	if sub == "" {
		writeError(w, http.StatusUnauthorized, i18n.Tr(r, "oauth.invalid_token"))
		return
	}

	// Nome pode vir do primeiro login ou ser vazio
	// Apple só envia o nome no primeiro login
	name := ""

	// Criar ou atualizar usuário
	user, err := h.store.CreateOrUpdateSocialUser(
		storage.AuthProviderApple,
		sub,
		email,
		name,
		"", // Apple não fornece avatar
	)
	if err != nil {
		h.auditLogger.LogSecurity(security.EventSuspiciousActivity, clientIP, map[string]interface{}{
			"provider": "apple",
			"error":    err.Error(),
		})
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "auth.create_error"))
		return
	}

	// Criar sessão JWT
	if err := h.setSession(w, user.ID, user.Email, r); err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "auth.session_error"))
		return
	}

	// Log de sucesso
	h.auditLogger.LogAuth(security.EventLoginSuccess, user.ID, clientIP, r.UserAgent(), "success", map[string]interface{}{
		"provider": "apple",
	})

	// Verificar se é admin
	isAdmin := checkIsAdmin(user.Email)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"name":       user.Name,
			"avatar_url": user.AvatarURL,
			"provider":   user.Provider,
			"is_admin":   isAdmin,
		},
	})
}

// validateAppleToken valida um ID token da Apple
func (h *Handler) validateAppleToken(idToken string) (jwt.MapClaims, error) {
	// Buscar chaves públicas da Apple
	keys, err := fetchApplePublicKeys()
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar chaves Apple: %w", err)
	}

	// Parse do token (sem validar ainda)
	token, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear token: %w", err)
	}

	// Buscar a chave correta pelo kid
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("kid não encontrado no header")
	}

	var publicKey *rsa.PublicKey
	for _, key := range keys.Keys {
		if key.Kid == kid {
			publicKey, err = key.ToRSAPublicKey()
			if err != nil {
				return nil, fmt.Errorf("erro ao converter chave: %w", err)
			}
			break
		}
	}

	if publicKey == nil {
		return nil, errors.New("chave pública não encontrada")
	}

	// Validar token com a chave pública
	validatedToken, err := jwt.Parse(idToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", t.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao validar token: %w", err)
	}

	claims, ok := validatedToken.Claims.(jwt.MapClaims)
	if !ok || !validatedToken.Valid {
		return nil, errors.New("claims inválidos")
	}

	// Verificar issuer
	if iss, _ := claims["iss"].(string); iss != "https://appleid.apple.com" {
		return nil, errors.New("issuer inválido")
	}

	// Verificar audience
	if aud, _ := claims["aud"].(string); aud != h.appleClientID {
		return nil, errors.New("audience inválido")
	}

	return claims, nil
}

// appleKeys representa as chaves públicas da Apple
type appleKeys struct {
	Keys []appleKey `json:"keys"`
}

type appleKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (k *appleKey) ToRSAPublicKey() (*rsa.PublicKey, error) {
	// Decodificar N (modulus)
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar N: %w", err)
	}

	// Decodificar E (exponent)
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar E: %w", err)
	}

	// Converter E para int
	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}, nil
}

func fetchApplePublicKeys() (*appleKeys, error) {
	resp, err := http.Get("https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var keys appleKeys
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, err
	}

	return &keys, nil
}

// =============================================================================
// STATUS
// =============================================================================

// Status retorna o status das integrações OAuth
//
// Endpoint: GET /api/auth/oauth/status
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"google": map[string]interface{}{
			"enabled":   h.googleClientID != "",
			"client_id": h.googleClientID,
		},
		"apple": map[string]interface{}{
			"enabled":   h.appleClientID != "",
			"client_id": h.appleClientID,
		},
	})
}

// =============================================================================
// FUNÇÕES AUXILIARES
// =============================================================================

// setSession cria um token JWT e define o cookie de sessão
func (h *Handler) setSession(w http.ResponseWriter, userID, email string, r *http.Request) error {
	now := time.Now()
	sessionDuration := 7 * 24 * time.Hour

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   now.Add(sessionDuration).Unix(),
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
		"jti":   fmt.Sprintf("%d%s", now.UnixNano(), randomChars(4)),
	})

	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "famli_session",
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureContext(r),
		SameSite: http.SameSiteLaxMode,
		Expires:  now.Add(7 * 24 * time.Hour),
		MaxAge:   7 * 24 * 60 * 60,
	})

	return nil
}

func isSecureContext(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}
	return false
}

func randomChars(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}

func checkIsAdmin(email string) bool {
	adminEmails := os.Getenv("ADMIN_EMAILS")
	if adminEmails == "" {
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

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	security.SetJSONHeaders(w)
	w.WriteHeader(status)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
