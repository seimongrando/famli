// =============================================================================
// FAMLI - Módulo de Validação e Sanitização
// =============================================================================
// Este módulo fornece funções para validar e sanitizar inputs do usuário.
//
// OWASP A03:2021 – Injection
// - Sanitização de HTML para prevenir XSS
// - Validação de formatos (email, telefone, etc.)
// - Limite de tamanho de inputs
//
// OWASP A07:2021 – Identification and Authentication Failures
// - Validação de força de senha
// - Validação de formato de email
// =============================================================================

package security

import (
	"errors"
	"html"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

// =============================================================================
// ERROS DE VALIDAÇÃO
// =============================================================================

var (
	// ErrInvalidEmail indica email com formato inválido
	ErrInvalidEmail = errors.New("formato de e-mail inválido")

	// ErrWeakPassword indica senha que não atende os requisitos
	ErrWeakPassword = errors.New("senha não atende os requisitos de segurança")

	// ErrInputTooLong indica input que excede o limite
	ErrInputTooLong = errors.New("texto excede o limite permitido")

	// ErrInvalidPhone indica telefone com formato inválido
	ErrInvalidPhone = errors.New("formato de telefone inválido")

	// ErrEmptyInput indica input vazio quando obrigatório
	ErrEmptyInput = errors.New("campo obrigatório não preenchido")

	// ErrInvalidCharacters indica caracteres não permitidos
	ErrInvalidCharacters = errors.New("caracteres não permitidos")
)

// =============================================================================
// LIMITES DE TAMANHO
// =============================================================================

// Limites máximos para diferentes tipos de campos
// Previne ataques de denial of service e buffer overflow
const (
	MaxEmailLength    = 254   // RFC 5321
	MaxPasswordLength = 128   // Limite razoável
	MinPasswordLength = 8     // Mínimo de segurança
	MaxNameLength     = 100   // Nome de usuário
	MaxTitleLength    = 200   // Título de item
	MaxContentLength  = 50000 // Conteúdo de item (50KB)
	MaxPhoneLength    = 20    // Telefone internacional
	MaxURLLength      = 2048  // URL
)

// =============================================================================
// VALIDAÇÃO DE EMAIL
// =============================================================================

// ValidateEmail valida o formato de um endereço de email
//
// Verificações:
// - Formato RFC 5322
// - Tamanho máximo
// - Domínio válido
//
// Parâmetros:
//   - email: endereço de email a ser validado
//
// Retorna:
//   - string: email normalizado (lowercase, trimmed)
//   - error: erro se a validação falhar
func ValidateEmail(email string) (string, error) {
	// Remover espaços
	email = strings.TrimSpace(email)

	// Verificar se está vazio
	if email == "" {
		return "", ErrEmptyInput
	}

	// Verificar tamanho
	if len(email) > MaxEmailLength {
		return "", ErrInputTooLong
	}

	// Validar formato usando net/mail
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", ErrInvalidEmail
	}

	// Extrair apenas o endereço (sem nome)
	email = addr.Address

	// Normalizar para lowercase
	email = strings.ToLower(email)

	// Verificar se tem domínio válido (pelo menos um ponto após @)
	parts := strings.Split(email, "@")
	if len(parts) != 2 || !strings.Contains(parts[1], ".") {
		return "", ErrInvalidEmail
	}

	// Verificar extensão do domínio
	domainParts := strings.Split(parts[1], ".")
	lastPart := domainParts[len(domainParts)-1]
	if len(lastPart) < 2 {
		return "", ErrInvalidEmail
	}

	return email, nil
}

// =============================================================================
// VALIDAÇÃO DE SENHA
// =============================================================================

// PasswordStrength representa a força de uma senha
type PasswordStrength int

const (
	PasswordWeak   PasswordStrength = iota // Não atende requisitos mínimos
	PasswordFair                           // Atende requisitos mínimos
	PasswordGood                           // Boa força
	PasswordStrong                         // Força excelente
)

// PasswordRequirements define os requisitos de senha
type PasswordRequirements struct {
	MinLength        int  // Tamanho mínimo
	RequireUppercase bool // Requer letra maiúscula
	RequireLowercase bool // Requer letra minúscula
	RequireDigit     bool // Requer número
	RequireSpecial   bool // Requer caractere especial
}

// DefaultPasswordRequirements são os requisitos padrão do Famli
// Balanceamos segurança com usabilidade para público 50+
var DefaultPasswordRequirements = PasswordRequirements{
	MinLength:        8,
	RequireUppercase: false, // Flexível para público-alvo
	RequireLowercase: true,
	RequireDigit:     true,
	RequireSpecial:   false, // Flexível para público-alvo
}

// ValidatePassword valida a força de uma senha
//
// Requisitos padrão:
// - Mínimo 8 caracteres
// - Pelo menos uma letra minúscula
// - Pelo menos um número
//
// Parâmetros:
//   - password: senha a ser validada
//
// Retorna:
//   - PasswordStrength: nível de força da senha
//   - error: erro se não atender requisitos mínimos
func ValidatePassword(password string) (PasswordStrength, error) {
	return ValidatePasswordWithRequirements(password, DefaultPasswordRequirements)
}

// ValidatePasswordWithRequirements valida senha com requisitos customizados
func ValidatePasswordWithRequirements(password string, req PasswordRequirements) (PasswordStrength, error) {
	// Verificar tamanho máximo (previne DOS)
	if len(password) > MaxPasswordLength {
		return PasswordWeak, ErrInputTooLong
	}

	// Verificar tamanho mínimo
	if len(password) < req.MinLength {
		return PasswordWeak, ErrWeakPassword
	}

	// Contar tipos de caracteres
	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Verificar requisitos obrigatórios
	if req.RequireUppercase && !hasUpper {
		return PasswordWeak, ErrWeakPassword
	}
	if req.RequireLowercase && !hasLower {
		return PasswordWeak, ErrWeakPassword
	}
	if req.RequireDigit && !hasDigit {
		return PasswordWeak, ErrWeakPassword
	}
	if req.RequireSpecial && !hasSpecial {
		return PasswordWeak, ErrWeakPassword
	}

	// Calcular força
	score := 0
	if hasUpper {
		score++
	}
	if hasLower {
		score++
	}
	if hasDigit {
		score++
	}
	if hasSpecial {
		score++
	}
	if len(password) >= 12 {
		score++
	}
	if len(password) >= 16 {
		score++
	}

	switch {
	case score >= 5:
		return PasswordStrong, nil
	case score >= 3:
		return PasswordGood, nil
	case score >= 2:
		return PasswordFair, nil
	default:
		return PasswordWeak, ErrWeakPassword
	}
}

// =============================================================================
// SANITIZAÇÃO DE TEXTO
// =============================================================================

// SanitizeText sanitiza texto para prevenir XSS e injection
//
// Operações:
// - Escape de HTML entities
// - Remoção de caracteres de controle
// - Trim de espaços
// - Limite de tamanho
//
// Parâmetros:
//   - text: texto a ser sanitizado
//   - maxLength: tamanho máximo permitido (0 = sem limite)
//
// Retorna:
//   - string: texto sanitizado
func SanitizeText(text string, maxLength int) string {
	// Remover caracteres de controle (exceto newline e tab)
	text = removeControlChars(text)

	// Escape HTML para prevenir XSS
	text = html.EscapeString(text)

	// Trim espaços
	text = strings.TrimSpace(text)

	// Aplicar limite de tamanho
	if maxLength > 0 && len(text) > maxLength {
		// Cortar em boundary de rune (não quebrar caracteres UTF-8)
		text = truncateString(text, maxLength)
	}

	return text
}

// SanitizeName sanitiza nomes de usuário
func SanitizeName(name string) string {
	name = SanitizeText(name, MaxNameLength)

	// Remover caracteres especiais exceto espaço, hífen e apóstrofo
	// Permite: "Maria José", "O'Connor", "Jean-Pierre"
	reg := regexp.MustCompile(`[^a-zA-ZÀ-ÿ\s\-']`)
	name = reg.ReplaceAllString(name, "")

	// Normalizar espaços múltiplos
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")

	return strings.TrimSpace(name)
}

// SanitizeTitle sanitiza títulos de itens
func SanitizeTitle(title string) string {
	return SanitizeText(title, MaxTitleLength)
}

// SanitizeContent sanitiza conteúdo de itens
func SanitizeContent(content string) string {
	return SanitizeText(content, MaxContentLength)
}

// =============================================================================
// VALIDAÇÃO DE TELEFONE
// =============================================================================

// ValidatePhone valida formato de telefone
//
// Aceita formatos:
// - +5511999999999 (internacional)
// - 11999999999 (nacional)
// - (11) 99999-9999 (formatado)
//
// Parâmetros:
//   - phone: número de telefone
//
// Retorna:
//   - string: telefone normalizado (apenas dígitos com +)
//   - error: erro se o formato for inválido
func ValidatePhone(phone string) (string, error) {
	if phone == "" {
		return "", nil // Telefone é opcional
	}

	// Verificar tamanho
	if len(phone) > MaxPhoneLength {
		return "", ErrInputTooLong
	}

	// Remover caracteres não numéricos (exceto +)
	normalized := ""
	for i, char := range phone {
		if char == '+' && i == 0 {
			normalized += string(char)
		} else if unicode.IsDigit(char) {
			normalized += string(char)
		}
	}

	// Verificar tamanho mínimo (DDD + número)
	digitsOnly := strings.TrimPrefix(normalized, "+")
	if len(digitsOnly) < 10 {
		return "", ErrInvalidPhone
	}

	// Adicionar código do Brasil se não tiver código de país
	if !strings.HasPrefix(normalized, "+") {
		normalized = "+55" + normalized
	}

	return normalized, nil
}

// =============================================================================
// VALIDAÇÃO DE URL
// =============================================================================

// ValidateURL valida e sanitiza URLs
// Previne SSRF verificando se a URL é segura
func ValidateURL(urlStr string) (string, error) {
	if urlStr == "" {
		return "", nil
	}

	// Verificar tamanho
	if len(urlStr) > MaxURLLength {
		return "", ErrInputTooLong
	}

	// Trim espaços
	urlStr = strings.TrimSpace(urlStr)

	// Verificar protocolo permitido
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return "", errors.New("URL deve começar com http:// ou https://")
	}

	// Bloquear IPs privados (previne SSRF)
	privatePatterns := []string{
		"://localhost",
		"://127.",
		"://10.",
		"://192.168.",
		"://172.16.", "://172.17.", "://172.18.", "://172.19.",
		"://172.20.", "://172.21.", "://172.22.", "://172.23.",
		"://172.24.", "://172.25.", "://172.26.", "://172.27.",
		"://172.28.", "://172.29.", "://172.30.", "://172.31.",
		"://0.0.0.0",
		"://[::1]",
	}

	for _, pattern := range privatePatterns {
		if strings.Contains(urlStr, pattern) {
			return "", errors.New("URL não permitida")
		}
	}

	return urlStr, nil
}

// =============================================================================
// FUNÇÕES AUXILIARES
// =============================================================================

// removeControlChars remove caracteres de controle exceto newline e tab
func removeControlChars(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r' {
			return -1
		}
		return r
	}, s)
}

// truncateString trunca string respeitando boundaries de rune
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Encontrar último boundary de rune antes do limite
	truncated := s[:maxLen]
	for len(truncated) > 0 {
		r, size := []rune(truncated)[len([]rune(truncated))-1], 1
		if r != '?' { // Se não é um caractere inválido, está ok
			break
		}
		truncated = truncated[:len(truncated)-size]
	}

	return truncated
}

// ContainsSQLInjection verifica padrões comuns de SQL injection
// NOTA: Isto é uma camada adicional, não substitui prepared statements
func ContainsSQLInjection(input string) bool {
	patterns := []string{
		"(?i)(union\\s+select)",
		"(?i)(select\\s+.*\\s+from)",
		"(?i)(insert\\s+into)",
		"(?i)(delete\\s+from)",
		"(?i)(drop\\s+table)",
		"(?i)(update\\s+.*\\s+set)",
		"(?i)('\\s*or\\s+')",
		"(?i)(--)",
		"(?i)(/\\*.*\\*/)",
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}

	return false
}
