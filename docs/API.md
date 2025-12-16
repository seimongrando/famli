# 📡 API Reference

Este documento descreve todos os endpoints da API REST do Famli.

---

## 📋 Índice

1. [Visão Geral](#visão-geral)
2. [Autenticação](#autenticação)
3. [Endpoints](#endpoints)
4. [Códigos de Erro](#códigos-de-erro)
5. [Rate Limiting](#rate-limiting)

---

## Visão Geral

### Base URL

```
http://localhost:8080/api
```

### Content-Type

```
Content-Type: application/json
Accept: application/json
```

### Autenticação

A maioria dos endpoints requer autenticação via cookie JWT:

```
Cookie: famli_session=<jwt_token>
```

---

## Autenticação

### POST /api/auth/register

Criar nova conta.

**Request:**
```json
{
  "email": "usuario@email.com",
  "password": "senha123",
  "name": "Nome do Usuário"
}
```

**Response 201:**
```json
{
  "user": {
    "id": "usr_abc123",
    "email": "usuario@email.com",
    "name": "Nome do Usuário",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Headers de Resposta:**
```
Set-Cookie: famli_session=<jwt>; Path=/; HttpOnly; Secure; SameSite=Lax
```

**Erros:**
- `400`: Email inválido ou senha fraca
- `409`: Email já cadastrado
- `429`: Rate limit excedido

---

### POST /api/auth/login

Fazer login.

**Request:**
```json
{
  "email": "usuario@email.com",
  "password": "senha123"
}
```

**Response 200:**
```json
{
  "user": {
    "id": "usr_abc123",
    "email": "usuario@email.com",
    "name": "Nome do Usuário"
  }
}
```

**Erros:**
- `400`: Credenciais inválidas
- `429`: Rate limit excedido (muitas tentativas)

---

### POST /api/auth/logout

Fazer logout.

**Requer autenticação:** ✅

**Response 200:**
```json
{
  "message": "Logout realizado com sucesso"
}
```

---

### GET /api/auth/me

Obter dados do usuário autenticado.

**Requer autenticação:** ✅

**Response 200:**
```json
{
  "user": {
    "id": "usr_abc123",
    "email": "usuario@email.com",
    "name": "Nome do Usuário",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

---

## Caixa Famli

### GET /api/box/items

Listar todos os itens do usuário.

**Requer autenticação:** ✅

**Response 200:**
```json
{
  "items": [
    {
      "id": "itm_abc123",
      "type": "info",
      "title": "Plano de Saúde",
      "content": "Número: 123456...",
      "category": "saúde",
      "is_important": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1
}
```

---

### POST /api/box/items

Criar novo item.

**Requer autenticação:** ✅

**Request:**
```json
{
  "type": "info",
  "title": "Plano de Saúde",
  "content": "Número do cartão: 123456...",
  "category": "saúde",
  "is_important": true
}
```

**Tipos válidos:**
- `info`: Informação importante
- `memory`: Memória/mensagem
- `note`: Nota pessoal
- `access`: Instruções de acesso
- `routine`: Rotina
- `location`: Localização

**Categorias válidas:**
- `saúde`
- `finanças`
- `família`
- `documentos`
- `memórias`
- `outros`

**Response 201:**
```json
{
  "id": "itm_abc123",
  "type": "info",
  "title": "Plano de Saúde",
  "content": "Número do cartão: 123456...",
  "category": "saúde",
  "is_important": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Erros:**
- `400`: Dados inválidos
- `401`: Não autenticado

---

### PUT /api/box/items/{itemID}

Atualizar item existente.

**Requer autenticação:** ✅

**Request:**
```json
{
  "title": "Plano de Saúde Atualizado",
  "content": "Novo conteúdo...",
  "is_important": false
}
```

**Response 200:**
```json
{
  "id": "itm_abc123",
  "type": "info",
  "title": "Plano de Saúde Atualizado",
  "content": "Novo conteúdo...",
  "updated_at": "2024-01-15T11:00:00Z"
}
```

**Erros:**
- `404`: Item não encontrado

---

### DELETE /api/box/items/{itemID}

Excluir item.

**Requer autenticação:** ✅

**Response 200:**
```json
{
  "message": "Item excluído com sucesso"
}
```

---

## Guardiões

### GET /api/guardians

Listar pessoas de confiança.

**Requer autenticação:** ✅

**Response 200:**
```json
{
  "guardians": [
    {
      "id": "grd_abc123",
      "name": "Maria Silva",
      "email": "maria@email.com",
      "phone": "+5511999999999",
      "relationship": "filho",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1
}
```

---

### POST /api/guardians

Adicionar pessoa de confiança.

**Requer autenticação:** ✅

**Request:**
```json
{
  "name": "Maria Silva",
  "email": "maria@email.com",
  "phone": "+5511999999999",
  "relationship": "filho"
}
```

**Relacionamentos válidos:**
- `filho`
- `neto`
- `conjuge`
- `irmao`
- `amigo`
- `outro`

**Response 201:**
```json
{
  "id": "grd_abc123",
  "name": "Maria Silva",
  "email": "maria@email.com",
  "phone": "+5511999999999",
  "relationship": "filho",
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

### DELETE /api/guardians/{guardianID}

Remover pessoa de confiança.

**Requer autenticação:** ✅

**Response 200:**
```json
{
  "message": "Guardião removido com sucesso"
}
```

---

## Guia Famli

### GET /api/guide/cards

Listar cards do guia.

**Requer autenticação:** ✅

**Response 200:**
```json
{
  "cards": [
    {
      "id": "welcome",
      "title": "Comece por aqui",
      "description": "Dê o primeiro passo...",
      "order": 1
    }
  ]
}
```

---

### GET /api/guide/progress

Obter progresso do usuário.

**Requer autenticação:** ✅

**Response 200:**
```json
{
  "progress": {
    "welcome": "completed",
    "people": "started",
    "locations": "pending"
  },
  "completed": 1,
  "total": 6
}
```

---

### POST /api/guide/progress/{cardID}

Marcar progresso em um card.

**Requer autenticação:** ✅

**Request:**
```json
{
  "status": "completed"
}
```

**Status válidos:**
- `pending`
- `started`
- `completed`
- `skipped`

**Response 200:**
```json
{
  "card_id": "welcome",
  "status": "completed"
}
```

---

## Assistente

### POST /api/assistant

Enviar pergunta para o assistente.

**Requer autenticação:** ✅

**Request:**
```json
{
  "input": "Como faço para começar?"
}
```

**Response 200:**
```json
{
  "reply": "Que bom que você quer começar! O primeiro passo é..."
}
```

---

## Configurações

### GET /api/settings

Obter configurações do usuário.

**Requer autenticação:** ✅

**Response 200:**
```json
{
  "language": "pt-BR",
  "notifications_enabled": true,
  "emergency_protocol_enabled": false
}
```

---

### PUT /api/settings

Atualizar configurações.

**Requer autenticação:** ✅

**Request:**
```json
{
  "language": "en",
  "notifications_enabled": false
}
```

**Response 200:**
```json
{
  "language": "en",
  "notifications_enabled": false,
  "emergency_protocol_enabled": false
}
```

---

## WhatsApp

### GET /api/whatsapp/status

Verificar status da integração WhatsApp.

**Response 200:**
```json
{
  "enabled": true,
  "phone_number": "whatsapp:+14155238886"
}
```

---

### POST /api/whatsapp/link

Vincular conta ao WhatsApp.

**Requer autenticação:** ✅

**Request:**
```json
{
  "phone": "+5511999999999"
}
```

**Response 200:**
```json
{
  "message": "Código de vinculação enviado",
  "phone": "+5511999999999"
}
```

---

### DELETE /api/whatsapp/link

Desvincular conta do WhatsApp.

**Requer autenticação:** ✅

**Response 200:**
```json
{
  "message": "Conta desvinculada"
}
```

---

## Códigos de Erro

| Código | Descrição |
|--------|-----------|
| 200 | Sucesso |
| 201 | Criado com sucesso |
| 400 | Requisição inválida |
| 401 | Não autenticado |
| 403 | Acesso negado |
| 404 | Não encontrado |
| 409 | Conflito (ex: email já existe) |
| 429 | Rate limit excedido |
| 500 | Erro interno |

### Formato de Erro

```json
{
  "error": "Mensagem de erro"
}
```

---

## Rate Limiting

| Endpoint | Limite | Janela |
|----------|--------|--------|
| POST /api/auth/login | 5 | 1 minuto |
| POST /api/auth/register | 3 | 1 hora |
| Outros endpoints | 60 | 1 minuto |

**Headers de Rate Limit:**
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 58
X-RateLimit-Reset: 1642255200
```

**Resposta 429:**
```json
{
  "error": "Muitas requisições. Tente novamente em 60 segundos."
}
```

---

*Última atualização: Dezembro 2024*

