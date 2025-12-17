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

	// Box Items
	GetBoxItems(userID string) ([]*BoxItem, error)
	ListBoxItems(userID string) []*BoxItem
	GetBoxItem(userID, itemID string) (*BoxItem, error)
	CreateBoxItem(userID string, item *BoxItem) (*BoxItem, error)
	UpdateBoxItem(userID, itemID string, updates *BoxItem) (*BoxItem, error)
	DeleteBoxItem(userID, itemID string) error

	// Guardians
	GetGuardians(userID string) ([]*Guardian, error)
	ListGuardians(userID string) []*Guardian
	CreateGuardian(userID string, guardian *Guardian) (*Guardian, error)
	UpdateGuardian(userID, guardianID string, updates *Guardian) (*Guardian, error)
	DeleteGuardian(userID, guardianID string) error

	// Guide Progress
	GetGuideProgress(userID string) map[string]*GuideProgress
	UpdateGuideProgress(userID, cardID, status string) (*GuideProgress, error)

	// Settings
	GetSettings(userID string) *Settings
	UpdateSettings(userID string, updates *Settings) *Settings

	// Admin
	GetStats() *Stats
	ListUsers() []*User
}

// Garantir que as implementações satisfazem a interface
var _ Store = (*MemoryStore)(nil)
var _ Store = (*PostgresStore)(nil)
