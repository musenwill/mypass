package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"

	"github.com/musenwill/mypass/errs"
)

type CryptoApi interface {
	GenKey256(message []byte) []byte
	GenHMacKey256(pincode, message []byte) []byte
	Encrypt(content []byte) ([]byte, error)
	Decrypt(content []byte) ([]byte, error)
}

func NewCrypto(message []byte) CryptoApi {
	c := &crypto{}
	c.key = c.GenKey256(message)
	return c
}

func NewHMacCrypto(pincode, token []byte) CryptoApi {
	c := &crypto{}
	c.key = c.GenHMacKey256(pincode, token)
	return c
}

type crypto struct {
	key []byte
}

func ensureCryptoImplApi() {
	var _ CryptoApi = &crypto{}
}

func (p *crypto) GenKey256(message []byte) []byte {
	h := sha256.New()
	h.Write(message)
	return h.Sum(nil)
}

func (p *crypto) GenHMacKey256(pincode, message []byte) []byte {
	h := hmac.New(sha256.New, pincode)
	h.Write(message)
	return h.Sum(nil)
}

func (p *crypto) Encrypt(content []byte) ([]byte, error) {
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

func (p *crypto) Decrypt(content []byte) ([]byte, error) {
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, p.key[:blockSize])
	origData := make([]byte, len(content))
	blockMode.CryptBlocks(origData, content)
	return p.pkcs7UnPadding(origData)
}

func (p *crypto) pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (p *crypto) pkcs7UnPadding(origData []byte) ([]byte, error) {
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
