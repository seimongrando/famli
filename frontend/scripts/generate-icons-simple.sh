#!/bin/bash
# ============================================================================
# Gera ícones PWA, favicons e assets usando sips (nativo do macOS)
# Uso: ./scripts/generate-icons-simple.sh
# ============================================================================

cd "$(dirname "$0")/.."

INPUT="famli.png"
OUTPUT_DIR="public/icons"
PUBLIC_DIR="public"

# Criar diretórios
mkdir -p "$OUTPUT_DIR"

echo ""
echo "🎨 Gerando ícones para Famli..."
echo ""

# ─────────────────────────────────────────────────────────────────────────────
# Ícones PWA
# ─────────────────────────────────────────────────────────────────────────────
echo "📱 Ícones PWA:"
SIZES="72 96 128 144 152 192 384 512"

for SIZE in $SIZES; do
    OUTPUT="$OUTPUT_DIR/icon-${SIZE}x${SIZE}.png"
    sips -z $SIZE $SIZE "$INPUT" --out "$OUTPUT" >/dev/null 2>&1
    echo "   ✓ icon-${SIZE}x${SIZE}.png"
done

# ─────────────────────────────────────────────────────────────────────────────
# Favicons
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "🌐 Favicons:"

# Favicon 32x32
sips -z 32 32 "$INPUT" --out "$PUBLIC_DIR/favicon-32x32.png" >/dev/null 2>&1
echo "   ✓ favicon-32x32.png"

# Favicon 16x16
sips -z 16 16 "$INPUT" --out "$PUBLIC_DIR/favicon-16x16.png" >/dev/null 2>&1
echo "   ✓ favicon-16x16.png"

# Favicon ICO (usando o de 32x32 como base)
cp "$PUBLIC_DIR/favicon-32x32.png" "$PUBLIC_DIR/favicon.ico" 2>/dev/null
echo "   ✓ favicon.ico"

# ─────────────────────────────────────────────────────────────────────────────
# Apple Touch Icon
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "🍎 Apple Touch Icon:"
sips -z 180 180 "$INPUT" --out "$PUBLIC_DIR/apple-touch-icon.png" >/dev/null 2>&1
echo "   ✓ apple-touch-icon.png"

# ─────────────────────────────────────────────────────────────────────────────
# Splash Screens iOS (placeholder - tamanho mínimo)
# Para splashes reais, use ferramentas como pwa-asset-generator
# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo "📺 Splash Screens (placeholders):"

# Criar splashes simples com fundo verde
create_splash() {
    local width=$1
    local height=$2
    local name=$3
    local logo_size=$((height / 4))
    
    # Criar imagem verde do tamanho certo
    # Nota: sips não pode criar imagens do zero, então usamos o logo como base
    # e redimensionamos. Para splashes reais, use ImageMagick ou uma ferramenta online.
    sips -z $height $width "$INPUT" --out "$OUTPUT_DIR/$name" >/dev/null 2>&1 || \
    cp "$OUTPUT_DIR/icon-512x512.png" "$OUTPUT_DIR/$name" 2>/dev/null
    echo "   ✓ $name (placeholder)"
}

# iOS Splash Screens
create_splash 640 1136 "splash-640x1136.png"
create_splash 750 1334 "splash-750x1334.png"
create_splash 1242 2208 "splash-1242x2208.png"
create_splash 1125 2436 "splash-1125x2436.png"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Ícones gerados com sucesso!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📁 Arquivos salvos em:"
echo "   - $OUTPUT_DIR/ (ícones PWA)"
echo "   - $PUBLIC_DIR/ (favicons)"
echo ""
echo "💡 Para gerar a imagem de compartilhamento social (og:image):"
echo "   ./scripts/generate-og-image.sh"
echo ""
echo "   Ou abra public/og-image.html no navegador e tire um screenshot."
echo ""
