package storage

import "time"

// =============================================================================
// PAGINAÇÃO
// =============================================================================

// PaginationParams define os parâmetros de paginação (cursor-based)
// Cursor-based é mais eficiente que OFFSET para grandes datasets
type PaginationParams struct {
	Cursor string `json:"cursor,omitempty"` // ID do último item (para next page)
	Limit  int    `json:"limit"`            // Número de itens por página (max 50)
}

// PaginatedResult representa o resultado paginado
type PaginatedResult[T any] struct {
	Items      []T    `json:"items"`                 // Itens da página atual
	NextCursor string `json:"next_cursor,omitempty"` // Cursor para próxima página
	HasMore    bool   `json:"has_more"`              // Indica se há mais páginas
	Total      int    `json:"total,omitempty"`       // Total de itens (opcional)
}

// DefaultPageSize é o tamanho padrão de página
const DefaultPageSize = 20

// MaxPageSize é o tamanho máximo de página
const MaxPageSize = 50

// NormalizePagination normaliza os parâmetros de paginação
func NormalizePagination(p *PaginationParams) *PaginationParams {
	if p == nil {
		return &PaginationParams{Limit: DefaultPageSize}
	}
	if p.Limit <= 0 {
		p.Limit = DefaultPageSize
	}
	if p.Limit > MaxPageSize {
		p.Limit = MaxPageSize
	}
	return p
}

// =============================================================================
// USUÁRIOS
// =============================================================================

// User representa um usuário do Famli
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name,omitempty"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// ItemType define os tipos de itens na Caixa Famli
type ItemType string

const (
	ItemTypeInfo     ItemType = "info"     // Informação importante
	ItemTypeMemory   ItemType = "memory"   // Memória/mensagem
	ItemTypeNote     ItemType = "note"     // Nota pessoal
	ItemTypeAccess   ItemType = "access"   // Instruções de acesso (não senhas!)
	ItemTypeRoutine  ItemType = "routine"  // Rotina que não pode parar
	ItemTypeLocation ItemType = "location" // Onde estão as coisas
)

// BoxItem representa um item na Caixa Famli
// Campos sensíveis (Title, Content, Recipient) são armazenados criptografados
type BoxItem struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Type        ItemType  `json:"type"`
	Title       string    `json:"title"`               // Criptografado no banco
	Content     string    `json:"content"`             // Criptografado no banco
	Category    string    `json:"category,omitempty"`  // saúde, finanças, família, etc.
	Recipient   string    `json:"recipient,omitempty"` // Criptografado no banco
	IsImportant bool      `json:"is_important"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BoxItemSummary é uma versão resumida do item para listagens
// Não inclui o Content completo para economizar dados
type BoxItemSummary struct {
	ID          string    `json:"id"`
	Type        ItemType  `json:"type"`
	Title       string    `json:"title"`
	Category    string    `json:"category,omitempty"`
	IsImportant bool      `json:"is_important"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Guardian representa uma pessoa de confiança
type Guardian struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone,omitempty"`
	Relationship string    `json:"relationship,omitempty"` // filho, neto, amigo, etc.
	Role         string    `json:"role"`                   // viewer, coauthor (futuro)
	Notes        string    `json:"notes,omitempty"`        // explicação do papel
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GuideCard representa um card do Guia Famli
type GuideCard struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Order       int    `json:"order"`
	ItemType    string `json:"item_type,omitempty"` // tipo de item relacionado
}

// GuideProgress armazena o progresso do usuário no Guia
type GuideProgress struct {
	UserID      string    `json:"user_id"`
	CardID      string    `json:"card_id"`
	Status      string    `json:"status"` // pending, started, completed, skipped
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// Settings armazena as configurações do usuário
type Settings struct {
	UserID                   string `json:"user_id"`
	EmergencyProtocolEnabled bool   `json:"emergency_protocol_enabled"`
	NotificationsEnabled     bool   `json:"notifications_enabled"`
	Theme                    string `json:"theme"` // light, dark, auto
}

// UserDataExport representa todos os dados do usuário para exportação (LGPD)
type UserDataExport struct {
	User       *User            `json:"user"`
	Items      []*BoxItem       `json:"items"`
	Guardians  []*Guardian      `json:"guardians"`
	Progress   []*GuideProgress `json:"guide_progress"`
	Settings   *Settings        `json:"settings"`
	ExportedAt time.Time        `json:"exported_at"`
}

// =============================================================================
// FEEDBACK
// =============================================================================

// FeedbackType define os tipos de feedback
type FeedbackType string

const (
	FeedbackSuggestion FeedbackType = "suggestion" // Sugestão
	FeedbackProblem    FeedbackType = "problem"    // Problema/Bug
	FeedbackPraise     FeedbackType = "praise"     // Elogio
	FeedbackQuestion   FeedbackType = "question"   // Dúvida
)

// Feedback representa um feedback do usuário
type Feedback struct {
	ID        string       `json:"id"`
	UserID    string       `json:"user_id"`
	UserEmail string       `json:"user_email,omitempty"` // Para exibição no admin
	Type      FeedbackType `json:"type"`
	Message   string       `json:"message"`
	Page      string       `json:"page,omitempty"` // Página onde estava
	UserAgent string       `json:"user_agent,omitempty"`
	Status    string       `json:"status"` // pending, reviewed, resolved
	AdminNote string       `json:"admin_note,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// =============================================================================
// ANALYTICS
// =============================================================================

// AnalyticsEventType define os tipos de eventos rastreados
type AnalyticsEventType string

const (
	EventPageView       AnalyticsEventType = "page_view"
	EventLogin          AnalyticsEventType = "login"
	EventRegister       AnalyticsEventType = "register"
	EventCreateItem     AnalyticsEventType = "create_item"
	EventEditItem       AnalyticsEventType = "edit_item"
	EventDeleteItem     AnalyticsEventType = "delete_item"
	EventCreateGuardian AnalyticsEventType = "create_guardian"
	EventCompleteGuide  AnalyticsEventType = "complete_guide"
	EventExportData     AnalyticsEventType = "export_data"
	EventSendFeedback   AnalyticsEventType = "send_feedback"
)

// AnalyticsEvent representa um evento de analytics
type AnalyticsEvent struct {
	ID        string             `json:"id"`
	UserID    string             `json:"user_id,omitempty"`
	EventType AnalyticsEventType `json:"event_type"`
	Page      string             `json:"page,omitempty"`
	Details   map[string]string  `json:"details,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
}

// AnalyticsSummary representa o resumo de analytics para o dashboard
type AnalyticsSummary struct {
	// Usuários
	TotalUsers       int `json:"total_users"`
	ActiveToday      int `json:"active_today"`
	ActiveThisWeek   int `json:"active_this_week"`
	NewUsersToday    int `json:"new_users_today"`
	NewUsersThisWeek int `json:"new_users_this_week"`

	// Ações
	TotalItems        int `json:"total_items"`
	ItemsCreatedToday int `json:"items_created_today"`
	TotalGuardians    int `json:"total_guardians"`

	// Engajamento
	EventsToday  int            `json:"events_today"`
	EventsByType map[string]int `json:"events_by_type"`

	// Feedbacks
	TotalFeedbacks   int `json:"total_feedbacks"`
	PendingFeedbacks int `json:"pending_feedbacks"`
}
