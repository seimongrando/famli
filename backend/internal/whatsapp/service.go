// =============================================================================
// FAMLI - Serviço de Processamento WhatsApp
// =============================================================================
// Este arquivo contém a lógica principal de processamento de mensagens.
// Ele interpreta o que o usuário enviou e toma a ação apropriada.
//
// Fluxo principal:
// 1. Mensagem chega via webhook
// 2. Identificamos o usuário pela sessão ou número
// 3. Processamos baseado no tipo de mensagem e estado atual
// 4. Salvamos na Caixa Famli se necessário
// 5. Enviamos resposta de confirmação
// =============================================================================

package whatsapp

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"famli/internal/storage"
)

// =============================================================================
// SERVIÇO PRINCIPAL
// =============================================================================

// Service gerencia toda a lógica de processamento de mensagens WhatsApp
type Service struct {
	// store é o armazenamento de dados do Famli
	store storage.Store

	// client é o cliente para enviar mensagens via Twilio
	client *TwilioClient

	// sessions armazena as sessões ativas dos usuários
	// Chave: número de telefone (ex: +5511999999999)
	sessions map[string]*UserSession

	// phoneToUser mapeia número de telefone para ID de usuário Famli
	// Permite vincular um número WhatsApp a uma conta Famli
	phoneToUser map[string]string

	// mu protege o acesso concorrente aos maps
	mu sync.RWMutex

	// config é a configuração do serviço
	config *Config
}

// NewService cria uma nova instância do serviço WhatsApp
//
// Parâmetros:
//   - store: armazenamento de dados do Famli
//   - config: configuração com credenciais Twilio
//
// Retorna:
//   - *Service: instância configurada do serviço
func NewService(store storage.Store, config *Config) *Service {
	var client *TwilioClient
	if config != nil && config.Enabled {
		client = NewTwilioClient(config.TwilioAccountSid, config.TwilioAuthToken, config.TwilioPhoneNumber)
	}

	return &Service{
		store:       store,
		client:      client,
		sessions:    make(map[string]*UserSession),
		phoneToUser: make(map[string]string),
		config:      config,
	}
}

// =============================================================================
// PROCESSAMENTO DE MENSAGENS
// =============================================================================

// ProcessMessage é o ponto de entrada principal para processar mensagens recebidas
//
// Parâmetros:
//   - msg: mensagem recebida do webhook Twilio
//
// Retorna:
//   - string: resposta a ser enviada ao usuário
//   - error: erro se houver falha no processamento
func (s *Service) ProcessMessage(msg *IncomingMessage) (string, error) {
	// Extrair número limpo (sem prefixo whatsapp:)
	phone := cleanPhoneNumber(msg.From)

	log.Printf("[WhatsApp] Mensagem recebida: tipo=%s, mídia=%d", msg.GetMessageType(), msg.NumMedia)

	// Obter ou criar sessão do usuário
	session := s.getOrCreateSession(phone)
	session.LastMessageAt = time.Now()

	// Verificar se é um comando especial
	if cmd := s.parseCommand(msg.Body); cmd != "" {
		return s.handleCommand(session, cmd, msg)
	}

	// Processar baseado no tipo de mensagem
	msgType := msg.GetMessageType()

	switch msgType {
	case MessageTypeText:
		return s.processTextMessage(session, msg)

	case MessageTypeImage:
		return s.processImageMessage(session, msg)

	case MessageTypeAudio:
		return s.processAudioMessage(session, msg)

	case MessageTypeDocument:
		return s.processDocumentMessage(session, msg)

	case MessageTypeLocation:
		return s.processLocationMessage(session, msg)

	default:
		return s.getHelpMessage(), nil
	}
}

// =============================================================================
// PROCESSAMENTO POR TIPO
// =============================================================================

// processTextMessage processa mensagens de texto
// Pode ser uma nota, memória ou informação a ser guardada
func (s *Service) processTextMessage(session *UserSession, msg *IncomingMessage) (string, error) {
	text := strings.TrimSpace(msg.Body)

	// Se não está vinculado, pedir para vincular
	if session.UserID == "" {
		return s.handleUnlinkedUser(session, text)
	}

	// Verificar estado da sessão
	switch session.State {
	case "awaiting_category":
		return s.handleCategorySelection(session, text)

	case "awaiting_confirmation":
		return s.handleConfirmation(session, text)

	default:
		// Estado idle - interpretar como novo item
		return s.startNewItem(session, text, "text")
	}
}

// processImageMessage processa imagens enviadas
// Salva como uma memória visual ou documento
func (s *Service) processImageMessage(session *UserSession, msg *IncomingMessage) (string, error) {
	if session.UserID == "" {
		return "📸 Vi sua foto! Para salvá-la no Famli, primeiro vincule seu número.\n\nDigite *vincular* para começar.", nil
	}

	// Criar item com a imagem
	caption := msg.Body
	if caption == "" {
		caption = "Foto enviada via WhatsApp"
	}

	// Iniciar processo de salvamento
	session.PendingItem = &PendingBoxItem{
		Content:   caption,
		Type:      "memory",
		MediaUrl:  msg.MediaUrl,
		MediaType: msg.MediaContentType,
		Title:     generateTitleFromContent(caption, 50),
	}
	session.State = "awaiting_category"
	s.saveSession(session)

	return fmt.Sprintf(
		"📸 *Foto recebida!*\n\n"+
			"Legenda: _%s_\n\n"+
			"Em qual categoria você quer guardar?\n\n"+
			"1️⃣ Família\n"+
			"2️⃣ Saúde\n"+
			"3️⃣ Finanças\n"+
			"4️⃣ Documentos\n"+
			"5️⃣ Memórias\n\n"+
			"_Responda com o número ou nome da categoria_",
		truncate(caption, 100),
	), nil
}

// processAudioMessage processa mensagens de voz
// No futuro, pode transcrever o áudio automaticamente
func (s *Service) processAudioMessage(session *UserSession, msg *IncomingMessage) (string, error) {
	if session.UserID == "" {
		return "🎤 Recebi seu áudio! Para salvá-lo, vincule seu número primeiro.\n\nDigite *vincular* para começar.", nil
	}

	// Por enquanto, salvar como nota de áudio
	// TODO: Implementar transcrição com Whisper/similar
	session.PendingItem = &PendingBoxItem{
		Content:   "Mensagem de voz enviada via WhatsApp",
		Type:      "note",
		MediaUrl:  msg.MediaUrl,
		MediaType: "audio",
		Title:     fmt.Sprintf("Áudio de %s", time.Now().Format("02/01/2006 15:04")),
	}
	session.State = "awaiting_category"
	s.saveSession(session)

	return "🎤 *Áudio recebido!*\n\n" +
		"Em qual categoria você quer guardar?\n\n" +
		"1️⃣ Família\n" +
		"2️⃣ Saúde\n" +
		"3️⃣ Finanças\n" +
		"4️⃣ Documentos\n" +
		"5️⃣ Memórias\n\n" +
		"_Responda com o número ou nome da categoria_", nil
}

// processDocumentMessage processa documentos (PDFs, etc.)
func (s *Service) processDocumentMessage(session *UserSession, msg *IncomingMessage) (string, error) {
	if session.UserID == "" {
		return "📄 Recebi seu documento! Para salvá-lo, vincule seu número primeiro.\n\nDigite *vincular* para começar.", nil
	}

	caption := msg.Body
	if caption == "" {
		caption = "Documento enviado via WhatsApp"
	}

	session.PendingItem = &PendingBoxItem{
		Content:   caption,
		Type:      "info",
		MediaUrl:  msg.MediaUrl,
		MediaType: "document",
		Title:     generateTitleFromContent(caption, 50),
	}
	session.State = "awaiting_category"
	s.saveSession(session)

	return "📄 *Documento recebido!*\n\n" +
		"Em qual categoria você quer guardar?\n\n" +
		"1️⃣ Família\n" +
		"2️⃣ Saúde\n" +
		"3️⃣ Finanças\n" +
		"4️⃣ Documentos\n" +
		"5️⃣ Memórias\n\n" +
		"_Responda com o número ou nome da categoria_", nil
}

// processLocationMessage processa localizações compartilhadas
func (s *Service) processLocationMessage(session *UserSession, msg *IncomingMessage) (string, error) {
	if session.UserID == "" {
		return "📍 Recebi a localização! Para salvá-la, vincule seu número primeiro.\n\nDigite *vincular* para começar.", nil
	}

	// Criar conteúdo com coordenadas
	content := fmt.Sprintf("Localização: %s, %s\nGoogle Maps: https://maps.google.com/?q=%s,%s",
		msg.Latitude, msg.Longitude, msg.Latitude, msg.Longitude)

	session.PendingItem = &PendingBoxItem{
		Content:  content,
		Type:     "location",
		Title:    "Localização importante",
		Category: "família",
	}
	session.State = "awaiting_confirmation"
	s.saveSession(session)

	return fmt.Sprintf(
		"📍 *Localização recebida!*\n\n"+
			"Coordenadas: %s, %s\n\n"+
			"Quer salvar como \"Localização importante\"?\n\n"+
			"✅ Responda *sim* para confirmar\n"+
			"✏️ Ou digite um título diferente",
		msg.Latitude, msg.Longitude,
	), nil
}

// =============================================================================
// FLUXO DE CRIAÇÃO DE ITEM
// =============================================================================

// startNewItem inicia o processo de criar um novo item na Caixa Famli
func (s *Service) startNewItem(session *UserSession, content string, contentType string) (string, error) {
	// Detectar automaticamente o tipo de item baseado no conteúdo
	itemType := detectItemType(content)
	title := generateTitleFromContent(content, 50)

	session.PendingItem = &PendingBoxItem{
		Content: content,
		Type:    itemType,
		Title:   title,
	}
	session.State = "awaiting_category"
	s.saveSession(session)

	return fmt.Sprintf(
		"📝 *Vou guardar isso para você!*\n\n"+
			"_%s_\n\n"+
			"Em qual categoria?\n\n"+
			"1️⃣ Família\n"+
			"2️⃣ Saúde\n"+
			"3️⃣ Finanças\n"+
			"4️⃣ Documentos\n"+
			"5️⃣ Memórias\n\n"+
			"_Responda com o número ou digite a categoria_",
		truncate(content, 200),
	), nil
}

// handleCategorySelection processa a seleção de categoria pelo usuário
func (s *Service) handleCategorySelection(session *UserSession, input string) (string, error) {
	category := parseCategory(input)

	if session.PendingItem == nil {
		session.State = "idle"
		s.saveSession(session)
		return "Ops! Algo deu errado. Envie sua mensagem novamente.", nil
	}

	session.PendingItem.Category = category
	session.State = "awaiting_confirmation"
	s.saveSession(session)

	return fmt.Sprintf(
		"✨ *Confirme os dados:*\n\n"+
			"📌 *Título:* %s\n"+
			"📁 *Categoria:* %s\n"+
			"📝 *Conteúdo:* _%s_\n\n"+
			"✅ Responda *sim* para salvar\n"+
			"❌ Responda *não* para cancelar\n"+
			"✏️ Ou digite um novo título",
		session.PendingItem.Title,
		category,
		truncate(session.PendingItem.Content, 150),
	), nil
}

// handleConfirmation processa a confirmação ou alteração do item
func (s *Service) handleConfirmation(session *UserSession, input string) (string, error) {
	inputLower := strings.ToLower(strings.TrimSpace(input))

	if session.PendingItem == nil {
		session.State = "idle"
		s.saveSession(session)
		return "Ops! Algo deu errado. Envie sua mensagem novamente.", nil
	}

	switch inputLower {
	case "sim", "s", "yes", "y", "confirmar", "ok":
		// Salvar o item na Caixa Famli
		return s.saveItemToBox(session)

	case "não", "nao", "n", "no", "cancelar":
		session.PendingItem = nil
		session.State = "idle"
		s.saveSession(session)
		return "❌ Cancelado! Se precisar de algo, é só me mandar uma mensagem.", nil

	default:
		// Usuário digitou um novo título
		session.PendingItem.Title = input
		return fmt.Sprintf(
			"✏️ *Título atualizado!*\n\n"+
				"📌 *Título:* %s\n"+
				"📁 *Categoria:* %s\n\n"+
				"✅ Responda *sim* para salvar\n"+
				"❌ Responda *não* para cancelar",
			session.PendingItem.Title,
			session.PendingItem.Category,
		), nil
	}
}

// saveItemToBox salva o item pendente na Caixa Famli
func (s *Service) saveItemToBox(session *UserSession) (string, error) {
	if session.PendingItem == nil || session.UserID == "" {
		return "Ops! Algo deu errado. Tente novamente.", nil
	}

	// Criar o item no storage
	item := &storage.BoxItem{
		Type:        storage.ItemType(session.PendingItem.Type),
		Title:       session.PendingItem.Title,
		Content:     session.PendingItem.Content,
		Category:    session.PendingItem.Category,
		IsImportant: false,
	}

	// Se tem mídia, adicionar à descrição
	if session.PendingItem.MediaUrl != "" {
		item.Content = fmt.Sprintf("%s\n\n[Mídia: %s]", item.Content, session.PendingItem.MediaUrl)
	}

	// Salvar no store
	created, err := s.store.CreateBoxItem(session.UserID, item)
	if err != nil {
		log.Printf("[WhatsApp] Erro ao salvar item: %v", err)
		return "😕 Desculpe, não consegui salvar. Tente novamente em alguns instantes.", nil
	}

	// Limpar sessão
	session.PendingItem = nil
	session.State = "idle"
	s.saveSession(session)

	return fmt.Sprintf(
		"✅ *Guardado com sucesso!*\n\n"+
			"📌 *%s*\n"+
			"📁 Categoria: %s\n\n"+
			"Você pode ver tudo na sua Caixa Famli:\n"+
			"🔗 famli.net/minha-caixa\n\n"+
			"_Continue me enviando o que quiser guardar!_ 💚",
		created.Title,
		created.Category,
	), nil
}

// =============================================================================
// COMANDOS
// =============================================================================

// parseCommand verifica se a mensagem é um comando conhecido
func (s *Service) parseCommand(text string) Command {
	textLower := strings.ToLower(strings.TrimSpace(text))

	// Comandos podem começar com / ou não
	textLower = strings.TrimPrefix(textLower, "/")

	switch textLower {
	case "ajuda", "help", "?", "oi", "olá", "ola", "menu":
		return CommandHelp
	case "guardar", "salvar", "save":
		return CommandSave
	case "listar", "ver", "list", "lista":
		return CommandList
	case "cancelar", "cancel", "parar", "sair":
		return CommandCancel
	case "status", "conta":
		return CommandStatus
	case "vincular", "conectar", "link", "login":
		return CommandLink
	default:
		return ""
	}
}

// handleCommand processa comandos especiais
func (s *Service) handleCommand(session *UserSession, cmd Command, msg *IncomingMessage) (string, error) {
	switch cmd {
	case CommandHelp:
		return s.getHelpMessage(), nil

	case CommandSave:
		return "📝 *Modo guardar ativado!*\n\n" +
			"Me envie o que você quer guardar:\n" +
			"• Uma mensagem de texto\n" +
			"• Uma foto\n" +
			"• Um áudio\n" +
			"• Um documento\n\n" +
			"_Estou esperando..._", nil

	case CommandList:
		return s.handleListCommand(session)

	case CommandCancel:
		session.PendingItem = nil
		session.State = "idle"
		s.saveSession(session)
		return "✅ Operação cancelada! Se precisar de algo, é só me chamar.", nil

	case CommandStatus:
		return s.handleStatusCommand(session)

	case CommandLink:
		return s.handleLinkCommand(session)

	default:
		return s.getHelpMessage(), nil
	}
}

// handleListCommand lista os últimos itens salvos pelo usuário
func (s *Service) handleListCommand(session *UserSession) (string, error) {
	if session.UserID == "" {
		return "Para ver seus itens, primeiro vincule seu número.\n\nDigite *vincular* para começar.", nil
	}

	items, err := s.store.GetBoxItems(session.UserID)
	if err != nil || len(items) == 0 {
		return "📭 Sua Caixa Famli está vazia!\n\nMe envie algo para guardar.", nil
	}

	// Mostrar os últimos 5 itens
	response := "📦 *Seus últimos itens:*\n\n"
	limit := 5
	if len(items) < limit {
		limit = len(items)
	}

	for i := 0; i < limit; i++ {
		item := items[i]
		emoji := getCategoryEmoji(item.Category)
		response += fmt.Sprintf("%s *%s*\n   _%s_\n\n", emoji, item.Title, truncate(item.Content, 50))
	}

	response += fmt.Sprintf("_Total: %d itens_\n\n🔗 Ver tudo: famli.net/minha-caixa", len(items))
	return response, nil
}

// handleStatusCommand mostra o status da conta
func (s *Service) handleStatusCommand(session *UserSession) (string, error) {
	if session.UserID == "" {
		return "📱 *Status: Não vinculado*\n\n" +
			"Seu WhatsApp ainda não está conectado a uma conta Famli.\n\n" +
			"Digite *vincular* para conectar.", nil
	}

	// Contar itens do usuário
	items, _ := s.store.GetBoxItems(session.UserID)
	itemCount := len(items)

	return fmt.Sprintf(
		"📱 *Status: Conectado* ✅\n\n"+
			"📦 Itens na Caixa: %d\n"+
			"📅 Última atividade: %s\n\n"+
			"🔗 Acesse: famli.net/minha-caixa",
		itemCount,
		session.LastMessageAt.Format("02/01/2006 15:04"),
	), nil
}

// handleLinkCommand inicia o processo de vincular número à conta Famli
func (s *Service) handleLinkCommand(session *UserSession) (string, error) {
	if session.UserID != "" {
		return "✅ Seu WhatsApp já está conectado!\n\n" +
			"Se quiser trocar de conta, acesse famli.net/configuracoes", nil
	}

	// Gerar código de vinculação (6 dígitos)
	// TODO: Implementar sistema real de códigos com expiração
	code := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)

	return fmt.Sprintf(
		"🔗 *Vincular WhatsApp ao Famli*\n\n"+
			"1️⃣ Acesse *famli.net*\n"+
			"2️⃣ Faça login na sua conta\n"+
			"3️⃣ Vá em *Configurações > WhatsApp*\n"+
			"4️⃣ Digite o código: *%s*\n\n"+
			"_O código expira em 10 minutos_",
		code,
	), nil
}

// handleUnlinkedUser trata mensagens de usuários não vinculados
func (s *Service) handleUnlinkedUser(session *UserSession, text string) (string, error) {
	return fmt.Sprintf(
		"👋 *Olá!* Sou o assistente do Famli.\n\n"+
			"Vi que você enviou:\n_%s_\n\n"+
			"Para guardar isso na sua Caixa Famli, preciso conectar seu WhatsApp à sua conta.\n\n"+
			"Digite *vincular* para começar!\n\n"+
			"_Não tem conta? Crie em famli.net_ 💚",
		truncate(text, 100),
	), nil
}

// =============================================================================
// MENSAGENS PADRÃO
// =============================================================================

// getHelpMessage retorna a mensagem de ajuda
func (s *Service) getHelpMessage() string {
	return "🏠 *Famli - Seu assistente de memórias*\n\n" +
		"Guarde o que importa diretamente pelo WhatsApp!\n\n" +
		"*O que você pode fazer:*\n\n" +
		"📝 Enviar *textos* para guardar\n" +
		"📸 Enviar *fotos* e memórias\n" +
		"🎤 Enviar *áudios* e notas de voz\n" +
		"📄 Enviar *documentos*\n" +
		"📍 Compartilhar *localizações*\n\n" +
		"*Comandos úteis:*\n\n" +
		"• *ajuda* - Esta mensagem\n" +
		"• *listar* - Ver últimos itens\n" +
		"• *vincular* - Conectar à conta\n" +
		"• *status* - Ver seu status\n" +
		"• *cancelar* - Cancelar operação\n\n" +
		"_É só me enviar o que quiser guardar!_ 💚"
}

// =============================================================================
// GERENCIAMENTO DE SESSÕES
// =============================================================================

// getOrCreateSession obtém ou cria uma sessão para o número
func (s *Service) getOrCreateSession(phone string) *UserSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, ok := s.sessions[phone]; ok {
		return session
	}

	// Criar nova sessão
	session := &UserSession{
		PhoneNumber: phone,
		State:       "idle",
		CreatedAt:   time.Now(),
	}

	// Verificar se o número já está vinculado a um usuário
	if userID, ok := s.phoneToUser[phone]; ok {
		session.UserID = userID
	}

	s.sessions[phone] = session
	return session
}

// saveSession salva a sessão atualizada
func (s *Service) saveSession(session *UserSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.PhoneNumber] = session
}

// LinkPhoneToUser vincula um número de telefone a um usuário Famli
func (s *Service) LinkPhoneToUser(phone, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	phone = cleanPhoneNumber(phone)
	s.phoneToUser[phone] = userID

	// Atualizar sessão se existir
	if session, ok := s.sessions[phone]; ok {
		session.UserID = userID
	}

	log.Printf("[WhatsApp] Número %s vinculado ao usuário %s", maskPhone(phone), userID)
}

// =============================================================================
// ENVIO DE MENSAGENS
// =============================================================================

// SendMessage envia uma mensagem para um número
func (s *Service) SendMessage(to, body string) error {
	if s.client == nil {
		log.Printf("[WhatsApp] Cliente não configurado, mensagem não enviada")
		return nil
	}

	return s.client.SendMessage(to, body)
}

// NotifyGuardians notifica os guardiões de um usuário
// Usado para alertas importantes
func (s *Service) NotifyGuardians(userID, message string) error {
	guardians, err := s.store.GetGuardians(userID)
	if err != nil {
		return err
	}

	for _, guardian := range guardians {
		if guardian.Phone != "" {
			if err := s.SendMessage(guardian.Phone, message); err != nil {
				log.Printf("[WhatsApp] Erro ao notificar guardião %s: %v", guardian.ID, err)
			}
		}
	}

	return nil
}

// =============================================================================
// FUNÇÕES AUXILIARES
// =============================================================================

// cleanPhoneNumber remove o prefixo whatsapp: do número
func cleanPhoneNumber(phone string) string {
	return strings.TrimPrefix(phone, "whatsapp:")
}

// truncate trunca uma string para o tamanho máximo especificado
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// generateTitleFromContent gera um título a partir do conteúdo
func generateTitleFromContent(content string, maxLen int) string {
	// Pegar primeira linha ou primeiras palavras
	lines := strings.Split(content, "\n")
	title := strings.TrimSpace(lines[0])

	// Limitar tamanho
	if len(title) > maxLen {
		// Tentar cortar em uma palavra
		words := strings.Fields(title)
		title = ""
		for _, word := range words {
			if len(title)+len(word)+1 > maxLen {
				break
			}
			if title != "" {
				title += " "
			}
			title += word
		}
	}

	if title == "" {
		title = "Item sem título"
	}

	return title
}

// detectItemType detecta o tipo de item baseado no conteúdo
func detectItemType(content string) string {
	contentLower := strings.ToLower(content)

	// Palavras-chave para cada tipo
	keywords := map[string][]string{
		"memory": {"lembro", "memória", "memória", "saudade", "querido", "amor", "filho", "neto", "família"},
		"info":   {"importante", "conta", "banco", "senha", "cpf", "documento", "cartão"},
		"access": {"login", "senha", "acesso", "usuário", "email"},
		"note":   {"nota", "lembrete", "anotar", "não esquecer"},
	}

	for itemType, words := range keywords {
		for _, word := range words {
			if strings.Contains(contentLower, word) {
				return itemType
			}
		}
	}

	return "note" // Padrão
}

// parseCategory converte entrada do usuário para categoria
func parseCategory(input string) string {
	inputLower := strings.ToLower(strings.TrimSpace(input))

	categories := map[string]string{
		"1": "família", "familia": "família", "fam": "família",
		"2": "saúde", "saude": "saúde", "sau": "saúde",
		"3": "finanças", "financas": "finanças", "fin": "finanças", "dinheiro": "finanças",
		"4": "documentos", "docs": "documentos", "doc": "documentos",
		"5": "memórias", "memorias": "memórias", "mem": "memórias", "memoria": "memórias",
	}

	if cat, ok := categories[inputLower]; ok {
		return cat
	}

	return "outros"
}

// getCategoryEmoji retorna o emoji para uma categoria
func getCategoryEmoji(category string) string {
	emojis := map[string]string{
		"família":    "👨‍👩‍👧‍👦",
		"saúde":      "🏥",
		"finanças":   "💰",
		"documentos": "📄",
		"memórias":   "💝",
		"outros":     "📌",
	}

	if emoji, ok := emojis[category]; ok {
		return emoji
	}
	return "📌"
}
