#!/bin/bash
# ============================================================================
# Gera imagem de compartilhamento social (og:image) - 1200x630px
# 
# Opção 1: Usar o template HTML
#   - Abra public/og-image.html no navegador
#   - Tire um screenshot de 1200x630px
#   - Salve como public/og-image.png
#
# Opção 2: Usar serviços online
#   - https://og-playground.vercel.app/
#   - https://www.bannerbear.com/
#   - https://placid.app/
#
# Opção 3: Usar este script com ImageMagick (se instalado)
# ============================================================================

cd "$(dirname "$0")/.."

OUTPUT="public/og-image.png"
LOGO="famli.png"

# Verificar se ImageMagick está instalado
if ! command -v magick &> /dev/null && ! command -v convert &> /dev/null; then
    echo "⚠️  ImageMagick não encontrado."
    echo ""
    echo "Para gerar a imagem OG, você pode:"
    echo ""
    echo "  1. Instalar ImageMagick:"
    echo "     brew install imagemagick"
    echo ""
    echo "  2. Usar o template HTML:"
    echo "     Abra public/og-image.html no navegador e tire um screenshot"
    echo ""
    echo "  3. Usar um serviço online como:"
    echo "     https://og-playground.vercel.app/"
    echo ""
    exit 1
fi

echo "🎨 Gerando imagem de compartilhamento social..."

# Determinar comando (magick para v7+, convert para v6)
if command -v magick &> /dev/null; then
    CMD="magick"
else
    CMD="convert"
fi

# Criar imagem OG
$CMD -size 1200x630 \
    -define gradient:angle=135 \
    gradient:'#355d4a-#1f3a2d' \
    -gravity center \
    \( "$LOGO" -resize 200x200 \) -geometry +0-80 -composite \
    -gravity center \
    -font "Helvetica-Bold" -pointsize 48 -fill white \
    -annotate +0+100 "Guarde o que importa para quem você ama" \
    -font "Helvetica" -pointsize 28 -fill '#f4a285' \
    -annotate +0+160 "Memórias • Documentos • Orientações" \
    -font "Helvetica" -pointsize 22 -fill 'rgba(255,255,255,0.6)' \
    -gravity south -annotate +0+40 "famli.net" \
    "$OUTPUT"

if [ $? -eq 0 ]; then
    echo "✅ Imagem gerada: $OUTPUT"
else
    echo "❌ Erro ao gerar imagem"
    echo ""
    echo "Use o template HTML como alternativa:"
    echo "  Abra public/og-image.html no navegador"
    exit 1
fi


