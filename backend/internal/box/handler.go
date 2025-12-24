// =============================================================================
// FAMLI - Handler da Caixa Famli
// =============================================================================
// Este módulo gerencia os itens da Caixa Famli (memórias, documentos, notas).
//
// Segurança implementada:
// - Validação de inputs (OWASP A03)
// - Sanitização contra XSS
// - Limite de tamanho de conteúdo
// - Auditoria de acesso a dados
// - Isolamento por usuário (A01)
// =============================================================================

package box

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"famli/internal/auth"
	"famli/internal/i18n"
	"famli/internal/security"
	"famli/internal/storage"
)

// =============================================================================
// HANDLER
// =============================================================================

// Handler gerencia endpoints da Caixa Famli
type Handler struct {
	// store é o armazenamento de dados
	store storage.Store

	// auditLogger registra eventos de acesso
	auditLogger *security.AuditLogger
}

// NewHandler cria uma nova instância do handler
//
// Parâmetros:
//   - store: armazenamento de dados
//
// Retorna:
//   - *Handler: handler configurado
func NewHandler(store storage.Store) *Handler {
	return &Handler{
		store:       store,
		auditLogger: security.GetAuditLogger(),
	}
}

// =============================================================================
// PAYLOADS
// =============================================================================

// itemPayload representa o payload de criação/atualização de item
type itemPayload struct {
	Type        storage.ItemType `json:"type"`
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	Category    string           `json:"category,omitempty"`
	Recipient   string           `json:"recipient,omitempty"`
	IsImportant bool             `json:"is_important"`
	IsShared    bool             `json:"is_shared"` // Compartilhado com guardiões
	GuardianIDs []string         `json:"guardian_ids,omitempty"`
}

// validate valida e sanitiza o payload
//
// Retorna:
//   - string: mensagem de erro (vazia se válido)
func (p *itemPayload) validate(r *http.Request) string {
	// Sanitizar título
	p.Title = security.SanitizeTitle(p.Title)
	if p.Title == "" {
		return i18n.Tr(r, "box.title_required")
	}

	// Verificar tamanho do título
	if len(p.Title) > security.MaxTitleLength {
		return i18n.Tr(r, "box.title_too_long")
	}

	// Sanitizar conteúdo
	p.Content = security.SanitizeContent(p.Content)

	// Verificar tamanho do conteúdo
	if len(p.Content) > security.MaxContentLength {
		return i18n.Tr(r, "box.content_too_long")
	}

	// Sanitizar categoria
	p.Category = sanitizeCategory(p.Category)

	// Sanitizar destinatário
	p.Recipient = security.SanitizeName(p.Recipient)

	// Validar tipo
	if !isValidItemType(p.Type) {
		p.Type = storage.ItemTypeInfo
	}

	// Verificar por tentativas de injection
	if security.ContainsSQLInjection(p.Title) || security.ContainsSQLInjection(p.Content) {
		return i18n.Tr(r, "box.invalid_detected")
	}

	if !p.IsShared {
		p.GuardianIDs = nil
		return ""
	}

	if len(p.GuardianIDs) > 0 {
		unique := make([]string, 0, len(p.GuardianIDs))
		seen := make(map[string]struct{}, len(p.GuardianIDs))
		for _, id := range p.GuardianIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			unique = append(unique, id)
		}
		p.GuardianIDs = unique
	}

	return ""
}

// =============================================================================
// ENDPOINTS
// =============================================================================

// List retorna todos os itens da Caixa Famli do usuário
//
// Endpoint: GET /api/box/items
//
// Segurança:
// - Requer autenticação JWT
// - Retorna apenas itens do usuário autenticado (A01)
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	clientIP := security.GetClientIP(r)

	// Parâmetros de paginação
	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")

	limit := storage.DefaultPageSize
	if limitStr != "" {
		if parsed, err := parseInt(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Buscar itens com paginação
	params := &storage.PaginationParams{
		Cursor: cursor,
		Limit:  limit,
	}

	result, err := h.store.ListBoxItemsPaginated(userID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "box.list_error"))
		return
	}

	// Contar total (opcional, apenas na primeira página)
	var total int
	if cursor == "" {
		total, _ = h.store.CountBoxItems(userID)
	}

	// Registrar acesso (auditoria)
	h.auditLogger.LogDataAccess(userID, clientIP, "box/items", "list", "success")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items":       result.Items,
		"next_cursor": result.NextCursor,
		"has_more":    result.HasMore,
		"total":       total,
	})
}

// Create adiciona um novo item à Caixa Famli
//
// Endpoint: POST /api/box/items
//
// Segurança:
// - Requer autenticação JWT
// - Validação e sanitização de inputs
// - Limite de tamanho de conteúdo
// - Auditoria de criação
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	clientIP := security.GetClientIP(r)

	// Limitar tamanho do body (previne DoS)
	r.Body = http.MaxBytesReader(w, r.Body, 100*1024) // 100KB max

	// Decodificar payload
	var payload itemPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "box.invalid_content"))
		return
	}

	// Validar e sanitizar
	if errMsg := payload.validate(r); errMsg != "" {
		writeError(w, http.StatusBadRequest, errMsg)
		return
	}

	// Criar item
	item := &storage.BoxItem{
		Type:        payload.Type,
		Title:       payload.Title,
		Content:     payload.Content,
		Category:    payload.Category,
		Recipient:   payload.Recipient,
		IsImportant: payload.IsImportant,
		IsShared:    payload.IsShared,
		GuardianIDs: payload.GuardianIDs,
	}

	idempotencyKey := getIdempotencyKey(r)
	if len(idempotencyKey) > 120 {
		idempotencyKey = idempotencyKey[:120]
	}

	var itemID string
	if idempotencyKey != "" {
		itemID = fmt.Sprintf("itm_%d", time.Now().UnixNano())
		existingID, inserted, err := h.store.RegisterIdempotencyKey(userID, idempotencyKey, "box_item", itemID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, i18n.Tr(r, "box.save_error"))
			return
		}
		if !inserted {
			existing, err := h.store.GetBoxItem(userID, existingID)
			if err != nil {
				writeError(w, http.StatusConflict, i18n.Tr(r, "box.save_error"))
				return
			}
			w.Header().Set("Idempotency-Replayed", "true")
			writeJSON(w, http.StatusOK, existing)
			return
		}
	}

	var created *storage.BoxItem
	var err error
	if idempotencyKey != "" {
		created, err = h.store.CreateBoxItemWithID(userID, item, itemID)
	} else {
		created, err = h.store.CreateBoxItem(userID, item)
	}
	if err != nil {
		h.auditLogger.LogDataAccess(userID, clientIP, "box/items", "create", "failure")
		if idempotencyKey != "" {
			_ = h.store.DeleteIdempotencyKey(userID, idempotencyKey, "box_item")
		}
		writeError(w, http.StatusInternalServerError, i18n.Tr(r, "box.save_error"))
		return
	}

	// Registrar criação (auditoria)
	h.auditLogger.LogDataAccess(userID, clientIP, "box/items/"+created.ID, "create", "success")

	writeJSON(w, http.StatusCreated, created)
}

// Update modifica um item existente
//
// Endpoint: PUT /api/box/items/{itemID}
//
// Segurança:
// - Requer autenticação JWT
// - Verifica propriedade do item (A01)
// - Validação e sanitização de inputs
// - Auditoria de atualização
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	clientIP := security.GetClientIP(r)
	itemID := chi.URLParam(r, "itemID")

	// Sanitizar itemID (previne path traversal)
	itemID = sanitizeID(itemID)

	// Limitar tamanho do body
	r.Body = http.MaxBytesReader(w, r.Body, 100*1024)

	// Decodificar payload
	var payload itemPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "box.invalid_content"))
		return
	}

	// Validar e sanitizar
	if errMsg := payload.validate(r); errMsg != "" {
		writeError(w, http.StatusBadRequest, errMsg)
		return
	}

	// Atualizar item
	updates := &storage.BoxItem{
		Type:        payload.Type,
		Title:       payload.Title,
		Content:     payload.Content,
		Category:    payload.Category,
		Recipient:   payload.Recipient,
		IsImportant: payload.IsImportant,
		IsShared:    payload.IsShared,
		GuardianIDs: payload.GuardianIDs,
	}

	updated, err := h.store.UpdateBoxItem(userID, itemID, updates)
	if err != nil {
		// Não revelar se o item existe mas pertence a outro usuário
		h.auditLogger.LogSecurity(security.EventUnauthorizedAccess, clientIP, map[string]interface{}{
			"user_id":  userID,
			"item_id":  itemID,
			"resource": "box/items",
		})
		writeError(w, http.StatusNotFound, i18n.Tr(r, "box.not_found"))
		return
	}

	// Registrar atualização (auditoria)
	h.auditLogger.LogDataAccess(userID, clientIP, "box/items/"+itemID, "update", "success")

	writeJSON(w, http.StatusOK, updated)
}

// Delete remove um item da Caixa Famli
//
// Endpoint: DELETE /api/box/items/{itemID}
//
// Segurança:
// - Requer autenticação JWT
// - Verifica propriedade do item (A01)
// - Auditoria de deleção
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	clientIP := security.GetClientIP(r)
	itemID := chi.URLParam(r, "itemID")

	// Sanitizar itemID
	itemID = sanitizeID(itemID)

	// Deletar item
	if err := h.store.DeleteBoxItem(userID, itemID); err != nil {
		h.auditLogger.LogSecurity(security.EventUnauthorizedAccess, clientIP, map[string]interface{}{
			"user_id":  userID,
			"item_id":  itemID,
			"resource": "box/items",
		})
		writeError(w, http.StatusNotFound, i18n.Tr(r, "box.not_found"))
		return
	}

	// Registrar deleção (auditoria)
	h.auditLogger.LogDataAccess(userID, clientIP, "box/items/"+itemID, "delete", "success")

	writeJSON(w, http.StatusOK, map[string]string{"message": i18n.Tr(r, "box.deleted")})
}

// =============================================================================
// ASSISTENTE
// =============================================================================

// Assistant responde perguntas do usuário de forma gentil
//
// Endpoint: POST /api/assistant
//
// Segurança:
// - Requer autenticação JWT
// - Sanitização de input
// - Limite de tamanho
func (h *Handler) Assistant(w http.ResponseWriter, r *http.Request) {
	// Limitar tamanho do body
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024) // 10KB max

	var payload struct {
		Input string `json:"input"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "assistant.empty_input"))
		return
	}

	// Sanitizar e validar input
	input := security.SanitizeText(payload.Input, 1000)
	if strings.TrimSpace(input) == "" {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "assistant.empty_input"))
		return
	}

	// Verificar por conteúdo malicioso
	if security.ContainsSQLInjection(input) {
		writeError(w, http.StatusBadRequest, i18n.Tr(r, "box.invalid_query"))
		return
	}

	reply := buildAssistantReply(r, input)
	writeJSON(w, http.StatusOK, map[string]string{"reply": reply})
}

// buildAssistantReply gera resposta do assistente baseada na pergunta
func buildAssistantReply(r *http.Request, input string) string {
	normalized := strings.ToLower(input)

	switch {
	case strings.Contains(normalized, "começar") || strings.Contains(normalized, "primeiro") ||
		strings.Contains(normalized, "start") || strings.Contains(normalized, "first"):
		return i18n.Tr(r, "assistant.start")

	case strings.Contains(normalized, "senha") || strings.Contains(normalized, "acesso") ||
		strings.Contains(normalized, "password") || strings.Contains(normalized, "access"):
		return i18n.Tr(r, "assistant.passwords")

	case strings.Contains(normalized, "pessoa") || strings.Contains(normalized, "confiança") ||
		strings.Contains(normalized, "person") || strings.Contains(normalized, "trust"):
		return i18n.Tr(r, "assistant.guardians")

	case strings.Contains(normalized, "documento") || strings.Contains(normalized, "saúde") || strings.Contains(normalized, "plano") ||
		strings.Contains(normalized, "document") || strings.Contains(normalized, "health") || strings.Contains(normalized, "plan"):
		return i18n.Tr(r, "assistant.documents")

	case strings.Contains(normalized, "mensagem") || strings.Contains(normalized, "memória") || strings.Contains(normalized, "recado") ||
		strings.Contains(normalized, "message") || strings.Contains(normalized, "memory"):
		return i18n.Tr(r, "assistant.memories")

	case strings.Contains(normalized, "seguro") || strings.Contains(normalized, "privacidade") ||
		strings.Contains(normalized, "security") || strings.Contains(normalized, "privacy"):
		return i18n.Tr(r, "assistant.security")

	case strings.Contains(normalized, "ajuda") || strings.Contains(normalized, "como") ||
		strings.Contains(normalized, "help") || strings.Contains(normalized, "how"):
		return i18n.Tr(r, "assistant.help")

	default:
		return i18n.Tr(r, "assistant.default")
	}
}

// =============================================================================
// FUNÇÕES AUXILIARES
// =============================================================================

// writeJSON escreve resposta JSON com headers de segurança
func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	security.SetJSONHeaders(w)
	w.WriteHeader(status)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

// writeError escreve resposta de erro JSON
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func getIdempotencyKey(r *http.Request) string {
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		key = strings.TrimSpace(r.Header.Get("X-Idempotency-Key"))
	}
	if len(key) > 120 {
		key = key[:120]
	}
	return key
}

// parseInt converte string para int
func parseInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	result := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}

// sanitizeID sanitiza IDs para prevenir path traversal
func sanitizeID(id string) string {
	// Remover caracteres perigosos
	id = strings.ReplaceAll(id, "..", "")
	id = strings.ReplaceAll(id, "/", "")
	id = strings.ReplaceAll(id, "\\", "")
	id = strings.ReplaceAll(id, "\x00", "")

	// Manter apenas caracteres alfanuméricos e underscore
	result := ""
	for _, c := range id {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			result += string(c)
		}
	}

	return result
}

// sanitizeCategory sanitiza e normaliza categoria
func sanitizeCategory(category string) string {
	category = strings.TrimSpace(strings.ToLower(category))

	// Categorias válidas
	validCategories := map[string]string{
		"saude":      "saúde",
		"saúde":      "saúde",
		"financas":   "finanças",
		"finanças":   "finanças",
		"familia":    "família",
		"família":    "família",
		"documentos": "documentos",
		"memorias":   "memórias",
		"memórias":   "memórias",
		"outros":     "outros",
	}

	if normalized, ok := validCategories[category]; ok {
		return normalized
	}

	// Se não for uma categoria válida, retornar "outros"
	if category != "" {
		return "outros"
	}

	return ""
}

// isValidItemType verifica se o tipo de item é válido
func isValidItemType(t storage.ItemType) bool {
	validTypes := map[storage.ItemType]bool{
		storage.ItemTypeInfo:     true,
		storage.ItemTypeMemory:   true,
		storage.ItemTypeNote:     true,
		storage.ItemTypeAccess:   true,
		storage.ItemTypeRoutine:  true,
		storage.ItemTypeLocation: true,
	}
	return validTypes[t]
}
