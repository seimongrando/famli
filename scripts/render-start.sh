#!/usr/bin/env bash
# ==============================================================================
# Famli - Start Script para Render
# ==============================================================================
set -e

ROOT_DIR=$(pwd)

export STATIC_DIR="$ROOT_DIR/frontend/dist"
export PORT="${PORT:-10000}"
export ENV="${ENV:-production}"

echo "🚀 Iniciando Famli..."
echo "   Porta: $PORT"
echo "   Ambiente: $ENV"
echo "   Frontend: $STATIC_DIR"

# Verificar se DATABASE_URL está configurado
if [ -n "$DATABASE_URL" ]; then
    echo "   Database: PostgreSQL (conectado)"
else
    echo "   ⚠️  DATABASE_URL não configurado - usando memória"
fi

# Verificar se há admins configurados
if [ -n "$ADMIN_EMAILS" ]; then
    echo "   Admins: configurados"
else
    echo "   ⚠️  ADMIN_EMAILS não configurado"
fi

exec "$ROOT_DIR/server"
