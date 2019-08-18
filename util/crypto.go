package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
)

type CryptoApi interface {
	GenKey256(pincode, message []byte) []byte
	Encrypt(content []byte) ([]byte, error)
	Decrypt(content []byte) ([]byte, error)
}

func NewCrypto(pincode, token []byte) CryptoApi {
	c := &crypto{}
	c.key = c.GenKey256(pincode, token)
	return c
}

type crypto struct {
	key []byte
}

func ensureCryptoImplApi() {
	var _ CryptoApi = &crypto{}
}

func (p *crypto) GenKey256(pincode, message []byte) []byte {
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
	origData = p.pkcs7UnPadding(origData)
	return origData, nil
}

func (p *crypto) pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (p *crypto) pkcs7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
