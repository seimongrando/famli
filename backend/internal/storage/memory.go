package storage

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound      = errors.New("não encontrado")
	ErrAlreadyExists = errors.New("já existe")
	ErrInvalidData   = errors.New("dados inválidos")
)

// MemoryStore implementa armazenamento em memória para o MVP
type MemoryStore struct {
	mu sync.RWMutex

	users        map[string]*User
	usersByEmail map[string]string
	items        map[string]map[string]*BoxItem       // userID -> itemID -> item
	guardians    map[string]map[string]*Guardian      // userID -> guardianID -> guardian
	progress     map[string]map[string]*GuideProgress // userID -> cardID -> progress
	settings     map[string]*Settings

	userSeq     int64
	itemSeq     int64
	guardianSeq int64
}

// NewMemoryStore cria uma nova instância do store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:        make(map[string]*User),
		usersByEmail: make(map[string]string),
		items:        make(map[string]map[string]*BoxItem),
		guardians:    make(map[string]map[string]*Guardian),
		progress:     make(map[string]map[string]*GuideProgress),
		settings:     make(map[string]*Settings),
	}
}

// ============ USERS ============

func (s *MemoryStore) CreateUser(email, hashedPassword, name string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return nil, ErrInvalidData
	}
	if _, exists := s.usersByEmail[normalized]; exists {
		return nil, ErrAlreadyExists
	}

	s.userSeq++
	user := &User{
		ID:        fmt.Sprintf("usr_%d", s.userSeq),
		Email:     email,
		Name:      name,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}

	s.users[user.ID] = user
	s.usersByEmail[normalized] = user.ID
	return user, nil
}

func (s *MemoryStore) GetUserByEmail(email string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	normalized := strings.ToLower(strings.TrimSpace(email))
	userID, ok := s.usersByEmail[normalized]
	if !ok {
		return nil, false
	}
	user, exists := s.users[userID]
	return user, exists
}

func (s *MemoryStore) GetUserByID(id string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	return user, ok
}

// DeleteUser remove um usuário e todos os seus dados (LGPD: Direito ao esquecimento)
func (s *MemoryStore) DeleteUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return ErrNotFound
	}

	// Remover referência por email
	normalized := strings.ToLower(strings.TrimSpace(user.Email))
	delete(s.usersByEmail, normalized)

	// Remover todos os dados relacionados (cascata)
	delete(s.items, userID)
	delete(s.guardians, userID)
	delete(s.progress, userID)
	delete(s.settings, userID)

	// Remover o usuário
	delete(s.users, userID)

	return nil
}

// ExportUserData exporta todos os dados do usuário (LGPD: Portabilidade)
func (s *MemoryStore) ExportUserData(userID string) (*UserDataExport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, ErrNotFound
	}

	// Criar cópia sem senha
	userCopy := &User{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	}

	// Coletar itens
	items := make([]*BoxItem, 0)
	for _, item := range s.items[userID] {
		copyItem := *item
		items = append(items, &copyItem)
	}

	// Coletar guardiões
	guardians := make([]*Guardian, 0)
	for _, g := range s.guardians[userID] {
		copyG := *g
		guardians = append(guardians, &copyG)
	}

	// Coletar progresso
	progress := make([]*GuideProgress, 0)
	for _, p := range s.progress[userID] {
		copyP := *p
		progress = append(progress, &copyP)
	}

	// Coletar configurações
	var settings *Settings
	if s, exists := s.settings[userID]; exists {
		copyS := *s
		settings = &copyS
	}

	return &UserDataExport{
		User:       userCopy,
		Items:      items,
		Guardians:  guardians,
		Progress:   progress,
		Settings:   settings,
		ExportedAt: time.Now(),
	}, nil
}

// ============ BOX ITEMS ============

// GetBoxItems retorna os itens de um usuário (alias para compatibilidade)
func (s *MemoryStore) GetBoxItems(userID string) ([]*BoxItem, error) {
	items := s.ListBoxItems(userID)
	return items, nil
}

func (s *MemoryStore) ListBoxItems(userID string) []*BoxItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userItems := s.items[userID]
	result := make([]*BoxItem, 0, len(userItems))
	for _, item := range userItems {
		copyItem := *item
		result = append(result, &copyItem)
	}
	return result
}

func (s *MemoryStore) GetBoxItem(userID, itemID string) (*BoxItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userItems, ok := s.items[userID]
	if !ok {
		return nil, ErrNotFound
	}
	item, exists := userItems[itemID]
	if !exists {
		return nil, ErrNotFound
	}
	copyItem := *item
	return &copyItem, nil
}

func (s *MemoryStore) CreateBoxItem(userID string, item *BoxItem) (*BoxItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.itemSeq++
	now := time.Now()
	item.ID = fmt.Sprintf("itm_%d", s.itemSeq)
	item.UserID = userID
	item.CreatedAt = now
	item.UpdatedAt = now

	if _, ok := s.items[userID]; !ok {
		s.items[userID] = make(map[string]*BoxItem)
	}
	s.items[userID][item.ID] = item
	copyItem := *item
	return &copyItem, nil
}

func (s *MemoryStore) UpdateBoxItem(userID, itemID string, updates *BoxItem) (*BoxItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userItems, ok := s.items[userID]
	if !ok {
		return nil, ErrNotFound
	}
	item, exists := userItems[itemID]
	if !exists {
		return nil, ErrNotFound
	}

	item.Title = updates.Title
	item.Content = updates.Content
	item.Type = updates.Type
	item.Category = updates.Category
	item.Recipient = updates.Recipient
	item.IsImportant = updates.IsImportant
	item.UpdatedAt = time.Now()

	copyItem := *item
	return &copyItem, nil
}

func (s *MemoryStore) DeleteBoxItem(userID, itemID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userItems, ok := s.items[userID]
	if !ok {
		return ErrNotFound
	}
	if _, exists := userItems[itemID]; !exists {
		return ErrNotFound
	}
	delete(userItems, itemID)
	return nil
}

// ListBoxItemsPaginated lista itens com paginação (cursor-based)
func (s *MemoryStore) ListBoxItemsPaginated(userID string, params *PaginationParams) (*PaginatedResult[*BoxItemSummary], error) {
	params = NormalizePagination(params)

	s.mu.RLock()
	defer s.mu.RUnlock()

	userItems := s.items[userID]

	// Converter para slice e ordenar por ID desc
	var allItems []*BoxItem
	for _, item := range userItems {
		copyItem := *item
		allItems = append(allItems, &copyItem)
	}

	// Ordenar por ID desc (simples)
	for i := 0; i < len(allItems); i++ {
		for j := i + 1; j < len(allItems); j++ {
			if allItems[i].ID < allItems[j].ID {
				allItems[i], allItems[j] = allItems[j], allItems[i]
			}
		}
	}

	// Aplicar cursor
	startIdx := 0
	if params.Cursor != "" {
		for i, item := range allItems {
			if item.ID == params.Cursor {
				startIdx = i + 1
				break
			}
		}
	}

	// Paginar
	endIdx := startIdx + params.Limit + 1
	if endIdx > len(allItems) {
		endIdx = len(allItems)
	}

	pagedItems := allItems[startIdx:endIdx]
	hasMore := len(pagedItems) > params.Limit
	if hasMore {
		pagedItems = pagedItems[:params.Limit]
	}

	// Converter para BoxItemSummary
	summaries := make([]*BoxItemSummary, len(pagedItems))
	for i, item := range pagedItems {
		summaries[i] = &BoxItemSummary{
			ID:          item.ID,
			Type:        item.Type,
			Title:       item.Title,
			Category:    item.Category,
			IsImportant: item.IsImportant,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	var nextCursor string
	if hasMore && len(summaries) > 0 {
		nextCursor = summaries[len(summaries)-1].ID
	}

	return &PaginatedResult[*BoxItemSummary]{
		Items:      summaries,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// CountBoxItems conta o total de itens de um usuário
func (s *MemoryStore) CountBoxItems(userID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items[userID]), nil
}

// ============ GUARDIANS ============

// GetGuardians retorna os guardiões de um usuário (alias para compatibilidade)
func (s *MemoryStore) GetGuardians(userID string) ([]*Guardian, error) {
	guardians := s.ListGuardians(userID)
	return guardians, nil
}

func (s *MemoryStore) ListGuardians(userID string) []*Guardian {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userGuardians := s.guardians[userID]
	result := make([]*Guardian, 0, len(userGuardians))
	for _, guardian := range userGuardians {
		copyGuardian := *guardian
		result = append(result, &copyGuardian)
	}
	return result
}

func (s *MemoryStore) CreateGuardian(userID string, guardian *Guardian) (*Guardian, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.guardianSeq++
	now := time.Now()
	guardian.ID = fmt.Sprintf("grd_%d", s.guardianSeq)
	guardian.UserID = userID
	guardian.CreatedAt = now
	guardian.UpdatedAt = now

	if guardian.Role == "" {
		guardian.Role = "viewer"
	}

	if _, ok := s.guardians[userID]; !ok {
		s.guardians[userID] = make(map[string]*Guardian)
	}
	s.guardians[userID][guardian.ID] = guardian
	copyGuardian := *guardian
	return &copyGuardian, nil
}

func (s *MemoryStore) UpdateGuardian(userID, guardianID string, updates *Guardian) (*Guardian, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userGuardians, ok := s.guardians[userID]
	if !ok {
		return nil, ErrNotFound
	}
	guardian, exists := userGuardians[guardianID]
	if !exists {
		return nil, ErrNotFound
	}

	guardian.Name = updates.Name
	guardian.Email = updates.Email
	guardian.Phone = updates.Phone
	guardian.Relationship = updates.Relationship
	guardian.Notes = updates.Notes
	guardian.UpdatedAt = time.Now()

	copyGuardian := *guardian
	return &copyGuardian, nil
}

func (s *MemoryStore) DeleteGuardian(userID, guardianID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userGuardians, ok := s.guardians[userID]
	if !ok {
		return ErrNotFound
	}
	if _, exists := userGuardians[guardianID]; !exists {
		return ErrNotFound
	}
	delete(userGuardians, guardianID)
	return nil
}

// ListGuardiansPaginated lista guardiões com paginação
func (s *MemoryStore) ListGuardiansPaginated(userID string, params *PaginationParams) (*PaginatedResult[*Guardian], error) {
	params = NormalizePagination(params)

	s.mu.RLock()
	defer s.mu.RUnlock()

	userGuardians := s.guardians[userID]

	// Converter para slice
	var allGuardians []*Guardian
	for _, g := range userGuardians {
		copyG := *g
		allGuardians = append(allGuardians, &copyG)
	}

	// Ordenar por ID desc
	for i := 0; i < len(allGuardians); i++ {
		for j := i + 1; j < len(allGuardians); j++ {
			if allGuardians[i].ID < allGuardians[j].ID {
				allGuardians[i], allGuardians[j] = allGuardians[j], allGuardians[i]
			}
		}
	}

	// Aplicar cursor
	startIdx := 0
	if params.Cursor != "" {
		for i, g := range allGuardians {
			if g.ID == params.Cursor {
				startIdx = i + 1
				break
			}
		}
	}

	// Paginar
	endIdx := startIdx + params.Limit + 1
	if endIdx > len(allGuardians) {
		endIdx = len(allGuardians)
	}

	pagedGuardians := allGuardians[startIdx:endIdx]
	hasMore := len(pagedGuardians) > params.Limit
	if hasMore {
		pagedGuardians = pagedGuardians[:params.Limit]
	}

	var nextCursor string
	if hasMore && len(pagedGuardians) > 0 {
		nextCursor = pagedGuardians[len(pagedGuardians)-1].ID
	}

	return &PaginatedResult[*Guardian]{
		Items:      pagedGuardians,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// CountGuardians conta o total de guardiões de um usuário
func (s *MemoryStore) CountGuardians(userID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.guardians[userID]), nil
}

// ============ GUIDE PROGRESS ============

func (s *MemoryStore) GetGuideProgress(userID string) map[string]*GuideProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userProgress := s.progress[userID]
	result := make(map[string]*GuideProgress)
	for cardID, p := range userProgress {
		copyProgress := *p
		result[cardID] = &copyProgress
	}
	return result
}

func (s *MemoryStore) UpdateGuideProgress(userID, cardID, status string) (*GuideProgress, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.progress[userID]; !ok {
		s.progress[userID] = make(map[string]*GuideProgress)
	}

	progress := &GuideProgress{
		UserID: userID,
		CardID: cardID,
		Status: status,
	}
	if status == "completed" {
		progress.CompletedAt = time.Now()
	}

	s.progress[userID][cardID] = progress
	return progress, nil
}

// ============ SETTINGS ============

func (s *MemoryStore) GetSettings(userID string) *Settings {
	s.mu.Lock()
	defer s.mu.Unlock()

	settings, ok := s.settings[userID]
	if !ok {
		settings = &Settings{
			UserID:               userID,
			NotificationsEnabled: true,
			Theme:                "light",
		}
		s.settings[userID] = settings
	}
	copySettings := *settings
	return &copySettings
}

func (s *MemoryStore) UpdateSettings(userID string, updates *Settings) *Settings {
	s.mu.Lock()
	defer s.mu.Unlock()

	updates.UserID = userID
	s.settings[userID] = updates
	copySettings := *updates
	return &copySettings
}

// ============ ADMIN / ESTATÍSTICAS ============

// Stats representa estatísticas do sistema
type Stats struct {
	TotalUsers      int            `json:"total_users"`
	TotalItems      int            `json:"total_items"`
	TotalGuardians  int            `json:"total_guardians"`
	ItemsByType     map[string]int `json:"items_by_type"`
	ItemsByCategory map[string]int `json:"items_by_category"`
	RecentSignups   int            `json:"recent_signups"` // Últimos 7 dias
}

// GetStats retorna estatísticas gerais do sistema
func (s *MemoryStore) GetStats() *Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &Stats{
		ItemsByType:     make(map[string]int),
		ItemsByCategory: make(map[string]int),
	}

	// Contar usuários
	stats.TotalUsers = len(s.users)

	// Contar inscrições recentes (últimos 7 dias)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	for _, user := range s.users {
		if user.CreatedAt.After(sevenDaysAgo) {
			stats.RecentSignups++
		}
	}

	// Contar itens e categorias
	for _, userItems := range s.items {
		for _, item := range userItems {
			stats.TotalItems++
			stats.ItemsByType[string(item.Type)]++
			if item.Category != "" {
				stats.ItemsByCategory[item.Category]++
			}
		}
	}

	// Contar guardiões
	for _, userGuardians := range s.guardians {
		stats.TotalGuardians += len(userGuardians)
	}

	return stats
}

// ListUsers retorna lista de todos os usuários (para admin)
func (s *MemoryStore) ListUsers() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		// Criar cópia sem a senha
		copyUser := &User{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			// Password NÃO incluído
		}
		users = append(users, copyUser)
	}
	return users
}
