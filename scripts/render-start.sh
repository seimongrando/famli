#!/bin/bash
# ==============================================================================
# Famli - Start Script para Render
# ==============================================================================
# Este script é usado no campo "Start Command" do Render.
#
# Variáveis de ambiente necessárias no Render:
#   - ENV=production
#   - JWT_SECRET=<seu-segredo-jwt>
#   - ENCRYPTION_KEY=<sua-chave-criptografia>
#   - PORT (definido automaticamente pelo Render)
# ==============================================================================

set -e

echo "🏠 Iniciando servidor Famli..."
echo ""

# Diretório raiz (onde o script está sendo executado)
ROOT_DIR=$(pwd)

# Verificar se o binário existe
if [ ! -f "$ROOT_DIR/server" ]; then
    echo "❌ ERRO: Binário 'server' não encontrado em $ROOT_DIR"
    echo "   Conteúdo do diretório:"
    ls -la "$ROOT_DIR"
    exit 1
fi

# Verificar se o frontend foi buildado
if [ ! -d "$ROOT_DIR/frontend/dist" ]; then
    echo "❌ ERRO: Frontend não encontrado em $ROOT_DIR/frontend/dist"
    exit 1
fi

# Configurar variáveis de ambiente
export STATIC_DIR="$ROOT_DIR/frontend/dist"
export PORT=${PORT:-8080}
export ENV=${ENV:-production}

echo "📋 Configuração:"
echo "   - Ambiente: $ENV"
echo "   - Porta: $PORT"
echo "   - Frontend: $STATIC_DIR"
echo "   - Binário: $ROOT_DIR/server"
echo ""

# Verificar se as variáveis obrigatórias estão definidas
if [ -z "$JWT_SECRET" ]; then
    echo "⚠️  AVISO: JWT_SECRET não definido, usando valor padrão (inseguro!)"
fi

if [ -z "$ENCRYPTION_KEY" ]; then
    echo "⚠️  AVISO: ENCRYPTION_KEY não definido, usando valor padrão (inseguro!)"
fi

echo "🚀 Executando servidor..."
echo ""

# Executar o servidor (exec substitui o processo shell pelo servidor)
exec "$ROOT_DIR/server"
