package i18n

import (
	"net/http"
	"strings"
)

// Messages armazena as traduções
type Messages map[string]string

// Translations contém todas as traduções por idioma
var Translations = map[string]Messages{
	"pt-BR": {
		// =======================================================================
		// AUTH - Autenticação
		// =======================================================================
		"auth.invalid_data":        "Dados inválidos.",
		"auth.email_required":      "Preencha e-mail e senha.",
		"auth.email_invalid":       "E-mail inválido.",
		"auth.password_weak":       "Senha precisa ter no mínimo 8 caracteres com letras e números.",
		"auth.prepare_error":       "Não foi possível preparar sua conta.",
		"auth.email_exists":        "Não foi possível criar a conta. Tente outro e-mail.",
		"auth.create_error":        "Não foi possível criar a conta.",
		"auth.session_error":       "Não foi possível iniciar a sessão.",
		"auth.not_found":           "Conta não encontrada.",
		"auth.invalid_credentials": "E-mail ou senha incorretos.",
		"auth.session_expired":     "Sessão expirada.",
		"auth.session_invalid":     "Sessão inválida.",
		"auth.logout_success":      "Sessão encerrada.",
		"auth.rate_limit":          "Muitas tentativas. Aguarde alguns minutos.",
		"auth.user_not_found":      "Usuário não encontrado.",
		"auth.password_incorrect":  "Senha incorreta.",
		"auth.delete_confirm":      "Texto de confirmação incorreto.",
		"auth.delete_error":        "Não foi possível excluir a conta.",
		"auth.delete_success":      "Conta excluída com sucesso. Todos os dados foram removidos.",
		"auth.export_error":        "Não foi possível exportar os dados.",
		"auth.internal_error":      "Não foi possível processar a solicitação.",

		// =======================================================================
		// BOX - Itens da Caixa Famli
		// =======================================================================
		"box.invalid_content":  "Conteúdo inválido.",
		"box.title_required":   "Dê um título ao que você quer guardar.",
		"box.title_too_long":   "Título muito longo.",
		"box.content_too_long": "Conteúdo muito longo.",
		"box.invalid_detected": "Conteúdo inválido detectado.",
		"box.save_error":       "Não foi possível salvar.",
		"box.list_error":       "Não foi possível carregar os itens.",
		"box.not_found":        "Item não encontrado.",
		"box.deleted":          "Item removido.",
		"box.invalid_query":    "Consulta inválida.",

		// =======================================================================
		// GUARDIANS - Pessoas de Confiança
		// =======================================================================
		"guardian.invalid_data":   "Dados inválidos.",
		"guardian.name_required":  "Informe o nome da pessoa.",
		"guardian.add_error":      "Não foi possível adicionar a pessoa.",
		"guardian.not_found":      "Pessoa não encontrada.",
		"guardian.deleted":        "Pessoa removida.",
		"guardian.notes_too_long": "As notas são muito longas. Máximo de 1000 caracteres.",
		"guardian.pin_too_short":  "O PIN deve ter pelo menos 4 caracteres.",
		"guardian.pin_required":   "PIN obrigatório para criar a pessoa de confiança.",

		// =======================================================================
		// SETTINGS - Configurações
		// =======================================================================
		"settings.invalid_data": "Dados inválidos.",

		// =======================================================================
		// GUIDE - Guia Famli
		// =======================================================================
		"guide.invalid_data":   "Dados inválidos.",
		"guide.invalid_status": "Status inválido.",
		"guide.progress_error": "Não foi possível salvar o progresso.",

		// =======================================================================
		// ADMIN - Administração
		// =======================================================================
		"admin.not_authenticated": "Não autenticado.",
		"admin.user_not_found":    "Usuário não encontrado.",
		"admin.access_denied":     "Acesso não permitido.",

		// =======================================================================
		// ASSISTANT - Assistente
		// =======================================================================
		"assistant.empty_input": "Envie uma mensagem.",
		"assistant.start":       "Que bom que você está aqui! Sugiro começar pelo mais simples: registre o contato de uma pessoa de confiança. Pode ser um filho, neto ou amigo próximo. Assim, se precisar, alguém saberá que você está cuidando do que importa.",
		"assistant.passwords":   "Aqui no Famli você não guarda as senhas em si, mas explica onde elas estão. Por exemplo: 'Minhas senhas ficam no aplicativo 1Password, no celular. O e-mail de recuperação é fulano@email.com'. Assim fica seguro e alguém de confiança consegue ajudar se precisar.",
		"assistant.guardians":   "Pessoas de confiança são familiares ou amigos com quem você pode compartilhar informações quando quiser. No momento, elas não têm acesso automático às suas informações — só você decide o que compartilhar.",
		"assistant.documents":   "Você pode registrar informações sobre documentos, planos de saúde e seguros. Basta criar uma nova informação e explicar onde estão os documentos físicos ou digitais, e quem contatar em caso de necessidade.",
		"assistant.memories":    "As memórias são um espaço especial para deixar mensagens, histórias e recados para quem você ama. Pode escrever para uma pessoa específica ou deixar algo geral. É o coração do Famli.",
		"assistant.security":    "Seus dados são seus. Nada é compartilhado automaticamente e você pode apagar tudo quando quiser. Não vendemos nem usamos suas informações para marketing. Adicionar alguém como pessoa de confiança não dá acesso automático às suas informações.",
		"assistant.help":        "Estou aqui para ajudar! Você pode me perguntar sobre: como começar, como registrar informações importantes, como adicionar pessoas de confiança, ou como deixar mensagens para quem você ama.",
		"assistant.default":     "Entendi. Estou aqui para ajudar você a organizar o que é importante. Você pode guardar informações, indicar pessoas de confiança ou deixar memórias e mensagens. O que gostaria de fazer?",

		// =======================================================================
		// FEEDBACK
		// =======================================================================
		"feedback.invalid_data":     "Dados inválidos.",
		"feedback.save_error":       "Não foi possível enviar o feedback.",
		"feedback.update_error":     "Não foi possível atualizar o feedback.",
		"feedback.not_found":        "Feedback não encontrado.",
		"feedback.type_required":    "Selecione o tipo de feedback.",
		"feedback.send_success":     "Feedback enviado com sucesso!",
		"feedback.update_success":   "Feedback atualizado com sucesso.",
		"feedback.message_too_long": "A mensagem é muito longa. Máximo de 2000 caracteres.",

		// =======================================================================
		// ANALYTICS
		// =======================================================================
		"analytics.invalid_data": "Dados inválidos.",
		"analytics.track_error":  "Não foi possível registrar o evento.",

		// =======================================================================
		// OAUTH - Login Social
		// =======================================================================
		"oauth.google_not_configured": "Login com Google não está configurado.",
		"oauth.apple_not_configured":  "Login com Apple não está configurado.",
		"oauth.token_required":        "Token de autenticação é obrigatório.",
		"oauth.invalid_token":         "Token de autenticação inválido.",
		"oauth.email_not_verified":    "O e-mail precisa estar verificado.",

		// =======================================================================
		// SHARE - Compartilhamento com Guardiões
		// =======================================================================
		"share.invalid_data": "Dados inválidos.",
		"share.create_error": "Não foi possível criar o link.",
		"share.list_error":   "Não foi possível listar os links.",
		"share.not_found":    "Link não encontrado.",
		"share.deleted":      "Link removido com sucesso.",
		"share.link_expired": "Este link expirou ou não está mais disponível.",
		"share.invalid_pin":  "PIN incorreto.",
		"share.pin_required": "PIN obrigatório para acessar este link.",
		"share.access_error": "Não foi possível acessar o conteúdo.",

		// =======================================================================
		// PASSWORD RESET - Recuperação de Senha
		// =======================================================================
		"password.reset_sent":    "Se o e-mail existir, você receberá instruções para redefinir sua senha.",
		"password.reset_invalid": "Link de redefinição inválido ou expirado.",
		"password.reset_success": "Senha alterada com sucesso!",
		"password.reset_error":   "Não foi possível alterar a senha.",

		// =======================================================================
		// GUIDE CARDS - Títulos e descrições do Guia Famli
		// =======================================================================
		"guide.card.welcome.title":         "Comece por aqui",
		"guide.card.welcome.description":   "Dê o primeiro passo: registre algo simples, como o telefone de emergência ou um contato importante.",
		"guide.card.people.title":          "Pessoas importantes",
		"guide.card.people.description":    "Quem são as pessoas que devem ser avisadas se você precisar de ajuda? Registre aqui seus contatos de confiança.",
		"guide.card.locations.title":       "Onde estão as coisas importantes",
		"guide.card.locations.description": "Documentos, chaves, cartões... Explique onde estão as coisas que alguém precisaria encontrar.",
		"guide.card.routines.title":        "Rotina que não pode parar",
		"guide.card.routines.description":  "Medicamentos, contas automáticas, pets... O que precisa continuar funcionando mesmo se você não estiver por perto?",
		"guide.card.access.title":          "Como acessar suas coisas",
		"guide.card.access.description":    "Explique onde estão suas senhas (não as senhas em si!) e como alguém de confiança pode ajudar a acessar.",
		"guide.card.memories.title":        "Notas pessoais e memórias",
		"guide.card.memories.description":  "Mensagens, histórias, recados... Um espaço para deixar algo especial para quem você ama.",
	},
	"en": {
		// =======================================================================
		// AUTH - Authentication
		// =======================================================================
		"auth.invalid_data":        "Invalid data.",
		"auth.email_required":      "Please fill in email and password.",
		"auth.email_invalid":       "Invalid email.",
		"auth.password_weak":       "Password must have at least 8 characters with letters and numbers.",
		"auth.prepare_error":       "Unable to prepare your account.",
		"auth.email_exists":        "Unable to create account. Try another email.",
		"auth.create_error":        "Unable to create account.",
		"auth.session_error":       "Unable to start session.",
		"auth.not_found":           "Account not found.",
		"auth.invalid_credentials": "Invalid email or password.",
		"auth.session_expired":     "Session expired.",
		"auth.session_invalid":     "Invalid session.",
		"auth.logout_success":      "Session ended.",
		"auth.rate_limit":          "Too many attempts. Please wait a few minutes.",
		"auth.user_not_found":      "User not found.",
		"auth.password_incorrect":  "Incorrect password.",
		"auth.delete_confirm":      "Incorrect confirmation text.",
		"auth.delete_error":        "Unable to delete account.",
		"auth.delete_success":      "Account deleted successfully. All data has been removed.",
		"auth.export_error":        "Unable to export data.",
		"auth.internal_error":      "Unable to process the request.",

		// =======================================================================
		// BOX - Famli Box Items
		// =======================================================================
		"box.invalid_content":  "Invalid content.",
		"box.title_required":   "Give a title to what you want to store.",
		"box.title_too_long":   "Title is too long.",
		"box.content_too_long": "Content is too long.",
		"box.invalid_detected": "Invalid content detected.",
		"box.save_error":       "Unable to save.",
		"box.list_error":       "Unable to load items.",
		"box.not_found":        "Item not found.",
		"box.deleted":          "Item removed.",
		"box.invalid_query":    "Invalid query.",

		// =======================================================================
		// GUARDIANS - Trusted People
		// =======================================================================
		"guardian.invalid_data":   "Invalid data.",
		"guardian.name_required":  "Please provide the person's name.",
		"guardian.add_error":      "Unable to add person.",
		"guardian.not_found":      "Person not found.",
		"guardian.deleted":        "Person removed.",
		"guardian.notes_too_long": "Notes are too long. Maximum 1000 characters.",
		"guardian.pin_too_short":  "PIN must be at least 4 characters.",
		"guardian.pin_required":   "A PIN is required to create a trusted person.",

		// =======================================================================
		// SETTINGS - Settings
		// =======================================================================
		"settings.invalid_data": "Invalid data.",

		// =======================================================================
		// GUIDE - Famli Guide
		// =======================================================================
		"guide.invalid_data":   "Invalid data.",
		"guide.invalid_status": "Invalid status.",
		"guide.progress_error": "Unable to save progress.",

		// =======================================================================
		// ADMIN - Administration
		// =======================================================================
		"admin.not_authenticated": "Not authenticated.",
		"admin.user_not_found":    "User not found.",
		"admin.access_denied":     "Access denied.",

		// =======================================================================
		// ASSISTANT - Assistant
		// =======================================================================
		"assistant.empty_input": "Send a message.",
		"assistant.start":       "Great that you're here! I suggest starting with something simple: register a trusted person's contact. It could be a son, grandchild, or close friend. That way, if needed, someone will know you're taking care of what matters.",
		"assistant.passwords":   "Here at Famli you don't store the passwords themselves, but explain where they are. For example: 'My passwords are in the 1Password app, on my phone. The recovery email is someone@email.com'. This way it's secure and a trusted person can help if needed.",
		"assistant.guardians":   "Trusted people are family members or friends you can share information with when you want. At the moment, they don't have automatic access to your information — only you decide what to share.",
		"assistant.documents":   "You can register information about documents, health plans, and insurance. Just create a new information and explain where the physical or digital documents are, and who to contact if needed.",
		"assistant.memories":    "Memories are a special space to leave messages, stories, and notes for those you love. You can write to a specific person or leave something general. It's the heart of Famli.",
		"assistant.security":    "Your data is yours. Nothing is shared automatically and you can delete everything whenever you want. We don't sell or use your information for marketing. Adding someone as a trusted person doesn't give automatic access to your information.",
		"assistant.help":        "I'm here to help! You can ask me about: how to start, how to register important information, how to add trusted people, or how to leave messages for those you love.",
		"assistant.default":     "I understand. I'm here to help you organize what's important. You can store information, indicate trusted people, or leave memories and messages. What would you like to do?",

		// =======================================================================
		// FEEDBACK
		// =======================================================================
		"feedback.invalid_data":     "Invalid data.",
		"feedback.save_error":       "Unable to send feedback.",
		"feedback.update_error":     "Unable to update feedback.",
		"feedback.not_found":        "Feedback not found.",
		"feedback.type_required":    "Please select a feedback type.",
		"feedback.send_success":     "Feedback sent successfully!",
		"feedback.update_success":   "Feedback updated successfully.",
		"feedback.message_too_long": "Message is too long. Maximum 2000 characters.",

		// =======================================================================
		// ANALYTICS
		// =======================================================================
		"analytics.invalid_data": "Invalid data.",
		"analytics.track_error":  "Unable to record event.",

		// =======================================================================
		// OAUTH - Social Login
		// =======================================================================
		"oauth.google_not_configured": "Google login is not configured.",
		"oauth.apple_not_configured":  "Apple login is not configured.",
		"oauth.token_required":        "Authentication token is required.",
		"oauth.invalid_token":         "Invalid authentication token.",
		"oauth.email_not_verified":    "Email must be verified.",

		// =======================================================================
		// SHARE - Sharing with Guardians
		// =======================================================================
		"share.invalid_data": "Invalid data.",
		"share.create_error": "Unable to create link.",
		"share.list_error":   "Unable to list links.",
		"share.not_found":    "Link not found.",
		"share.deleted":      "Link removed successfully.",
		"share.link_expired": "This link has expired or is no longer available.",
		"share.invalid_pin":  "Incorrect PIN.",
		"share.pin_required": "A PIN is required to access this link.",
		"share.access_error": "Unable to access content.",

		// =======================================================================
		// PASSWORD RESET - Password Recovery
		// =======================================================================
		"password.reset_sent":    "If the email exists, you will receive instructions to reset your password.",
		"password.reset_invalid": "Invalid or expired reset link.",
		"password.reset_success": "Password changed successfully!",
		"password.reset_error":   "Unable to change password.",

		// =======================================================================
		// GUIDE CARDS - Famli Guide titles and descriptions
		// =======================================================================
		"guide.card.welcome.title":         "Start here",
		"guide.card.welcome.description":   "Take the first step: register something simple, like an emergency phone number or an important contact.",
		"guide.card.people.title":          "Important people",
		"guide.card.people.description":    "Who should be notified if you need help? Register your trusted contacts here.",
		"guide.card.locations.title":       "Where important things are",
		"guide.card.locations.description": "Documents, keys, cards... Explain where things are that someone might need to find.",
		"guide.card.routines.title":        "Routines that can't stop",
		"guide.card.routines.description":  "Medications, automatic bills, pets... What needs to keep running even if you're not around?",
		"guide.card.access.title":          "How to access your things",
		"guide.card.access.description":    "Explain where your passwords are (not the passwords themselves!) and how a trusted person can help access them.",
		"guide.card.memories.title":        "Personal notes and memories",
		"guide.card.memories.description":  "Messages, stories, notes... A space to leave something special for those you love.",
	},
}

// GetLocale extrai o idioma do header Accept-Language
func GetLocale(r *http.Request) string {
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang == "" {
		return "pt-BR"
	}

	// Parse simples do Accept-Language
	langs := strings.Split(acceptLang, ",")
	for _, lang := range langs {
		lang = strings.TrimSpace(strings.Split(lang, ";")[0])

		if strings.HasPrefix(lang, "pt") {
			return "pt-BR"
		}
		if strings.HasPrefix(lang, "en") {
			return "en"
		}
	}

	return "pt-BR"
}

// T retorna a tradução para uma chave
func T(locale, key string) string {
	if msgs, ok := Translations[locale]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}

	// Fallback para pt-BR
	if msgs, ok := Translations["pt-BR"]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}

	return key
}

// Tr é um helper que pega o locale do request
func Tr(r *http.Request, key string) string {
	return T(GetLocale(r), key)
}
