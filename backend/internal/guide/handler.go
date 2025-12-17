package guide

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"famli/internal/auth"
	"famli/internal/storage"
)

// Cards pré-definidos do Guia Famli
var defaultCards = []storage.GuideCard{
	{
		ID:          "welcome",
		Title:       "Comece por aqui",
		Description: "Dê o primeiro passo: registre algo simples, como o telefone de emergência ou um contato importante.",
		Icon:        "👋",
		Order:       1,
		ItemType:    "info",
	},
	{
		ID:          "people",
		Title:       "Pessoas importantes",
		Description: "Quem são as pessoas que devem ser avisadas se você precisar de ajuda? Registre aqui seus contatos de confiança.",
		Icon:        "👥",
		Order:       2,
		ItemType:    "guardian",
	},
	{
		ID:          "locations",
		Title:       "Onde estão as coisas importantes",
		Description: "Documentos, chaves, cartões... Explique onde estão as coisas que alguém precisaria encontrar.",
		Icon:        "📍",
		Order:       3,
		ItemType:    "location",
	},
	{
		ID:          "routines",
		Title:       "Rotina que não pode parar",
		Description: "Medicamentos, contas automáticas, pets... O que precisa continuar funcionando mesmo se você não estiver por perto?",
		Icon:        "🔄",
		Order:       4,
		ItemType:    "routine",
	},
	{
		ID:          "access",
		Title:       "Como acessar suas coisas",
		Description: "Explique onde estão suas senhas (não as senhas em si!) e como alguém de confiança pode ajudar a acessar.",
		Icon:        "🔑",
		Order:       5,
		ItemType:    "access",
	},
	{
		ID:          "memories",
		Title:       "Notas pessoais e memórias",
		Description: "Mensagens, histórias, recados... Um espaço para deixar algo especial para quem você ama.",
		Icon:        "💝",
		Order:       6,
		ItemType:    "memory",
	},
}

type Handler struct {
	store storage.Store
}

func NewHandler(store storage.Store) *Handler {
	return &Handler{store: store}
}

// ListCards retorna os cards do Guia Famli
func (h *Handler) ListCards(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"cards": defaultCards,
	})
}

// GetProgress retorna o progresso do usuário no guia
func (h *Handler) GetProgress(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	progress := h.store.GetGuideProgress(userID)

	// Montar resposta com status de cada card
	cardsProgress := make([]map[string]interface{}, len(defaultCards))
	for i, card := range defaultCards {
		status := "pending"
		if p, ok := progress[card.ID]; ok {
			status = p.Status
		}
		cardsProgress[i] = map[string]interface{}{
			"card_id": card.ID,
			"status":  status,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"progress": cardsProgress,
	})
}

// MarkCardProgress atualiza o progresso de um card
func (h *Handler) MarkCardProgress(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	cardID := chi.URLParam(r, "cardID")

	var payload struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Dados inválidos.")
		return
	}

	// Validar status
	validStatuses := map[string]bool{
		"pending":   true,
		"started":   true,
		"completed": true,
		"skipped":   true,
	}
	if !validStatuses[payload.Status] {
		writeError(w, http.StatusBadRequest, "Status inválido.")
		return
	}

	progress, err := h.store.UpdateGuideProgress(userID, cardID, payload.Status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Erro ao salvar progresso.")
		return
	}

	writeJSON(w, http.StatusOK, progress)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
