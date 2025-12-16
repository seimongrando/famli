/**
 * Script para gerar ícones PWA e App a partir do logo principal
 * 
 * Uso: node scripts/generate-icons.js
 * 
 * Requer: npm install sharp --save-dev
 */

import sharp from 'sharp';
import { mkdir } from 'fs/promises';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const rootDir = join(__dirname, '..');

// Tamanhos de ícones necessários para PWA
const iconSizes = [72, 96, 128, 144, 152, 192, 384, 512];

// Tamanhos de splash screens iOS
const splashScreens = [
  { width: 640, height: 1136, name: 'splash-640x1136.png' },
  { width: 750, height: 1334, name: 'splash-750x1334.png' },
  { width: 1242, height: 2208, name: 'splash-1242x2208.png' },
  { width: 1125, height: 2436, name: 'splash-1125x2436.png' },
];

async function generateIcons() {
  const inputPath = join(rootDir, 'famli.png');
  const outputDir = join(rootDir, 'public', 'icons');

  // Criar diretório de saída
  await mkdir(outputDir, { recursive: true });

  console.log('🎨 Gerando ícones PWA...\n');

  // Gerar ícones de diferentes tamanhos
  for (const size of iconSizes) {
    const outputPath = join(outputDir, `icon-${size}x${size}.png`);
    
    await sharp(inputPath)
      .resize(size, size, {
        fit: 'contain',
        background: { r: 53, g: 93, b: 74, alpha: 1 } // #355d4a
      })
      .png()
      .toFile(outputPath);
    
    console.log(`  ✓ icon-${size}x${size}.png`);
  }

  console.log('\n🖼️  Gerando splash screens iOS...\n');

  // Gerar splash screens
  for (const splash of splashScreens) {
    const outputPath = join(outputDir, splash.name);
    
    // Criar splash com logo centralizado
    const logoSize = Math.min(splash.width, splash.height) * 0.3;
    
    // Primeiro, redimensionar o logo
    const resizedLogo = await sharp(inputPath)
      .resize(Math.round(logoSize), Math.round(logoSize), {
        fit: 'contain',
        background: { r: 0, g: 0, b: 0, alpha: 0 }
      })
      .toBuffer();

    // Criar o splash screen com fundo e logo centralizado
    await sharp({
      create: {
        width: splash.width,
        height: splash.height,
        channels: 4,
        background: { r: 53, g: 93, b: 74, alpha: 1 } // #355d4a
      }
    })
      .composite([{
        input: resizedLogo,
        gravity: 'center'
      }])
      .png()
      .toFile(outputPath);
    
    console.log(`  ✓ ${splash.name}`);
  }

  // Gerar favicon
  const faviconPath = join(rootDir, 'public', 'favicon.ico');
  await sharp(inputPath)
    .resize(32, 32)
    .toFile(faviconPath);
  console.log('\n  ✓ favicon.ico');

  console.log('\n✅ Todos os ícones foram gerados com sucesso!');
  console.log(`   Arquivos salvos em: ${outputDir}\n`);
}

generateIcons().catch(err => {
  console.error('❌ Erro ao gerar ícones:', err.message);
  process.exit(1);
});

