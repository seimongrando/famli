package storage

import "time"

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
type BoxItem struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Type        ItemType  `json:"type"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Category    string    `json:"category,omitempty"`  // saúde, finanças, família, etc.
	Recipient   string    `json:"recipient,omitempty"` // para quem é (memórias)
	IsImportant bool      `json:"is_important"`
	CreatedAt   time.Time `json:"created_at"`
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
