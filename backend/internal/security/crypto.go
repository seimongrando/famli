// =============================================================================
// FAMLI - Módulo de Criptografia
// =============================================================================
// Este módulo fornece funções para criptografar e descriptografar dados
// sensíveis armazenados no sistema.
//
// OWASP A02:2021 – Cryptographic Failures
// - Usa AES-256-GCM para criptografia simétrica
// - Chave derivada usando Argon2id (resistente a GPU attacks)
// - Nonce único por operação (previne replay attacks)
//
// Dados que DEVEM ser criptografados:
// - Conteúdo de itens marcados como sensíveis
// - Informações de acesso (instruções de login, etc.)
// - Dados de saúde
// - Informações financeiras
// =============================================================================

package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"

	"golang.org/x/crypto/argon2"
)

// =============================================================================
// ERROS
// =============================================================================

var (
	// ErrEncryptionFailed indica falha na criptografia
	ErrEncryptionFailed = errors.New("falha ao criptografar dados")

	// ErrDecryptionFailed indica falha na descriptografia
	ErrDecryptionFailed = errors.New("falha ao descriptografar dados")

	// ErrInvalidCiphertext indica texto cifrado inválido
	ErrInvalidCiphertext = errors.New("texto cifrado inválido")

	// ErrKeyNotSet indica que a chave não foi configurada
	ErrKeyNotSet = errors.New("chave de criptografia não configurada")
)

// =============================================================================
// CONFIGURAÇÃO ARGON2
// =============================================================================

// Parâmetros Argon2id recomendados pelo OWASP
// Estes valores oferecem boa segurança com performance aceitável
const (
	argon2Time    = 3         // Número de iterações
	argon2Memory  = 64 * 1024 // 64 MB de memória
	argon2Threads = 4         // Threads paralelas
	argon2KeyLen  = 32        // Tamanho da chave (256 bits para AES-256)
	saltLen       = 16        // Tamanho do salt
)

// =============================================================================
// ENCRYPTOR
// =============================================================================

// Encryptor fornece criptografia segura para dados sensíveis
type Encryptor struct {
	// masterKey é a chave mestra derivada da senha/segredo
	masterKey []byte

	// salt usado na derivação da chave
	salt []byte

	// mu protege acesso concorrente
	mu sync.RWMutex
}

// NewEncryptor cria um novo encryptor com a chave fornecida
//
// Parâmetros:
//   - secretKey: segredo usado para derivar a chave de criptografia
//     IMPORTANTE: Use um segredo forte (mínimo 32 caracteres)
//
// Retorna:
//   - *Encryptor: encryptor configurado
//   - error: erro se a configuração falhar
func NewEncryptor(secretKey string) (*Encryptor, error) {
	if len(secretKey) < 16 {
		return nil, errors.New("chave de criptografia muito curta (mínimo 16 caracteres)")
	}

	// Gerar salt aleatório
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("erro ao gerar salt: %w", err)
	}

	// Derivar chave usando Argon2id
	key := argon2.IDKey(
		[]byte(secretKey),
		salt,
		argon2Time,
		argon2Memory,
		argon2Threads,
		argon2KeyLen,
	)

	return &Encryptor{
		masterKey: key,
		salt:      salt,
	}, nil
}

// NewEncryptorWithSalt cria um encryptor usando um salt existente
// Use isto para descriptografar dados previamente criptografados
func NewEncryptorWithSalt(secretKey string, salt []byte) (*Encryptor, error) {
	if len(secretKey) < 16 {
		return nil, errors.New("chave de criptografia muito curta")
	}

	key := argon2.IDKey(
		[]byte(secretKey),
		salt,
		argon2Time,
		argon2Memory,
		argon2Threads,
		argon2KeyLen,
	)

	return &Encryptor{
		masterKey: key,
		salt:      salt,
	}, nil
}

// =============================================================================
// CRIPTOGRAFIA
// =============================================================================

// Encrypt criptografa dados usando AES-256-GCM
//
// AES-GCM fornece:
// - Confidencialidade (dados cifrados)
// - Integridade (detecta alterações)
// - Autenticação (verifica origem)
//
// Formato do resultado:
// base64(nonce || ciphertext || tag)
//
// Parâmetros:
//   - plaintext: dados a serem criptografados
//
// Retorna:
//   - string: dados criptografados em base64
//   - error: erro se a criptografia falhar
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.masterKey == nil {
		return "", ErrKeyNotSet
	}

	// Criar cipher AES
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	// Criar GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	// Gerar nonce aleatório
	// IMPORTANTE: Nonce deve ser único para cada operação
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	// Criptografar
	// O resultado inclui: nonce + ciphertext + authentication tag
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Codificar em base64 para armazenamento seguro
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt descriptografa dados criptografados com Encrypt
//
// Parâmetros:
//   - ciphertext: dados criptografados em base64
//
// Retorna:
//   - string: dados originais
//   - error: erro se a descriptografia falhar
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.masterKey == nil {
		return "", ErrKeyNotSet
	}

	// Decodificar base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("%w: base64 inválido", ErrInvalidCiphertext)
	}

	// Criar cipher AES
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	// Criar GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	// Verificar tamanho mínimo
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	// Extrair nonce e ciphertext
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// Descriptografar e verificar integridade
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		// Não revelar detalhes do erro (pode ser timing attack)
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// GetSalt retorna o salt usado na derivação da chave
// Necessário para persistir e restaurar o encryptor
func (e *Encryptor) GetSalt() []byte {
	e.mu.RLock()
	defer e.mu.RUnlock()

	saltCopy := make([]byte, len(e.salt))
	copy(saltCopy, e.salt)
	return saltCopy
}

// =============================================================================
// FUNÇÕES AUXILIARES
// =============================================================================

// GenerateRandomKey gera uma chave aleatória segura
// Use para gerar segredos de produção
func GenerateRandomKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashSensitiveData cria um hash unidirecional de dados sensíveis
// Use para indexação sem expor o dado original
func HashSensitiveData(data string, salt []byte) string {
	hash := argon2.IDKey(
		[]byte(data),
		salt,
		1, // Menos iterações para hashing rápido
		64*1024,
		4,
		32,
	)
	return base64.StdEncoding.EncodeToString(hash)
}
