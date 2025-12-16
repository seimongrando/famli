package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "userID"

// JWTMiddleware valida o token JWT no cookie
func JWTMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("famli_session")
			if err != nil {
				http.Error(w, `{"error":"Sessão não encontrada"}`, http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

			if err != nil || !token.Valid {
				http.Error(w, `{"error":"Sessão inválida"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"Sessão inválida"}`, http.StatusUnauthorized)
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				http.Error(w, `{"error":"Sessão inválida"}`, http.StatusUnauthorized)
				return
			}

			expFloat, ok := claims["exp"].(float64)
			if !ok || int64(expFloat) < time.Now().Unix() {
				http.Error(w, `{"error":"Sessão expirada"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
