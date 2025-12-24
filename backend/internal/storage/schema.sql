-- =============================================================================
-- FAMLI - Schema do Banco de Dados PostgreSQL
-- =============================================================================
-- Este arquivo contém o schema completo para o Famli.
-- Execute este script para criar todas as tabelas necessárias.
--
-- Uso: psql -d famli -f schema.sql
-- =============================================================================

-- Extensão para UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- TABELA: users
-- =============================================================================
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(50) PRIMARY KEY DEFAULT CONCAT('usr_', uuid_generate_v4()::text),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255),
    password VARCHAR(255) NOT NULL,
    terms_accepted BOOLEAN DEFAULT FALSE,
    terms_accepted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(LOWER(email));

-- =============================================================================
-- TABELA: system_config
-- =============================================================================
CREATE TABLE IF NOT EXISTS system_config (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- TABELA: idempotency_keys
-- =============================================================================
CREATE TABLE IF NOT EXISTS idempotency_keys (
    user_id VARCHAR(50) NOT NULL,
    key VARCHAR(120) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, key, resource_type)
);

CREATE INDEX IF NOT EXISTS idx_idempotency_created ON idempotency_keys(created_at DESC);

-- =============================================================================
-- TABELA: box_items
-- =============================================================================
CREATE TABLE IF NOT EXISTS box_items (
    id VARCHAR(50) PRIMARY KEY DEFAULT CONCAT('itm_', uuid_generate_v4()::text),
    user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL DEFAULT 'info',
    title VARCHAR(512) NOT NULL,
    content VARCHAR(10000),
    category VARCHAR(50),
    recipient VARCHAR(512),
    is_important BOOLEAN DEFAULT FALSE,
    is_shared BOOLEAN DEFAULT FALSE,
    guardian_ids TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_box_items_user ON box_items(user_id);
CREATE INDEX IF NOT EXISTS idx_box_items_type ON box_items(type);

-- =============================================================================
-- TABELA: guardians
-- =============================================================================
CREATE TABLE IF NOT EXISTS guardians (
    id VARCHAR(50) PRIMARY KEY DEFAULT CONCAT('grd_', uuid_generate_v4()::text),
    user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(512) NOT NULL,
    email VARCHAR(512),
    phone VARCHAR(128),
    relationship VARCHAR(255),
    role VARCHAR(50) DEFAULT 'viewer',
    notes VARCHAR(512),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_guardians_user ON guardians(user_id);

-- =============================================================================
-- TABELA: feedbacks
-- =============================================================================
CREATE TABLE IF NOT EXISTS feedbacks (
    id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(50) REFERENCES users(id) ON DELETE SET NULL,
    user_email VARCHAR(255),
    type VARCHAR(50) NOT NULL DEFAULT 'suggestion',
    message VARCHAR(2000) NOT NULL,
    page VARCHAR(255),
    user_agent VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    admin_note VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_feedbacks_user ON feedbacks(user_id);
CREATE INDEX IF NOT EXISTS idx_feedbacks_status ON feedbacks(status);
CREATE INDEX IF NOT EXISTS idx_feedbacks_created ON feedbacks(created_at DESC);

-- =============================================================================
-- TABELA: guide_progress
-- =============================================================================
CREATE TABLE IF NOT EXISTS guide_progress (
    user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, card_id)
);

-- =============================================================================
-- TABELA: settings
-- =============================================================================
CREATE TABLE IF NOT EXISTS settings (
    user_id VARCHAR(50) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    emergency_protocol_enabled BOOLEAN DEFAULT FALSE,
    notifications_enabled BOOLEAN DEFAULT TRUE,
    theme VARCHAR(20) DEFAULT 'light',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- FUNÇÃO: Atualizar updated_at automaticamente
-- =============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers para atualizar updated_at
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_box_items_updated_at ON box_items;
CREATE TRIGGER update_box_items_updated_at
    BEFORE UPDATE ON box_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_guardians_updated_at ON guardians;
CREATE TRIGGER update_guardians_updated_at
    BEFORE UPDATE ON guardians
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_guide_progress_updated_at ON guide_progress;
CREATE TRIGGER update_guide_progress_updated_at
    BEFORE UPDATE ON guide_progress
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_settings_updated_at ON settings;
CREATE TRIGGER update_settings_updated_at
    BEFORE UPDATE ON settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
