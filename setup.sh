#!/bin/bash
# ==============================================================================
# Famli - Script de Setup
# ==============================================================================
# Este script configura o ambiente de desenvolvimento do Famli.
#
# Uso:
#   ./setup.sh        - Configuração completa
#   ./setup.sh --help - Ver opções
#
# O que este script faz:
#   1. Verifica se Node.js, npm e Go estão instalados
#   2. Instala dependências do frontend (npm install)
#   3. Tenta gerar ícones PWA
#   4. Faz build do frontend (npm run build)
#   5. Atualiza módulos Go (go mod tidy)
#
# Documentação: docs/DEVELOPMENT.md
# ==============================================================================

set -e

# ==============================================================================
# CORES E SÍMBOLOS
# ==============================================================================

GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

CHECK="✓"
CROSS="✗"
WARN="⚠"

# ==============================================================================
# FUNÇÕES AUXILIARES
# ==============================================================================

# Exibe mensagem de sucesso
success() {
    echo -e "${GREEN}${CHECK}${NC} $1"
}

# Exibe mensagem de erro
error() {
    echo -e "${RED}${CROSS}${NC} $1"
}

# Exibe mensagem de aviso
warn() {
    echo -e "${YELLOW}${WARN}${NC} $1"
}

# Exibe mensagem de info
info() {
    echo -e "${YELLOW}→${NC} $1"
}

# Exibe mensagem de passo
step() {
    echo ""
    echo -e "${BLUE}[$1]${NC} $2"
}

# Verifica se um comando existe
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# ==============================================================================
# HELP
# ==============================================================================

show_help() {
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                    Famli - Setup Script                          ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "Uso: ./setup.sh [opções]"
    echo ""
    echo "Opções:"
    echo "  --help, -h     Mostra esta mensagem"
    echo "  --check        Apenas verifica dependências"
    echo "  --skip-icons   Pula geração de ícones"
    echo "  --skip-build   Pula build do frontend"
    echo ""
    echo "Exemplos:"
    echo "  ./setup.sh              # Setup completo"
    echo "  ./setup.sh --check      # Verifica dependências"
    echo "  ./setup.sh --skip-icons # Setup sem gerar ícones"
    echo ""
    exit 0
}

# ==============================================================================
# VERIFICAÇÃO DE DEPENDÊNCIAS
# ==============================================================================

check_dependencies() {
    echo ""
    echo -e "${BLUE}🔍 Verificando dependências...${NC}"
    echo ""

    local all_ok=true

    # Node.js
    if command_exists node; then
        success "Node.js: $(node -v)"
    else
        error "Node.js: não encontrado"
        echo "   Instale em: https://nodejs.org/"
        all_ok=false
    fi

    # npm
    if command_exists npm; then
        success "npm: $(npm -v)"
    else
        error "npm: não encontrado"
        all_ok=false
    fi

    # Go
    if command_exists go; then
        success "Go: $(go version | awk '{print $3}')"
    else
        error "Go: não encontrado"
        echo "   Instale em: https://go.dev/"
        all_ok=false
    fi

    # Git (opcional mas recomendado)
    if command_exists git; then
        success "Git: $(git --version | awk '{print $3}')"
    else
        warn "Git: não encontrado (opcional)"
    fi

    echo ""

    if [ "$all_ok" = false ]; then
        error "Algumas dependências estão faltando. Instale-as e tente novamente."
        exit 1
    fi
}

# ==============================================================================
# MAIN
# ==============================================================================

main() {
    # Processar argumentos
    SKIP_ICONS=false
    SKIP_BUILD=false
    CHECK_ONLY=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            --help|-h)
                show_help
                ;;
            --check)
                CHECK_ONLY=true
                shift
                ;;
            --skip-icons)
                SKIP_ICONS=true
                shift
                ;;
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            *)
                error "Opção desconhecida: $1"
                show_help
                ;;
        esac
    done

    # Header
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║              🏠 Famli - Configuração Inicial                     ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════╝${NC}"

    # Navegar para o diretório do script
    SCRIPT_DIR=$(dirname "$0")
    cd "$SCRIPT_DIR"

    # Verificar dependências
    check_dependencies

    # Se for apenas verificação, parar aqui
    if [ "$CHECK_ONLY" = true ]; then
        success "Todas as dependências estão instaladas!"
        exit 0
    fi

    # ==============================================================================
    # INSTALAÇÃO
    # ==============================================================================

    step "1/4" "Instalando dependências do frontend..."
    cd frontend
    npm install --silent
    success "Dependências npm instaladas"
    cd ..

    # ==============================================================================
    # ÍCONES
    # ==============================================================================

    if [ "$SKIP_ICONS" = false ]; then
        step "2/4" "Gerando ícones PWA..."
        cd frontend
        if npm install sharp --save-dev --silent 2>/dev/null && node scripts/generate-icons.js 2>/dev/null; then
            success "Ícones gerados"
        else
            warn "Ícones: usando placeholders (sharp não disponível)"
        fi
        cd ..
    else
        step "2/4" "Pulando geração de ícones..."
        warn "Ícones não gerados (--skip-icons)"
    fi

    # ==============================================================================
    # BUILD FRONTEND
    # ==============================================================================

    if [ "$SKIP_BUILD" = false ]; then
        step "3/4" "Construindo frontend..."
        cd frontend
        npm run build --silent
        success "Frontend construído em frontend/dist/"
        cd ..
    else
        step "3/4" "Pulando build do frontend..."
        warn "Frontend não construído (--skip-build)"
    fi

    # ==============================================================================
    # MÓDULOS GO
    # ==============================================================================

    step "4/4" "Atualizando módulos Go..."
    cd backend
    go mod tidy
    success "Módulos Go atualizados"
    cd ..

    # ==============================================================================
    # FINALIZAÇÃO
    # ==============================================================================

    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                     ✅ Setup Completo!                            ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${GREEN}Comandos disponíveis:${NC}"
    echo ""
    echo -e "  ${BLUE}make run${NC}            - Iniciar servidor de produção"
    echo -e "  ${BLUE}make dev${NC}            - Modo desenvolvimento (hot reload)"
    echo -e "  ${BLUE}make mobile-setup${NC}   - Configurar apps Android/iOS"
    echo -e "  ${BLUE}make help${NC}           - Ver todos os comandos"
    echo ""
    echo -e "${GREEN}Acesse:${NC} ${BLUE}http://localhost:8080${NC}"
    echo ""
    echo -e "${YELLOW}📚 Documentação:${NC} docs/DEVELOPMENT.md"
    echo ""
}

# Executar
main "$@"
