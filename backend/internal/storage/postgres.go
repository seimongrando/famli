// =============================================================================
// FAMLI - PostgreSQL Storage
// =============================================================================
// Implementação de persistência usando PostgreSQL com:
// - Criptografia de dados sensíveis (AES-256-GCM)
// - Paginação cursor-based para performance
// - Campos específicos (sem SELECT *)
//
// Variáveis de ambiente:
// - DATABASE_URL: URL de conexão do PostgreSQL
// - ENCRYPTION_KEY: Chave para criptografia (mínimo 32 caracteres)
//
// Exemplo:
//   DATABASE_URL=postgres://user:pass@host:5432/famli?sslmode=require
// =============================================================================

package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"famli/internal/security"

	_ "github.com/lib/pq"
)

// PostgresStore implementa armazenamento com PostgreSQL
// Dados sensíveis são criptografados antes de serem salvos
type PostgresStore struct {
	db        *sql.DB
	encryptor *security.Encryptor
}

// NewPostgresStore cria uma nova conexão com PostgreSQL
// Inicializa também o encryptor para dados sensíveis
func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao PostgreSQL: %w", err)
	}

	// Configurar pool de conexões para performance
	db.SetMaxOpenConns(25)                 // Máximo de conexões abertas
	db.SetMaxIdleConns(5)                  // Conexões ociosas no pool
	db.SetConnMaxLifetime(5 * time.Minute) // Tempo de vida da conexão

	// Testar conexão
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erro ao pingar PostgreSQL: %w", err)
	}

	// Inicializar encryptor para dados sensíveis
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		encryptionKey = os.Getenv("JWT_SECRET") // Fallback para compatibilidade
	}
	if encryptionKey == "" {
		encryptionKey = "famli-dev-encryption-key-32chars!" // Apenas para desenvolvimento
		log.Println("⚠️  AVISO: Usando chave de criptografia padrão (apenas para desenvolvimento)")
	}

	encryptor, err := security.NewEncryptor(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar encryptor: %w", err)
	}

	store := &PostgresStore{
		db:        db,
		encryptor: encryptor,
	}

	// Executar migrações
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("erro nas migrações: %w", err)
	}

	log.Println("✅ PostgreSQL conectado com sucesso")
	log.Println("🔐 Criptografia de dados habilitada")
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

		// =======================================================================
		// ÍNDICES PARA PERFORMANCE
		// =======================================================================
		// Índices de usuários
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(LOWER(email))`,
		`CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC)`,

		// Índices de box_items (performance em listagens e filtros)
		`CREATE INDEX IF NOT EXISTS idx_box_items_user ON box_items(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_box_items_user_type ON box_items(user_id, type)`,
		`CREATE INDEX IF NOT EXISTS idx_box_items_user_created ON box_items(user_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_box_items_important ON box_items(user_id, is_important) WHERE is_important = TRUE`,

		// Índices de guardians
		`CREATE INDEX IF NOT EXISTS idx_guardians_user ON guardians(user_id)`,

		// Índices de guide_progress (para verificar progresso rapidamente)
		`CREATE INDEX IF NOT EXISTS idx_guide_progress_user ON guide_progress(user_id)`,

		// =======================================================================
		// AUDITORIA E SEGURANÇA
		// =======================================================================
		// Tabela de auditoria para rastrear ações sensíveis (LGPD)
		`CREATE TABLE IF NOT EXISTS audit_log (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(50),
			action VARCHAR(100) NOT NULL,
			resource_type VARCHAR(50),
			resource_id VARCHAR(50),
			ip_address VARCHAR(45),
			user_agent TEXT,
			details JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_log(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_log(action)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_log(created_at DESC)`,

		// Tabela para tokens de exclusão (confirmação de exclusão de conta)
		`CREATE TABLE IF NOT EXISTS deletion_tokens (
			id VARCHAR(100) PRIMARY KEY,
			user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_deletion_tokens_user ON deletion_tokens(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_deletion_tokens_expires ON deletion_tokens(expires_at)`,
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
	var name sql.NullString

	err := s.db.QueryRow(`
		SELECT id, email, name, password, created_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &name, &user.Password, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		log.Printf("[PostgreSQL] Erro ao buscar usuário por ID: %v", err)
		return nil, false
	}

	user.Name = name.String
	return &user, true
}

// DeleteUser remove um usuário e todos os seus dados (LGPD: Direito ao esquecimento)
// Devido ao ON DELETE CASCADE, todos os dados relacionados são removidos automaticamente
func (s *PostgresStore) DeleteUser(userID string) error {
	result, err := s.db.Exec(`DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("erro ao deletar usuário: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	log.Printf("[PostgreSQL] Usuário %s e todos os dados relacionados foram deletados (LGPD)", userID)
	return nil
}

// ExportUserData exporta todos os dados do usuário (LGPD: Portabilidade)
func (s *PostgresStore) ExportUserData(userID string) (*UserDataExport, error) {
	user, found := s.GetUserByID(userID)
	if !found {
		return nil, ErrNotFound
	}

	// Limpar senha do export
	user.Password = ""

	items := s.ListBoxItems(userID)
	guardians := s.ListGuardians(userID)
	progressMap := s.GetGuideProgress(userID)
	settings := s.GetSettings(userID)

	// Converter map de progresso para slice
	progress := make([]*GuideProgress, 0, len(progressMap))
	for _, p := range progressMap {
		progress = append(progress, p)
	}

	return &UserDataExport{
		User:       user,
		Items:      items,
		Guardians:  guardians,
		Progress:   progress,
		Settings:   settings,
		ExportedAt: time.Now(),
	}, nil
}

// ============================================================================
// HELPERS DE CRIPTOGRAFIA
// ============================================================================

// encryptSensitive criptografa um valor se não estiver vazio
func (s *PostgresStore) encryptSensitive(value string) string {
	if value == "" || s.encryptor == nil {
		return value
	}
	encrypted, err := s.encryptor.Encrypt(value)
	if err != nil {
		log.Printf("[PostgreSQL] Aviso: falha ao criptografar valor: %v", err)
		return value // Retorna valor original se falhar
	}
	return encrypted
}

// decryptSensitive descriptografa um valor se estiver criptografado
func (s *PostgresStore) decryptSensitive(value string) string {
	if value == "" || s.encryptor == nil {
		return value
	}
	// Tenta descriptografar - se falhar, assume que é texto plano (migração)
	decrypted, err := s.encryptor.Decrypt(value)
	if err != nil {
		// Pode ser dado antigo não criptografado
		return value
	}
	return decrypted
}

// ============================================================================
// BOX ITEMS
// ============================================================================

// GetBoxItems retorna itens (compatibilidade - sem paginação)
func (s *PostgresStore) GetBoxItems(userID string) ([]*BoxItem, error) {
	return s.ListBoxItems(userID), nil
}

// ListBoxItems lista todos os itens (CUIDADO: sem paginação, usar apenas para exportação)
func (s *PostgresStore) ListBoxItems(userID string) []*BoxItem {
	// Query com campos específicos (não usa SELECT *)
	rows, err := s.db.Query(`
		SELECT id, user_id, type, title, content, category, recipient, is_important, created_at, updated_at
		FROM box_items 
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT 1000
	`, userID)
	if err != nil {
		log.Printf("[PostgreSQL] Erro ao listar itens: %v", err)
		return []*BoxItem{}
	}
	defer rows.Close()

	var items []*BoxItem
	for rows.Next() {
		var item BoxItem
		var title, content, category, recipient sql.NullString
		err := rows.Scan(
			&item.ID, &item.UserID, &item.Type, &title,
			&content, &category, &recipient,
			&item.IsImportant, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			log.Printf("[PostgreSQL] Erro ao escanear item: %v", err)
			continue
		}
		// Descriptografar dados sensíveis
		item.Title = s.decryptSensitive(title.String)
		item.Content = s.decryptSensitive(content.String)
		item.Category = category.String
		item.Recipient = s.decryptSensitive(recipient.String)
		items = append(items, &item)
	}

	return items
}

// ListBoxItemsPaginated lista itens com paginação (método preferido)
// Usa cursor-based pagination para melhor performance
func (s *PostgresStore) ListBoxItemsPaginated(userID string, params *PaginationParams) (*PaginatedResult[*BoxItemSummary], error) {
	params = NormalizePagination(params)

	// Query paginada - busca limit+1 para detectar hasMore
	var rows *sql.Rows
	var err error

	if params.Cursor != "" {
		// Buscar itens após o cursor (baseado no ID)
		rows, err = s.db.Query(`
			SELECT id, type, title, category, is_important, updated_at
			FROM box_items 
			WHERE user_id = $1 AND id < $2
			ORDER BY id DESC
			LIMIT $3
		`, userID, params.Cursor, params.Limit+1)
	} else {
		// Primeira página
		rows, err = s.db.Query(`
			SELECT id, type, title, category, is_important, updated_at
			FROM box_items 
			WHERE user_id = $1
			ORDER BY id DESC
			LIMIT $2
		`, userID, params.Limit+1)
	}

	if err != nil {
		return nil, fmt.Errorf("erro ao listar itens paginados: %w", err)
	}
	defer rows.Close()

	var items []*BoxItemSummary
	for rows.Next() {
		var item BoxItemSummary
		var title, category sql.NullString
		err := rows.Scan(
			&item.ID, &item.Type, &title, &category,
			&item.IsImportant, &item.UpdatedAt,
		)
		if err != nil {
			log.Printf("[PostgreSQL] Erro ao escanear item: %v", err)
			continue
		}
		item.Title = s.decryptSensitive(title.String)
		item.Category = category.String
		items = append(items, &item)
	}

	// Verificar se há mais páginas
	hasMore := len(items) > params.Limit
	if hasMore {
		items = items[:params.Limit] // Remove o item extra
	}

	// Próximo cursor
	var nextCursor string
	if hasMore && len(items) > 0 {
		nextCursor = items[len(items)-1].ID
	}

	return &PaginatedResult[*BoxItemSummary]{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// CountBoxItems conta o total de itens de um usuário
func (s *PostgresStore) CountBoxItems(userID string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM box_items WHERE user_id = $1
	`, userID).Scan(&count)
	return count, err
}

// GetBoxItem busca um item específico por ID
func (s *PostgresStore) GetBoxItem(userID, itemID string) (*BoxItem, error) {
	var item BoxItem
	var title, content, category, recipient sql.NullString

	// Query com campos específicos (não usa SELECT *)
	err := s.db.QueryRow(`
		SELECT id, user_id, type, title, content, category, recipient, is_important, created_at, updated_at
		FROM box_items 
		WHERE user_id = $1 AND id = $2
	`, userID, itemID).Scan(
		&item.ID, &item.UserID, &item.Type, &title,
		&content, &category, &recipient,
		&item.IsImportant, &item.CreatedAt, &item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Descriptografar dados sensíveis
	item.Title = s.decryptSensitive(title.String)
	item.Content = s.decryptSensitive(content.String)
	item.Category = category.String
	item.Recipient = s.decryptSensitive(recipient.String)
	return &item, nil
}

// CreateBoxItem cria um novo item com dados criptografados
func (s *PostgresStore) CreateBoxItem(userID string, item *BoxItem) (*BoxItem, error) {
	id := fmt.Sprintf("itm_%d", time.Now().UnixNano())
	now := time.Now()

	// Criptografar dados sensíveis antes de salvar
	encTitle := s.encryptSensitive(item.Title)
	encContent := s.encryptSensitive(item.Content)
	encRecipient := s.encryptSensitive(item.Recipient)

	_, err := s.db.Exec(`
		INSERT INTO box_items (id, user_id, type, title, content, category, recipient, is_important, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, id, userID, item.Type, encTitle, encContent, item.Category, encRecipient, item.IsImportant, now, now)

	if err != nil {
		return nil, err
	}

	item.ID = id
	item.UserID = userID
	item.CreatedAt = now
	item.UpdatedAt = now
	return item, nil
}

// UpdateBoxItem atualiza um item existente com dados criptografados
func (s *PostgresStore) UpdateBoxItem(userID, itemID string, updates *BoxItem) (*BoxItem, error) {
	// Criptografar dados sensíveis antes de atualizar
	encTitle := s.encryptSensitive(updates.Title)
	encContent := s.encryptSensitive(updates.Content)
	encRecipient := s.encryptSensitive(updates.Recipient)

	result, err := s.db.Exec(`
		UPDATE box_items 
		SET title = $1, content = $2, type = $3, category = $4, recipient = $5, is_important = $6, updated_at = $7
		WHERE user_id = $8 AND id = $9
	`, encTitle, encContent, updates.Type, updates.Category, encRecipient, updates.IsImportant, time.Now(), userID, itemID)

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

// GetGuardians retorna guardiões (compatibilidade)
func (s *PostgresStore) GetGuardians(userID string) ([]*Guardian, error) {
	return s.ListGuardians(userID), nil
}

// ListGuardians lista todos os guardiões (CUIDADO: sem paginação)
func (s *PostgresStore) ListGuardians(userID string) []*Guardian {
	rows, err := s.db.Query(`
		SELECT id, user_id, name, email, phone, relationship, role, notes, created_at, updated_at
		FROM guardians 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`, userID)
	if err != nil {
		log.Printf("[PostgreSQL] Erro ao listar guardiões: %v", err)
		return []*Guardian{}
	}
	defer rows.Close()

	var guardians []*Guardian
	for rows.Next() {
		var g Guardian
		var name, email, phone, relationship, notes sql.NullString
		err := rows.Scan(
			&g.ID, &g.UserID, &name, &email, &phone,
			&relationship, &g.Role, &notes,
			&g.CreatedAt, &g.UpdatedAt,
		)
		if err != nil {
			log.Printf("[PostgreSQL] Erro ao escanear guardião: %v", err)
			continue
		}
		// Descriptografar dados sensíveis (PII)
		g.Name = s.decryptSensitive(name.String)
		g.Email = s.decryptSensitive(email.String)
		g.Phone = s.decryptSensitive(phone.String)
		g.Relationship = relationship.String
		g.Notes = s.decryptSensitive(notes.String)
		guardians = append(guardians, &g)
	}

	return guardians
}

// ListGuardiansPaginated lista guardiões com paginação
func (s *PostgresStore) ListGuardiansPaginated(userID string, params *PaginationParams) (*PaginatedResult[*Guardian], error) {
	params = NormalizePagination(params)

	var rows *sql.Rows
	var err error

	if params.Cursor != "" {
		rows, err = s.db.Query(`
			SELECT id, user_id, name, email, phone, relationship, role, notes, created_at, updated_at
			FROM guardians 
			WHERE user_id = $1 AND id < $2
			ORDER BY id DESC
			LIMIT $3
		`, userID, params.Cursor, params.Limit+1)
	} else {
		rows, err = s.db.Query(`
			SELECT id, user_id, name, email, phone, relationship, role, notes, created_at, updated_at
			FROM guardians 
			WHERE user_id = $1
			ORDER BY id DESC
			LIMIT $2
		`, userID, params.Limit+1)
	}

	if err != nil {
		return nil, fmt.Errorf("erro ao listar guardiões paginados: %w", err)
	}
	defer rows.Close()

	var guardians []*Guardian
	for rows.Next() {
		var g Guardian
		var name, email, phone, relationship, notes sql.NullString
		err := rows.Scan(
			&g.ID, &g.UserID, &name, &email, &phone,
			&relationship, &g.Role, &notes,
			&g.CreatedAt, &g.UpdatedAt,
		)
		if err != nil {
			continue
		}
		g.Name = s.decryptSensitive(name.String)
		g.Email = s.decryptSensitive(email.String)
		g.Phone = s.decryptSensitive(phone.String)
		g.Relationship = relationship.String
		g.Notes = s.decryptSensitive(notes.String)
		guardians = append(guardians, &g)
	}

	hasMore := len(guardians) > params.Limit
	if hasMore {
		guardians = guardians[:params.Limit]
	}

	var nextCursor string
	if hasMore && len(guardians) > 0 {
		nextCursor = guardians[len(guardians)-1].ID
	}

	return &PaginatedResult[*Guardian]{
		Items:      guardians,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// CountGuardians conta o total de guardiões de um usuário
func (s *PostgresStore) CountGuardians(userID string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM guardians WHERE user_id = $1
	`, userID).Scan(&count)
	return count, err
}

// CreateGuardian cria um novo guardião com dados criptografados
func (s *PostgresStore) CreateGuardian(userID string, guardian *Guardian) (*Guardian, error) {
	id := fmt.Sprintf("grd_%d", time.Now().UnixNano())
	now := time.Now()
	role := guardian.Role
	if role == "" {
		role = "viewer"
	}

	// Criptografar dados sensíveis (PII)
	encName := s.encryptSensitive(guardian.Name)
	encEmail := s.encryptSensitive(guardian.Email)
	encPhone := s.encryptSensitive(guardian.Phone)
	encNotes := s.encryptSensitive(guardian.Notes)

	_, err := s.db.Exec(`
		INSERT INTO guardians (id, user_id, name, email, phone, relationship, role, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, id, userID, encName, encEmail, encPhone, guardian.Relationship, role, encNotes, now, now)

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

// UpdateGuardian atualiza um guardião com dados criptografados
func (s *PostgresStore) UpdateGuardian(userID, guardianID string, updates *Guardian) (*Guardian, error) {
	// Criptografar dados sensíveis (PII)
	encName := s.encryptSensitive(updates.Name)
	encEmail := s.encryptSensitive(updates.Email)
	encPhone := s.encryptSensitive(updates.Phone)
	encNotes := s.encryptSensitive(updates.Notes)

	result, err := s.db.Exec(`
		UPDATE guardians 
		SET name = $1, email = $2, phone = $3, relationship = $4, notes = $5, updated_at = $6
		WHERE user_id = $7 AND id = $8
	`, encName, encEmail, encPhone, updates.Relationship, encNotes, time.Now(), userID, guardianID)

	if err != nil {
		return nil, err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, ErrNotFound
	}

	// Buscar guardião atualizado (descriptografado pelo GetGuardian interno)
	var g Guardian
	var name, email, phone, relationship, notes sql.NullString
	err = s.db.QueryRow(`
		SELECT id, user_id, name, email, phone, relationship, role, notes, created_at, updated_at
		FROM guardians WHERE user_id = $1 AND id = $2
	`, userID, guardianID).Scan(
		&g.ID, &g.UserID, &name, &email, &phone,
		&relationship, &g.Role, &notes,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	// Descriptografar dados sensíveis
	g.Name = s.decryptSensitive(name.String)
	g.Email = s.decryptSensitive(email.String)
	g.Phone = s.decryptSensitive(phone.String)
	g.Relationship = relationship.String
	g.Notes = s.decryptSensitive(notes.String)
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
