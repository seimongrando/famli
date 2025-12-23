// =============================================================================
// FAMLI - Cliente Twilio para WhatsApp
// =============================================================================
// Este arquivo contém o cliente para comunicação com a API do Twilio.
// O Twilio atua como intermediário entre o Famli e o WhatsApp.
//
// Configuração necessária:
// 1. Criar conta em twilio.com
// 2. Ativar WhatsApp Sandbox (para testes)
// 3. Obter Account SID e Auth Token
// 4. Configurar webhook apontando para /api/whatsapp/webhook
//
// Documentação:
// - https://www.twilio.com/docs/whatsapp
// - https://www.twilio.com/docs/whatsapp/sandbox
// =============================================================================

package whatsapp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// =============================================================================
// CLIENTE TWILIO
// =============================================================================

// TwilioClient é o cliente para comunicação com a API do Twilio
type TwilioClient struct {
	// accountSid é o identificador único da conta Twilio
	// Encontrado em: twilio.com/console
	accountSid string

	// authToken é o token de autenticação
	// Encontrado em: twilio.com/console
	authToken string

	// fromNumber é o número do WhatsApp Twilio (sandbox ou verificado)
	// Formato: whatsapp:+14155238886 (sandbox) ou whatsapp:+5511999999999
	fromNumber string

	// httpClient é o cliente HTTP para fazer requisições
	httpClient *http.Client
}

// NewTwilioClient cria uma nova instância do cliente Twilio
//
// Parâmetros:
//   - accountSid: SID da conta Twilio
//   - authToken: Token de autenticação
//   - fromNumber: Número WhatsApp do Twilio (com prefixo whatsapp:)
//
// Retorna:
//   - *TwilioClient: cliente configurado
func NewTwilioClient(accountSid, authToken, fromNumber string) *TwilioClient {
	return &TwilioClient{
		accountSid: accountSid,
		authToken:  authToken,
		fromNumber: fromNumber,
		httpClient: &http.Client{},
	}
}

// =============================================================================
// ENVIO DE MENSAGENS
// =============================================================================

// SendMessage envia uma mensagem de texto para um número WhatsApp
//
// Parâmetros:
//   - to: número de destino (formato: +5511999999999 ou whatsapp:+5511999999999)
//   - body: texto da mensagem
//
// Retorna:
//   - error: erro se houver falha no envio
func (c *TwilioClient) SendMessage(to, body string) error {
	// Garantir formato correto do número
	if !strings.HasPrefix(to, "whatsapp:") {
		to = "whatsapp:" + to
	}

	// URL da API do Twilio
	apiURL := fmt.Sprintf(
		"https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json",
		c.accountSid,
	)

	// Dados do formulário
	data := url.Values{}
	data.Set("To", to)
	data.Set("From", c.fromNumber)
	data.Set("Body", body)

	// Criar requisição
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %w", err)
	}

	// Headers
	req.SetBasicAuth(c.accountSid, c.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Enviar requisição
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar mensagem: %w", err)
	}
	defer resp.Body.Close()

	// Verificar resposta
	if resp.StatusCode >= 400 {
		_, _ = io.ReadAll(resp.Body)
		log.Printf("[Twilio] Erro na API: status=%d", resp.StatusCode)
		return fmt.Errorf("erro da API Twilio: status %d", resp.StatusCode)
	}

	log.Printf("[Twilio] Mensagem enviada para %s", maskPhone(to))
	return nil
}

// SendMessageWithMedia envia uma mensagem com mídia anexada
//
// Parâmetros:
//   - to: número de destino
//   - body: texto da mensagem
//   - mediaURL: URL pública da mídia (imagem, áudio, documento)
//
// Retorna:
//   - error: erro se houver falha no envio
func (c *TwilioClient) SendMessageWithMedia(to, body, mediaURL string) error {
	if !strings.HasPrefix(to, "whatsapp:") {
		to = "whatsapp:" + to
	}

	apiURL := fmt.Sprintf(
		"https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json",
		c.accountSid,
	)

	data := url.Values{}
	data.Set("To", to)
	data.Set("From", c.fromNumber)
	data.Set("Body", body)
	data.Set("MediaUrl", mediaURL)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.SetBasicAuth(c.accountSid, c.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar mensagem: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		_, _ = io.ReadAll(resp.Body)
		log.Printf("[Twilio] Erro na API: status=%d", resp.StatusCode)
		return fmt.Errorf("erro da API Twilio: status %d", resp.StatusCode)
	}

	log.Printf("[Twilio] Mensagem com mídia enviada para %s", maskPhone(to))
	return nil
}

// =============================================================================
// VALIDAÇÃO DE WEBHOOK
// =============================================================================

// ValidateWebhookSignature valida a assinatura de um webhook do Twilio
// Isso garante que a requisição realmente veio do Twilio
//
// Parâmetros:
//   - signature: valor do header X-Twilio-Signature
//   - url: URL completa do webhook
//   - params: parâmetros do POST
//
// Retorna:
//   - bool: true se a assinatura é válida
//
// Nota: Para simplificar o MVP, esta validação está desabilitada.
// Em produção, SEMPRE valide a assinatura!
// Documentação: https://www.twilio.com/docs/usage/security
func (c *TwilioClient) ValidateWebhookSignature(signature, webhookURL string, params map[string]string) bool {
	// TODO: Implementar validação real de assinatura HMAC-SHA1
	// Para o MVP, retornamos true sempre
	// Em produção, usar a biblioteca oficial do Twilio para validação
	return true
}

// =============================================================================
// PARSING DE WEBHOOK
// =============================================================================

// ParseWebhookRequest converte uma requisição HTTP do webhook em IncomingMessage
//
// O Twilio envia os dados como application/x-www-form-urlencoded
// com os seguintes campos principais:
//   - MessageSid: ID único da mensagem
//   - From: número do remetente (whatsapp:+5511...)
//   - To: número do destinatário (nosso número)
//   - Body: conteúdo da mensagem
//   - NumMedia: quantidade de arquivos anexados
//   - MediaUrl0, MediaContentType0: dados da primeira mídia
//
// Parâmetros:
//   - r: requisição HTTP do webhook
//
// Retorna:
//   - *IncomingMessage: mensagem parseada
//   - error: erro se houver falha no parsing
func ParseWebhookRequest(r *http.Request) (*IncomingMessage, error) {
	// Parse do formulário
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("erro ao parsear formulário: %w", err)
	}

	// Extrair número de mídias
	numMedia := 0
	if nm := r.FormValue("NumMedia"); nm != "" {
		fmt.Sscanf(nm, "%d", &numMedia)
	}

	msg := &IncomingMessage{
		MessageSid:  r.FormValue("MessageSid"),
		From:        r.FormValue("From"),
		To:          r.FormValue("To"),
		Body:        r.FormValue("Body"),
		NumMedia:    numMedia,
		ProfileName: r.FormValue("ProfileName"),
	}

	// Se tem mídia, pegar a primeira
	if numMedia > 0 {
		msg.MediaUrl = r.FormValue("MediaUrl0")
		msg.MediaContentType = r.FormValue("MediaContentType0")
	}

	// Localização (se enviada)
	msg.Latitude = r.FormValue("Latitude")
	msg.Longitude = r.FormValue("Longitude")

	return msg, nil
}

// =============================================================================
// RESPOSTAS DO WEBHOOK
// =============================================================================

// TwiMLResponse representa uma resposta TwiML para o webhook
// TwiML é a linguagem XML do Twilio para instruções
type TwiMLResponse struct {
	Message string
}

// ToXML converte a resposta para formato TwiML XML
//
// Exemplo de saída:
//
//	<?xml version="1.0" encoding="UTF-8"?>
//	<Response>
//	    <Message>Sua mensagem aqui</Message>
//	</Response>
func (t *TwiMLResponse) ToXML() string {
	// Escapar caracteres especiais XML
	message := escapeXML(t.Message)

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Message>%s</Message>
</Response>`, message)
}

// ToJSON converte a resposta para formato JSON (alternativa ao TwiML)
func (t *TwiMLResponse) ToJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"message": t.Message,
	})
}

// escapeXML escapa caracteres especiais para XML
func escapeXML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(s)
}
