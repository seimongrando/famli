#!/bin/bash
# ==============================================================================
# Famli - Build Script para Render
# ==============================================================================
# Este script é usado no campo "Build Command" do Render.
#
# No Render, configure:
#   Build Command: ./scripts/render-build.sh
#   Start Command: ./scripts/render-start.sh
#
# O que este script faz:
#   1. Instala dependências do Node.js (frontend)
#   2. Faz build do frontend (Vue + Vite)
#   3. Compila o backend (Go)
# ==============================================================================

set -e  # Parar em caso de erro

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║              🏠 Famli - Build para Render                        ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# ==============================================================================
# FRONTEND
# ==============================================================================

echo "📦 [1/3] Instalando dependências do frontend..."
cd frontend
npm ci --production=false
echo "✓ Dependências instaladas"
echo ""

echo "🔨 [2/3] Construindo frontend..."
npm run build
echo "✓ Frontend construído em frontend/dist/"
echo ""

cd ..

# ==============================================================================
# BACKEND
# ==============================================================================

echo "🔨 [3/3] Compilando backend..."
cd backend

# Configurar cache do Go (Render tem disco efêmero)
export GOCACHE=$(pwd)/.gocache
mkdir -p $GOCACHE

# Build do binário
go build -ldflags="-s -w" -o server .

echo "✓ Backend compilado em backend/server"
echo ""

# ==============================================================================
# FINALIZAÇÃO
# ==============================================================================

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║                     ✅ Build Completo!                           ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""
echo "Arquivos gerados:"
echo "  - frontend/dist/ (frontend estático)"
echo "  - backend/server (binário do backend)"
echo ""

