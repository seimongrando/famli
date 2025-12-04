package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"legacybridge/middleware"
	"legacybridge/store"
)

// GuardiansHandler manages trusted contacts.
type GuardiansHandler struct {
	store *store.MemoryStore
}

// NewGuardiansHandler builds a handler.
func NewGuardiansHandler(store *store.MemoryStore) *GuardiansHandler {
	return &GuardiansHandler{store: store}
}

type guardianPayload struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// List returns saved guardians.
func (h *GuardiansHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"guardians": h.store.ListGuardians(userID),
	})
}

// Create adds a guardian.
func (h *GuardiansHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var payload guardianPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Dados inválidos.")
		return
	}
	if payload.Name == "" || payload.Email == "" {
		writeError(w, http.StatusBadRequest, "Informe nome e e-mail.")
		return
	}
	guardian := h.store.AddGuardian(userID, payload.Name, payload.Email)
	writeJSON(w, http.StatusCreated, guardian)
}

// Delete removes a guardian.
func (h *GuardiansHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	guardianID := chi.URLParam(r, "guardianID")
	if err := h.store.DeleteGuardian(userID, guardianID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Pessoa removida."})
}
