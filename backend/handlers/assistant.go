package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

// AssistantHandler returns warm canned answers that imitate a concierge.
type AssistantHandler struct{}

// NewAssistantHandler builds the handler.
func NewAssistantHandler() *AssistantHandler {
	return &AssistantHandler{}
}

type assistantPayload struct {
	Input string `json:"input"`
}

type assistantResponse struct {
	Reply string `json:"reply"`
}

// Handle answers the prompt with human language examples.
func (h *AssistantHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var payload assistantPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || strings.TrimSpace(payload.Input) == "" {
		writeError(w, http.StatusBadRequest, "Envie uma mensagem simples.")
		return
	}

	reply := buildSuggestion(payload.Input)
	writeJSON(w, http.StatusOK, assistantResponse{Reply: reply})
}

func buildSuggestion(input string) string {
	normalized := strings.ToLower(input)

	switch {
	case strings.Contains(normalized, "apólice") || strings.Contains(normalized, "seguro"):
		return "Perfeito! Clique em 'Adicionar item', escolha o tipo 'Documento' e cole as informações da apólice. Assim tudo fica guardado em um só lugar."
	case strings.Contains(normalized, "pessoa") || strings.Contains(normalized, "confiança"):
		return "Ótima ideia incluir alguém de confiança. Toque em 'Pessoas de confiança' e depois em 'Adicionar' para informar nome e e-mail."
	case strings.Contains(normalized, "protocolo") || strings.Contains(normalized, "emergência"):
		return "Para o protocolo de emergência, abra 'Configurações' e ative o botão 'Avisar em caso de emergência'. Você controla tudo."
	case strings.Contains(normalized, "ajuda") || strings.Contains(normalized, "como"):
		return "Estou aqui para guiar você. Diga o que deseja guardar ou organizar que eu mostro o próximo passo."
	default:
		return "Tudo certo. Escolha entre guardar informações, indicar uma pessoa de confiança ou ajustar configurações. Estou acompanhando você em cada etapa."
	}
}
