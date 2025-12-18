#!/bin/bash
# ==============================================================================
# Famli - Script de Start Rápido
# ==============================================================================
# Inicia o servidor Famli após verificar se o setup foi feito.
#
# Uso:
#   ./start.sh       - Iniciar servidor
#   ./start.sh dev   - Modo desenvolvimento
# ==============================================================================

set -e

BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m'

# Navegar para o diretório do script
SCRIPT_DIR=$(dirname "$0")
cd "$SCRIPT_DIR"

# Verificar se o frontend foi buildado
if [ ! -d "frontend/dist" ]; then
    echo ""
    echo -e "${YELLOW}⚠ Frontend não encontrado. Executando setup...${NC}"
    echo ""
    ./setup.sh
fi

# Determinar modo
MODE=${1:-run}

case $MODE in
    dev|development)
        echo ""
        echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${BLUE}║             🔧 Famli - Modo Desenvolvimento                      ║${NC}"
        echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════╝${NC}"
        echo ""
        echo -e "${GREEN}Frontend:${NC} ${BLUE}http://localhost:5173${NC} (Hot Reload)"
        echo -e "${GREEN}Backend:${NC}  ${BLUE}http://localhost:8080${NC} (API)"
        echo ""
        echo -e "${YELLOW}Pressione Ctrl+C para parar.${NC}"
        echo ""
        make dev
        ;;
    run|prod|production|*)
        echo ""
        echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${BLUE}║                    🏠 Servidor Famli                             ║${NC}"
        echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════╝${NC}"
        echo ""
        echo -e "${GREEN}Acesse:${NC} ${BLUE}http://localhost:8080${NC}"
        echo ""
        echo -e "${YELLOW}Pressione Ctrl+C para parar.${NC}"
        echo ""
        cd backend && go run main.go
        ;;
esac


