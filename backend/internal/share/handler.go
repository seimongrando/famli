// =============================================================================
// FAMLI - Handler de Compartilhamento
// =============================================================================
// Gerencia links de compartilhamento para guardiões acessarem informações.
//
// Endpoints:
// - POST /api/share/links - Criar link de compartilhamento
// - GET /api/share/links - Listar links do usuário
// - DELETE /api/share/links/:id - Remover link
// - GET /api/shared/:token - Acessar conteúdo compartilhado (público)
// - POST /api/shared/:token/verify - Verificar PIN (se necessário)
//
// Tipos de link:
// - normal: Acesso a categorias selecionadas
// - emergency: Acesso em caso de emergência
// - memorial: Acesso completo após falecimento
// =============================================================================

package share

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"famli/internal/auth"
	"famli/internal/i18n"
	"famli/internal/security"
	"famli/internal/storage"
)

// Handler gerencia operações de compartilhamento
type Handler struct {
	store       storage.Store
	auditLogger *security.AuditLogger
}

// NewHandler cria uma nova instância do handler
func NewHandler(store storage.Store) *Handler {
	return &Handler{
		store:       store,
		auditLogger: security.GetAuditLogger(),
	}
}

// =============================================================================
// PAYLOADS
// =============================================================================

// CreateLinkRequest representa o payload para criar um link
type CreateLinkRequest struct {
	Name        string   `json:"name"`                   // Nome identificador
	GuardianID  string   `json:"guardian_id,omitempty"`  // Guardião específico (deprecated)
	GuardianIDs []string `json:"guardian_ids,omitempty"` // Guardiões específicos
	Type        string   `json:"type"`                   // normal, emergency, memorial
	Categories  []string `json:"categories,omitempty"`   // Categorias permitidas
	PIN         string   `json:"pin,omitempty"`          // PIN opcional
	ExpiresIn   int      `json:"expires_in,omitempty"`   // Dias até expirar (0 = nunca)
	MaxUses     int      `json:"max_uses,omitempty"`     // Máximo de usos (0 = ilimitado)
}

// ShareLinkResponse representa a resposta com o link criado
type ShareLinkResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	URL        string     `json:"url"`
	Categories []string   `json:"categories,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	MaxUses    int        `json:"max_uses"`
	UsageCount int        `json:"usage_count"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
}

// VerifyPINRequest representa o payload para verificar PIN
type VerifyPINRequest struct {
	PIN string `json:"pin"`
}

// =============================================================================
// ENDPOINTS AUTENTICADOS
// =============================================================================

// CreateLink cria um novo link de compartilhamento
// POST /api/share/links
func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	clientIP := security.GetClientIP(r)

	var req CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "share.invalid_data"))
		return
	}

	// Validar nome
	if req.Name == "" {
		req.Name = "Link de Compartilhamento"
	}
	if len(req.Name) > 100 {
		req.Name = req.Name[:100]
	}

	// Validar tipo
	linkType := storage.ShareLinkNormal
	switch req.Type {
	case "emergency":
		linkType = storage.ShareLinkEmergency
	case "memorial":
		linkType = storage.ShareLinkMemorial
	}

	// Gerar token seguro
	token := generateSecureToken()

	// Hash do PIN se fornecido
	var pinHash string
	if req.PIN != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.DefaultCost)
		if err == nil {
			pinHash = string(hash)
		}
	}

	// Calcular expiração
	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		exp := time.Now().AddDate(0, 0, req.ExpiresIn)
		expiresAt = &exp
	}

	// Consolidar GuardianID e GuardianIDs
	guardianIDs := req.GuardianIDs
	if req.GuardianID != "" && len(guardianIDs) == 0 {
		guardianIDs = []string{req.GuardianID}
	}

	now := time.Now()
	link := &storage.ShareLink{
		ID:          uuid.New().String(),
		UserID:      userID,
		GuardianID:  req.GuardianID,
		GuardianIDs: guardianIDs,
		Token:       token,
		Type:        linkType,
		Name:        req.Name,
		PIN:         pinHash,
		Categories:  req.Categories,
		ExpiresAt:   expiresAt,
		MaxUses:     req.MaxUses,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.store.CreateShareLink(link); err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "share.create_error"))
		return
	}

	// Log de auditoria
	h.auditLogger.LogDataAccess(userID, clientIP, "share/links/"+link.ID, "create", "success")

	// Construir URL
	baseURL := getBaseURL(r)
	shareURL := baseURL + "/compartilhado/" + token

	writeJSON(w, http.StatusCreated, ShareLinkResponse{
		ID:         link.ID,
		Name:       link.Name,
		Type:       string(link.Type),
		URL:        shareURL,
		Categories: link.Categories,
		ExpiresAt:  link.ExpiresAt,
		MaxUses:    link.MaxUses,
		UsageCount: link.UsageCount,
		IsActive:   link.IsActive,
		CreatedAt:  link.CreatedAt,
	})
}

// ListLinks lista os links do usuário
// GET /api/share/links
func (h *Handler) ListLinks(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)

	links, err := h.store.GetShareLinksByUser(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "share.list_error"))
		return
	}

	// Converter para response (sem expor tokens)
	baseURL := getBaseURL(r)
	var responses []ShareLinkResponse
	for _, link := range links {
		responses = append(responses, ShareLinkResponse{
			ID:         link.ID,
			Name:       link.Name,
			Type:       string(link.Type),
			URL:        baseURL + "/compartilhado/" + link.Token,
			Categories: link.Categories,
			ExpiresAt:  link.ExpiresAt,
			MaxUses:    link.MaxUses,
			UsageCount: link.UsageCount,
			IsActive:   link.IsActive,
			CreatedAt:  link.CreatedAt,
		})
	}

	if responses == nil {
		responses = []ShareLinkResponse{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"links": responses,
	})
}

// DeleteLink remove um link
// DELETE /api/share/links/:id
func (h *Handler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	linkID := chi.URLParam(r, "id")
	clientIP := security.GetClientIP(r)

	if err := h.store.DeleteShareLink(userID, linkID); err != nil {
		writeError(w, http.StatusNotFound, i18n.Tr(r, "share.not_found"))
		return
	}

	h.auditLogger.LogDataAccess(userID, clientIP, "share/links/"+linkID, "delete", "success")

	writeJSON(w, http.StatusOK, map[string]string{
		"message": i18n.Tr(r, "share.deleted"),
	})
}

// =============================================================================
// ENDPOINTS PÚBLICOS (Acesso via Link)
// =============================================================================

// AccessShared acessa o conteúdo compartilhado
// GET /api/shared/:token
func (h *Handler) AccessShared(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	clientIP := security.GetClientIP(r)

	// Buscar link
	link, err := h.store.GetShareLinkByToken(token)
	if err != nil {
		writeError(w, http.StatusNotFound, i18n.Tr(r, "share.link_expired"))
		return
	}

	// Verificar expiração
	if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now()) {
		writeError(w, http.StatusGone, i18n.Tr(r, "share.link_expired"))
		return
	}

	// Verificar limite de uso
	if link.MaxUses > 0 && link.UsageCount >= link.MaxUses {
		writeError(w, http.StatusGone, i18n.Tr(r, "share.link_expired"))
		return
	}

	// Verificar se precisa de PIN
	if link.PIN != "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"requires_pin": true,
			"link_type":    link.Type,
		})
		return
	}

	// Buscar dados do usuário
	sharedView, err := h.getSharedContent(link)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "share.access_error"))
		return
	}

	// Registrar acesso
	h.recordAccess(link, clientIP, r.UserAgent())

	writeJSON(w, http.StatusOK, sharedView)
}

// VerifyPIN verifica o PIN e retorna o conteúdo
// POST /api/shared/:token/verify
func (h *Handler) VerifyPIN(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	clientIP := security.GetClientIP(r)

	var req VerifyPINRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "share.invalid_data"))
		return
	}

	// Buscar link
	link, err := h.store.GetShareLinkByToken(token)
	if err != nil {
		writeError(w, http.StatusNotFound, i18n.Tr(r, "share.link_expired"))
		return
	}

	// Verificar PIN
	if err := bcrypt.CompareHashAndPassword([]byte(link.PIN), []byte(req.PIN)); err != nil {
		writeError(w, http.StatusUnauthorized, i18n.Tr(r, "share.invalid_pin"))
		return
	}

	// Buscar dados
	sharedView, err := h.getSharedContent(link)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "share.access_error"))
		return
	}

	// Registrar acesso
	h.recordAccess(link, clientIP, r.UserAgent())

	writeJSON(w, http.StatusOK, sharedView)
}

// =============================================================================
// FUNÇÕES AUXILIARES
// =============================================================================

// getSharedContent retorna o conteúdo baseado no tipo de link
func (h *Handler) getSharedContent(link *storage.ShareLink) (*storage.SharedView, error) {
	// Buscar usuário
	user, ok := h.store.GetUserByID(link.UserID)
	if !ok {
		return nil, storage.ErrNotFound
	}

	// Buscar itens
	allItems, err := h.store.GetBoxItems(link.UserID)
	if err != nil {
		return nil, err
	}

	// Filtrar por categoria se necessário
	var items []*storage.BoxItem
	for _, item := range allItems {
		if len(link.Categories) == 0 || contains(link.Categories, item.Category) {
			items = append(items, item)
		}
	}

	view := &storage.SharedView{
		UserName:   user.Name,
		Items:      items,
		LinkType:   link.Type,
		AccessedAt: time.Now(),
	}

	// Adicionar guardiões baseado no tipo de link e filtro
	allGuardians := h.store.ListGuardians(link.UserID)

	// Se há guardiões específicos selecionados, filtrar
	if len(link.GuardianIDs) > 0 {
		var filteredGuardians []*storage.Guardian
		for _, g := range allGuardians {
			for _, id := range link.GuardianIDs {
				if g.ID == id {
					filteredGuardians = append(filteredGuardians, g)
					break
				}
			}
		}
		view.Guardians = filteredGuardians

		// Nome do guardião para mensagem personalizada
		if len(filteredGuardians) == 1 {
			view.GuardianName = filteredGuardians[0].Name
		}
	}

	// Adicionar todos guardiões em modo memorial
	if link.Type == storage.ShareLinkMemorial {
		if len(link.GuardianIDs) == 0 {
			view.Guardians = allGuardians
		}
		view.UserEmail = user.Email
		view.Message = "Este é o memorial de " + user.Name + ". As informações aqui foram deixadas para ajudar você."
	}

	// Mensagem para modo emergência
	if link.Type == storage.ShareLinkEmergency {
		view.Message = "Acesso de emergência às informações de " + user.Name + "."
	}

	return view, nil
}

// recordAccess registra um acesso ao link
func (h *Handler) recordAccess(link *storage.ShareLink, ip, userAgent string) {
	// Incrementar contador
	h.store.IncrementShareLinkUsage(link.ID)

	// Registrar detalhes do acesso
	access := &storage.ShareLinkAccess{
		ID:          uuid.New().String(),
		ShareLinkID: link.ID,
		IPAddress:   ip,
		UserAgent:   userAgent,
		AccessedAt:  time.Now(),
	}
	h.store.RecordShareLinkAccess(access)

	// Log de auditoria
	h.auditLogger.LogDataAccess(link.UserID, ip, "shared/"+link.ID, "access", "success")
}

// generateSecureToken gera um token seguro para o link
func generateSecureToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:16]) // 32 caracteres hex
}

// =============================================================================
// ENDPOINT DE ACESSO DO GUARDIÃO (Nova Arquitetura)
// =============================================================================

// GuardianAccessResponse representa a resposta para acesso do guardião
type GuardianAccessResponse struct {
	Guardian   *GuardianInfo     `json:"guardian"`
	Owner      *OwnerInfo        `json:"owner"`
	Items      []*SharedItemInfo `json:"items"`
	AccessType string            `json:"access_type"`
	AccessedAt time.Time         `json:"accessed_at"`
}

// GuardianInfo representa info do guardião
type GuardianInfo struct {
	Name         string `json:"name"`
	Relationship string `json:"relationship"`
}

// OwnerInfo representa info do dono da caixa
type OwnerInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// SharedItemInfo representa um item compartilhado
type SharedItemInfo struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Category    string    `json:"category,omitempty"`
	Recipient   string    `json:"recipient,omitempty"`
	IsImportant bool      `json:"is_important,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// AccessGuardianView permite acessar itens compartilhados via token do guardião
// GET /api/guardian-access/:token
func (h *Handler) AccessGuardianView(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "share.invalid_token"))
		return
	}

	// Buscar guardião pelo token
	guardian, err := h.store.GetGuardianByAccessToken(token)
	if err != nil {
		writeError(w, http.StatusNotFound, i18n.Tr(r, "share.link_not_found"))
		return
	}

	// Buscar dono da caixa
	owner, found := h.store.GetUserByID(guardian.UserID)
	if !found {
		writeError(w, http.StatusNotFound, i18n.Tr(r, "share.link_not_found"))
		return
	}

	// SEGURANÇA: Verificar se o guardião tem PIN configurado
	// Se tiver, exigir verificação antes de mostrar conteúdo
	if guardian.AccessPIN != "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"requires_pin": true,
			"guardian": map[string]string{
				"name":         guardian.Name,
				"relationship": guardian.Relationship,
			},
			"owner": map[string]string{
				"name": owner.Name,
			},
		})
		return
	}

	// Retornar conteúdo completo
	h.returnGuardianContent(w, r, guardian, owner)
}

// VerifyGuardianPIN verifica o PIN do guardião e retorna o conteúdo
// POST /api/guardian-access/:token/verify
func (h *Handler) VerifyGuardianPIN(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "share.invalid_token"))
		return
	}

	var req VerifyPINRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "share.invalid_data"))
		return
	}

	// Buscar guardião pelo token
	guardian, err := h.store.GetGuardianByAccessToken(token)
	if err != nil {
		writeError(w, http.StatusNotFound, i18n.Tr(r, "share.link_not_found"))
		return
	}

	// Verificar PIN
	if err := bcrypt.CompareHashAndPassword([]byte(guardian.AccessPIN), []byte(req.PIN)); err != nil {
		writeError(w, http.StatusUnauthorized, i18n.Tr(r, "share.invalid_pin"))
		return
	}

	// Buscar dono da caixa
	owner, found := h.store.GetUserByID(guardian.UserID)
	if !found {
		writeError(w, http.StatusNotFound, i18n.Tr(r, "share.link_not_found"))
		return
	}

	// Retornar conteúdo completo
	h.returnGuardianContent(w, r, guardian, owner)
}

// returnGuardianContent retorna o conteúdo da caixa para o guardião
func (h *Handler) returnGuardianContent(w http.ResponseWriter, r *http.Request, guardian *storage.Guardian, owner *storage.User) {
	// IMPORTANTE: Buscar apenas itens COMPARTILHADOS (is_shared = true)
	// Itens não compartilhados são privados e não devem ser expostos
	sharedItems := h.store.ListSharedItems(guardian.UserID)

	// Converter para resposta
	items := make([]*SharedItemInfo, 0, len(sharedItems))
	for _, item := range sharedItems {
		items = append(items, &SharedItemInfo{
			ID:          item.ID,
			Type:        string(item.Type),
			Title:       item.Title,
			Content:     item.Content,
			Category:    item.Category,
			Recipient:   item.Recipient,
			IsImportant: item.IsImportant,
			CreatedAt:   item.CreatedAt,
		})
	}

	response := &GuardianAccessResponse{
		Guardian: &GuardianInfo{
			Name:         guardian.Name,
			Relationship: guardian.Relationship,
		},
		Owner: &OwnerInfo{
			Name:  owner.Name,
			Email: owner.Email,
		},
		Items:      items,
		AccessType: string(guardian.AccessType),
		AccessedAt: time.Now(),
	}

	// Log de acesso
	if h.auditLogger != nil {
		h.auditLogger.Log(security.AuditEvent{
			Type:     security.EventDataAccess,
			UserID:   guardian.UserID,
			ClientIP: security.GetClientIP(r),
			Resource: "guardian_access",
			Action:   "view_shared_items",
			Result:   "success",
			Details: map[string]interface{}{
				"guardian_id":   guardian.ID,
				"guardian_name": guardian.Name,
				"items_count":   len(items),
			},
		})
	}

	writeJSON(w, http.StatusOK, response)
}

// getBaseURL retorna a URL base da aplicação
func getBaseURL(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	return scheme + "://" + r.Host
}

// contains verifica se um slice contém um valor
func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// writeJSON escreve uma resposta JSON
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// writeError escreve uma resposta de erro
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
