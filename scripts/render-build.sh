#!/usr/bin/env bash
# ==============================================================================
# Famli - Build Script para Render
# ==============================================================================
set -e

echo "🏠 Famli - Build para Render"
echo ""

ROOT_DIR=$(pwd)

# Frontend
echo "📦 Instalando dependências do frontend..."
cd "$ROOT_DIR/frontend"
npm ci

echo "🔨 Construindo frontend..."
npm run build

# Backend
echo "🔨 Compilando backend..."
cd "$ROOT_DIR/backend"
GOCACHE="$ROOT_DIR/.gocache" go build -ldflags="-s -w" -o "$ROOT_DIR/server" .
chmod +x "$ROOT_DIR/server"

echo ""
echo "✅ Build completo!"
echo "   - Frontend: $ROOT_DIR/frontend/dist/"
echo "   - Backend: $ROOT_DIR/server"
