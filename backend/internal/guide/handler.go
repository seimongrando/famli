package guide

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"famli/internal/auth"
	"famli/internal/i18n"
	"famli/internal/storage"
)

// CardConfig define a configuração de um card do guia (sem textos)
type CardConfig struct {
	ID       string
	Icon     string
	Order    int
	ItemType string
}

// cardConfigs são as configurações dos cards (sem textos para i18n)
var cardConfigs = []CardConfig{
	{ID: "welcome", Icon: "👋", Order: 1, ItemType: "info"},
	{ID: "people", Icon: "👥", Order: 2, ItemType: "guardian"},
	{ID: "locations", Icon: "📍", Order: 3, ItemType: "location"},
	{ID: "routines", Icon: "🔄", Order: 4, ItemType: "routine"},
	{ID: "access", Icon: "🔑", Order: 5, ItemType: "access"},
	{ID: "memories", Icon: "💝", Order: 6, ItemType: "memory"},
}

// getLocalizedCards retorna os cards do guia traduzidos para o locale do request
func getLocalizedCards(r *http.Request) []storage.GuideCard {
	cards := make([]storage.GuideCard, len(cardConfigs))
	for idx, cfg := range cardConfigs {
		cards[idx] = storage.GuideCard{
			ID:          cfg.ID,
			Title:       i18n.Tr(r, "guide.card."+cfg.ID+".title"),
			Description: i18n.Tr(r, "guide.card."+cfg.ID+".description"),
			Icon:        cfg.Icon,
			Order:       cfg.Order,
			ItemType:    cfg.ItemType,
		}
	}
	return cards
}

type Handler struct {
	store storage.Store
}

func NewHandler(store storage.Store) *Handler {
	return &Handler{store: store}
}

// ListCards retorna os cards do Guia Famli (traduzidos)
func (h *Handler) ListCards(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"cards": getLocalizedCards(r),
	})
}

// GetProgress retorna o progresso do usuário no guia
func (h *Handler) GetProgress(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	progress := h.store.GetGuideProgress(userID)
	cards := getLocalizedCards(r)

	// Montar resposta com status de cada card
	cardsProgress := make([]map[string]interface{}, len(cards))
	for i, card := range cards {
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
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "guide.invalid_data"))
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
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "guide.invalid_status"))
		return
	}

	progress, err := h.store.UpdateGuideProgress(userID, cardID, payload.Status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "guide.progress_error"))
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
