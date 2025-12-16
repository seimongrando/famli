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
		// Auth
		"auth.invalid_data":     "Não foi possível entender os dados.",
		"auth.email_required":   "Preencha e-mail e senha.",
		"auth.prepare_error":    "Erro ao preparar sua conta.",
		"auth.email_exists":     "Este e-mail já está cadastrado.",
		"auth.create_error":     "Erro ao criar conta.",
		"auth.session_error":    "Erro ao criar sessão.",
		"auth.not_found":        "Conta não encontrada.",
		"auth.invalid_password": "Senha inválida.",
		"auth.session_expired":  "Sessão expirada.",
		"auth.session_invalid":  "Sessão inválida.",
		"auth.logout_success":   "Sessão encerrada.",

		// Box Items
		"box.invalid_content": "Conteúdo inválido.",
		"box.title_required":  "Dê um título ao que você quer guardar.",
		"box.save_error":      "Erro ao guardar.",
		"box.not_found":       "Item não encontrado.",
		"box.deleted":         "Item removido.",

		// Guardians
		"guardian.invalid_data":  "Dados inválidos.",
		"guardian.name_required": "Informe o nome da pessoa.",
		"guardian.add_error":     "Erro ao adicionar pessoa.",
		"guardian.not_found":     "Pessoa não encontrada.",
		"guardian.deleted":       "Pessoa removida.",

		// Settings
		"settings.invalid_data": "Dados inválidos.",

		// Guide
		"guide.invalid_data":   "Dados inválidos.",
		"guide.invalid_status": "Status inválido.",
		"guide.progress_error": "Erro ao salvar progresso.",

		// Assistant
		"assistant.empty_input": "Envie uma mensagem.",
		"assistant.start":       "Que bom que você está aqui! Sugiro começar pelo mais simples: registre o contato de uma pessoa de confiança. Pode ser um filho, neto ou amigo próximo. Assim, se precisar, alguém saberá que você está cuidando de tudo.",
		"assistant.passwords":   "Aqui no Famli você não guarda as senhas em si, mas explica onde elas estão. Por exemplo: 'Minhas senhas ficam no aplicativo 1Password, no celular. O e-mail de recuperação é fulano@email.com'. Assim fica seguro e alguém de confiança consegue ajudar se precisar.",
		"assistant.guardians":   "Pessoas de confiança são familiares ou amigos que você autoriza a serem avisados se um dia precisar de ajuda. No momento, elas não têm acesso automático às suas informações — só você decide o que compartilhar.",
		"assistant.documents":   "Você pode registrar informações sobre documentos, planos de saúde e seguros. Basta criar uma nova informação e explicar onde estão os documentos físicos ou digitais, e quem contatar em caso de necessidade.",
		"assistant.memories":    "As memórias são um espaço especial para deixar mensagens, histórias e recados para quem você ama. Pode escrever para uma pessoa específica ou deixar algo geral. É o coração do Famli.",
		"assistant.security":    "Seus dados são seus. Nada é compartilhado automaticamente e você pode apagar tudo quando quiser. Não vendemos nem usamos suas informações para marketing. Adicionar alguém como pessoa de confiança não dá acesso automático às suas informações.",
		"assistant.help":        "Estou aqui para ajudar! Você pode me perguntar sobre: como começar, como registrar informações importantes, como adicionar pessoas de confiança, ou como deixar mensagens para quem você ama.",
		"assistant.default":     "Entendi. Estou aqui para ajudar você a organizar o que é importante. Você pode guardar informações, indicar pessoas de confiança ou deixar memórias e mensagens. O que gostaria de fazer?",
	},
	"en": {
		// Auth
		"auth.invalid_data":     "Unable to understand the data.",
		"auth.email_required":   "Please fill in email and password.",
		"auth.prepare_error":    "Error preparing your account.",
		"auth.email_exists":     "This email is already registered.",
		"auth.create_error":     "Error creating account.",
		"auth.session_error":    "Error creating session.",
		"auth.not_found":        "Account not found.",
		"auth.invalid_password": "Invalid password.",
		"auth.session_expired":  "Session expired.",
		"auth.session_invalid":  "Invalid session.",
		"auth.logout_success":   "Session ended.",

		// Box Items
		"box.invalid_content": "Invalid content.",
		"box.title_required":  "Give a title to what you want to store.",
		"box.save_error":      "Error saving.",
		"box.not_found":       "Item not found.",
		"box.deleted":         "Item removed.",

		// Guardians
		"guardian.invalid_data":  "Invalid data.",
		"guardian.name_required": "Please provide the person's name.",
		"guardian.add_error":     "Error adding person.",
		"guardian.not_found":     "Person not found.",
		"guardian.deleted":       "Person removed.",

		// Settings
		"settings.invalid_data": "Invalid data.",

		// Guide
		"guide.invalid_data":   "Invalid data.",
		"guide.invalid_status": "Invalid status.",
		"guide.progress_error": "Error saving progress.",

		// Assistant
		"assistant.empty_input": "Send a message.",
		"assistant.start":       "Great that you're here! I suggest starting with something simple: register a trusted person's contact. It could be a son, grandchild, or close friend. That way, if needed, someone will know you're taking care of everything.",
		"assistant.passwords":   "Here at Famli you don't store the passwords themselves, but explain where they are. For example: 'My passwords are in the 1Password app, on my phone. The recovery email is someone@email.com'. This way it's secure and a trusted person can help if needed.",
		"assistant.guardians":   "Trusted people are family members or friends you authorize to be notified if you ever need help. At the moment, they don't have automatic access to your information — only you decide what to share.",
		"assistant.documents":   "You can register information about documents, health plans, and insurance. Just create a new information and explain where the physical or digital documents are, and who to contact if needed.",
		"assistant.memories":    "Memories are a special space to leave messages, stories, and notes for those you love. You can write to a specific person or leave something general. It's the heart of Famli.",
		"assistant.security":    "Your data is yours. Nothing is shared automatically and you can delete everything whenever you want. We don't sell or use your information for marketing. Adding someone as a trusted person doesn't give automatic access to your information.",
		"assistant.help":        "I'm here to help! You can ask me about: how to start, how to register important information, how to add trusted people, or how to leave messages for those you love.",
		"assistant.default":     "I understand. I'm here to help you organize what's important. You can store information, indicate trusted people, or leave memories and messages. What would you like to do?",
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
