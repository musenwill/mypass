package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
)

type Api interface {
	KeyGen256(pincode string, message []byte) []byte
	Encrypt(key, content []byte) ([]byte, error)
	Decrypt(key, content []byte) ([]byte, error)
}

type crypto struct{}

func ensureCryptoImplApi() {
	var _ Api = &crypto{}
}

func (p *crypto) KeyGen256(pincode string, message []byte) []byte {
	h := hmac.New(sha256.New, []byte(pincode))
	h.Write(message)
	return h.Sum(nil)
}

func (p *crypto) Encrypt(key, content []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData := p.pkcs7Padding(content, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func (p *crypto) Decrypt(key, content []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
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
