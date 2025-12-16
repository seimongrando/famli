# 🚀 Guia de Deploy em Produção

Este documento descreve como fazer deploy do Famli em um ambiente de produção.

---

## 📋 Índice

1. [Requisitos](#requisitos)
2. [Build de Produção](#build-de-produção)
3. [Configuração](#configuração)
4. [Deploy Local](#deploy-local)
5. [Deploy em Cloud](#deploy-em-cloud)
6. [Docker](#docker)
7. [Monitoramento](#monitoramento)
8. [Backup](#backup)

---

## Requisitos

### Servidor

| Recurso | Mínimo | Recomendado |
|---------|--------|-------------|
| CPU | 1 vCPU | 2 vCPU |
| RAM | 512 MB | 1 GB |
| Disco | 1 GB | 5 GB |
| SO | Ubuntu 22.04+ | Ubuntu 24.04 |

### Software

- Go 1.21+ (para compilação)
- Node.js 18+ (para build do frontend)
- Nginx ou Caddy (proxy reverso)
- Certbot ou Caddy (SSL)

---

## Build de Produção

### Opção 1: Build Completo

```bash
# Clone o repositório
git clone https://github.com/seu-usuario/famli.git
cd famli

# Build completo
make build

# Arquivos gerados:
# - frontend/dist/  (frontend estático)
# - backend/famli   (binário do backend)
```

### Opção 2: Build Manual

```bash
# Frontend
cd frontend
npm ci --production
npm run build
cd ..

# Backend
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o famli main.go
```

### Opção 3: Cross-compilation

```bash
# Build para Linux (de qualquer SO)
cd backend
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o famli-linux-amd64 main.go
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o famli-linux-arm64 main.go
```

---

## Configuração

### Variáveis de Ambiente (Obrigatórias)

```bash
# Arquivo: /etc/famli/famli.env

# Ambiente
ENV=production

# Servidor
PORT=8080
STATIC_DIR=/opt/famli/frontend/dist

# Segurança (GERAR NOVOS VALORES!)
JWT_SECRET=<gerar-com-openssl-rand-base64-48>
ENCRYPTION_KEY=<gerar-com-openssl-rand-base64-48>

# WhatsApp (opcional)
TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_AUTH_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_PHONE_NUMBER=whatsapp:+14155238886
WEBHOOK_BASE_URL=https://famli.net
```

### Gerar Segredos

```bash
# JWT_SECRET (64 caracteres)
openssl rand -base64 48

# ENCRYPTION_KEY (64 caracteres)
openssl rand -base64 48
```

⚠️ **IMPORTANTE**: Use valores diferentes em produção! Nunca reutilize segredos de desenvolvimento.

---

## Deploy Local

### 1. Preparar Diretório

```bash
# Criar estrutura
sudo mkdir -p /opt/famli
sudo mkdir -p /etc/famli
sudo mkdir -p /var/log/famli

# Copiar arquivos
sudo cp -r frontend/dist /opt/famli/frontend/
sudo cp backend/famli /opt/famli/

# Criar arquivo de configuração
sudo nano /etc/famli/famli.env
```

### 2. Criar Serviço Systemd

```bash
# Arquivo: /etc/systemd/system/famli.service
sudo nano /etc/systemd/system/famli.service
```

```ini
[Unit]
Description=Famli Server
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/famli
ExecStart=/opt/famli/famli
Restart=always
RestartSec=5
EnvironmentFile=/etc/famli/famli.env

# Limites de segurança
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
ReadWritePaths=/var/log/famli

# Logging
StandardOutput=append:/var/log/famli/famli.log
StandardError=append:/var/log/famli/famli.log

[Install]
WantedBy=multi-user.target
```

### 3. Iniciar Serviço

```bash
# Recarregar systemd
sudo systemctl daemon-reload

# Habilitar inicio automático
sudo systemctl enable famli

# Iniciar serviço
sudo systemctl start famli

# Verificar status
sudo systemctl status famli
```

### 4. Configurar Nginx (Proxy Reverso)

```bash
sudo nano /etc/nginx/sites-available/famli
```

```nginx
server {
    listen 80;
    server_name famli.net www.famli.net;
    
    # Redirecionar para HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name famli.net www.famli.net;

    # SSL (Certbot irá configurar)
    ssl_certificate /etc/letsencrypt/live/famli.net/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/famli.net/privkey.pem;
    
    # Headers de segurança (redundância com app)
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Gzip
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;
    gzip_min_length 1000;

    # Proxy para o backend
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Cache de assets estáticos
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
        proxy_pass http://127.0.0.1:8080;
        expires 30d;
        add_header Cache-Control "public, no-transform";
    }
}
```

### 5. Configurar SSL

```bash
# Ativar site
sudo ln -s /etc/nginx/sites-available/famli /etc/nginx/sites-enabled/

# Testar configuração
sudo nginx -t

# Obter certificado SSL
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d famli.net -d www.famli.net

# Reiniciar Nginx
sudo systemctl restart nginx
```

### Alternativa: Caddy (SSL automático)

```bash
# Instalar Caddy
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy
```

```bash
# Arquivo: /etc/caddy/Caddyfile
sudo nano /etc/caddy/Caddyfile
```

```caddyfile
famli.net, www.famli.net {
    reverse_proxy localhost:8080

    encode gzip

    header {
        X-Frame-Options "DENY"
        X-Content-Type-Options "nosniff"
        X-XSS-Protection "1; mode=block"
    }

    log {
        output file /var/log/caddy/famli.log
    }
}
```

```bash
sudo systemctl restart caddy
```

---

## Deploy em Cloud

### AWS (EC2 + RDS)

```bash
# 1. Criar instância EC2 (t3.small)
# 2. Configurar Security Groups (80, 443, 22)
# 3. Instalar dependências

# Conectar via SSH
ssh -i sua-key.pem ubuntu@ec2-xx-xx-xx-xx.compute-1.amazonaws.com

# Instalar Go
sudo snap install go --classic

# Instalar Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Seguir passos de "Deploy Local"
```

### Google Cloud (Cloud Run)

```bash
# Build Docker
docker build -t gcr.io/seu-projeto/famli .

# Push para Container Registry
docker push gcr.io/seu-projeto/famli

# Deploy
gcloud run deploy famli \
  --image gcr.io/seu-projeto/famli \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars "ENV=production,JWT_SECRET=xxx"
```

### DigitalOcean (App Platform)

```yaml
# .do/app.yaml
name: famli
services:
  - name: api
    source_dir: /
    build_command: make build
    run_command: ./backend/famli
    environment_slug: go
    instance_size_slug: basic-xs
    envs:
      - key: ENV
        value: production
      - key: JWT_SECRET
        value: ${JWT_SECRET}
        type: SECRET
```

---

## Docker

### Dockerfile

```dockerfile
# Arquivo: Dockerfile
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.21-alpine AS backend-builder
WORKDIR /app
COPY backend/go.* ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o famli main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /app/famli .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
ENV ENV=production
ENV STATIC_DIR=/app/frontend/dist
EXPOSE 8080
CMD ["./famli"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  famli:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENV=production
      - JWT_SECRET=${JWT_SECRET}
      - ENCRYPTION_KEY=${ENCRYPTION_KEY}
    volumes:
      - famli_data:/app/data
    restart: unless-stopped

volumes:
  famli_data:
```

### Comandos Docker

```bash
# Build
docker build -t famli .

# Run
docker run -d \
  --name famli \
  -p 8080:8080 \
  -e ENV=production \
  -e JWT_SECRET=xxx \
  -e ENCRYPTION_KEY=xxx \
  famli

# Com docker-compose
docker-compose up -d
```

---

## Monitoramento

### Logs

```bash
# Ver logs do serviço
sudo journalctl -u famli -f

# Ver últimas 100 linhas
sudo tail -100 /var/log/famli/famli.log

# Buscar erros
sudo grep -i error /var/log/famli/famli.log
```

### Health Check

```bash
# Endpoint de health (criar se necessário)
curl http://localhost:8080/api/health

# Verificar status HTTP
curl -I https://famli.net
```

### Monitoramento Externo

- **Uptime**: UptimeRobot, Pingdom
- **APM**: Datadog, New Relic
- **Logs**: Logtail, Papertrail
- **Alertas**: PagerDuty, Opsgenie

---

## Backup

### MVP (MemoryStore)

⚠️ **Aviso**: No MVP, os dados estão em memória e são perdidos ao reiniciar.

### Futuro (PostgreSQL)

```bash
# Backup diário
pg_dump -h localhost -U famli -d famli > backup_$(date +%Y%m%d).sql

# Restaurar
psql -h localhost -U famli -d famli < backup_20240115.sql

# Automatizar com cron
0 2 * * * pg_dump -h localhost -U famli -d famli | gzip > /backups/famli_$(date +\%Y\%m\%d).sql.gz
```

---

## Checklist de Deploy

### Antes do Deploy

- [ ] Build de produção testado localmente
- [ ] Variáveis de ambiente configuradas
- [ ] Segredos gerados (JWT_SECRET, ENCRYPTION_KEY)
- [ ] DNS configurado
- [ ] Firewall configurado (80, 443)

### Durante o Deploy

- [ ] Copiar arquivos para servidor
- [ ] Configurar serviço systemd
- [ ] Configurar proxy reverso (Nginx/Caddy)
- [ ] Obter certificado SSL
- [ ] Testar endpoints

### Após o Deploy

- [ ] Verificar logs
- [ ] Testar funcionalidades principais
- [ ] Configurar monitoramento
- [ ] Configurar backups
- [ ] Documentar acesso ao servidor

---

## Rollback

```bash
# Se algo der errado, reverter para versão anterior

# 1. Parar serviço
sudo systemctl stop famli

# 2. Restaurar backup do binário
sudo cp /opt/famli/backups/famli-v1.0.0 /opt/famli/famli

# 3. Reiniciar
sudo systemctl start famli

# 4. Verificar logs
sudo journalctl -u famli -f
```

---

*Última atualização: Dezembro 2024*

