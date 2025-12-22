// =============================================================================
// FAMLI - Interface de Storage
// =============================================================================
// Define a interface comum para todos os backends de armazenamento.
// Isso permite trocar entre MemoryStore (desenvolvimento) e PostgresStore
// (produção) sem modificar o código dos handlers.
// =============================================================================

package storage

// Store define a interface para armazenamento de dados
type Store interface {
	// Users
	CreateUser(email, hashedPassword, name string) (*User, error)
	GetUserByEmail(email string) (*User, bool)
	GetUserByID(id string) (*User, bool)
	UpdateUserPassword(userID, hashedPassword string) error
	UpdateUserLocale(userID, locale string) error // Atualiza idioma preferido
	DeleteUser(userID string) error               // LGPD: Direito ao esquecimento

	// Social Auth (Google, Apple)
	CreateOrUpdateSocialUser(provider AuthProvider, providerID, email, name, avatarURL string) (*User, error)
	GetUserByProvider(provider AuthProvider, providerID string) (*User, bool)
	LinkSocialProvider(userID string, provider AuthProvider, providerID string) error

	// Box Items (métodos legacy para compatibilidade)
	GetBoxItems(userID string) ([]*BoxItem, error)
	ListBoxItems(userID string) []*BoxItem
	GetBoxItem(userID, itemID string) (*BoxItem, error)
	CreateBoxItem(userID string, item *BoxItem) (*BoxItem, error)
	UpdateBoxItem(userID, itemID string, updates *BoxItem) (*BoxItem, error)
	DeleteBoxItem(userID, itemID string) error

	// Box Items (métodos paginados - preferir estes)
	ListBoxItemsPaginated(userID string, params *PaginationParams) (*PaginatedResult[*BoxItemSummary], error)
	CountBoxItems(userID string) (int, error)

	// Guardians (métodos legacy para compatibilidade)
	GetGuardians(userID string) ([]*Guardian, error)
	ListGuardians(userID string) []*Guardian
	CreateGuardian(userID string, guardian *Guardian) (*Guardian, error)
	UpdateGuardian(userID, guardianID string, updates *Guardian) (*Guardian, error)
	DeleteGuardian(userID, guardianID string) error

	// Guardians (métodos paginados)
	ListGuardiansPaginated(userID string, params *PaginationParams) (*PaginatedResult[*Guardian], error)
	CountGuardians(userID string) (int, error)

	// Guardian Access (acesso via token do guardião)
	GetGuardianByAccessToken(token string) (*Guardian, error)
	ListSharedItems(userID string) []*BoxItem // Lista itens com is_shared = true

	// Guide Progress
	GetGuideProgress(userID string) map[string]*GuideProgress
	UpdateGuideProgress(userID, cardID, status string) (*GuideProgress, error)

	// Settings
	GetSettings(userID string) *Settings
	UpdateSettings(userID string, updates *Settings) *Settings

	// Admin
	GetStats() *Stats
	ListUsers() []*User

	// User Data Export (LGPD: Portabilidade)
	ExportUserData(userID string) (*UserDataExport, error)

	// Feedback
	CreateFeedback(f *Feedback) error
	ListFeedbacks(status string, limit int) ([]*Feedback, error)
	UpdateFeedbackStatus(id, status, adminNote string) error
	GetFeedbackStats() (total, pending int)

	// Analytics
	TrackEvent(e *AnalyticsEvent) error
	GetAnalyticsSummary() *AnalyticsSummary
	GetRecentEvents(limit int) ([]*AnalyticsEvent, error)
	GetDailyStats(days int) ([]map[string]interface{}, error)

	// Share Links (Compartilhamento com Guardiões)
	CreateShareLink(link *ShareLink) error
	GetShareLinkByToken(token string) (*ShareLink, error)
	GetShareLinksByUser(userID string) ([]*ShareLink, error)
	UpdateShareLink(link *ShareLink) error
	DeleteShareLink(userID, linkID string) error
	RecordShareLinkAccess(access *ShareLinkAccess) error
	IncrementShareLinkUsage(linkID string) error

	// Password Reset (Recuperação de Senha)
	CreatePasswordResetToken(token *PasswordResetToken) error
	GetPasswordResetToken(tokenHash string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(tokenID string) error
	CleanupExpiredPasswordResetTokens() error

	// Emergency Protocol (Protocolo de Emergência)
	GetEmergencyProtocol(userID string) (*EmergencyProtocol, error)
	UpdateEmergencyProtocol(protocol *EmergencyProtocol) error

	// Maintenance
	CleanupOldLogs(retentionDays int) error
}

// Garantir que as implementações satisfazem a interface
var _ Store = (*MemoryStore)(nil)
var _ Store = (*PostgresStore)(nil)
