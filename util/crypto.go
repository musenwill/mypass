package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"

	"github.com/musenwill/mypass/errs"
)

const (
	// KeyLen is the length of every key, AES-256 needs it.
	KeyLen = 32
	// NonceLen is the length of the nonce, a random one is used on every encryption.
	NonceLen = 12
)

// CryptoApi encrypts content with authenticated encryption. The ciphertext reads
// back only with the same key and aad, any other key or any change on the
// ciphertext fails the decryption.
type CryptoApi interface {
	Encrypt(content []byte) ([]byte, error)
	Decrypt(content []byte) ([]byte, error)
}

// NewCrypto returns a CryptoApi of AES-256-GCM, key must be KeyLen bytes.
// aad is authenticated but not encrypted, it binds the ciphertext to its context.
func NewCrypto(key, aad []byte) CryptoApi {
	return &aeadCrypto{key, aad}
}

// RecordKey derives the key of an account from the token key, one per title.
// A password sealed for an account never opens under another one.
func RecordKey(tokenKey []byte, title string) []byte {
	return hmacSha256(tokenKey, []byte(title))
}

// RandomBytes returns n unpredictable bytes.
func RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

func hmacSha256(key, message []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(message)
	return h.Sum(nil)
}

type aeadCrypto struct {
	key []byte
	aad []byte
}

var _ CryptoApi = (*aeadCrypto)(nil)

func (p *aeadCrypto) aead() (cipher.AEAD, error) {
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// Encrypt returns nonce || ciphertext || tag
func (p *aeadCrypto) Encrypt(content []byte) ([]byte, error) {
	aead, err := p.aead()
	if err != nil {
		return nil, err
	}

	nonce, err := RandomBytes(NonceLen)
	if err != nil {
		return nil, err
	}

	return aead.Seal(nonce, nonce, content, p.aad), nil
}

func (p *aeadCrypto) Decrypt(content []byte) ([]byte, error) {
	aead, err := p.aead()
	if err != nil {
		return nil, err
	}
	if len(content) < NonceLen {
		return nil, errs.DecryptError
	}

	plain, err := aead.Open(nil, content[:NonceLen], content[NonceLen:], p.aad)
	if err != nil {
		return nil, errs.DecryptError
	}
	return plain, nil
}
