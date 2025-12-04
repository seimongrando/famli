package handlers

import (
	"encoding/json"
	"net/http"

	"legacybridge/middleware"
	"legacybridge/models"
	"legacybridge/store"
)

// SettingsHandler exposes the emergency protocol toggle.
type SettingsHandler struct {
	store *store.MemoryStore
}

// NewSettingsHandler builds an instance.
func NewSettingsHandler(store *store.MemoryStore) *SettingsHandler {
	return &SettingsHandler{store: store}
}

// Get returns the stored settings or defaults.
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	writeJSON(w, http.StatusOK, h.store.GetSettings(userID))
}

// Save updates the emergency protocol flag.
func (h *SettingsHandler) Save(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var payload models.Settings
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Dados inválidos.")
		return
	}
	payload.UserID = userID
	updated := h.store.UpdateSettings(userID, &payload)
	writeJSON(w, http.StatusOK, updated)
}
