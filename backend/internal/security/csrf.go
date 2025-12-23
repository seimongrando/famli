// =============================================================================
// FAMLI - Protecao CSRF
// =============================================================================
// Verifica Origin/Referer para requisicoes mutantes (POST/PUT/PATCH/DELETE).
// A API usa cookies, entao este middleware evita que outros sites abusem
// das credenciais do usuario em navegadores.
// =============================================================================

package security

import (
	"net/http"
	"net/url"
	"strings"
)

// CSRFMiddleware valida Origin/Referer para metodos inseguros.
// allowMissing permite requisicoes sem Origin/Referer (ex.: clients nao-browser).
func CSRFMiddleware(allowedOrigins []string, allowMissing bool) func(http.Handler) http.Handler {
	allowed := normalizeOrigins(allowedOrigins)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isUnsafeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			origin := normalizeOrigin(r.Header.Get("Origin"))
			if origin == "" {
				origin = normalizeOrigin(refererOrigin(r.Header.Get("Referer")))
			}

			if origin == "" {
				if allowMissing {
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, `{"error":"CSRF check failed"}`, http.StatusForbidden)
				return
			}

			if origin == "null" {
				http.Error(w, `{"error":"CSRF check failed"}`, http.StatusForbidden)
				return
			}

			hostOrigin := requestOrigin(r)
			if origin == hostOrigin || allowed[origin] {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, `{"error":"CSRF check failed"}`, http.StatusForbidden)
		})
	}
}

func isUnsafeMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func normalizeOrigins(origins []string) map[string]bool {
	out := make(map[string]bool, len(origins))
	for _, origin := range origins {
		norm := normalizeOrigin(origin)
		if norm != "" {
			out[norm] = true
		}
	}
	return out
}

func normalizeOrigin(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func refererOrigin(referer string) string {
	if referer == "" {
		return ""
	}
	parsed, err := url.Parse(referer)
	if err != nil {
		return ""
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func requestOrigin(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	return scheme + "://" + r.Host
}
