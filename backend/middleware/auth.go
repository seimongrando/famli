package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ContextKey is used to ensure context values stay scoped.
type ContextKey string

const (
	// ContextUserIDKey stores the authenticated user ID inside the request context.
	ContextUserIDKey ContextKey = "userID"
)

// NewJWTMiddleware validates the session stored inside the httpOnly cookie.
func NewJWTMiddleware(secret string) func(http.Handler) http.Handler {
	if strings.TrimSpace(secret) == "" {
		secret = "insecure-dev-secret"
		os.Setenv("JWT_SECRET", secret)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("legacybridge_session")
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
			if err != nil || !token.Valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			expFloat, ok := claims["exp"].(float64)
			if !ok || int64(expFloat) < time.Now().Unix() {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserIDKey, sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID pulls the authenticated user ID from the request context.
func GetUserID(r *http.Request) string {
	value := r.Context().Value(ContextUserIDKey)
	if value == nil {
		return ""
	}
	if userID, ok := value.(string); ok {
		return userID
	}
	return ""
}
