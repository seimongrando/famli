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
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"famli/internal/security"

	"github.com/lib/pq"
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

	store := &PostgresStore{
		db: db,
	}

	// Executar migrações
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("erro nas migrações: %w", err)
	}

	// Inicializar encryptor para dados sensíveis
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		encryptionKey = os.Getenv("JWT_SECRET") // Fallback para compatibilidade
	}
	if encryptionKey == "" {
		encryptionKey = "famli-dev-encryption-key-32chars!" // Apenas para desenvolvimento
		// Aviso será logado pelo main.go
	}

	salt, err := store.getOrCreateEncryptionSalt()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter salt de criptografia: %w", err)
	}

	encryptor, err := security.NewEncryptorWithSalt(encryptionKey, salt)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar encryptor: %w", err)
	}
	store.encryptor = encryptor

	return store, nil
}

// migrate executa as migrações do banco
func (s *PostgresStore) migrate() error {
	migrations := []string{
		// Extensão UUID (ignora erro se não houver privilégio)
		`DO $$
		BEGIN
			CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		EXCEPTION
			WHEN insufficient_privilege THEN
				RAISE NOTICE 'uuid-ossp extension not available';
		END
		$$;`,

		// Tabela users
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(50) PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255),
			password VARCHAR(255) NOT NULL,
			terms_accepted BOOLEAN DEFAULT FALSE,
			terms_accepted_at TIMESTAMP,
			locale VARCHAR(10) DEFAULT 'pt-BR',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		// Migration: adicionar coluna locale se não existir
		`DO $$ 
		BEGIN 
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'users' AND column_name = 'locale') 
			THEN 
				ALTER TABLE users ADD COLUMN locale VARCHAR(10) DEFAULT 'pt-BR';
			END IF;
		END $$`,

		// Configuração do sistema (ex: salt de criptografia)
		`CREATE TABLE IF NOT EXISTS system_config (
			key VARCHAR(100) PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Idempotência (evita criação duplicada por retries)
		`CREATE TABLE IF NOT EXISTS idempotency_keys (
			user_id VARCHAR(50) NOT NULL,
			key VARCHAR(120) NOT NULL,
			resource_type VARCHAR(50) NOT NULL,
			resource_id VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, key, resource_type)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_idempotency_created ON idempotency_keys(created_at DESC)`,

		// Tabela box_items (com limite de 10KB para content)
		`CREATE TABLE IF NOT EXISTS box_items (
			id VARCHAR(50) PRIMARY KEY,
			user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL DEFAULT 'info',
			title VARCHAR(512) NOT NULL,
			content VARCHAR(10000),
			category VARCHAR(50),
			recipient VARCHAR(512),
			is_important BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Tabela guardians (notas limitadas a 1KB)
		`CREATE TABLE IF NOT EXISTS guardians (
			id VARCHAR(50) PRIMARY KEY,
			user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(512) NOT NULL,
			email VARCHAR(512),
			phone VARCHAR(128),
			relationship VARCHAR(255),
			role VARCHAR(20) DEFAULT 'viewer',
			notes VARCHAR(512),
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
		// Colunas para Social Auth (Google, Apple)
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS provider VARCHAR(50) DEFAULT 'email'`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS provider_id VARCHAR(255)`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT`,

		// Índices de usuários
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(LOWER(email))`,
		`CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_users_provider ON users(provider, provider_id) WHERE provider_id IS NOT NULL`,

		// Índices de box_items (performance em listagens e filtros)
		`CREATE INDEX IF NOT EXISTS idx_box_items_user ON box_items(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_box_items_user_type ON box_items(user_id, type)`,
		`CREATE INDEX IF NOT EXISTS idx_box_items_user_created ON box_items(user_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_box_items_important ON box_items(user_id, is_important) WHERE is_important = TRUE`,

		// Índices de guardians
		`CREATE INDEX IF NOT EXISTS idx_guardians_user ON guardians(user_id)`,

		// Migrações: Compartilhamento integrado
		`ALTER TABLE box_items ADD COLUMN IF NOT EXISTS is_shared BOOLEAN DEFAULT FALSE`,
		`ALTER TABLE box_items ADD COLUMN IF NOT EXISTS guardian_ids TEXT[]`,
		`ALTER TABLE guardians ADD COLUMN IF NOT EXISTS access_token VARCHAR(100)`,
		`ALTER TABLE guardians ADD COLUMN IF NOT EXISTS access_pin VARCHAR(255)`,
		`ALTER TABLE guardians ADD COLUMN IF NOT EXISTS access_type VARCHAR(20) DEFAULT 'normal'`,
		`CREATE INDEX IF NOT EXISTS idx_guardians_token ON guardians(access_token) WHERE access_token IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_box_items_shared ON box_items(user_id, is_shared) WHERE is_shared = TRUE`,
		`CREATE INDEX IF NOT EXISTS idx_box_items_guardian_ids ON box_items USING GIN (guardian_ids)`,
		`CREATE INDEX IF NOT EXISTS idx_box_items_user_updated ON box_items(user_id, updated_at DESC)`,

		// Índices de guide_progress (para verificar progresso rapidamente)
		`CREATE INDEX IF NOT EXISTS idx_guide_progress_user ON guide_progress(user_id)`,

		// =======================================================================
		// AUDITORIA E SEGURANÇA
		// =======================================================================
		// Tabela de auditoria para rastrear ações sensíveis (LGPD)
		// Removido user_agent para economizar espaço no banco
		`CREATE TABLE IF NOT EXISTS audit_log (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(50),
			action VARCHAR(100) NOT NULL,
			resource_type VARCHAR(50),
			resource_id VARCHAR(50),
			ip_address VARCHAR(45),
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

		// =======================================================================
		// FEEDBACK
		// =======================================================================
		// Feedbacks com limites de tamanho para economizar espaço
		`CREATE TABLE IF NOT EXISTS feedbacks (
			id VARCHAR(50) PRIMARY KEY,
			user_id VARCHAR(50) REFERENCES users(id) ON DELETE SET NULL,
			user_email VARCHAR(255),
			type VARCHAR(50) NOT NULL DEFAULT 'suggestion',
			message VARCHAR(2000) NOT NULL,
			page VARCHAR(100),
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			admin_note VARCHAR(500),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_feedbacks_user ON feedbacks(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_feedbacks_status ON feedbacks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_feedbacks_created ON feedbacks(created_at DESC)`,
		`ALTER TABLE feedbacks ADD COLUMN IF NOT EXISTS user_agent VARCHAR(255)`,

		// =======================================================================
		// ANALYTICS (com limpeza automática de eventos antigos)
		// =======================================================================
		`CREATE TABLE IF NOT EXISTS analytics_events (
			id VARCHAR(50) PRIMARY KEY,
			user_id VARCHAR(50),
			event_type VARCHAR(50) NOT NULL,
			page VARCHAR(100),
			details JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_user ON analytics_events(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_type ON analytics_events(event_type)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_created ON analytics_events(created_at DESC)`,
		// Nota: Índice parcial com CURRENT_DATE não é permitido (não-IMMUTABLE)
		// Consultas usam WHERE created_at >= date_trunc('day', CURRENT_TIMESTAMP) no runtime

		// =======================================================================
		// COMPARTILHAMENTO E ACESSO (Links para Guardiões)
		// =======================================================================
		`CREATE TABLE IF NOT EXISTS share_links (
			id VARCHAR(50) PRIMARY KEY,
			user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			guardian_id VARCHAR(50) REFERENCES guardians(id) ON DELETE SET NULL,
			token VARCHAR(100) NOT NULL UNIQUE,
			type VARCHAR(20) NOT NULL DEFAULT 'normal',
			name VARCHAR(255) NOT NULL,
			pin_hash VARCHAR(255),
			categories TEXT[],
			expires_at TIMESTAMP,
			max_uses INT DEFAULT 0,
			usage_count INT DEFAULT 0,
			last_used_at TIMESTAMP,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_share_links_user ON share_links(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_share_links_token ON share_links(token) WHERE is_active = TRUE`,
		`CREATE INDEX IF NOT EXISTS idx_share_links_guardian ON share_links(guardian_id)`,
		`ALTER TABLE share_links ADD COLUMN IF NOT EXISTS guardian_ids TEXT[]`,

		// Registro de acessos aos links
		`CREATE TABLE IF NOT EXISTS share_link_accesses (
			id VARCHAR(50) PRIMARY KEY,
			share_link_id VARCHAR(50) NOT NULL REFERENCES share_links(id) ON DELETE CASCADE,
			ip_address VARCHAR(45),
			user_agent VARCHAR(500),
			accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_share_accesses_link ON share_link_accesses(share_link_id)`,

		// Ajustes de tamanho para campos curtos (com espaço para criptografia)
		`ALTER TABLE box_items ALTER COLUMN title TYPE VARCHAR(512)`,
		`ALTER TABLE box_items ALTER COLUMN recipient TYPE VARCHAR(512)`,
		`ALTER TABLE guardians ALTER COLUMN name TYPE VARCHAR(512)`,
		`ALTER TABLE guardians ALTER COLUMN email TYPE VARCHAR(512)`,
		`ALTER TABLE guardians ALTER COLUMN phone TYPE VARCHAR(128)`,
		`ALTER TABLE guardians ALTER COLUMN relationship TYPE VARCHAR(255)`,
		`ALTER TABLE guardians ALTER COLUMN notes TYPE VARCHAR(512)`,
		`ALTER TABLE share_links ALTER COLUMN name TYPE VARCHAR(255)`,
		`ALTER TABLE feedbacks ALTER COLUMN page TYPE VARCHAR(255)`,

		// =======================================================================
		// RECUPERAÇÃO DE SENHA
		// =======================================================================
		`CREATE TABLE IF NOT EXISTS password_reset_tokens (
			id VARCHAR(50) PRIMARY KEY,
			user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash VARCHAR(255) NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			used_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_password_reset_user ON password_reset_tokens(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_password_reset_expires ON password_reset_tokens(expires_at)`,

		// =======================================================================
		// PROTOCOLO DE EMERGÊNCIA
		// =======================================================================
		`CREATE TABLE IF NOT EXISTS emergency_protocols (
			user_id VARCHAR(50) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			is_active BOOLEAN DEFAULT FALSE,
			activated_at TIMESTAMP,
			activated_by VARCHAR(50),
			deactivated_at TIMESTAMP,
			reason VARCHAR(500),
			notify_guardians BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
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

// CleanupOldLogs remove logs e analytics antigos para economizar espaço
// Deve ser chamado periodicamente (ex: diariamente)
func (s *PostgresStore) CleanupOldLogs(retentionDays int) error {
	if retentionDays < 7 {
		retentionDays = 7 // Mínimo de 7 dias
	}

	queries := []string{
		// Limpar audit_log antigos (manter últimos N dias)
		fmt.Sprintf(`DELETE FROM audit_log WHERE created_at < NOW() - INTERVAL '%d days'`, retentionDays),

		// Limpar analytics_events antigos
		fmt.Sprintf(`DELETE FROM analytics_events WHERE created_at < NOW() - INTERVAL '%d days'`, retentionDays),

		// Limpar acessos de links compartilhados antigos
		fmt.Sprintf(`DELETE FROM share_link_accesses WHERE accessed_at < NOW() - INTERVAL '%d days'`, retentionDays),

		// Limpar deletion_tokens expirados
		`DELETE FROM deletion_tokens WHERE expires_at < NOW()`,

		// Limpar tokens de reset de senha expirados ou usados
		`DELETE FROM password_reset_tokens WHERE expires_at < NOW() OR used_at IS NOT NULL`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("cleanup error: %w", err)
		}
	}

	return nil
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
	var locale sql.NullString
	err := s.db.QueryRow(`
		SELECT id, email, name, password, locale, created_at
		FROM users WHERE LOWER(email) = $1
	`, normalized).Scan(&user.ID, &user.Email, &user.Name, &user.Password, &locale, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		// Erro de banco é tratado como usuário não encontrado
		// O chamador deve decidir como lidar com isso
		return nil, false
	}

	if locale.Valid {
		user.Locale = locale.String
	}

	return &user, true
}

func (s *PostgresStore) GetUserByID(id string) (*User, bool) {
	var user User
	var name, locale sql.NullString

	err := s.db.QueryRow(`
		SELECT id, email, name, password, locale, created_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &name, &user.Password, &locale, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		// Erro de banco é tratado como usuário não encontrado
		return nil, false
	}

	user.Name = name.String
	if locale.Valid {
		user.Locale = locale.String
	}
	return &user, true
}

// UpdateUserPassword atualiza a senha de um usuário
func (s *PostgresStore) UpdateUserPassword(userID, hashedPassword string) error {
	result, err := s.db.Exec(`UPDATE users SET password = $1, updated_at = $2 WHERE id = $3`,
		hashedPassword, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("erro ao atualizar senha: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// UpdateUserLocale atualiza o idioma preferido do usuário
func (s *PostgresStore) UpdateUserLocale(userID, locale string) error {
	result, err := s.db.Exec(`UPDATE users SET locale = $1, updated_at = $2 WHERE id = $3`,
		locale, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("erro ao atualizar locale: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
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

	// Auditoria de deleção deve ser feita pelo chamador
	return nil
}

// ============================================================================
// SOCIAL AUTH (Google, Apple)
// ============================================================================

// CreateOrUpdateSocialUser cria ou atualiza um usuário via login social
func (s *PostgresStore) CreateOrUpdateSocialUser(provider AuthProvider, providerID, email, name, avatarURL string) (*User, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	now := time.Now()

	// Primeiro, tentar encontrar pelo provider + providerID
	existingUser, found := s.GetUserByProvider(provider, providerID)
	if found {
		// Atualizar informações se necessário
		_, err := s.db.Exec(`
			UPDATE users SET name = $1, avatar_url = $2, updated_at = $3
			WHERE id = $4
		`, name, avatarURL, now, existingUser.ID)
		if err != nil {
			return nil, fmt.Errorf("erro ao atualizar usuário social: %w", err)
		}
		existingUser.Name = name
		existingUser.AvatarURL = avatarURL
		return existingUser, nil
	}

	// Verificar se já existe usuário com este email
	existingByEmail, foundByEmail := s.GetUserByEmail(normalized)
	if foundByEmail {
		// Vincular o provedor ao usuário existente
		err := s.LinkSocialProvider(existingByEmail.ID, provider, providerID)
		if err != nil {
			return nil, err
		}
		// Atualizar avatar se não tiver
		if existingByEmail.AvatarURL == "" && avatarURL != "" {
			s.db.Exec(`UPDATE users SET avatar_url = $1, updated_at = $2 WHERE id = $3`, avatarURL, now, existingByEmail.ID)
		}
		existingByEmail.Provider = provider
		existingByEmail.ProviderID = providerID
		existingByEmail.AvatarURL = avatarURL
		return existingByEmail, nil
	}

	// Criar novo usuário
	id := fmt.Sprintf("usr_%d", now.UnixNano())
	_, err := s.db.Exec(`
		INSERT INTO users (id, email, name, password, provider, provider_id, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, '', $4, $5, $6, $7, $8)
	`, id, email, name, provider, providerID, avatarURL, now, now)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("erro ao criar usuário social: %w", err)
	}

	return &User{
		ID:         id,
		Email:      email,
		Name:       name,
		Provider:   provider,
		ProviderID: providerID,
		AvatarURL:  avatarURL,
		CreatedAt:  now,
	}, nil
}

// GetUserByProvider busca um usuário pelo provedor de autenticação
func (s *PostgresStore) GetUserByProvider(provider AuthProvider, providerID string) (*User, bool) {
	var user User
	var name, avatarURL sql.NullString

	err := s.db.QueryRow(`
		SELECT id, email, name, password, provider, provider_id, avatar_url, created_at
		FROM users WHERE provider = $1 AND provider_id = $2
	`, provider, providerID).Scan(
		&user.ID, &user.Email, &name, &user.Password,
		&user.Provider, &user.ProviderID, &avatarURL, &user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		return nil, false
	}

	user.Name = name.String
	user.AvatarURL = avatarURL.String
	return &user, true
}

// LinkSocialProvider vincula um provedor social a um usuário existente
func (s *PostgresStore) LinkSocialProvider(userID string, provider AuthProvider, providerID string) error {
	result, err := s.db.Exec(`
		UPDATE users SET provider = $1, provider_id = $2, updated_at = $3
		WHERE id = $4
	`, provider, providerID, time.Now(), userID)

	if err != nil {
		return fmt.Errorf("erro ao vincular provedor social: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

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
// Em caso de erro, falha fechado para evitar persistir PII em texto puro.
func (s *PostgresStore) encryptSensitive(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if s.encryptor == nil {
		return "", fmt.Errorf("encryptor not configured")
	}
	encrypted, err := s.encryptor.Encrypt(value)
	if err != nil {
		return "", err
	}
	return "enc:" + encrypted, nil
}

// decryptSensitive descriptografa um valor se estiver criptografado
func (s *PostgresStore) decryptSensitive(value string) string {
	if value == "" || s.encryptor == nil {
		return value
	}

	encryptedValue := value
	if strings.HasPrefix(value, "enc:") {
		encryptedValue = strings.TrimPrefix(value, "enc:")
		decrypted, err := s.encryptor.Decrypt(encryptedValue)
		if err != nil {
			return ""
		}
		return decrypted
	}

	if !isLikelyCiphertext(encryptedValue) {
		return value
	}

	decrypted, err := s.encryptor.Decrypt(encryptedValue)
	if err != nil {
		return ""
	}
	return decrypted
}

func (s *PostgresStore) getOrCreateEncryptionSalt() ([]byte, error) {
	envSalt := strings.TrimSpace(os.Getenv("ENCRYPTION_SALT"))
	if envSalt != "" {
		return decodeSalt(envSalt)
	}

	var stored string
	err := s.db.QueryRow(`SELECT value FROM system_config WHERE key = $1`, "encryption_salt").Scan(&stored)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if stored == "" {
		raw := make([]byte, 16)
		if _, err := rand.Read(raw); err != nil {
			return nil, err
		}
		stored = base64.StdEncoding.EncodeToString(raw)
		_, err := s.db.Exec(`
			INSERT INTO system_config (key, value, updated_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = $3
		`, "encryption_salt", stored, time.Now())
		if err != nil {
			return nil, err
		}
	}

	return decodeSalt(stored)
}

func decodeSalt(value string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return nil, fmt.Errorf("invalid ENCRYPTION_SALT")
	}
	if len(decoded) < 16 {
		return nil, fmt.Errorf("invalid ENCRYPTION_SALT length")
	}
	return decoded, nil
}

func isLikelyCiphertext(value string) bool {
	if len(value) < 24 {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return false
	}
	// nonce (12) + tag (16) + payload (>=0)
	return len(decoded) >= 28
}

// ============================================================================
// BOX ITEMS
// ============================================================================

// GetBoxItems retorna itens (compatibilidade - sem paginação)
func (s *PostgresStore) GetBoxItems(userID string) ([]*BoxItem, error) {
	return s.ListBoxItems(userID), nil
}

// ListBoxItems lista todos os itens (CUIDADO: sem paginação, usar apenas para exportação)
// Em caso de erro, retorna lista vazia (comportamento esperado para exportação)
func (s *PostgresStore) ListBoxItems(userID string) []*BoxItem {
	// Query com campos específicos (não usa SELECT *)
	rows, err := s.db.Query(`
		SELECT id, user_id, type, title, content, category, recipient, is_important, is_shared, guardian_ids, created_at, updated_at
		FROM box_items 
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT 1000
	`, userID)
	if err != nil {
		return []*BoxItem{}
	}
	defer rows.Close()

	var items []*BoxItem
	for rows.Next() {
		var item BoxItem
		var title, content, category, recipient sql.NullString
		var guardianIDs pq.StringArray
		err := rows.Scan(
			&item.ID, &item.UserID, &item.Type, &title,
			&content, &category, &recipient,
			&item.IsImportant, &item.IsShared, &guardianIDs, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			// Pular itens com erro de leitura
			continue
		}
		// Descriptografar dados sensíveis
		item.Title = s.decryptSensitive(title.String)
		item.Content = s.decryptSensitive(content.String)
		item.Category = category.String
		item.Recipient = s.decryptSensitive(recipient.String)
		item.GuardianIDs = guardianIDs
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
			SELECT id, type, title, category, is_important, is_shared, guardian_ids, updated_at
			FROM box_items 
			WHERE user_id = $1 AND id < $2
			ORDER BY id DESC
			LIMIT $3
		`, userID, params.Cursor, params.Limit+1)
	} else {
		// Primeira página
		rows, err = s.db.Query(`
			SELECT id, type, title, category, is_important, is_shared, guardian_ids, updated_at
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
		var guardianIDs pq.StringArray
		err := rows.Scan(
			&item.ID, &item.Type, &title, &category,
			&item.IsImportant, &item.IsShared, &guardianIDs, &item.UpdatedAt,
		)
		if err != nil {
			// Pular itens com erro de leitura
			continue
		}
		item.Title = s.decryptSensitive(title.String)
		item.Category = category.String
		item.GuardianIDs = guardianIDs
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
	var guardianIDs pq.StringArray

	// Query com campos específicos (não usa SELECT *)
	err := s.db.QueryRow(`
		SELECT id, user_id, type, title, content, category, recipient, is_important, is_shared, guardian_ids, created_at, updated_at
		FROM box_items 
		WHERE user_id = $1 AND id = $2
	`, userID, itemID).Scan(
		&item.ID, &item.UserID, &item.Type, &title,
		&content, &category, &recipient,
		&item.IsImportant, &item.IsShared, &guardianIDs, &item.CreatedAt, &item.UpdatedAt,
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
	item.GuardianIDs = guardianIDs
	return &item, nil
}

// CreateBoxItem cria um novo item com dados criptografados
func (s *PostgresStore) CreateBoxItem(userID string, item *BoxItem) (*BoxItem, error) {
	id := fmt.Sprintf("itm_%d", time.Now().UnixNano())
	return s.CreateBoxItemWithID(userID, item, id)
}

// CreateBoxItemWithID cria um novo item com ID pré-definido (idempotência).
func (s *PostgresStore) CreateBoxItemWithID(userID string, item *BoxItem, itemID string) (*BoxItem, error) {
	now := time.Now()

	// Criptografar dados sensíveis antes de salvar
	encTitle, err := s.encryptSensitive(item.Title)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar título: %w", err)
	}
	encContent, err := s.encryptSensitive(item.Content)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar conteúdo: %w", err)
	}
	encRecipient, err := s.encryptSensitive(item.Recipient)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar destinatário: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO box_items (id, user_id, type, title, content, category, recipient, is_important, is_shared, guardian_ids, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, itemID, userID, item.Type, encTitle, encContent, item.Category, encRecipient, item.IsImportant, item.IsShared, pq.Array(item.GuardianIDs), now, now)

	if err != nil {
		return nil, err
	}

	item.ID = itemID
	item.UserID = userID
	item.CreatedAt = now
	item.UpdatedAt = now
	return item, nil
}

// UpdateBoxItem atualiza um item existente com dados criptografados
func (s *PostgresStore) UpdateBoxItem(userID, itemID string, updates *BoxItem) (*BoxItem, error) {
	// Criptografar dados sensíveis antes de atualizar
	encTitle, err := s.encryptSensitive(updates.Title)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar título: %w", err)
	}
	encContent, err := s.encryptSensitive(updates.Content)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar conteúdo: %w", err)
	}
	encRecipient, err := s.encryptSensitive(updates.Recipient)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar destinatário: %w", err)
	}

	result, err := s.db.Exec(`
		UPDATE box_items 
		SET title = $1, content = $2, type = $3, category = $4, recipient = $5, is_important = $6, is_shared = $7, guardian_ids = $8, updated_at = $9
		WHERE user_id = $10 AND id = $11
	`, encTitle, encContent, updates.Type, updates.Category, encRecipient, updates.IsImportant, updates.IsShared, pq.Array(updates.GuardianIDs), time.Now(), userID, itemID)

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
// Em caso de erro, retorna lista vazia
func (s *PostgresStore) ListGuardians(userID string) []*Guardian {
	rows, err := s.db.Query(`
		SELECT id, user_id, name, email, phone, relationship, role, notes, access_token, access_pin, access_type, created_at, updated_at
		FROM guardians 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`, userID)
	if err != nil {
		return []*Guardian{}
	}
	defer rows.Close()

	var guardians []*Guardian
	for rows.Next() {
		var g Guardian
		var name, email, phone, relationship, notes, accessToken, accessPIN, accessType sql.NullString
		err := rows.Scan(
			&g.ID, &g.UserID, &name, &email, &phone,
			&relationship, &g.Role, &notes, &accessToken, &accessPIN, &accessType,
			&g.CreatedAt, &g.UpdatedAt,
		)
		if err != nil {
			// Pular guardiões com erro de leitura
			continue
		}
		// Descriptografar dados sensíveis (PII)
		g.Name = s.decryptSensitive(name.String)
		g.Email = s.decryptSensitive(email.String)
		g.Phone = s.decryptSensitive(phone.String)
		g.Relationship = relationship.String
		g.Notes = s.decryptSensitive(notes.String)
		g.AccessToken = accessToken.String
		g.AccessPIN = accessPIN.String
		g.HasPIN = accessPIN.String != ""
		g.AccessType = GuardianAccessType(accessType.String)

		s.ensureGuardianAccessToken(&g)

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
			SELECT id, user_id, name, email, phone, relationship, role, notes, access_token, access_type, created_at, updated_at
			FROM guardians 
			WHERE user_id = $1 AND id < $2
			ORDER BY id DESC
			LIMIT $3
		`, userID, params.Cursor, params.Limit+1)
	} else {
		rows, err = s.db.Query(`
			SELECT id, user_id, name, email, phone, relationship, role, notes, access_token, access_type, created_at, updated_at
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
		var name, email, phone, relationship, notes, accessToken, accessType sql.NullString
		err := rows.Scan(
			&g.ID, &g.UserID, &name, &email, &phone,
			&relationship, &g.Role, &notes, &accessToken, &accessType,
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
		g.AccessToken = accessToken.String
		g.AccessType = GuardianAccessType(accessType.String)
		s.ensureGuardianAccessToken(&g)
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

func (s *PostgresStore) ensureGuardianAccessToken(g *Guardian) {
	if g == nil {
		return
	}
	if g.AccessToken == "" || isLegacyAccessToken(g.AccessToken) {
		newToken := generateAccessToken()
		g.AccessToken = newToken
		// Atualizar no banco em background
		go func(id, token string) {
			s.db.Exec(`UPDATE guardians SET access_token = $1 WHERE id = $2`, token, id)
		}(g.ID, newToken)
	}
}

func isLegacyAccessToken(token string) bool {
	return strings.HasPrefix(token, "gat_")
}

// CountGuardians conta o total de guardiões de um usuário
func (s *PostgresStore) CountGuardians(userID string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM guardians WHERE user_id = $1
	`, userID).Scan(&count)
	return count, err
}

// GetGuardianByAccessToken busca um guardião pelo seu token de acesso
func (s *PostgresStore) GetGuardianByAccessToken(token string) (*Guardian, error) {
	var g Guardian
	var name, email, phone, relationship, notes, accessToken, accessPIN, accessType sql.NullString

	err := s.db.QueryRow(`
		SELECT id, user_id, name, email, phone, relationship, role, notes, access_token, access_pin, access_type, created_at, updated_at
		FROM guardians 
		WHERE access_token = $1
	`, token).Scan(
		&g.ID, &g.UserID, &name, &email, &phone,
		&relationship, &g.Role, &notes, &accessToken, &accessPIN, &accessType,
		&g.CreatedAt, &g.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	g.Name = s.decryptSensitive(name.String)
	g.Email = s.decryptSensitive(email.String)
	g.Phone = s.decryptSensitive(phone.String)
	g.Relationship = relationship.String
	g.Notes = s.decryptSensitive(notes.String)
	g.AccessToken = accessToken.String
	g.AccessPIN = accessPIN.String
	g.HasPIN = accessPIN.String != ""
	g.AccessType = GuardianAccessType(accessType.String)
	return &g, nil
}

// ListSharedItems lista itens compartilhados de um usuário
func (s *PostgresStore) ListSharedItems(userID string) []*BoxItem {
	rows, err := s.db.Query(`
		SELECT id, user_id, type, title, content, category, recipient, is_important, is_shared, guardian_ids, created_at, updated_at
		FROM box_items 
		WHERE user_id = $1 AND is_shared = TRUE
		ORDER BY updated_at DESC
		LIMIT 100
	`, userID)
	if err != nil {
		return []*BoxItem{}
	}
	defer rows.Close()

	var items []*BoxItem
	for rows.Next() {
		var item BoxItem
		var title, content, category, recipient sql.NullString
		var guardianIDs pq.StringArray
		err := rows.Scan(
			&item.ID, &item.UserID, &item.Type, &title,
			&content, &category, &recipient,
			&item.IsImportant, &item.IsShared, &guardianIDs, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			continue
		}
		item.Title = s.decryptSensitive(title.String)
		item.Content = s.decryptSensitive(content.String)
		item.Category = category.String
		item.Recipient = s.decryptSensitive(recipient.String)
		item.GuardianIDs = guardianIDs
		items = append(items, &item)
	}

	return items
}

// CreateGuardian cria um novo guardião com dados criptografados
func (s *PostgresStore) CreateGuardian(userID string, guardian *Guardian) (*Guardian, error) {
	id := fmt.Sprintf("grd_%d", time.Now().UnixNano())
	return s.CreateGuardianWithID(userID, guardian, id)
}

func (s *PostgresStore) CreateGuardianWithID(userID string, guardian *Guardian, guardianID string) (*Guardian, error) {
	now := time.Now()
	role := guardian.Role
	if role == "" {
		role = "viewer"
	}

	// Gerar access_token único para o guardião
	accessToken := generateAccessToken()
	accessType := guardian.AccessType
	if accessType == "" {
		accessType = GuardianAccessNormal
	}

	// Criptografar dados sensíveis (PII)
	encName, err := s.encryptSensitive(guardian.Name)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar nome: %w", err)
	}
	encEmail, err := s.encryptSensitive(guardian.Email)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar email: %w", err)
	}
	encPhone, err := s.encryptSensitive(guardian.Phone)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar telefone: %w", err)
	}
	encNotes, err := s.encryptSensitive(guardian.Notes)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar notas: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO guardians (id, user_id, name, email, phone, relationship, role, notes, access_token, access_pin, access_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, guardianID, userID, encName, encEmail, encPhone, guardian.Relationship, role, encNotes, accessToken, guardian.AccessPIN, accessType, now, now)

	if err != nil {
		return nil, err
	}

	guardian.ID = guardianID
	guardian.UserID = userID
	guardian.Role = role
	guardian.AccessToken = accessToken
	guardian.HasPIN = guardian.AccessPIN != ""
	guardian.AccessType = accessType
	guardian.CreatedAt = now
	guardian.UpdatedAt = now
	return guardian, nil
}

// generateAccessToken gera um token único URL-safe de 20 caracteres
func generateAccessToken() string {
	b := make([]byte, 15)
	rand.Read(b)
	token := base64.URLEncoding.EncodeToString(b)
	// Remover caracteres especiais e limitar a 20 chars
	token = strings.ReplaceAll(token, "-", "")
	token = strings.ReplaceAll(token, "_", "")
	if len(token) > 20 {
		token = token[:20]
	}
	return token
}

// UpdateGuardian atualiza um guardião com dados criptografados
func (s *PostgresStore) UpdateGuardian(userID, guardianID string, updates *Guardian) (*Guardian, error) {
	// Criptografar dados sensíveis (PII)
	encName, err := s.encryptSensitive(updates.Name)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar nome: %w", err)
	}
	encEmail, err := s.encryptSensitive(updates.Email)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar email: %w", err)
	}
	encPhone, err := s.encryptSensitive(updates.Phone)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar telefone: %w", err)
	}
	encNotes, err := s.encryptSensitive(updates.Notes)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar notas: %w", err)
	}

	var result sql.Result

	// Se PIN foi fornecido, atualizar também
	if updates.AccessPIN != "" {
		result, err = s.db.Exec(`
			UPDATE guardians 
			SET name = $1, email = $2, phone = $3, relationship = $4, notes = $5, access_pin = $6, updated_at = $7
			WHERE user_id = $8 AND id = $9
		`, encName, encEmail, encPhone, updates.Relationship, encNotes, updates.AccessPIN, time.Now(), userID, guardianID)
	} else {
		result, err = s.db.Exec(`
			UPDATE guardians 
			SET name = $1, email = $2, phone = $3, relationship = $4, notes = $5, updated_at = $6
			WHERE user_id = $7 AND id = $8
		`, encName, encEmail, encPhone, updates.Relationship, encNotes, time.Now(), userID, guardianID)
	}

	if err != nil {
		return nil, err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, ErrNotFound
	}

	// Buscar guardião atualizado
	var g Guardian
	var name, email, phone, relationship, notes, accessToken, accessPIN, accessType sql.NullString
	err = s.db.QueryRow(`
		SELECT id, user_id, name, email, phone, relationship, role, notes, access_token, access_pin, access_type, created_at, updated_at
		FROM guardians WHERE user_id = $1 AND id = $2
	`, userID, guardianID).Scan(
		&g.ID, &g.UserID, &name, &email, &phone,
		&relationship, &g.Role, &notes, &accessToken, &accessPIN, &accessType,
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
	g.AccessToken = accessToken.String
	g.AccessPIN = accessPIN.String
	g.HasPIN = accessPIN.String != ""
	g.AccessType = GuardianAccessType(accessType.String)
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

// RegisterIdempotencyKey registra uma chave idempotente para um recurso.
func (s *PostgresStore) RegisterIdempotencyKey(userID, key, resourceType, resourceID string) (string, bool, error) {
	result, err := s.db.Exec(`
		INSERT INTO idempotency_keys (user_id, key, resource_type, resource_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING
	`, userID, key, resourceType, resourceID, time.Now())
	if err != nil {
		return "", false, err
	}

	rows, _ := result.RowsAffected()
	if rows == 1 {
		return "", true, nil
	}

	var existingID string
	err = s.db.QueryRow(`
		SELECT resource_id FROM idempotency_keys
		WHERE user_id = $1 AND key = $2 AND resource_type = $3
	`, userID, key, resourceType).Scan(&existingID)
	if err == sql.ErrNoRows {
		return "", false, ErrNotFound
	}
	if err != nil {
		return "", false, err
	}
	return existingID, false, nil
}

// DeleteIdempotencyKey remove uma chave idempotente (ex.: após falha).
func (s *PostgresStore) DeleteIdempotencyKey(userID, key, resourceType string) error {
	_, err := s.db.Exec(`
		DELETE FROM idempotency_keys WHERE user_id = $1 AND key = $2 AND resource_type = $3
	`, userID, key, resourceType)
	return err
}

// ============================================================================
// GUIDE PROGRESS
// ============================================================================

func (s *PostgresStore) GetGuideProgress(userID string) map[string]*GuideProgress {
	rows, err := s.db.Query(`
		SELECT user_id, card_id, status, completed_at
		FROM guide_progress WHERE user_id = $1 LIMIT 50
	`, userID)
	if err != nil {
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

	// Itens por tipo (máximo 20 tipos diferentes)
	rows, _ := s.db.Query(`SELECT type, COUNT(*) FROM box_items GROUP BY type LIMIT 20`)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var itemType string
			var count int
			rows.Scan(&itemType, &count)
			stats.ItemsByType[itemType] = count
		}
	}

	// Itens por categoria (máximo 20 categorias)
	rows2, _ := s.db.Query(`SELECT category, COUNT(*) FROM box_items WHERE category != '' GROUP BY category LIMIT 20`)
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
		SELECT id, email, name, created_at FROM users ORDER BY created_at DESC LIMIT 500
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

// ============================================================================
// FEEDBACK
// ============================================================================

// CreateFeedback salva um novo feedback
func (s *PostgresStore) CreateFeedback(f *Feedback) error {
	_, err := s.db.Exec(`
		INSERT INTO feedbacks (id, user_id, user_email, type, message, page, user_agent, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, f.ID, f.UserID, f.UserEmail, f.Type, f.Message, f.Page, f.UserAgent, f.Status, f.CreatedAt, f.UpdatedAt)
	return err
}

// ListFeedbacks retorna todos os feedbacks (para admin)
func (s *PostgresStore) ListFeedbacks(status string, limit int) ([]*Feedback, error) {
	var query string
	var args []interface{}

	if status != "" && status != "all" {
		query = `
			SELECT id, user_id, user_email, type, message, page, user_agent, status, admin_note, created_at, updated_at
			FROM feedbacks WHERE status = $1 ORDER BY created_at DESC LIMIT $2
		`
		args = []interface{}{status, limit}
	} else {
		query = `
			SELECT id, user_id, user_email, type, message, page, user_agent, status, admin_note, created_at, updated_at
			FROM feedbacks ORDER BY created_at DESC LIMIT $1
		`
		args = []interface{}{limit}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []*Feedback
	for rows.Next() {
		var f Feedback
		var userID, userEmail, page, userAgent, adminNote sql.NullString
		err := rows.Scan(&f.ID, &userID, &userEmail, &f.Type, &f.Message, &page, &userAgent, &f.Status, &adminNote, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			continue
		}
		f.UserID = userID.String
		f.UserEmail = userEmail.String
		f.Page = page.String
		f.UserAgent = userAgent.String
		f.AdminNote = adminNote.String
		feedbacks = append(feedbacks, &f)
	}

	return feedbacks, nil
}

// UpdateFeedbackStatus atualiza o status de um feedback
func (s *PostgresStore) UpdateFeedbackStatus(id, status, adminNote string) error {
	_, err := s.db.Exec(`
		UPDATE feedbacks SET status = $1, admin_note = $2, updated_at = $3 WHERE id = $4
	`, status, adminNote, time.Now(), id)
	return err
}

// GetFeedbackStats retorna estatísticas de feedbacks
func (s *PostgresStore) GetFeedbackStats() (total, pending int) {
	s.db.QueryRow(`SELECT COUNT(*) FROM feedbacks`).Scan(&total)
	s.db.QueryRow(`SELECT COUNT(*) FROM feedbacks WHERE status = 'pending'`).Scan(&pending)
	return
}

// ============================================================================
// ANALYTICS
// ============================================================================

// TrackEvent registra um evento de analytics
func (s *PostgresStore) TrackEvent(e *AnalyticsEvent) error {
	detailsJSON, _ := json.Marshal(e.Details)
	_, err := s.db.Exec(`
		INSERT INTO analytics_events (id, user_id, event_type, page, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, e.ID, e.UserID, e.EventType, e.Page, detailsJSON, e.CreatedAt)
	return err
}

// GetAnalyticsSummary retorna o resumo de analytics
func (s *PostgresStore) GetAnalyticsSummary() *AnalyticsSummary {
	summary := &AnalyticsSummary{
		EventsByType: make(map[string]int),
	}

	// Total de usuários
	s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&summary.TotalUsers)

	// Usuários novos hoje
	s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE`).Scan(&summary.NewUsersToday)

	// Usuários novos esta semana
	s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'`).Scan(&summary.NewUsersThisWeek)

	// Usuários ativos hoje (com eventos)
	s.db.QueryRow(`SELECT COUNT(DISTINCT user_id) FROM analytics_events WHERE created_at >= CURRENT_DATE AND user_id IS NOT NULL`).Scan(&summary.ActiveToday)

	// Usuários ativos esta semana
	s.db.QueryRow(`SELECT COUNT(DISTINCT user_id) FROM analytics_events WHERE created_at >= CURRENT_DATE - INTERVAL '7 days' AND user_id IS NOT NULL`).Scan(&summary.ActiveThisWeek)

	// Total de itens
	s.db.QueryRow(`SELECT COUNT(*) FROM box_items`).Scan(&summary.TotalItems)

	// Itens criados hoje
	s.db.QueryRow(`SELECT COUNT(*) FROM box_items WHERE created_at >= CURRENT_DATE`).Scan(&summary.ItemsCreatedToday)

	// Total de guardiões
	s.db.QueryRow(`SELECT COUNT(*) FROM guardians`).Scan(&summary.TotalGuardians)

	// Eventos hoje
	s.db.QueryRow(`SELECT COUNT(*) FROM analytics_events WHERE created_at >= CURRENT_DATE`).Scan(&summary.EventsToday)

	// Eventos por tipo (últimos 7 dias, máximo 30 tipos)
	rows, err := s.db.Query(`
		SELECT event_type, COUNT(*) as count
		FROM analytics_events
		WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
		GROUP BY event_type
		ORDER BY count DESC
		LIMIT 30
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var eventType string
			var count int
			rows.Scan(&eventType, &count)
			summary.EventsByType[eventType] = count
		}
	}

	// Feedbacks
	summary.TotalFeedbacks, summary.PendingFeedbacks = s.GetFeedbackStats()

	return summary
}

// GetRecentEvents retorna os eventos mais recentes
func (s *PostgresStore) GetRecentEvents(limit int) ([]*AnalyticsEvent, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, event_type, page, details, created_at
		FROM analytics_events
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*AnalyticsEvent
	for rows.Next() {
		var e AnalyticsEvent
		var userID, page sql.NullString
		var detailsJSON []byte
		err := rows.Scan(&e.ID, &userID, &e.EventType, &page, &detailsJSON, &e.CreatedAt)
		if err != nil {
			continue
		}
		e.UserID = userID.String
		e.Page = page.String
		if len(detailsJSON) > 0 {
			json.Unmarshal(detailsJSON, &e.Details)
		}
		events = append(events, &e)
	}

	return events, nil
}

// GetDailyStats retorna estatísticas diárias para gráficos
func (s *PostgresStore) GetDailyStats(days int) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(`
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as events,
			COUNT(DISTINCT user_id) as users
		FROM analytics_events
		WHERE created_at >= CURRENT_DATE - $1 * INTERVAL '1 day'
		GROUP BY DATE(created_at)
		ORDER BY date
	`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []map[string]interface{}
	for rows.Next() {
		var date time.Time
		var events, users int
		rows.Scan(&date, &events, &users)
		stats = append(stats, map[string]interface{}{
			"date":   date.Format("2006-01-02"),
			"events": events,
			"users":  users,
		})
	}

	return stats, nil
}

// ============================================================================
// SHARE LINKS (Compartilhamento com Guardiões)
// ============================================================================

// CreateShareLink cria um novo link de compartilhamento
func (s *PostgresStore) CreateShareLink(link *ShareLink) error {
	_, err := s.db.Exec(`
		INSERT INTO share_links (id, user_id, guardian_id, guardian_ids, token, type, name, pin_hash, categories, expires_at, max_uses, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, link.ID, link.UserID, nullString(link.GuardianID), pq.Array(link.GuardianIDs), link.Token, link.Type, link.Name,
		nullString(link.PIN), pq.Array(link.Categories), link.ExpiresAt, link.MaxUses, link.IsActive, link.CreatedAt, link.UpdatedAt)
	return err
}

// GetShareLinkByToken busca um link pelo token
func (s *PostgresStore) GetShareLinkByToken(token string) (*ShareLink, error) {
	var link ShareLink
	var guardianID, pinHash sql.NullString
	var expiresAt, lastUsedAt sql.NullTime
	var categories, guardianIDs pq.StringArray

	err := s.db.QueryRow(`
		SELECT id, user_id, guardian_id, guardian_ids, token, type, name, pin_hash, categories, expires_at, max_uses, usage_count, last_used_at, is_active, created_at, updated_at
		FROM share_links
		WHERE token = $1 AND is_active = TRUE
	`, token).Scan(&link.ID, &link.UserID, &guardianID, &guardianIDs, &link.Token, &link.Type, &link.Name,
		&pinHash, &categories, &expiresAt, &link.MaxUses, &link.UsageCount, &lastUsedAt, &link.IsActive, &link.CreatedAt, &link.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	link.GuardianID = guardianID.String
	link.GuardianIDs = guardianIDs
	link.PIN = pinHash.String
	link.Categories = categories
	if expiresAt.Valid {
		link.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		link.LastUsedAt = &lastUsedAt.Time
	}

	return &link, nil
}

// GetShareLinksByUser lista todos os links de um usuário
func (s *PostgresStore) GetShareLinksByUser(userID string) ([]*ShareLink, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, guardian_id, guardian_ids, token, type, name, categories, expires_at, max_uses, usage_count, last_used_at, is_active, created_at, updated_at
		FROM share_links
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*ShareLink
	for rows.Next() {
		var link ShareLink
		var guardianID sql.NullString
		var expiresAt, lastUsedAt sql.NullTime
		var categories, guardianIDs pq.StringArray

		err := rows.Scan(&link.ID, &link.UserID, &guardianID, &guardianIDs, &link.Token, &link.Type, &link.Name,
			&categories, &expiresAt, &link.MaxUses, &link.UsageCount, &lastUsedAt, &link.IsActive, &link.CreatedAt, &link.UpdatedAt)
		if err != nil {
			continue
		}

		link.GuardianID = guardianID.String
		link.GuardianIDs = guardianIDs
		link.Categories = categories
		if expiresAt.Valid {
			link.ExpiresAt = &expiresAt.Time
		}
		if lastUsedAt.Valid {
			link.LastUsedAt = &lastUsedAt.Time
		}
		links = append(links, &link)
	}

	return links, nil
}

// UpdateShareLink atualiza um link
func (s *PostgresStore) UpdateShareLink(link *ShareLink) error {
	_, err := s.db.Exec(`
		UPDATE share_links SET name = $1, categories = $2, expires_at = $3, max_uses = $4, is_active = $5, updated_at = $6
		WHERE id = $7 AND user_id = $8
	`, link.Name, pq.Array(link.Categories), link.ExpiresAt, link.MaxUses, link.IsActive, time.Now(), link.ID, link.UserID)
	return err
}

// DeleteShareLink remove um link
func (s *PostgresStore) DeleteShareLink(userID, linkID string) error {
	result, err := s.db.Exec(`DELETE FROM share_links WHERE id = $1 AND user_id = $2`, linkID, userID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// RecordShareLinkAccess registra um acesso a um link
func (s *PostgresStore) RecordShareLinkAccess(access *ShareLinkAccess) error {
	_, err := s.db.Exec(`
		INSERT INTO share_link_accesses (id, share_link_id, ip_address, user_agent, accessed_at)
		VALUES ($1, $2, $3, $4, $5)
	`, access.ID, access.ShareLinkID, access.IPAddress, access.UserAgent, access.AccessedAt)
	return err
}

// IncrementShareLinkUsage incrementa o contador de uso
func (s *PostgresStore) IncrementShareLinkUsage(linkID string) error {
	_, err := s.db.Exec(`
		UPDATE share_links SET usage_count = usage_count + 1, last_used_at = $1 WHERE id = $2
	`, time.Now(), linkID)
	return err
}

// ============================================================================
// PASSWORD RESET (Recuperação de Senha)
// ============================================================================

// CreatePasswordResetToken cria um token de recuperação
func (s *PostgresStore) CreatePasswordResetToken(token *PasswordResetToken) error {
	// Invalidar tokens anteriores do usuário
	s.db.Exec(`UPDATE password_reset_tokens SET used_at = $1 WHERE user_id = $2 AND used_at IS NULL`, time.Now(), token.UserID)

	_, err := s.db.Exec(`
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, token.ID, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt)
	return err
}

// GetPasswordResetToken busca um token válido
func (s *PostgresStore) GetPasswordResetToken(tokenHash string) (*PasswordResetToken, error) {
	var token PasswordResetToken
	var usedAt sql.NullTime

	err := s.db.QueryRow(`
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > $2
	`, tokenHash, time.Now()).Scan(&token.ID, &token.UserID, &token.Token, &token.ExpiresAt, &usedAt, &token.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if usedAt.Valid {
		token.UsedAt = &usedAt.Time
	}

	return &token, nil
}

// MarkPasswordResetTokenUsed marca um token como usado
func (s *PostgresStore) MarkPasswordResetTokenUsed(tokenID string) error {
	_, err := s.db.Exec(`UPDATE password_reset_tokens SET used_at = $1 WHERE id = $2`, time.Now(), tokenID)
	return err
}

// CleanupExpiredPasswordResetTokens remove tokens expirados
func (s *PostgresStore) CleanupExpiredPasswordResetTokens() error {
	_, err := s.db.Exec(`DELETE FROM password_reset_tokens WHERE expires_at < $1 OR used_at IS NOT NULL`, time.Now().Add(-24*time.Hour))
	return err
}

// ============================================================================
// EMERGENCY PROTOCOL (Protocolo de Emergência)
// ============================================================================

// GetEmergencyProtocol busca o protocolo de emergência de um usuário
func (s *PostgresStore) GetEmergencyProtocol(userID string) (*EmergencyProtocol, error) {
	var protocol EmergencyProtocol
	var activatedAt, deactivatedAt sql.NullTime
	var activatedBy, reason sql.NullString

	err := s.db.QueryRow(`
		SELECT user_id, is_active, activated_at, activated_by, deactivated_at, reason, notify_guardians
		FROM emergency_protocols WHERE user_id = $1
	`, userID).Scan(&protocol.UserID, &protocol.IsActive, &activatedAt, &activatedBy, &deactivatedAt, &reason, &protocol.NotifyGuardians)

	if err == sql.ErrNoRows {
		// Retornar protocolo padrão (não ativado)
		return &EmergencyProtocol{UserID: userID, IsActive: false, NotifyGuardians: true}, nil
	}
	if err != nil {
		return nil, err
	}

	if activatedAt.Valid {
		protocol.ActivatedAt = &activatedAt.Time
	}
	if deactivatedAt.Valid {
		protocol.DeactivatedAt = &deactivatedAt.Time
	}
	protocol.ActivatedBy = activatedBy.String
	protocol.Reason = reason.String

	return &protocol, nil
}

// UpdateEmergencyProtocol atualiza o protocolo de emergência
func (s *PostgresStore) UpdateEmergencyProtocol(protocol *EmergencyProtocol) error {
	_, err := s.db.Exec(`
		INSERT INTO emergency_protocols (user_id, is_active, activated_at, activated_by, deactivated_at, reason, notify_guardians, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id) DO UPDATE SET 
			is_active = $2, activated_at = $3, activated_by = $4, deactivated_at = $5, reason = $6, notify_guardians = $7, updated_at = $8
	`, protocol.UserID, protocol.IsActive, protocol.ActivatedAt, nullString(protocol.ActivatedBy),
		protocol.DeactivatedAt, nullString(protocol.Reason), protocol.NotifyGuardians, time.Now())
	return err
}

// nullString retorna sql.NullString para strings vazias
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
