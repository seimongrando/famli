// =============================================================================
// FAMLI - Feedback Handler
// =============================================================================
// Gerencia o sistema de feedback dos usuários
//
// Endpoints:
// - POST /api/feedback - Envia um feedback
// - GET /api/admin/feedbacks - Lista feedbacks (admin only)
// - PATCH /api/admin/feedbacks/:id - Atualiza status do feedback (admin)
//
// Tipos de feedback:
// - suggestion: Sugestão de melhoria
// - problem: Problema ou bug reportado
// - praise: Elogio ou agradecimento
// - question: Dúvida
// =============================================================================

package feedback

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"famli/internal/auth"
	"famli/internal/i18n"
	"famli/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// writeError escreve resposta de erro JSON internacionalizada
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Handler gerencia operações de feedback
type Handler struct {
	store storage.Store
}

// NewHandler cria uma nova instância do handler
func NewHandler(store storage.Store) *Handler {
	return &Handler{store: store}
}

// CreateFeedbackRequest representa o payload para criar feedback
type CreateFeedbackRequest struct {
	Type    string `json:"type"`    // suggestion, problem, praise, question
	Message string `json:"message"` // Mensagem do feedback
	Page    string `json:"page"`    // Página onde o usuário estava
}

// UpdateFeedbackRequest representa o payload para atualizar feedback
type UpdateFeedbackRequest struct {
	Status    string `json:"status"`     // pending, reviewed, resolved
	AdminNote string `json:"admin_note"` // Nota do admin
}

// Create cria um novo feedback
// POST /api/feedback
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	// Obter user ID e email do contexto usando as funções do pacote auth
	userID := auth.GetUserID(r)
	userEmail := auth.GetUserEmail(r)

	var req CreateFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "feedback.invalid_data"))
		return
	}

	// Validações
	if req.Message == "" {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "feedback.invalid_data"))
		return
	}

	if len(req.Message) > 5000 {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "feedback.invalid_data"))
		return
	}

	// Validar tipo
	validTypes := map[string]bool{
		"suggestion": true,
		"problem":    true,
		"praise":     true,
		"question":   true,
	}
	if req.Type == "" {
		req.Type = "suggestion"
	}
	if !validTypes[req.Type] {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "feedback.type_required"))
		return
	}

	// Criar feedback
	feedback := &storage.Feedback{
		ID:        uuid.New().String(),
		UserID:    userID,
		UserEmail: userEmail,
		Type:      storage.FeedbackType(req.Type),
		Message:   req.Message,
		Page:      req.Page,
		UserAgent: r.Header.Get("User-Agent"),
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Salvar feedback
	if err := h.store.CreateFeedback(feedback); err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "feedback.save_error"))
		return
	}

	// Responder com sucesso
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": i18n.Tr(r, "feedback.send_success"),
		"id":      feedback.ID,
	})
}

// List lista todos os feedbacks (admin only)
// GET /api/admin/feedbacks?status=pending&limit=50
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	feedbacks, err := h.store.ListFeedbacks(status, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "feedback.save_error"))
		return
	}

	if feedbacks == nil {
		feedbacks = []*storage.Feedback{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feedbacks)
}

// Update atualiza o status de um feedback (admin only)
// PATCH /api/admin/feedbacks/:id
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "feedback.not_found"))
		return
	}

	var req UpdateFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "feedback.invalid_data"))
		return
	}

	// Validar status
	validStatuses := map[string]bool{
		"pending":  true,
		"reviewed": true,
		"resolved": true,
	}
	if !validStatuses[req.Status] {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "feedback.invalid_data"))
		return
	}

	if err := h.store.UpdateFeedbackStatus(id, req.Status, req.AdminNote); err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "feedback.update_error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": i18n.Tr(r, "feedback.update_success"),
	})
}

// GetStats retorna estatísticas de feedback (para admin dashboard)
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	total, pending := h.store.GetFeedbackStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"total":   total,
		"pending": pending,
	})
}
