#!/bin/bash
# ==============================================================================
# Famli - Start Script para Render
# ==============================================================================
# Este script é usado no campo "Start Command" do Render.
#
# No Render, configure:
#   Build Command: ./scripts/render-build.sh
#   Start Command: ./scripts/render-start.sh
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

# Definir diretório do frontend relativo ao backend
export STATIC_DIR=../frontend/dist

# O Render define a variável PORT automaticamente
# Se não estiver definida, usar 8080 como padrão
export PORT=${PORT:-8080}

# Garantir que estamos em produção
export ENV=${ENV:-production}

echo "📋 Configuração:"
echo "   - Ambiente: $ENV"
echo "   - Porta: $PORT"
echo "   - Frontend: $STATIC_DIR"
echo ""

# Iniciar servidor
cd backend
exec ./server

