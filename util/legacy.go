package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"

	"github.com/musenwill/mypass/errs"
)

// The v1 crypto, before authenticated encryption and slow key derivation.
// It is kept only to read a v1 store in `mypass upgrade`, never use it for new data.
// There is no integrity check, a wrong key may decrypt into garbage without an error.

// NewLegacyCrypto encrypted the whole store, key is sha256 of the pincode.
func NewLegacyCrypto(pincode []byte) CryptoApi {
	digest := sha256.Sum256(pincode)
	return &legacyCrypto{key: digest[:]}
}

// NewLegacyHMacCrypto encrypted one record, key is hmac of the title and the token.
func NewLegacyHMacCrypto(title, token []byte) CryptoApi {
	return &legacyCrypto{key: hmacSha256(title, token)}
}

type legacyCrypto struct {
	key []byte
}

var _ CryptoApi = (*legacyCrypto)(nil)

func (p *legacyCrypto) Encrypt(content []byte) ([]byte, error) {
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData := p.pkcs7Padding(content, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, p.key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func (p *legacyCrypto) Decrypt(content []byte) ([]byte, error) {
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	if len(content) == 0 || len(content)%blockSize != 0 {
		return nil, errs.DecryptError
	}
	blockMode := cipher.NewCBCDecrypter(block, p.key[:blockSize])
	origData := make([]byte, len(content))
	blockMode.CryptBlocks(origData, content)
	return p.pkcs7UnPadding(origData)
}

func (p *legacyCrypto) pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// The check is weak on purpose, it is what v1 did. Callers of the legacy crypto
// must validate the plaintext themselves.
func (p *legacyCrypto) pkcs7UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length <= 0 {
		return origData, nil
	}
	unpadding := int(origData[length-1])
	if length-unpadding < 0 {
		return nil, errs.DecryptError
	}

	return origData[:(length - unpadding)], nil
}
