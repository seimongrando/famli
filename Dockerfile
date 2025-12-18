# ==============================================================================
# Famli - Dockerfile
# ==============================================================================
# Build multi-stage para criar uma imagem otimizada.
#
# Uso:
#   docker build -t famli .
#   docker run -p 8080:8080 -e ENV=production famli
#
# Build com argumentos:
#   docker build --build-arg VERSION=1.0.0 -t famli:1.0.0 .
# ==============================================================================

# ==============================================================================
# STAGE 1: Build do Frontend
# ==============================================================================
FROM node:20-alpine AS frontend-builder

# Definir diretório de trabalho
WORKDIR /app/frontend

# Copiar arquivos de dependências
COPY frontend/package*.json ./

# Instalar dependências (usando ci para builds mais rápidos)
RUN npm ci --silent

# Copiar código fonte
COPY frontend/ ./

# Build do frontend
RUN npm run build

# ==============================================================================
# STAGE 2: Build do Backend
# ==============================================================================
FROM golang:1.21-alpine AS backend-builder

# Definir diretório de trabalho
WORKDIR /app

# Argumentos de build
ARG VERSION=dev
ARG BUILD_TIME

# Copiar arquivos de dependências
COPY backend/go.mod backend/go.sum ./

# Baixar dependências
RUN go mod download

# Copiar código fonte
COPY backend/ ./

# Build do binário
# -s -w: Remove debug info para binário menor
# -X: Injeta variáveis de versão
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o famli main.go

# ==============================================================================
# STAGE 3: Imagem Final (Produção)
# ==============================================================================
FROM alpine:3.19

# Metadados
LABEL maintainer="Famli <famli@famli.me>"
LABEL description="Famli - Organize o que importa"
LABEL version="${VERSION:-dev}"

# Instalar certificados SSL e timezone
RUN apk --no-cache add ca-certificates tzdata

# Criar usuário não-root para segurança
RUN addgroup -S famli && adduser -S famli -G famli

# Definir diretório de trabalho
WORKDIR /app

# Copiar binário do backend
COPY --from=backend-builder /app/famli .

# Copiar frontend buildado
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Definir permissões
RUN chown -R famli:famli /app

# Trocar para usuário não-root
USER famli

# Variáveis de ambiente padrão
ENV ENV=production
ENV PORT=8080
ENV STATIC_DIR=/app/frontend/dist

# Expor porta
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Comando de inicialização
CMD ["./famli"]


