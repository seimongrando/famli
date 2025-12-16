#!/usr/bin/env bash
# ==============================================================================
# Famli - Start Script para Render
# ==============================================================================
set -e

ROOT_DIR=$(pwd)

export STATIC_DIR="$ROOT_DIR/frontend/dist"
export PORT="${PORT:-8080}"
export ENV="${ENV:-production}"

echo "🚀 Iniciando Famli..."
echo "   Porta: $PORT"
echo "   Frontend: $STATIC_DIR"

exec "$ROOT_DIR/server"
