// =============================================================================
// FAMLI - Integração WhatsApp
// =============================================================================
// Este pacote gerencia toda a comunicação via WhatsApp, permitindo que
// usuários interajam com o Famli através de mensagens de texto, fotos e áudios.
//
// Funcionalidades:
// - Receber mensagens via webhook (Twilio)
// - Processar texto, imagens e áudios
// - Salvar conteúdo na Caixa Famli
// - Enviar respostas e confirmações
// - Notificar guardiões quando necessário
// =============================================================================

package whatsapp

import "time"

// =============================================================================
// TIPOS DE MENSAGEM
// =============================================================================

// MessageType define os tipos de mensagens que podemos receber do WhatsApp
type MessageType string

const (
	// MessageTypeText representa uma mensagem de texto simples
	MessageTypeText MessageType = "text"

	// MessageTypeImage representa uma imagem enviada pelo usuário
	MessageTypeImage MessageType = "image"

	// MessageTypeAudio representa um áudio/mensagem de voz
	MessageTypeAudio MessageType = "audio"

	// MessageTypeDocument representa um documento (PDF, etc.)
	MessageTypeDocument MessageType = "document"

	// MessageTypeLocation representa uma localização compartilhada
	MessageTypeLocation MessageType = "location"
)

// =============================================================================
// MENSAGEM RECEBIDA
// =============================================================================

// IncomingMessage representa uma mensagem recebida do WhatsApp via Twilio
// Esta estrutura é preenchida a partir do webhook do Twilio
type IncomingMessage struct {
	// MessageSid é o ID único da mensagem no Twilio
	MessageSid string `json:"message_sid"`

	// From é o número de telefone do remetente (formato: whatsapp:+5511999999999)
	From string `json:"from"`

	// To é o número de telefone do destinatário (nosso número Twilio)
	To string `json:"to"`

	// Body é o conteúdo textual da mensagem
	Body string `json:"body"`

	// NumMedia indica quantos arquivos de mídia estão anexados
	NumMedia int `json:"num_media"`

	// MediaContentType é o tipo MIME da mídia (ex: image/jpeg, audio/ogg)
	MediaContentType string `json:"media_content_type,omitempty"`

	// MediaUrl é a URL para download da mídia
	MediaUrl string `json:"media_url,omitempty"`

	// Latitude é a latitude se for uma mensagem de localização
	Latitude string `json:"latitude,omitempty"`

	// Longitude é a longitude se for uma mensagem de localização
	Longitude string `json:"longitude,omitempty"`

	// ProfileName é o nome do perfil do WhatsApp do remetente
	ProfileName string `json:"profile_name,omitempty"`

	// ReceivedAt é quando a mensagem foi recebida pelo nosso sistema
	ReceivedAt time.Time `json:"received_at"`
}

// GetMessageType determina o tipo de mensagem baseado no conteúdo
// Retorna o tipo apropriado para processamento
func (m *IncomingMessage) GetMessageType() MessageType {
	// Se tem mídia anexada, verificar o tipo
	if m.NumMedia > 0 {
		switch {
		case contains(m.MediaContentType, "image"):
			return MessageTypeImage
		case contains(m.MediaContentType, "audio"):
			return MessageTypeAudio
		case contains(m.MediaContentType, "application"):
			return MessageTypeDocument
		}
	}

	// Se tem coordenadas, é localização
	if m.Latitude != "" && m.Longitude != "" {
		return MessageTypeLocation
	}

	// Padrão: mensagem de texto
	return MessageTypeText
}

// =============================================================================
// MENSAGEM DE SAÍDA
// =============================================================================

// OutgoingMessage representa uma mensagem a ser enviada para o WhatsApp
type OutgoingMessage struct {
	// To é o número de destino (formato: whatsapp:+5511999999999)
	To string `json:"to"`

	// Body é o texto da mensagem
	Body string `json:"body"`

	// MediaUrl é uma URL de mídia a ser enviada (opcional)
	MediaUrl string `json:"media_url,omitempty"`
}

// =============================================================================
// SESSÃO DO USUÁRIO
// =============================================================================

// UserSession armazena o estado da conversa com um usuário
// Permite manter contexto entre mensagens
type UserSession struct {
	// PhoneNumber é o número do usuário (chave)
	PhoneNumber string `json:"phone_number"`

	// UserID é o ID do usuário no Famli (se vinculado)
	UserID string `json:"user_id,omitempty"`

	// State é o estado atual da conversa
	// Valores: "idle", "awaiting_title", "awaiting_category", "awaiting_confirmation"
	State string `json:"state"`

	// PendingItem armazena dados temporários de um item sendo criado
	PendingItem *PendingBoxItem `json:"pending_item,omitempty"`

	// LastMessageAt é quando a última mensagem foi recebida
	LastMessageAt time.Time `json:"last_message_at"`

	// CreatedAt é quando a sessão foi criada
	CreatedAt time.Time `json:"created_at"`
}

// PendingBoxItem armazena dados de um item que está sendo criado via WhatsApp
type PendingBoxItem struct {
	// Content é o conteúdo principal (texto, URL da imagem, etc.)
	Content string `json:"content"`

	// Type é o tipo do item (info, memory, note, etc.)
	Type string `json:"type"`

	// Title é o título do item (pode ser gerado automaticamente)
	Title string `json:"title"`

	// Category é a categoria (saúde, finanças, família, etc.)
	Category string `json:"category"`

	// MediaUrl é a URL da mídia se houver
	MediaUrl string `json:"media_url,omitempty"`

	// MediaType é o tipo da mídia
	MediaType string `json:"media_type,omitempty"`
}

// =============================================================================
// CONFIGURAÇÃO
// =============================================================================

// Config armazena a configuração do serviço WhatsApp
type Config struct {
	// TwilioAccountSid é o SID da conta Twilio
	TwilioAccountSid string

	// TwilioAuthToken é o token de autenticação do Twilio
	TwilioAuthToken string

	// TwilioPhoneNumber é o número de telefone do Twilio para WhatsApp
	// Formato: whatsapp:+14155238886 (sandbox) ou seu número verificado
	TwilioPhoneNumber string

	// WebhookBaseURL é a URL base para webhooks (ex: https://famli.me)
	WebhookBaseURL string

	// Enabled indica se a integração está ativa
	Enabled bool
}

// =============================================================================
// COMANDOS RECONHECIDOS
// =============================================================================

// Command representa um comando que o usuário pode enviar
type Command string

const (
	// CommandHelp mostra a ajuda
	CommandHelp Command = "ajuda"

	// CommandSave inicia o processo de salvar algo
	CommandSave Command = "guardar"

	// CommandList lista os últimos itens salvos
	CommandList Command = "listar"

	// CommandCancel cancela a operação atual
	CommandCancel Command = "cancelar"

	// CommandStatus mostra o status da conta
	CommandStatus Command = "status"

	// CommandLink vincula o número a uma conta Famli
	CommandLink Command = "vincular"
)

// =============================================================================
// FUNÇÕES AUXILIARES
// =============================================================================

// contains verifica se uma string contém outra (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsLower(s, substr))
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if toLower(s[i:i+len(substr)]) == toLower(substr) {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
