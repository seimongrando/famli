# ==============================================================================
# Famli - Makefile
# ==============================================================================
# Este arquivo contém todos os comandos para desenvolvimento, build e deploy.
#
# Uso rápido:
#   make setup   - Configuração inicial (primeira vez)
#   make dev     - Modo desenvolvimento (hot reload)
#   make run     - Modo produção local
#   make build   - Build completo para deploy
#
# Para ver todos os comandos: make help
# ==============================================================================

.PHONY: help setup macos-bootstrap dev dev-db run run-memory run-db build clean \
        frontend-install frontend-dev frontend-build frontend-icons frontend-lint \
        backend-run backend-build backend-test backend-lint \
        mobile-setup mobile-android mobile-ios mobile-sync \
        docker-build docker-run docker-stop docker-up docker-down \
        db-up db-down db-reset \
        check-deps

# ==============================================================================
# VARIÁVEIS
# ==============================================================================

# Cores para output
GREEN  := \033[0;32m
YELLOW := \033[0;33m
BLUE   := \033[0;34m
RED    := \033[0;31m
NC     := \033[0m

# Diretórios
FRONTEND_DIR := frontend
BACKEND_DIR := backend
DOCS_DIR := docs

# Banco local (PostgreSQL via docker-compose)
LOCAL_DATABASE_URL := postgres://famli:famli_dev_password@localhost:5432/famli?sslmode=disable

# Versão (para builds)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# ==============================================================================
# HELP
# ==============================================================================

help:
	@echo ""
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║                        Famli - Comandos                          ║$(NC)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(GREEN)🚀 Quick Start:$(NC)"
	@echo "  make setup          - Configuração inicial completa"
	@echo "  make macos-bootstrap - Instala dependências no macOS (Homebrew)"
	@echo "  make dev            - Inicia ambiente de desenvolvimento"
	@echo "  make dev-db         - Desenvolvimento com PostgreSQL local"
	@echo "  make run            - Build + servidor local (memória por padrão)"
	@echo "  make run-memory     - Build + servidor local (força memória)"
	@echo "  make run-db         - Build + servidor local (PostgreSQL)"
	@echo ""
	@echo "$(GREEN)🔨 Build:$(NC)"
	@echo "  make build          - Build completo (frontend + backend)"
	@echo "  make frontend-build - Build apenas do frontend (PWA)"
	@echo "  make backend-build  - Compila binário do backend"
	@echo ""
	@echo "$(GREEN)📱 Mobile (Capacitor):$(NC)"
	@echo "  make mobile-setup   - Configura projeto mobile (Android + iOS)"
	@echo "  make mobile-android - Build e abre projeto Android"
	@echo "  make mobile-ios     - Build e abre projeto iOS (requer macOS)"
	@echo "  make mobile-sync    - Sincroniza código com projetos nativos"
	@echo ""
	@echo "$(GREEN)🧪 Qualidade:$(NC)"
	@echo "  make test           - Roda todos os testes"
	@echo "  make lint           - Verifica código (lint)"
	@echo "  make check-deps     - Verifica dependências instaladas"
	@echo ""
	@echo "$(GREEN)🐳 Docker:$(NC)"
	@echo "  make docker-up      - Inicia Famli + PostgreSQL (recomendado)"
	@echo "  make docker-down    - Para todos os serviços"
	@echo "  make docker-build   - Build da imagem Docker"
	@echo "  make db-up          - Inicia apenas PostgreSQL (dev local)"
	@echo "  make db-down        - Para PostgreSQL"
	@echo "  make db-reset       - Reseta PostgreSQL (remove dados)"
	@echo ""
	@echo "$(GREEN)🧹 Utilidades:$(NC)"
	@echo "  make frontend-icons - Gera ícones PWA/App"
	@echo "  make clean          - Remove arquivos de build"
	@echo ""
	@echo "$(YELLOW)📚 Documentação:$(NC) docs/DEVELOPMENT.md"
	@echo ""

# ==============================================================================
# VERIFICAÇÃO DE DEPENDÊNCIAS
# ==============================================================================

check-deps:
	@echo ""
	@echo "$(BLUE)🔍 Verificando dependências...$(NC)"
	@echo ""
	@command -v node >/dev/null 2>&1 && echo "$(GREEN)✓$(NC) Node.js: $$(node -v)" || echo "$(RED)✗$(NC) Node.js: não encontrado"
	@command -v npm >/dev/null 2>&1 && echo "$(GREEN)✓$(NC) npm: $$(npm -v)" || echo "$(RED)✗$(NC) npm: não encontrado"
	@command -v go >/dev/null 2>&1 && echo "$(GREEN)✓$(NC) Go: $$(go version | awk '{print $$3}')" || echo "$(RED)✗$(NC) Go: não encontrado"
	@command -v git >/dev/null 2>&1 && echo "$(GREEN)✓$(NC) Git: $$(git --version | awk '{print $$3}')" || echo "$(RED)✗$(NC) Git: não encontrado"
	@echo ""

# ==============================================================================
# BOOTSTRAP (macOS)
# ==============================================================================

macos-bootstrap:
	@bash -c 'set -e; \
	if [ "$$(uname)" != "Darwin" ]; then \
		echo "macos-bootstrap é apenas para macOS."; exit 0; \
	fi; \
	if ! command -v brew >/dev/null 2>&1; then \
		echo "🍎 Instalando Homebrew..."; \
		/bin/bash -c "$$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"; \
		if [ -x /opt/homebrew/bin/brew ]; then eval "$$(/opt/homebrew/bin/brew shellenv)"; fi; \
		if [ -x /usr/local/bin/brew ]; then eval "$$(/usr/local/bin/brew shellenv)"; fi; \
	fi; \
	echo "🍺 Atualizando Homebrew..."; \
	brew update --quiet; \
	for pkg in node go git; do \
		if ! brew list --formula $$pkg >/dev/null 2>&1; then \
			echo "📦 Instalando $$pkg..."; \
			brew install $$pkg; \
		fi; \
	done'

# ==============================================================================
# SETUP INICIAL
# ==============================================================================

setup: macos-bootstrap check-deps
	@echo ""
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║                 🚀 Configurando Projeto Famli                     ║$(NC)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)📦 [1/4] Instalando dependências do frontend...$(NC)"
	@cd $(FRONTEND_DIR) && npm install --silent
	@echo "$(GREEN)    ✓ Dependências npm instaladas$(NC)"
	@echo ""
	@echo "$(YELLOW)🎨 [2/4] Tentando gerar ícones PWA...$(NC)"
	@cd $(FRONTEND_DIR) && (npm install sharp --save-dev --silent 2>/dev/null && node scripts/generate-icons.js 2>/dev/null) || echo "$(YELLOW)    ⚠ Ícones: usando placeholders (sharp não disponível)$(NC)"
	@echo ""
	@echo "$(YELLOW)🔨 [3/4] Construindo frontend...$(NC)"
	@cd $(FRONTEND_DIR) && npm run build --silent
	@echo "$(GREEN)    ✓ Frontend construído em $(FRONTEND_DIR)/dist/$(NC)"
	@echo ""
	@echo "$(YELLOW)📚 [4/4] Atualizando módulos Go...$(NC)"
	@cd $(BACKEND_DIR) && go mod tidy
	@echo "$(GREEN)    ✓ Módulos Go atualizados$(NC)"
	@echo ""
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║                     ✅ Setup Completo!                            ║$(NC)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(GREEN)Próximos passos:$(NC)"
	@echo "  $(BLUE)make run$(NC)   → Iniciar servidor de produção"
	@echo "  $(BLUE)make dev$(NC)   → Modo desenvolvimento (hot reload)"
	@echo ""
	@echo "$(GREEN)Acesse:$(NC) $(BLUE)http://localhost:8080$(NC)"
	@echo ""

# ==============================================================================
# DESENVOLVIMENTO
# ==============================================================================

dev:
	@echo ""
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║                 🔧 Modo Desenvolvimento                           ║$(NC)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)Encerrando processos anteriores nas portas 8080 e 5173...$(NC)"
	@lsof -ti tcp:8080 | xargs -r kill -9 2>/dev/null || true
	@lsof -ti tcp:5173 | xargs -r kill -9 2>/dev/null || true
	@echo "$(YELLOW)Iniciando serviços...$(NC)"
	@echo ""
	@echo "  Backend:  $(BLUE)http://localhost:8080$(NC)  (API)"
	@echo "  Frontend: $(BLUE)http://localhost:5173$(NC)  (Hot Reload)"
	@echo ""
	@echo "$(YELLOW)Pressione Ctrl+C para parar.$(NC)"
	@echo ""
	@cd $(BACKEND_DIR) && go run main.go &
	@cd $(FRONTEND_DIR) && npm run dev

dev-db:
	@echo ""
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║            🔧 Modo Desenvolvimento (PostgreSQL)                   ║$(NC)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)Encerrando processos anteriores nas portas 8080 e 5173...$(NC)"
	@lsof -ti tcp:8080 | xargs -r kill -9 2>/dev/null || true
	@lsof -ti tcp:5173 | xargs -r kill -9 2>/dev/null || true
	@echo "$(YELLOW)Certifique-se de ter o PostgreSQL em execução:$(NC)"
	@echo "  $(BLUE)make db-up$(NC)"
	@echo ""
	@echo "  Backend:  $(BLUE)http://localhost:8080$(NC)  (API)"
	@echo "  Frontend: $(BLUE)http://localhost:5173$(NC)  (Hot Reload)"
	@echo ""
	@echo "$(YELLOW)Pressione Ctrl+C para parar.$(NC)"
	@echo ""
	@cd $(BACKEND_DIR) && DATABASE_URL=$(LOCAL_DATABASE_URL) go run main.go &
	@cd $(FRONTEND_DIR) && npm run dev

# ==============================================================================
# PRODUÇÃO
# ==============================================================================

run: frontend-build
	@echo ""
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║                 🏠 Servidor Famli                                 ║$(NC)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)Encerrando processos anteriores na porta 8080...$(NC)"
	@lsof -ti tcp:8080 | xargs -r kill -9 2>/dev/null || true
	@echo "$(GREEN)Acesse:$(NC) $(BLUE)http://localhost:8080$(NC)"
	@echo ""
	@echo "$(YELLOW)Pressione Ctrl+C para parar.$(NC)"
	@echo ""
	@cd $(BACKEND_DIR) && go run main.go

run-memory: frontend-build
	@echo ""
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║            🏠 Servidor Famli (Memória)                            ║$(NC)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)Encerrando processos anteriores na porta 8080...$(NC)"
	@lsof -ti tcp:8080 | xargs -r kill -9 2>/dev/null || true
	@echo "$(GREEN)Acesse:$(NC) $(BLUE)http://localhost:8080$(NC)"
	@echo "$(YELLOW)Storage: memória (dados serão perdidos ao reiniciar)$(NC)"
	@echo ""
	@echo "$(YELLOW)Pressione Ctrl+C para parar.$(NC)"
	@echo ""
	@cd $(BACKEND_DIR) && DATABASE_URL= go run main.go

run-db: frontend-build
	@echo ""
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║         🏠 Servidor Famli (PostgreSQL)                            ║$(NC)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)Encerrando processos anteriores na porta 8080...$(NC)"
	@lsof -ti tcp:8080 | xargs -r kill -9 2>/dev/null || true
	@echo "$(YELLOW)Certifique-se de ter o PostgreSQL em execução:$(NC)"
	@echo "  $(BLUE)make db-up$(NC)"
	@echo ""
	@echo "$(GREEN)Acesse:$(NC) $(BLUE)http://localhost:8080$(NC)"
	@echo ""
	@echo "$(YELLOW)Pressione Ctrl+C para parar.$(NC)"
	@echo ""
	@cd $(BACKEND_DIR) && DATABASE_URL=$(LOCAL_DATABASE_URL) go run main.go

build: frontend-build backend-build
	@echo ""
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║                     ✅ Build Completo                            ║$(NC)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(GREEN)Arquivos gerados:$(NC)"
	@echo "  Frontend: $(BLUE)$(FRONTEND_DIR)/dist/$(NC)"
	@echo "  Backend:  $(BLUE)$(BACKEND_DIR)/famli$(NC)"
	@echo ""
	@echo "$(GREEN)Para rodar em produção:$(NC)"
	@echo "  $(BLUE)./$(BACKEND_DIR)/famli$(NC)"
	@echo ""

# ==============================================================================
# FRONTEND
# ==============================================================================

frontend-install:
	@echo "$(YELLOW)📦 Instalando dependências do frontend...$(NC)"
	@cd $(FRONTEND_DIR) && npm install

frontend-dev:
	@cd $(FRONTEND_DIR) && npm run dev

frontend-build:
	@echo "$(YELLOW)🔨 Construindo frontend (PWA)...$(NC)"
	@cd $(FRONTEND_DIR) && npm run build
	@echo "$(GREEN)✓ Frontend construído em $(FRONTEND_DIR)/dist/$(NC)"

frontend-icons:
	@echo "$(YELLOW)🎨 Gerando ícones PWA e App...$(NC)"
	@cd $(FRONTEND_DIR) && npm install sharp --save-dev 2>/dev/null || true
	@cd $(FRONTEND_DIR) && node scripts/generate-icons.js

frontend-lint:
	@echo "$(YELLOW)🔍 Verificando código do frontend...$(NC)"
	@cd $(FRONTEND_DIR) && npm run lint

# ==============================================================================
# BACKEND
# ==============================================================================

backend-run:
	@cd $(BACKEND_DIR) && go run main.go

backend-build:
	@echo "$(YELLOW)🔨 Compilando backend...$(NC)"
	@cd $(BACKEND_DIR) && go build -ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)" -o famli main.go
	@echo "$(GREEN)✓ Backend compilado em $(BACKEND_DIR)/famli$(NC)"

backend-test:
	@echo "$(YELLOW)🧪 Rodando testes do backend...$(NC)"
	@cd $(BACKEND_DIR) && go test -v ./...

backend-lint:
	@echo "$(YELLOW)🔍 Verificando código do backend...$(NC)"
	@cd $(BACKEND_DIR) && go vet ./...
	@command -v golangci-lint >/dev/null 2>&1 && cd $(BACKEND_DIR) && golangci-lint run || echo "$(YELLOW)⚠ golangci-lint não instalado$(NC)"

# ==============================================================================
# TESTES E QUALIDADE
# ==============================================================================

test: backend-test
	@echo "$(GREEN)✓ Todos os testes passaram$(NC)"

lint: frontend-lint backend-lint
	@echo "$(GREEN)✓ Lint completo$(NC)"

# ==============================================================================
# MOBILE (CAPACITOR)
# ==============================================================================

mobile-setup: frontend-build
	@echo ""
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║                 📱 Configurando Projetos Mobile                   ║$(NC)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)🤖 Adicionando plataforma Android...$(NC)"
	@cd $(FRONTEND_DIR) && npx cap add android 2>/dev/null || echo "$(YELLOW)    ℹ Android já configurado$(NC)"
	@echo ""
	@echo "$(YELLOW)🍎 Adicionando plataforma iOS...$(NC)"
	@cd $(FRONTEND_DIR) && npx cap add ios 2>/dev/null || echo "$(YELLOW)    ℹ iOS já configurado$(NC)"
	@echo ""
	@echo "$(YELLOW)🔄 Sincronizando código...$(NC)"
	@cd $(FRONTEND_DIR) && npx cap sync
	@echo ""
	@echo "$(GREEN)✅ Projetos mobile configurados!$(NC)"
	@echo ""
	@echo "$(GREEN)Próximos passos:$(NC)"
	@echo "  $(BLUE)make mobile-android$(NC) → Abrir Android Studio"
	@echo "  $(BLUE)make mobile-ios$(NC)     → Abrir Xcode (macOS)"
	@echo ""

mobile-android: frontend-build
	@echo "$(YELLOW)🤖 Preparando build Android...$(NC)"
	@cd $(FRONTEND_DIR) && npx cap sync android
	@echo "$(GREEN)✅ Abrindo Android Studio...$(NC)"
	@cd $(FRONTEND_DIR) && npx cap open android

mobile-ios: frontend-build
	@echo "$(YELLOW)🍎 Preparando build iOS...$(NC)"
	@cd $(FRONTEND_DIR) && npx cap sync ios
	@echo "$(GREEN)✅ Abrindo Xcode...$(NC)"
	@cd $(FRONTEND_DIR) && npx cap open ios

mobile-sync: frontend-build
	@echo "$(YELLOW)🔄 Sincronizando código com projetos nativos...$(NC)"
	@cd $(FRONTEND_DIR) && npx cap sync
	@echo "$(GREEN)✓ Código sincronizado$(NC)"

# ==============================================================================
# DOCKER
# ==============================================================================

docker-build:
	@echo "$(YELLOW)🐳 Construindo imagem Docker...$(NC)"
	docker build -t famli:$(VERSION) -t famli:latest .
	@echo "$(GREEN)✓ Imagem construída: famli:$(VERSION)$(NC)"

docker-run:
	@echo "$(YELLOW)🐳 Iniciando container...$(NC)"
	docker run -d \
		--name famli \
		-p 8080:8080 \
		-e ENV=production \
		famli:latest
	@echo "$(GREEN)✓ Container iniciado$(NC)"
	@echo "$(GREEN)Acesse:$(NC) $(BLUE)http://localhost:8080$(NC)"

docker-stop:
	@echo "$(YELLOW)🐳 Parando container...$(NC)"
	docker stop famli 2>/dev/null || true
	docker rm famli 2>/dev/null || true
	@echo "$(GREEN)✓ Container parado$(NC)"

# Docker Compose (com PostgreSQL)
docker-up:
	@echo "$(YELLOW)🐳 Iniciando Famli + PostgreSQL...$(NC)"
	docker-compose up -d
	@echo ""
	@echo "$(GREEN)✓ Serviços iniciados$(NC)"
	@echo "$(GREEN)Acesse:$(NC) $(BLUE)http://localhost:8080$(NC)"
	@echo ""
	@echo "$(YELLOW)Comandos úteis:$(NC)"
	@echo "  docker-compose logs -f      $(GREEN)# Ver logs$(NC)"
	@echo "  docker-compose down         $(GREEN)# Parar$(NC)"
	@echo "  docker-compose down -v      $(GREEN)# Parar e remover dados$(NC)"

docker-down:
	@echo "$(YELLOW)🐳 Parando serviços...$(NC)"
	docker-compose down
	@echo "$(GREEN)✓ Serviços parados$(NC)"

# Apenas PostgreSQL (para desenvolvimento local)
db-up:
	@echo "$(YELLOW)🐘 Iniciando PostgreSQL...$(NC)"
	docker-compose up -d postgres
	@echo ""
	@echo "$(GREEN)✓ PostgreSQL iniciado$(NC)"
	@echo ""
	@echo "$(YELLOW)Connection string:$(NC)"
	@echo "  $(BLUE)postgres://famli:famli_dev_password@localhost:5432/famli?sslmode=disable$(NC)"
	@echo ""
	@echo "$(YELLOW)Para conectar via psql:$(NC)"
	@echo "  $(BLUE)psql postgres://famli:famli_dev_password@localhost:5432/famli$(NC)"

db-down:
	@echo "$(YELLOW)🐘 Parando PostgreSQL...$(NC)"
	docker-compose stop postgres
	@echo "$(GREEN)✓ PostgreSQL parado$(NC)"

db-reset:
	@echo "$(YELLOW)🐘 Resetando PostgreSQL (remove dados!)...$(NC)"
	docker-compose down -v postgres
	docker-compose up -d postgres
	@echo "$(GREEN)✓ PostgreSQL resetado$(NC)"

# ==============================================================================
# LIMPEZA
# ==============================================================================

clean:
	@echo "$(YELLOW)🧹 Limpando arquivos de build...$(NC)"
	@rm -rf $(FRONTEND_DIR)/dist
	@rm -rf $(FRONTEND_DIR)/android
	@rm -rf $(FRONTEND_DIR)/ios
	@rm -f $(BACKEND_DIR)/famli
	@echo "$(GREEN)✓ Limpeza concluída$(NC)"

clean-all: clean
	@echo "$(YELLOW)🧹 Limpeza completa (incluindo node_modules)...$(NC)"
	@rm -rf $(FRONTEND_DIR)/node_modules
	@echo "$(GREEN)✓ Limpeza completa$(NC)"
