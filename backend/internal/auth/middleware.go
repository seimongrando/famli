// =============================================================================
// FAMLI - JWT Middleware
// =============================================================================
// Middleware para autenticação JWT com renovação automática de sessão.
//
// Funcionalidades:
// - Valida token JWT no cookie
// - Renova automaticamente sessões próximas de expirar
// - Adiciona user_id e user_email ao contexto
// =============================================================================

package auth

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	userIDKey    contextKey = "userID"
	userEmailKey contextKey = "user_email"
)

// Constantes de tempo para renovação de sessão
const (
	sessionDuration       = 7 * 24 * time.Hour // 7 dias
	renewalThreshold      = 24 * time.Hour     // Renovar se faltam menos de 24h
	sessionCheckThreshold = 6 * time.Hour      // Log se faltam menos de 6h
)

// JWTMiddleware valida o token JWT no cookie e renova automaticamente
// Logs são minimizados para evitar custos - apenas erros importantes são logados
func JWTMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("famli_session")
			if err != nil {
				// Não logar - é normal não ter cookie em algumas situações
				http.Error(w, `{"error":"Sessão não encontrada","code":"SESSION_NOT_FOUND"}`, http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

			if err != nil {
				// Limpar cookie inválido (não logar - pode ser token expirado normal)
				clearSessionCookie(w, r)
				http.Error(w, `{"error":"Sessão inválida","code":"SESSION_INVALID"}`, http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				clearSessionCookie(w, r)
				http.Error(w, `{"error":"Sessão inválida","code":"SESSION_INVALID"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				clearSessionCookie(w, r)
				http.Error(w, `{"error":"Sessão inválida","code":"SESSION_INVALID"}`, http.StatusUnauthorized)
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				clearSessionCookie(w, r)
				http.Error(w, `{"error":"Sessão inválida","code":"SESSION_INVALID"}`, http.StatusUnauthorized)
				return
			}

			expFloat, ok := claims["exp"].(float64)
			if !ok {
				clearSessionCookie(w, r)
				http.Error(w, `{"error":"Sessão inválida","code":"SESSION_INVALID"}`, http.StatusUnauthorized)
				return
			}

			expTime := time.Unix(int64(expFloat), 0)
			now := time.Now()

			// Verificar se expirou
			if expTime.Before(now) {
				clearSessionCookie(w, r)
				http.Error(w, `{"error":"Sessão expirada","code":"SESSION_EXPIRED"}`, http.StatusUnauthorized)
				return
			}

			// Calcular tempo restante
			timeRemaining := expTime.Sub(now)

			// Renovar automaticamente se faltam menos de 24h (sem logar - operação normal)
			if timeRemaining < renewalThreshold {
				renewSession(w, r, sub, secret)
			}

			// Extrair email se presente
			email, _ := claims["email"].(string)

			// Adicionar ao contexto
			ctx := context.WithValue(r.Context(), userIDKey, sub)
			ctx = context.WithValue(ctx, userEmailKey, email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// renewSession renova o token JWT e o cookie de sessão
func renewSession(w http.ResponseWriter, r *http.Request, userID string, secret string) {
	now := time.Now()

	// Novo token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": now.Add(sessionDuration).Unix(),
		"iat": now.Unix(),
		"nbf": now.Unix(),
	})

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Printf("[AUTH] Erro ao renovar sessão: %v", err)
		return
	}

	// Definir novo cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "famli_session",
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureContextMiddleware(r),
		SameSite: http.SameSiteLaxMode,
		Expires:  now.Add(sessionDuration),
		MaxAge:   int(sessionDuration.Seconds()),
	})
	// Não logar renovação - é uma operação normal e frequente
}

// clearSessionCookie limpa o cookie de sessão
func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "famli_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecureContextMiddleware(r),
		SameSite: http.SameSiteLaxMode,
	})
}

// isSecureContextMiddleware verifica se deve usar cookie seguro
func isSecureContextMiddleware(r *http.Request) bool {
	// Em produção, usar HTTPS
	if r.TLS != nil {
		return true
	}
	// Verificar header de proxy (quando atrás de load balancer)
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}
	return false
}

// GetUserID extrai o ID do usuário do contexto
func GetUserID(r *http.Request) string {
	value := r.Context().Value(userIDKey)
	if value == nil {
		return ""
	}
	if userID, ok := value.(string); ok {
		return userID
	}
	return ""
}

// GetUserEmail extrai o email do usuário do contexto
func GetUserEmail(r *http.Request) string {
	value := r.Context().Value(userEmailKey)
	if value == nil {
		return ""
	}
	if email, ok := value.(string); ok {
		return email
	}
	return ""
}
