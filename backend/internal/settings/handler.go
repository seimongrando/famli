package settings

import (
	"encoding/json"
	"net/http"

	"famli/internal/auth"
	"famli/internal/storage"
)

type Handler struct {
	store *storage.MemoryStore
}

func NewHandler(store *storage.MemoryStore) *Handler {
	return &Handler{store: store}
}

type settingsPayload struct {
	EmergencyProtocolEnabled bool   `json:"emergency_protocol_enabled"`
	NotificationsEnabled     bool   `json:"notifications_enabled"`
	Theme                    string `json:"theme"`
}

// Get retorna as configurações do usuário
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	settings := h.store.GetSettings(userID)

	writeJSON(w, http.StatusOK, settings)
}

// Update atualiza as configurações
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)

	var payload settingsPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Dados inválidos.")
		return
	}

	updates := &storage.Settings{
		EmergencyProtocolEnabled: payload.EmergencyProtocolEnabled,
		NotificationsEnabled:     payload.NotificationsEnabled,
		Theme:                    payload.Theme,
	}

	if updates.Theme == "" {
		updates.Theme = "light"
	}

	updated := h.store.UpdateSettings(userID, updates)
	writeJSON(w, http.StatusOK, updated)
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
