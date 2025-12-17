// =============================================================================
// FAMLI - PostgreSQL Storage
// =============================================================================
// Implementação de persistência usando PostgreSQL.
//
// Variáveis de ambiente:
// - DATABASE_URL: URL de conexão do PostgreSQL
//
// Exemplo:
//   DATABASE_URL=postgres://user:pass@host:5432/famli?sslmode=require
// =============================================================================

package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// PostgresStore implementa armazenamento com PostgreSQL
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore cria uma nova conexão com PostgreSQL
func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao PostgreSQL: %w", err)
	}

	// Configurar pool de conexões
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Testar conexão
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erro ao pingar PostgreSQL: %w", err)
	}

	store := &PostgresStore{db: db}

	// Executar migrações
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("erro nas migrações: %w", err)
	}

	log.Println("✅ PostgreSQL conectado com sucesso")
	return store, nil
}

// migrate executa as migrações do banco
func (s *PostgresStore) migrate() error {
	migrations := []string{
		// Extensão UUID
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,

		// Tabela users
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(50) PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255),
			password VARCHAR(255) NOT NULL,
			terms_accepted BOOLEAN DEFAULT FALSE,
			terms_accepted_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Tabela box_items
		`CREATE TABLE IF NOT EXISTS box_items (
			id VARCHAR(50) PRIMARY KEY,
			user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL DEFAULT 'info',
			title VARCHAR(500) NOT NULL,
			content TEXT,
			category VARCHAR(100),
			recipient VARCHAR(255),
			is_important BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Tabela guardians
		`CREATE TABLE IF NOT EXISTS guardians (
			id VARCHAR(50) PRIMARY KEY,
			user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255),
			phone VARCHAR(50),
			relationship VARCHAR(100),
			role VARCHAR(50) DEFAULT 'viewer',
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Tabela guide_progress
		`CREATE TABLE IF NOT EXISTS guide_progress (
			user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			card_id VARCHAR(50) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			completed_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, card_id)
		)`,

		// Tabela settings
		`CREATE TABLE IF NOT EXISTS settings (
			user_id VARCHAR(50) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			emergency_protocol_enabled BOOLEAN DEFAULT FALSE,
			notifications_enabled BOOLEAN DEFAULT TRUE,
			theme VARCHAR(20) DEFAULT 'light',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Índices
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(LOWER(email))`,
		`CREATE INDEX IF NOT EXISTS idx_box_items_user ON box_items(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_guardians_user ON guardians(user_id)`,
	}

	for _, migration := range migrations {
		if _, err := s.db.Exec(migration); err != nil {
			return fmt.Errorf("erro na migração: %w\nSQL: %s", err, migration)
		}
	}

	return nil
}

// Close fecha a conexão com o banco
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// ============================================================================
// USERS
// ============================================================================

func (s *PostgresStore) CreateUser(email, hashedPassword, name string) (*User, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return nil, ErrInvalidData
	}

	id := fmt.Sprintf("usr_%d", time.Now().UnixNano())
	now := time.Now()

	_, err := s.db.Exec(`
		INSERT INTO users (id, email, name, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, email, name, hashedPassword, now, now)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}

	return &User{
		ID:        id,
		Email:     email,
		Name:      name,
		Password:  hashedPassword,
		CreatedAt: now,
	}, nil
}

func (s *PostgresStore) GetUserByEmail(email string) (*User, bool) {
	normalized := strings.ToLower(strings.TrimSpace(email))

	var user User
	err := s.db.QueryRow(`
		SELECT id, email, name, password, created_at
		FROM users WHERE LOWER(email) = $1
	`, normalized).Scan(&user.ID, &user.Email, &user.Name, &user.Password, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		log.Printf("[PostgreSQL] Erro ao buscar usuário por email: %v", err)
		return nil, false
	}

	return &user, true
}

func (s *PostgresStore) GetUserByID(id string) (*User, bool) {
	var user User
	err := s.db.QueryRow(`
		SELECT id, email, name, password, created_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.Name, &user.Password, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		log.Printf("[PostgreSQL] Erro ao buscar usuário por ID: %v", err)
		return nil, false
	}

	return &user, true
}

// ============================================================================
// BOX ITEMS
// ============================================================================

func (s *PostgresStore) GetBoxItems(userID string) ([]*BoxItem, error) {
	return s.ListBoxItems(userID), nil
}

func (s *PostgresStore) ListBoxItems(userID string) []*BoxItem {
	rows, err := s.db.Query(`
		SELECT id, user_id, type, title, content, category, recipient, is_important, created_at, updated_at
		FROM box_items WHERE user_id = $1
		ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		log.Printf("[PostgreSQL] Erro ao listar itens: %v", err)
		return []*BoxItem{}
	}
	defer rows.Close()

	var items []*BoxItem
	for rows.Next() {
		var item BoxItem
		var content, category, recipient sql.NullString
		err := rows.Scan(
			&item.ID, &item.UserID, &item.Type, &item.Title,
			&content, &category, &recipient,
			&item.IsImportant, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			log.Printf("[PostgreSQL] Erro ao escanear item: %v", err)
			continue
		}
		item.Content = content.String
		item.Category = category.String
		item.Recipient = recipient.String
		items = append(items, &item)
	}

	return items
}

func (s *PostgresStore) GetBoxItem(userID, itemID string) (*BoxItem, error) {
	var item BoxItem
	var content, category, recipient sql.NullString

	err := s.db.QueryRow(`
		SELECT id, user_id, type, title, content, category, recipient, is_important, created_at, updated_at
		FROM box_items WHERE user_id = $1 AND id = $2
	`, userID, itemID).Scan(
		&item.ID, &item.UserID, &item.Type, &item.Title,
		&content, &category, &recipient,
		&item.IsImportant, &item.CreatedAt, &item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	item.Content = content.String
	item.Category = category.String
	item.Recipient = recipient.String
	return &item, nil
}

func (s *PostgresStore) CreateBoxItem(userID string, item *BoxItem) (*BoxItem, error) {
	id := fmt.Sprintf("itm_%d", time.Now().UnixNano())
	now := time.Now()

	_, err := s.db.Exec(`
		INSERT INTO box_items (id, user_id, type, title, content, category, recipient, is_important, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, id, userID, item.Type, item.Title, item.Content, item.Category, item.Recipient, item.IsImportant, now, now)

	if err != nil {
		return nil, err
	}

	item.ID = id
	item.UserID = userID
	item.CreatedAt = now
	item.UpdatedAt = now
	return item, nil
}

func (s *PostgresStore) UpdateBoxItem(userID, itemID string, updates *BoxItem) (*BoxItem, error) {
	result, err := s.db.Exec(`
		UPDATE box_items 
		SET title = $1, content = $2, type = $3, category = $4, recipient = $5, is_important = $6, updated_at = $7
		WHERE user_id = $8 AND id = $9
	`, updates.Title, updates.Content, updates.Type, updates.Category, updates.Recipient, updates.IsImportant, time.Now(), userID, itemID)

	if err != nil {
		return nil, err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, ErrNotFound
	}

	return s.GetBoxItem(userID, itemID)
}

func (s *PostgresStore) DeleteBoxItem(userID, itemID string) error {
	result, err := s.db.Exec(`
		DELETE FROM box_items WHERE user_id = $1 AND id = $2
	`, userID, itemID)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// ============================================================================
// GUARDIANS
// ============================================================================

func (s *PostgresStore) GetGuardians(userID string) ([]*Guardian, error) {
	return s.ListGuardians(userID), nil
}

func (s *PostgresStore) ListGuardians(userID string) []*Guardian {
	rows, err := s.db.Query(`
		SELECT id, user_id, name, email, phone, relationship, role, notes, created_at, updated_at
		FROM guardians WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		log.Printf("[PostgreSQL] Erro ao listar guardiões: %v", err)
		return []*Guardian{}
	}
	defer rows.Close()

	var guardians []*Guardian
	for rows.Next() {
		var g Guardian
		var email, phone, relationship, notes sql.NullString
		err := rows.Scan(
			&g.ID, &g.UserID, &g.Name, &email, &phone,
			&relationship, &g.Role, &notes,
			&g.CreatedAt, &g.UpdatedAt,
		)
		if err != nil {
			log.Printf("[PostgreSQL] Erro ao escanear guardião: %v", err)
			continue
		}
		g.Email = email.String
		g.Phone = phone.String
		g.Relationship = relationship.String
		g.Notes = notes.String
		guardians = append(guardians, &g)
	}

	return guardians
}

func (s *PostgresStore) CreateGuardian(userID string, guardian *Guardian) (*Guardian, error) {
	id := fmt.Sprintf("grd_%d", time.Now().UnixNano())
	now := time.Now()
	role := guardian.Role
	if role == "" {
		role = "viewer"
	}

	_, err := s.db.Exec(`
		INSERT INTO guardians (id, user_id, name, email, phone, relationship, role, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, id, userID, guardian.Name, guardian.Email, guardian.Phone, guardian.Relationship, role, guardian.Notes, now, now)

	if err != nil {
		return nil, err
	}

	guardian.ID = id
	guardian.UserID = userID
	guardian.Role = role
	guardian.CreatedAt = now
	guardian.UpdatedAt = now
	return guardian, nil
}

func (s *PostgresStore) UpdateGuardian(userID, guardianID string, updates *Guardian) (*Guardian, error) {
	result, err := s.db.Exec(`
		UPDATE guardians 
		SET name = $1, email = $2, phone = $3, relationship = $4, notes = $5, updated_at = $6
		WHERE user_id = $7 AND id = $8
	`, updates.Name, updates.Email, updates.Phone, updates.Relationship, updates.Notes, time.Now(), userID, guardianID)

	if err != nil {
		return nil, err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, ErrNotFound
	}

	// Buscar guardião atualizado
	var g Guardian
	var email, phone, relationship, notes sql.NullString
	err = s.db.QueryRow(`
		SELECT id, user_id, name, email, phone, relationship, role, notes, created_at, updated_at
		FROM guardians WHERE user_id = $1 AND id = $2
	`, userID, guardianID).Scan(
		&g.ID, &g.UserID, &g.Name, &email, &phone,
		&relationship, &g.Role, &notes,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	g.Email = email.String
	g.Phone = phone.String
	g.Relationship = relationship.String
	g.Notes = notes.String
	return &g, nil
}

func (s *PostgresStore) DeleteGuardian(userID, guardianID string) error {
	result, err := s.db.Exec(`
		DELETE FROM guardians WHERE user_id = $1 AND id = $2
	`, userID, guardianID)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// ============================================================================
// GUIDE PROGRESS
// ============================================================================

func (s *PostgresStore) GetGuideProgress(userID string) map[string]*GuideProgress {
	rows, err := s.db.Query(`
		SELECT user_id, card_id, status, completed_at
		FROM guide_progress WHERE user_id = $1
	`, userID)
	if err != nil {
		log.Printf("[PostgreSQL] Erro ao buscar progresso: %v", err)
		return map[string]*GuideProgress{}
	}
	defer rows.Close()

	progress := make(map[string]*GuideProgress)
	for rows.Next() {
		var p GuideProgress
		var completedAt sql.NullTime
		err := rows.Scan(&p.UserID, &p.CardID, &p.Status, &completedAt)
		if err != nil {
			continue
		}
		if completedAt.Valid {
			p.CompletedAt = completedAt.Time
		}
		progress[p.CardID] = &p
	}

	return progress
}

func (s *PostgresStore) UpdateGuideProgress(userID, cardID, status string) (*GuideProgress, error) {
	now := time.Now()
	var completedAt *time.Time
	if status == "completed" {
		completedAt = &now
	}

	// Upsert
	_, err := s.db.Exec(`
		INSERT INTO guide_progress (user_id, card_id, status, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, card_id) 
		DO UPDATE SET status = $3, completed_at = $4, updated_at = $6
	`, userID, cardID, status, completedAt, now, now)

	if err != nil {
		return nil, err
	}

	progress := &GuideProgress{
		UserID: userID,
		CardID: cardID,
		Status: status,
	}
	if completedAt != nil {
		progress.CompletedAt = *completedAt
	}

	return progress, nil
}

// ============================================================================
// SETTINGS
// ============================================================================

func (s *PostgresStore) GetSettings(userID string) *Settings {
	var settings Settings
	err := s.db.QueryRow(`
		SELECT user_id, emergency_protocol_enabled, notifications_enabled, theme
		FROM settings WHERE user_id = $1
	`, userID).Scan(&settings.UserID, &settings.EmergencyProtocolEnabled, &settings.NotificationsEnabled, &settings.Theme)

	if err == sql.ErrNoRows {
		// Criar configurações padrão
		settings = Settings{
			UserID:               userID,
			NotificationsEnabled: true,
			Theme:                "light",
		}
		s.db.Exec(`
			INSERT INTO settings (user_id, notifications_enabled, theme)
			VALUES ($1, $2, $3)
		`, userID, true, "light")
	}

	return &settings
}

func (s *PostgresStore) UpdateSettings(userID string, updates *Settings) *Settings {
	s.db.Exec(`
		INSERT INTO settings (user_id, emergency_protocol_enabled, notifications_enabled, theme)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) 
		DO UPDATE SET emergency_protocol_enabled = $2, notifications_enabled = $3, theme = $4
	`, userID, updates.EmergencyProtocolEnabled, updates.NotificationsEnabled, updates.Theme)

	updates.UserID = userID
	return updates
}

// ============================================================================
// ADMIN / ESTATÍSTICAS
// ============================================================================

func (s *PostgresStore) GetStats() *Stats {
	stats := &Stats{
		ItemsByType:     make(map[string]int),
		ItemsByCategory: make(map[string]int),
	}

	// Total de usuários
	s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)

	// Usuários recentes (7 dias)
	s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE created_at > NOW() - INTERVAL '7 days'`).Scan(&stats.RecentSignups)

	// Total de itens
	s.db.QueryRow(`SELECT COUNT(*) FROM box_items`).Scan(&stats.TotalItems)

	// Total de guardiões
	s.db.QueryRow(`SELECT COUNT(*) FROM guardians`).Scan(&stats.TotalGuardians)

	// Itens por tipo
	rows, _ := s.db.Query(`SELECT type, COUNT(*) FROM box_items GROUP BY type`)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var itemType string
			var count int
			rows.Scan(&itemType, &count)
			stats.ItemsByType[itemType] = count
		}
	}

	// Itens por categoria
	rows2, _ := s.db.Query(`SELECT category, COUNT(*) FROM box_items WHERE category != '' GROUP BY category`)
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var category string
			var count int
			rows2.Scan(&category, &count)
			stats.ItemsByCategory[category] = count
		}
	}

	return stats
}

func (s *PostgresStore) ListUsers() []*User {
	rows, err := s.db.Query(`
		SELECT id, email, name, created_at FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return []*User{}
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		var name sql.NullString
		rows.Scan(&user.ID, &user.Email, &name, &user.CreatedAt)
		user.Name = name.String
		users = append(users, &user)
	}

	return users
}
