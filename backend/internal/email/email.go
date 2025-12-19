// =============================================================================
// FAMLI - Serviço de Email
// =============================================================================
// Componente genérico para envio de emails. Suporta múltiplos provedores
// através de uma interface comum, facilitando a troca no futuro.
//
// Provedores suportados:
// - Mailtrap (padrão para MVP)
// - SendGrid (futuro)
// - AWS SES (futuro)
// - SMTP genérico (futuro)
//
// Variáveis de ambiente:
// - EMAIL_PROVIDER: "mailtrap" (padrão), "sendgrid", "ses", "smtp"
// - MAILTRAP_API_TOKEN: Token da API do Mailtrap
// - MAILTRAP_SANDBOX: "true" para usar sandbox (testes), "false" para produção
// - MAILTRAP_INBOX_ID: ID da inbox (obrigatório para sandbox)
// - EMAIL_FROM: Email remetente (ex: noreply@famli.me)
// - EMAIL_FROM_NAME: Nome do remetente (ex: Famli)
// =============================================================================

package email

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// =============================================================================
// INTERFACE
// =============================================================================

// Provider define a interface para provedores de email
type Provider interface {
	Send(email *Email) error
	Name() string
}

// Email representa um email a ser enviado
type Email struct {
	To       string            // Destinatário
	ToName   string            // Nome do destinatário
	Subject  string            // Assunto
	HTML     string            // Corpo HTML
	Text     string            // Corpo texto (fallback)
	Metadata map[string]string // Metadados opcionais
}

// Service gerencia o envio de emails
type Service struct {
	provider Provider
	from     string
	fromName string
}

// =============================================================================
// SERVICE
// =============================================================================

// NewService cria uma nova instância do serviço de email
func NewService() *Service {
	providerName := os.Getenv("EMAIL_PROVIDER")
	if providerName == "" {
		providerName = "mailtrap"
	}

	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		from = "noreply@famli.me"
	}

	fromName := os.Getenv("EMAIL_FROM_NAME")
	if fromName == "" {
		fromName = "Famli"
	}

	var provider Provider
	switch providerName {
	case "mailtrap":
		provider = NewMailtrapProvider()
	// Adicionar outros provedores no futuro:
	// case "sendgrid":
	//     provider = NewSendGridProvider()
	// case "ses":
	//     provider = NewSESProvider()
	default:
		provider = NewMailtrapProvider()
	}

	return &Service{
		provider: provider,
		from:     from,
		fromName: fromName,
	}
}

// IsConfigured retorna se o serviço está configurado
func (s *Service) IsConfigured() bool {
	return s.provider != nil
}

// GetProviderName retorna o nome do provedor atual
func (s *Service) GetProviderName() string {
	if s.provider == nil {
		return "none"
	}
	return s.provider.Name()
}

// Send envia um email
func (s *Service) Send(email *Email) error {
	if s.provider == nil {
		return fmt.Errorf("email provider not configured")
	}
	return s.provider.Send(email)
}

// =============================================================================
// TEMPLATES DE EMAIL
// =============================================================================

// SendPasswordReset envia email de recuperação de senha
// locale: idioma do usuário ("pt-BR", "en", etc.)
func (s *Service) SendPasswordReset(to, toName, resetLink, locale string) error {
	var subject, html, text string

	// Logo SVG inline (compatível com email)
	logo := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 92 81" width="80" height="70">
		<path d="M0 13C0 5.82 5.82 0 13 0H55C62.18 0 68 5.82 68 13V49C68 56.18 62.18 62 55 62H40L34 75L28 62H13C5.82 62 0 56.18 0 49V13Z" fill="#355d4a"/>
		<path d="M34 52C34 52 52.5 38.5 52.5 26C52.5 20 48 15 42 15C37.5 15 34 18 34 18C34 18 30.5 15 26 15C20 15 15.5 20 15.5 26C15.5 38.5 34 52 34 52Z" fill="#f4a285"/>
	</svg>`

	if strings.HasPrefix(locale, "en") {
		// Template em Inglês
		subject = "🔐 Reset your password - Famli"
		html = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Password - Famli</title>
    <link href="https://fonts.googleapis.com/css2?family=Nunito:wght@400;600;700&display=swap" rel="stylesheet">
</head>
<body style="margin: 0; padding: 0; font-family: 'Nunito', -apple-system, BlinkMacSystemFont, sans-serif; background-color: #faf8f5;">
    <table width="100%%" cellpadding="0" cellspacing="0" style="max-width: 600px; margin: 0 auto; padding: 24px;">
        <tr>
            <td style="background: #2d5a47; padding: 40px; text-align: center; border-radius: 20px 20px 0 0;">
                <div style="margin-bottom: 16px;">%s</div>
                <h1 style="color: white; margin: 0; font-size: 32px; font-weight: 700;">famli</h1>
                <p style="color: rgba(255,255,255,0.85); margin: 8px 0 0; font-size: 16px;">Organizing what matters, with care.</p>
            </td>
        </tr>
        <tr>
            <td style="background: white; padding: 40px; border-radius: 0 0 20px 20px; box-shadow: 0 4px 24px rgba(44, 42, 38, 0.08);">
                <h2 style="color: #2c2a26; margin: 0 0 20px; font-size: 24px; font-weight: 600;">Hello%s!</h2>
                
                <p style="color: #5c584f; font-size: 17px; line-height: 1.6;">
                    We received a request to reset your Famli account password.
                </p>
                
                <p style="color: #5c584f; font-size: 17px; line-height: 1.6;">
                    Click the button below to create a new password:
                </p>
                
                <div style="text-align: center; margin: 32px 0;">
                    <a href="%s" style="display: inline-block; background: #e07b39; color: white; padding: 16px 36px; text-decoration: none; border-radius: 12px; font-weight: 700; font-size: 17px;">
                        Reset My Password
                    </a>
                </div>
                
                <p style="color: #6b665c; font-size: 15px; line-height: 1.6;">
                    This link expires in <strong>1 hour</strong>. If you didn't request a password reset, please ignore this email.
                </p>
                
                <p style="color: #6b665c; font-size: 14px; line-height: 1.6;">
                    If the button doesn't work, copy and paste this link in your browser:<br>
                    <a href="%s" style="color: #2d5a47; word-break: break-all;">%s</a>
                </p>
                
                <hr style="border: none; border-top: 1px solid #e5ddd0; margin: 32px 0;">
                
                <p style="color: #6b665c; font-size: 13px; text-align: center;">
                    This email was sent by Famli. If you don't have an account, please ignore this message.
                </p>
            </td>
        </tr>
    </table>
</body>
</html>
`, logo, getNameGreeting(toName), resetLink, resetLink, resetLink)

		text = fmt.Sprintf(`
Hello%s!

We received a request to reset your Famli account password.

Click the link below to create a new password:
%s

This link expires in 1 hour. If you didn't request a password reset, please ignore this email.

--
Famli - Organizing what matters, with care.
`, getNameGreeting(toName), resetLink)

	} else {
		// Template em Português (padrão)
		subject = "🔐 Redefinir sua senha - Famli"
		html = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Redefinir Senha - Famli</title>
    <link href="https://fonts.googleapis.com/css2?family=Nunito:wght@400;600;700&display=swap" rel="stylesheet">
</head>
<body style="margin: 0; padding: 0; font-family: 'Nunito', -apple-system, BlinkMacSystemFont, sans-serif; background-color: #faf8f5;">
    <table width="100%%" cellpadding="0" cellspacing="0" style="max-width: 600px; margin: 0 auto; padding: 24px;">
        <tr>
            <td style="background: #2d5a47; padding: 40px; text-align: center; border-radius: 20px 20px 0 0;">
                <div style="margin-bottom: 16px;">%s</div>
                <h1 style="color: white; margin: 0; font-size: 32px; font-weight: 700;">famli</h1>
                <p style="color: rgba(255,255,255,0.85); margin: 8px 0 0; font-size: 16px;">Organizando o que importa, com carinho.</p>
            </td>
        </tr>
        <tr>
            <td style="background: white; padding: 40px; border-radius: 0 0 20px 20px; box-shadow: 0 4px 24px rgba(44, 42, 38, 0.08);">
                <h2 style="color: #2c2a26; margin: 0 0 20px; font-size: 24px; font-weight: 600;">Olá%s!</h2>
                
                <p style="color: #5c584f; font-size: 17px; line-height: 1.6;">
                    Recebemos uma solicitação para redefinir a senha da sua conta Famli.
                </p>
                
                <p style="color: #5c584f; font-size: 17px; line-height: 1.6;">
                    Clique no botão abaixo para criar uma nova senha:
                </p>
                
                <div style="text-align: center; margin: 32px 0;">
                    <a href="%s" style="display: inline-block; background: #e07b39; color: white; padding: 16px 36px; text-decoration: none; border-radius: 12px; font-weight: 700; font-size: 17px;">
                        Redefinir Minha Senha
                    </a>
                </div>
                
                <p style="color: #6b665c; font-size: 15px; line-height: 1.6;">
                    Este link expira em <strong>1 hora</strong>. Se você não solicitou a redefinição de senha, ignore este email.
                </p>
                
                <p style="color: #6b665c; font-size: 14px; line-height: 1.6;">
                    Se o botão não funcionar, copie e cole este link no seu navegador:<br>
                    <a href="%s" style="color: #2d5a47; word-break: break-all;">%s</a>
                </p>
                
                <hr style="border: none; border-top: 1px solid #e5ddd0; margin: 32px 0;">
                
                <p style="color: #6b665c; font-size: 13px; text-align: center;">
                    Este email foi enviado pelo Famli. Se você não tem uma conta, por favor ignore esta mensagem.
                </p>
            </td>
        </tr>
    </table>
</body>
</html>
`, logo, getNameGreeting(toName), resetLink, resetLink, resetLink)

		text = fmt.Sprintf(`
Olá%s!

Recebemos uma solicitação para redefinir a senha da sua conta Famli.

Clique no link abaixo para criar uma nova senha:
%s

Este link expira em 1 hora. Se você não solicitou a redefinição de senha, ignore este email.

--
Famli - Organizando o que importa, com carinho.
`, getNameGreeting(toName), resetLink)
	}

	return s.Send(&Email{
		To:      to,
		ToName:  toName,
		Subject: subject,
		HTML:    html,
		Text:    text,
	})
}

// SendWelcome envia email de boas-vindas
// SendWelcome envia email de boas-vindas
// locale: idioma do usuário ("pt-BR", "en", etc.)
func (s *Service) SendWelcome(to, toName, locale string) error {
	var subject, html, text string

	// Logo SVG inline
	logo := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 92 81" width="80" height="70">
		<path d="M0 13C0 5.82 5.82 0 13 0H55C62.18 0 68 5.82 68 13V49C68 56.18 62.18 62 55 62H40L34 75L28 62H13C5.82 62 0 56.18 0 49V13Z" fill="#355d4a"/>
		<path d="M34 52C34 52 52.5 38.5 52.5 26C52.5 20 48 15 42 15C37.5 15 34 18 34 18C34 18 30.5 15 26 15C20 15 15.5 20 15.5 26C15.5 38.5 34 52 34 52Z" fill="#f4a285"/>
	</svg>`

	if strings.HasPrefix(locale, "en") {
		subject = "🏠 Welcome to Famli!"
		html = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="https://fonts.googleapis.com/css2?family=Nunito:wght@400;600;700&display=swap" rel="stylesheet">
</head>
<body style="margin: 0; padding: 0; font-family: 'Nunito', -apple-system, BlinkMacSystemFont, sans-serif; background-color: #faf8f5;">
    <table width="100%%" cellpadding="0" cellspacing="0" style="max-width: 600px; margin: 0 auto; padding: 24px;">
        <tr>
            <td style="background: #2d5a47; padding: 40px; text-align: center; border-radius: 20px 20px 0 0;">
                <div style="margin-bottom: 16px;">%s</div>
                <h1 style="color: white; margin: 0; font-size: 32px; font-weight: 700;">Welcome to famli!</h1>
            </td>
        </tr>
        <tr>
            <td style="background: white; padding: 40px; border-radius: 0 0 20px 20px; box-shadow: 0 4px 24px rgba(44, 42, 38, 0.08);">
                <h2 style="color: #2c2a26; margin: 0 0 20px; font-size: 24px;">Hello%s! 👋</h2>
                
                <p style="color: #5c584f; font-size: 17px; line-height: 1.6;">
                    Your account was created successfully! Famli is the place to organize memories, documents and guidance for your loved ones.
                </p>
                
                <p style="color: #5c584f; font-size: 17px; line-height: 1.6;">
                    Start by adding your first information - it can be something simple like an emergency contact or a special memory.
                </p>
                
                <div style="text-align: center; margin: 32px 0;">
                    <a href="https://famli.me/my-box" style="display: inline-block; background: #e07b39; color: white; padding: 16px 36px; text-decoration: none; border-radius: 12px; font-weight: 700; font-size: 17px;">
                        Access My Box
                    </a>
                </div>
                
                <p style="color: #6b665c; font-size: 15px;">
                    With care,<br>
                    <strong style="color: #2d5a47;">The Famli Team</strong>
                </p>
            </td>
        </tr>
    </table>
</body>
</html>
`, logo, getNameGreeting(toName))
		text = fmt.Sprintf("Hello%s! Your Famli account was created successfully. Access: https://famli.me/my-box", getNameGreeting(toName))
	} else {
		subject = "🏠 Bem-vindo ao Famli!"
		html = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="https://fonts.googleapis.com/css2?family=Nunito:wght@400;600;700&display=swap" rel="stylesheet">
</head>
<body style="margin: 0; padding: 0; font-family: 'Nunito', -apple-system, BlinkMacSystemFont, sans-serif; background-color: #faf8f5;">
    <table width="100%%" cellpadding="0" cellspacing="0" style="max-width: 600px; margin: 0 auto; padding: 24px;">
        <tr>
            <td style="background: #2d5a47; padding: 40px; text-align: center; border-radius: 20px 20px 0 0;">
                <div style="margin-bottom: 16px;">%s</div>
                <h1 style="color: white; margin: 0; font-size: 32px; font-weight: 700;">Bem-vindo ao famli!</h1>
            </td>
        </tr>
        <tr>
            <td style="background: white; padding: 40px; border-radius: 0 0 20px 20px; box-shadow: 0 4px 24px rgba(44, 42, 38, 0.08);">
                <h2 style="color: #2c2a26; margin: 0 0 20px; font-size: 24px;">Olá%s! 👋</h2>
                
                <p style="color: #5c584f; font-size: 17px; line-height: 1.6;">
                    Sua conta foi criada com sucesso! O Famli é o lugar para organizar memórias, documentos e orientações para seus entes queridos.
                </p>
                
                <p style="color: #5c584f; font-size: 17px; line-height: 1.6;">
                    Comece adicionando suas primeiras informações - pode ser algo simples como um contato de emergência ou uma memória especial.
                </p>
                
                <div style="text-align: center; margin: 32px 0;">
                    <a href="https://famli.me/minha-caixa" style="display: inline-block; background: #e07b39; color: white; padding: 16px 36px; text-decoration: none; border-radius: 12px; font-weight: 700; font-size: 17px;">
                        Acessar Minha Caixa
                    </a>
                </div>
                
                <p style="color: #6b665c; font-size: 15px;">
                    Com carinho,<br>
                    <strong style="color: #2d5a47;">Equipe Famli</strong>
                </p>
            </td>
        </tr>
    </table>
</body>
</html>
`, logo, getNameGreeting(toName))
		text = fmt.Sprintf("Olá%s! Sua conta Famli foi criada com sucesso. Acesse: https://famli.me/minha-caixa", getNameGreeting(toName))
	}

	return s.Send(&Email{
		To:      to,
		ToName:  toName,
		Subject: subject,
		HTML:    html,
		Text:    text,
	})
}

// =============================================================================
// MAILTRAP PROVIDER
// =============================================================================

// URLs da API Mailtrap
const (
	MailtrapSandboxBaseURL = "https://sandbox.api.mailtrap.io/api/send"
	MailtrapProductionURL  = "https://send.api.mailtrap.io/api/send"
)

// MailtrapProvider implementa o envio via Mailtrap
type MailtrapProvider struct {
	apiToken  string
	apiURL    string // URL da API (sandbox ou produção)
	inboxID   string // Inbox ID (apenas para sandbox)
	from      string
	fromName  string
	isSandbox bool
	client    *http.Client
}

// NewMailtrapProvider cria um novo provider Mailtrap
func NewMailtrapProvider() *MailtrapProvider {
	// Verifica se deve usar sandbox (default: false para produção)
	sandbox := os.Getenv("MAILTRAP_SANDBOX")
	isSandbox := sandbox == "true" || sandbox == "1"
	inboxID := os.Getenv("MAILTRAP_INBOX_ID") // Necessário apenas para sandbox

	apiURL := MailtrapProductionURL
	if isSandbox {
		// Sandbox requer inbox_id na URL
		if inboxID != "" {
			apiURL = fmt.Sprintf("%s/%s", MailtrapSandboxBaseURL, inboxID)
		} else {
			apiURL = MailtrapSandboxBaseURL // Vai falhar se inbox_id não for fornecido
		}
	}

	return &MailtrapProvider{
		apiToken:  os.Getenv("MAILTRAP_API_TOKEN"),
		apiURL:    apiURL,
		inboxID:   inboxID,
		from:      os.Getenv("EMAIL_FROM"),
		fromName:  os.Getenv("EMAIL_FROM_NAME"),
		isSandbox: isSandbox,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name retorna o nome do provider
func (p *MailtrapProvider) Name() string {
	if p.isSandbox {
		return "mailtrap-sandbox"
	}
	return "mailtrap"
}

// IsSandbox retorna se está usando o ambiente de sandbox
func (p *MailtrapProvider) IsSandbox() bool {
	return p.isSandbox
}

// Send envia um email via Mailtrap
func (p *MailtrapProvider) Send(email *Email) error {
	if p.apiToken == "" {
		return fmt.Errorf("MAILTRAP_API_TOKEN not configured")
	}

	// Sandbox requer inbox_id
	if p.isSandbox && p.inboxID == "" {
		return fmt.Errorf("MAILTRAP_INBOX_ID is required for sandbox mode")
	}

	// Gerar Message-ID único para evitar filtros de spam
	messageID := fmt.Sprintf("<%d.%s@famli.me>", time.Now().UnixNano(), generateRandomID(12))

	// Payload da API Mailtrap
	payload := map[string]interface{}{
		"from": map[string]string{
			"email": p.from,
			"name":  p.fromName,
		},
		"to": []map[string]string{
			{
				"email": email.To,
				"name":  email.ToName,
			},
		},
		"subject": email.Subject,
		"html":    email.HTML,
		"text":    email.Text,
		"headers": map[string]string{
			"Message-ID": messageID,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling email: %w", err)
	}

	// Criar request (usa URL configurada: sandbox ou produção)
	req, err := http.NewRequest("POST", p.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiToken)

	// Enviar
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}
	defer resp.Body.Close()

	// Verificar resposta
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailtrap error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// =============================================================================
// HELPERS
// =============================================================================

// generateRandomID gera um ID aleatório para Message-ID
func generateRandomID(length int) string {
	b := make([]byte, length/2+1)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}

func getNameGreeting(name string) string {
	if name == "" {
		return ""
	}
	return " " + name
}

// RenderTemplate renderiza um template HTML com dados
func RenderTemplate(tmpl string, data interface{}) (string, error) {
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
