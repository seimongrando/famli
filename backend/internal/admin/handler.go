// =============================================================================
// FAMLI - Handler de Administração
// =============================================================================
// Este módulo gerencia os endpoints administrativos para monitoramento
// e gestão do sistema.
//
// Funcionalidades:
// - Dashboard com estatísticas gerais
// - Health check do sistema
// - Listagem de usuários (sem dados sensíveis)
// - Métricas de uso
//
// Segurança:
// - Requer autenticação admin (email em lista permitida)
// - Não expõe dados sensíveis dos usuários
// - Rate limiting aplicado
// =============================================================================

package admin

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"famli/internal/auth"
	"famli/internal/security"
	"famli/internal/storage"
)

// =============================================================================
// CONFIGURAÇÃO
// =============================================================================

// getAdminEmails retorna a lista de emails de administradores
// Lê dinamicamente a variável de ambiente a cada chamada
// Formato: ADMIN_EMAILS=admin1@email.com,admin2@email.com
func getAdminEmails() []string {
	emails := os.Getenv("ADMIN_EMAILS")
	if emails == "" {
		return []string{}
	}

	result := []string{}
	for _, email := range strings.Split(emails, ",") {
		email = strings.TrimSpace(email)
		if email != "" {
			result = append(result, strings.ToLower(email))
		}
	}
	return result
}

// =============================================================================
// HANDLER
// =============================================================================

// Handler gerencia endpoints administrativos
type Handler struct {
	store       *storage.MemoryStore
	startTime   time.Time
	auditLogger *security.AuditLogger
}

// NewHandler cria uma nova instância do handler admin
func NewHandler(store *storage.MemoryStore) *Handler {
	return &Handler{
		store:       store,
		startTime:   time.Now(),
		auditLogger: security.GetAuditLogger(),
	}
}

// =============================================================================
// MIDDLEWARE DE ADMIN
// =============================================================================

// AdminOnly é um middleware que verifica se o usuário é admin
func (h *Handler) AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obter userID do contexto usando a função correta do pacote auth
		userID := auth.GetUserID(r)
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "Não autenticado")
			return
		}

		// Buscar usuário
		user, ok := h.store.GetUserByID(userID)
		if !ok {
			writeError(w, http.StatusUnauthorized, "Usuário não encontrado")
			return
		}

		// Verificar se é admin
		if !isAdmin(user.Email) {
			h.auditLogger.LogSecurity(security.EventUnauthorizedAccess, security.GetClientIP(r), map[string]interface{}{
				"user_id":  userID,
				"email":    user.Email,
				"resource": "admin",
			})
			writeError(w, http.StatusForbidden, "Acesso negado")
			return
		}

		// Registrar acesso admin
		h.auditLogger.LogDataAccess(userID, security.GetClientIP(r), "admin", "access", "success")

		next.ServeHTTP(w, r)
	})
}

// isAdmin verifica se o email está na lista de admins
// Lê a variável de ambiente ADMIN_EMAILS dinamicamente
func isAdmin(email string) bool {
	email = strings.ToLower(email)
	adminEmails := getAdminEmails()
	env := os.Getenv("ENV")

	// Debug log usando AuditLogger para garantir visibilidade
	auditLogger := security.GetAuditLogger()
	auditLogger.Log(security.AuditEvent{
		Type:     security.EventDataAccess,
		Severity: security.SeverityInfo,
		Resource: "admin_check",
		Action:   "verify_permission",
		Result:   "checking",
		Details: map[string]interface{}{
			"email":        email,
			"admin_emails": adminEmails,
			"env":          env,
		},
	})

	for _, adminEmail := range adminEmails {
		if email == adminEmail {
			auditLogger.Log(security.AuditEvent{
				Type:     security.EventDataAccess,
				Severity: security.SeverityInfo,
				Resource: "admin_check",
				Action:   "verify_permission",
				Result:   "granted",
				Details:  map[string]interface{}{"email": email},
			})
			return true
		}
	}

	// Em desenvolvimento, se não houver admins configurados, permitir qualquer usuário autenticado
	if len(adminEmails) == 0 && env != "production" {
		auditLogger.Log(security.AuditEvent{
			Type:     security.EventDataAccess,
			Severity: security.SeverityInfo,
			Resource: "admin_check",
			Action:   "verify_permission",
			Result:   "granted_dev_mode",
			Details:  map[string]interface{}{"reason": "no admins configured in dev"},
		})
		return true
	}

	auditLogger.Log(security.AuditEvent{
		Type:     security.EventUnauthorizedAccess,
		Severity: security.SeverityWarning,
		Resource: "admin_check",
		Action:   "verify_permission",
		Result:   "denied",
		Details: map[string]interface{}{
			"email":        email,
			"admin_emails": adminEmails,
		},
	})
	return false
}

// =============================================================================
// ENDPOINTS
// =============================================================================

// Dashboard retorna estatísticas gerais do sistema
//
// Endpoint: GET /api/admin/dashboard
//
// Resposta:
//   - users: número total de usuários
//   - items: número total de itens
//   - guardians: número total de guardiões
//   - activity: atividade recente
//   - config: configurações do admin (para debug)
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	stats := h.store.GetStats()

	// Calcular métricas adicionais
	avgItemsPerUser := float64(0)
	if stats.TotalUsers > 0 {
		avgItemsPerUser = float64(stats.TotalItems) / float64(stats.TotalUsers)
	}

	// Obter configurações para debug
	adminEmails := getAdminEmails()
	env := os.Getenv("ENV")

	dashboard := map[string]interface{}{
		"overview": map[string]interface{}{
			"total_users":        stats.TotalUsers,
			"total_items":        stats.TotalItems,
			"total_guardians":    stats.TotalGuardians,
			"avg_items_per_user": avgItemsPerUser,
		},
		"items_by_type":     stats.ItemsByType,
		"items_by_category": stats.ItemsByCategory,
		"recent_signups":    stats.RecentSignups,
		"config": map[string]interface{}{
			"admin_emails": adminEmails,
			"environment":  env,
		},
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, dashboard)
}

// Health retorna o status de saúde do sistema
//
// Endpoint: GET /api/admin/health
//
// Resposta:
//   - status: "healthy" ou "degraded"
//   - uptime: tempo de atividade
//   - memory: uso de memória
//   - goroutines: número de goroutines
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	uptime := time.Since(h.startTime)

	health := map[string]interface{}{
		"status": "healthy",
		"uptime": map[string]interface{}{
			"seconds": int64(uptime.Seconds()),
			"human":   formatDuration(uptime),
		},
		"memory": map[string]interface{}{
			"alloc_mb":       float64(memStats.Alloc) / 1024 / 1024,
			"total_alloc_mb": float64(memStats.TotalAlloc) / 1024 / 1024,
			"sys_mb":         float64(memStats.Sys) / 1024 / 1024,
			"num_gc":         memStats.NumGC,
		},
		"runtime": map[string]interface{}{
			"goroutines": runtime.NumGoroutine(),
			"cpus":       runtime.NumCPU(),
			"go_version": runtime.Version(),
		},
		"storage": map[string]interface{}{
			"type":   "memory",
			"status": "ok",
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, health)
}

// Users retorna lista de usuários (sem dados sensíveis)
//
// Endpoint: GET /api/admin/users
//
// Query params:
//   - page: página (default: 1)
//   - limit: itens por página (default: 20, max: 100)
func (h *Handler) Users(w http.ResponseWriter, r *http.Request) {
	users := h.store.ListUsers()

	// Mapear para formato seguro (sem senhas)
	safeUsers := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		itemCount := len(h.store.ListBoxItems(user.ID))
		guardianCount := len(h.store.ListGuardians(user.ID))

		safeUsers = append(safeUsers, map[string]interface{}{
			"id":              user.ID,
			"email":           maskEmail(user.Email),
			"name":            user.Name,
			"created_at":      user.CreatedAt.Format(time.RFC3339),
			"items_count":     itemCount,
			"guardians_count": guardianCount,
			"is_admin":        isAdmin(user.Email),
		})
	}

	response := map[string]interface{}{
		"users": safeUsers,
		"total": len(safeUsers),
	}

	writeJSON(w, http.StatusOK, response)
}

// Activity retorna atividade recente do sistema
//
// Endpoint: GET /api/admin/activity
func (h *Handler) Activity(w http.ResponseWriter, r *http.Request) {
	// Obter eventos de auditoria recentes
	events := h.auditLogger.GetRecentEvents(50)

	// Converter para formato de resposta
	activities := make([]map[string]interface{}, 0, len(events))
	for _, event := range events {
		activities = append(activities, map[string]interface{}{
			"id":        event.ID,
			"type":      string(event.Type),
			"severity":  string(event.Severity),
			"timestamp": event.Timestamp.Format(time.RFC3339),
			"client_ip": maskIP(event.ClientIP),
			"result":    event.Result,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"activities": activities,
		"total":      len(activities),
	})
}

// =============================================================================
// HEALTH CHECK PÚBLICO (sem autenticação)
// =============================================================================

// PublicHealth retorna status básico (para load balancers)
//
// Endpoint: GET /api/health
func (h *Handler) PublicHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
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

// writeError escreve erro JSON
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// maskEmail mascara parte do email para privacidade
func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}

	name := parts[0]
	domain := parts[1]

	if len(name) <= 2 {
		return name[:1] + "***@" + domain
	}

	return name[:2] + "***@" + domain
}

// maskIP mascara parte do IP para privacidade
func maskIP(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return ip // IPv6 ou formato desconhecido
	}
	return parts[0] + "." + parts[1] + ".***. ***"
}

// formatDuration formata duração em formato legível
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return strings.TrimSpace(strings.Join([]string{
			itoa(days) + "d",
			itoa(hours) + "h",
			itoa(minutes) + "m",
		}, " "))
	}

	if hours > 0 {
		return itoa(hours) + "h " + itoa(minutes) + "m"
	}

	return itoa(minutes) + "m"
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
