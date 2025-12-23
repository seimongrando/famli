// =============================================================================
// FAMLI - Auditoria e Logging de Segurança
// =============================================================================
// Este módulo implementa logging de eventos de segurança para:
//
// OWASP A09:2021 – Security Logging and Monitoring Failures
// - Registrar tentativas de autenticação
// - Detectar padrões suspeitos
// - Fornecer trilha de auditoria
//
// Eventos registrados:
// - Login bem-sucedido/falho
// - Criação de conta
// - Alteração de senha
// - Acesso a dados sensíveis
// - Tentativas de acesso não autorizado
// - Rate limit excedido
// =============================================================================

package security

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// TIPOS DE EVENTO
// =============================================================================

// AuditEventType define os tipos de eventos de auditoria
type AuditEventType string

const (
	// Autenticação
	EventLoginSuccess   AuditEventType = "LOGIN_SUCCESS"
	EventLoginFailed    AuditEventType = "LOGIN_FAILED"
	EventLogout         AuditEventType = "LOGOUT"
	EventRegister       AuditEventType = "REGISTER"
	EventPasswordChange AuditEventType = "PASSWORD_CHANGE"
	EventPasswordReset  AuditEventType = "PASSWORD_RESET"
	EventSessionExpired AuditEventType = "SESSION_EXPIRED"

	// Acesso a dados
	EventDataAccess      AuditEventType = "DATA_ACCESS"
	EventDataCreate      AuditEventType = "DATA_CREATE"
	EventDataUpdate      AuditEventType = "DATA_UPDATE"
	EventDataDelete      AuditEventType = "DATA_DELETE"
	EventSensitiveAccess AuditEventType = "SENSITIVE_ACCESS"

	// Segurança
	EventRateLimitExceeded  AuditEventType = "RATE_LIMIT_EXCEEDED"
	EventUnauthorizedAccess AuditEventType = "UNAUTHORIZED_ACCESS"
	EventSuspiciousActivity AuditEventType = "SUSPICIOUS_ACTIVITY"
	EventTokenInvalid       AuditEventType = "TOKEN_INVALID"

	// WhatsApp
	EventWhatsAppLink            AuditEventType = "WHATSAPP_LINK"
	EventWhatsAppUnlink          AuditEventType = "WHATSAPP_UNLINK"
	EventWhatsAppMessageReceived AuditEventType = "WHATSAPP_MESSAGE_RECEIVED"
	EventWhatsAppWebhookReceived AuditEventType = "WHATSAPP_WEBHOOK_RECEIVED"

	// LGPD - Direitos do Titular
	EventAccountDeletion AuditEventType = "ACCOUNT_DELETION" // Direito ao esquecimento
	EventDataExport      AuditEventType = "DATA_EXPORT"      // Direito à portabilidade
)

// AuditSeverity define a severidade do evento
type AuditSeverity string

const (
	SeverityInfo     AuditSeverity = "INFO"
	SeverityWarning  AuditSeverity = "WARNING"
	SeverityError    AuditSeverity = "ERROR"
	SeverityCritical AuditSeverity = "CRITICAL"
)

// =============================================================================
// EVENTO DE AUDITORIA
// =============================================================================

// AuditEvent representa um evento de auditoria
type AuditEvent struct {
	// ID único do evento
	ID string `json:"id"`

	// Timestamp do evento
	Timestamp time.Time `json:"timestamp"`

	// Tipo do evento
	Type AuditEventType `json:"type"`

	// Severidade
	Severity AuditSeverity `json:"severity"`

	// ID do usuário (se aplicável)
	UserID string `json:"user_id,omitempty"`

	// IP do cliente
	ClientIP string `json:"client_ip"`

	// User-Agent do cliente
	UserAgent string `json:"user_agent,omitempty"`

	// Recurso acessado
	Resource string `json:"resource,omitempty"`

	// Ação realizada
	Action string `json:"action,omitempty"`

	// Resultado (success, failure, blocked)
	Result string `json:"result"`

	// Detalhes adicionais (não incluir dados sensíveis!)
	Details map[string]interface{} `json:"details,omitempty"`

	// Request ID para correlação
	RequestID string `json:"request_id,omitempty"`
}

// =============================================================================
// LOGGER DE AUDITORIA
// =============================================================================

// AuditLogger gerencia o logging de eventos de segurança
type AuditLogger struct {
	// events armazena eventos recentes para análise
	events []AuditEvent

	// maxEvents é o número máximo de eventos mantidos em memória
	maxEvents int

	// mu protege acesso concorrente
	mu sync.RWMutex

	// alertThresholds define limiares para alertas
	alertThresholds map[AuditEventType]int

	// eventCounts conta eventos por tipo para detecção de anomalias
	eventCounts map[string]int

	// lastReset é quando os contadores foram resetados
	lastReset time.Time
}

// NewAuditLogger cria um novo logger de auditoria
func NewAuditLogger() *AuditLogger {
	al := &AuditLogger{
		events:          make([]AuditEvent, 0, 1000),
		maxEvents:       10000,
		alertThresholds: make(map[AuditEventType]int),
		eventCounts:     make(map[string]int),
		lastReset:       time.Now(),
	}

	// Definir limiares de alerta
	al.alertThresholds[EventLoginFailed] = 10        // 10 falhas por minuto
	al.alertThresholds[EventRateLimitExceeded] = 50  // 50 rate limits por minuto
	al.alertThresholds[EventUnauthorizedAccess] = 20 // 20 acessos não autorizados

	// Iniciar goroutine de reset de contadores
	go al.resetCounters()

	return al
}

// =============================================================================
// LOGGING
// =============================================================================

// Log registra um evento de auditoria
//
// Parâmetros:
//   - event: evento a ser registrado
func (al *AuditLogger) Log(event AuditEvent) {
	al.mu.Lock()
	defer al.mu.Unlock()

	// Adicionar timestamp se não definido
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Gerar ID se não definido
	if event.ID == "" {
		event.ID = generateEventID()
	}

	// Adicionar ao buffer
	al.events = append(al.events, event)

	// Manter tamanho máximo
	if len(al.events) > al.maxEvents {
		// Remover eventos mais antigos
		al.events = al.events[len(al.events)-al.maxEvents:]
	}

	// Contar eventos para detecção de anomalias
	key := string(event.Type) + "_" + event.ClientIP
	al.eventCounts[key]++

	// Verificar limiares de alerta
	al.checkAlertThreshold(event)

	// Log para saída padrão (em produção, enviar para sistema centralizado)
	al.logToOutput(event)
}

// LogAuth registra evento de autenticação
func (al *AuditLogger) LogAuth(eventType AuditEventType, userID, clientIP, userAgent, result string, details map[string]interface{}) {
	severity := SeverityInfo
	if eventType == EventLoginFailed || eventType == EventUnauthorizedAccess {
		severity = SeverityWarning
	}

	al.Log(AuditEvent{
		Type:      eventType,
		Severity:  severity,
		UserID:    userID,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Result:    result,
		Details:   details,
	})
}

// LogDataAccess registra acesso a dados
func (al *AuditLogger) LogDataAccess(userID, clientIP, resource, action, result string) {
	al.Log(AuditEvent{
		Type:     EventDataAccess,
		Severity: SeverityInfo,
		UserID:   userID,
		ClientIP: clientIP,
		Resource: resource,
		Action:   action,
		Result:   result,
	})
}

// LogSecurity registra evento de segurança
func (al *AuditLogger) LogSecurity(eventType AuditEventType, clientIP string, details map[string]interface{}) {
	severity := SeverityWarning
	if eventType == EventSuspiciousActivity || eventType == EventUnauthorizedAccess {
		severity = SeverityError
	}

	al.Log(AuditEvent{
		Type:     eventType,
		Severity: severity,
		ClientIP: clientIP,
		Result:   "blocked",
		Details:  details,
	})
}

// =============================================================================
// ANÁLISE E ALERTAS
// =============================================================================

// checkAlertThreshold verifica se o limiar de alerta foi atingido
func (al *AuditLogger) checkAlertThreshold(event AuditEvent) {
	threshold, ok := al.alertThresholds[event.Type]
	if !ok {
		return
	}

	key := string(event.Type) + "_" + event.ClientIP
	count := al.eventCounts[key]

	if count >= threshold {
		// Gerar alerta
		al.generateAlert(event, count, threshold)
	}
}

// generateAlert gera um alerta de segurança
func (al *AuditLogger) generateAlert(event AuditEvent, count, threshold int) {
	alert := AuditEvent{
		Type:     EventSuspiciousActivity,
		Severity: SeverityCritical,
		ClientIP: event.ClientIP,
		Result:   "alert",
		Details: map[string]interface{}{
			"trigger_event": event.Type,
			"count":         count,
			"threshold":     threshold,
			"message":       "Limiar de segurança excedido",
		},
	}

	// Log do alerta
	log.Printf("[SECURITY ALERT] %s de %s: %d eventos (limiar: %d)",
		event.Type, maskIP(event.ClientIP), count, threshold)

	// Em produção, aqui enviaria para:
	// - Sistema de monitoramento (Datadog, Prometheus, etc.)
	// - Canal Slack/Discord
	// - Email de segurança
	// - SIEM

	al.events = append(al.events, alert)
}

// resetCounters reseta contadores periodicamente
func (al *AuditLogger) resetCounters() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		al.mu.Lock()
		al.eventCounts = make(map[string]int)
		al.lastReset = time.Now()
		al.mu.Unlock()
	}
}

// =============================================================================
// CONSULTA
// =============================================================================

// GetRecentEvents retorna eventos recentes
func (al *AuditLogger) GetRecentEvents(limit int) []AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	if limit <= 0 || limit > len(al.events) {
		limit = len(al.events)
	}

	// Retornar os mais recentes
	start := len(al.events) - limit
	if start < 0 {
		start = 0
	}

	result := make([]AuditEvent, limit)
	copy(result, al.events[start:])
	return result
}

// GetEventsByUser retorna eventos de um usuário específico
func (al *AuditLogger) GetEventsByUser(userID string, limit int) []AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	result := make([]AuditEvent, 0)
	for i := len(al.events) - 1; i >= 0 && len(result) < limit; i-- {
		if al.events[i].UserID == userID {
			result = append(result, al.events[i])
		}
	}

	return result
}

// GetEventsByIP retorna eventos de um IP específico
func (al *AuditLogger) GetEventsByIP(clientIP string, limit int) []AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	result := make([]AuditEvent, 0)
	for i := len(al.events) - 1; i >= 0 && len(result) < limit; i-- {
		if al.events[i].ClientIP == clientIP {
			result = append(result, al.events[i])
		}
	}

	return result
}

// GetSecurityEvents retorna apenas eventos de segurança
func (al *AuditLogger) GetSecurityEvents(limit int) []AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	securityTypes := map[AuditEventType]bool{
		EventLoginFailed:        true,
		EventRateLimitExceeded:  true,
		EventUnauthorizedAccess: true,
		EventSuspiciousActivity: true,
		EventTokenInvalid:       true,
	}

	result := make([]AuditEvent, 0)
	for i := len(al.events) - 1; i >= 0 && len(result) < limit; i-- {
		if securityTypes[al.events[i].Type] {
			result = append(result, al.events[i])
		}
	}

	return result
}

// =============================================================================
// OUTPUT
// =============================================================================

// logToOutput envia evento para saída de log
func (al *AuditLogger) logToOutput(event AuditEvent) {
	sanitized := sanitizeAuditEvent(event)
	// Formato JSON estruturado para parsing fácil
	jsonBytes, err := json.Marshal(sanitized)
	if err != nil {
		log.Printf("[AUDIT] Erro ao serializar evento: %v", err)
		return
	}

	// Prefixo baseado na severidade
	prefix := "[AUDIT]"
	switch event.Severity {
	case SeverityWarning:
		prefix = "[AUDIT-WARN]"
	case SeverityError:
		prefix = "[AUDIT-ERROR]"
	case SeverityCritical:
		prefix = "[AUDIT-CRITICAL]"
	}

	log.Printf("%s %s", prefix, string(jsonBytes))
}

// =============================================================================
// FUNÇÕES AUXILIARES
// =============================================================================

func sanitizeAuditEvent(event AuditEvent) AuditEvent {
	sanitized := event
	sanitized.ClientIP = maskIP(event.ClientIP)
	sanitized.UserAgent = ""
	sanitized.Details = sanitizeDetails(event.Details)
	return sanitized
}

func sanitizeDetails(details map[string]interface{}) map[string]interface{} {
	if len(details) == 0 {
		return nil
	}

	sanitized := make(map[string]interface{}, len(details))
	for key, value := range details {
		lowerKey := strings.ToLower(key)
		if isSensitiveKey(lowerKey) {
			sanitized[key] = "[redacted]"
			continue
		}
		switch v := value.(type) {
		case string:
			if isSensitiveValue(v) {
				sanitized[key] = "[redacted]"
			} else if len(v) > 120 {
				sanitized[key] = v[:120] + "…"
			} else {
				sanitized[key] = v
			}
		default:
			sanitized[key] = v
		}
	}

	return sanitized
}

func isSensitiveKey(key string) bool {
	sensitiveKeys := []string{
		"email",
		"phone",
		"token",
		"pin",
		"password",
		"name",
		"message",
		"content",
		"body",
		"recipient",
	}
	for _, candidate := range sensitiveKeys {
		if strings.Contains(key, candidate) {
			return true
		}
	}
	return false
}

func isSensitiveValue(value string) bool {
	if strings.Contains(value, "@") {
		return true
	}
	if strings.HasPrefix(value, "+") && len(value) > 8 {
		return true
	}
	return false
}

func maskIP(value string) string {
	if value == "" {
		return ""
	}
	if strings.Contains(value, ".") {
		parts := strings.Split(value, ".")
		if len(parts) == 4 {
			return parts[0] + "." + parts[1] + ".x.x"
		}
	}
	if strings.Contains(value, ":") {
		parts := strings.Split(value, ":")
		if len(parts) >= 2 {
			return parts[0] + ":" + parts[1] + "::"
		}
	}
	return "x.x.x.x"
}

// generateEventID gera um ID único para o evento
func generateEventID() string {
	now := time.Now()
	return now.Format("20060102150405") + "-" + randomString(6)
}

// randomString gera string aleatória
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// =============================================================================
// SINGLETON GLOBAL
// =============================================================================

var (
	globalAuditLogger *AuditLogger
	auditLoggerOnce   sync.Once
)

// GetAuditLogger retorna a instância global do logger de auditoria
func GetAuditLogger() *AuditLogger {
	auditLoggerOnce.Do(func() {
		globalAuditLogger = NewAuditLogger()
	})
	return globalAuditLogger
}
