# 🔐 Configuração de Login Social (OAuth)

Este documento descreve como configurar a integração com Google e Apple para login social no Famli.

---

## 📋 Índice

1. [Visão Geral](#visão-geral)
2. [Google Sign-In](#google-sign-in)
3. [Apple Sign-In](#apple-sign-in)
4. [Variáveis de Ambiente](#variáveis-de-ambiente)
5. [Testando](#testando)

---

## Visão Geral

O Famli suporta login via:
- **Google** - Google Identity Services (GIS)
- **Apple** - Sign in with Apple

Os botões de login social aparecem automaticamente na página de autenticação quando configurados.

### Fluxo de Autenticação

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Frontend  │────▶│   Google/   │────▶│   Backend   │
│  (AuthPage) │     │   Apple     │     │  (OAuth)    │
└─────────────┘     └─────────────┘     └─────────────┘
      │                    │                   │
      │  1. Clica botão    │                   │
      │─────────────────▶  │                   │
      │                    │                   │
      │  2. Login no       │                   │
      │     provedor       │                   │
      │◀─────────────────  │                   │
      │                    │                   │
      │  3. Recebe token   │                   │
      │◀─────────────────  │                   │
      │                    │                   │
      │  4. Envia token ao backend             │
      │────────────────────────────────────────▶
      │                                        │
      │  5. Backend valida token               │
      │                                        │
      │  6. Cria/atualiza usuário              │
      │                                        │
      │  7. Retorna sessão JWT                 │
      │◀────────────────────────────────────────
```

---

## Google Sign-In

### Passo 1: Criar Projeto no Google Cloud Console

1. Acesse: https://console.cloud.google.com
2. Crie um novo projeto ou selecione um existente
3. Vá em **APIs & Services** → **Credentials**

### Passo 2: Configurar Tela de Consentimento OAuth

1. **APIs & Services** → **OAuth consent screen**
2. Escolha **External** (para usuários fora da organização)
3. Preencha:
   - **App name**: Famli
   - **User support email**: seu-email@exemplo.com
   - **App logo**: (opcional)
   - **App domain**: https://famli.net (ou seu domínio)
   - **Developer contact**: seu-email@exemplo.com
4. **Scopes**: Adicione `email` e `profile`
5. **Test users**: Adicione seu email para testes
6. Clique em **Publish App** quando estiver pronto para produção

### Passo 3: Criar Credenciais OAuth 2.0

1. **APIs & Services** → **Credentials**
2. **Create Credentials** → **OAuth client ID**
3. **Application type**: Web application
4. **Name**: Famli Web
5. **Authorized JavaScript origins**:
   - `http://localhost:5173` (desenvolvimento)
   - `http://localhost:8080` (desenvolvimento)
   - `https://famli.net` (produção)
   - `https://www.famli.net` (produção)
6. **Authorized redirect URIs**: (não necessário para GIS)
7. **Create**
8. **Copie o Client ID** (formato: `xxxxxxxxxxxx.apps.googleusercontent.com`)

### Passo 4: Configurar no Famli

```bash
export GOOGLE_CLIENT_ID="xxxxxxxxxxxx.apps.googleusercontent.com"
```

---

## Apple Sign-In

### Passo 1: Apple Developer Account

1. Acesse: https://developer.apple.com
2. Você precisa de uma conta **Apple Developer** ($99/ano)

### Passo 2: Registrar App ID

1. **Certificates, Identifiers & Profiles** → **Identifiers**
2. **App IDs** → **+** (novo)
3. Selecione **App IDs**
4. Selecione **App**
5. Preencha:
   - **Description**: Famli
   - **Bundle ID**: `com.famli.app` (ou seu bundle)
6. **Capabilities**: Marque **Sign in with Apple**
7. **Continue** → **Register**

### Passo 3: Criar Service ID (para Web)

1. **Identifiers** → **+** (novo)
2. Selecione **Services IDs**
3. Preencha:
   - **Description**: Famli Web
   - **Identifier**: `com.famli.web` (único)
4. **Continue** → **Register**
5. Clique no Service ID criado
6. Marque **Sign in with Apple** → **Configure**
7. Configure:
   - **Primary App ID**: selecione o App ID criado
   - **Domains**: `famli.net` (seu domínio)
   - **Return URLs**: `https://famli.net/auth`
8. **Save** → **Continue** → **Save**

### Passo 4: Criar Key para Sign in with Apple

1. **Keys** → **+** (nova)
2. **Key Name**: Famli Sign In Key
3. Marque **Sign in with Apple** → **Configure**
4. **Primary App ID**: selecione o App ID
5. **Save** → **Continue** → **Register**
6. **Download** a chave (.p8) - **guarde com segurança!**
7. Anote o **Key ID** (10 caracteres)

### Passo 5: Configurar no Famli

```bash
export APPLE_CLIENT_ID="com.famli.web"
export APPLE_TEAM_ID="XXXXXXXXXX"  # Seu Team ID (visível no canto superior direito)
export APPLE_KEY_ID="XXXXXXXXXX"   # Key ID da chave criada
export APPLE_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----
MIGTAgEAM... (conteúdo do arquivo .p8)
-----END PRIVATE KEY-----"
```

> ⚠️ **Importante**: A chave privada deve ser armazenada de forma segura. No Render, use uma variável de ambiente do tipo **Secret**.

---

## Variáveis de Ambiente

### Resumo

| Variável | Obrigatório | Descrição |
|----------|-------------|-----------|
| `GOOGLE_CLIENT_ID` | Não | Client ID do Google OAuth |
| `APPLE_CLIENT_ID` | Não | Service ID da Apple (ex: com.famli.web) |
| `APPLE_TEAM_ID` | Não* | Team ID da Apple Developer |
| `APPLE_KEY_ID` | Não* | Key ID da chave privada |
| `APPLE_PRIVATE_KEY` | Não* | Conteúdo da chave privada (.p8) |

> *Obrigatório se `APPLE_CLIENT_ID` estiver configurado

### Exemplo Completo (.env)

```bash
# OAuth - Google
GOOGLE_CLIENT_ID=123456789-xxxxx.apps.googleusercontent.com

# OAuth - Apple
APPLE_CLIENT_ID=com.famli.web
APPLE_TEAM_ID=ABCD123456
APPLE_KEY_ID=XYZ1234567
APPLE_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBHkwdwIBAQQg...
-----END PRIVATE KEY-----"
```

### No Render

1. **Dashboard** → Seu serviço → **Environment**
2. Adicione cada variável como **Secret**
3. Para `APPLE_PRIVATE_KEY`:
   - Cole o conteúdo completo do arquivo .p8
   - Inclua as linhas `-----BEGIN...` e `-----END...`

---

## Testando

### Verificar Status do OAuth

```bash
curl https://seu-dominio.com/api/auth/oauth/status
```

Resposta esperada:

```json
{
  "google": {
    "enabled": true,
    "client_id": "123456789-xxxxx.apps.googleusercontent.com"
  },
  "apple": {
    "enabled": true,
    "client_id": "com.famli.web"
  }
}
```

### Testar Login

1. Acesse a página de login do Famli
2. Se OAuth estiver configurado, os botões aparecerão automaticamente
3. Clique em "Continuar com Google" ou "Continuar com Apple"
4. Complete o fluxo de autenticação
5. Você será redirecionado para o dashboard

### Logs de Debug

Se algo não funcionar, verifique os logs do backend:

```bash
# Render
Dashboard → Logs

# Local
./server 2>&1 | grep -i oauth
```

---

## Troubleshooting

### Google

| Problema | Causa | Solução |
|----------|-------|---------|
| Botão não aparece | `GOOGLE_CLIENT_ID` vazio | Configure a variável |
| Erro de origem | Origem não autorizada | Adicione a URL em "Authorized JavaScript origins" |
| Popup bloqueado | Browser bloqueando | Desabilite bloqueador de popup |
| "access_denied" | App não publicado | Publique o app no OAuth consent screen |

### Apple

| Problema | Causa | Solução |
|----------|-------|---------|
| Botão não aparece | Variáveis não configuradas | Configure todas as variáveis Apple |
| "invalid_client" | Service ID incorreto | Verifique `APPLE_CLIENT_ID` |
| "invalid_grant" | Domínio não autorizado | Adicione o domínio no Service ID |
| Chave inválida | Formato incorreto | Inclua `-----BEGIN...` e `-----END...` |

---

## Segurança

- ✅ Tokens são validados no backend (nunca no frontend)
- ✅ Senhas não são armazenadas para usuários sociais
- ✅ Email é obtido diretamente do provedor (não pode ser falsificado)
- ✅ Usuários podem vincular conta social a conta existente (mesmo email)
- ✅ Sessão JWT segue as mesmas regras de login tradicional

---

*Última atualização: Dezembro 2024*

