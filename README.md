# 🏠 Famli

> Um espaço simples, seguro e humano para organizar informações, orientações e memórias importantes — para que quem você ama saiba exatamente o que fazer quando precisar.

[![License](https://img.shields.io/badge/license-Proprietary-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![Node](https://img.shields.io/badge/Node-18+-339933.svg)](https://nodejs.org/)

---

## 📋 Índice

- [Sobre o Projeto](#-sobre-o-projeto)
- [Quick Start](#-quick-start)
- [Funcionalidades](#-funcionalidades)
- [Documentação](#-documentação)
- [Comandos Úteis](#-comandos-úteis)
- [Segurança](#-segurança)
- [Roadmap](#-roadmap)

---

## 📖 Sobre o Projeto

O **Famli** resolve um problema real: pessoas 50+ acumulam ao longo da vida informações importantes espalhadas (saúde, finanças, contatos), responsabilidades que só aparecem em momentos críticos, e mensagens que fazem sentido apenas no contexto familiar.

### O Famli organiza cuidado, não apenas dados.

**Público-alvo:**
- 👴 Pessoas 50+ que querem deixar tudo organizado
- 👨‍👩‍👧 Famílias que não querem deixar ninguém no escuro
- 👩‍⚕️ Filhos, netos e cuidadores que precisam saber onde encontrar o que é importante

---

## 🚀 Quick Start

### Pré-requisitos

- [Node.js 18+](https://nodejs.org/)
- [Go 1.21+](https://go.dev/)
- [Docker](https://www.docker.com/) (para PostgreSQL)
- [Git](https://git-scm.com/)

### Opção 1: Docker (Recomendado)

```bash
# 1. Clone o repositório
git clone https://github.com/seu-usuario/famli.git
cd famli

# 2. Inicie com Docker Compose (inclui PostgreSQL)
docker-compose up -d

# 3. Pronto! Acesse:
open http://localhost:8080
```

### Opção 2: Desenvolvimento Local

```bash
# 1. Clone o repositório
git clone https://github.com/seu-usuario/famli.git
cd famli

# 2. Inicie apenas o PostgreSQL via Docker
docker-compose up -d postgres

# 3. Configure as variáveis de ambiente
cp env.example .env
# Edite .env e configure DATABASE_URL:
# DATABASE_URL=postgres://famli:famli_dev_password@localhost:5432/famli?sslmode=disable

# 4. Execute o setup (instala dependências + build)
./setup.sh

# 5. Inicie o servidor
make run
```

**Pronto!** Acesse [http://localhost:8080](http://localhost:8080)

### Modo Desenvolvimento (com hot reload)

```bash
# Terminal 1: PostgreSQL
docker-compose up -d postgres

# Terminal 2: Backend
cd backend && DATABASE_URL="postgres://famli:famli_dev_password@localhost:5432/famli?sslmode=disable" go run main.go

# Terminal 3: Frontend (hot reload)
cd frontend && npm run dev
```
Acesse [http://localhost:5173](http://localhost:5173)

### Sem Banco de Dados (Apenas para Testes)

Se não configurar `DATABASE_URL`, o Famli usa armazenamento em memória (dados perdidos ao reiniciar):

```bash
./setup.sh
make run
```

---

## ✨ Funcionalidades

### 📦 Caixa Famli
Um feed unificado com tudo que você quer guardar:
- Informações importantes
- Documentos e instruções
- Memórias e mensagens

### 🗺️ Guia Famli
Cards guiados para organizar aos poucos, sem pressa.

### 📱 WhatsApp
Adicione à sua Caixa enviando mensagens pelo WhatsApp.

### 👥 Pessoas de Confiança
Registre quem pode ajudar quando precisar.

### 🤖 Assistente
Um copilot gentil para ajudar a decidir o que guardar.

### 📲 Apps Mobile
PWA instalável + apps nativos via Capacitor.

---

## 📚 Documentação

| Documento | Descrição |
|-----------|-----------|
| [DEVELOPMENT.md](docs/DEVELOPMENT.md) | Guia completo de desenvolvimento |
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | Arquitetura e estrutura do código |
| [DEPLOYMENT.md](docs/DEPLOYMENT.md) | Deploy em produção |
| [SECURITY.md](SECURITY.md) | Práticas de segurança (OWASP) |

---

## 🛠️ Comandos Úteis

```bash
# Ver todos os comandos
make help

# Setup inicial
make setup          # Instala dependências + build
./setup.sh          # Alternativa com verificação

# Desenvolvimento
make dev            # Backend + Frontend (hot reload)
make run            # Servidor de produção

# Build
make build          # Build completo
make frontend-build # Apenas frontend (PWA)
make backend-build  # Apenas backend (binário)

# Mobile
make mobile-setup   # Configura Android + iOS
make mobile-android # Abre Android Studio
make mobile-ios     # Abre Xcode

# Outros
make clean          # Remove builds
make frontend-icons # Gera ícones
```

---

## 🔐 Segurança

O Famli segue as melhores práticas do **OWASP Top 10**:

| Proteção | Status |
|----------|--------|
| Controle de Acesso | ✅ |
| Criptografia | ✅ |
| Injection | ✅ |
| Rate Limiting | ✅ |
| Headers de Segurança | ✅ |
| Auditoria | ✅ |

Veja [SECURITY.md](SECURITY.md) para detalhes completos.

---

## 🗺️ Roadmap

- [x] MVP funcional
- [x] Internacionalização (PT-BR, EN)
- [x] PWA e suporte mobile
- [x] Integração WhatsApp
- [x] Segurança OWASP
- [ ] Validação com usuários reais
- [ ] Modo guardião (visualização)
- [ ] Co-autor de confiança
- [ ] Áudio e vídeo para memórias

---

## 📄 Licença

Proprietary - All rights reserved.

---

## 🤝 Contato

- **Email**: contato@famli.net
- **Website**: https://famli.net

---

*Famli — Organizar cuidado, não apenas dados.* 💚
