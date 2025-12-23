// =============================================================================
// FAMLI - Analytics Handler
// =============================================================================
// Gerencia o rastreamento de eventos e métricas de uso
//
// Endpoints:
// - POST /api/analytics/track - Rastreia um evento
// - GET /api/admin/analytics/summary - Resumo de analytics (admin)
// - GET /api/admin/analytics/events - Eventos recentes (admin)
// - GET /api/admin/analytics/daily - Estatísticas diárias (admin)
//
// Eventos rastreados:
// - page_view: Visualização de página
// - login: Login bem-sucedido
// - register: Novo cadastro
// - create_item: Criação de item
// - edit_item: Edição de item
// - delete_item: Exclusão de item
// - create_guardian: Criação de guardião
// - complete_guide: Conclusão de guia
// - export_data: Exportação de dados
// - send_feedback: Envio de feedback
// =============================================================================

package analytics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"famli/internal/auth"
	"famli/internal/i18n"
	"famli/internal/security"
	"famli/internal/storage"

	"github.com/google/uuid"
)

// writeError escreve resposta de erro JSON internacionalizada
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Handler gerencia operações de analytics
type Handler struct {
	store storage.Store
}

// NewHandler cria uma nova instância do handler
func NewHandler(store storage.Store) *Handler {
	return &Handler{store: store}
}

// TrackRequest representa o payload para rastrear um evento
type TrackRequest struct {
	EventType string            `json:"event_type"` // Tipo do evento
	Page      string            `json:"page"`       // Página atual
	Details   map[string]string `json:"details"`    // Detalhes adicionais
}

// Track rastreia um evento de analytics
// POST /api/analytics/track
func (h *Handler) Track(w http.ResponseWriter, r *http.Request) {
	// Obter user ID do contexto usando função do pacote auth
	userID := auth.GetUserID(r)

	var req TrackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "analytics.invalid_data"))
		return
	}

	// Validar tipo do evento
	validEvents := map[string]bool{
		"page_view":       true,
		"login":           true,
		"register":        true,
		"create_item":     true,
		"edit_item":       true,
		"delete_item":     true,
		"create_guardian": true,
		"complete_guide":  true,
		"export_data":     true,
		"send_feedback":   true,
	}

	if !validEvents[req.EventType] {
		// Ignorar eventos desconhecidos silenciosamente
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
		return
	}

	// Criar evento
	event := &storage.AnalyticsEvent{
		ID:        uuid.New().String(),
		UserID:    userID,
		EventType: storage.AnalyticsEventType(req.EventType),
		Page:      req.Page,
		Details:   sanitizeAnalyticsDetails(req.Details),
		CreatedAt: time.Now(),
	}

	// Salvar no banco (silenciosamente ignora erros - tracking não deve bloquear UX)
	_ = h.store.TrackEvent(event)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "tracked"})
}

const (
	maxAnalyticsDetailsEntries    = 20
	maxAnalyticsDetailKeyLength   = 64
	maxAnalyticsDetailValueLength = 255
	maxAnalyticsDetailsBytes      = 2048
)

func sanitizeAnalyticsDetails(details map[string]string) map[string]string {
	if len(details) == 0 {
		return nil
	}

	sanitized := make(map[string]string)
	totalBytes := 0

	for key, value := range details {
		if len(sanitized) >= maxAnalyticsDetailsEntries {
			break
		}
		cleanKey := security.SanitizeText(strings.TrimSpace(key), maxAnalyticsDetailKeyLength)
		if cleanKey == "" {
			continue
		}
		cleanValue := security.SanitizeText(strings.TrimSpace(value), maxAnalyticsDetailValueLength)

		itemBytes := len(cleanKey) + len(cleanValue)
		if totalBytes+itemBytes > maxAnalyticsDetailsBytes {
			break
		}
		sanitized[cleanKey] = cleanValue
		totalBytes += itemBytes
	}

	if len(sanitized) == 0 {
		return nil
	}

	return sanitized
}

// GetSummary retorna o resumo de analytics (admin only)
// GET /api/admin/analytics/summary
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary := h.store.GetAnalyticsSummary()

	if summary == nil {
		summary = &storage.AnalyticsSummary{
			EventsByType: make(map[string]int),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// GetRecentEvents retorna os eventos mais recentes (admin only)
// GET /api/admin/analytics/events?limit=50
func (h *Handler) GetRecentEvents(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
		limit = l
	}

	events, err := h.store.GetRecentEvents(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "analytics.track_error"))
		return
	}

	if events == nil {
		events = []*storage.AnalyticsEvent{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// GetDailyStats retorna estatísticas diárias (admin only)
// GET /api/admin/analytics/daily?days=7
func (h *Handler) GetDailyStats(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 7
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 30 {
		days = d
	}

	stats, err := h.store.GetDailyStats(days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "analytics.track_error"))
		return
	}

	if stats == nil {
		stats = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
