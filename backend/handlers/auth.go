package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"legacybridge/middleware"
	"legacybridge/store"
)

// AuthHandler manages registration and login.
type AuthHandler struct {
	store     *store.MemoryStore
	jwtSecret string
}

// NewAuthHandler wires the dependencies for auth endpoints.
func NewAuthHandler(store *store.MemoryStore, secret string) *AuthHandler {
	if secret == "" {
		secret = os.Getenv("JWT_SECRET")
	}
	if secret == "" {
		secret = "insecure-dev-secret"
	}
	return &AuthHandler{store: store, jwtSecret: secret}
}

type authPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles user creation with immediate login.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var payload authPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Não foi possível entender os dados.")
		return
	}
	if payload.Email == "" || payload.Password == "" {
		writeError(w, http.StatusBadRequest, "Preencha e-mail e senha.")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Erro ao preparar sua conta.")
		return
	}

	user, err := h.store.CreateUser(payload.Email, string(hashed))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.setSession(w, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "Erro ao criar sessão.")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	})
}

// Login validates credentials and returns a session cookie.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var payload authPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Não foi possível entender os dados.")
		return
	}

	user, ok := h.store.GetUserByEmail(payload.Email)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Conta não encontrada.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "Senha inválida.")
		return
	}

	if err := h.setSession(w, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "Erro ao criar sessão.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	})
}

// Me returns session information for the authenticated user.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Sessão expirada.")
		return
	}

	user, ok := h.store.GetUserByID(userID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Sessão inválida.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	})
}

// Logout clears the session cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "legacybridge_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Sessão encerrada."})
}

func (h *AuthHandler) setSession(w http.ResponseWriter, userID string) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "legacybridge_session",
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
	return nil
}
