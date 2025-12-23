// =============================================================================
// FAMLI - Request Logger com Redação de Dados Sensíveis
// =============================================================================
// Reduz dados sensíveis nos logs HTTP (tokens em URL, etc).
// =============================================================================

package security

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RedactingLogger cria um middleware de logging com redação de caminhos sensíveis.
func RedactingLogger() func(http.Handler) http.Handler {
	return middleware.RequestLogger(&RedactingLogFormatter{})
}

// RedactingLogFormatter implementa o formatter do chi.
type RedactingLogFormatter struct{}

// NewLogEntry cria uma nova entrada de log para a requisição.
func (f *RedactingLogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &redactingLogEntry{
		req:   r,
		start: time.Now(),
	}
}

type redactingLogEntry struct {
	req   *http.Request
	start time.Time
}

func (e *redactingLogEntry) Write(status, bytes int, _ http.Header, elapsed time.Duration, _ interface{}) {
	path := sanitizePath(e.req.URL.Path)
	reqID := middleware.GetReqID(e.req.Context())
	if reqID != "" {
		log.Printf("[REQ] %s %s %d %dB %s req_id=%s", e.req.Method, path, status, bytes, elapsed, reqID)
		return
	}
	log.Printf("[REQ] %s %s %d %dB %s", e.req.Method, path, status, bytes, elapsed)
}

func (e *redactingLogEntry) Panic(v interface{}, stack []byte) {
	log.Printf("[PANIC] %v", v)
}

func sanitizePath(path string) string {
	if path == "" {
		return path
	}

	parts := strings.Split(path, "/")
	for i := 0; i < len(parts); i++ {
		if parts[i] == "" {
			continue
		}
		prev := ""
		prev2 := ""
		if i > 0 {
			prev = parts[i-1]
		}
		if i > 1 {
			prev2 = parts[i-2]
		}

		if prev == "shared" || prev == "guardian-access" || prev == "g" || prev == "compartilhado" {
			parts[i] = ":token"
			continue
		}
		if prev2 == "api" && (prev == "shared" || prev == "guardian-access") {
			parts[i] = ":token"
		}
	}

	return strings.Join(parts, "/")
}
