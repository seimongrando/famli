#!/bin/bash
# ==============================================================================
# Famli - Build Script para Render
# ==============================================================================
# Este script é usado no campo "Build Command" do Render.
#
# No Render, configure:
#   Build Command: ./scripts/render-build.sh
#   Start Command: ./scripts/render-start.sh
# ==============================================================================

set -e  # Parar em caso de erro

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║              🏠 Famli - Build para Render                        ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# Diretório raiz do projeto
ROOT_DIR=$(pwd)
echo "📁 Diretório raiz: $ROOT_DIR"
echo ""

# ==============================================================================
# FRONTEND
# ==============================================================================

echo "📦 [1/3] Instalando dependências do frontend..."
cd "$ROOT_DIR/frontend"
npm ci --production=false --silent
echo "✓ Dependências instaladas"
echo ""

echo "🔨 [2/3] Construindo frontend..."
npm run build
echo "✓ Frontend construído em frontend/dist/"
echo ""

# Verificar se o build foi criado
if [ ! -d "$ROOT_DIR/frontend/dist" ]; then
    echo "❌ ERRO: frontend/dist não foi criado!"
    exit 1
fi

# ==============================================================================
# BACKEND
# ==============================================================================

echo "🔨 [3/3] Compilando backend..."
cd "$ROOT_DIR/backend"

# Configurar cache do Go
export GOCACHE="$ROOT_DIR/.gocache"
mkdir -p "$GOCACHE"

# Build do binário
# Colocar o binário na raiz do projeto para facilitar
go build -ldflags="-s -w" -o "$ROOT_DIR/server" .

echo "✓ Backend compilado em $ROOT_DIR/server"
echo ""

# Verificar se o binário foi criado
if [ ! -f "$ROOT_DIR/server" ]; then
    echo "❌ ERRO: binário 'server' não foi criado!"
    exit 1
fi

# Dar permissão de execução
chmod +x "$ROOT_DIR/server"

# ==============================================================================
# FINALIZAÇÃO
# ==============================================================================

cd "$ROOT_DIR"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║                     ✅ Build Completo!                           ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""
echo "📋 Arquivos gerados:"
ls -la "$ROOT_DIR/server"
ls -la "$ROOT_DIR/frontend/dist/" | head -5
echo ""
