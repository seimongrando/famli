// =============================================================================
// FAMLI - Handler HTTP para WhatsApp
// =============================================================================
// Este arquivo expõe os endpoints HTTP para integração com WhatsApp via Twilio.
//
// Endpoints:
// - POST /api/whatsapp/webhook  - Recebe mensagens do Twilio
// - GET  /api/whatsapp/webhook  - Validação do webhook (Twilio verification)
// - POST /api/whatsapp/link     - Vincula número WhatsApp a uma conta Famli
// - GET  /api/whatsapp/status   - Verifica status da integração
//
// Fluxo do Webhook:
// 1. Twilio recebe mensagem no WhatsApp
// 2. Twilio envia POST para nosso webhook
// 3. Processamos a mensagem e geramos resposta
// 4. Retornamos TwiML com a resposta
// 5. Twilio envia resposta ao usuário
// =============================================================================

package whatsapp

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"famli/internal/auth"
)

// =============================================================================
// HANDLER PRINCIPAL
// =============================================================================

// Handler gerencia todas as requisições HTTP relacionadas ao WhatsApp
type Handler struct {
	// service é o serviço de processamento de mensagens
	service *Service

	// config é a configuração do WhatsApp
	config *Config
}

// NewHandler cria uma nova instância do handler WhatsApp
//
// Parâmetros:
//   - service: serviço de processamento de mensagens
//   - config: configuração com credenciais Twilio
//
// Retorna:
//   - *Handler: handler configurado
func NewHandler(service *Service, config *Config) *Handler {
	return &Handler{
		service: service,
		config:  config,
	}
}

// =============================================================================
// WEBHOOK - RECEBER MENSAGENS
// =============================================================================

// Webhook é o endpoint principal que recebe mensagens do Twilio
//
// O Twilio envia um POST com dados da mensagem como form-urlencoded.
// Respondemos com TwiML contendo a mensagem de resposta.
//
// Endpoint: POST /api/whatsapp/webhook
// Content-Type: application/x-www-form-urlencoded
//
// Campos recebidos do Twilio:
//   - MessageSid: ID único da mensagem
//   - From: whatsapp:+5511999999999
//   - To: whatsapp:+14155238886 (nosso número)
//   - Body: conteúdo da mensagem
//   - NumMedia: quantidade de mídias anexadas
//   - MediaUrl0, MediaContentType0: dados da mídia
//
// Resposta: TwiML XML com a mensagem de resposta
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	// Verificar se a integração está habilitada
	if h.config == nil || !h.config.Enabled {
		log.Println("[WhatsApp] Webhook recebido mas integração está desabilitada")
		h.writeEmptyTwiML(w)
		return
	}

	// Parsear a mensagem recebida
	msg, err := ParseWebhookRequest(r)
	if err != nil {
		log.Printf("[WhatsApp] Erro ao parsear webhook: %v", err)
		h.writeErrorTwiML(w, "Desculpe, não consegui entender sua mensagem.")
		return
	}

	// Registrar timestamp de recebimento
	msg.ReceivedAt = time.Now()

	// Log da mensagem recebida (sem dados sensíveis)
	log.Printf("[WhatsApp] Mensagem de %s: tipo=%s, mídia=%d",
		maskPhone(msg.From),
		msg.GetMessageType(),
		msg.NumMedia,
	)

	// Processar a mensagem
	response, err := h.service.ProcessMessage(msg)
	if err != nil {
		log.Printf("[WhatsApp] Erro ao processar mensagem: %v", err)
		response = "Desculpe, tive um problema ao processar sua mensagem. Tente novamente."
	}

	// Enviar resposta como TwiML
	h.writeTwiML(w, response)
}

// WebhookVerify é usado pelo Twilio para verificar o webhook
// O Twilio faz um GET para validar que o endpoint existe
//
// Endpoint: GET /api/whatsapp/webhook
func (h *Handler) WebhookVerify(w http.ResponseWriter, r *http.Request) {
	// Simplesmente retornar OK
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Famli WhatsApp Webhook OK"))
}

// =============================================================================
// VINCULAÇÃO DE CONTA
// =============================================================================

// LinkPayload é o payload para vincular um número WhatsApp
type LinkPayload struct {
	// Code é o código de 6 dígitos gerado pelo usuário no WhatsApp
	Code string `json:"code"`

	// PhoneNumber é o número de telefone a ser vinculado
	PhoneNumber string `json:"phone_number"`
}

// Link vincula um número WhatsApp a uma conta Famli
//
// O usuário:
// 1. Digita "vincular" no WhatsApp e recebe um código
// 2. Acessa famli.net/configuracoes
// 3. Digita o código para vincular
//
// Endpoint: POST /api/whatsapp/link
// Autenticação: Requer JWT (usuário logado)
// Body: { "code": "123456", "phone_number": "+5511999999999" }
func (h *Handler) Link(w http.ResponseWriter, r *http.Request) {
	// Obter ID do usuário do contexto (requer autenticação)
	userID := auth.GetUserID(r)
	if userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Faça login para vincular seu WhatsApp")
		return
	}

	// Parsear payload
	var payload LinkPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	// Validar campos
	if payload.PhoneNumber == "" {
		writeJSONError(w, http.StatusBadRequest, "Número de telefone é obrigatório")
		return
	}

	// TODO: Validar código (implementar sistema de códigos com expiração)
	// Por enquanto, aceitamos qualquer código para testes

	// Vincular número ao usuário
	h.service.LinkPhoneToUser(payload.PhoneNumber, userID)

	// Enviar mensagem de confirmação no WhatsApp
	go func() {
		msg := "✅ *WhatsApp vinculado com sucesso!*\n\n" +
			"Agora você pode me enviar:\n" +
			"• Textos para guardar\n" +
			"• Fotos e memórias\n" +
			"• Áudios e documentos\n\n" +
			"_Experimente: me envie algo para guardar!_ 💚"

		if err := h.service.SendMessage(payload.PhoneNumber, msg); err != nil {
			log.Printf("[WhatsApp] Erro ao enviar confirmação de vinculação: %v", err)
		}
	}()

	// Responder sucesso
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "WhatsApp vinculado com sucesso!",
	})
}

// Unlink desvincula o WhatsApp de uma conta Famli
//
// Endpoint: DELETE /api/whatsapp/link
// Autenticação: Requer JWT
func (h *Handler) Unlink(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Faça login")
		return
	}

	// TODO: Implementar desvinculação
	// Por enquanto, apenas retornamos sucesso

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "WhatsApp desvinculado",
	})
}

// =============================================================================
// STATUS DA INTEGRAÇÃO
// =============================================================================

// Status retorna informações sobre a integração WhatsApp
//
// Endpoint: GET /api/whatsapp/status
//
// Resposta:
//
//	{
//	  "enabled": true,
//	  "phone_number": "+14155238886",
//	  "webhook_url": "https://famli.net/api/whatsapp/webhook"
//	}
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"enabled": h.config != nil && h.config.Enabled,
	}

	if h.config != nil && h.config.Enabled {
		// Mostrar apenas parte do número (privacidade)
		status["phone_number"] = maskPhone(h.config.TwilioPhoneNumber)
		status["webhook_url"] = h.config.WebhookBaseURL + "/api/whatsapp/webhook"
	}

	writeJSON(w, http.StatusOK, status)
}

// =============================================================================
// RESPOSTAS TWIML
// =============================================================================

// writeTwiML escreve uma resposta TwiML com a mensagem fornecida
func (h *Handler) writeTwiML(w http.ResponseWriter, message string) {
	response := &TwiMLResponse{Message: message}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response.ToXML()))
}

// writeEmptyTwiML escreve uma resposta TwiML vazia (sem mensagem de resposta)
func (h *Handler) writeEmptyTwiML(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Response></Response>`))
}

// writeErrorTwiML escreve uma resposta TwiML com mensagem de erro
func (h *Handler) writeErrorTwiML(w http.ResponseWriter, message string) {
	h.writeTwiML(w, message)
}

// =============================================================================
// RESPOSTAS JSON
// =============================================================================

// writeJSON escreve uma resposta JSON
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// writeJSONError escreve uma resposta JSON de erro
func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// =============================================================================
// FUNÇÕES AUXILIARES
// =============================================================================

// maskPhone mascara um número de telefone para logs/exibição
// Exemplo: +5511999999999 -> +55119****9999
func maskPhone(phone string) string {
	if len(phone) < 8 {
		return "****"
	}

	// Remover prefixo whatsapp:
	phone = cleanPhoneNumber(phone)

	// Mostrar início e fim, mascarar o meio
	if len(phone) > 8 {
		return phone[:len(phone)-8] + "****" + phone[len(phone)-4:]
	}

	return "****" + phone[len(phone)-4:]
}
