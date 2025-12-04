package store

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"legacybridge/models"
)

// MemoryStore provides a simple in-memory persistence layer for the MVP.
type MemoryStore struct {
	mu           sync.RWMutex
	users        map[string]*models.User
	usersByEmail map[string]string
	items        map[string]map[string]*models.LegacyItem
	guardians    map[string]map[string]*models.Guardian
	settings     map[string]*models.Settings
	userSeq      int64
	itemSeq      int64
	guardianSeq  int64
}

// NewMemoryStore builds a store with deterministic maps.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:        make(map[string]*models.User),
		usersByEmail: make(map[string]string),
		items:        make(map[string]map[string]*models.LegacyItem),
		guardians:    make(map[string]map[string]*models.Guardian),
		settings:     make(map[string]*models.Settings),
	}
}

// CreateUser registers a user if the email is not already used.
func (s *MemoryStore) CreateUser(email, hashedPassword string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return nil, errors.New("email cannot be empty")
	}
	if _, exists := s.usersByEmail[normalized]; exists {
		return nil, errors.New("email already registered")
	}

	s.userSeq++
	user := &models.User{
		ID:        fmt.Sprintf("usr_%d", s.userSeq),
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}

	s.users[user.ID] = user
	s.usersByEmail[normalized] = user.ID
	return user, nil
}

// GetUserByEmail fetches a user by normalized email.
func (s *MemoryStore) GetUserByEmail(email string) (*models.User, bool) {
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

// GetUserByID returns the user by identifier.
func (s *MemoryStore) GetUserByID(id string) (*models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	return user, ok
}

// ListLegacyItems returns the items for the user.
func (s *MemoryStore) ListLegacyItems(userID string) []*models.LegacyItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userItems := s.items[userID]
	result := make([]*models.LegacyItem, 0, len(userItems))
	for _, item := range userItems {
		// create copies to avoid external mutation
		copyItem := *item
		result = append(result, &copyItem)
	}
	return result
}

// AddLegacyItem stores a new record owned by the user.
func (s *MemoryStore) AddLegacyItem(userID, title, itemType, content string) *models.LegacyItem {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.itemSeq++
	item := &models.LegacyItem{
		ID:        fmt.Sprintf("itm_%d", s.itemSeq),
		UserID:    userID,
		Title:     title,
		Type:      itemType,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if _, ok := s.items[userID]; !ok {
		s.items[userID] = make(map[string]*models.LegacyItem)
	}
	s.items[userID][item.ID] = item
	return item
}

// UpdateLegacyItem modifies an existing record if owned by the user.
func (s *MemoryStore) UpdateLegacyItem(userID, itemID, title, itemType, content string) (*models.LegacyItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userItems, ok := s.items[userID]
	if !ok {
		return nil, errors.New("item not found")
	}
	item, exists := userItems[itemID]
	if !exists {
		return nil, errors.New("item not found")
	}

	item.Title = title
	item.Type = itemType
	item.Content = content
	return item, nil
}

// DeleteLegacyItem removes the record if it belongs to the user.
func (s *MemoryStore) DeleteLegacyItem(userID, itemID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userItems, ok := s.items[userID]
	if !ok {
		return errors.New("item not found")
	}
	if _, exists := userItems[itemID]; !exists {
		return errors.New("item not found")
	}
	delete(userItems, itemID)
	return nil
}

// ListGuardians returns trusted contacts for a user.
func (s *MemoryStore) ListGuardians(userID string) []*models.Guardian {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userGuardians := s.guardians[userID]
	result := make([]*models.Guardian, 0, len(userGuardians))
	for _, guardian := range userGuardians {
		copyGuardian := *guardian
		result = append(result, &copyGuardian)
	}
	return result
}

// AddGuardian adds a new trusted person.
func (s *MemoryStore) AddGuardian(userID, name, email string) *models.Guardian {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.guardianSeq++
	guardian := &models.Guardian{
		ID:     fmt.Sprintf("grd_%d", s.guardianSeq),
		UserID: userID,
		Name:   name,
		Email:  email,
	}
	if _, ok := s.guardians[userID]; !ok {
		s.guardians[userID] = make(map[string]*models.Guardian)
	}
	s.guardians[userID][guardian.ID] = guardian
	return guardian
}

// DeleteGuardian removes the guardian if owned by the user.
func (s *MemoryStore) DeleteGuardian(userID, guardianID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userGuardians, ok := s.guardians[userID]
	if !ok {
		return errors.New("guardian not found")
	}
	if _, exists := userGuardians[guardianID]; !exists {
		return errors.New("guardian not found")
	}
	delete(userGuardians, guardianID)
	return nil
}

// GetSettings returns the user's settings, defaulting to opt-out.
func (s *MemoryStore) GetSettings(userID string) *models.Settings {
	s.mu.Lock()
	defer s.mu.Unlock()

	settings, ok := s.settings[userID]
	if !ok {
		settings = &models.Settings{UserID: userID}
		s.settings[userID] = settings
	}
	copySettings := *settings
	return &copySettings
}

// UpdateSettings replaces the stored settings for the user.
func (s *MemoryStore) UpdateSettings(userID string, incoming *models.Settings) *models.Settings {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings[userID] = &models.Settings{
		UserID:                   userID,
		EmergencyProtocolEnabled: incoming.EmergencyProtocolEnabled,
	}
	copySettings := *s.settings[userID]
	return &copySettings
}
