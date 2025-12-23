package guardian

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"famli/internal/auth"
	"famli/internal/i18n"
	"famli/internal/security"
	"famli/internal/storage"
)

type Handler struct {
	store storage.Store
}

func NewHandler(store storage.Store) *Handler {
	return &Handler{store: store}
}

type guardianPayload struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone,omitempty"`
	Relationship string `json:"relationship,omitempty"`
	Notes        string `json:"notes,omitempty"`
	AccessPIN    string `json:"access_pin,omitempty"` // PIN de proteção para acesso
}

// List retorna todas as pessoas de confiança
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	guardians := h.store.ListGuardians(userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"guardians": guardians,
	})
}

// Create adiciona uma nova pessoa de confiança
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)

	var payload guardianPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "guardian.invalid_data"))
		return
	}

	payload.Name = security.SanitizeName(payload.Name)
	payload.Relationship = security.SanitizeText(payload.Relationship, security.MaxNameLength)
	payload.Notes = security.SanitizeText(payload.Notes, security.MaxNotesLength)

	if payload.Name == "" {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "guardian.name_required"))
		return
	}
	if payload.AccessPIN == "" {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "guardian.pin_required"))
		return
	}

	// Limitar tamanho das notas para economizar banco
	payload.Notes = strings.TrimSpace(payload.Notes)
	if len(payload.Notes) > security.MaxNotesLength {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "guardian.notes_too_long"))
		return
	}

	guardian := &storage.Guardian{
		Name:         payload.Name,
		Email:        payload.Email,
		Phone:        payload.Phone,
		Relationship: payload.Relationship,
		Notes:        payload.Notes,
		Role:         "viewer",
	}

	// Hash do PIN se fornecido
	if payload.AccessPIN != "" {
		if len(payload.AccessPIN) < 4 {
			writeError(w, http.StatusBadRequest, i18n.Tr(r, "guardian.pin_too_short"))
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(payload.AccessPIN), bcrypt.DefaultCost)
		if err == nil {
			guardian.AccessPIN = string(hash)
		}
	}

	created, err := h.store.CreateGuardian(userID, guardian)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "guardian.add_error"))
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// Update modifica uma pessoa de confiança
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	guardianID := chi.URLParam(r, "guardianID")

	var payload guardianPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "guardian.invalid_data"))
		return
	}

	payload.Name = security.SanitizeName(payload.Name)
	payload.Relationship = security.SanitizeText(payload.Relationship, security.MaxNameLength)
	payload.Notes = security.SanitizeText(payload.Notes, security.MaxNotesLength)

	if payload.Name == "" {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "guardian.name_required"))
		return
	}

	// Limitar tamanho das notas para economizar banco
	payload.Notes = strings.TrimSpace(payload.Notes)
	if len(payload.Notes) > security.MaxNotesLength {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "guardian.notes_too_long"))
		return
	}

	updates := &storage.Guardian{
		Name:         payload.Name,
		Email:        payload.Email,
		Phone:        payload.Phone,
		Relationship: payload.Relationship,
		Notes:        payload.Notes,
	}

	// Hash do PIN se fornecido
	if payload.AccessPIN != "" {
		if len(payload.AccessPIN) < 4 {
			writeError(w, http.StatusBadRequest, i18n.Tr(r, "guardian.pin_too_short"))
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(payload.AccessPIN), bcrypt.DefaultCost)
		if err == nil {
			updates.AccessPIN = string(hash)
		}
	}

	updated, err := h.store.UpdateGuardian(userID, guardianID, updates)
	if err != nil {
		writeError(w, http.StatusNotFound, i18n.Tr(r, "guardian.not_found"))
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// Delete remove uma pessoa de confiança
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	guardianID := chi.URLParam(r, "guardianID")

	if err := h.store.DeleteGuardian(userID, guardianID); err != nil {
		writeError(w, http.StatusNotFound, i18n.Tr(r, "guardian.not_found"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": i18n.Tr(r, "guardian.deleted")})
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
