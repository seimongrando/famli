# 🔐 Segurança do Famli

Este documento descreve as medidas de segurança implementadas no projeto Famli, seguindo as diretrizes do **OWASP Top 10 2021**.

## 📋 Índice

1. [Visão Geral](#visão-geral)
2. [OWASP Top 10 - Implementação](#owasp-top-10---implementação)
3. [Segurança do Backend](#segurança-do-backend)
4. [Segurança do Frontend](#segurança-do-frontend)
5. [Criptografia](#criptografia)
6. [Autenticação e Autorização](#autenticação-e-autorização)
7. [Auditoria e Monitoramento](#auditoria-e-monitoramento)
8. [Configurações de Produção](#configurações-de-produção)
9. [Checklist de Segurança](#checklist-de-segurança)

---

## Visão Geral

O Famli foi desenvolvido com segurança como prioridade desde o início. Dado que o sistema armazena informações pessoais e sensíveis (memórias, documentos, contatos), implementamos múltiplas camadas de proteção.

### Princípios de Segurança

- **Defense in Depth**: Múltiplas camadas de proteção
- **Least Privilege**: Acesso mínimo necessário
- **Secure by Default**: Configurações seguras por padrão
- **Zero Trust**: Não confiar implicitamente em nenhuma entrada

---

## OWASP Top 10 - Implementação

### A01:2021 – Broken Access Control ✅

**Riscos mitigados:**
- Acesso não autorizado a dados de outros usuários
- Escalação de privilégios
- Manipulação de IDs em URLs

**Implementação:**

| Medida | Arquivo | Descrição |
|--------|---------|-----------|
| Isolamento por usuário | `storage/memory.go` | Cada usuário só acessa seus próprios dados |
| JWT com verificação | `auth/middleware.go` | Validação de token em todas as requisições |
| IDs não sequenciais | `storage/memory.go` | IDs prefixados (usr_, itm_, grd_) |
| Sanitização de IDs | `box/handler.go` | Prevenção de path traversal |

```go
// Exemplo: Verificação de propriedade do item
func (s *MemoryStore) GetBoxItem(userID, itemID string) (*BoxItem, error) {
    userItems, ok := s.items[userID]  // Isolamento por usuário
    if !ok {
        return nil, ErrNotFound
    }
    // ...
}
```

---

### A02:2021 – Cryptographic Failures ✅

**Riscos mitigados:**
- Exposição de dados sensíveis
- Senhas armazenadas em texto plano
- Transmissão insegura de dados

**Implementação:**

| Medida | Arquivo | Descrição |
|--------|---------|-----------|
| bcrypt para senhas | `auth/handler.go` | Custo padrão (10) para hashing |
| AES-256-GCM | `security/crypto.go` | Criptografia de dados sensíveis |
| Argon2id | `security/crypto.go` | Derivação de chaves resistente a GPU |
| HTTPS forçado (prod) | `security/headers.go` | HSTS com preload |
| Cookies seguros | `auth/handler.go` | HttpOnly, Secure, SameSite |

```go
// Criptografia de dados sensíveis
encryptor, _ := security.NewEncryptor(secretKey)
encrypted, _ := encryptor.Encrypt(sensitiveData)
```

**Algoritmos utilizados:**
- **Senhas**: bcrypt (custo 10)
- **Dados sensíveis**: AES-256-GCM
- **Derivação de chaves**: Argon2id (3 iterações, 64MB memória)

---

### A03:2021 – Injection ✅

**Riscos mitigados:**
- SQL Injection
- XSS (Cross-Site Scripting)
- Command Injection

**Implementação:**

| Medida | Arquivo | Descrição |
|--------|---------|-----------|
| Sanitização de HTML | `security/validation.go` | Escape de entidades HTML |
| Validação de inputs | `security/validation.go` | Tipos, tamanhos, formatos |
| Detecção de SQL injection | `security/validation.go` | Padrões bloqueados |
| CSP headers | `security/headers.go` | Content-Security-Policy |

```go
// Sanitização de texto
sanitized := security.SanitizeText(input, maxLength)

// Verificação de SQL injection
if security.ContainsSQLInjection(input) {
    return "Conteúdo inválido"
}
```

**Limites de tamanho:**
- Email: 254 caracteres (RFC 5321)
- Senha: 8-128 caracteres
- Nome: 100 caracteres
- Título: 200 caracteres
- Conteúdo: 50.000 caracteres (50KB)

---

### A04:2021 – Insecure Design ✅

**Riscos mitigados:**
- Força bruta em login
- DoS (Denial of Service)
- Credential stuffing

**Implementação:**

| Medida | Arquivo | Descrição |
|--------|---------|-----------|
| Rate limiting por IP | `security/ratelimit.go` | Limite por endpoint |
| Bloqueio progressivo | `security/ratelimit.go` | Aumenta com falhas |
| Limites de requisição | `main.go` | Middleware global |
| MaxBytesReader | `box/handler.go` | Limite de body size |

**Configurações de Rate Limit:**

| Endpoint | Requisições | Janela | Bloqueio |
|----------|-------------|--------|----------|
| Login | 5 | 1 min | 15 min |
| Registro | 3 | 1 hora | 1 hora |
| API geral | 60 | 1 min | 5 min |
| Webhooks | 200 | 1 min | 1 min |

**Bloqueio progressivo (login):**
- 3 falhas → 1 minuto
- 5 falhas → 5 minutos
- 10 falhas → 30 minutos
- 15+ falhas → 1 hora

---

### A05:2021 – Security Misconfiguration ✅

**Riscos mitigados:**
- Headers de segurança faltando
- Informações expostas
- Configurações padrão inseguras

**Implementação:**

| Medida | Arquivo | Descrição |
|--------|---------|-----------|
| Security headers | `security/headers.go` | Middleware automático |
| Env-based config | `main.go` | Configurações por ambiente |
| No-cache para API | `security/headers.go` | Previne cache de dados |

**Headers de segurança configurados:**

```
Content-Security-Policy: default-src 'self'; script-src 'self'; ...
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
```

---

### A06:2021 – Vulnerable and Outdated Components ⚠️

**Status:** Parcialmente implementado

**Medidas tomadas:**
- Go modules com versões fixas
- npm audit em CI/CD (recomendado)
- Dependências mínimas

**Recomendações:**
```bash
# Backend - verificar vulnerabilidades
go mod tidy
govulncheck ./...

# Frontend - verificar vulnerabilidades
cd frontend
npm audit
npm audit fix
```

---

### A07:2021 – Identification and Authentication Failures ✅

**Riscos mitigados:**
- Senhas fracas
- Sessões roubadas
- Enumeração de usuários

**Implementação:**

| Medida | Arquivo | Descrição |
|--------|---------|-----------|
| Validação de senha | `security/validation.go` | Requisitos mínimos |
| Proteção timing attack | `auth/handler.go` | bcrypt sempre executado |
| Mensagens genéricas | `auth/handler.go` | Não revela se email existe |
| JWT com expiração | `auth/handler.go` | 7 dias |
| Cookie HttpOnly | `auth/handler.go` | Previne acesso JS |

**Requisitos de senha:**
- Mínimo 8 caracteres
- Pelo menos uma letra minúscula
- Pelo menos um número
- Máximo 128 caracteres

```go
// Proteção contra timing attack
var dummyHash = "$2a$10$dummy.hash.for.timing.attack.prevention"
passwordToCheck := dummyHash
if userExists {
    passwordToCheck = user.Password
}
bcrypt.CompareHashAndPassword([]byte(passwordToCheck), []byte(password))
```

---

### A08:2021 – Software and Data Integrity Failures ⚠️

**Status:** Parcialmente implementado

**Medidas tomadas:**
- SRI (Subresource Integrity) para CDN fonts
- Validação de JWT com algoritmo específico

**Recomendações para produção:**
- Implementar assinatura de código
- Usar cache com validação de integridade
- CI/CD com verificação de assinaturas

---

### A09:2021 – Security Logging and Monitoring Failures ✅

**Riscos mitigados:**
- Ataques não detectados
- Falta de trilha de auditoria
- Alertas de segurança ausentes

**Implementação:**

| Medida | Arquivo | Descrição |
|--------|---------|-----------|
| Audit logging | `security/audit.go` | Eventos de segurança |
| Detecção de anomalias | `security/audit.go` | Limiares de alerta |
| Logs estruturados | `security/audit.go` | JSON para parsing |
| Request ID | `main.go` | Correlação de logs |

**Eventos registrados:**
- LOGIN_SUCCESS / LOGIN_FAILED
- REGISTER
- LOGOUT
- DATA_ACCESS / DATA_CREATE / DATA_UPDATE / DATA_DELETE
- RATE_LIMIT_EXCEEDED
- UNAUTHORIZED_ACCESS
- SUSPICIOUS_ACTIVITY
- TOKEN_INVALID

**Exemplo de log:**
```json
{
  "id": "20240115143052-abc123",
  "timestamp": "2024-01-15T14:30:52Z",
  "type": "LOGIN_FAILED",
  "severity": "WARNING",
  "client_ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "result": "failure",
  "details": {
    "email": "us***@example.com"
  }
}
```

---

### A10:2021 – Server-Side Request Forgery (SSRF) ✅

**Riscos mitigados:**
- Requisições a URLs internas
- Acesso a metadados de cloud
- Port scanning interno

**Implementação:**

| Medida | Arquivo | Descrição |
|--------|---------|-----------|
| Validação de URL | `security/validation.go` | Bloqueio de IPs privados |
| Whitelist de protocolos | `security/validation.go` | Apenas http/https |

```go
// URLs bloqueadas
privatePatterns := []string{
    "://localhost",
    "://127.",
    "://10.",
    "://192.168.",
    "://172.16.", // ... até 172.31.
    "://0.0.0.0",
    "://[::1]",
}
```

---

## Segurança do Backend

### Estrutura de Arquivos

```
backend/internal/security/
├── audit.go       # Logging de eventos de segurança
├── crypto.go      # Criptografia AES-256-GCM
├── headers.go     # Headers HTTP de segurança
├── ratelimit.go   # Rate limiting por IP
└── validation.go  # Validação e sanitização
```

### Middlewares Aplicados

```go
r.Use(
    chimiddleware.RequestID,     // ID único por requisição
    chimiddleware.RealIP,        // IP real (proxy)
    chimiddleware.Logger,        // Log de requisições
    chimiddleware.Recoverer,     // Recuperar de panics
    security.HeadersMiddleware,  // Headers de segurança
    cors.Handler,                // CORS configurado
)
```

---

## Segurança do Frontend

### Content Security Policy

```
default-src 'self';
script-src 'self';
style-src 'self' 'unsafe-inline' https://fonts.googleapis.com;
font-src 'self' https://fonts.gstatic.com;
img-src 'self' data: https:;
connect-src 'self';
frame-ancestors 'none';
form-action 'self';
base-uri 'self';
object-src 'none';
upgrade-insecure-requests;
```

### Proteções Implementadas

| Proteção | Implementação |
|----------|---------------|
| XSS | CSP + sanitização de inputs |
| CSRF | SameSite cookies + origin check |
| Clickjacking | X-Frame-Options: DENY |
| HTTPS | upgrade-insecure-requests |

---

## Criptografia

### Algoritmos

| Uso | Algoritmo | Configuração |
|-----|-----------|--------------|
| Senhas | bcrypt | Custo 10 |
| Dados sensíveis | AES-256-GCM | Nonce único |
| Derivação de chaves | Argon2id | 3 iter, 64MB |
| JWT | HS256 | Segredo ≥32 chars |

### Dados Criptografados

Os seguintes tipos de dados são (ou devem ser) criptografados:
- Instruções de acesso (tipo "access")
- Informações de saúde
- Informações financeiras
- Dados marcados como sensíveis

---

## Autenticação e Autorização

### Fluxo de Autenticação

```
1. POST /api/auth/login
   ├─ Rate limit check
   ├─ Validação de email
   ├─ bcrypt compare (sempre executa)
   ├─ Gerar JWT
   └─ Set cookie HttpOnly

2. Requisições autenticadas
   ├─ Cookie famli_session
   ├─ JWT validation
   ├─ Extract user ID
   └─ Injetar no contexto
```

### Configuração de Cookies

```go
http.Cookie{
    Name:     "famli_session",
    Value:    jwtToken,
    Path:     "/",
    HttpOnly: true,              // Não acessível via JS
    Secure:   true,              // Apenas HTTPS (produção)
    SameSite: http.SameSiteLaxMode,
    Expires:  now.Add(7 * 24 * time.Hour),
}
```

---

## Auditoria e Monitoramento

### Eventos de Segurança

```go
const (
    EventLoginSuccess       // Login bem-sucedido
    EventLoginFailed        // Falha de login
    EventRateLimitExceeded  // Rate limit atingido
    EventUnauthorizedAccess // Acesso não autorizado
    EventSuspiciousActivity // Atividade suspeita
)
```

### Limiares de Alerta

| Evento | Limiar | Ação |
|--------|--------|------|
| LOGIN_FAILED | 10/min | Alerta + bloqueio |
| RATE_LIMIT_EXCEEDED | 50/min | Alerta |
| UNAUTHORIZED_ACCESS | 20/min | Alerta crítico |

---

## Configurações de Produção

### Variáveis de Ambiente Obrigatórias

```bash
# Segurança
ENV=production
JWT_SECRET=<segredo-aleatorio-minimo-32-caracteres>
ENCRYPTION_KEY=<chave-aleatoria-minimo-32-caracteres>

# Servidor
PORT=8080
STATIC_DIR=../frontend/dist

# WhatsApp (opcional)
TWILIO_ACCOUNT_SID=ACxxxxxxxxx
TWILIO_AUTH_TOKEN=xxxxxxxxx
TWILIO_PHONE_NUMBER=whatsapp:+14155238886
WEBHOOK_BASE_URL=https://famli.me
```

### Gerar Segredos Seguros

```bash
# Gerar JWT_SECRET (64 caracteres)
openssl rand -base64 48

# Gerar ENCRYPTION_KEY (64 caracteres)
openssl rand -base64 48
```

### Checklist de Deploy

- [ ] `ENV=production` definido
- [ ] `JWT_SECRET` com ≥32 caracteres aleatórios
- [ ] `ENCRYPTION_KEY` configurado
- [ ] HTTPS configurado (SSL/TLS)
- [ ] Firewall configurado (porta 443)
- [ ] Logs centralizados configurados
- [ ] Backup de dados configurado
- [ ] Monitoramento de alertas ativo

---

## Checklist de Segurança

### Antes do Deploy

- [ ] Todas as dependências atualizadas
- [ ] `npm audit` sem vulnerabilidades críticas
- [ ] `govulncheck` sem vulnerabilidades
- [ ] Variáveis de ambiente de produção definidas
- [ ] HTTPS configurado
- [ ] CSP testado e funcionando
- [ ] Rate limiting testado
- [ ] Logs de auditoria funcionando

### Testes de Segurança

```bash
# Testar headers de segurança
curl -I https://famli.me

# Verificar CSP
# Deve retornar Content-Security-Policy

# Testar rate limiting
for i in {1..10}; do curl -X POST https://famli.me/api/auth/login; done
# Deve retornar 429 Too Many Requests

# Testar HTTPS redirect
curl -I http://famli.me
# Deve redirecionar para https://
```

### Monitoramento Contínuo

- [ ] Alertas de segurança configurados
- [ ] Logs revisados periodicamente
- [ ] Dependências atualizadas mensalmente
- [ ] Testes de penetração anuais
- [ ] Revisão de código com foco em segurança

---

## Reportar Vulnerabilidades

Se você encontrar uma vulnerabilidade de segurança, por favor:

1. **NÃO** abra uma issue pública
2. Envie email para: famli@famli.me
3. Inclua:
   - Descrição detalhada
   - Passos para reproduzir
   - Impacto potencial
   - Sugestão de correção (se possível)

Respondemos em até 48 horas úteis.

---

*Última atualização: Dezembro 2024*
