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

// AuthProvider define os provedores de autenticação suportados
type AuthProvider string

const (
	AuthProviderEmail  AuthProvider = "email"  // Login com email/senha
	AuthProviderGoogle AuthProvider = "google" // Login com Google
	AuthProviderApple  AuthProvider = "apple"  // Login com Apple
)

// User representa um usuário do Famli
type User struct {
	ID         string       `json:"id"`
	Email      string       `json:"email"`
	Name       string       `json:"name,omitempty"`
	Password   string       `json:"-"`
	Provider   AuthProvider `json:"provider,omitempty"`    // Provedor de autenticação
	ProviderID string       `json:"provider_id,omitempty"` // ID do usuário no provedor
	AvatarURL  string       `json:"avatar_url,omitempty"`  // URL do avatar (Google/Apple)
	Locale     string       `json:"locale,omitempty"`      // Idioma preferido (ex: "pt-BR", "en")
	CreatedAt  time.Time    `json:"created_at"`
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
	IsShared    bool      `json:"is_shared"` // Se o item é visível para guardiões
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

// GuardianAccessType define os tipos de acesso do guardião
type GuardianAccessType string

const (
	GuardianAccessNormal    GuardianAccessType = "normal"    // Acesso normal (sempre disponível)
	GuardianAccessEmergency GuardianAccessType = "emergency" // Apenas em emergência
	GuardianAccessMemorial  GuardianAccessType = "memorial"  // Apenas após falecimento
)

// Guardian representa uma pessoa de confiança
type Guardian struct {
	ID           string             `json:"id"`
	UserID       string             `json:"user_id"`
	Name         string             `json:"name"`
	Email        string             `json:"email"`
	Phone        string             `json:"phone,omitempty"`
	Relationship string             `json:"relationship,omitempty"` // filho, neto, amigo, etc.
	Role         string             `json:"role"`                   // viewer, coauthor (futuro)
	Notes        string             `json:"notes,omitempty"`        // explicação do papel
	AccessToken  string             `json:"access_token"`           // Token único para acesso (sempre retornado)
	AccessPIN    string             `json:"-"`                      // PIN de proteção (hash) - não expor no JSON
	HasPIN       bool               `json:"has_pin"`                // Indica se tem PIN configurado
	AccessType   GuardianAccessType `json:"access_type,omitempty"`  // Tipo de acesso
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
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

// =============================================================================
// COMPARTILHAMENTO E ACESSO
// =============================================================================

// ShareLinkType define os tipos de link de compartilhamento
type ShareLinkType string

const (
	ShareLinkNormal    ShareLinkType = "normal"    // Acesso normal (guardião pode ver)
	ShareLinkEmergency ShareLinkType = "emergency" // Acesso de emergência (protocolo ativado)
	ShareLinkMemorial  ShareLinkType = "memorial"  // Acesso memorial (pós-falecimento)
)

// ShareLink representa um link de compartilhamento para guardiões
type ShareLink struct {
	ID          string        `json:"id"`
	UserID      string        `json:"user_id"`
	GuardianID  string        `json:"guardian_id,omitempty"`  // Se vinculado a um guardião específico (deprecated)
	GuardianIDs []string      `json:"guardian_ids,omitempty"` // Guardiões específicos que podem acessar
	Token       string        `json:"-"`                      // Token secreto (não expor na API)
	Type        ShareLinkType `json:"type"`
	Name        string        `json:"name"`       // Nome para identificar o link
	PIN         string        `json:"-"`          // PIN opcional para acesso (hash)
	Categories  []string      `json:"categories"` // Categorias permitidas (vazio = todas)
	ExpiresAt   *time.Time    `json:"expires_at"` // Nulo = nunca expira
	MaxUses     int           `json:"max_uses"`   // 0 = ilimitado
	UsageCount  int           `json:"usage_count"`
	LastUsedAt  *time.Time    `json:"last_used_at"`
	IsActive    bool          `json:"is_active"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// ShareLinkAccess registra cada acesso a um link de compartilhamento
type ShareLinkAccess struct {
	ID          string    `json:"id"`
	ShareLinkID string    `json:"share_link_id"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	AccessedAt  time.Time `json:"accessed_at"`
}

// PasswordResetToken representa um token de recuperação de senha
type PasswordResetToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Token     string     `json:"-"` // Token secreto (hash)
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// EmergencyProtocol representa o estado do protocolo de emergência
type EmergencyProtocol struct {
	UserID          string     `json:"user_id"`
	IsActive        bool       `json:"is_active"`
	ActivatedAt     *time.Time `json:"activated_at"`
	ActivatedBy     string     `json:"activated_by,omitempty"` // ID do guardião que ativou
	DeactivatedAt   *time.Time `json:"deactivated_at"`
	Reason          string     `json:"reason,omitempty"` // Motivo da ativação
	NotifyGuardians bool       `json:"notify_guardians"` // Notificar outros guardiões
}

// SharedView representa a visualização compartilhada para um guardião
type SharedView struct {
	UserName     string        `json:"user_name"`
	UserEmail    string        `json:"user_email,omitempty"`    // Apenas se autorizado
	GuardianName string        `json:"guardian_name,omitempty"` // Nome do guardião que está acessando
	Items        []*BoxItem    `json:"items"`
	Guardians    []*Guardian   `json:"guardians,omitempty"` // Apenas em modo memorial
	Message      string        `json:"message,omitempty"`   // Mensagem personalizada
	LinkType     ShareLinkType `json:"link_type"`
	AccessedAt   time.Time     `json:"accessed_at"`
}
