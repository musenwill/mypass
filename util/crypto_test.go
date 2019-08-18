package util

import (
	"strings"
	"testing"
)

func TestKenGen256(t *testing.T) {
	api := &crypto{}
	key := api.KeyGen256("hello", []byte("world"))

	if act, exp := len(key), 256/8; act != exp {
		t.Errorf("got %v key len expected %v", act, exp)
	}
}

func TestEncryptAndDecrypt(t *testing.T) {
	origin := "world"

	api := &crypto{}
	key := api.KeyGen256("hello", []byte(origin))

	cipherText, err := api.Encrypt(key, []byte(origin))
	if err != nil {
		t.Error(err)
	}
	plainText, err := api.Decrypt(key, cipherText)
	if err != nil {
		t.Error(err)
	}

	if act, exp := string(plainText), origin; act != exp {
		t.Errorf("got %v decrypted expected %v", act, exp)
	}
}

func BenchmarkKeyGen256(b *testing.B) {
	msg1024k := strings.Repeat("647eyrughq^&IYE9rt3qw48iu}ODF3-4eri0pwj{FDL:wq4e0$ ", 26215)
	msg1024k = msg1024k[:1024*1024]
	if act, exp := len(msg1024k), 1024*1024; act != exp {
		b.Errorf("got %v message length expected %v", act, exp)
	}

	api := &crypto{}
	for i := 0; i < b.N; i++ {
		api.KeyGen256("123456", []byte(msg1024k))
	}
}
