// =============================================================================
// FAMLI - Internacionalização de Meta Tags
// =============================================================================
// Este pacote gerencia as meta tags localizadas para compartilhamento social
// e SEO. As meta tags são injetadas no HTML antes de servir ao cliente.
//
// Idiomas suportados: pt-BR (padrão), en
// =============================================================================

package i18n

import (
	"net/http"
	"strings"
)

// MetaTags contém as meta tags traduzidas para um idioma
type MetaTags struct {
	Title       string
	Description string
	Keywords    string
	OGTitle     string
	OGDesc      string
	Language    string
	Locale      string
}

// Traduções das meta tags por idioma
var metaTagsTranslations = map[string]MetaTags{
	"pt-BR": {
		Title:       "Famli - Organize memórias e orientações para quem você ama",
		Description: "Transmita o que importa para as pessoas certas, quando for a hora. Organize memórias, documentos e orientações com cuidado, no seu tempo e com mais controle.",
		Keywords:    "memórias familiares, documentos importantes, organização familiar, legado, orientações familiares, planejamento familiar, segurança de dados",
		OGTitle:     "Famli - Organize memórias e orientações para quem você ama",
		OGDesc:      "Transmita o que importa para as pessoas certas, quando for a hora. Organize com cuidado, no seu tempo.",
		Language:    "Portuguese",
		Locale:      "pt_BR",
	},
	"en": {
		Title:       "Famli - Organize memories and guidance for your loved ones",
		Description: "Pass on what matters to the right people, when the time comes. Organize memories, documents, and guidance with care, at your own pace, with more control.",
		Keywords:    "family memories, important documents, family organization, legacy, family guidance, secure storage",
		OGTitle:     "Famli - Organize memories and guidance for your loved ones",
		OGDesc:      "Pass on what matters to the right people, when the time comes. Organize with care, at your own pace.",
		Language:    "English",
		Locale:      "en_US",
	},
}

// Textos originais em português (do index.html) para substituição
var originalTexts = struct {
	Title       string
	Description string
	Keywords    string
	OGTitle     string
	OGDesc      string
	Language    string
	Locale      string
	HTMLLang    string
}{
	Title:       "Famli - Organize memórias e orientações para quem você ama",
	Description: "Transmita o que importa para as pessoas certas, quando for a hora. Organize memórias, documentos e orientações com cuidado, no seu tempo e com mais controle.",
	Keywords:    "memórias familiares, documentos importantes, organização familiar, legado, orientações familiares, planejamento familiar, segurança de dados",
	OGTitle:     "Famli - Organize memórias e orientações para quem você ama",
	OGDesc:      "Transmita o que importa para as pessoas certas, quando for a hora. Organize com cuidado, no seu tempo.",
	Language:    "Portuguese",
	Locale:      "pt_BR",
	HTMLLang:    "pt-BR",
}

// GetPreferredLanguage detecta o idioma preferido do usuário pelo header Accept-Language
func GetPreferredLanguage(r *http.Request) string {
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang == "" {
		return "pt-BR"
	}

	// Parsear o header Accept-Language
	// Formato: pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7
	langs := strings.Split(acceptLang, ",")
	for _, lang := range langs {
		// Remover qualidade (q=)
		parts := strings.Split(strings.TrimSpace(lang), ";")
		langCode := strings.TrimSpace(parts[0])

		// Verificar se é inglês (prioridade para o primeiro idioma)
		if strings.HasPrefix(strings.ToLower(langCode), "en") {
			return "en"
		}

		// Verificar se é português
		if strings.HasPrefix(strings.ToLower(langCode), "pt") {
			return "pt-BR"
		}
	}

	return "pt-BR"
}

// GetMetaTags retorna as meta tags para o idioma especificado
func GetMetaTags(lang string) MetaTags {
	if meta, ok := metaTagsTranslations[lang]; ok {
		return meta
	}
	return metaTagsTranslations["pt-BR"]
}

// InjectMetaTags substitui as meta tags no HTML pelo idioma detectado
func InjectMetaTags(html string, lang string) string {
	// Se é português, não precisa substituir nada
	if lang == "pt-BR" {
		return html
	}

	meta := GetMetaTags(lang)

	// Substituições
	result := html

	// HTML lang
	result = strings.Replace(result,
		`<html lang="pt-BR">`,
		`<html lang="en-US">`,
		1)

	// Title
	result = strings.Replace(result,
		`<title>`+originalTexts.Title+`</title>`,
		`<title>`+meta.Title+`</title>`,
		1)

	// Description
	result = strings.Replace(result,
		`<meta name="description" content="`+originalTexts.Description+`" />`,
		`<meta name="description" content="`+meta.Description+`" />`,
		1)

	// Keywords
	result = strings.Replace(result,
		`<meta name="keywords" content="`+originalTexts.Keywords+`" />`,
		`<meta name="keywords" content="`+meta.Keywords+`" />`,
		1)

	// Language
	result = strings.Replace(result,
		`<meta name="language" content="Portuguese" />`,
		`<meta name="language" content="English" />`,
		1)

	// Open Graph title
	result = strings.Replace(result,
		`<meta property="og:title" content="`+originalTexts.OGTitle+`" />`,
		`<meta property="og:title" content="`+meta.OGTitle+`" />`,
		1)

	// Open Graph description
	result = strings.Replace(result,
		`<meta property="og:description" content="`+originalTexts.OGDesc+`" />`,
		`<meta property="og:description" content="`+meta.OGDesc+`" />`,
		1)

	// Open Graph locale
	result = strings.Replace(result,
		`<meta property="og:locale" content="pt_BR" />`,
		`<meta property="og:locale" content="en_US" />`,
		1)

	// Open Graph image alt
	result = strings.Replace(result,
		`<meta property="og:image:alt" content="`+originalTexts.OGTitle+`" />`,
		`<meta property="og:image:alt" content="`+meta.OGTitle+`" />`,
		1)

	// Twitter title
	result = strings.Replace(result,
		`<meta name="twitter:title" content="`+originalTexts.OGTitle+`" />`,
		`<meta name="twitter:title" content="`+meta.OGTitle+`" />`,
		1)

	// Twitter description
	result = strings.Replace(result,
		`<meta name="twitter:description" content="`+originalTexts.OGDesc+`" />`,
		`<meta name="twitter:description" content="`+meta.OGDesc+`" />`,
		1)

	// Twitter image alt
	result = strings.Replace(result,
		`<meta name="twitter:image:alt" content="`+originalTexts.OGTitle+`" />`,
		`<meta name="twitter:image:alt" content="`+meta.OGTitle+`" />`,
		1)

	return result
}
