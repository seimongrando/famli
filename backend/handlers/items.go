package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"legacybridge/middleware"
	"legacybridge/store"
)

// ItemsHandler manages CRUD operations for legacy items.
type ItemsHandler struct {
	store *store.MemoryStore
}

// NewItemsHandler builds an ItemsHandler.
func NewItemsHandler(store *store.MemoryStore) *ItemsHandler {
	return &ItemsHandler{store: store}
}

type itemPayload struct {
	Title   string `json:"title"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

// List returns all legacy items for the session user.
func (h *ItemsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": h.store.ListLegacyItems(userID),
	})
}

// Create stores a new item.
func (h *ItemsHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var payload itemPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Conteúdo inválido.")
		return
	}
	if payload.Title == "" {
		writeError(w, http.StatusBadRequest, "Dê um nome ao item.")
		return
	}
	item := h.store.AddLegacyItem(userID, payload.Title, payload.Type, payload.Content)
	writeJSON(w, http.StatusCreated, item)
}

// Update changes an existing item.
func (h *ItemsHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	itemID := chi.URLParam(r, "itemID")
	var payload itemPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Conteúdo inválido.")
		return
	}
	if payload.Title == "" {
		writeError(w, http.StatusBadRequest, "Dê um nome ao item.")
		return
	}
	updated, err := h.store.UpdateLegacyItem(userID, itemID, payload.Title, payload.Type, payload.Content)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Delete removes the selected item.
func (h *ItemsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	itemID := chi.URLParam(r, "itemID")
	if err := h.store.DeleteLegacyItem(userID, itemID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Item removido."})
}
