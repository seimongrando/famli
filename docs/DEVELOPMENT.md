# 🔧 Guia de Desenvolvimento

Este documento descreve como configurar e trabalhar no ambiente de desenvolvimento do Famli.

---

## 📋 Índice

1. [Pré-requisitos](#pré-requisitos)
2. [Configuração Inicial](#configuração-inicial)
3. [Estrutura do Projeto](#estrutura-do-projeto)
4. [Desenvolvimento](#desenvolvimento)
5. [Testes](#testes)
6. [Convenções de Código](#convenções-de-código)
7. [Troubleshooting](#troubleshooting)

---

## Pré-requisitos

### Obrigatórios

| Ferramenta | Versão | Verificar |
|------------|--------|-----------|
| Node.js | 18+ | `node -v` |
| npm | 9+ | `npm -v` |
| Go | 1.21+ | `go version` |
| Git | 2.0+ | `git --version` |

### Instalação das Dependências

**macOS (Homebrew):**
```bash
brew install node go git
```

**Ubuntu/Debian:**
```bash
# Node.js via NodeSource
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Go
sudo snap install go --classic

# Git
sudo apt-get install -y git
```

**Windows (Chocolatey):**
```powershell
choco install nodejs-lts golang git
```

### Opcionais (para desenvolvimento mobile)

| Ferramenta | Uso |
|------------|-----|
| Android Studio | Apps Android |
| Xcode (macOS) | Apps iOS |
| Java 17+ | Build Android |

---

## Configuração Inicial

### Opção 1: Setup Automático (Recomendado)

```bash
# Clone o repositório
git clone https://github.com/seu-usuario/famli.git
cd famli

# Execute o script de setup
./setup.sh
```

O script irá:
1. ✅ Verificar se todas as dependências estão instaladas
2. ✅ Instalar pacotes npm do frontend
3. ✅ Gerar ícones PWA
4. ✅ Fazer build do frontend
5. ✅ Baixar módulos Go do backend

### Opção 2: Setup Manual

```bash
# Clone
git clone https://github.com/seu-usuario/famli.git
cd famli

# Frontend
cd frontend
npm install
npm run build
cd ..

# Backend
cd backend
go mod tidy
cd ..
```

### Verificar Instalação

```bash
# Iniciar servidor
make run

# Acessar no navegador
open http://localhost:8080
```

---

## Estrutura do Projeto

```
famli/
├── 📁 backend/                 # API Go
│   ├── main.go                # Ponto de entrada
│   ├── go.mod                 # Dependências Go
│   └── 📁 internal/           # Código interno
│       ├── auth/              # Autenticação JWT
│       ├── box/               # Caixa Famli (itens)
│       ├── guardian/          # Pessoas de confiança
│       ├── guide/             # Guia Famli
│       ├── i18n/              # Internacionalização
│       ├── security/          # Segurança
│       ├── settings/          # Configurações
│       ├── storage/           # Persistência
│       └── whatsapp/          # Integração WhatsApp
│
├── 📁 frontend/               # Vue 3 + Vite
│   ├── package.json           # Dependências npm
│   ├── vite.config.js         # Config Vite + PWA
│   ├── capacitor.config.ts    # Config mobile
│   ├── index.html             # HTML principal
│   ├── 📁 public/             # Assets estáticos
│   │   ├── famli.png          # Logo principal
│   │   ├── icons/             # Ícones PWA
│   │   └── ...
│   ├── 📁 scripts/            # Scripts utilitários
│   └── 📁 src/
│       ├── main.js            # Entry point Vue
│       ├── App.vue            # Componente raiz
│       ├── 📁 components/     # Componentes Vue
│       ├── 📁 pages/          # Páginas (views)
│       ├── 📁 stores/         # Pinia stores
│       ├── 📁 i18n/           # Traduções
│       └── 📁 styles/         # CSS global
│
├── 📁 docs/                   # Documentação
│   ├── DEVELOPMENT.md         # Este arquivo
│   ├── ARCHITECTURE.md        # Arquitetura
│   └── DEPLOYMENT.md          # Deploy
│
├── Makefile                   # Comandos make
├── setup.sh                   # Script de setup
├── README.md                  # Documentação principal
└── SECURITY.md                # Segurança
```

---

## Desenvolvimento

### Modo Desenvolvimento (Recomendado)

O modo dev oferece **hot reload** para mudanças instantâneas:

```bash
make dev
```

Isso inicia:
- **Backend** em http://localhost:8080 (API)
- **Frontend** em http://localhost:5173 (com HMR)

> 💡 Acesse http://localhost:5173 para desenvolvimento com hot reload.

### Backend Standalone

```bash
# Terminal 1 - Backend
cd backend
go run main.go
```

### Frontend Standalone

```bash
# Terminal 2 - Frontend
cd frontend
npm run dev
```

### Variáveis de Ambiente

```bash
# Copie o exemplo (se existir)
cp .env.example .env

# Ou configure manualmente:
export PORT=8080
export ENV=development
export JWT_SECRET=dev-secret-change-in-production
export ENCRYPTION_KEY=dev-encryption-key-change-in-prod
```

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `PORT` | 8080 | Porta do servidor |
| `ENV` | development | Ambiente (development/production) |
| `JWT_SECRET` | (dev secret) | Segredo para tokens JWT |
| `ENCRYPTION_KEY` | (dev key) | Chave de criptografia |
| `STATIC_DIR` | ../frontend/dist | Diretório do frontend |
| `TWILIO_*` | - | Configurações WhatsApp |

---

## Testes

### Backend (Go)

```bash
cd backend
go test ./...
```

### Frontend (Vue)

```bash
cd frontend
npm run test        # Testes unitários
npm run test:e2e    # Testes E2E (se configurado)
```

### Linting

```bash
# Frontend
cd frontend
npm run lint

# Backend
cd backend
golangci-lint run
```

---

## Convenções de Código

### Go (Backend)

```go
// =============================================================================
// FAMLI - Nome do Módulo
// =============================================================================
// Descrição do que este arquivo faz.
//
// Funcionalidades:
// - Feature 1
// - Feature 2
// =============================================================================

package nomepacote

// Handler gerencia operações de X
type Handler struct {
    // store é o armazenamento de dados
    store *storage.MemoryStore
}

// NewHandler cria uma nova instância
//
// Parâmetros:
//   - store: armazenamento de dados
//
// Retorna:
//   - *Handler: handler configurado
func NewHandler(store *storage.MemoryStore) *Handler {
    return &Handler{store: store}
}
```

### Vue (Frontend)

```vue
<!-- =============================================================================
FAMLI - Nome do Componente
===============================================================================
Descrição do que este componente faz.

Props:
- prop1: Descrição
- prop2: Descrição

Emits:
- evento1: Quando X acontece
============================================================================= -->

<script setup>
// Imports organizados
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'

// =============================================================================
// PROPS E EMITS
// =============================================================================

const props = defineProps({
  // Descrição da prop
  nomeProp: {
    type: String,
    required: true
  }
})

// =============================================================================
// ESTADO
// =============================================================================

const estado = ref('')

// =============================================================================
// COMPUTED
// =============================================================================

const computedValue = computed(() => {
  return estado.value
})

// =============================================================================
// MÉTODOS
// =============================================================================

function handleClick() {
  // Implementação
}
</script>

<template>
  <div class="componente">
    <!-- Conteúdo -->
  </div>
</template>

<style scoped>
/* Estilos do componente */
</style>
```

### Commits

Seguimos o padrão [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: adiciona novo recurso
fix: corrige bug
docs: atualiza documentação
style: formatação (sem mudança de código)
refactor: refatoração de código
test: adiciona testes
chore: tarefas de manutenção
```

Exemplos:
```bash
git commit -m "feat: adiciona validação de senha com feedback visual"
git commit -m "fix: corrige redirecionamento após login"
git commit -m "docs: atualiza README com instruções de instalação"
```

---

## Troubleshooting

### Erro: "node: command not found"

**Causa:** Node.js não está instalado ou não está no PATH.

**Solução:**
```bash
# macOS
brew install node

# Ubuntu
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs
```

### Erro: "go: command not found"

**Causa:** Go não está instalado ou não está no PATH.

**Solução:**
```bash
# macOS
brew install go

# Ubuntu
sudo snap install go --classic

# Verificar GOPATH
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

### Erro: "Frontend não encontrado"

**Causa:** O frontend não foi buildado.

**Solução:**
```bash
cd frontend
npm install
npm run build
```

### Erro: Porta 8080 já em uso

**Causa:** Outro processo está usando a porta.

**Solução:**
```bash
# Encontrar processo
lsof -i :8080

# Matar processo
kill -9 <PID>

# Ou usar outra porta
PORT=3000 make run
```

### Erro: "sharp" não instala (ícones)

**Causa:** Versão do Node.js incompatível com o módulo sharp.

**Solução:**
```bash
# Usar Node.js 18+
nvm install 18
nvm use 18

# Ou ignorar e usar ícones placeholder
# Os ícones são opcionais no desenvolvimento
```

### Hot Reload não funciona

**Causa:** Problema com o Vite.

**Solução:**
```bash
# Limpar cache
rm -rf frontend/node_modules/.vite
rm -rf frontend/dist

# Reinstalar
cd frontend
npm install
npm run dev
```

---

## 📞 Suporte

Se encontrar problemas não listados aqui:

1. Verifique os [issues existentes](https://github.com/seu-usuario/famli/issues)
2. Crie um novo issue com:
   - Versão do Node/Go/npm
   - Sistema operacional
   - Passos para reproduzir
   - Mensagem de erro completa

---

*Última atualização: Dezembro 2024*

